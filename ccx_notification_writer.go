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

// Entry point to the notification writer service.
//
// The service contains consumer (usually Kafka consumer) that consumes
// messages from given source, processes those messages and stores them
// in configured data store.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Shopify/sarama"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Messages
const (
	versionMessage                           = "Notification writer version 1.0"
	authorsMessage                           = "Pavel Tisnovsky, Red Hat Inc."
	connectionToBrokerMessage                = "Connection to broker"
	operationFailedMessage                   = "Operation failed"
	notConnectedToBrokerMessage              = "Not connected to broker"
	brokerConnectionSuccessMessage           = "Broker connection OK"
	databaseCleanupOperationFailedMessage    = "Database cleanup operation failed"
	databaseDropTablesOperationFailedMessage = "Database drop tables operation failed"
)

// Configuration-related constants
const (
	configFileEnvVariableName = "NOTIFICATION_SERVICE_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// Exit codes
const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota
	// ExitStatusError is a general error code
	ExitStatusError
	// ExitStatusConsumerError is returned in case of any consumer-related error
	ExitStatusConsumerError
	// ExitStatusKafkaError is returned in case of any Kafka-related error
	ExitStatusKafkaError
	// ExitStatusStorageError is returned in case of any consumer-related error
	ExitStatusStorageError
	// ExitStatusHTTPServerError is returned in case the HTTP server can not be started
	ExitStatusHTTPServerError
)

// showVersion function displays version information.
func showVersion() {
	fmt.Println(versionMessage)
}

// showAuthors function displays information about authors.
func showAuthors() {
	fmt.Println(authorsMessage)
}

// tryToConnectToKafka function just tries connection to Kafka broker
func tryToConnectToKafka(config ConfigStruct) (int, error) {
	log.Info().Msg("Checking connection to Kafka")

	// prepare broker configuration
	brokerConfiguration := GetBrokerConfiguration(config)

	log.Info().Str("broker address", brokerConfiguration.Address).Msg("Broker address")

	// create new broker instance (w/o any checks)
	broker := sarama.NewBroker(brokerConfiguration.Address)

	// check broker connection
	err := broker.Open(nil)
	if err != nil {
		log.Error().Err(err).Msg(connectionToBrokerMessage)
		return ExitStatusKafkaError, err
	}

	// check if connection remain
	connected, err := broker.Connected()
	if err != nil {
		log.Error().Err(err).Msg(connectionToBrokerMessage)
		return ExitStatusKafkaError, err
	}
	if !connected {
		log.Error().Err(err).Msg(notConnectedToBrokerMessage)
		return ExitStatusConsumerError, err
	}

	log.Info().Msg(brokerConnectionSuccessMessage)

	// everything seems to be ok
	return ExitStatusOK, nil
}

// performDatabaseInitialization function performs database initialization -
// creates all tables in database.
func performDatabaseInitialization(config ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(config)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.DatabaseInitialization()
	if err != nil {
		log.Err(err).Msg("Database initialization operation failed")
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// performDatabaseCleanup function performs database cleanup - deletes content
// of all tables in database.
func performDatabaseCleanup(config ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(config)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.DatabaseCleanup()
	if err != nil {
		log.Err(err).Msg(databaseCleanupOperationFailedMessage)
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// performDatabaseDropTables function performs drop of all databases tables.
func performDatabaseDropTables(config ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(config)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.DatabaseDropTables()
	if err != nil {
		log.Err(err).Msg(databaseDropTablesOperationFailedMessage)
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// startService function tries to start the notification writer service.
func startService(config ConfigStruct) (int, error) {
	// configure metrics
	metricsConfig := GetMetricsConfiguration(config)
	if metricsConfig.Namespace != "" {
		log.Info().Str("namespace", metricsConfig.Namespace).Msg("Setting metrics namespace")
		AddMetricsWithNamespace(metricsConfig.Namespace)
	}

	// prepare the storage
	storageConfiguration := GetStorageConfiguration(config)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	// prepare HTTP server with metrics exposed
	err = startHTTPServer(metricsConfig.Address)
	if err != nil {
		log.Error().Err(err)
		return ExitStatusHTTPServerError, err
	}

	// prepare broker
	brokerConfiguration := GetBrokerConfiguration(config)

	// if broker is disabled, simply don't start it
	if brokerConfiguration.Enabled {
		err := startConsumer(brokerConfiguration, storage)
		if err != nil {
			log.Error().Err(err)
			return ExitStatusConsumerError, err
		}
	} else {
		log.Info().Msg("Broker is disabled, not starting it")
	}

	return ExitStatusOK, nil
}

// startConsumer function starts the Kafka consumer.
func startConsumer(config BrokerConfiguration, storage Storage) error {
	consumer, err := NewConsumer(config, storage)
	if err != nil {
		log.Error().Err(err).Msg("Construct broker failed")
		return err
	}
	consumer.Serve()
	return nil
}

// startHTTP server starts server with exposed metrics
func startHTTPServer(address string) error {
	// setup handlers
	http.Handle("/metrics", promhttp.Handler())

	// start the server
	go func() {
		log.Info().Str("HTTP server address", address).Msg("Starting HTTP server")
		err := http.ListenAndServe(address, nil)
		if err != nil {
			log.Error().Err(err).Msg("Listen and serve")
		}
	}()
	return nil
}

// doSelectedOperation function perform operation selected on command line.
// When no operation is specified, the Notification writer service is started
// instead.
func doSelectedOperation(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	switch {
	case cliFlags.showVersion:
		showVersion()
		return ExitStatusOK, nil
	case cliFlags.showAuthors:
		showAuthors()
		return ExitStatusOK, nil
	case cliFlags.checkConnectionToKafka:
		return tryToConnectToKafka(configuration)
	case cliFlags.performDatabaseInitialization:
		return performDatabaseInitialization(configuration)
	case cliFlags.performDatabaseCleanup:
		return performDatabaseCleanup(configuration)
	case cliFlags.performDatabaseDropTables:
		return performDatabaseDropTables(configuration)
	default:
		exitCode, err := startService(configuration)
		return exitCode, err
	}
	// this can not happen: return ExitStatusOK, nil
}

// main function is entry point to the Notification writer service.
func main() {
	var cliFlags CliFlags

	// define and parse all command line options
	flag.BoolVar(&cliFlags.performDatabaseInitialization, "db-init", false, "perform database initialization")
	flag.BoolVar(&cliFlags.performDatabaseCleanup, "db-cleanup", false, "perform database cleanup")
	flag.BoolVar(&cliFlags.performDatabaseDropTables, "db-drop-tables", false, "drop all tables from database")
	flag.BoolVar(&cliFlags.checkConnectionToKafka, "check-kafka", false, "check connection to Kafka")
	flag.BoolVar(&cliFlags.showVersion, "version", false, "show version")
	flag.BoolVar(&cliFlags.showAuthors, "authors", false, "show authors")
	flag.BoolVar(&cliFlags.showConfiguration, "show-configuration", false, "show configuration")
	flag.Parse()

	// config has exactly the same structure as *.toml file
	config, err := LoadConfiguration(configFileEnvVariableName, defaultConfigFileName)
	if err != nil {
		log.Err(err).Msg("Load configuration")
	}

	if config.Logging.Debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Debug().Msg("Started")

	// perform selected operation
	exitStatus, err := doSelectedOperation(config, cliFlags)
	if err != nil {
		log.Err(err).Msg("Do selected operation")
		os.Exit(exitStatus)
		return
	}

	log.Debug().Msg("Finished")
}
