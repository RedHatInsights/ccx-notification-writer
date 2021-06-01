# ccx-notification-writer
CCX notification writer service

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

## Usage

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
  -show-configuration
        show configuration
  -version
        show version
```

## Metrics

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
    - The total number of shrinked reports
* `notification_writer_stored_messages`
    - The total number of messages stored into database
* `notification_writer_stored_bytes`
    - The total number of bytes stored into database

### Retriewing metrics

```
curl localhost:8080/metrics | grep ^notification_writer
```

## Database

PostgreSQL database is used as a storage.

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
(4 rows)
```

## Database schema

### Table `new_reports`

This table contains new reports consumed from Kafka topic and stored to
database in shrinked format (some attributes are removed).

```
   Column     |            Type             | Modifiers
--------------+-----------------------------+-----------
 org_id       | integer                     | not null
 account_id   | integer                     | not null
 cluster      | character(36)               | not null
 report       | character varying           | not null
 updated_at   | timestamp without time zone | not null
 kafka_offset | bigint                      | not null default 0
Indexes:
    "new_reports_pkey" PRIMARY KEY, btree (org_id, cluster, updated_at)
    "report_kafka_offset_btree_idx" btree (kafka_offset)
```

### Table `reported`

Information of notifications reported to user or skipped due to some
conditions.

```
      Column       |            Type             | Modifiers
-------------------+-----------------------------+-----------
 org_id            | integer                     | not null
 account_id        | integer                     | not null
 cluster           | character(36)               | not null
 notification_type | integer                     | not null
 state             | integer                     | not null
 report            | character varying           | not null
 updated_at        | timestamp without time zone | not null
 notified_at       | timestamp without time zone | not null
 error_log         | character varying           | 

Indexes:
    "reported_pkey" PRIMARY KEY, btree (org_id, cluster)
Foreign-key constraints:
    "fk_notification_type" FOREIGN KEY (notification_type) REFERENCES notification_types(id)
    "fk_state" FOREIGN KEY (state) REFERENCES states(id)
```

### Table `notification_types`

This table contains list of all notification types used by Notification service.
Frequency can be specified as in `crontab` - https://crontab.guru/

```
  Column   |       Type        | Modifiers
-----------+-------------------+-----------
 id        | integer           | not null
 value     | character varying | not null
 frequency | character varying | not null
 comment   | character varying |
Indexes:
    "notification_types_pkey" PRIMARY KEY, btree (id)
Referenced by:
    TABLE "reported" CONSTRAINT "fk_notification_type" FOREIGN KEY (notification_type) REFERENCES notification_types(id)
```

Currently the following values are stored in this read-only table:

```
 id |  value  |  frequency  |               comment                
----+---------+-------------+--------------------------------------
  1 | instant | * * * * * * | instant notifications performed ASAP
  2 | instant | @weekly     | weekly summary
(2 rows)
```

### Table `states`

This table contains states for each row stored in `reported` table. User can be
notified about the report, report can be skipped if the same as previous,
skipped becuase of lower pripority, or can be in error state.

```
 Column  |       Type        | Modifiers
---------+-------------------+-----------
 id      | integer           | not null
 value   | character varying | not null
 comment | character varying |
Indexes:
    "states_pkey" PRIMARY KEY, btree (id)
Referenced by:
    TABLE "reported" CONSTRAINT "fk_state" FOREIGN KEY (state) REFERENCES states(id)
```

Currently the following values are stored in this read-only table:

```
 id | value |                   comment                   
----+-------+---------------------------------------------
  1 | sent  | notification has been sent to user
  2 | same  | skipped, report is the same as previous one
  3 | lower | skipped, all issues has low priority
  4 | error | notification delivery error
(4 rows)
```
