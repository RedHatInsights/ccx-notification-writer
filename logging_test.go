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
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func checkCapture(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// TestLogDuration check the logDuration function from the main module
func TestLogDuration(t *testing.T) {
	startTime := time.Date(2000, time.November, 10, 23, 0, 0, 0, time.UTC)
	endTime := time.Date(2000, time.November, 10, 23, 0, 1, 0, time.UTC)
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.LogDuration(startTime, endTime, 9999, "test message")
	})
	checkCapture(t, err)
	assert.Contains(t, output, "test message") // key
	assert.Contains(t, output, "9999")         // offset
	assert.Contains(t, output, "1000000")      // duration

}
