# ccx-notification-writer
CCX Notification Writer service

[![GoDoc](https://godoc.org/github.com/RedHatInsights/ccx-notification-writer?status.svg)](https://godoc.org/github.com/RedHatInsights/ccx-notification-writer)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/ccx-notification-writer/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/ccx-notification-writer)](https://goreportcard.com/report/github.com/RedHatInsights/ccx-notification-writer)
[![Build Status](https://ci.ext.devshift.net/buildStatus/icon?job=RedHatInsights-ccx-notification-writer-gh-build-master)](https://ci.ext.devshift.net/job/RedHatInsights-ccx-notification-writer-gh-build-master/)
[![Build Status](https://travis-ci.com/RedHatInsights/ccx-notification-writer.svg?branch=master)](https://travis-ci.com/RedHatInsights/ccx-notification-writer)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/RedHatInsights/ccx-notification-writer)
[![License](https://img.shields.io/badge/license-Apache-blue)](https://github.com/RedHatInsights/ccx-notification-writer/blob/master/LICENSE)

<!-- vim-markdown-toc GFM -->

* [Description](#description)
    * [Architecture](#architecture)
* [Building](#building)
    * [Makefile targets](#makefile-targets)
* [Configuration](#configuration)
* [Usage](#usage)
    * [All command line options](#all-command-line-options)
    * [Starting the service](#starting-the-service)
    * [Cleanup old records](#cleanup-old-records)
* [Metrics](#metrics)
    * [Exposed metrics](#exposed-metrics)
    * [Retriewing metrics](#retriewing-metrics)
* [Database](#database)
    * [Migrations](#migrations)
    * [Database schema](#database-schema)
    * [Check PostgreSQL status](#check-postgresql-status)
    * [Start PostgreSQL database](#start-postgresql-database)
    * [Login into the database](#login-into-the-database)
* [Testing](#testing)
* [Package manifest](#package-manifest)

<!-- vim-markdown-toc -->

## Description

The main task for this service is to listen to configured Kafka topic, consume
all messages from such topic, and write OCP results (in JSON format) with
additional information (like organization ID, cluster name, Kafka offset etc.)
into database table named `new_reports`. Multiple reports can be consumed and
written into the database for the same cluster, because the primary (compound)
key for `new_reports` table is set to the combination `(org_id, cluster,
updated_at)`. When some message does not conform to expected schema (for
example if `org_id` is missing), such message is dropped and the error is
stored into log.

Additionally this service exposes several metrics about consumed and processed
messages. These metrics can be aggregated by Prometheus and displayed by
Grafana tools.

### Architecture

Overall architecture and integration of this service into the whole pipeline is
described [in this document](https://redhatinsights.github.io/ccx-notification-writer/)

## Building

Use `make build` to build executable file with this service.

### Makefile targets

All Makefile targets:

```
Usage: make <OPTIONS> ... <TARGETS>

Available targets are:

clean                Run go clean
build                Keep this rule for compatibility
fmt                  Run go fmt -w for all sources
lint                 Run golint
vet                  Run go vet. Report likely mistakes in source code
cyclo                Run gocyclo
ineffassign          Run ineffassign checker
shellcheck           Run shellcheck
errcheck             Run errcheck
goconst              Run goconst checker
gosec                Run gosec checker
abcgo                Run ABC metrics checker
style                Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts
run                  Build the project and executes the binary
test                 Run the unit tests
profiler             Run the unit tests with profiler enabled
cover                Generate HTML pages with code coverage
coverage             Display code coverage on terminal
bdd_tests            Run BDD tests
before_commit        Checks done before commit
help                 Show this help screen
```

## Configuration

Configuration is described
[in this document](https://redhatinsights.github.io/ccx-notification-writer/configuration.html)

## Usage

Provided a valid configuration, you can start the service with `./ccx-notification-writer`

### All command line options

List of all available command line options:

```
  -authors
        show authors
  -check-kafka
        check connection to Kafka
  -db-cleanup
        perform database cleanup
  -db-drop-tables
        drop all tables from database
  -db-init
        perform database initialization
  -db-init-migration
        initialize migration
  -max-age string
        max age for displaying/cleaning old records
  -migrate string
        set database version
  -migration-info
        prints migration info
  -new-reports-cleanup
        perform new reports clean up
  -old-reports-cleanup
        perform old reports clean up
  -print-new-reports-for-cleanup
        print new reports to be cleaned up
  -print-old-reports-for-cleanup
        print old reports to be cleaned up
  -show-configuration
        show configuration
  -version
        show version
```

### Starting the service

In order to start the service, just `./ccx-notification-writer` is needed to be called from CLI.

### Cleanup old records

It is possible to cleanup old records from `new_reports` and `reported` tables. To do it, use the following CLI options:

```
./ccx-notification-writer -old-reports-cleanup --max-age="30 days"
```

to perform cleanup of `reported` table.

It is also possible to use following command

```
./ccx-notification-writer -new-reports-cleanup --max-age="30 days"
```

to perform cleanup of `new_reports` table.

Additionally it is possible to just display old reports without touching the database tables:

```
./ccx-notification-writer -print-old-reports-for-cleanup --max-age="30 days"
```

or in case of new reports:

```
./ccx-notification-writer -print-new-reports-for-cleanup --max-age="30 days"
```


## Metrics

It is possible to use `/metrics` REST API endpoint to read all metrics
exposed to Prometheus or to any tool that is compatible with it. Currently,
the following metrics are exposed:

### Exposed metrics

* `notification_writer_check_last_checked_timestamp`
    - The total number of messages with last checked timestamp
* `notification_writer_check_schema_version`
    - The total number of messages with successfull schema check
* `notification_writer_consumed_messages`
    - The total number of messages consumed from Kafka
* `notification_writer_consuming_errors`
    - The total number of errors during consuming messages from Kafka
* `notification_writer_marshal_report`
    - The total number of marshaled reports
* `notification_writer_parse_incoming_message`
    - The total number of parsed messages
* `notification_writer_shrink_report`
    - The total number of shrunk reports
* `notification_writer_stored_messages`
    - The total number of messages stored into database
* `notification_writer_stored_bytes`
    - The total number of bytes stored into database

### Retriewing metrics

For service running locally:

```
curl localhost:8080/metrics | grep ^notification_writer
```

## Database

PostgreSQL database is used as a storage for new reports and for already reported reports as well.

### Migrations

This service contains an implementation of a simple database migration
mechanism that allows semi-automatic transitions between various database
versions as well as building the latest version of the database from scratch.

Migration mechanism [is described here](https://redhatinsights.github.io/ccx-notification-writer/migrations.html)

### Database schema

Latest database schema is described [in this document](https://redhatinsights.github.io/ccx-notification-writer/database.html)

Please also look [at detailed schema
description](https://redhatinsights.github.io/ccx-notification-writer/db-description/)
for more details about tables, indexes, and keys.

### Check PostgreSQL status

```
service postgresql status
```

### Start PostgreSQL database

```
sudo service postgresql start
```

### Login into the database

```
psql --user postgres
```

List all databases:

```
\l
```

Select the right database:

```
\c notification
```

List of tables:

```
\dt

               List of relations
 Schema |        Name        | Type  |  Owner
--------+--------------------+-------+----------
 public | new_reports        | table | postgres
 public | notification_types | table | postgres
 public | reported           | table | postgres
 public | states             | table | postgres
 public | migration_info     | table | postgres
 public | event_targets      | table | postgres
(6 rows)
```

## Testing

Tests and its configuration is described [in this document](https://redhatinsights.github.io/ccx-notification-writer/testing.html)


## Package manifest

Package manifest is available at [docs/manifest.txt](docs/manifest.txt).
