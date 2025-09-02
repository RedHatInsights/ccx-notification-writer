// Copyright 2025 Red Hat, Inc
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

// Definition of up and down steps for migration #0008

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/mig_0008_add_performance_indexes.html

import (
	"database/sql"

	mig "github.com/RedHatInsights/insights-operator-utils/migrations"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

// mig0008AddPerformanceIndexes adds performance indexes to the reported table
// to fix slow query execution and cleanup operations.
//
// (CCXDEV-15345) This migration addresses:
// 1. Slow ReadLastNotifiedRecordForClusterList queries (20+ minutes)
// 2. Slow cleanup operations (12+ minutes for 7K rows)
// 3. Missing indexes on frequently filtered columns
var mig0008AddPerformanceIndexes = mig.Migration{
	// up step: migrate database to version #0008
	StepUp: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0008AddPerformanceIndexes stepUp function")

		// Add composite index for the main query performance
		// This index covers the WHERE clause: event_type_id, state, org_id, cluster
		query1 := `
			CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reported_event_state_org_cluster 
			ON reported(event_type_id, state, org_id, cluster);
		`
		_, err := executeQuery(tx, query1)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create idx_reported_event_state_org_cluster index")
			return err
		}
		log.Debug().Msg("Created idx_reported_event_state_org_cluster index successfully")

		// Add index for cleanup operations on updated_at
		// This index covers the WHERE clause: updated_at < NOW() - INTERVAL
		query2 := `
			CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reported_updated_at 
			ON reported(updated_at);
		`
		_, err = executeQuery(tx, query2)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create idx_reported_updated_at index")
			return err
		}
		log.Debug().Msg("Created idx_reported_updated_at index successfully")

		// everything's fine
		return nil
	},
	// down step: migrate database to version #0007
	StepDown: func(tx *sql.Tx, _ types.DBDriver) error {
		log.Debug().Msg("Executing mig0008AddPerformanceIndexes stepDown function")

		// Drop the indexes in reverse order
		statements := []string{
			"DROP INDEX IF EXISTS idx_reported_updated_at;",
			"DROP INDEX IF EXISTS idx_reported_event_state_org_cluster;",
		}

		for _, statement := range statements {
			_, err := executeQuery(tx, statement)
			if err != nil {
				log.Error().Err(err).Msg("Failed to drop index")
				return err
			}
			log.Debug().Msg("Index dropped successfully")
		}

		// everything's fine
		return nil
	},
}
