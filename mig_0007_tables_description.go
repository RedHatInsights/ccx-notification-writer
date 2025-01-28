// Copyright 2023 Red Hat, Inc
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

// Definition of up and down steps for migration #0007

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0007_tables_description.html

import (
	"database/sql"

	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0007TablesDescription updates description for all tables in CCX
// Notification database schema.
var mig0007TablesDescription = mig.Migration{
	// up step: migrate database to version #0007
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0007TablesDescription stepUp function")
		// explicit statements to set up table descriptions
		statements := []string{
			"COMMENT ON TABLE event_targets IS 'specification of all event targets currently supported';",
			"COMMENT ON TABLE migration_info IS 'information about the latest DB schema and migration status';",
			"COMMENT ON TABLE new_reports IS 'new reports consumed from Kafka topic and stored to database in shrunk format';",
			"COMMENT ON TABLE notification_types IS 'list of all notification types used by Notification service';",
			"COMMENT ON TABLE read_errors IS 'errors that are detected during new reports collection';",
			"COMMENT ON TABLE reported IS 'notifications reported to user or skipped due to some conditions';",
			"COMMENT ON TABLE states IS 'states for each row stored in reported table';",
		}

		for _, statement := range statements {
			// update comment/description
			_, err := executeQuery(tx, statement)

			// check for any error
			if err != nil {
				return err
			}
			log.Debug().Msg("Table description has been updated")
		}

		// everything's fine
		return nil
	},
	// up down: migrate database to version #0006
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0007TablesDescription stepDown function")
		// explicit statements to clean table descriptions
		statements := []string{
			"COMMENT ON TABLE event_targets IS null;",
			"COMMENT ON TABLE migration_info IS null;",
			"COMMENT ON TABLE new_reports IS null;",
			"COMMENT ON TABLE notification_types IS null;",
			"COMMENT ON TABLE read_errors IS null;",
			"COMMENT ON TABLE reported IS null;",
			"COMMENT ON TABLE states IS null;",
		}

		for _, statement := range statements {
			// delete comment/description
			_, err := executeQuery(tx, statement)

			// check for any error
			if err != nil {
				return err
			}
			log.Debug().Msg("Table description has been cleaned")
		}

		// everything's fine
		return nil
	},
}
