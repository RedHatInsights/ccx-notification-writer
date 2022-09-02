---
layout: page
nav_order: 5
---

# Migrations

This service contains an implementation of a simple database migration
mechanism that allows semi-automatic transitions between various database
versions as well as building the latest version of the database from scratch.

### Printing information about database migrations

```shell
./ccx-notification-writer migration-info
```

### Upgrading the database to the latest available migration

```shell
./ccx-notification-writer migrate latest
```

### Downgrading to the base (empty) database migration version

```shell
./ccx-notification-writer migrate 0
```

Before using the migration mechanism, it is first necessary to initialize the migration information
table `migration_info`. This can be done using the following command:

```shell
./ccx-notification-writer db-init-migration
```

New migrations must be added manually into the code, because it was decided
that modifying the list of migrations at runtime is undesirable.

To migrate the database to a certain version, in either direction (both upgrade and downgrade), use
the `migration.SetDBVersion(*sql.DB, migration.Version)` function.

**To upgrade the database to the highest available version, use
`migration.SetDBVersion(db, migration.GetMaxVersion())`.** This will automatically perform all the
necessary steps to migrate the database from its current version to the highest defined version.

See `migration.go` documentation for an overview of all available DB migration
functionality.

