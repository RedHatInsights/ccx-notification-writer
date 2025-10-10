/*
Copyright © 2024 Red Hat, Inc.

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

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-service/packages/producer/kafka/kafka_producer_test.html

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	main "github.com/RedHatInsights/ccx-notification-writer"
	"github.com/RedHatInsights/insights-operator-utils/tests/helpers"
	"github.com/stretchr/testify/assert"

	"github.com/rs/zerolog"
)

var (
	brokerCfg = main.BrokerConfiguration{
		Addresses: "localhost:9092",
		Topic:     "platform.notifications.ingress",
		Enabled:   true,
	}
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}


// Test Producer creation with a non accessible Kafka broker
func TestNewProducerBadBroker(t *testing.T) {
	const expectedErrorMessage1 = "kafka: client has run out of available brokers to talk to: dial tcp: missing address"
	const expectedErrorMessage2 = "connect: connection refused"

	_, err := main.NewProducer(&main.BrokerConfiguration{
		Addresses: "",
		Topic:     "whatever",
		Enabled:   true,
	})
	assert.EqualError(t, err, expectedErrorMessage1)

	_, err = main.NewProducer(&brokerCfg)
	assert.ErrorContains(t, err, expectedErrorMessage2)
}

// TestProducerClose makes sure it's possible to close the connection
func TestProducerClose(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	prod := main.Producer{
		Configuration: &brokerCfg,
		Producer:      mockProducer,
	}

	err := prod.Close()
	assert.NoError(t, err, "failed to close Kafka producer")
}

func TestProducerNew(t *testing.T) {
	mockBroker := sarama.NewMockBroker(t, 0)
	defer mockBroker.Close()

	handlerMap := map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
			SetLeader(brokerCfg.Topic, 0, mockBroker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(brokerCfg.Topic, 0, -1, 0).
			SetOffset(brokerCfg.Topic, 0, -2, 0),
		"FetchRequest": sarama.NewMockFetchResponse(t, 1),
		"FindCoordinatorRequest": sarama.NewMockFindCoordinatorResponse(t).
			SetCoordinator(sarama.CoordinatorGroup, "", mockBroker),
		"OffsetFetchRequest": sarama.NewMockOffsetFetchResponse(t).
			SetOffset("", brokerCfg.Topic, 0, 0, "", sarama.ErrNoError),
	}
	mockBroker.SetHandlerByMap(handlerMap)

	// Use NewProducer with MockBroker-compatible settings
	prod, err := main.NewProducer(&main.BrokerConfiguration{
		Addresses:                 mockBroker.Addr(),
		Topic:                     brokerCfg.Topic,
		DisableAPIVersionsRequest: true, // MockBroker doesn't support API version negotiation
	})
	helpers.FailOnError(t, err)
	helpers.FailOnError(t, prod.Close())
}

func TestProducerSendEmptyMessage(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	kafkaProducer := main.Producer{
		Configuration: &brokerCfg,
		Producer:      mockProducer,
	}

	msgBytes, err := json.Marshal("")
	helpers.FailOnError(t, err)

	_, _, err = kafkaProducer.ProduceMessage(msgBytes)
	assert.NoError(t, err, "Couldn't produce message with given broker configuration")
	helpers.FailOnError(t, kafkaProducer.Close())
}

func TestPayloadTrackerProducerNew(t *testing.T) {
	mockBroker := sarama.NewMockBroker(t, 0)
	defer mockBroker.Close()

	handlerMap := map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
			SetLeader(brokerCfg.Topic, 0, mockBroker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(brokerCfg.Topic, 0, -1, 0).
			SetOffset(brokerCfg.Topic, 0, -2, 0),
		"FetchRequest": sarama.NewMockFetchResponse(t, 1),
		"FindCoordinatorRequest": sarama.NewMockFindCoordinatorResponse(t).
			SetCoordinator(sarama.CoordinatorGroup, "", mockBroker),
		"OffsetFetchRequest": sarama.NewMockOffsetFetchResponse(t).
			SetOffset("", brokerCfg.Topic, 0, 0, "", sarama.ErrNoError),
	}
	mockBroker.SetHandlerByMap(handlerMap)

	// Use NewPayloadTrackerProducer with MockBroker-compatible settings
	prod, err := main.NewPayloadTrackerProducer(&main.ConfigStruct{
		Broker: main.BrokerConfiguration{
			Addresses:                 mockBroker.Addr(),
			Topic:                     brokerCfg.Topic,
			DisableAPIVersionsRequest: true, // MockBroker doesn't support API version negotiation
		},
		Tracker: main.TrackerConfiguration{
			ServiceName: "test_service",
			Topic:       "test_tracker_topic",
		},
	})
	helpers.FailOnError(t, err)
	helpers.FailOnError(t, prod.Close())
}

func TestPayloadTrackerEmptyRequestID(t *testing.T) {
	tracker := main.PayloadTrackerProducer{
		ServiceName: "test",
		Producer:    main.Producer{},
	}

	err := tracker.TrackPayload("", time.Now(), "any_status")
	assert.NoError(t, err, "No error should be returned for empty request ID")
}

func TestPayloadTrackerValidRequestID(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	p := main.Producer{
		Configuration: &brokerCfg,
		Producer:      mockProducer,
	}

	tracker := main.PayloadTrackerProducer{
		ServiceName: "test",
		Producer:    p,
	}

	err := tracker.TrackPayload("anything", time.Now(), "any_status")
	assert.NoError(t, err, "No error should be returned for empty request ID")

	helpers.FailOnError(t, tracker.Close())
}
