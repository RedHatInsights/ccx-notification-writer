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
