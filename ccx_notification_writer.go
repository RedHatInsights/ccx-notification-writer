/*
Copyright © 2021, 2022 Red Hat, Inc.

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

// Entry point to the notification writer service.
//
// The service contains consumer (usually Kafka consumer) that consumes
// messages from given source, processes those messages and stores them
// in configured data store.
package main

// Entry point to the CCX Notification writer service

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/ccx_notification_writer.html

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	utils "github.com/RedHatInsights/insights-operator-utils/migrations"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Shopify/sarama"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Messages
const (
	versionMessage                                          = "CCX Notification Writer version 1.0"
	authorsMessage                                          = "Pavel Tisnovsky, Red Hat Inc."
	connectionToBrokerMessage                               = "Connection to broker"
	operationFailedMessage                                  = "Operation failed"
	notConnectedToBrokerMessage                             = "Not connected to broker"
	brokerConnectionSuccessMessage                          = "Broker connection OK"
	databaseCleanupOperationFailedMessage                   = "Database cleanup operation failed"
	databaseDropTablesOperationFailedMessage                = "Database drop tables operation failed"
	databasePrintNewReportsForCleanupOperationFailedMessage = "Print records from `new_reports` table prepared for cleanup failed"
	databasePrintOldReportsForCleanupOperationFailedMessage = "Print records from `reported` table prepared for cleanup failed"
	databaseCleanupNewReportsOperationFailedMessage         = "Cleanup records from `new_reports` table failed"
	databaseCleanupOldReportsOperationFailedMessage         = "Cleanup records from `reported` table failed"
	rowsInsertedMessage                                     = "Rows inserted"
	rowsDeletedMessage                                      = "Rows deleted"
	rowsAffectedMessage                                     = "Rows affected"
	brokerAddress                                           = "Broker address"
	StorageHandleErr                                        = "unable to get storage handle"
)

// Configuration-related constants
const (
	configFileEnvVariableName = "CCX_NOTIFICATION_WRITER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// Exit codes
const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota
	// ExitStatusConsumerError is returned in case of any consumer-related error
	ExitStatusConsumerError
	// ExitStatusKafkaError is returned in case of any Kafka-related error
	ExitStatusKafkaError
	// ExitStatusStorageError is returned in case of any consumer-related error
	ExitStatusStorageError
	// ExitStatusHTTPServerError is returned in case the HTTP server can not be started
	ExitStatusHTTPServerError
	// ExitStatusMigrationError is returned in case of an error while attempting to perform DB migrations
	ExitStatusMigrationError
)

// showVersion function displays version information.
func showVersion() {
	fmt.Println(versionMessage)
}

// showAuthors function displays information about authors.
func showAuthors() {
	fmt.Println(authorsMessage)
}

// showConfiguration function displays actual configuration.
func showConfiguration(configuration ConfigStruct) {
	brokerConfig := GetBrokerConfiguration(configuration)
	log.Info().
		Str(brokerAddress, brokerConfig.Address).
		Str("Security protocol", brokerConfig.SecurityProtocol).
		Str("Cert path", brokerConfig.CertPath).
		Str("Sasl mechanism", brokerConfig.SaslMechanism).
		Str("Sasl username", brokerConfig.SaslUsername).
		Str("Topic", brokerConfig.Topic).
		Str("Group", brokerConfig.Group).
		Bool("Enabled", brokerConfig.Enabled).
		Msg("Broker configuration")

	storageConfig := GetStorageConfiguration(configuration)
	log.Info().
		Str("Driver", storageConfig.Driver).
		Str("DB Name", storageConfig.PGDBName).
		Str("Username", storageConfig.PGUsername). // password is omitted on purpose
		Str("Host", storageConfig.PGHost).
		Int("Port", storageConfig.PGPort).
		Bool("LogSQLQueries", storageConfig.LogSQLQueries).
		Msg("Storage configuration")

	loggingConfig := GetLoggingConfiguration(configuration)
	log.Info().
		Str("Level", loggingConfig.LogLevel).
		Bool("Pretty colored debug logging", loggingConfig.Debug).
		Msg("Logging configuration")

	metricsConfig := GetMetricsConfiguration(configuration)
	log.Info().
		Str("Namespace", metricsConfig.Namespace).
		Str("Address", metricsConfig.Address).
		Msg("Metrics configuration")
}

// tryToConnectToKafka function just tries connection to Kafka broker
func tryToConnectToKafka(configuration ConfigStruct) (int, error) {
	log.Info().Msg("Checking connection to Kafka")

	// prepare broker configuration
	brokerConfiguration := GetBrokerConfiguration(configuration)

	log.Info().Str("broker address", brokerConfiguration.Address).Msg(brokerAddress)

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
func performDatabaseInitialization(configuration ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
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

// performDatabaseInitMigration function initialize migration table
func performDatabaseInitMigration(configuration ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.DatabaseInitMigration()
	if err != nil {
		log.Err(err).Msg("Database migration initialization operation failed")
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// performDatabaseCleanup function performs database cleanup - deletes content
// of all tables in database.
func performDatabaseCleanup(configuration ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
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
func performDatabaseDropTables(configuration ConfigStruct) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
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

// printNewReportsForCleanup function print all reports for `new_reports` table
// that are older than specified max age.
func printNewReportsForCleanup(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Error().Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.PrintNewReportsForCleanup(cliFlags.MaxAge)
	if err != nil {
		log.Error().Err(err).Msg(databasePrintNewReportsForCleanupOperationFailedMessage)
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// performNewReportsCleanup function deletes all reports from `new_reports`
// table that are older than specified max age.
func performNewReportsCleanup(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Error().Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	affected, err := storage.CleanupNewReports(cliFlags.MaxAge)
	if err != nil {
		log.Error().Err(err).Msg(databaseCleanupNewReportsOperationFailedMessage)
		return ExitStatusStorageError, err
	}
	log.Info().Int(rowsDeletedMessage, affected).Msg("Cleanup `new_reports` finished")

	return ExitStatusOK, nil
}

// printOldReportsForCleanup function print all reports for `reported` table
// that are older than specified max age.
func printOldReportsForCleanup(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Error().Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	err = storage.PrintOldReportsForCleanup(cliFlags.MaxAge)
	if err != nil {
		log.Error().Err(err).Msg(databasePrintOldReportsForCleanupOperationFailedMessage)
		return ExitStatusStorageError, err
	}

	return ExitStatusOK, nil
}

// performOldReportsCleanup function deletes all reports from `reported` table
// that are older than specified max age.
func performOldReportsCleanup(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(storageConfiguration)
	if err != nil {
		log.Error().Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	affected, err := storage.CleanupOldReports(cliFlags.MaxAge)
	if err != nil {
		log.Error().Err(err).Msg(databaseCleanupOldReportsOperationFailedMessage)
		return ExitStatusStorageError, err
	}
	log.Info().Int(rowsDeletedMessage, affected).Msg("Cleanup `reported` finished")

	return ExitStatusOK, nil
}

// startService function tries to start the notification writer service.
func startService(configuration ConfigStruct) (int, error) {
	// show configuration at startup
	showConfiguration(configuration)

	// configure metrics
	metricsConfig := GetMetricsConfiguration(configuration)
	if metricsConfig.Namespace != "" {
		log.Info().Str("namespace", metricsConfig.Namespace).Msg("Setting metrics namespace")
		AddMetricsWithNamespace(metricsConfig.Namespace)
	}

	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
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
	brokerConfiguration := GetBrokerConfiguration(configuration)

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
func startConsumer(brokerConfiguration BrokerConfiguration, storage Storage) error {
	consumer, err := NewConsumer(brokerConfiguration, storage)
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
		err := http.ListenAndServe(address, nil) // #nosec G114
		if err != nil {
			log.Error().Err(err).Msg("Listen and serve")
		}
	}()
	return nil
}

// doSelectedOperation function perform operation selected on command line.
// When no operation is specified, the Notification writer service is started
// instead.
//
//gocyclo:ignore
func doSelectedOperation(configuration ConfigStruct, cliFlags CliFlags) (int, error) {
	switch {
	case cliFlags.ShowVersion:
		showVersion()
		return ExitStatusOK, nil
	case cliFlags.ShowAuthors:
		showAuthors()
		return ExitStatusOK, nil
	case cliFlags.ShowConfiguration:
		showConfiguration(configuration)
		return ExitStatusOK, nil
	case cliFlags.CheckConnectionToKafka:
		return tryToConnectToKafka(configuration)
	case cliFlags.PerformDatabaseInitialization:
		return performDatabaseInitialization(configuration)
	case cliFlags.PerformDatabaseCleanup:
		return performDatabaseCleanup(configuration)
	case cliFlags.PerformDatabaseDropTables:
		return performDatabaseDropTables(configuration)
	case cliFlags.PerformDatabaseInitMigration:
		return performDatabaseInitMigration(configuration)
	case cliFlags.PrintNewReportsForCleanup:
		return printNewReportsForCleanup(configuration, cliFlags)
	case cliFlags.PerformNewReportsCleanup:
		return performNewReportsCleanup(configuration, cliFlags)
	case cliFlags.PrintOldReportsForCleanup:
		return printOldReportsForCleanup(configuration, cliFlags)
	case cliFlags.PerformOldReportsCleanup:
		return performOldReportsCleanup(configuration, cliFlags)
	case cliFlags.MigrationInfo:
		return PrintMigrationInfo(configuration)
	case cliFlags.PerformMigrations != "":
		return PerformMigrations(configuration, cliFlags.PerformMigrations)
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
	flag.BoolVar(&cliFlags.PerformDatabaseInitialization, "db-init", false, "perform database initialization")
	flag.BoolVar(&cliFlags.PerformDatabaseCleanup, "db-cleanup", false, "perform database cleanup")
	flag.BoolVar(&cliFlags.PerformDatabaseDropTables, "db-drop-tables", false, "drop all tables from database")
	flag.BoolVar(&cliFlags.PerformDatabaseInitMigration, "db-init-migration", false, "initialize migration")
	flag.BoolVar(&cliFlags.CheckConnectionToKafka, "check-kafka", false, "check connection to Kafka")
	flag.BoolVar(&cliFlags.ShowVersion, "version", false, "show version")
	flag.BoolVar(&cliFlags.ShowAuthors, "authors", false, "show authors")
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.PrintNewReportsForCleanup, "print-new-reports-for-cleanup", false, "print new reports to be cleaned up")
	flag.BoolVar(&cliFlags.PerformNewReportsCleanup, "new-reports-cleanup", false, "perform new reports clean up")
	flag.BoolVar(&cliFlags.PrintOldReportsForCleanup, "print-old-reports-for-cleanup", false, "print old reports to be cleaned up")
	flag.BoolVar(&cliFlags.PerformOldReportsCleanup, "old-reports-cleanup", false, "perform old reports clean up")
	flag.BoolVar(&cliFlags.MigrationInfo, "migration-info", false, "prints migration info")
	flag.StringVar(&cliFlags.MaxAge, "max-age", "", "max age for displaying/cleaning old records")
	flag.StringVar(&cliFlags.PerformMigrations, "migrate", "", "set database version")
	flag.Parse()

	// configuration has exactly the same structure as *.toml file
	configuration, err := LoadConfiguration(configFileEnvVariableName, defaultConfigFileName)
	if err != nil {
		log.Err(err).Msg("Load configuration")
	}

	if configuration.Logging.Debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Debug().Msg("Started")

	// override default value read from configuration file
	if cliFlags.MaxAge == "" {
		cliFlags.MaxAge = "7 days"
	}

	// perform selected operation
	exitStatus, err := doSelectedOperation(configuration, cliFlags)
	if err != nil {
		log.Err(err).Msg("Do selected operation")
		os.Exit(exitStatus)
		return
	}

	log.Debug().Msg("Finished")
}

// PrintMigrationInfo function prints information about current DB migration
// version without making any modifications.
func PrintMigrationInfo(configuration ConfigStruct) (int, error) {
	storage, err := NewStorage(configuration.Storage)
	if err != nil {
		log.Error().Err(err).Msg(StorageHandleErr)
		return ExitStatusMigrationError, err
	}
	currMigVer, err := utils.GetDBVersion(storage.connection)
	if err != nil {
		log.Error().Err(err).Msg("Unable to get current DB version")
		return ExitStatusMigrationError, err
	}

	log.Info().Msgf("Current DB version: %d", currMigVer)
	log.Info().Msgf("Maximum available version: %d", utils.GetMaxVersion())
	return ExitStatusOK, nil
}

// PerformMigrations migrates the database to the version
// specified in params
func PerformMigrations(configuration ConfigStruct, migParam string) (exitStatus int, err error) {
	// init migration utils
	utils.Set(All())

	// get db handle
	storage, err := NewStorage(configuration.Storage)
	if err != nil {
		log.Error().Err(err).Msg(StorageHandleErr)
		exitStatus = ExitStatusMigrationError
		return
	}

	// parse migration params
	var desiredVersion utils.Version
	if migParam == "latest" {
		desiredVersion = utils.GetMaxVersion()
	} else {
		vers, convErr := strconv.Atoi(migParam)
		if err != nil {
			log.Error().Err(err).Msgf("Unable to parse migration version: %v", migParam)
			exitStatus = ExitStatusMigrationError
			err = convErr
			return
		}
		desiredVersion = utils.Version(vers)
	}

	err = Migrate(storage.Connection(), desiredVersion)
	if err != nil {
		log.Error().Err(err).Msg("migration failure")
		exitStatus = ExitStatusMigrationError
		return
	}

	return
}
