// Copyright 2021 Red Hat, Inc
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

// CliFlags represents structure holding all command line arguments/flags.
type CliFlags struct {
	performDatabaseCleanup        bool
	performDatabaseInitialization bool
	performDatabaseDropTables     bool
	checkConnectionToKafka        bool
	showVersion                   bool
	showAuthors                   bool
	showConfiguration             bool
}

// RequestID is used to store the request ID supplied in input Kafka records as
// a unique identifier of payloads. Empty string represents a missing request
// ID.
type RequestID string

// KafkaOffset type for kafka offset
type KafkaOffset int64

// DBDriver type for db driver enum
type DBDriver int

const (
	// DBDriverSQLite3 shows that db driver is sqlite
	DBDriverSQLite3 DBDriver = iota
	// DBDriverPostgres shows that db driver is postgres
	DBDriverPostgres
	// DBDriverGeneral general sql(used for mock now)
	DBDriverGeneral
)
