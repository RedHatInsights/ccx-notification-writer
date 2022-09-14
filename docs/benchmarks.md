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
[tests/benchmark.tom](https://github.com/RedHatInsights/ccx-notification-writer/blob/master/tests/benchmark.toml)
file.

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
