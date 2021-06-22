# Copyright © 2021 Pavel Tisnovsky, Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Database-related operations performed by BDD tests."""

import subprocess
import psycopg2
from psycopg2.errors import UndefinedTable


from behave import given, then, when


@given(u"Postgres is running")
def check_if_postgres_is_running(context):
    """Check if Postgresql service is active."""
    p = subprocess.Popen(["systemctl", "is-active", "--quiet", "postgresql"])
    assert p is not None

    # interact with the process:
    p.communicate()

    # check the return code
    assert p.returncode == 0, \
        "Postgresql service not running: got return code {code}".format(code=p.returncode)


@when(u"I connect to database named {database} as user {user} with password {password}")
def connect_to_database(context, database, user, password):
    """Perform connection to selected database."""
    connection_string = "dbname={} user={} password={}".format(database, user, password)
    context.connection = psycopg2.connect(connection_string)


@then(u"I should be able to connect to such database")
def check_connection(context):
    """Chck the connection to database."""
    assert context.connection is not None, "connection should be established"


@when(u"I close database connection")
def disconnect_from_database(context):
    """Close the connection to database."""
    context.connection.close()
    context.connection = None


@then(u"I should be disconnected")
def check_disconnection(context):
    """Check that the connection has been closed."""
    assert context.connection is None, "connection should be closed"


@given(u"CCX Notification Writer database is created for user {user} with password {password}")
def database_is_created(context, user, password):
    """Perform connection to CCX Notification Writer database to check its ability."""
    connect_to_database(context, "notification", user, password)


@given(u"CXX Notification Writer database contains all required tables")
def database_contains_all_tables(context):
    """Check if CCX Notification Writer database contains all required tables."""
    raise NotImplementedError(u'STEP: Given CXX Notification Writer database contains all tables')


@when(u"I select all rows from table {table}")
def select_all_rows_from_table(context, table):
    """Select number of all rows from given table."""
    cursor = context.connection.cursor()
    try:
        cursor.execute("SELECT count(*) as cnt from {}".format(table))
        results = cursor.fetchone()
        assert len(results) == 1, "Wrong number of records returned: {}".format(len(results))
        context.query_count = results[0]
    except Exception as e:
        raise e


@then(u"I should get {expected_count:d} rows")
def check_rows_count(context, expected_count):
    """Check if expected number of rows were returned."""
    assert context.query_count == expected_count, \
        "Wrong number of rows returned: {} instead of {}".format(context.query_count, expected_count)
