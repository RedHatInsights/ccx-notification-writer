// Copyright 2021 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// An implementation of various logging strategies supported by this service.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/logging.html

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

// logMessageInfo function records log info about message consumed from given
// topic, partition, and offset.
func logMessageInfo(consumer *KafkaConsumer, originalMessage *sarama.ConsumerMessage, parsedMessage IncomingMessage, event string) {
	log.Info().
		Int(offsetKey, int(originalMessage.Offset)).
		Int(partitionKey, int(originalMessage.Partition)).
		Str(topicKey, consumer.Configuration.Topic).
		Int(organizationKey, int(*parsedMessage.Organization)).
		Str(clusterKey, string(*parsedMessage.ClusterName)).
		Int(versionKey, int(parsedMessage.Version)).
		Msg(event)
}

// logUnparsedMessageError function records log info about consumed message
// that can not be parsed.
func logUnparsedMessageError(consumer *KafkaConsumer, originalMessage *sarama.ConsumerMessage, event string, err error) {
	log.Error().
		Int(offsetKey, int(originalMessage.Offset)).
		Str(topicKey, consumer.Configuration.Topic).
		Err(err).
		Msg(event)
}

// logMessageError function records log info about consumed message that
// contain (any) improper data.
func logMessageError(consumer *KafkaConsumer, originalMessage *sarama.ConsumerMessage, parsedMessage IncomingMessage, event string, err error) {
	log.Error().
		Int(offsetKey, int(originalMessage.Offset)).
		Str(topicKey, consumer.Configuration.Topic).
		Int(organizationKey, int(*parsedMessage.Organization)).
		Str(clusterKey, string(*parsedMessage.ClusterName)).
		Int(versionKey, int(parsedMessage.Version)).
		Err(err).
		Msg(event)
}

// logMessageWarning function records log info about consumed message that
// contain (any) data that are not 100% correct.
func logMessageWarning(consumer *KafkaConsumer, originalMessage *sarama.ConsumerMessage, parsedMessage IncomingMessage, event string) {
	log.Warn().
		Int(offsetKey, int(originalMessage.Offset)).
		Int(partitionKey, int(originalMessage.Partition)).
		Str(topicKey, consumer.Configuration.Topic).
		Int(organizationKey, int(*parsedMessage.Organization)).
		Str(clusterKey, string(*parsedMessage.ClusterName)).
		Int(versionKey, int(parsedMessage.Version)).
		Msg(event)
}

// logDuration function records log info about duration of any task/process.
func logDuration(tStart time.Time, tEnd time.Time, offset int64, key string) {
	duration := tEnd.Sub(tStart)
	log.Info().Int64(durationKey, duration.Microseconds()).Int64(offsetKey, offset).Msg(key)
}
