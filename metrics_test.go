/*
Copyright © 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// TestAddMetricsWithNamespace function checks the basic behaviour of function
// AddMetricsWithNamespace from `metrics.go`
func TestAddMetricsWithNamespace(t *testing.T) {
	// add all metrics into the namespace "foobar"
	main.AddMetricsWithNamespace("foobar")

	// check the registration
	assert.NotNil(t, main.ConsumedMessages)
	assert.NotNil(t, main.ConsumingErrors)
	assert.NotNil(t, main.ParsedIncomingMessage)
	assert.NotNil(t, main.CheckSchemaVersion)
	assert.NotNil(t, main.MarshalReport)
	assert.NotNil(t, main.ShrinkReport)
	assert.NotNil(t, main.CheckLastCheckedTimestamp)
	assert.NotNil(t, main.StoredMessages)
	assert.NotNil(t, main.StoredBytes)
}
