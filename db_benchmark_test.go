/*
Copyright © 2022 Red Hat, Inc.

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

package main_test

// Benchmark for storage backend - the Postgres or RDS
//
// Please note that database schema needs to be prepared before benchmarks and
// needs to be migrated to latest version.
//
// Look into README.md for detailed information how to check DB status and how
// to migrate the database.
//
// Benchmarks use reports represented as JSON files that are stored in
// tests/reports/ subdirectory.
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/db_benchmark_test.html

import (
	"os"
	"testing"
	"time"

	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// Configuration-related constants
const (
	configFileEnvName = "CCX_NOTIFICATION_WRITER_CONFIG_FILE"
	configFileName    = "tests/benchmark"
	reportDirectory   = "tests/reports/"
)

// SQL statements
//
// Table reported V1 is reported table without event_type column and constraint for this column
// Table reported V2 is reported table with event_type column and constraint for this column
const (
	dropTableReportedV1 = `DROP TABLE IF EXISTS reported_benchmark_1;`
	dropTableReportedV2 = `DROP TABLE IF EXISTS reported_benchmark_2;`

	createTableReportedV1 = `
                CREATE TABLE IF NOT EXISTS reported_benchmark_1 (
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
                );`

	createTableReportedV2 = `
                CREATE TABLE IF NOT EXISTS reported_benchmark_2 (
                    org_id            integer not null,
                    account_number    integer not null,
                    cluster           character(36) not null,
                    notification_type integer not null,
                    state             integer not null,
                    report            varchar not null,
                    updated_at        timestamp not null,
                    notified_at       timestamp not null,
		    error_log         varchar,
		    event_type_id     integer,
                
                    PRIMARY KEY (org_id, cluster, notified_at),
                    CONSTRAINT fk_notification_type
                        foreign key(notification_type)
                        references notification_types(id),
                    CONSTRAINT fk_state
                        foreign key (state)
                        references states(id),
		    CONSTRAINT fk_event_type
		        foreign key (event_type_id)
			references event_targets(id)
                );`

	// Index for the reported table used in benchmarks for
	// notified_at column
	createIndexReportedNotifiedAtDescV1 = `
                CREATE INDEX IF NOT EXISTS notified_at_desc_idx
		    ON reported_benchmark_1
		 USING btree (notified_at DESC);
        `

	// Index for the reported table used in benchmarks for
	// notified_at column
	createIndexReportedNotifiedAtDescV2 = `
                CREATE INDEX IF NOT EXISTS notified_at_desc_idx
		    ON reported_benchmark_2
		 USING btree (notified_at DESC);
        `

	// Index for the reported table for updated_at_desc column
	createIndexReportedUpdatedAtAscV1 = `
                CREATE INDEX IF NOT EXISTS updated_at_desc_idx
		    ON reported_benchmark_1
		 USING btree (updated_at ASC);
        `

	// Index for the reported table for updated_at_desc column
	createIndexReportedUpdatedAtAscV2 = `
                CREATE INDEX IF NOT EXISTS updated_at_desc_idx
		    ON reported_benchmark_2
		 USING btree (updated_at ASC);
        `

	// Optional index for the reported table for event_type_id
	// column. Index type is set to BTree.
	// (https://www.postgresql.org/docs/current/indexes-types.html for more info)
	createIndexReportedV2EventTypeBtree = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING btree (event_type_id ASC);
        `

	// Optional index for the reported table for event_type_id
	// column. Index type is set to hash.
	// (https://www.postgresql.org/docs/current/indexes-types.html for more info)
	createIndexReportedV2EventTypeHash = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING hash (event_type_id ASC);
        `

	// Optional index for the reported table
	// column. Index type is set to BRIN.
	// (https://www.postgresql.org/docs/current/indexes-types.html for more info)
	createIndexReportedV2EventTypeBrin = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING brin (event_type_id ASC);
        `

	insertIntoReportedV1Statement = `
            INSERT INTO reported_benchmark_1
            (org_id, account_number, cluster, notification_type, state, report, updated_at, notified_at, error_log)
            VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	insertIntoReportedV2Statement = `
            INSERT INTO reported_benchmark_2
            (org_id, account_number, cluster, notification_type, state, report, updated_at, notified_at, error_log, event_type_id)
            VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
)

// insertIntoReportedFunc type represents any function to be called to insert
// records into reported table
type insertIntoReportedFunc func(b *testing.B, connection *sql.DB, i int, report *string)

// initLogging function initializes logging that's used internally by functions
// from github.com/RedHatInsights/ccx-notification-writer package
func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// loadConfiguration function loads configuration prepared to be used by
// benchmarks
func loadConfiguration() (main.ConfigStruct, error) {
	os.Clearenv()

	err := os.Setenv(configFileEnvName, configFileName)
	if err != nil {
		return main.ConfigStruct{}, err
	}

	config, err := main.LoadConfiguration(configFileEnvName, configFileName)
	if err != nil {
		return main.ConfigStruct{}, err
	}

	return config, nil
}

// initializeStorage function initializes storage and perform connection to
// database
func initializeStorage(config main.ConfigStruct) (*sql.DB, error) {
	storageConfiguration := main.GetStorageConfiguration(config)

	// perform storage initialization
	storage, err := main.NewStorage(storageConfiguration)
	if err != nil {
		return nil, err
	}

	// for benchmarks, we just need to have connection to DB, not the whole
	// Storage structure
	connection := storage.Connection()

	// check if connection has been established
	if storage.Connection() == nil {
		return nil, err
	}

	return connection, nil
}

// unfortunatelly there seems to be no other way how to handle per-benchmark
// setup properly than to use global variables
var connection *sql.DB
var initialized bool = false

// setup function performs all DB initializations needed for DB benchmarks. Can
// be called from any benchmark. In case of any error detected, benchmarks fail
// immediatelly.
func setup(b *testing.B) *sql.DB {
	if initialized {
		return connection
	}

	initLogging()

	config, err := loadConfiguration()
	if err != nil {
		b.Fatal(err)
	}

	connection, err = initializeStorage(config)
	if err != nil {
		b.Fatal(err)
	}

	initialized = true

	return connection
}

// execSQLStatement function executes any provided SQL statement. In case of
// any error detected, benchmarks fail immediatelly.
func execSQLStatement(b *testing.B, connection *sql.DB, statement string) {
	_, err := connection.Exec(statement)

	// check for any error, possible to exit immediatelly
	if err != nil {
		b.Fatal(err)
	}
}

// insertIntoReportedV1Statement function inserts one new record into reported
// table v1. In case of any error detected, benchmarks fail immediatelly.
func insertIntoReportedV1(b *testing.B, connection *sql.DB, i int, report *string) {
	// following columns needs to be updated with data:
	// 1 | org_id            | integer                     | not null  |
	// 2 | account_number    | integer                     | not null  |
	// 3 | cluster           | character(36)               | not null  |
	// 4 | notification_type | integer                     | not null  |
	// 5 | state             | integer                     | not null  |
	// 6 | report            | character varying           | not null  |
	// 7 | updated_at        | timestamp without time zone | not null  |
	// 8 | notified_at       | timestamp without time zone | not null  |
	// 9 | error_log         | character varying           |           |

	orgID := i % 1000                  // limited number of org IDs
	accountNumber := orgID + 1         // can be different than org ID
	clusterName := uuid.New().String() // unique
	notificationTypeID := 1            // instant report
	stateID := 1 + i%4                 // just four states can be used
	updatedAt := time.Now()            // don't have to be unique
	notifiedAt := time.Now()           // don't have to be unique
	errorLog := ""                     // usually empty

	_, err := connection.Exec(insertIntoReportedV1Statement, orgID,
		accountNumber, clusterName, notificationTypeID, stateID,
		report, updatedAt, notifiedAt, errorLog)

	// check for any error, possible to exit immediatelly
	if err != nil {
		b.Fatal(err)
	}
}

// insertIntoReportedV2Statement function inserts one new record into reported
// table v2. In case of any error detected, benchmarks fail immediatelly.
func insertIntoReportedV2(b *testing.B, connection *sql.DB, i int, report *string) {
	// following columns needs to be updated with data:
	// 1 | org_id            | integer                     | not null  |
	// 2 | account_number    | integer                     | not null  |
	// 3 | cluster           | character(36)               | not null  |
	// 4 | notification_type | integer                     | not null  |
	// 5 | state             | integer                     | not null  |
	// 6 | report            | character varying           | not null  |
	// 7 | updated_at        | timestamp without time zone | not null  |
	// 8 | notified_at       | timestamp without time zone | not null  |
	// 9 | error_log         | character varying           |           |
	// 10| event_type_id     | integer                     |           |

	orgID := i % 1000                  // limited number of org IDs
	accountNumber := orgID + 1         // can be different than org ID
	clusterName := uuid.New().String() // unique
	notificationTypeID := 1            // instant report
	stateID := 1 + i%4                 // just four states can be used
	updatedAt := time.Now()            // don't have to be unique
	notifiedAt := time.Now()           // don't have to be unique
	errorLog := ""                     // usually empty
	eventTypeID := i%2 + 1             // just two event type targets possible

	_, err := connection.Exec(insertIntoReportedV2Statement, orgID,
		accountNumber, clusterName, notificationTypeID, stateID,
		report, updatedAt, notifiedAt, errorLog, eventTypeID)

	// check for any error, possible to exit immediatelly
	if err != nil {
		b.Fatal(err)
	}
}

// runBenchmarkInsertIntoReportedTable function perform several inserts into
// reported table v1 or v2 (depending on injected function). In case of any
// error detected, benchmarks fail immediatelly.
func runBenchmarkInsertIntoReportedTable(b *testing.B, insertFunction insertIntoReportedFunc,
	initStatements []string, scaleFactor int, report *string) {
	// retrieve DB connection
	connection := setup(b)
	if connection == nil {
		b.Fatal()
	}

	// run all init SQL statements
	for _, sqlStatement := range initStatements {
		execSQLStatement(b, connection, sqlStatement)
	}

	// good citizens cleanup properly
	//defer execSQLStatement(b, connection, dropTableReportedV1)

	// time to start benchmark
	b.ResetTimer()

	// perform DB benchmark
	for i := 0; i < b.N; i++ {
		for j := 0; j < scaleFactor; j++ {
			insertFunction(b, connection, i, report)
		}
	}
}

// readReport function tries to read report from file. Benchmark will fail in
// any error.
func readReport(b *testing.B, filename string) string {
	content, err := os.ReadFile(reportDirectory + filename)
	if err != nil {
		b.Fatal(err)
	}
	return string(content)
}

// BenchmarkInsertEmptyReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertEmptyReportIntoReportedTableV1(b *testing.B) {
	report := ""

	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV1, initStatements, 1, &report)
}

// BenchmarkInsertEmptyReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertEmptyReportIntoReportedTableV2(b *testing.B) {
	report := ""

	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV2, initStatements, 1, &report)
}

// BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV1(b *testing.B) {
	report := readReport(b, "analysis_metadata_only.json")

	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV1, initStatements, 1, &report)
}

// BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV2(b *testing.B) {
	report := readReport(b, "analysis_metadata_only.json")

	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV2, initStatements, 1, &report)
}

// BenchmarkInsertSmallReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertSmallReportIntoReportedTableV1(b *testing.B) {
	report := readReport(b, "small_size.json")

	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV1, initStatements, 1, &report)
}

// BenchmarkInsertSmallReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertSmallReportIntoReportedTableV2(b *testing.B) {
	report := readReport(b, "small_size.json")

	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV2, initStatements, 1, &report)
}

// BenchmarkInsertMiddleReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertMiddleReportIntoReportedTableV1(b *testing.B) {
	report := readReport(b, "middle_size.json")

	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV1, initStatements, 1, &report)
}

// BenchmarkInsertMiddleReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertMiddleReportIntoReportedTableV2(b *testing.B) {
	report := readReport(b, "middle_size.json")

	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV2, initStatements, 1, &report)
}

// BenchmarkInsertLargeReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertLargeReportIntoReportedTableV1(b *testing.B) {
	report := readReport(b, "large_size.json")

	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV1, initStatements, 1, &report)
}

// BenchmarkInsertLargeReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertLargeReportIntoReportedTableV2(b *testing.B) {
	report := readReport(b, "large_size.json")

	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	runBenchmarkInsertIntoReportedTable(b, insertIntoReportedV2, initStatements, 1, &report)
}
