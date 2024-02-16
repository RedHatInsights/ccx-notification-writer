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

package main_test

// Unit test definitions for functions and methods defined in source file
// logging.go
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/logging_test.html

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// Constants used by various tests
const (
	testOrganizationID = 6502
	testClusterName    = "thisIsClusterName"
	testTopicName      = "thisIsTopicName"
	testError          = "this is test error only"
	testEventMessage   = "this is event message"
)

// init function is called before tests
func init() {
	// set default logging level regardles of config made in code
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// checkCapture function perform checks if standard output has been captured by
// unit test
func checkCapture(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// Construct consumer used only for tests.
func constructTestConsumer() *main.KafkaConsumer {
	// mocked broker configuration
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses: "address",
		Topic:     testTopicName,
		Group:     "group",
		Enabled:   true,
	}

	// construct mocked consumer
	return &main.KafkaConsumer{
		Configuration: brokerConfiguration,
		ConsumerGroup: nil,
	}
}

// Construct parsed message used only for tests.
func constructParsedMessage() main.IncomingMessage {
	// mocked message
	var orgID main.OrgID = main.OrgID(testOrganizationID)
	var clusterName main.ClusterName = main.ClusterName(testClusterName)

	// construct mocked parsed message
	return main.IncomingMessage{
		Organization: &orgID,
		ClusterName:  &clusterName,
		Version:      99,
	}
}

// TestLogDuration check the logDuration function from the main module
func TestLogDuration(t *testing.T) {
	startTime := time.Date(2000, time.November, 10, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, time.November, 10, 23, 0, 1, 0, time.UTC)

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogDuration(startTime, endTime, 9999, "test message")
	})

	// check the captured text
	checkCapture(t, err)

	// expected text(s) writen onto error output
	assert.Contains(t, output, "test message") // key
	assert.Contains(t, output, "9999")         // offset
	assert.Contains(t, output, "1000000")      // duration
}

// TestLogMessageInfo check the logMessageInfo function from the main module
func TestLogMessageInfo(t *testing.T) {
	consumer := constructTestConsumer()
	originalMessage := sarama.ConsumerMessage{}
	parsedMessage := constructParsedMessage()

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogMessageInfo(consumer, &originalMessage, parsedMessage, testEventMessage)
	})

	// check the captured text
	checkCapture(t, err)

	// expected text(s) writen onto error output
	assert.Contains(t, output, "organization")
	assert.Contains(t, output, testClusterName)
	assert.Contains(t, output, testTopicName)
	assert.Contains(t, output, testEventMessage)
	assert.Contains(t, output, "6502")
	assert.Contains(t, output, "99") // version
}

// TestLogMessageError check the logMessageError function from the main module
func TestLogMessageError(t *testing.T) {
	consumer := constructTestConsumer()
	originalMessage := sarama.ConsumerMessage{}
	parsedMessage := constructParsedMessage()

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogMessageError(consumer, &originalMessage, parsedMessage, testEventMessage, errors.New(testError))
	})

	// check the captured text
	checkCapture(t, err)

	// expected text(s) writen onto error output
	assert.Contains(t, output, "organization")
	assert.Contains(t, output, testClusterName)
	assert.Contains(t, output, testTopicName)
	assert.Contains(t, output, testError)
	assert.Contains(t, output, testEventMessage)
	assert.Contains(t, output, "6502")
	assert.Contains(t, output, "99") // version
}

// TestLogUnparsedMessageError check the logUnparsedMessageError function from
// the main module
func TestLogUnparsedMessageError(t *testing.T) {
	consumer := constructTestConsumer()
	originalMessage := sarama.ConsumerMessage{}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogUnparsedMessageError(consumer, &originalMessage, testEventMessage, errors.New(testError))
	})

	// check the captured text
	checkCapture(t, err)

	// expected text(s) writen onto error output
	assert.Contains(t, output, testTopicName)
	assert.Contains(t, output, testError)
	assert.Contains(t, output, testEventMessage)
}

// TestLogMessageWarning check the logMessageWarning function from the main
// module
func TestLogMessageWarning(t *testing.T) {
	consumer := constructTestConsumer()
	originalMessage := sarama.ConsumerMessage{}
	parsedMessage := constructParsedMessage()

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogMessageWarning(consumer, &originalMessage, parsedMessage, testEventMessage)
	})

	// check the captured text
	checkCapture(t, err)

	// expected text(s) writen onto error output
	assert.Contains(t, output, "organization")
	assert.Contains(t, output, testClusterName)
	assert.Contains(t, output, testTopicName)
	assert.Contains(t, output, testEventMessage)
	assert.Contains(t, output, "6502")
	assert.Contains(t, output, "99") // version
}
