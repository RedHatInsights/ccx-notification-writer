---
layout: default
nav_order: 0
---

# Description

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
