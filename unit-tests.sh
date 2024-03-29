#!/usr/bin/env bash
# Copyright 2021 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

STORAGE=$1

function run_unit_tests() {
    # shellcheck disable=SC2046
    if ! go test -v -timeout 2m -coverprofile coverage.out $(go list ./... | grep -v tests | tr '\n' ' ')
    then
        echo "unit tests failed"
        exit 1
    fi
}

function check_composer() {
    if command -v docker-compose > /dev/null; then
        COMPOSER=docker-compose
    elif command -v podman-compose > /dev/null; then
        COMPOSER=podman-compose
    else
        echo "Please, install docker-compose or podman-compose to run this tests"
        exit 1
    fi
}

path_to_config=$(pwd)/config-devel.toml
export CCX_NOTIFICATION_WRITER_CONFIG_FILE="$path_to_config"
export CCX_NOTIFICATION_WRITER__TESTS_DB="postgres"
export CCX_NOTIFICATION_WRITER__TESTS_DB_ADMIN_PASS="admin"
run_unit_tests
