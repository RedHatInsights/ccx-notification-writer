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

## Database benchmarks

Database benchmarks can be run against local or remote database. The
configuration is stored in
[tests/benchmark.toml](https://github.com/RedHatInsights/ccx-notification-writer/blob/master/tests/benchmark.toml)
file.

### Number of records to be inserted into tested database table

In order to be able to check time complexity of SELECT/INSERT/DELETE operations
it is needed to have a technology to specify number of inserted records. It can
be specified via the followin CLI arguments during test/benchmark invocation:

1. `-min-reports`
1. `-max-reports`
1. `-reports-step`

These three paramers can be used to create a sequence compatible with Python's `range` generator.

Additionally it is possible to specify inserted reports counts explicitly:

1. `reports-count=1,2,5,1000

Please note that CLI arguments for tests/benchmarks need to be prepended by `-args`

#### Examples

```bash
go test -run=^$
go test -bench="BenchmarkDelete.*" -count=1 -benchtime=1s -timeout 100m -run=^$
go test -bench="BenchmarkDelete.*" -count=1 -benchtime=2s -timeout 200m -run=^$$ -args -min-reports=3 -max-reports=7 -reports-step=2
go test -bench="BenchmarkDelete.*" -count=1 -benchtime=2s -timeout 200m -run=^$$ -args -reports-count=1,2,5
go test -bench="BenchmarkDelete.*" -count=1 -benchtime=2s -timeout 200m -run=^$$ -args -min-reports=1 -max-reports=10 -reports-step=2 -reports-count=1,2,5
```

### Configuring remote database

For benchmarking against remote database (which is real-world scenarion) the
database needs to be configured to allow remote access:

1. Update file `/var/lib/pgsql/data/pg_hba.conf` to contain:

```
# "local" is for Unix domain socket connections only
local   all             all                                     password
# IPv4 local connections:
host    all             all             127.0.0.1/32            password
# IPv6 local connections:
host    all             all             ::1/128                 password
# Remote access
host    all             all             0.0.0.0/0               password
```

2. Update file `/var/lib/pgsql/data/postgresql.conf` to contain:

```
listen_addresses = '*'          # what IP address(es) to listen on;
```

3. Restart database by using the following command:

```
sudo systemctl restart postgresql
```



## Currently implemented benchmarks

* [db_benchmark_test.go](./packages/db_benchmark_test.html)
