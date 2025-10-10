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

// Package kafka contains an implementation of Producer interface that can be
// used to produce (that is send) messages to properly configured Kafka broker.
package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-service/producer
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-service/packages/producer/kafka/kafka_producer.html

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/IBM/sarama"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/rs/zerolog/log"
)

const (
	// StatusReceived is reported when a new payload is received.
	StatusReceived = "received"
	// StatusMessageProcessed is reported when the message of a payload has been processed.
	StatusMessageProcessed = "processed"
	// StatusSuccess is reported upon a successful handling of a payload.
	StatusSuccess = "success"
	// StatusError is reported when the handling of a payload fails for any reason.
	StatusError = "error"
)

// Producer is a holder of a Sarama Producer and its configuration
type Producer struct {
	Configuration *BrokerConfiguration
	Producer      sarama.SyncProducer
}

// PayloadTrackerProducer a holder of a Producer and the service name used
// when sending tracking information
type PayloadTrackerProducer struct {
	ServiceName string
	Producer
}

// PayloadTrackerMessage represents content of messages
// sent to the Payload Tracker topic in Kafka.
type PayloadTrackerMessage struct {
	Service   string `json:"service"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Date      string `json:"date"`
}

// NewProducer instantiates a Kafka Producer
func NewProducer(config *BrokerConfiguration) (*Producer, error) {
	saramaConfig, err := saramaProducerConfigFromBrokerConfig(config)
	if err != nil {
		log.Error().Err(err).Msg("Unable to create a valid Kafka configuration")
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(strings.Split(config.Addresses, ","), saramaConfig)
	if err != nil {
		log.Error().Str("Kafka address", config.Addresses).Err(err).Msg("unable to start a Kafka producer")
		return nil, err
	}

	return &Producer{
		Configuration: config,
		Producer:      producer,
	}, nil
}

// NewPayloadTrackerProducer instantiates a PayloadTrackerProducer
func NewPayloadTrackerProducer(config *ConfigStruct) (*PayloadTrackerProducer, error) {
	kafkaConfig := config.Broker
	kafkaConfig.Topic = config.Tracker.Topic

	producer, err := NewProducer(&kafkaConfig)
	if err != nil {
		return nil, err
	}

	return &PayloadTrackerProducer{
		ServiceName: config.Tracker.ServiceName,
		Producer:    *producer,
	}, nil
}

// ProduceMessage produces message to selected topic. That function returns
// partition ID and offset of new message or an error value in case of any
// problem on broker side.
func (producer *Producer) ProduceMessage(msg []byte) (partitionID int32, offset int64, err error) {
	// no-op when producer is disabled
	// (this logic allows us to enable/disable producer on the fly
	if !producer.Configuration.Enabled {
		return
	}

	producerMsg := &sarama.ProducerMessage{
		Topic: producer.Configuration.Topic,
		Value: sarama.ByteEncoder(msg),
	}

	partitionID, offset, err = producer.Producer.SendMessage(producerMsg)
	if err != nil {
		log.Error().Err(err).Msg("failed to produce message to Kafka")
	} else {
		log.Info().Int("partition", int(partitionID)).Int("offset", int(offset)).Msg("message sent")
	}
	return
}

// Close allow the Sarama producer to be gracefully closed
func (producer *Producer) Close() error {
	log.Info().Msg("Shutting down kafka producer")
	if err := producer.Producer.Close(); err != nil {
		log.Error().Err(err).Msg("unable to close Kafka producer")
		return err
	}

	return nil
}

// TrackPayload publishes the status of a payload with the given request ID to
// the payload tracker Kafka topic. If the request ID is empty, the payload
// will not be tracked and the event is logged as a warning.
func (producer *PayloadTrackerProducer) TrackPayload(reqID types.RequestID, timestamp time.Time, status string) error {
	if len(reqID) == 0 {
		log.Warn().Msg("request ID is missing, null or empty")
		return nil
	}

	trackerMsg := PayloadTrackerMessage{
		Service:   producer.ServiceName,
		RequestID: string(reqID),
		Status:    status,
		Date:      timestamp.UTC().Format(time.RFC3339Nano),
	}

	jsonBytes, err := json.Marshal(trackerMsg)
	if err != nil {
		return err
	}
	_, _, err = producer.Producer.ProduceMessage(jsonBytes)
	if err != nil {
		log.Error().Err(err).Msgf(
			"unable to produce payload tracker message (request ID: '%s', timestamp: %v, status: '%s')",
			reqID, timestamp, status)
		return err
	}

	return nil
}
