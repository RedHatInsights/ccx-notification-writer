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

### Clowder configuration

In Clowder environment, some configuration options are injected automatically.
Currently Kafka broker configuration is injected this side. To test this
behavior, it is possible to specify path to Clowder-related configuration file
via `AGG_CONFIG` environment variable:

```
export ACG_CONFIG="clowder_config.json"
```
