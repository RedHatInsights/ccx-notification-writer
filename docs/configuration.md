---
layout: page
nav_order: 3
---

# Configuration

Configuration is done by toml config, default one is `config.toml` in working directory,
but it can be overwritten by `NOTIFICATION_WRITER_CONFIG_FILE` env var.

Also each key in config can be overwritten by corresponding env var. For example if you have config

```toml
[broker]
address = "localhost:9092"
security_protocol = "PLAINTEXT"
sasl_mechanism = "not-used"
sasl_username = "not-used"
sasl_password = "not-used"
topic = "ccx.ocp.results"
group = "test-consumer-group"
enabled = true

[storage]
db_driver = "postgres"
pg_username = "postgres"
pg_password = "postgres"
pg_host = "localhost"
pg_port = 5432
pg_db_name = "notification"
pg_params = "sslmode=disable"
log_sql_queries = true

[logging]
debug = true
log_level = ""

[metrics]
namespace = "notification_writer"
address = ":8080"
```

and environment variables

```shell
NOTIFICATION_WRITER__STORAGE__DB_DRIVER="postgres"
NOTIFICATION_WRITER__STORAGE__PG_PASSWORD="your secret password"
```

the actual driver will be postgres with password "your secret password"

It's very useful for deploying docker containers and keeping some of your configuration
outside of main config file(like passwords).

## Environment variables

List of all environment variables that can be used to override configuration settings stored in config file:

```
CCX_NOTIFICATION_WRITER__BROKER__ADDRESS
CCX_NOTIFICATION_WRITER__BROKER__SECURITY_PROTOCOL
CCX_NOTIFICATION_WRITER__BROKER__SASL_MECHANISM
CCX_NOTIFICATION_WRITER__BROKER__SASL_USERNAME
CCX_NOTIFICATION_WRITER__BROKER__SASL_PASSWORD
CCX_NOTIFICATION_WRITER__BROKER__TOPIC
CCX_NOTIFICATION_WRITER__BROKER__GROUP
CCX_NOTIFICATION_WRITER__BROKER__ENABLED
CCX_NOTIFICATION_WRITER__STORAGE__DB_DRIVER
CCX_NOTIFICATION_WRITER__STORAGE__PG_USERNAME
CCX_NOTIFICATION_WRITER__STORAGE__PG_PASSWORD
CCX_NOTIFICATION_WRITER__STORAGE__PG_HOST
CCX_NOTIFICATION_WRITER__STORAGE__PG_PORT
CCX_NOTIFICATION_WRITER__STORAGE__PG_DB_NAME
CCX_NOTIFICATION_WRITER__STORAGE__PG_PARAMS
CCX_NOTIFICATION_WRITER__STORAGE__LOG_SQL_QUERIES
CCX_NOTIFICATION_WRITER__LOGGING__DEBUG
CCX_NOTIFICATION_WRITER__LOGGING__LOG_LEVEL
CCX_NOTIFICATION_WRITER__METRICS__NAMESPACE
CCX_NOTIFICATION_WRITER__METRICS__ADDRESS
```

### Clowder configuration

In Clowder environment, some configuration options are injected automatically.
Currently Kafka broker configuration is injected this side. To test this
behavior, it is possible to specify path to Clowder-related configuration file
via `AGG_CONFIG` environment variable:

```
export ACG_CONFIG="clowder_config.json"
```

An example of Clowder configuration (please note that URLs and other attributes are not real ones):

```json
{
  "BOPURL": "http://ephemeral_service_url:1234",
  "database": {
    "adminPassword": "admin password",
    "adminUsername": "admin username",
    "hostname": "host name of machine with DB",
    "name": "ccx-notification-db",
    "username": "username to connect to database",
    "password": "password to connect to database",
    "port": 5432,
    "sslMode": "disable"
  },
  "endpoints": [
    {
      "app": "notifications-backend",
      "hostname": "notifications-backend-service URL",
      "name": "service",
      "port": 8000
    },
    {
      "app": "notifications-engine",
      "hostname": "notifications-engine-service URL",
      "name": "service",
      "port": 8000
    }
  ],
  "kafka": {
    "brokers": [
      {
        "hostname": "hostname of machine with running Kafka broker",
        "port": 9092
      }
    ],
    "topics": [
      {
        "name": "ccx.ocp.results",
        "requestedName": "ccx.ocp.results"
      }
    ]
  },
  "logging": {
    "cloudwatch": {
      "accessKeyId": "",
      "logGroup": "",
      "region": "",
      "secretAccessKey": ""
    },
    "type": "null"
  },
  "metricsPath": "/metrics",
  "metricsPort": 9000,
  "privatePort": 10000,
  "publicPort": 8000,
  "webPort": 8000
}
```
