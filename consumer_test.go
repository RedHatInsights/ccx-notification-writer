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

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// Variables used by unit tests
var (
	ExpectedOrgID         = main.OrgID(1)
	ExpectedAccountNumber = main.AccountNumber(1234)
	ExpectedClusterName   = main.ClusterName("84f7eedc-0dd8-49cd-9d4d-f6646df3a5bc")
	LastCheckedAt         = time.Unix(25, 0).UTC()

	ConsumerReport = `{
		"fingerprints": [],
		"info": [],
		"reports": [],
		"skips": [],
		"system": {}
	}`
)

// TestNewConsumerBadBroker function checks the consumer creation by
// using a non accessible Kafka broker.
func TestNewConsumerBadBroker(t *testing.T) {
	const expectedErr = "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)"

	// invalid broker configuration
	var brokerConfiguration = main.BrokerConfiguration{
		Address: "",
		Topic:   "whatever",
		Group:   "whatever",
		Enabled: true,
	}

	// dummy storage not really useable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.EqualError(t, err, expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
	)
}

// TestNewConsumerLocalBroker function checks the consumer creation by using a
// non accessible Kafka broker. This test assumes there is no local Kafka
// instance currently running
func TestNewConsumerLocalBroker(t *testing.T) {
	const expectedErr = "kafka: client has run out of available brokers to talk to (Is your cluster reachable?)"

	// valid broker configuration for local Kafka instance
	var brokerConfiguration = main.BrokerConfiguration{
		Address: "localhost:9092",
		Topic:   "platform.notifications.ingress",
		Group:   "",
		Enabled: true,
	}

	// dummy storage not really useable as the driver is not specified
	dummyStorage := main.NewFromConnection(nil, 1)

	// try to construct new consumer
	mockConsumer, err := main.NewConsumer(brokerConfiguration, dummyStorage)

	// check that error is really reported
	assert.EqualError(t, err, expectedErr)

	// test the return value
	assert.Equal(
		t,
		(*main.KafkaConsumer)(nil),
		mockConsumer,
		"consumer.New should return nil instead of Consumer implementation",
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

// TestParseMessageWithoutReport checks the parsing of improperly
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
