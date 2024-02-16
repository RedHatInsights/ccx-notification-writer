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
// consumer.go
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/consumer_test.html

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// Constants used by unit tests
const (
	ConsumerReport = `{
		"fingerprints": [],
		"info": [],
		"reports": [],
		"skips": [],
		"system": {}
	}`
)

// Variables used by unit tests
var (
	ExpectedOrgID         = main.OrgID(1)
	ExpectedAccountNumber = main.AccountNumber(1234)
	ExpectedClusterName   = main.ClusterName("84f7eedc-0dd8-49cd-9d4d-f6646df3a5bc")
	LastCheckedAt         = time.Unix(25, 0).UTC()
)

// TestNewConsumerBadBroker function checks the consumer creation by
// using a non-accessible Kafka broker.
func TestNewConsumerBadBrokerNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to: dial tcp: missing address"

	// invalid broker configuration
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses: "",
		Topic:     "whatever",
		Group:     "whatever",
		Enabled:   true,
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerLocalBroker function checks the consumer creation by using a
// non-accessible Kafka broker. This test assumes there is no local Kafka
// instance currently running
func TestNewConsumerLocalBrokerNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses: "localhost:9092",
		Topic:     "platform.notifications.ingress",
		Group:     "",
		Enabled:   true,
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerConsumerGroup function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. Consumer group is enabled and setup for this test.
func TestNewConsumerSaramaConfigNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses: "localhost:9092",
		Topic:     "platform.notifications.ingress",
		Group:     "test-group",
		Enabled:   true,
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerTLSEnabled function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. TSL is enabled in broker configuration.
func TestNewConsumerTLSEnabledNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SSL",
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerSASLEnabled function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. SASL is enabled in broker configuration.
func TestNewConsumerSASLEnabledNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SASL_",
		SaslUsername:     "sasl_user",
		SaslPassword:     "sasl_password",
		SaslMechanism:    "",
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestSaramaConfigFromBrokerWithSASLEnabled function checks that the Sarama config returned
// for a broker configuration with SASL enabled contains the expected fields
func TestSaramaConfigFromBrokerWithSASLEnabledNoSASLMechanism(t *testing.T) {
	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SASL_",
		SaslUsername:     "sasl_user",
		SaslPassword:     "sasl_password",
		SaslMechanism:    "",
	}

	saramaConfig, err := main.SaramaConfigFromBrokerConfig(&brokerConfiguration)
	assert.Nil(t, err)
	assert.True(t, saramaConfig.Net.SASL.Enable)
	assert.Equal(t, saramaConfig.Net.SASL.User, brokerConfiguration.SaslUsername)
	assert.Equal(t, saramaConfig.Net.SASL.Password, brokerConfiguration.SaslPassword)
	assert.Nil(t, saramaConfig.Net.SASL.SCRAMClientGeneratorFunc, "SCRAM client generator function should not be created with given config")
}

// TestSaramaConfigFromBrokerWithSASLEnabled function checks that the Sarama config returned
// for a broker configuration with SASL enabled using SCRAM authentication mechanism
// contains expected fields
func TestSaramaConfigFromBrokerWithSASLEnabledSCRAMAuth(t *testing.T) {
	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SASL_",
		SaslUsername:     "sasl_user",
		SaslPassword:     "sasl_password",
		SaslMechanism:    sarama.SASLTypeSCRAMSHA512,
	}

	saramaConfig, err := main.SaramaConfigFromBrokerConfig(&brokerConfiguration)
	assert.Nil(t, err)
	assert.True(t, saramaConfig.Net.SASL.Enable)
	assert.Equal(t, saramaConfig.Net.SASL.User, brokerConfiguration.SaslUsername)
	assert.Equal(t, saramaConfig.Net.SASL.Password, brokerConfiguration.SaslPassword)
	assert.NotNil(t, saramaConfig.Net.SASL.SCRAMClientGeneratorFunc, "SCRAM client generator function should have been created with given config")
}

// TestSaramaConfigFromBrokerWithSASLEnabled function checks that the Sarama config returned
// for a broker configuration with SASL enabled using unhandled authentication mechanism
// contains expected fields
func TestSaramaConfigFromBrokerWithSASLEnabledUnexpectedAuthMechanism(t *testing.T) {
	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SASL_",
		SaslUsername:     "sasl_user",
		SaslPassword:     "sasl_password",
		SaslMechanism:    sarama.SASLTypeSCRAMSHA256,
	}

	saramaConfig, err := main.SaramaConfigFromBrokerConfig(&brokerConfiguration)
	assert.Nil(t, err)
	assert.True(t, saramaConfig.Net.SASL.Enable)
	assert.Equal(t, saramaConfig.Net.SASL.User, brokerConfiguration.SaslUsername)
	assert.Equal(t, saramaConfig.Net.SASL.Password, brokerConfiguration.SaslPassword)
	assert.Nil(t, saramaConfig.Net.SASL.SCRAMClientGeneratorFunc, "SCRAM client generator function should not be created with given config")
}

// TestNewConsumerSASLEnabled function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. SASL is enabled in broker configuration.
func TestNewConsumerSASLEnabled(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SASL_",
		SaslUsername:     "sasl_user",
		SaslPassword:     "sasl_password",
		SaslMechanism:    "",
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerCertPath function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. Valid cert. path is provided by tests.
func TestNewConsumerCertPathNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "kafka: client has run out of available brokers to talk to"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses: "localhost:9092",
		Topic:     "platform.notifications.ingress",
		Group:     "",
		Enabled:   true,
		CertPath:  "testdata/cert.pem",
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"invalid cert path",
	)
}

// TestNewConsumerInvalidCertPath function checks the consumer creation by using a
// non-accesible Kafka broker. This test assumes there is no local Kafka
// instance currently running. Invalid cert. path is provided by tests.
func TestNewConsumerInvalidCertPathNonAccessibleBroker(t *testing.T) {
	// expected error message
	const expectedErr = "open /foo/bar/baz: no such file or directory"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Addresses:        "localhost:9092",
		Topic:            "platform.notifications.ingress",
		Group:            "",
		Enabled:          true,
		SecurityProtocol: "SSL",
		CertPath:         "/foo/bar/baz",
	}

	// dummy storage not really usable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(&brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.Contains(t, err.Error(), expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"invalid cert path",
	)
}

// TestParseEmptyMessage checks how empty message is handled by
// consumer.
func TestParseEmptyMessage(t *testing.T) {
	// empty message to be parsed
	const emptyMessage = ""

	// try to parse the message
	_, err := main.ParseMessage([]byte(emptyMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "unexpected end of JSON input")
}

// TestParseMessageWithWrongContent checks how message with wrong
// (unexpected) content is handled by consumer.
func TestParseMessageWithWrongContent(t *testing.T) {
	// JSON-encoded message with unexpected content
	const message = `{"this":"is", "not":"expected content"}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(message))

	// check for error - it should be reported
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required attribute")
}

// TestParseMessageWithImproperJSON checks how message with wrong
// content which is not parseable as JSON is processed.
func TestParseMessageWithImproperJSON(t *testing.T) {
	// message not in JSON format
	const message = `"this_is_not_json_dude"`

	// try to parse the message
	_, err := main.ParseMessage([]byte(message))

	// check for error - it should be reported
	assert.EqualError(
		t,
		err,
		"json: cannot unmarshal string into Go value of type main.IncomingMessage",
	)
}

// TestParseProperMessage checks the parsing of properly declared
// message.
func TestParseProperMessage(t *testing.T) {
	// message in JSON format with all required attributes
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	message, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should not be reported
	assert.Nil(t, err)

	// check returned values
	assert.Equal(t, main.OrgID(1), *message.Organization)
	assert.Equal(t, ExpectedClusterName, *message.ClusterName)
	assert.Equal(t, ExpectedAccountNumber, *message.AccountNumber)
}

// TestParseProperMessageWrongOrgID checks the parsing of message
// with wrong organization ID.
func TestParseProperMessageWrongOrgID(t *testing.T) {
	// message with wrong organization ID attribute.
	message := `{
		"OrgID": "foobar",
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(message))

	// check for error - it should be reported
	assert.EqualError(
		t,
		err,
		"json: cannot unmarshal string into Go struct field IncomingMessage.OrgID of type main.OrgID",
	)
}

// TestParseProperMessageWrongAccountNumber checks the parsing of message
// with wrong organization ID.
func TestParseProperMessageWrongAccountNumber(t *testing.T) {
	// message with wrong account number attribute.
	message := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": "foobar",
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(message))

	// check for error - it should be reported
	assert.EqualError(
		t,
		err,
		"json: cannot unmarshal string into Go struct field IncomingMessage.AccountNumber of type main.AccountNumber",
	)
}

// TestParseProperMessageWrongClusterName checks the parsing of message
// with wrong cluster name.
func TestParseProperMessageWrongClusterName(t *testing.T) {
	// message with wrong cluster name
	message := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "this is not an UUID",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(message))

	// check for error - it should be reported
	assert.EqualError(t, err, "cluster name is not a UUID")
}

// TestParseMessageWithoutOrgID checks the parsing of improperly
// declared message - OrgID attribute is missing.
func TestParseMessageWithoutOrgID(t *testing.T) {
	// message without OrgID attribute
	ConsumerMessage := `{
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "missing required attribute 'OrgID'")
}

// TestParseMessageWithoutAccountNumber checks the parsing of improperly
// declared message - AccountNumber attribute is missing.
func TestParseMessageWithoutAccountNumber(t *testing.T) {
	// message without AccountNumber attribute
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "missing required attribute 'AccountNumber'")
}

// TestParseMessageWithoutClusterName checks the parsing of improperly
// declared message - ClusterName attribute is missing.
func TestParseMessageWithoutClusterName(t *testing.T) {
	// message without ClusterName attribute
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "missing required attribute 'ClusterName'")
}

// TestParseMessageWithoutReport checks the parsing of improperly
// declared message - Report attribute is missing.
func TestParseMessageWithoutReport(t *testing.T) {
	// message without Report attribute
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "missing required attribute 'Report'")
}

// TestParseMessageEmptyReport checks the parsing of improperly
// declared message - report attribute is empty.
func TestParseMessageEmptyReport(t *testing.T) {
	// message with empty Report attribute
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report": {},
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "Improper report structure, missing key fingerprints")
}

// TestParseMessageNullReport checks the parsing of improperly
// declared message - Report attribute is null.
func TestParseMessageNullReport(t *testing.T) {
	// message with empty Report attribute
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report": null,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	// try to parse the message
	_, err := main.ParseMessage([]byte(ConsumerMessage))

	// check for error - it should be reported
	assert.EqualError(t, err, "missing required attribute 'Report'")
}

// NewDummyConsumer function constructs new instance of (not running)
// KafkaConsumer.
func NewDummyConsumer(s main.Storage) *main.KafkaConsumer {
	brokerCfg := main.BrokerConfiguration{
		Addresses: "localhost:1234",
		Topic:     "topic",
		Group:     "group",
	}
	return &main.KafkaConsumer{
		Configuration: brokerCfg,
		Storage:       s,
		Ready:         make(chan bool),
	}
}

// TestProcessEmptyMessage check the behaviour of function ProcessMessage with
// empty message on input.
func TestProcessEmptyMessage(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// prepare an empty message
	message := sarama.ConsumerMessage{}

	// try to process the message
	_, err := dummyConsumer.ProcessMessage(&message)

	// check for errors - it should be reported
	assert.EqualError(t, err, "unexpected end of JSON input")

	// nothing should be written into storage
	assert.Equal(t, 0, mockStorage.writeReportCalled)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

// TestProcessMessageWithWrongDateFormat check the behaviour of function
// ProcessMessage with message with wrong date.
func TestProcessMessageWithWrongDateFormat(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// prepare a message
	message := sarama.ConsumerMessage{}

	// fill in a message payload
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "2020.01.23 16:15:59"
	}`
	message.Value = []byte(ConsumerMessage)

	// try to process the message
	_, err := dummyConsumer.ProcessMessage(&message)

	// check for errors - it should be reported
	assert.EqualError(t, err, "parsing time \"2020.01.23 16:15:59\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \".01.23 16:15:59\" as \"-\"")

	// nothing should be written into storage
	assert.Equal(t, 0, mockStorage.writeReportCalled)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

// TestProcessMessageFromFuture check the behaviour of function ProcessMessage
// with message with wrong date.
func TestProcessMessageFromFuture(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// prepare a message
	message := sarama.ConsumerMessage{}

	// fill in a message payload
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "2099-01-01T23:59:59.999999999Z"
	}`
	message.Value = []byte(ConsumerMessage)

	// try to process the message
	_, err := dummyConsumer.ProcessMessage(&message)

	// check for errors - it should be reported
	assert.EqualError(t, err, "got a message from the future")

	// nothing should be written into storage
	assert.Equal(t, 0, mockStorage.writeReportCalled)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

// TestProcessCorrectMessage check the behaviour of function ProcessMessage for
// correct message.
func TestProcessCorrectMessage(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// prepare a message
	message := sarama.ConsumerMessage{}

	// fill in a message payload
	ConsumerMessage := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`
	message.Value = []byte(ConsumerMessage)

	// message is correct -> one record should be written into the database
	_, err := dummyConsumer.ProcessMessage(&message)

	// check for error - it should not be reported
	assert.Nil(t, err)

	// one record should be written into the storage
	assert.Equal(t, 1, mockStorage.writeReportCalled)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

// TestConsumerSetup function checks the method KafkaConsumer.Setup().
func TestConsumerSetup(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// channel that needs to be closed in Setup
	dummyConsumer.Ready = make(chan bool)

	// try to setup the consumer (without consumer group)
	err := dummyConsumer.Setup(nil)

	// and check for any error
	assert.Nil(t, err)
}

// TestConsumerCleanup function checks the method KafkaConsumer.Cleanup().
func TestConsumerCleanup(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// try to cleanup the consumer (without consumer group)
	err := dummyConsumer.Cleanup(nil)

	// and check for any error
	assert.Nil(t, err)
}

// TestConsumerClose function checks the method KafkaConsumer.Close().
func TestConsumerClose(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// try to close the consumer (without consumer group)
	err := dummyConsumer.Close()

	// and check for any error
	assert.Nil(t, err)
}

// TestConsumerCloseCancel function checks the method KafkaConsumer.Close().
func TestConsumerCloseCancel(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// setup cancel hook
	_, cancel := context.WithCancel(context.Background())
	dummyConsumer.Cancel = cancel

	// try to close the consumer (without consumer group)
	err := dummyConsumer.Close()

	// and check for any error
	assert.Nil(t, err)
}

// TestHandleNilMessage function checks the method
// KafkaConsumer.HandleMessage() for nil input.
func TestHandleNilMessage(_ *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	// nil message
	dummyConsumer.HandleMessage(nil)
}

// TestHandleEmptyMessage function checks the method
// KafkaConsumer.HandleMessage() for empty message value.
func TestHandleEmptyMessage(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	message := sarama.ConsumerMessage{}

	message.Value = []byte("")

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// empty message
	dummyConsumer.HandleMessage(&message)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(1), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

// TestHandleCorrectMessage function checks the method
// KafkaConsumer.HandleMessage() for correct input.
func TestHandleCorrectMessage(t *testing.T) {
	// construct mock storage
	mockStorage := NewMockStorage()

	// construct dummy consumer
	dummyConsumer := NewDummyConsumer(&mockStorage)

	message := sarama.ConsumerMessage{}
	value := `{
		"OrgID": ` + fmt.Sprint(ExpectedOrgID) + `,
		"AccountNumber": ` + fmt.Sprint(ExpectedAccountNumber) + `,
		"ClusterName": "` + string(ExpectedClusterName) + `",
		"Report":` + ConsumerReport + `,
		"LastChecked": "` + LastCheckedAt.UTC().Format(time.RFC3339) + `"
	}`

	message.Value = []byte(value)

	// counter checks
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())

	// correct message
	dummyConsumer.HandleMessage(&message)

	// counter checks
	assert.Equal(t, uint64(1), dummyConsumer.GetNumberOfSuccessfullyConsumedMessages())
	assert.Equal(t, uint64(0), dummyConsumer.GetNumberOfErrorsConsumingMessages())
}

func testShrinkedMessage(t *testing.T, inputMessage main.Report) {
	assert.NotContains(t, inputMessage, "system")
	assert.NotContains(t, inputMessage, "fingerprints")
	assert.NotContains(t, inputMessage, "skips")
	assert.NotContains(t, inputMessage, "info")
	assert.NotContains(t, inputMessage, "pass")
	assert.NotContains(t, inputMessage, "analysis_metadata")
}

// TestShrinkEmptyMessage tests the function to remove unneeded attributes from incoming message
func TestShrinkEmptyMessage(t *testing.T) {
	var inputMessage main.Report = make(map[string]*json.RawMessage)
	main.ShrinkMessage(&inputMessage)
	testShrinkedMessage(t, inputMessage)
}

// TestShrinkMessageWithAnalysisMetadata tests the function to remove unneeded attributes from incoming message
func TestShrinkMessageWithAnalysisMetadata(t *testing.T) {
	var inputMessage main.Report = make(map[string]*json.RawMessage)
	analysisMetadata := json.RawMessage("{}")
	inputMessage["analysis_metadata"] = &analysisMetadata

	main.ShrinkMessage(&inputMessage)
	testShrinkedMessage(t, inputMessage)
}
