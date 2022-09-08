---
layout: page
nav_order: 11
---

# Benchmarks

## Preparation steps

Please note that database schema needs to be prepared before benchmarks and
needs to be migrated to latest version.

It is needed to perform the following steps:

1. build CCX Notification Writer
1. run `./ccx-notification-writer -db-init`
1. run `./ccx-notification-writer -db-init-migration`
1. run `./ccx-notification-writer -migrate latest`

## Running benchmarks
Benchmarks can be started from command line by the following command:

```
make benchmark
```

## Currently implemented benchmarks

* [db_benchmark_test.go](./packages/db_benchmark_test.html)
