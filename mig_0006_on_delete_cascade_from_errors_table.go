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

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0005_create_read_errors_table.html

import (
	"database/sql"
	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0006OnCascadeDeleteFromErrorsTable migration creates table named `event_targets`.
// This table contains all event receivers, for example Notification Backend
// and ServiceLog
var mig0006OnCascadeDeleteFromErrorsTable = mig.Migration{
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0006OnCascadeDeleteFromErrorsTable stepUp function")
		query := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT read_errors_org_id_fkey,
		    ADD CONSTRAINT  read_errors_org_id_fkey
		       FOREIGN KEY (org_id, cluster, updated_at)
		       REFERENCES  new_reports(org_id, cluster, updated_at)
		       ON DELETE CASCADE;
		`
		_, err := executeQuery(tx, query)
		if err == nil {
			log.Debug().Msg("Constraints for table read_errors updated successfully: added on delete cascade")
		}
		return err
	},
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0006OnCascadeDeleteFromErrorsTable stepDown function")
		query := `
		    ALTER TABLE read_errors
		    DROP CONSTRAINT read_errors_org_id_fkey,
		    ADD CONSTRAINT  read_errors_org_id_fkey
		       FOREIGN KEY (org_id, cluster, updated_at)
		       REFERENCES  new_reports(org_id, cluster, updated_at);
		`
		_, err := executeQuery(tx, query)
		if err == nil {
			log.Debug().Msg("Constraints for table read_errors updated successfully: dropped on delete cascade")
		}
		return err
	},
}
