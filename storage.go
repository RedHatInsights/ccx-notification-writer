// Copyright 2021 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// This source file contains an implementation of interface between Go code and
// (almost any) SQL database like PostgreSQL, SQLite, or MariaDB.
//
// It is possible to configure connection to selected database by using
// StorageConfiguration structure. Currently that structure contains two
// configurable parameter:
//
// Driver - a SQL driver, like "sqlite3", "pq" etc.
// DataSource - specification of data source. The content of this parameter depends on the database used.

import (
	"errors"
	"fmt"
	"time"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"
)

// Table creation-related scripts
const (
	createTableNotificationTypes = `
                create table notification_types (
                    id          integer not null,
                    value       varchar not null,
                    frequency   varchar not null,
                    comment     varchar,
                
                    PRIMARY KEY (id)
                );
`

	createTableStates = `
                create table states (
                    id          integer not null,
                    value       varchar not null,
                    comment     varchar,
                
                    PRIMARY KEY (id)
                );
`

	createTableReported = `
                create table reported (
                    org_id            integer not null,
                    account_number    integer not null,
                    cluster           character(36) not null,
                    notification_type integer not null,
                    state             integer not null,
                    report            varchar not null,
                    updated_at        timestamp not null,
                
                    PRIMARY KEY (org_id, cluster),
                    CONSTRAINT fk_notification_type
                        foreign key(notification_type)
                        references notification_types(id),
                    CONSTRAINT fk_state
                        foreign key (state)
                        references states(id)
                );
`

	createTableNewReports = `
                create table new_reports (
                    org_id            integer not null,
                    account_number    integer not null,
                    cluster           character(36) not null,
                    report            varchar not null,
                    updated_at        timestamp not null,
                    kafka_offset      bigint not null default 0,
                
                    PRIMARY KEY (org_id, cluster, updated_at)
                );
`

	createIndexKafkaOffset = "CREATE INDEX report_kafka_offset_btree_idx ON new_reports (kafka_offset)"
)

// SQL statements
const (
	InsertNewReportStatement = `
		INSERT INTO new_reports(org_id, account_number, cluster, report, updated_at)
		VALUES ($1, $2, $3, $4, $5);
	`
)

// Messages
const (
	SQLStatementMessage = "SQL statement"
	StatementMessage    = "Statement"
)

// Storage represents an interface to almost any database or storage system
type Storage interface {
	Close() error
	WriteReportForCluster(
		orgID OrgID,
		accountNumber AccountNumber,
		clusterName ClusterName,
		report ClusterReport,
		collectedAtTime time.Time,
	) error
	DatabaseInitialization() error
	DatabaseCleanup() error
	DatabaseDropTables() error
}

// DBStorage is an implementation of Storage interface that use selected SQL like database
// like SQLite, PostgreSQL, MariaDB, RDS etc. That implementation is based on the standard
// sql package. It is possible to configure connection via Configuration structure.
// SQLQueriesLog is log for sql queries, default is nil which means nothing is logged
type DBStorage struct {
	connection   *sql.DB
	dbDriverType DBDriver
}

// ErrOldReport is an error returned if a more recent already
// exists on the storage while attempting to write a report for a cluster.
var ErrOldReport = errors.New("More recent report already exists in storage")

// tableNames contains names of all tables in the database.
var tableNames []string

// initStatements contains all statements used to initialize database
var initStatements []string

// NewStorage function creates and initializes a new instance of Storage interface
func NewStorage(configuration StorageConfiguration) (*DBStorage, error) {
	log.Info().Msg("Initializing connection to storage")

	driverType, driverName, dataSource, err := initAndGetDriver(configuration)
	if err != nil {
		log.Error().Err(err).Msg("Unsupported driver")
		return nil, err
	}

	log.Info().
		Str("driver", driverName).
		Str("datasource", dataSource).
		Msg("Making connection to data storage")

	// prepare connection
	connection, err := sql.Open(driverName, dataSource)
	if err != nil {
		log.Error().Err(err).Msg("Can not connect to data storage")
		return nil, err
	}

	// lazy initialization (TODO: use init function instead?)
	tableNames = []string{
		"new_reports",
		"reported",
		"notification_types",
		"states",
	}

	// lazy initialization (TODO: use init function instead?)
	initStatements = []string{
		createTableNotificationTypes,
		createTableStates,
		createTableReported,
		createTableNewReports,
		createIndexKafkaOffset,
	}

	log.Info().Msg("Connection to storage established")
	return NewFromConnection(connection, driverType), nil
}

// NewFromConnection function creates and initializes a new instance of Storage interface from prepared connection
func NewFromConnection(connection *sql.DB, dbDriverType DBDriver) *DBStorage {
	return &DBStorage{
		connection:   connection,
		dbDriverType: dbDriverType,
	}
}

// initAndGetDriver initializes driver(with logs if logSQLQueries is true),
// checks if it's supported and returns driver type, driver name, dataSource and error
func initAndGetDriver(configuration StorageConfiguration) (driverType DBDriver, driverName string, dataSource string, err error) {
	//var driver sql_driver.Driver
	driverName = configuration.Driver

	switch driverName {
	case "sqlite3":
		driverType = DBDriverSQLite3
		//driver = &sqlite3.SQLiteDriver{}
		// dataSource = configuration.SQLiteDataSource
	case "postgres":
		driverType = DBDriverPostgres
		//driver = &pq.Driver{}
		dataSource = fmt.Sprintf(
			"postgresql://%v:%v@%v:%v/%v?%v",
			configuration.PGUsername,
			configuration.PGPassword,
			configuration.PGHost,
			configuration.PGPort,
			configuration.PGDBName,
			configuration.PGParams,
		)
	default:
		err = fmt.Errorf("driver %v is not supported", driverName)
		return
	}

	return
}

// Close method closes the connection to database. Needs to be called at the end of application lifecycle.
func (storage DBStorage) Close() error {
	log.Info().Msg("Closing connection to data storage")
	if storage.connection != nil {
		err := storage.connection.Close()
		if err != nil {
			log.Error().Err(err).Msg("Can not close connection to data storage")
			return err
		}
	}
	return nil
}

// WriteReportForCluster writes result (health status) for selected cluster for given organization
func (storage DBStorage) WriteReportForCluster(
	orgID OrgID,
	accountNumber AccountNumber,
	clusterName ClusterName,
	report ClusterReport,
	lastCheckedTime time.Time,
) error {
	if storage.dbDriverType != DBDriverSQLite3 && storage.dbDriverType != DBDriverPostgres {
		return fmt.Errorf("Writing report with DB %v is not supported", storage.dbDriverType)
	}

	// Begin a new transaction.
	tx, err := storage.connection.Begin()
	if err != nil {
		return err
	}

	err = func(tx *sql.Tx) error {
		err = storage.insertReport(tx, orgID, accountNumber, clusterName, report, lastCheckedTime)
		if err != nil {
			return err
		}

		return nil
	}(tx)

	finishTransaction(tx, err)

	return err
}

func (storage DBStorage) insertReport(
	tx *sql.Tx,
	orgID OrgID,
	accountNumber AccountNumber,
	clusterName ClusterName,
	report ClusterReport,
	lastCheckedTime time.Time,
) error {
	_, err := tx.Exec(InsertNewReportStatement, orgID, accountNumber, clusterName, report, lastCheckedTime)
	if err != nil {
		log.Err(err).
			Int("org", int(orgID)).
			Int("account", int(accountNumber)).
			Str("cluster", string(clusterName)).
			Str("last checked", lastCheckedTime.String()).
			Msg("Unable to insert the cluster report")
		return err
	}

	return nil
}

// finishTransaction finishes the transaction depending on err. err == nil -> commit, err != nil -> rollback
func finishTransaction(tx *sql.Tx, err error) {
	if err != nil {
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			log.Err(rollbackError).Msg("Error when trying to rollback a transaction")
		}
	} else {
		commitError := tx.Commit()
		if commitError != nil {
			log.Err(commitError).Msg("Error when trying to commit a transaction")
		}
	}
}

func closeRows(rows *sql.Rows) {
	_ = rows.Close()
}

func tablesRelatedOperation(storage DBStorage, cmd func(string) string) error {
	// Begin a new transaction.
	tx, err := storage.connection.Begin()
	if err != nil {
		return err
	}

	err = func(tx *sql.Tx) error {

		// perform some operation to all tables
		for _, tableName := range tableNames {
			// it is not possible to use parameter for table name or a key
			// disable "G202 (CWE-89): SQL string concatenation (Confidence: HIGH, Severity: MEDIUM)"
			// #nosec G202
			sqlStatement := cmd(tableName)
			log.Info().Str(StatementMessage, sqlStatement).Msg(SQLStatementMessage)
			// println(sqlStatement)

			// perform the SQL statement in transaction
			_, err := tx.Exec(sqlStatement)
			if err != nil {
				return err
			}
		}

		return nil
	}(tx)

	finishTransaction(tx, err)

	return err
}

func deleteFromTableStatement(tableName string) string {
	return "DELETE FROM " + tableName + ";"
}

// DatabaseCleanup method performs database cleanup - deletes content of all
// tables in database.
func (storage DBStorage) DatabaseCleanup() error {
	err := tablesRelatedOperation(storage, deleteFromTableStatement)
	return err
}

func dropTableStatement(tableName string) string {
	return "DROP TABLE " + tableName + ";"
}

// DatabaseDropTables method performs database drop for all tables in database.
func (storage DBStorage) DatabaseDropTables() error {
	err := tablesRelatedOperation(storage, dropTableStatement)
	return err
}

// DatabaseInitialization method performs database initialization - creates all
// tables in database.
func (storage DBStorage) DatabaseInitialization() error {
	// Begin a new transaction.
	tx, err := storage.connection.Begin()
	if err != nil {
		return err
	}

	err = func(tx *sql.Tx) error {

		// databaze initialization
		for _, sqlStatement := range initStatements {
			log.Info().Str(StatementMessage, sqlStatement).Msg(SQLStatementMessage)
			// println(sqlStatement)

			// perform the SQL statement in transaction
			_, err := tx.Exec(sqlStatement)
			if err != nil {
				return err
			}
		}

		return nil
	}(tx)

	finishTransaction(tx, err)

	return err
}
