/*
Copyright © 2021 Red Hat, Inc.

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

import (
	"encoding/json"
)

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
	MaxAge                        string
}

// RequestID data type is used to store the request ID supplied in input Kafka
// records as a unique identifier of payloads. Empty string represents a
// missing request ID.
type RequestID string

// KafkaOffset is a data type representing offset in Kafka topic.
type KafkaOffset int64

// OrgID data type represents organization ID.
type OrgID uint32

// AccountNumber data type represents account number for a given report.
type AccountNumber uint32

// ClusterName data type represents name of cluster in format
// c8590f31-e97e-4b85-b506-c45ce1911a12 (ie. in UUID format).
type ClusterName string

// RuleID represents type for rule id.
type RuleID string

// ErrorKey represents type for error key.
type ErrorKey string

// ClusterReport represents the whole cluster report.
type ClusterReport string

// SchemaVersion is just a constant integer for now, max value 255. If we one
// day need more versions or combination of versions, it would be better
// consider upgrading to semantic versioning.
type SchemaVersion uint8

// ReportItem data structure represents a single (hit) rule of the string
// encoded report.
type ReportItem struct {
	Module       RuleID          `json:"component"`
	ErrorKey     ErrorKey        `json:"key"`
	TemplateData json.RawMessage `json:"details"`
}

// DBDriver type for db driver enum.
type DBDriver int
