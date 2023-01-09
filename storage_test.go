/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

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

// Unit test definitions for functions and methods defined in source file
// storage.go
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/storage_test.html

import (
	"errors"
	"testing"
	"time"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// wrongDatabaseDriver is any integer value different from DBDriverSQLite3 and
// DBDriverPostgres
// (for selected DB operations)
const wrongDatabaseDriver = 10

// mustCreateMockConnection function tries to create a new mock connection and
// checks if the operation was finished without problems.
func mustCreateMockConnection(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	// try to initialize new mock connection
	connection, mock, err := sqlmock.New()

	// check the status
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return connection, mock
}

// checkConnectionClose function perform mocked DB closing operation and checks
// if the connection is properly closed from unit tests.
func checkConnectionClose(t *testing.T, connection *sql.DB) {
	// connection to mocked DB needs to be closed properly
	err := connection.Close()

	// check the error status
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}
}

// checkAllExpectations function checks if all database-related operations have
// been really met.
func checkAllExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	// check if all expectations were met
	err := mock.ExpectationsWereMet()

	// check the error status
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestGetLatestKafkaOffset function checks the method
// Storage.GetLatestKafkaOffset.
func TestGetLatestKafkaOffset(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"kafka_offset"})
	rows.AddRow(42)

	// expected query performed by tested function
	expectedQuery := "SELECT COALESCE\\(MAX\\(kafka_offset\\), 0\\) FROM new_reports;"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	offset, err := storage.GetLatestKafkaOffset()
	if err != nil {
		t.Errorf("error was not expected while getting latest Kafka offset: %s", err)
	}

	// check the org ID returned from tested function
	if offset != 42 {
		t.Errorf("wrong Kafka offset returned: %d", offset)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestGetLatestKafkaOffsetOnError function checks the method
// Storage.GetLatestKafkaOffset when error is returned.
func TestGetLatestKafkaOffsetOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"kafka_offset"})
	rows.AddRow(42)

	// expected query performed by tested function
	expectedQuery := "SELECT COALESCE\\(MAX\\(kafka_offset\\), 0\\) FROM new_reports;"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 0)

	// call the tested method
	_, err := storage.GetLatestKafkaOffset()
	if err == nil {
		t.Errorf("error was expected while getting latest Kafka offset: %s", err)
	}

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintNewReportsForCleanup function checks the method
// Storage.PrintNewReportsForCleanup.
func TestPrintNewReportsForCleanup(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id", "account_number", "cluster", "updated_at", "kafka_offset"})
	updatedAt := time.Now()
	rows.AddRow(1, 1000, "cluster_name", updatedAt, 42)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, kafka_offset FROM new_reports WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintNewReportsForCleanup("1 day")
	if err != nil {
		t.Errorf("error was not expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintNewReportsForCleanupOnScanError function checks the method
// Storage.PrintNewReportsForCleanup.
func TestPrintNewReportsForCleanupOnScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id", "account_number", "cluster", "updated_at", "kafka_offset"})
	updatedAt := time.Now()
	rows.AddRow(1, "this is not integer", "cluster_name", updatedAt, 42)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, kafka_offset FROM new_reports WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintNewReportsForCleanup("1 day")
	if err == nil {
		t.Errorf("error was expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintNewReportsForCleanupOnError function checks the method
// Storage.PrintNewReportsForCleanup.
func TestPrintNewReportsForCleanupOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, kafka_offset FROM new_reports WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintNewReportsForCleanup("1 day")
	if err == nil {
		t.Errorf("error was expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintOldReportsForCleanup function checks the method
// Storage.PrintOldReportsForCleanup.
func TestPrintOldReportsForCleanup(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id", "account_number", "cluster", "updated_at", "kafka_offset"})
	updatedAt := time.Now()
	rows.AddRow(1, 1000, "cluster_name", updatedAt, 42)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, 0 FROM reported WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintOldReportsForCleanup("1 day")
	if err != nil {
		t.Errorf("error was not expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintOldReportsForCleanupOnScanError function checks the method
// Storage.PrintOldReportsForCleanup.
func TestPrintOldReportsForCleanupOnScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id", "account_number", "cluster", "updated_at", "kafka_offset"})
	updatedAt := time.Now()
	rows.AddRow(1, "this is not integer", "cluster_name", updatedAt, 42)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, 0 FROM reported WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintOldReportsForCleanup("1 day")
	if err == nil {
		t.Errorf("error was expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestPrintOldReportsForCleanupOnError function checks the method
// Storage.PrintOldReportsForCleanup.
func TestPrintOldReportsForCleanupOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedQuery := "SELECT org_id, account_number, cluster, updated_at, 0 FROM reported WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY updated_at"

	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.PrintOldReportsForCleanup("1 day")
	if err == nil {
		t.Errorf("error was expected while printing old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestCleanupNewReports function checks the method Storage.CleanupNewReports.
func TestCleanupNewReports(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "DELETE FROM new_reports WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL"

	mock.ExpectExec(expectedStatement).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	deleted, err := storage.CleanupNewReports("1 day")
	if err != nil {
		t.Errorf("error was not expected while cleaning old reports: %s", err)
	}

	// check number of returned rows
	if deleted != 1 {
		t.Errorf("one row should be deleted, but %d rows were deleted", deleted)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestCleanupNewReportsOnError function checks the method
// Storage.CleanupNewReports.
func TestCleanupNewReportsOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "DELETE FROM new_reports WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL"

	mock.ExpectExec(expectedStatement).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	_, err := storage.CleanupNewReports("1 day")
	if err == nil {
		t.Errorf("error was not expected while cleaning old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestCleanupOldReports function checks the method Storage.CleanupOldReports.
func TestCleanupOldReports(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "DELETE FROM reported WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL"

	mock.ExpectExec(expectedStatement).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	deleted, err := storage.CleanupOldReports("1 day")
	if err != nil {
		t.Errorf("error was not expected while cleaning old reports: %s", err)
	}

	// check number of returned rows
	if deleted != 1 {
		t.Errorf("one row should be deleted, but %d rows were deleted", deleted)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestCleanupOldReportsOnError function checks the method
// Storage.CleanupNewReports.
func TestCleanupOldReportsOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "DELETE FROM reported WHERE updated_at < NOW\\(\\) - \\$1::INTERVAL"

	mock.ExpectExec(expectedStatement).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	_, err := storage.CleanupOldReports("1 day")
	if err == nil {
		t.Errorf("error was not expected while cleaning old reports: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestWriteReportForCluster function checks the method
// Storage.WriteReportForCluster.
func TestWriteReportForCluster(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "INSERT INTO new_reports\\(org_id, account_number, cluster, report, updated_at, kafka_offset\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\);"

	mock.ExpectBegin()
	mock.ExpectExec(expectedStatement).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.WriteReportForCluster(1, 2, "foo", "", time.Now(), 42)
	if err != nil {
		t.Errorf("error was not expected while writing report for cluster: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestWriteReportForClusterOnError function checks the method
// Storage.WriteReportForCluster.
func TestWriteReportForClusterOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	expectedStatement := "INSERT INTO new_reports\\(org_id, account_number, cluster, report, updated_at, kafka_offset\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\);"

	mock.ExpectBegin()
	mock.ExpectExec(expectedStatement).WillReturnError(mockedError)
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.WriteReportForCluster(1, 2, "foo", "", time.Now(), 42)
	if err == nil {
		t.Errorf("error was expected while writing report for cluster: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestWriteReportForClusterWrongDriver function checks the method
// Storage.WriteReportForCluster.
func TestWriteReportForClusterWrongDriver(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, wrongDatabaseDriver)

	// call the tested method
	err := storage.WriteReportForCluster(1, 2, "foo", "", time.Now(), 42)
	if err == nil {
		t.Errorf("error was expected while writing report for cluster")
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestGetDatabaseVersionInfo function checks the method
// Storage.getDatabaseVersionInfo, the happy path in this case.
func TestGetDatabaseVersionInfo(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected database version
	expectedVersion := 42

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	rowsCount.AddRow(1)

	// first expected SQL statement
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM migration_info;").WillReturnRows(rowsCount)

	// prepare mocked result for SQL query
	rowsVersion := sqlmock.NewRows([]string{"count"})
	rowsVersion.AddRow(expectedVersion)

	// second expected SQL statement
	mock.ExpectQuery("SELECT version FROM migration_info LIMIT 1;").WillReturnRows(rowsVersion)

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	version, err := storage.GetDatabaseVersionInfo()
	if err != nil {
		t.Errorf("error was not expected while initializing database: %s", err)
	}

	// check the returned version
	assert.Equal(t, expectedVersion, version)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestGetDatabaseVersionInfoNoVersion function checks the method
// Storage.getDatabaseVersionInfo when no version is stored in the
// database.
func TestGetDatabaseVersionInfoNoVersion(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected database version
	expectedVersion := -1

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	rowsCount.AddRow(0)

	// first expected SQL statement
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM migration_info;").WillReturnRows(rowsCount)

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	version, err := storage.GetDatabaseVersionInfo()
	if err != nil {
		t.Errorf("error was not expected while initializing database: %s", err)
	}

	// check the returned version
	assert.Equal(t, expectedVersion, version)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestGetDatabaseVersionInfoFirstReadFailure function checks the method
// Storage.getDatabaseVersionInfo.
func TestGetDatabaseVersionInfoFirstReadFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	rowsCount.AddRow(1)

	// first expected SQL statement
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM migration_info;").WillReturnError(mockedError)

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	version, err := storage.GetDatabaseVersionInfo()
	if err == nil {
		t.Errorf("error was expected while initializing database: %s", err)
	}

	// version returned in case of error
	expectedVersion := -1

	// check the returned version
	assert.Equal(t, expectedVersion, version)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestGetDatabaseVersionInfoSecondReadFailure function checks the method
// Storage.getDatabaseVersionInfo.
func TestGetDatabaseVersionInfoSecondReadFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	rowsCount.AddRow(1)

	// first expected SQL statement
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM migration_info;").WillReturnRows(rowsCount)

	// second expected SQL statement
	mock.ExpectQuery("SELECT version FROM migration_info LIMIT 1;").WillReturnError(mockedError)

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	version, err := storage.GetDatabaseVersionInfo()
	if err == nil {
		t.Errorf("error was expected while initializing database: %s", err)
	}

	// version returned in case of error
	expectedVersion := -1

	// check the returned version
	assert.Equal(t, expectedVersion, version)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestDatabaseInitialization function checks the method
// Storage.DatabaseInitialization.
func TestDatabaseInitialization(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow(0)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM states;").WillReturnRows(rows)
	mock.ExpectCommit()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.DatabaseInitialization()
	if err != nil {
		t.Errorf("error was not expected while initializing database: %s", err)
	}

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestDatabaseCleanup function checks the method Storage.DatabaseCleanup.
func TestDatabaseCleanup(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// note that list of statements is not initialized so just empty
	// transaction operations are expected there
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.DatabaseCleanup()
	if err != nil {
		t.Errorf("error was not expected while cleaning up database: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestDatabaseDropTables function checks the method
// Storage.DatabaseDropTables.
func TestDatabaseDropTables(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// note that list of statements is not initialized so just empty
	// transaction operations are expected there
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.DatabaseDropTables()
	if err != nil {
		t.Errorf("error was not expected while dropping database tables: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestDatabaseDropIndexes function checks the method
// Storage.DatabaseDropIndexes.
func TestDatabaseDropIndexes(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// note that list of statements is not initialized so just empty
	// transaction operations are expected there
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 1)

	// call the tested method
	err := storage.DatabaseDropIndexes()
	if err != nil {
		t.Errorf("error was not expected while dropping database indexes: %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestDropTableStatement function checks the helper function
// dropTableStatement.
func TestDropTableStatement(t *testing.T) {
	const expected = "DROP TABLE FOOBAR;"
	actual := main.DropTableStatement("FOOBAR")
	assert.Equal(t, actual, expected)
}

// TestDropIndexStatement function checks the helper function
// dropIndexStatement.
func TestDropIndexStatement(t *testing.T) {
	const expected = "DROP INDEX IF EXISTS FOOBAR;"
	actual := main.DropIndexStatement("FOOBAR")
	assert.Equal(t, actual, expected)
}

// TestDeleteFromTableStatement functions checks the helper function
// deleteFromTableStatement.
func TestDeleteFromTableStatement(t *testing.T) {
	const expected = "DELETE FROM FOOBAR;"
	actual := main.DeleteFromTableStatement("FOOBAR")
	assert.Equal(t, actual, expected)
}

// TestNewStorage checks whether constructor for new storage returns error for improper storage configuration
func TestNewStorageError(t *testing.T) {
	_, err := main.NewStorage(&main.StorageConfiguration{
		Driver: "non existing driver",
	})
	assert.EqualError(t, err, "driver non existing driver is not supported")
}

// TestNewStoragePostgreSQL function tests creating new storage with logs
func TestNewStoragePostgreSQL(t *testing.T) {
	_, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "postgres",
		PGUsername:    "user",
		PGPassword:    "password",
		PGHost:        "nowhere",
		PGPort:        1234,
		PGDBName:      "test",
		PGParams:      "",
		LogSQLQueries: true,
	})

	// we just happen to make connection without trying to actually connect
	assert.Nil(t, err)
}

// TestNewStorageSQLite3 function tests creating new storage with logs
func TestNewStorageSQLite3(t *testing.T) {
	_, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "sqlite3",
		LogSQLQueries: true,
	})

	// we just happen to make connection without trying to actually connect
	assert.Nil(t, err)
}

// TestClose function tests database close operation.
func TestClose(t *testing.T) {
	storage, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "sqlite3",
		LogSQLQueries: true,
	})

	// we just happen to make connection without trying to actually connect
	assert.NoError(t, err)

	// try to close the storage
	err = storage.Close()

	// it should not fail
	assert.NoError(t, err)
}

// TestConnection function checks the method Storage.Connection.
func TestConnection(t *testing.T) {
	storage, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "sqlite3",
		LogSQLQueries: false,
	})

	// we just happen to make connection without trying to actually connect
	assert.NoError(t, err)

	// try to retrieve connection
	returned := storage.Connection()
	assert.NotNil(t, returned)

	// try to close the storage
	err = storage.Close()

	// it should not fail
	assert.NoError(t, err)
}
