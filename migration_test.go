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

	utils.Set(main.All())

	// migration should end with error

	// migrate to version 1
	assert.Error(t, main.Migrate(connection, 1), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

func Test0001MigrationStepUp(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 1))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 0))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 2))

	// check if all expectations were met
	checkAllExpectations(t, mock)

}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 1))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 3))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 2))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 4))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test the 4th migration in case the table update fails.
func Test0004MigrationStepUp2(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// table update will fail
	mock.ExpectExec(expectedUpdate1).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	utils.Set(main.All())

	// migration should end with error
	assert.Error(t, main.Migrate(connection, 4), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 3))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// Test the 4th migration in case the table update fails.
func Test0004MigrationStepDown2(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery0).WillReturnRows(count)
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectBegin()

	// alter table will fail
	mock.ExpectExec(expectedAlter).WillReturnError(mockedError)

	// so we expect roll back instead of transaction commit
	mock.ExpectRollback()
	mock.ExpectClose()

	utils.Set(main.All())

	// migration should end with error
	assert.Error(t, main.Migrate(connection, 3), mockedError)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 5))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 4))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 6))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 5))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 7))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

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

	utils.Set(main.All())
	assert.NoError(t, main.Migrate(connection, 6))

	// check if all expectations were met
	checkAllExpectations(t, mock)
}
