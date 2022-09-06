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

package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/migration.html

import (
	"database/sql"

	utils "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
)

// migrations is a list of migrations that, when applied in their order,
// create the most recent version of the database from scratch.
var migrations = []utils.Migration{
	mig0001CreateEventTargetsTbl,
	mig0002AddEventTargetCol,
	mig0003PopulateEventTables,
	mig0004PopulateEventTables,
}

// All returns "migration" , the list of implemented utils.Migration
func All() []utils.Migration {
	return migrations
}

// Migrate interfaces with migration utils to update
// the database db with the specified target version
func Migrate(db *sql.DB, target utils.Version) error {
	err := utils.SetDBVersion(db, types.DBDriverPostgres, target)
	defer func() {
		err = db.Close()
	}()
	return err
}
