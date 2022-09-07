/*
Copyright © 2022 Red Hat, Inc.

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

package main_test

// Benchmark for storage backend - the Postgres or RDS
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/db_benchmark_test.html

import (
	"os"
	"testing"

	"database/sql"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

const (
	configFileEnvName = "CCX_NOTIFICATION_WRITER_CONFIG_FILE"
	configFileName    = "tests/benchmark"
)

// initLogging function initializes logging
func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// loadConfiguration function loads configuration prepared to be used by
// benchmarks
func loadConfiguration() (main.ConfigStruct, error) {
	os.Clearenv()

	err := os.Setenv(configFileEnvName, configFileName)
	if err != nil {
		return main.ConfigStruct{}, err
	}

	config, err := main.LoadConfiguration(configFileEnvName, configFileName)
	if err != nil {
		return main.ConfigStruct{}, err
	}

	return config, nil
}

// initializeStorage function initializes storage and perform connection to
// database
func initializeStorage(config main.ConfigStruct) (*sql.DB, error) {
	storageConfiguration := main.GetStorageConfiguration(config)
	storage, err := main.NewStorage(storageConfiguration)
	if err != nil {
		return nil, err
	}

	connection := storage.Connection()

	// check if connection has been established
	if storage.Connection() == nil {
		return nil, err
	}

	return connection, nil
}

// unfortunatelly there seems to be no other way how to handle per-benchmark
// setup properly than to use global variables
var connection *sql.DB
var initialized bool = false

// setup function performs all DB initializations needed for DB benchmarks. Can
// be called from any benchmark.
func setup(b *testing.B) *sql.DB {
	if initialized {
		return connection
	}

	initLogging()

	config, err := loadConfiguration()
	if err != nil {
		b.Fatal(err)
	}

	connection, err = initializeStorage(config)
	if err != nil {
		b.Fatal(err)
	}

	initialized = true

	return connection
}

// BenchmarkInsertIntoReportedTableV1 checks the speed of inserting into
// reported table without event_type column
func BenchmarkInsertIntoReportedTableV1(b *testing.B) {
	report := ""

	runBenchmarkInsertIntoReportedTableV1(b, 1, &report)
}

// BenchmarkInsertIntoReportedTableV2 checks the speed of inserting into
// reported table with event_type column
func BenchmarkInsertIntoReportedTableV2(b *testing.B) {
	report := ""

	runBenchmarkInsertIntoReportedTableV2(b, 1, &report)
}
