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

package main_test

// Unit test definitions for functions and methods defined in source file
// ccx_notification_writer.go
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/ccx_notification_writer_test.html

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// TestShowVersion checks the function showVersion
func TestShowVersion(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowVersion()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "CCX Notification Writer version")
}

// TestShowAuthors checks the function showAuthors
func TestShowAuthors(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowAuthors()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Pavel Tisnovsky")
	assert.Contains(t, output, "Red Hat Inc.")
}

// TestShowConfiguration checks the function ShowConfiguration
func TestShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Broker = main.BrokerConfiguration{
		Address: "broker_address",
		Topic:   "broker_topic",
	}
	configuration.Metrics = main.MetricsConfiguration{
		Namespace: "metrics_namespace",
	}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.ShowConfiguration(configuration)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "broker_address")
	assert.Contains(t, output, "broker_topic")
	assert.Contains(t, output, "metrics_namespace")
}

// TestDoSelectedOperationShowVersion checks the function showVersion called
// via doSelectedOperation function
func TestDoSelectedOperationShowVersion(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       true,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(configuration, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "CCX Notification Writer version")
}

// TestDoSelectedOperationShowAuthors checks the function showAuthors called
// via doSelectedOperation function
func TestDoSelectedOperationShowAuthors(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       true,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(configuration, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Pavel Tisnovsky")
	assert.Contains(t, output, "Red Hat Inc.")
}

// TestDoSelectedOperationShowConfiguration checks the function
// showConfiguration called via doSelectedOperation function
func TestDoSelectedOperationShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Broker = main.BrokerConfiguration{
		Address: "broker_address",
		Topic:   "broker_topic",
	}
	configuration.Metrics = main.MetricsConfiguration{
		Namespace: "metrics_namespace",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: true,
	}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		code, err := main.DoSelectedOperation(configuration, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "broker_address")
	assert.Contains(t, output, "broker_topic")
	assert.Contains(t, output, "metrics_namespace")
}
