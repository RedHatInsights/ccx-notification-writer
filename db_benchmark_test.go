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
// Additional info about benchmarks can be found at
// https://redhatinsights.github.io/ccx-notification-writer/benchmarks.html
//
// Benchmarks use reports represented as JSON files that are stored in
// tests/reports/ subdirectory.
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/db_benchmark_test.html

import (
	"fmt"
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

// JSON files containing reports
const (
	reportWithAnalysisMetadataOnlyInJSON = "analysis_metadata_only.json"
	smallReportInJSON                    = "small_size.json"
	middleReportInJSON                   = "middle_size.json"
	largeReportInJSON                    = "large_size.json"
)

// SQL statements
//
// Table reported V1 is reported table without event_type column and constraint for this column
// Table reported V2 is reported table with event_type column and constraint for this column
const (
	noOpStatement = ``

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
	createIndexReportedEventTypeBtreeV2 = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING btree (event_type_id ASC);
        `

	// Optional index for the reported table for event_type_id
	// column. Index type is set to hash.
	// (https://www.postgresql.org/docs/current/indexes-types.html for more info)
	createIndexReportedEventTypeHashV2 = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING hash (event_type_id);
        `

	// Optional index for the reported table
	// column. Index type is set to BRIN.
	// (https://www.postgresql.org/docs/current/indexes-types.html for more info)
	createIndexReportedEventTypeBrinV2 = `
                CREATE INDEX IF NOT EXISTS event_type_idx
		    ON reported_benchmark_2
		 USING brin (event_type_id);
        `

	// Insert one record into reported table w/o event_type_id column
	insertIntoReportedV1Statement = `
            INSERT INTO reported_benchmark_1
            (org_id, account_number, cluster, notification_type, state, report, updated_at, notified_at, error_log)
            VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Insert one record into reported table with event_type_id column
	insertIntoReportedV2Statement = `
            INSERT INTO reported_benchmark_2
            (org_id, account_number, cluster, notification_type, state, report, updated_at, notified_at, error_log, event_type_id)
            VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// SQL query used to display older records from reported table
	displayOldRecordsFromReportedTableV1 = `
                SELECT org_id, account_number, cluster, updated_at, 0
                  FROM reported_benchmark_1
                 WHERE updated_at < NOW() - $1::INTERVAL
                 ORDER BY updated_at
        `

	// SQL query used to display older records from reported table
	displayOldRecordsFromReportedTableV2 = `
                SELECT org_id, account_number, cluster, updated_at, 0
                  FROM reported_benchmark_2
                 WHERE updated_at < NOW() - $1::INTERVAL
                 ORDER BY updated_at
        `

	// SQL query used to display older records from reported table
	deleteOldRecordsFromReportedTableV1 = `
                DELETE
                  FROM reported_benchmark_1
                 WHERE updated_at < NOW() - $1::INTERVAL
        `

	// SQL query used to display older records from reported table
	deleteOldRecordsFromReportedTableV2 = `
                DELETE
                  FROM reported_benchmark_2
                 WHERE updated_at < NOW() - $1::INTERVAL
        `

	// vacuuming statements for both tables used by benchmarks
	vacuumTableReportedV1Statement = `vacuum reported_benchmark_1`
	vacuumTableReportedV2Statement = `vacuum reported_benchmark_2`
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

	// connection should be established at this moment
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

	// next time, benchmarks will use already established connection to
	// database
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

	// perform insert
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

	// perform insert
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

// runBenchmarkSelectOrDeleteFromReportedTable function perform several inserts
// into reported table v1 or v2 (depending on injected function) and the run
// queries or delete statements against such table. In case of any error
// detected, benchmarks fail immediatelly.
func runBenchmarkSelectOrDeleteFromReportedTable(b *testing.B, insertFunction insertIntoReportedFunc,
	selectStatement string, deleteStatement string,
	initStatements []string, reportsCount int, report *string) {
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

	// fill-in the table (no part of benchmark, so don't measure time there)
	for i := 0; i < reportsCount; i++ {
		insertFunction(b, connection, i, report)
	}

	// time to start benchmark
	b.ResetTimer()

	// perform DB benchmark
	for i := 0; i < b.N; i++ {
		// perform benchmarks for SELECT statement (query)
		if selectStatement != noOpStatement {
			rows, err := connection.Query(selectStatement, "8 days")
			if err != nil {
				b.Fatal(err)
			}
			if rows == nil {
				b.Fatal("no rows selected")
			}

			err = rows.Close()
			if err != nil {
				b.Fatal(err)
			}
		}
		// perform benchmarks for DELETE statement
		if deleteStatement != noOpStatement {
			_, err := connection.Exec(deleteStatement, "8 days")
			if err != nil {
				b.Fatal(err)
			}
			// fill-in the table again (no part of benchmark, so don't measure time there)
			b.StopTimer()
			for i := 0; i < reportsCount; i++ {
				insertFunction(b, connection, i, report)
			}
			b.StartTimer()
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

// getIndicesForReportedTableV1 is helper function to return map with all
// possible indices combinations
func getIndicesForReportedTableV1() map[string][]string {
	indices := make(map[string][]string)

	indices["no indices]"] = []string{}

	indices["notified_at_desc"] = []string{
		createIndexReportedNotifiedAtDescV1}

	indices["updated_at_asc"] = []string{
		createIndexReportedUpdatedAtAscV1}

	indices["notified_at_desc and updated_at_asc"] = []string{
		createIndexReportedNotifiedAtDescV1,
		createIndexReportedUpdatedAtAscV1}

	return indices
}

// getIndicesForReportedTableV2 is helper function to return map with all
// possible indices combinations
func getIndicesForReportedTableV2() map[string][]string {
	indices := make(map[string][]string)

	indices["no indices]"] = []string{}

	indices["notified_at_desc"] = []string{
		createIndexReportedNotifiedAtDescV2}

	indices["updated_at_asc"] = []string{
		createIndexReportedUpdatedAtAscV2}

	indices["notified_at_desc and updated_at_asc"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedUpdatedAtAscV2}

	indices["notified_at_desc and event_type(BTree)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedEventTypeBtreeV2}

	indices["notified_at_desc and event_type(hash)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedEventTypeHashV2}

	indices["notified_at_desc and event_type(BRIN)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedEventTypeBrinV2}

	indices["updated_at_asc and event_type(BTree)"] = []string{
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeBtreeV2}

	indices["updated_at_asc and event_type(hash)"] = []string{
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeHashV2}

	indices["updated_at_asc and event_type(BRIN)"] = []string{
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeBrinV2}

	indices["notified_at_desc and updated_at_asc and event_type(BTree)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeBtreeV2}

	indices["notified_at_desc and updated_at_asc and event_type(hash)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeHashV2}

	indices["notified_at_desc and updated_at_asc and event_type(BRIN)"] = []string{
		createIndexReportedNotifiedAtDescV2,
		createIndexReportedUpdatedAtAscV2,
		createIndexReportedEventTypeBrinV2}
	return indices
}

// benchmarkInsertReportsIntoReportedTableImpl is an implementation of
// benchmark to insert reports into reported table with or without
// event_type_id column. Table with all possible indices combination is tested.
func benchmarkInsertReportsIntoReportedTableImpl(
	b *testing.B,
	insertFunction insertIntoReportedFunc,
	initStatements []string,
	indices map[string][]string,
	report string) {

	// try all indices combinations
	for description, indexStatements := range indices {
		// new benchmark
		b.Run(description, func(b *testing.B) {
			// prepare all SQL statements to be run before benchmark
			sqlStatements := make([]string, len(initStatements)+len(indexStatements))

			// add all init statements
			sqlStatements = append(sqlStatements, initStatements...)

			// add all statements to create indices
			sqlStatements = append(sqlStatements, indexStatements...)

			// now everything's ready -> run benchmark
			runBenchmarkInsertIntoReportedTable(b, insertFunction, sqlStatements, 1, &report)
		})
	}
}

// benchmarkSelectOrDeleteOldReportsFromReportedTableImpl is an implementation
// of benchmark to query reports from reported or delete reports from such
// table with or without event_type_id column. Table with all possible indices
// combination is tested.
func benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(
	b *testing.B,
	insertFunction insertIntoReportedFunc,
	selectStatement string,
	deleteStatement string,
	initStatements []string,
	indices map[string][]string,
	report string) {

	// number of reports to be insterted into the table
	var possibleReportsCount []int = []int{1, 100, 1000}

	// try all indices combinations
	for description, indexStatements := range indices {
		// benchmark with various reports count stored in table
		for _, reportsCount := range possibleReportsCount {
			// new benchmark
			benchmarkName := fmt.Sprintf("%s with %d reports", description, reportsCount)
			b.Run(benchmarkName, func(b *testing.B) {
				// prepare all SQL statements to be run before benchmark
				sqlStatements := make([]string, len(initStatements)+len(indexStatements))

				// add all init statements
				sqlStatements = append(sqlStatements, initStatements...)

				// add all statements to create indices
				sqlStatements = append(sqlStatements, indexStatements...)

				// now everything's ready -> run benchmark
				runBenchmarkSelectOrDeleteFromReportedTable(b, insertFunction,
					selectStatement, deleteStatement,
					sqlStatements, reportsCount, &report)
			})
		}
	}
}

// BenchmarkInsertReportsIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertReportsIntoReportedTableV1(b *testing.B) {
	// try to insert empty reports
	report := ""

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV1,
		initStatements, indices, report)
}

// BenchmarkInsertReportsIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertReportsIntoReportedTableV2(b *testing.B) {
	// try to insert empty reports
	report := ""

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table with event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV2,
		initStatements, indices, report)
}

// BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV1(b *testing.B) {
	// report with size approx 0.5kB
	report := readReport(b, reportWithAnalysisMetadataOnlyInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV1,
		initStatements, indices, report)
}

// BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertAnalysisMetadataOnlyReportIntoReportedTableV2(b *testing.B) {
	// report with size approx 0.5kB
	report := readReport(b, reportWithAnalysisMetadataOnlyInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table with event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV2,
		initStatements, indices, report)
}

// BenchmarkInsertSmallReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertSmallReportIntoReportedTableV1(b *testing.B) {
	// report with size approx 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV1,
		initStatements, indices, report)
}

// BenchmarkInsertSmallReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertSmallReportIntoReportedTableV2(b *testing.B) {
	// report with size approx 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table with event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV2,
		initStatements, indices, report)
}

// BenchmarkInsertMiddleReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertMiddleReportIntoReportedTableV1(b *testing.B) {
	// report with size approx 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV1,
		initStatements, indices, report)
}

// BenchmarkInsertMiddleReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertMiddleReportIntoReportedTableV2(b *testing.B) {
	// report with size approx 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table with event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV2,
		initStatements, indices, report)
}

// BenchmarkInsertLargeReportIntoReportedTableV1 checks the speed of inserting
// into reported table without event_type column
func BenchmarkInsertLargeReportIntoReportedTableV1(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV1,
		initStatements, indices, report)
}

// BenchmarkInsertLargeReportIntoReportedTableV2 checks the speed of inserting
// into reported table with event_type column
func BenchmarkInsertLargeReportIntoReportedTableV2(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table with event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkInsertReportsIntoReportedTableImpl(b, insertIntoReportedV2,
		initStatements, indices, report)
}

// BenchmarkSelectOldEmptyRecordsFromReportedTableV1 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldEmptyRecordsFromReportedTableV1(b *testing.B) {
	// empty report
	report := ``

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, displayOldRecordsFromReportedTableV1, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldEmptyRecordsFromReportedTableV2 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldEmptyRecordsFromReportedTableV2(b *testing.B) {
	// empty report
	report := ``

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, displayOldRecordsFromReportedTableV2, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldSmallRecordsFromReportedTableV1 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldSmallRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, displayOldRecordsFromReportedTableV1, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldSmallRecordsFromReportedTableV2 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldSmallRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, displayOldRecordsFromReportedTableV2, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldMiddleRecordsFromReportedTableV1 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldMiddleRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, displayOldRecordsFromReportedTableV1, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldMiddleRecordsFromReportedTableV2 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldMiddleRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, displayOldRecordsFromReportedTableV2, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldLargeRecordsFromReportedTableV1 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldLargeRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, displayOldRecordsFromReportedTableV1, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkSelectOldLargeRecordsFromReportedTableV2 tests the query to reported
// table used inside CCX Notification Service
func BenchmarkSelectOldLargeRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table with event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, displayOldRecordsFromReportedTableV2, noOpStatement,
		initStatements, indices, report)
}

// BenchmarkDeleteOldEmptyRecordsFromReportedTableV1 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldEmptyRecordsFromReportedTableV1(b *testing.B) {
	// empty report
	report := ``

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, noOpStatement, deleteOldRecordsFromReportedTableV1,
		initStatements, indices, report)
}

// BenchmarkDeleteOldEmptyRecordsFromReportedTableV2 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldEmptyRecordsFromReportedTableV2(b *testing.B) {
	// empty report
	report := ``

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, noOpStatement, deleteOldRecordsFromReportedTableV2,
		initStatements, indices, report)
}

// BenchmarkDeleteOldSmallRecordsFromReportedTableV1 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldSmallRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, noOpStatement, deleteOldRecordsFromReportedTableV1,
		initStatements, indices, report)
}

// BenchmarkDeleteOldSmallRecordsFromReportedTableV2 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldSmallRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 2kB
	report := readReport(b, smallReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, noOpStatement, deleteOldRecordsFromReportedTableV2,
		initStatements, indices, report)
}

// BenchmarkDeleteOldMiddleRecordsFromReportedTableV1 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldMiddleRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, noOpStatement, deleteOldRecordsFromReportedTableV1,
		initStatements, indices, report)
}

// BenchmarkDeleteOldMiddleRecordsFromReportedTableV2 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldMiddleRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 4kB
	report := readReport(b, middleReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, noOpStatement, deleteOldRecordsFromReportedTableV2,
		initStatements, indices, report)
}

// BenchmarkDeleteOldLargeRecordsFromReportedTableV1 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldLargeRecordsFromReportedTableV1(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV1,
		createTableReportedV1,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV1()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV1, noOpStatement, deleteOldRecordsFromReportedTableV1,
		initStatements, indices, report)
}

// BenchmarkDeleteOldLargeRecordsFromReportedTableV2 tests the deletion statement
// from reported table used inside CCX Notification Service
func BenchmarkDeleteOldLargeRecordsFromReportedTableV2(b *testing.B) {
	// report with size over 8kB
	report := readReport(b, largeReportInJSON)

	// default init statements for reported table without event_type_id column
	initStatements := []string{
		dropTableReportedV2,
		createTableReportedV2,
	}

	// all possible indices combinatiors for reported table without event_type_id column
	indices := getIndicesForReportedTableV2()

	// run benchmarks with various combination of indices
	benchmarkSelectOrDeleteOldReportsFromReportedTableImpl(b,
		insertIntoReportedV2, noOpStatement, deleteOldRecordsFromReportedTableV2,
		initStatements, indices, report)
}
