/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// Declaration of data types used by CCX Notification Writer.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/types.html

// CliFlags represents structure holding all command line arguments and flags.
type CliFlags struct {
	PerformDatabaseCleanup        bool
	PerformDatabaseInitialization bool
	PerformDatabaseInitMigration  bool
	PerformDatabaseDropTables     bool
	CheckConnectionToKafka        bool
	ShowVersion                   bool
	ShowAuthors                   bool
	ShowConfiguration             bool
	PrintNewReportsForCleanup     bool
	PerformNewReportsCleanup      bool
	PrintOldReportsForCleanup     bool
	PerformOldReportsCleanup      bool
	PrintReadErrorsForCleanup     bool
	PerformReadErrorsCleanup      bool
	MigrationInfo                 bool
	MaxAge                        string
	PerformMigrations             string
}

// RequestID data type is used to store the request ID supplied in input Kafka
// records as a unique identifier of payloads. Empty string represents a
// missing request ID.
type RequestID string

// ClusterReport represents the whole cluster report.
type ClusterReport string

// SchemaVersion is just a constant integer for now, max value 255. If we one
// day need more versions or combination of versions, it would be better
// consider upgrading to semantic versioning.
//
// TODO: provide expected schema version in configuration file
type SchemaVersion uint8

// DBDriver type for db driver enum.
type DBDriver int
