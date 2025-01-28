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

package main

// Definition of up and down steps for migration #0003

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0003_populate_event_tables.html

import (
	"database/sql"

	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0003PopulateEventTables migration add all existing event targets into
// `event_targets` table. Currently Notification Backend and ServiceLog targets
// exist and will be inserted to the table.
//
// Also `reported` table content is changed by this migration: all existing
// records are updated to contain foreign key to `event_target == notification
// backend`.
var mig0003PopulateEventTables = mig.Migration{
	// up step: migrate database to version #0003
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0003PopulateEventTables stepUp function")
		query := `
 			INSERT INTO event_targets (id, name, metainfo) 
 				VALUES (1, 'notifications backend', 'the target of the report is the ccx notification service back end'),
 				       (2, 'service log', 'the target of the report is the ServiceLog')`
		result, err := executeQuery(tx, query)
		if err == nil {
			rows, _ := result.RowsAffected()
			log.Debug().Int64(rowsInsertedMessage, rows).Msg("Table event_targets altered successfully")
		}
		return err
	},
	// up down: migrate database to version #0002
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0003PopulateEventTables stepDown function")
		query := "DELETE FROM event_targets"
		result, err := executeQuery(tx, query)
		if err == nil {
			rows, _ := result.RowsAffected()
			log.Debug().Int64(rowsDeletedMessage, rows).Msg("Table event_targets cleaned up successfully")
		}
		return err
	},
}
