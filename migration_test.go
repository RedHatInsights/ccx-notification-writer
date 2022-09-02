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

func Test0001Migration(t *testing.T) {
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

func Test0002Migration(t *testing.T) {
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
