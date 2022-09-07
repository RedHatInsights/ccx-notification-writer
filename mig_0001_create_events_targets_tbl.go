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
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0001_create_events_targets_tbl.html

import (
	"database/sql"
	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0001CreateEventTargetsTbl migration creates table named `event_targets`.
// This table contains all event receivers, for example Notification Backend
// and ServiceLog
var mig0001CreateEventTargetsTbl = mig.Migration{
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0001CreateEventTargetsTbl stepUp function")
		query := "CREATE TABLE IF NOT EXISTS event_targets (" +
			"id INTEGER NOT NULL, " +
			"name VARCHAR NOT NULL UNIQUE, " +
			"metainfo VARCHAR NOT NULL UNIQUE, " +
			"PRIMARY KEY(id))"
		log.Debug().Str("query", query).Msg("Executing")
		_, err := tx.Exec(query)
		if err == nil {
			log.Debug().Msg("Table event_targets created successfully")
		}
		return err
	},
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0001CreateEventTargetsTbl stepDown function")
		query := "DROP TABLE event_targets"
		log.Debug().Str("query", query).Msg("")
		_, err := tx.Exec(query)
		if err == nil {
			log.Debug().Msg("Table event_targets dropped successfully")
		}
		return err
	},
}
