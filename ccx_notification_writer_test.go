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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// TestShowVersion checks the function showVersion
func TestShowVersion(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowVersion()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Notification writer version")
}
