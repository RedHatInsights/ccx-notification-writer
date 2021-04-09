# ccx-notification-writer
CCX notification writer service

## Usage

```
  -authors
        show authors
  -check-kafka
        check connection to Kafka
  -db-clenaup
        perform database cleanup
  -db-drop-tables
        drop all tables from database
  -db-init
        perform database initialization
  -show-configuration
        show configuration
  -version
        show cleaner version
```

## Database schema

### Table `new_reports`

```
   Column   |            Type             | Modifiers
------------+-----------------------------+-----------
 org_id     | integer                     | not null
 cluster    | character(36)               | not null
 updated_at | timestamp without time zone | not null
Indexes:
    "new_reports_pkey" PRIMARY KEY, btree (org_id, cluster)
```

### Table `states`

```
          Table "public.states"
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

### Table `reported`

```
      Column       |            Type             | Modifiers
-------------------+-----------------------------+-----------
 org_id            | integer                     | not null
 cluster           | character(36)               | not null
 notification_type | integer                     | not null
 state             | integer                     | not null
 report            | character varying           | not null
 updated_at        | timestamp without time zone | not null
Indexes:
    "reported_pkey" PRIMARY KEY, btree (org_id, cluster)
Foreign-key constraints:
    "fk_notification_type" FOREIGN KEY (notification_type) REFERENCES notification_types(id)
    "fk_state" FOREIGN KEY (state) REFERENCES states(id)
```

### Table `notification_types`

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
