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

package main

// This file contains interface for any consumer that is able to process
// messages. It also contains implementation of Apache Kafka consumer.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/consumer.html

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/IBM/sarama"
	tlsutils "github.com/RedHatInsights/insights-operator-utils/tls"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Common constants used in the following code
const (
	// key for topic name used in structured log messages
	topicKey = "topic"

	// key for broker group name used in structured log messages
	groupKey = "group"

	// key for message offset used in structured log messages
	offsetKey = "offset"

	// key for message partition used in structured log messages
	partitionKey = "partition"

	// key for organization ID used in structured log messages
	organizationKey = "organization"

	// key for cluster ID used in structured log messages
	clusterKey = "cluster"

	// key for data schema version message type used in structured log messages
	versionKey = "version"

	// key for duration message type used in structured log messages
	durationKey = "duration"
)

// Attribute names that are used in incoming messages stored as JSONs
const (
	systemAttribute           = "system"
	fingerprintsAttribute     = "fingerprints"
	skipsAttribute            = "skips"
	infoAttribute             = "info"
	passAttribute             = "pass"
	analysisMetadataAttribute = "analysis_metadata"
	reportsAttribute          = "reports"
)

// CurrentSchemaVersion represents the currently supported data schema version
//
// TODO: make this value configurable
const CurrentSchemaVersion = types.SchemaVersion(2)

// SaramaVersion is the version of Kafka API used in sarama client
var SaramaVersion = sarama.V3_8_0_0

// Report represents report send in a message consumed from any broker
type Report map[string]*json.RawMessage

// IncomingMessage data structure is representation of message consumed from
// any broker. Some values might be missing in incorrectly formatted message so
// pointers are used to be able to distinguish true values from nils.
type IncomingMessage struct {
	Organization  *types.OrgID         `json:"OrgID"`
	AccountNumber *types.AccountNumber `json:"AccountNumber"`
	ClusterName   *types.ClusterName   `json:"ClusterName"`
	Report        *Report              `json:"Report"`
	// LastChecked is a date in format "2020-01-23T16:15:59.478901889Z"
	LastChecked string              `json:"LastChecked"`
	Version     types.SchemaVersion `json:"Version"`
	RequestID   types.RequestID     `json:"RequestId"`
}

// KafkaConsumer in an implementation of Consumer interface
// Example:
//
// kafkaConsumer, err := consumer.New(brokerCfg, storage)
//
//	if err != nil {
//	    panic(err)
//	}
//
// kafkaConsumer.Serve()
//
// err := kafkaConsumer.Stop()
//
//	if err != nil {
//	    panic(err)
//	}
type KafkaConsumer struct {
	Configuration                        BrokerConfiguration
	ConsumerGroup                        sarama.ConsumerGroup
	Storage                              Storage
	Tracker                              *PayloadTrackerProducer
	numberOfSuccessfullyConsumedMessages uint64
	numberOfErrorsConsumingMessages      uint64
	Ready                                chan bool
	Cancel                               context.CancelFunc
}

// DefaultSaramaConfig is a config which will be used by default
// here you can use specific version of a protocol for example
// useful for testing
var DefaultSaramaConfig *sarama.Config

// NewConsumer constructs new implementation of Consumer interface
func NewConsumer(brokerConfiguration *BrokerConfiguration, storage Storage) (*KafkaConsumer, error) {
	return NewWithSaramaConfig(brokerConfiguration, DefaultSaramaConfig, storage)
}

// NewWithSaramaConfig constructs new implementation of Consumer interface with
// custom Sarama configuration.
func NewWithSaramaConfig(
	brokerConfiguration *BrokerConfiguration,
	saramaConfig *sarama.Config,
	storage Storage,
) (*KafkaConsumer, error) {
	var err error
	// check if custom Sarama configuration is provided
	if saramaConfig == nil {
		// read configuration provided via configuration file and/or
		// environment variables
		saramaConfig, err = saramaConfigFromBrokerConfig(brokerConfiguration)
		if err != nil {
			log.Error().Err(err).Msg("unable to create sarama configuration from current broker configuration")
			return nil, err
		}
	}

	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(brokerConfiguration.Addresses, ","), brokerConfiguration.Group, saramaConfig)
	if err != nil {
		return nil, err
	}

	// construct Apache Kafka consumer
	consumer := &KafkaConsumer{
		Configuration:                        *brokerConfiguration,
		ConsumerGroup:                        consumerGroup,
		Storage:                              storage,
		numberOfSuccessfullyConsumedMessages: 0,
		numberOfErrorsConsumingMessages:      0,
		Ready:                                make(chan bool),
	}

	return consumer, nil
}

// Serve method starts listening for messages and processing them. It blocks
// current thread.
func (consumer *KafkaConsumer) Serve() {
	ctx, cancel := context.WithCancel(context.Background())
	consumer.Cancel = cancel

	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims.
			if err := consumer.ConsumerGroup.Consume(ctx, []string{consumer.Configuration.Topic}, consumer); err != nil {
				log.Fatal().Err(err).Msg("Unable to recreate Kafka session")
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Info().Err(ctx.Err()).Msg("Stopping consumer")
				return
			}

			log.Info().Msg("Created new kafka session")

			// consumer is prepared to accept messages
			consumer.Ready = make(chan bool)
		}
	}()

	// Await till the consumer has been set up
	log.Info().Msg("Waiting for consumer to become ready")
	<-consumer.Ready
	log.Info().Msg("Finished waiting for consumer to become ready")

	// Actual processing is done in goroutine created by Sarama
	// (see ConsumeClaim below)
	log.Info().Msg("Started serving consumer")
	<-ctx.Done()
	log.Info().Msg("Context cancelled, exiting")

	cancel()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	log.Info().Msg("New session has been setup")
	// Mark the consumer as ready
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Info().Msg("New session has been finished")
	return nil
}

// ConsumeClaim starts a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Info().
		Int64(offsetKey, claim.InitialOffset()).
		Msg("Starting messages loop")

	// try to retrieve offset of latest message consumed
	latestMessageOffset, err := consumer.Storage.GetLatestKafkaOffset()
	if err != nil {
		log.Error().Msg("Unable to get latest offset")
		latestMessageOffset = 0
	}
	log.Info().
		Int64("Offset in DB", int64(latestMessageOffset)).
		Msg("Latest offset read from database")

	// start consuming messages
	for message := range claim.Messages() {
		msgOffset := types.KafkaOffset(message.Offset)
		// skip over old (already consumed messages)
		if msgOffset <= latestMessageOffset {
			log.Warn().
				Int64(offsetKey, message.Offset).
				Msg("This offset was already processed")
		}

		// handle new message
		consumer.HandleMessage(message)

		session.MarkMessage(message, "")
		if msgOffset > latestMessageOffset {
			latestMessageOffset = msgOffset
			log.Info().
				Int64(offsetKey, int64(latestMessageOffset)).
				Msg("Updating latest message offset")
		}
	}

	return nil
}

// Close method closes all resources used by consumer
func (consumer *KafkaConsumer) Close() error {
	if consumer.Cancel != nil {
		consumer.Cancel()
	}

	// close consumer group(s)
	if consumer.ConsumerGroup != nil {
		if err := consumer.ConsumerGroup.Close(); err != nil {
			log.Error().
				Err(err).
				Msg("Unable to close consumer group")
		}
	}

	if consumer.Tracker != nil {
		if err := consumer.Tracker.Close(); err != nil {
			log.Error().Err(err).Msg("unable to close payload tracker Kafka producer")
		}
	}

	return nil
}

// GetNumberOfSuccessfullyConsumedMessages returns number of consumed messages
// since creating KafkaConsumer object
func (consumer *KafkaConsumer) GetNumberOfSuccessfullyConsumedMessages() uint64 {
	return consumer.numberOfSuccessfullyConsumedMessages
}

// GetNumberOfErrorsConsumingMessages returns number of errors during consuming messages
// since creating KafkaConsumer object
func (consumer *KafkaConsumer) GetNumberOfErrorsConsumingMessages() uint64 {
	return consumer.numberOfErrorsConsumingMessages
}

// HandleMessage method handles the message and does all checking, logging,
// metrics update, etc
func (consumer *KafkaConsumer) HandleMessage(msg *sarama.ConsumerMessage) {
	if msg == nil {
		log.Error().Msg("nil message")
		return
	}

	log.Info().
		Int64(offsetKey, msg.Offset).
		Int32(partitionKey, msg.Partition).
		Str(topicKey, msg.Topic).
		Time("message_timestamp", msg.Timestamp).
		Msg("Started processing message")

	// update metric
	ConsumedMessages.Inc()

	// try to process the message
	startTime := time.Now()
	requestID, err := consumer.ProcessMessage(msg)
	timeAfterProcessingMessage := time.Now()
	messageProcessingDuration := timeAfterProcessingMessage.Sub(startTime).Seconds()

	_ = consumer.Tracker.TrackPayload(requestID, timeAfterProcessingMessage, StatusMessageProcessed)

	log.Info().
		Int64(offsetKey, msg.Offset).
		Int32(partitionKey, msg.Partition).
		Str("Request ID", string(requestID)).
		Str(topicKey, msg.Topic).
		Msgf("Processing of message took '%v' seconds", messageProcessingDuration)

	// Something went wrong while processing the message.
	if err != nil {
		// update metric
		ConsumingErrors.Inc()

		log.Error().
			Err(err).
			Msg("Error processing message consumed from Kafka")
		consumer.numberOfErrorsConsumingMessages++

		_ = consumer.Tracker.TrackPayload(requestID, timeAfterProcessingMessage, StatusError)
	} else {
		// The message was processed successfully.
		consumer.numberOfSuccessfullyConsumedMessages++
		_ = consumer.Tracker.TrackPayload(requestID, timeAfterProcessingMessage, StatusSuccess)

	}

	totalMessageDuration := time.Since(startTime)
	log.Info().
		Int64("duration", totalMessageDuration.Milliseconds()).
		Int64(offsetKey, msg.Offset).
		Msg("Message consumed")
}

// checkMessageVersion function verifies incoming data's version is the
// expected one
func checkMessageVersion(consumer *KafkaConsumer, message *IncomingMessage, msg *sarama.ConsumerMessage) {
	if message.Version != CurrentSchemaVersion {
		const warning = "Received data with unexpected version."
		logMessageWarning(consumer, msg, *message, warning)
	}
}

// shrinkMessage function shrink the original message by removing unused parts.
func shrinkMessage(message *Report) {
	// delete all unneeded 'root' attributes
	tryToDeleteAttribute(message, systemAttribute)
	tryToDeleteAttribute(message, fingerprintsAttribute)
	tryToDeleteAttribute(message, skipsAttribute)
	tryToDeleteAttribute(message, infoAttribute)
	tryToDeleteAttribute(message, passAttribute)
	tryToDeleteAttribute(message, analysisMetadataAttribute)
}

// tryToDeleteAttribute function deletes selected attribute from input map. If
// attribute does not exists, it is skipped silently.
func tryToDeleteAttribute(message *Report, attributeName string) {
	_, found := (*message)[attributeName]
	if found {
		delete(*message, attributeName)
	}
	// let's ingore 'not-found' state as we just need to remove the
	// attribute, not to check message schema
}

// ProcessMessage method processes an incoming message
func (consumer *KafkaConsumer) ProcessMessage(msg *sarama.ConsumerMessage) (types.RequestID, error) {
	tStart := time.Now()

	// Step #1: parse the incomming message
	log.Info().
		Int(offsetKey, int(msg.Offset)).
		Str(topicKey, consumer.Configuration.Topic).
		Str(groupKey, consumer.Configuration.Group).
		Msg("Consumed")
	message, err := parseMessage(msg.Value)
	if err != nil {
		logUnparsedMessageError(consumer, msg, "Error parsing message from Kafka", err)
		return message.RequestID, err
	}

	// update metric - number of parsed messages
	ParsedIncomingMessage.Inc()

	logMessageInfo(consumer, msg, message, "Read")

	tRead := time.Now()

	_ = consumer.Tracker.TrackPayload(message.RequestID, tRead, StatusReceived)

	// Step #2: check message (schema) version
	checkMessageVersion(consumer, &message, msg)

	// update metric - number of messages with successful schema check
	CheckSchemaVersion.Inc()

	// Step #3: marshall report into byte slice to figure out original length
	reportAsBytes, err := json.Marshal(*message.Report)
	if err != nil {
		logMessageError(consumer, msg, message, "Error marshalling report", err)
		return message.RequestID, err
	}

	// update metric - number of marshaled reports
	MarshalReport.Inc()

	logMessageInfo(consumer, msg, message, "Marshalled")
	tMarshalled := time.Now()

	// Step #4: shrink the Report structure
	logMessageInfo(consumer, msg, message, "Shrinking message")
	shrinkMessage(message.Report)
	shrunkAsBytes, err := json.Marshal(*message.Report)
	if err != nil {
		logMessageError(consumer, msg, message, "Error marshalling skrinked report", err)
		return message.RequestID, err
	}
	logShrunkMessage(reportAsBytes, shrunkAsBytes)

	// update metric - number of shrunk reports
	ShrinkReport.Inc()

	tShrunk := time.Now()

	// Step #5: check the last checked timestamp
	lastCheckedTime, err := time.Parse(time.RFC3339Nano, message.LastChecked)
	if err != nil {
		logMessageError(consumer, msg, message, "Error parsing date from message", err)
		return message.RequestID, err
	}

	lastCheckedTimestampLagMinutes := time.Since(lastCheckedTime).Minutes()
	if lastCheckedTimestampLagMinutes < 0 {
		errorMessage := "got a message from the future"
		logMessageError(consumer, msg, message, errorMessage, nil)
		return message.RequestID, errors.New(errorMessage)
	}

	// update metric - number of messages with last checked timestamp
	CheckLastCheckedTimestamp.Inc()

	logMessageInfo(consumer, msg, message, "Time ok")

	tTimeCheck := time.Now()

	kafkaOffset := types.KafkaOffset(msg.Offset)

	// Step #6: write the shrunk report into storage (database)
	err = consumer.Storage.WriteReportForCluster(
		*message.Organization,
		*message.AccountNumber,
		*message.ClusterName,
		types.ClusterReport(shrunkAsBytes),
		tTimeCheck,
		kafkaOffset,
	)
	if err != nil {
		if err == ErrOldReport {
			logMessageInfo(consumer, msg, message, "Skipping because a more recent report already exists for this cluster")
			return message.RequestID, nil
		}

		logMessageError(consumer, msg, message, "Error writing report to database", err)
		return message.RequestID, err
	}

	// update metric - number of messages stored into database
	StoredMessages.Inc()

	// update metric - number of bytes stored into database
	// beware: counter value is represented as float64, not as bytes as you'd expect
	StoredBytes.Add(float64(len(shrunkAsBytes)))

	logMessageInfo(consumer, msg, message, "Stored")
	tStored := time.Now()

	// Step #7: print durations of all previous steps

	// log durations for every message consumption steps
	logDuration(tStart, tRead, msg.Offset, "Read duration")
	logDuration(tRead, tMarshalled, msg.Offset, "Marshalling duration")
	logDuration(tMarshalled, tShrunk, msg.Offset, "Shrinking duration")
	logDuration(tShrunk, tTimeCheck, msg.Offset, "Time check duration")
	logDuration(tTimeCheck, tStored, msg.Offset, "DB store duration")

	// message has been parsed and stored into storage
	return message.RequestID, nil
}

// logshrunkMessage function prints/logs information about status of
// shrinking the message.
func logShrunkMessage(reportAsBytes, shrunkAsBytes []byte) {
	orig := len(reportAsBytes)
	shrunk := len(shrunkAsBytes)
	percentage := 100.0 * shrunk / orig
	log.Info().
		Int("Original size", len(reportAsBytes)).
		Int("Shrunk size", len(shrunkAsBytes)).
		Int("Ratio (%)", percentage).
		Msg("Message shrunk")
}

// checkReportStructure function checks if the report has correct structure
func checkReportStructure(r Report) error {
	// the structure is not well defined yet, so all we should do is to check if all keys are there
	expectedKeys := []string{
		fingerprintsAttribute,
		reportsAttribute,
		systemAttribute,
	}

	// 'skips' key is now optional, we should not expect it anymore:
	// https://github.com/RedHatInsights/insights-results-aggregator/issues/1206
	// Simialrly, 'info'  key is now optional too.
	// https://github.com/RedHatInsights/insights-results-aggregator/pull/1996
	// expectedKeys := []string{"fingerprints", "info", "reports", "skips", "system"}

	// check if the structure contains all expected keys
	for _, expectedKey := range expectedKeys {
		_, found := r[expectedKey]
		if !found {
			return errors.New("Improper report structure, missing key " + expectedKey)
		}
	}
	return nil
}

// parseMessage function tries to parse incoming message and read all required
// attributes from it
func parseMessage(messageValue []byte) (IncomingMessage, error) {
	var deserialized IncomingMessage

	err := json.Unmarshal(messageValue, &deserialized)
	if err != nil {
		return deserialized, err
	}

	if deserialized.Organization == nil {
		return deserialized, errors.New("missing required attribute 'OrgID'")
	}
	if deserialized.AccountNumber == nil {
		return deserialized, errors.New("missing required attribute 'AccountNumber'")
	}
	if deserialized.ClusterName == nil {
		return deserialized, errors.New("missing required attribute 'ClusterName'")
	}
	if deserialized.Report == nil {
		return deserialized, errors.New("missing required attribute 'Report'")
	}

	_, err = uuid.Parse(string(*deserialized.ClusterName))

	if err != nil {
		return deserialized, errors.New("cluster name is not a UUID")
	}

	err = checkReportStructure(*deserialized.Report)
	if err != nil {
		log.Err(err).
			Msgf("Deserialized report read from message with improper structure: %v", *deserialized.Report)
		return deserialized, err
	}

	return deserialized, nil
}

// saramaConfigFromBrokerConfig function reads broker configuration and
// construct configuration compatible with Sarama library
func saramaConfigFromBrokerConfig(brokerConfiguration *BrokerConfiguration) (*sarama.Config, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = SaramaVersion

	// Set API version request behavior based on configuration
	// When DisableAPIVersionsRequest is true, skip API version negotiation
	// (useful for MockBroker testing or legacy brokers)
	saramaConfig.ApiVersionsRequest = !brokerConfiguration.DisableAPIVersionsRequest

	/* TODO: we need to do it in production code
	if brokerCfg.Timeout > 0 {
		saramaConfig.Net.DialTimeout = brokerConfiguration.Timeout
		saramaConfig.Net.ReadTimeout = brokerConfiguration.Timeout
		saramaConfig.Net.WriteTimeout = brokerConfiguration.Timeout
	}
	*/
	if strings.Contains(brokerConfiguration.SecurityProtocol, SSLProtocol) {
		saramaConfig.Net.TLS.Enable = true
	}
	if strings.EqualFold(brokerConfiguration.SecurityProtocol, SSLProtocol) && brokerConfiguration.CertPath != "" {
		tlsConfig, err := tlsutils.NewTLSConfig(brokerConfiguration.CertPath)
		if err != nil {
			log.Error().Msgf("Unable to load TLS config for %s cert", brokerConfiguration.CertPath)
			return nil, err
		}
		saramaConfig.Net.TLS.Config = tlsConfig
	} else if strings.HasPrefix(brokerConfiguration.SecurityProtocol, "SASL_") {
		log.Info().Msg("Configuring SASL authentication")
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = brokerConfiguration.SaslUsername
		saramaConfig.Net.SASL.Password = brokerConfiguration.SaslPassword
		saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(brokerConfiguration.SaslMechanism)

		if strings.EqualFold(brokerConfiguration.SaslMechanism, sarama.SASLTypeSCRAMSHA512) {
			log.Info().Msg("Configuring SCRAM-SHA512")
			saramaConfig.Net.SASL.Handshake = true
			saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &SCRAMClient{HashGeneratorFcn: sha512.New}
			}
		}
	}
	return saramaConfig, nil
}

func saramaProducerConfigFromBrokerConfig(brokerConfiguration *BrokerConfiguration) (*sarama.Config, error) {
	saramaConfig, err := saramaConfigFromBrokerConfig(brokerConfiguration)
	if err != nil {
		return nil, err
	}
	saramaConfig.Producer.Return.Successes = true
	return saramaConfig, nil
}
