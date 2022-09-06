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

// Exit codes
const (
	exitStatusConfigurationError = iota
	exisStatusConnectionInitializationError
	exisStatusConnectionClosingError
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

// BenchmarkFoo is dummy ATM
func BenchmarkFoo(b *testing.B) {
	connection := setup(b)
	if connection == nil {
		b.Fatal()
	}

	b.ResetTimer()

	// perform DB benchmark
	for i := 0; i < b.N; i++ {
	}
}

// BenchmarkBar is dummy ATM
func BenchmarkBar(b *testing.B) {
	connection := setup(b)
	if connection == nil {
		b.Fatal()
	}

	b.ResetTimer()

	// perform DB benchmark
	for i := 0; i < b.N; i++ {
	}
}
