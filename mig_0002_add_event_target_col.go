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

import (
	"database/sql"

	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
)

var mig0002AddEventTargetCol = mig.Migration{
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		_, err := tx.Exec(`
		ALTER TABLE reported 
		ADD COLUMN IF NOT EXISTS event_type_id INTEGER,
		ADD FOREIGN KEY (event_type_id) REFERENCES event_targets(id)
	`)
		return err
	},
	StepDown: func(tx *sql.Tx, driver types.DBDriver) error {
		_, err := tx.Exec(`
		ALTER TABLE reported DROP COLUMN event_type_id
	`)
		return err
	},
}
