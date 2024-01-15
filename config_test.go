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

package main_test

// Unit test definitions for functions and methods defined in source file
// config.go
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/config_test.html

import (
	"fmt"
	"os"

	"testing"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// init function is called before tests
func init() {
	// set default logging level regardles of config made in code
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// mustLoadConfiguration function loads configuration file or the actual test
// will fail
func mustLoadConfiguration(envVar string) {
	_, err := main.LoadConfiguration(envVar, "tests/config1")
	if err != nil {
		panic(err)
	}
}

// mustSetEnv function set specified environment variable or the actual test
// will fail
func mustSetEnv(t *testing.T, key, val string) {
	err := os.Setenv(key, val)
	assert.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}
}

// TestLoadDefaultConfiguration test loads a configuration file for testing
// with check that load was correct
func TestLoadDefaultConfiguration(_ *testing.T) {
	os.Clearenv()
	mustLoadConfiguration("nonExistingEnvVar")
}

// TestLoadConfigurationFromEnvVariable tests loading the config. file for
// testing from an environment variable
func TestLoadConfigurationFromEnvVariable(t *testing.T) {
	os.Clearenv()

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	mustLoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE")
}

// TestLoadConfigurationNonEnvVarUnknownConfigFile tests loading an unexisting
// config file when no environment variable is provided
func TestLoadConfigurationNonEnvVarUnknownConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "foobar")
	assert.Nil(t, err)
}

// TestLoadConfigurationBadConfigFile tests loading an unexisting config file
// when no environment variable is provided
func TestLoadConfigurationBadConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "tests/config3")
	assert.Contains(t, err.Error(), `fatal error config file: While parsing config:`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests loading a
// non-existent configuration file set in environment
func TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig(t *testing.T) {
	os.Clearenv()

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests that if env
// var is provided, it must point to a valid config file
func TestLoadingConfigurationEnvVariableBadValueDefaultConfigFailure(t *testing.T) {
	os.Clearenv()

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config1")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadBrokerConfiguration tests loading the broker configuration sub-tree
func TestLoadBrokerConfiguration(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	brokerCfg := main.GetBrokerConfiguration(&config)

	assert.Equal(t, []string{"localhost:29092"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
}

// TestLoadStorageConfiguration tests loading the storage configuration
// subtree
func TestLoadStorageConfiguration(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	storageCfg := main.GetStorageConfiguration(&config)

	assert.Equal(t, "sqlite3", storageCfg.Driver)
	assert.Equal(t, "user", storageCfg.PGUsername)
	assert.Equal(t, "password", storageCfg.PGPassword)
	assert.Equal(t, "localhost", storageCfg.PGHost)
	assert.Equal(t, 5432, storageCfg.PGPort)
	assert.Equal(t, "notifications", storageCfg.PGDBName)
	assert.Equal(t, "", storageCfg.PGParams)
	assert.Equal(t, true, storageCfg.LogSQLQueries)
}

// TestLoadLoggingConfiguration tests loading the logging configuration sub-tree
func TestLoadLoggingConfiguration(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	loggingCfg := main.GetLoggingConfiguration(&config)

	assert.Equal(t, true, loggingCfg.Debug)
	assert.Equal(t, "", loggingCfg.LogLevel)
}

// TestLoadMetricsConfiguration tests loading the metrics configuration sub-tree
func TestLoadMetricsConfiguration(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	metricsCfg := main.GetMetricsConfiguration(&config)

	assert.Equal(t, "notification_writer", metricsCfg.Namespace)
	assert.Equal(t, ":8080", metricsCfg.Address)
}

// TestLoadConfigurationFromEnvVariableClowderEnabled tests loading the config.
// file for testing from an environment variable. Clowder config is enabled in
// this case.
func TestLoadConfigurationFromEnvVariableClowderEnabled(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	// explicit database configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// check if database configuration has been loaded properly
	dbCfg := main.GetStorageConfiguration(&config)
	assert.Equal(t, testDB, dbCfg.PGDBName)
}

// TestLoadConfigurationNoKafkaBroker test if number of configured brokers are
// tested properly when no broker config exists
func TestLoadConfigurationNoKafkaBroker(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{}, // no brokers in configuration
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	// check broker configuration
	assert.Equal(t, []string{"localhost:29092"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)
}

// TestLoadConfigurationKafkaBrokerEmptyConfig test if empty broker config is
// loaded via Clowder
func TestLoadConfigurationKafkaBrokerEmptyConfig(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	// just one empty broker configuration
	var brokersConfig = []clowder.BrokerConfig{
		{},
	}

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{
			Brokers: brokersConfig,
			Topics:  []clowder.TopicConfig{},
		},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	print(clowder.KafkaServers)
	print(clowder.LoadedConfig.Kafka)
	// check broker configuration
	assert.Equal(t, []string{""}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)
}

// TestLoadConfigurationKafkaBrokerNoPort test loading broker configuration w/o port
func TestLoadConfigurationKafkaBrokerNoPort(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	// just one non-empty broker configuration
	var brokersConfig = []clowder.BrokerConfig{
		{
			Hostname: "test",
			Port:     nil}, // port is not set
	}

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{
			Brokers: brokersConfig},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	// check broker configuration
	// no port should be set
	assert.Equal(t, []string{"test"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)
}

// TestLoadConfigurationKafkaBrokerPort test loading broker port
func TestLoadConfigurationKafkaBrokerPort(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	var port = 1234

	// just one non-empty broker configuration
	var brokersConfig []clowder.BrokerConfig = []clowder.BrokerConfig{
		{
			Hostname: "test",
			Port:     &port}, // port is set
	}

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{
			Brokers: brokersConfig},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	// check broker configuration
	assert.Equal(t, []string{"test:1234"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)
}

// TestLoadConfigurationKafkaBrokerAuthConfigMissingSASL test loading broker auth. config
// is correct but missing SASL configuration
func TestLoadConfigurationKafkaBrokerAuthConfigMissingSASL(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	var port = 1234
	var authType clowder.BrokerConfigAuthtype

	var brokersConfig []clowder.BrokerConfig = []clowder.BrokerConfig{
		{
			Hostname: "test",
			Port:     &port,
			Authtype: &authType,
		},
	}

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{
			Brokers: brokersConfig,
		},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	// check broker configuration
	assert.Equal(t, []string{"test:1234"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)

	// additionally SASL config should be set too
	assert.Equal(t, "", brokerCfg.SaslUsername)
	assert.Equal(t, "", brokerCfg.SaslPassword)
	assert.Equal(t, "", brokerCfg.SaslMechanism)
	assert.Equal(t, "", brokerCfg.SecurityProtocol)
}

// TestLoadConfigurationKafkaBrokerAuthConfig test loading broker auth. config
// is correct
func TestLoadConfigurationKafkaBrokerAuthConfig(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	var port = 1234
	var authType clowder.BrokerConfigAuthtype

	var username = "username"
	var password = "password"
	var saslMechanism = "mechanism"
	var securityProtocol = "security_protocol"

	var brokersConfig []clowder.BrokerConfig = []clowder.BrokerConfig{
		{
			Hostname:         "test",
			Port:             &port,
			Authtype:         &authType,
			SecurityProtocol: &securityProtocol,
			// proper SASL configuration
			Sasl: &clowder.KafkaSASLConfig{
				Username:      &username,
				Password:      &password,
				SaslMechanism: &saslMechanism},
		},
	}

	// explicit database and broker configuration
	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
		Kafka: &clowder.KafkaConfig{
			Brokers: brokersConfig,
		},
	}

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, "CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	// load configuration using Clowder config
	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	// retrieve broker configuration that was just loaded
	brokerCfg := main.GetBrokerConfiguration(&config)

	// check broker configuration
	assert.Equal(t, []string{"test:1234"}, brokerCfg.Addresses)
	assert.Equal(t, "ccx_test_notifications", brokerCfg.Topic)
	assert.Equal(t, "test-consumer-group", brokerCfg.Group)
	assert.True(t, brokerCfg.Enabled)

	// additionally SASL config should be set too
	assert.Equal(t, username, brokerCfg.SaslUsername)
	assert.Equal(t, password, brokerCfg.SaslPassword)
	assert.Equal(t, saslMechanism, brokerCfg.SaslMechanism)
	assert.Equal(t, securityProtocol, brokerCfg.SecurityProtocol)
}

// TestLoadConfigurationKafkaTopicUpdatedFromClowder tests that when applying the config,
// if the Clowder config is enabled, the Kafka topics are replaced by the ones defined in
// LoadedConfig.Kafka.Topics if found, and used as-is if not.
func TestLoadConfigurationKafkaTopicUpdatedFromClowder(t *testing.T) {
	os.Clearenv()
	hostname := "kafka"
	port := 9092
	topicName := "ccx_test_notifications"
	newTopicName := "the.clowder.kafka.topic"

	// explicit database and broker config
	clowder.LoadedConfig = &clowder.AppConfig{
		Kafka: &clowder.KafkaConfig{
			Brokers: []clowder.BrokerConfig{
				{
					Hostname: hostname,
					Port:     &port,
				},
			},
		},
	}

	clowder.KafkaTopics = make(map[string]clowder.TopicConfig)
	clowder.KafkaTopics[topicName] = clowder.TopicConfig{
		Name:          newTopicName,
		RequestedName: topicName,
	}

	// set environment variable that points to Clowder configuration file
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")

	config, err := main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	brokerCfg := main.GetBrokerConfiguration(&config)
	assert.Equal(t, []string{fmt.Sprintf("%s:%d", hostname, port)}, brokerCfg.Addresses)
	assert.Equal(t, newTopicName, brokerCfg.Topic)

	// config with different broker configuration, broker's hostname taken from clowder, but no topic to map to
	topicName = "test_notification_topic"

	config, err = main.LoadConfiguration("CCX_NOTIFICATION_WRITER_CONFIG_FILE", "tests/config1")
	assert.NoError(t, err, "Failed loading configuration file")

	brokerCfg = main.GetBrokerConfiguration(&config)
	assert.Equal(t, []string{fmt.Sprintf("%s:%d", hostname, port)}, brokerCfg.Addresses)
	assert.Equal(t, topicName, brokerCfg.Topic)
}

// TestGetStorageConfigurationIsImmutable checks if function
// GetStorageConfiguration is not mutable
func TestGetStorageConfigurationIsImmutable(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")

	// load configuration with check if loading was ok
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	// clone the configuration
	origConfig := config

	// call the tested function
	main.GetStorageConfiguration(&config)

	// and compare original configuration with possibly mutated one
	assert.Equal(t, config, origConfig, "GetStorageConfiguration must not be mutable")
}

// TestGetLoggingConfigurationIsImmutable checks if function
// GetLoggingConfiguration is not mutable
func TestGetLoggingConfigurationIsImmutable(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")

	// load configuration with check if loading was ok
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	// clone the configuration
	origConfig := config

	// call the tested function
	main.GetLoggingConfiguration(&config)

	// and compare original configuration with possibly mutated one
	assert.Equal(t, config, origConfig, "GetLoggingConfiguration must not be mutable")
}

// TestGetBrokerConfigurationIsImmutable checks if function
// GetBrokerConfiguration is not mutable
func TestGetBrokerConfigurationIsImmutable(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")

	// load configuration with check if loading was ok
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	// clone the configuration
	origConfig := config

	// call the tested function
	main.GetBrokerConfiguration(&config)

	// and compare original configuration with possibly mutated one
	assert.Equal(t, config, origConfig, "GetBrokerConfiguration must not be mutable")
}

// TestGetMetricsConfigurationIsImmutable checks if function
// GetMetricsConfiguration is not mutable
func TestGetMetricsConfigurationIsImmutable(t *testing.T) {
	envVar := "CCX_NOTIFICATION_WRITER_CONFIG_FILE"

	// set environment variable that points to config file
	// (without extension)
	mustSetEnv(t, envVar, "tests/config2")

	// load configuration with check if loading was ok
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	// clone the configuration
	origConfig := config

	// call the tested function
	main.GetMetricsConfiguration(&config)

	// and compare original configuration with possibly mutated one
	assert.Equal(t, config, origConfig, "GetMetricsConfiguration must not be mutable")
}
