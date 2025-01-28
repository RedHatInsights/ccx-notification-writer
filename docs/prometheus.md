---
layout: page
nav_order: 7
---

# Prometheus metrics

It is possible to use `/metrics` REST API endpoint to read all metrics exposed to Prometheus
or to any tool that is compatible with it.
Currently, the following metrics are exposed:

1. `consumed_messages` - the total number of messages consumed from Kafka
1. `consuming_errors` - the total number of errors during consuming messages from Kafka
1. `parse_incoming_message` - the total number of parsed messages
1. `check_schema_version` - the total number of messages with successful schema check
1. `marshal_report` - the total number of marshaled reports
1. `shrink_report` - the total number of shrunk reports
1. `check_last_checked_timestamp` - the total number of messages with last checked timestamp
1. `stored_messages` - the total number of messages stored into database
1. `stored_bytes` - the total number of bytes stored into database



## Metrics namespace

As explained in the [configuration](./configuration) section of this
documentation, a namespace can be provided in order to act as a prefix to the
metric name. If no namespace is provided in the configuration, the metrics will
be exposed as described in this documentation.
