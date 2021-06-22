Feature: Ability to display old records stored in database

  Scenario: Check the ability to display old records from `new_reports` table if the table is empty.
    Given Postgres is running
      And CCX Notification Writer database is created for user postgres with password postgres
      And CCX Notification Writer database is empty
     When I select all rows from table new_reports
     Then I should get 0 rows
     When I select all rows from table reported
     Then I should get 0 rows
     When I close database connection
     Then I should be disconnected
     When I start the CCX Notification Writer with the --db-drop-tables command line flag
     Then the process should exit with status code set to 0
