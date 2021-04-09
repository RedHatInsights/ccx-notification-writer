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
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Messages
const (
	versionMessage = "Notification writer version 1.0"
	authorsMessage = "Pavel Tisnovsky, Red Hat Inc."
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
	// ExitStatusStorageError is returned in case of any consumer-related error
	ExitStatusStorageError
)

// showVersion function displays version information.
func showVersion() {
	fmt.Println(versionMessage)
}

// showAuthors function displays information about authors.
func showAuthors() {
	fmt.Println(authorsMessage)
}

// startService function tries to start the notification writer service.
func startService(config ConfigStruct) int {
	brokerConf := GetBrokerConfiguration(config)

	// if broker is disabled, simply don't start it
	if brokerConf.Enabled {
		err := startConsumer(brokerConf)
		if err != nil {
			log.Error().Err(err)
			return ExitStatusConsumerError
		}
	} else {
		log.Info().Msg("Broker is disabled, not starting it")
	}

	return ExitStatusOK
}

// startConsumer function starts the Kafka consumer.
func startConsumer(config BrokerConfiguration) error {
	consumer, err := NewConsumer(config)
	if err != nil {
		log.Error().Err(err).Msg("Construct broker")
		return err
	}
	consumer.Serve()
	return nil
}

// doSelectedOperation function perform operation selected on command line.
// When no operation is specified, the Notification writer service is started
// instead.
func doSelectedOperation(configuration ConfigStruct, cliFlags CliFlags) error {
	switch {
	case cliFlags.showVersion:
		showVersion()
		return nil
	case cliFlags.showAuthors:
		showAuthors()
		return nil
	default:
		brokerConfiguration := GetBrokerConfiguration(configuration)
		err := startConsumer(brokerConfiguration)
		return err
	}
	return nil
}

// main function is entry point to the Notification writer service.
func main() {
	var cliFlags CliFlags

	// define and parse all command line options
	flag.BoolVar(&cliFlags.performDatabaseInitialization, "db-init", false, "perform database initialization")
	flag.BoolVar(&cliFlags.performDatabaseCleanup, "db-clenaup", false, "perform database cleanup")
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

	storage, err := NewStorage(GetStorageConfiguration(config))
	fmt.Println(storage)
	if err != nil {
		log.Err(err).Msg("Operation failed")
	}

	// perform selected operation
	err = doSelectedOperation(config, cliFlags)
	if err != nil {
		log.Err(err).Msg("Operation failed")
	}

	log.Debug().Msg("Finished")
}
