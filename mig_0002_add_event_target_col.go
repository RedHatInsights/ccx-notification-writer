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
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0002_add_event_target_col.html

import (
	"database/sql"

	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0002AddEventTargetCol migration updates scheme of table named `reported`.
// New column containing foreign keys to `event_targets(id)` is added to the
// table.
var mig0002AddEventTargetCol = mig.Migration{
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0002AddEventTargetCol stepUp function")
		query := `
			ALTER TABLE reported 
			ADD COLUMN IF NOT EXISTS event_type_id INTEGER,
			ADD FOREIGN KEY (event_type_id) REFERENCES event_targets(id)
 		`
		_, err := executeQuery(tx, query)
		if err == nil {
			log.Debug().Msg("Table reported altered successfully")
		}
		return err
	},
	StepDown: func(tx *sql.Tx, driver types.DBDriver) error {
		log.Debug().Msg("Executing mig0002AddEventTargetCol stepDown function")
		query := "ALTER TABLE reported DROP COLUMN event_type_id"
		_, err := executeQuery(tx, query)
		if err == nil {
			log.Debug().Msg("Column event_type_id dropped successfully")
		}
		return err
	},
}
