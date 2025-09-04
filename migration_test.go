// Copyright 2020, 2021, 2022, 2023 Red Hat, Inc
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

package main_test

// Tests if all migrations can be performed and giving the correct sequence of
// SQL statements.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/migration_test.html

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	main "github.com/RedHatInsights/ccx-notification-writer"
	utils "github.com/RedHatInsights/insights-operator-utils/migrations"
	"github.com/stretchr/testify/assert"
)

// SQL statements used in multiple tests
const (
	createIndexEventStateOrgCluster = "CREATE INDEX IF NOT EXISTS idx_reported_event_state_org_cluster ON reported\\(event_type_id, state, org_id, cluster\\);"
	createIndexUpdatedAt            = "CREATE INDEX IF NOT EXISTS idx_reported_updated_at ON reported\\(updated_at\\);"
	dropIndexEventStateOrgCluster   = "DROP INDEX IF EXISTS idx_reported_event_state_org_cluster;"
	dropIndexUpdatedAt              = "DROP INDEX IF EXISTS idx_reported_updated_at;"
)

// TestMigrationErrorDuringQueryingMigrationInfo1 test checks proper handling
// query errors during retrieving migration info from database (concretely
// during checking if migration table contains any value).
func TestMigrationErrorDuringQueryingMigrationInfo1(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected SQL query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"

	// the (only) query will throw an error
	mock.ExpectQuery(expectedQuery0).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 1
	assert.Error(t, main.Migrate(connection, 1), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// TestMigrationErrorDuringQueryingMigrationInfo2 test checks proper handling
// query error during retrieving migration info from database (concretely
// during reading actual database version).
func TestMigrationErrorDuringQueryingMigrationInfo2(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected SQL queries performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"

	// first query will succeed
	// second query will throw an error
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 1
	assert.Error(t, main.Migrate(connection, 1), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0001MigrationStepUp test checks migration #1, step up part.
func Test0001MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	// initial database version
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("0")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS event_targets .*"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate).WillReturnResult(resultCreate)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 1
	assert.NoError(t, main.Migrate(connection, 1))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0001MigrationStepUpOnMigrationFailure test checks migration #1 in case
// the table create fails.
func Test0001MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	// initial database version
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("0")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS event_targets .*"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// table create will fail
	mock.ExpectExec(expectedCreate).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 1
	assert.Error(t, main.Migrate(connection, 1), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0001MigrationStepDown test checks migration #1, step down part.
func Test0001MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("1")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDrop := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "DROP TABLE event_targets"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedDrop).WillReturnResult(resultDrop)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 0
	assert.NoError(t, main.Migrate(connection, 0))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0001MigrationStepDownOnMigrationFailure test checks migration #1, step
// down part in case table update fails
func Test0001MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("1")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "DROP TABLE event_targets"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// drop table will fail
	mock.ExpectExec(expectedDrop).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 0
	assert.Error(t, main.Migrate(connection, 0), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0002MigrationStepUp test checks migration #2, step up part.
func Test0002MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("0")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)
	resultAlter := sqlmock.NewResult(0, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS event_targets .*"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"
	expectedAlter := "ALTER TABLE reported ADD COLUMN .*"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate).WillReturnResult(resultCreate)
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 2
	assert.NoError(t, main.Migrate(connection, 2))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0002MigrationStepUpOnMigrationFailure test checks migration #2 in case
// the migration fails.
func Test0002MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	// initial database version
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("1")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := "ALTER TABLE reported ADD COLUMN .*"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// alter table will fail
	mock.ExpectExec(expectedAlter).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 2
	assert.Error(t, main.Migrate(connection, 2), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0002MigrationStepDown test checks migration #2, step down part.
func Test0002MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("2")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDrop := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "ALTER TABLE reported DROP COLUMN event_type_id"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedDrop).WillReturnResult(resultDrop)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 1
	assert.NoError(t, main.Migrate(connection, 1))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0002MigrationStepDownOnMigrationFailure test checks migration #2 in case
// the migration fails.
func Test0002MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("2")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "ALTER TABLE reported DROP COLUMN event_type_id"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// drop table will fail
	mock.ExpectExec(expectedDrop).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 1
	assert.Error(t, main.Migrate(connection, 1), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0003MigrationStepUp test checks migration #3, step up part.
func Test0003MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("0")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)
	resultAlter := sqlmock.NewResult(0, 1)
	resultInsert := sqlmock.NewResult(1, 2)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS event_targets .*"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"
	expectedInsert := "INSERT INTO event_targets \\(id, name, metainfo\\) VALUES .*"
	expectedAlter := "ALTER TABLE reported ADD COLUMN .*"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate).WillReturnResult(resultCreate)
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedInsert).WillReturnResult(resultInsert)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 3
	assert.NoError(t, main.Migrate(connection, 3))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0003MigrationStepUpOnMigrationFailure test checks migration #1 in case
// the migration fails.
func Test0003MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("0")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate := sqlmock.NewResult(0, 1)
	resultAlter := sqlmock.NewResult(0, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS event_targets .*"
	expectedInsert := "INSERT INTO event_targets \\(id, name, metainfo\\) VALUES .*"
	expectedAlter := "ALTER TABLE reported ADD COLUMN .*"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate).WillReturnResult(resultCreate)
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)

	// insert into table will fail
	mock.ExpectExec(expectedInsert).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// migration should end with error

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 3
	assert.Error(t, main.Migrate(connection, 3), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0003MigrationStepUp test checks migration #3, step down part.
func Test0003MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("3")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDelete := sqlmock.NewResult(0, 2)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDelete := "DELETE FROM event_targets"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedDelete).WillReturnResult(resultDelete)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 2
	assert.NoError(t, main.Migrate(connection, 2))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0003MigrationStepDownOnMigrationFailure test checks migration #1 in case
// the migration fails.
func Test0003MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("3")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDelete := "DELETE FROM event_targets"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// delete from table will fail
	mock.ExpectExec(expectedDelete).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 2
	assert.Error(t, main.Migrate(connection, 2), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0004MigrationStepUp test checks migration #4, step up part.
func Test0004MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("3")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultUpdate1 := sqlmock.NewResult(1, 1)
	resultAlter := sqlmock.NewResult(0, 1)
	resultUpdate2 := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedUpdate1 := "UPDATE reported SET event_type_id = 1 WHERE event_type_id IS NULL"
	expectedAlter := "ALTER TABLE reported ALTER COLUMN event_type_id SET NOT NULL"
	expectedUpdate2 := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedUpdate1).WillReturnResult(resultUpdate1)
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedUpdate2).WillReturnResult(resultUpdate2)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 4))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test the 4th migration in case the table update fails.
func Test0004MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("3")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedUpdate1 := "UPDATE reported SET event_type_id = 1 WHERE event_type_id IS NULL"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// table update will fail
	mock.ExpectExec(expectedUpdate1).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 4
	assert.Error(t, main.Migrate(connection, 4), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0004MigrationStepDown test checks migration #4, step down part.
func Test0004MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("4")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultAlter := sqlmock.NewResult(0, 2)
	resultUpdate1 := sqlmock.NewResult(1, 1)
	resultUpdate2 := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := "ALTER TABLE reported ALTER COLUMN event_type_id DROP NOT NULL"
	expectedUpdate1 := "UPDATE reported SET event_type_id = NULL WHERE event_type_id = 1"
	expectedUpdate2 := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedUpdate1).WillReturnResult(resultUpdate1)
	mock.ExpectExec(expectedUpdate2).WillReturnResult(resultUpdate2)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 3
	assert.NoError(t, main.Migrate(connection, 3))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test the 4th migration in case the table update fails.
func Test0004MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("4")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := "ALTER TABLE reported ALTER COLUMN event_type_id DROP NOT NULL"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// alter table will fail
	mock.ExpectExec(expectedAlter).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 3
	assert.Error(t, main.Migrate(connection, 3), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0005MigrationStepUp test checks migration #5, step up part.
func Test0005MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("4")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS read_errors .*"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate).WillReturnResult(resultCreate)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 5
	assert.NoError(t, main.Migrate(connection, 5))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0005MigrationStepUpOnMigrationFailure test checks migration #5 in case
// the migration fails.
func Test0005MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	// initial database version
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("4")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate := "CREATE TABLE IF NOT EXISTS read_errors .*"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// table create will fail
	mock.ExpectExec(expectedCreate).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 5
	assert.Error(t, main.Migrate(connection, 5), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0005MigrationStepDown test checks migration #5, step down part.
func Test0005MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("5")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDrop := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "DROP TABLE read_errors"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedDrop).WillReturnResult(resultDrop)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 4
	assert.NoError(t, main.Migrate(connection, 4))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0005MigrationStepDownOnMigrationFailure test checks migration #5 in case
// the migration fails.
func Test0005MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("5")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop := "DROP TABLE read_errors"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// drop table will fail
	mock.ExpectExec(expectedDrop).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migration should end with error

	// migrate to version 4
	assert.Error(t, main.Migrate(connection, 4), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0006MigrationStepUp test checks migration #6, step up part.
func Test0006MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("5")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultAlter := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT IF EXISTS read_errors_org_id_fkey,
		    DROP CONSTRAINT IF EXISTS read_errors_org_id_cluster_updated_at_fkey,
		    ADD CONSTRAINT  read_errors_org_id_cluster_updated_at_fkey
		       FOREIGN KEY \(org_id, cluster, updated_at\)
		       REFERENCES  new_reports\(org_id, cluster, updated_at\)
		       ON DELETE CASCADE;
		`
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 6
	assert.NoError(t, main.Migrate(connection, 6))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0006MigrationStepUpOnMigrationFailure test checks migration #6 in case
// the migration fails.
func Test0006MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("5")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT IF EXISTS read_errors_org_id_fkey,
		    DROP CONSTRAINT IF EXISTS read_errors_org_id_cluster_updated_at_fkey,
		    ADD CONSTRAINT  read_errors_org_id_cluster_updated_at_fkey
		       FOREIGN KEY \(org_id, cluster, updated_at\)
		       REFERENCES  new_reports\(org_id, cluster, updated_at\)
		       ON DELETE CASCADE;
		`

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// alter table will fail
	mock.ExpectExec(expectedAlter).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 6
	assert.Error(t, main.Migrate(connection, 6), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0006MigrationStepDown test checks migration #6, step down part.
func Test0006MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("6")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultAlter := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT read_errors_org_id_cluster_updated_at_fkey,
		    ADD CONSTRAINT  read_errors_org_id_cluster_updated_at_fkey
		       FOREIGN KEY \(org_id, cluster, updated_at\)
		       REFERENCES  new_reports\(org_id, cluster, updated_at\);
		`
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedAlter).WillReturnResult(resultAlter)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 5
	assert.NoError(t, main.Migrate(connection, 5))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0006MigrationStepDownOnMigrationFailure test checks migration #6 in case
// the migration fails.
func Test0006MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("6")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedAlter := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT read_errors_org_id_cluster_updated_at_fkey,
		    ADD CONSTRAINT  read_errors_org_id_cluster_updated_at_fkey
		       FOREIGN KEY \(org_id, cluster, updated_at\)
		       REFERENCES  new_reports\(org_id, cluster, updated_at\);
		`

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// alter table will fail
	mock.ExpectExec(expectedAlter).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 5
	assert.Error(t, main.Migrate(connection, 5), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0007MigrationStepUp test checks migration #7, step up part.
func Test0007MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("6")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultStatement := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedStatements := []string{
		"COMMENT ON TABLE event_targets IS 'specification of all event targets currently supported';",
		"COMMENT ON TABLE migration_info IS 'information about the latest DB schema and migration status';",
		"COMMENT ON TABLE new_reports IS 'new reports consumed from Kafka topic and stored to database in shrunk format';",
		"COMMENT ON TABLE notification_types IS 'list of all notification types used by Notification service';",
		"COMMENT ON TABLE read_errors IS 'errors that are detected during new reports collection';",
		"COMMENT ON TABLE reported IS 'notifications reported to user or skipped due to some conditions';",
		"COMMENT ON TABLE states IS 'states for each row stored in reported table';",
	}
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	for _, expectedStatement := range expectedStatements {
		mock.ExpectExec(expectedStatement).WillReturnResult(resultStatement)
	}
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 7
	assert.NoError(t, main.Migrate(connection, 7))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0007MigrationStepUpOnMigrationFailure test checks migration #7 in case
// the migration fails.
func Test0007MigrationStepUpOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("6")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedStatement := "COMMENT ON TABLE event_targets IS 'specification of all event targets currently supported';"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// statement to change comment will fail
	mock.ExpectExec(expectedStatement).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 7
	assert.Error(t, main.Migrate(connection, 7), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0007MigrationStepDown test checks migration #7, step down part.
func Test0007MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("7")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultStatement := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"
	expectedStatements := []string{
		"COMMENT ON TABLE event_targets IS null;",
		"COMMENT ON TABLE migration_info IS null;",
		"COMMENT ON TABLE new_reports IS null;",
		"COMMENT ON TABLE notification_types IS null;",
		"COMMENT ON TABLE read_errors IS null;",
		"COMMENT ON TABLE reported IS null;",
		"COMMENT ON TABLE states IS null;",
	}
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	for _, expectedStatement := range expectedStatements {
		mock.ExpectExec(expectedStatement).WillReturnResult(resultStatement)
	}
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 6
	assert.NoError(t, main.Migrate(connection, 6))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0007MigrationStepDownOnMigrationFailure test checks migration #7 in case
// the migration fails.
func Test0007MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("7")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedStatement := "COMMENT ON TABLE event_targets IS null;"

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// statement to change comment will fail
	mock.ExpectExec(expectedStatement).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 6
	assert.Error(t, main.Migrate(connection, 6), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0008MigrationStepUp test checks migration #8, step up part.
func Test0008MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("7")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate1 := sqlmock.NewResult(0, 1)
	resultCreate2 := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate1 := createIndexEventStateOrgCluster
	expectedCreate2 := createIndexUpdatedAt
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedCreate1).WillReturnResult(resultCreate1)
	mock.ExpectExec(expectedCreate2).WillReturnResult(resultCreate2)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 8
	assert.NoError(t, main.Migrate(connection, 8))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0008MigrationStepUpOnMigrationFailure1 test checks migration #8 in case
// the first index creation fails.
func Test0008MigrationStepUpOnMigrationFailure1(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("7")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate1 := createIndexEventStateOrgCluster

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// first index creation will fail
	mock.ExpectExec(expectedCreate1).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 8
	assert.Error(t, main.Migrate(connection, 8), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0008MigrationStepUpOnMigrationFailure2 test checks migration #8 in case
// the second index creation fails.
func Test0008MigrationStepUpOnMigrationFailure2(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("7")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultCreate1 := sqlmock.NewResult(0, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedCreate1 := createIndexEventStateOrgCluster
	expectedCreate2 := createIndexUpdatedAt

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// first index creation will succeed
	mock.ExpectExec(expectedCreate1).WillReturnResult(resultCreate1)

	// second index creation will fail
	mock.ExpectExec(expectedCreate2).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 8
	assert.Error(t, main.Migrate(connection, 8), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0008MigrationStepDown test checks migration #8, step down part.
func Test0008MigrationStepDown(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("8")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDrop1 := sqlmock.NewResult(0, 1)
	resultDrop2 := sqlmock.NewResult(0, 1)
	resultUpdate := sqlmock.NewResult(1, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop1 := dropIndexUpdatedAt
	expectedDrop2 := dropIndexEventStateOrgCluster
	expectedUpdate := "UPDATE migration_info SET version=\\$1;"

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(expectedDrop1).WillReturnResult(resultDrop1)
	mock.ExpectExec(expectedDrop2).WillReturnResult(resultDrop2)
	mock.ExpectExec(expectedUpdate).WillReturnResult(resultUpdate)
	mock.ExpectCommit()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 7
	assert.NoError(t, main.Migrate(connection, 7))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test0008MigrationStepDownOnMigrationFailure test checks migration #8 in case
// the index drop fails.
func Test0008MigrationStepDownOnMigrationFailure(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"version"})
	rows.AddRow("8")

	count := sqlmock.NewRows([]string{"count"})
	count.AddRow("1")

	resultDrop1 := sqlmock.NewResult(0, 1)

	// expected query performed by tested function
	expectedQuery0 := "SELECT COUNT\\(\\*\\) FROM migration_info;"
	expectedQuery1 := "SELECT version FROM migration_info;"
	expectedDrop1 := dropIndexUpdatedAt
	expectedDrop2 := dropIndexEventStateOrgCluster

	// queries to retrieve DB version should succeed
	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// first index drop will succeed
	mock.ExpectExec(expectedDrop1).WillReturnResult(resultDrop1)

	// second index drop will fail
	mock.ExpectExec(expectedDrop2).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	// prepare list of all migrations
	utils.Set(main.All())

	// migrate to version 7
	assert.Error(t, main.Migrate(connection, 7), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}
