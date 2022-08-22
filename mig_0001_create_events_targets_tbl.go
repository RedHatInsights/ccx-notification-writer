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

// mig0001CreateEventTargetsTbl migration creates table named `event_targets`.
// This table contains all event receivers, for example Notification Backend
// and ServiceLog
var mig0001CreateEventTargetsTbl = mig.Migration{
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		_, err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS event_targets (
				id              INTEGER NOT NULL,
				name            VARCHAR NOT NULL UNIQUE,
				metainfo        VARCHAR NOT NULL UNIQUE,
				PRIMARY KEY(id)
			)`)
		return err
	},
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		_, err := tx.Exec(`DROP TABLE event_targets`)
		return err
	},
}
