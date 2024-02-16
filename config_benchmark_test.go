/*
Copyright © 2022, 2023 Red Hat, Inc.

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

// Benchmark for config module
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/config_benchmark_test.html

import (
	"testing"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// mustLoadBenchmarkConfiguration helper function loads configuration to be
// used by benchmarks.
func mustLoadBenchmarkConfiguration(b *testing.B) main.ConfigStruct {
	configuration, err := loadConfiguration()
	if err != nil {
		b.Fatal(err)
	}
	return configuration
}

// BenchmarkGetMetricsConfiguration measures the speed of
// GetMetricsConfiguration function from the main module.
func BenchmarkGetMetricsConfiguration(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

	// run the benchmarked code specified amount of times
	for i := 0; i < b.N; i++ {
		// call benchmarked function
		m := main.GetMetricsConfiguration(&configuration)

		b.StopTimer()
		if m.Namespace != "notification_writer" {
			b.Fatal("Wrong configuration: namespace = '" + m.Namespace + "'")
		}
		if m.Address != ":8080" {
			b.Fatal("Wrong configuration: address = '" + m.Address + "'")
		}
		b.StartTimer()
	}
}

// BenchmarkGetBrokerConfiguration measures the speed of
// GetBrokerConfiguration function from the main module.
func BenchmarkGetBrokerConfiguration(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

	// run the benchmarked code specified amount of times
	for i := 0; i < b.N; i++ {
		// call benchmarked function
		m := main.GetBrokerConfiguration(&configuration)

		b.StopTimer()
		if m.Addresses != "kafka:29092" {
			b.Fatal("Wrong configuration: addresses = '" + m.Addresses + "'")
		}
		b.StartTimer()
	}
}

// BenchmarkGetLoggingConfiguration measures the speed of
// GetLoggingConfiguration function from the main module.
func BenchmarkGetLoggingConfiguration(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

	// run the benchmarked code specified amount of times
	for i := 0; i < b.N; i++ {
		// call benchmarked function
		m := main.GetLoggingConfiguration(&configuration)

		b.StopTimer()
		if !m.Debug {
			b.Fatal("Wrong configuration: debug is set to false")
		}
		if m.LogLevel != "" {
			b.Fatal("Wrong configuration: loglevel = '" + m.LogLevel + "'")
		}
		b.StartTimer()
	}
}

// BenchmarkGetStorageConfiguration measures the speed of
// GetStorageConfiguration function from the main module.
func BenchmarkGetStorageConfiguration(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

	// run the benchmarked code specified amount of times
	for i := 0; i < b.N; i++ {
		// call benchmarked function
		m := main.GetStorageConfiguration(&configuration)

		b.StopTimer()
		if m.Driver != "postgres" {
			b.Fatal("Wrong configuration: driver = '" + m.Driver + "'")
		}
		b.StartTimer()
	}
}
