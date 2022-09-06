// Copyright 2020, 2021, 2022 Red Hat, Inc
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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	main "github.com/RedHatInsights/ccx-notification-writer"
	utils "github.com/RedHatInsights/insights-operator-utils/migrations"
	"github.com/stretchr/testify/assert"
)

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
