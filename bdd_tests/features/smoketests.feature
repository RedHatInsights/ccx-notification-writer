Feature: Basic set of smoke tests


  Scenario: Check if CCX Notification Writer application is available
    Given the system is in default state
     When I look for executable file ccx-notification-writer
     Then I should find that file on PATH


  Scenario: Check if CCX Notification Writer displays help message
    Given the system is in default state
     When I start the CCX Notification Writer with the --help command line flag
     Then I should see help messages displayed on standard output


  Scenario: Check if CCX Notification Writer displays authors
    Given the system is in default state
     When I start the CCX Notification Writer with the --authors command line flag
     Then I should see info about authors displayed on standard output


  Scenario: Check if Postgres database is available
    Given the system is in default state
     When I connect to database named test as user postgres with password postgres
     Then I should be able to connect to such database
     When I close database connection
     Then I should be disconnected
