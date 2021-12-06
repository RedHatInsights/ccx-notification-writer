/*
Copyright © 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// This source file contains an implementation of interface between Go code and
// (almost any) SQL database like PostgreSQL, SQLite, or MariaDB.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/storage.html

// It is possible to configure connection to selected database by using
// StorageConfiguration structure. Currently that structure contains two
// configurable parameter:
//
// Driver - a SQL driver, like "sqlite3", "pq" etc.
// DataSource - specification of data source. The content of this parameter depends on the database used.

import (
	"errors"
	"fmt"
	"math"
	"time"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"
)

// Table creation-related scripts, queries, index creation etc.
const (
	// This table contains information about DB schema version and about
	// migration status.
	createTableMigrationInfo = `
                CREATE TABLE IF NOT EXISTS migration_info (
                    version     integer not null
                );
`

	// This table contains list of all notification types used by
	// Notification service. Frequency can be specified as in `crontab` -
	// https://crontab.guru/
	createTableNotificationTypes = `
                CREATE TABLE notification_types (
                    id          integer not null,
                    value       varchar not null,
                    frequency   varchar not null,
                    comment     varchar,
                
                    PRIMARY KEY (id)
                );
`

	// This table contains states for each row stored in `reported` table.
	// User can be notified about the report, report can be skipped if the
	// same as previous, skipped because of lower pripority, or can be in
	// error state.
	createTableStates = `
                CREATE TABLE states (
                    id          integer not null,
                    value       varchar not null,
                    comment     varchar,
                
                    PRIMARY KEY (id)
                );
`

	// Information of notifications reported to user or skipped due to some
	// conditions.
	createTableReported = `
                CREATE TABLE reported (
                    org_id            integer not null,
                    account_number    integer not null,
                    cluster           character(36) not null,
                    notification_type integer not null,
                    state             integer not null,
                    report            varchar not null,
                    updated_at        timestamp not null,
                    notified_at       timestamp not null,
		    error_log         varchar,
                
                    PRIMARY KEY (org_id, cluster, notified_at),
                    CONSTRAINT fk_notification_type
                        foreign key(notification_type)
                        references notification_types(id),
                    CONSTRAINT fk_state
                        foreign key (state)
                        references states(id)
                );
`

	// This table contains new reports consumed from Kafka topic and stored
	// to database in shrunk format (some attributes are removed).
	createTableNewReports = `
                CREATE TABLE new_reports (
                    org_id            integer not null,
                    account_number    integer not null,
                    cluster           character(36) not null,
                    report            varchar not null,
                    updated_at        timestamp not null,
                    kafka_offset      bigint not null default 0,
                
                    PRIMARY KEY (org_id, cluster, updated_at)
                );
`

	// Index for the new_reports table
	createIndexKafkaOffset = `
                CREATE INDEX report_kafka_offset_btree_idx
		       ON new_reports (kafka_offset)
`

	// Index for the new_reports table
	createIndexNewReportsOrgID = `
                CREATE INDEX new_reports_org_id_idx
		    ON new_reports
		 USING btree (org_id);
`

	// Index for the new_reports table
	createIndexNewReportsCluster = `
                CREATE INDEX new_reports_cluster_idx
		    ON new_reports
		 USING btree (cluster);
`
	// Index for the new_reports table
	createIndexNewReportsUpdatedAtAsc = `
                CREATE INDEX new_reports_updated_at_asc_idx
		       ON new_reports USING btree (updated_at ASC);
`

	// Index for the new_reports table
	createIndexNewReportsUpdatedAtDesc = `
                CREATE INDEX new_reports_updated_at_desc_idx
		       ON new_reports USING btree (updated_at DESC);
`

	// Index for the reported table
	createIndexReportedNotifiedAtDesc = `
                CREATE INDEX notified_at_desc_idx
		    ON reported
		 USING btree (notified_at DESC);
`

	// Index for the reported table
	createIndexReportedUpdatedAtAsc = `
                CREATE INDEX updated_at_desc_idx
		    ON reported
		 USING btree (updated_at ASC);
`

	// Index for the notification_types table
	createIndexNotificationTypesID = `
                CREATE INDEX notification_types_id_idx
		       ON notification_types USING btree (id ASC);
`

	// Get 0 if DB version is not inserted, 1 instead
	isDatabaseVersionExist = `SELECT count(*) FROM migration_info;`

	// Retrieve DB version
	getDatabaseVersion = `SELECT version FROM migration_info LIMIT 1;`

	// Display older records from new_reports table
	displayOldRecordsFromNewReportsTable = `
                SELECT org_id, account_number, cluster, updated_at, kafka_offset
		  FROM new_reports
		 WHERE updated_at < NOW() - $1::INTERVAL
		 ORDER BY updated_at
`

	// Delete older records from new_reports table
	deleteOldRecordsFromNewReportsTable = `
                DELETE
		  FROM new_reports
		 WHERE updated_at < NOW() - $1::INTERVAL
`

	// Display older records from reported table
	displayOldRecordsFromReportedTable = `
                SELECT org_id, account_number, cluster, updated_at, 0
		  FROM reported
		 WHERE updated_at < NOW() - $1::INTERVAL
		 ORDER BY updated_at
`

	// Delete older records from reported table
	deleteOldRecordsFromReportedTable = `
                DELETE
		  FROM reported
		 WHERE updated_at < NOW() - $1::INTERVAL
`
	// Value to be stored in migration_info table
	insertMigrationVersion = `
                INSERT INTO migration_info (version)
		            VALUES (1);
`
	// Value to be stored in notification_types table
	insertInstantReport = `
                INSERT INTO notification_types (id, value, frequency, comment)
		            VALUES (1, 'instant', '* * * * * *', 'instant notifications performed ASAP');
`
	// Value to be stored in notification_types table
	insertWeeklySummary = `
                INSERT INTO notification_types (id, value, frequency, comment)
		            VALUES (2, 'weekly', '@weekly', 'weekly summary');
`
	// Value to be stored in states table
	insertSentState = `
                INSERT INTO states (id, value, comment)
		            VALUES (1, 'sent', 'notification has been sent to user');
`
	// Value to be stored in states table
	insertSentSame = `
                INSERT INTO states (id, value, comment)
		            VALUES (2, 'same', 'skipped, report is the same as previous one');
`

	// Value to be stored in states table
	insertSentLowPriority = `
                INSERT INTO states (id, value, comment)
		            VALUES (3, 'lower', 'skipped, all issues has low priority');
`

	// Value to be stored in states table
	insertSentError = `
                INSERT INTO states (id, value, comment)
		            VALUES (4, 'error', 'notification delivery error');
`
)

// SQL statements
const (
	InsertNewReportStatement = `
		INSERT INTO new_reports(org_id, account_number, cluster, report, updated_at, kafka_offset)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
)

// Messages
const (
	SQLStatementMessage       = "SQL statement"
	StatementMessage          = "Statement"
	OrgIDMessage              = "Organization ID"
	AccountNumberMessage      = "Account number"
	ClusterNameMessage        = "Cluster name"
	UpdatedAtMessage          = "Updated at"
	UnableToCloseDBRowsHandle = "Unable to close the DB rows handle"
	AgeMessage                = "Age"
	MaxAgeAttribute           = "max age"
	VersionMessage            = "Retrieve database version"
)

// Other constants
const (
	DatabaseVersion = 1
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
		kafkaOffset KafkaOffset,
	) error
	DatabaseInitialization() error
	DatabaseCleanup() error
	DatabaseDropTables() error
	DatabaseDropIndexes() error
	DatabaseInitMigration() error
	GetLatestKafkaOffset() (KafkaOffset, error)
	PrintNewReportsForCleanup(maxAge string) error
	CleanupNewReports(maxAge string) (int, error)
	PrintOldReportsForCleanup(maxAge string) error
	CleanupOldReports(maxAge string) (int, error)
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

// indexNames contains names of all indexes in the database.
var indexNames []string

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
		"migration_info",
		"new_reports",
		"reported",
		"notification_types",
		"states",
	}

	// lazy initialization (TODO: use init function instead?)
	indexNames = []string{
		"report_kafka_offset_btree_idx",
	}

	// lazy initialization (TODO: use init function instead?)
	initStatements = []string{
		// tables
		createTableNotificationTypes,
		createTableStates,
		createTableReported,
		createTableNewReports,

		// indexes
		createIndexKafkaOffset,
		createIndexNewReportsOrgID,
		createIndexNewReportsCluster,
		createIndexNewReportsUpdatedAtAsc,
		createIndexNewReportsUpdatedAtDesc,
		createIndexReportedNotifiedAtDesc,
		createIndexReportedUpdatedAtAsc,
		createIndexNotificationTypesID,

		// records
		insertMigrationVersion,
		insertInstantReport,
		insertWeeklySummary,
		insertSentState,
		insertSentSame,
		insertSentLowPriority,
		insertSentError,
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
	// var driver sql_driver.Driver
	driverName = configuration.Driver

	switch driverName {
	case "sqlite3":
		driverType = DBDriverSQLite3
	case "postgres":
		driverType = DBDriverPostgres
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
	kafkaOffset KafkaOffset,
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
		err = storage.insertReport(tx, orgID, accountNumber, clusterName, report, lastCheckedTime, kafkaOffset)
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
	kafkaOffset KafkaOffset,
) error {
	_, err := tx.Exec(InsertNewReportStatement, orgID, accountNumber, clusterName, report, lastCheckedTime, kafkaOffset)
	if err != nil {
		log.Err(err).
			Int("org", int(orgID)).
			Int("account", int(accountNumber)).
			Int64("kafka offset", int64(kafkaOffset)).
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
	// it is not possible to use parameter for table name or a key
	// disable "G202 (CWE-89): SQL string concatenation (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G202
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

func dropIndexStatement(indexName string) string {
	return "DROP INDEX IF EXISTS " + indexName + ";"
}

// DatabaseDropIndexes method performs database drop for all tables in database.
func (storage DBStorage) DatabaseDropIndexes() error {
	err := tablesRelatedOperation(storage, dropIndexStatement)
	return err
}

// GetDatabaseVersionInfo method tries to retrieve database version from
// migration table.
func (storage DBStorage) GetDatabaseVersionInfo() (int, error) {
	// check if version info is stored in the database
	var count int
	err := storage.connection.QueryRow(isDatabaseVersionExist).Scan(&count)
	if err != nil {
		return -1, err
	}

	// table exists, but does not containing DB version
	if count == 0 {
		return -1, nil
	}

	// process version info
	var version int
	err = storage.connection.QueryRow(getDatabaseVersion).Scan(&version)
	if err != nil {
		return -1, err
	}

	return version, nil
}

// DatabaseInitMigration method initializes migration_info table
func (storage DBStorage) DatabaseInitMigration() error {
	// Begin a new transaction.
	tx, err := storage.connection.Begin()
	if err != nil {
		return err
	}

	err = func(tx *sql.Tx) error {
		// try to retrieve database version info
		version, err := storage.GetDatabaseVersionInfo()

		// it is possible (and expected) that version can not be read
		if err != nil {
			// just log the error - it is expected
			log.Info().Str(VersionMessage, createTableMigrationInfo).Msg(SQLStatementMessage)
		} else {
			// migration table already exists and contains the right version
			if version == DatabaseVersion {
				return nil
			}
		}

		// migration_info table initialization
		log.Info().Str(StatementMessage, createTableMigrationInfo).Msg(SQLStatementMessage)

		// perform the SQL statement in transaction
		_, err = tx.Exec(createTableMigrationInfo)
		if err != nil {
			return err
		}
		return nil
	}(tx)

	finishTransaction(tx, err)

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
		// try to retrieve database version info
		version, err := storage.GetDatabaseVersionInfo()
		if err != nil {
			log.Error().Err(err).Msg("DB version can not be retrieved")
			log.Error().Msg("Try to run CCX Notification service with --db-init-migration")
			return err
		}

		// skip rest of DB initialization, if already initialized
		if version == DatabaseVersion {
			log.Info().Msg("Database is already initialized")
			return nil
		}

		// databaze initialization
		for _, sqlStatement := range initStatements {
			log.Info().Str(StatementMessage, sqlStatement).Msg(SQLStatementMessage)

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

// GetLatestKafkaOffset returns latest kafka offset from report table
func (storage DBStorage) GetLatestKafkaOffset() (KafkaOffset, error) {
	var offset KafkaOffset
	err := storage.connection.QueryRow("SELECT COALESCE(MAX(kafka_offset), 0) FROM new_reports;").Scan(&offset)
	return offset, err
}

// PrintNewReports method prints all reports from selected table older than
// specified relative time
func (storage DBStorage) PrintNewReports(maxAge, query, tableName string) error {
	log.Info().
		Str(MaxAgeAttribute, maxAge).
		Str("select statement", query).
		Msg("PrintReportsForCleanup operation")

	rows, err := storage.connection.Query(query, maxAge)
	if err != nil {
		return err
	}
	// used to compute a real record age
	now := time.Now()

	// iterate over all old records
	for rows.Next() {
		var (
			orgID         int
			accountNumber int
			clusterName   string
			updatedAt     time.Time
			kafkaOffset   int64
		)

		// read one old record from the report table
		if err := rows.Scan(&orgID, &accountNumber, &clusterName, &updatedAt, &kafkaOffset); err != nil {
			// close the result set in case of any error
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg(UnableToCloseDBRowsHandle)
			}
			return err
		}

		// compute the real record age
		age := int(math.Ceil(now.Sub(updatedAt).Hours() / 24)) // in days

		// prepare for the report
		updatedAtF := updatedAt.Format(time.RFC3339)

		// just print the report
		log.Info().
			Int(OrgIDMessage, orgID).
			Int(AccountNumberMessage, accountNumber).
			Str(ClusterNameMessage, clusterName).
			Str(UpdatedAtMessage, updatedAtF).
			Int(AgeMessage, age).
			Msg("Old report from `" + tableName + "` table")
	}
	return nil
}

// PrintNewReportsForCleanup method prints all reports from `new_reports` table
// older than specified relative time
func (storage DBStorage) PrintNewReportsForCleanup(maxAge string) error {
	return storage.PrintNewReports(maxAge, displayOldRecordsFromNewReportsTable, "new_reports")
}

// PrintOldReportsForCleanup method prints all reports from `reported` table
// older than specified relative time
func (storage DBStorage) PrintOldReportsForCleanup(maxAge string) error {
	return storage.PrintNewReports(maxAge, displayOldRecordsFromReportedTable, "reported")
}

// Cleanup method deletes all reports older than specified
// relative time
func (storage DBStorage) Cleanup(maxAge, statement string) (int, error) {
	log.Info().
		Str(MaxAgeAttribute, maxAge).
		Str("delete statement", statement).
		Msg("Cleanup operation")

	// perform the SQL statement
	result, err := storage.connection.Exec(statement, maxAge)
	if err != nil {
		return 0, err
	}

	// read number of affected (deleted) rows
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// CleanupNewReports method deletes all reports from `new_reports` table older
// than specified relative time
func (storage DBStorage) CleanupNewReports(maxAge string) (int, error) {
	return storage.Cleanup(maxAge, deleteOldRecordsFromNewReportsTable)
}

// CleanupOldReports method deletes all reports from `reported` table older
// than specified relative time
func (storage DBStorage) CleanupOldReports(maxAge string) (int, error) {
	return storage.Cleanup(maxAge, deleteOldRecordsFromReportedTable)
}
