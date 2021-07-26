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

package main

// This source file contains definition of data type named ConfigStruct that
// represents configuration of Notification Writer service. This source file
// also contains function named LoadConfiguration that can be used to load
// configuration from provided configuration file and/or from environment
// variables. Additionally several specific functions named
// GetStorageConfiguration, GetLoggingConfiguration, and GetBrokerConfiguration
// are to be used to return specific configuration options.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/config.html

// Default name of configuration file is config.toml
// It can be changed via environment variable NOTIFICATION_WRITER_CONFIG_FILE

// An example of configuration file that can be used in devel environment:
//
// [broker]
// address = "kafka:29092"
// topic = "ccx.ocp.results"
// group = "aggregator"
// enabled = true
//
// [storage]
// db_driver = "postgres"
// pg_username = "user"
// pg_password = "password"
// pg_host = "localhost"
// pg_port = 5432
// pg_db_name = "notification"
// pg_params = "sslmode=disable"
// log_sql_queries = true
//
// [logging]
// debug = true
// log_level = ""
//
// Environment variables that can be used to override configuration file settings:
// TBD

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"

	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	filenameAttribute               = "filename"
	parsingConfigurationFileMessage = "parsing configuration file"
)

// ConfigStruct is a structure holding the whole notification service
// configuration
type ConfigStruct struct {
	Broker  BrokerConfiguration  `mapstructure:"broker"  toml:"broker"`
	Storage StorageConfiguration `mapstructure:"storage" toml:"storage"`
	Logging LoggingConfiguration `mapstructure:"logging" toml:"logging"`
	Metrics MetricsConfiguration `mapstructure:"metrics" toml:"metrics"`
}

// MetricsConfiguration holds metrics related configuration
type MetricsConfiguration struct {
	Namespace string `mapstructure:"namespace" toml:"namespace"`
	Address   string `mapstructure:"address" toml:"address"`
}

// LoggingConfiguration represents configuration for logging in general
type LoggingConfiguration struct {
	// Debug enables pretty colored logging
	Debug bool `mapstructure:"debug" toml:"debug"`

	// LogLevel sets logging level to show. Possible values are:
	// "debug"
	// "info"
	// "warn", "warning"
	// "error"
	// "fatal"
	//
	// logging level won't be changed if value is not one of listed above
	LogLevel string `mapstructure:"log_level" toml:"log_level"`
}

// BrokerConfiguration represents configuration for the broker
type BrokerConfiguration struct {
	// Address represents Kafka address
	Address string `mapstructure:"address" toml:"address"`
	// Topic is name of Kafka topic
	Topic string `mapstructure:"topic" toml:"topic"`
	// Group is name of Kafka group
	Group string `mapstructure:"group" toml:"group"`
	// Enabled is set to true if Kafka consumer is to be enabled
	Enabled bool `mapstructure:"enabled" toml:"enabled"`
}

// StorageConfiguration represents configuration of data storage
type StorageConfiguration struct {
	Driver        string `mapstructure:"db_driver"       toml:"db_driver"`
	PGUsername    string `mapstructure:"pg_username"     toml:"pg_username"`
	PGPassword    string `mapstructure:"pg_password"     toml:"pg_password"`
	PGHost        string `mapstructure:"pg_host"         toml:"pg_host"`
	PGPort        int    `mapstructure:"pg_port"         toml:"pg_port"`
	PGDBName      string `mapstructure:"pg_db_name"      toml:"pg_db_name"`
	PGParams      string `mapstructure:"pg_params"       toml:"pg_params"`
	LogSQLQueries bool   `mapstructure:"log_sql_queries" toml:"log_sql_queries"`
}

// LoadConfiguration loads configuration from defaultConfigFile, file set in
// configFileEnvVariableName or from env
func LoadConfiguration(configFileEnvVariableName string, defaultConfigFile string) (ConfigStruct, error) {
	var config ConfigStruct

	// env. variable holding name of configuration file
	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		log.Info().Str(filenameAttribute, configFile).Msg(parsingConfigurationFileMessage)
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		log.Info().Str(filenameAttribute, defaultConfigFile).Msg(parsingConfigurationFileMessage)
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	// try to read the whole configuration
	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// If config file is not present (which might be correct in
		// some environment) we need to read configuration from
		// environment variables. The problem is that Viper is not
		// smart enough to understand the structure of config by
		// itself, so we need to read fake config file
		fakeTomlConfigWriter := new(bytes.Buffer)

		err := toml.NewEncoder(fakeTomlConfigWriter).Encode(config)
		if err != nil {
			return config, err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))
		if err != nil {
			return config, err
		}
	} else if err != nil {
		// error is processed on caller side
		return config, fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env

	const envPrefix = "CCX_NOTIFICATION_WRITER_"

	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	if clowder.IsClowderEnabled() {
		// can not use Zerolog at this moment!
		fmt.Println("Clowder is enabled")

		// TODO: insert logic to replace SELECTED configuration variables
	} else {
		// can not use Zerolog at this moment!
		fmt.Println("Clowder is disabled")
	}

	// everything's should be ok
	return config, nil
}

// GetStorageConfiguration returns storage configuration
func GetStorageConfiguration(config ConfigStruct) StorageConfiguration {
	return config.Storage
}

// GetLoggingConfiguration returns logging configuration
func GetLoggingConfiguration(config ConfigStruct) LoggingConfiguration {
	return config.Logging
}

// GetBrokerConfiguration returns broker configuration
func GetBrokerConfiguration(config ConfigStruct) BrokerConfiguration {
	return config.Broker
}

// GetMetricsConfiguration returns metrics configuration
func GetMetricsConfiguration(config ConfigStruct) MetricsConfiguration {
	return config.Metrics
}
