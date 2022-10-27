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

// Benchmark for config module

import (
	"testing"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

func mustLoadBenchmarkConfiguration(b *testing.B) main.ConfigStruct {
	configuration, err := loadConfiguration()
	if err != nil {
		b.Fatal(err)
	}
	return configuration
}

// BenchmarkGetMetricsConfigurarion measures the speed of
// GetMetricsConfiguration function from the main module.
func BenchmarkGetMetricsConfigurarion(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

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

// BenchmarkGetBrokerConfigurarion measures the speed of
// GetBrokerConfiguration function from the main module.
func BenchmarkGetBrokerConfigurarion(b *testing.B) {
	initLogging()
	configuration := mustLoadBenchmarkConfiguration(b)

	for i := 0; i < b.N; i++ {
		// call benchmarked function
		m := main.GetBrokerConfiguration(&configuration)

		b.StopTimer()
		if m.Address != "kafka:29092" {
			b.Fatal("Wrong configuration: address = '" + m.Address + "'")
		}
		b.StartTimer()
	}

}
