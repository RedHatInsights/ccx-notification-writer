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

package main

// File metrics contains all metrics that needs to be exposed to Prometheus and
// indirectly to Grafana.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/ccx-notification-writer/
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/metrics.html

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics names
const (
	ConsumedMessagesName          = "consumed_messages"
	ConsumingErrorsName           = "consuming_errors"
	ParsedIncomingMessageName     = "parse_incoming_message"
	CheckSchemaVersionName        = "check_schema_version"
	MarshalReportName             = "marshal_report"
	ShrinkReportName              = "shrink_report"
	CheckLastCheckedTimestampName = "check_last_checked_timestamp"
	StoredMessagesName            = "stored_messages"
	StoredBytesName               = "stored_bytes"
)

// Metrics helps
const (
	ConsumedMessagesHelp          = "The total number of messages consumed from Kafka"
	ConsumingErrorsHelp           = "The total number of errors during consuming messages from Kafka"
	ParsedIncomingMessageHelp     = "The total number of parsed messages"
	CheckSchemaVersionHelp        = "The total number of messages with successful schema check"
	MarshalReportHelp             = "The total number of marshaled reports"
	ShrinkReportHelp              = "The total number of shrunk reports"
	CheckLastCheckedTimestampHelp = "The total number of messages with last checked timestamp"
	StoredMessagesHelp            = "The total number of messages stored into database"
	StoredBytesHelp               = "The total number of bytes stored into database"
)

// ConsumedMessages shows number of messages consumed from Kafka by aggregator
var ConsumedMessages = promauto.NewCounter(prometheus.CounterOpts{
	Name: ConsumedMessagesName,
	Help: ConsumedMessagesHelp,
})

// ConsumingErrors shows the total number of errors during consuming messages from Kafka
var ConsumingErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: ConsumingErrorsName,
	Help: ConsumingErrorsHelp,
})

// ParsedIncomingMessage shows the number of parsed messages
var ParsedIncomingMessage = promauto.NewCounter(prometheus.CounterOpts{
	Name: ParsedIncomingMessageName,
	Help: ParsedIncomingMessageHelp,
})

// CheckSchemaVersion shows the number of messages with correct schema version
var CheckSchemaVersion = promauto.NewCounter(prometheus.CounterOpts{
	Name: CheckSchemaVersionName,
	Help: CheckSchemaVersionHelp,
})

// MarshalReport shows the number of successfully marshaled reports
var MarshalReport = promauto.NewCounter(prometheus.CounterOpts{
	Name: MarshalReportName,
	Help: MarshalReportHelp,
})

// ShrinkReport shows the number of messages with shrunk report
var ShrinkReport = promauto.NewCounter(prometheus.CounterOpts{
	Name: ShrinkReportName,
	Help: ShrinkReportHelp,
})

// CheckLastCheckedTimestamp shows the number of messages with correct timestamp
var CheckLastCheckedTimestamp = promauto.NewCounter(prometheus.CounterOpts{
	Name: CheckLastCheckedTimestampName,
	Help: CheckLastCheckedTimestampHelp,
})

// StoredMessages shows number of messages stored into database
var StoredMessages = promauto.NewCounter(prometheus.CounterOpts{
	Name: StoredMessagesName,
	Help: StoredMessagesHelp,
})

// StoredBytes shows number of bytes stored into database
var StoredBytes = promauto.NewCounter(prometheus.CounterOpts{
	Name: StoredBytesName,
	Help: StoredBytesHelp,
})

// AddMetricsWithNamespace register the desired metrics using a given namespace
func AddMetricsWithNamespace(namespace string) {
	// exposed metrics

	// first unregister all metrics
	prometheus.Unregister(ConsumedMessages)
	prometheus.Unregister(ConsumingErrors)
	prometheus.Unregister(ParsedIncomingMessage)
	prometheus.Unregister(CheckSchemaVersion)
	prometheus.Unregister(MarshalReport)
	prometheus.Unregister(ShrinkReport)
	prometheus.Unregister(CheckLastCheckedTimestamp)

	// and registrer them again
	ConsumedMessages = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      ConsumedMessagesName,
		Help:      ConsumedMessagesHelp,
	})

	ConsumingErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      ConsumingErrorsName,
		Help:      ConsumingErrorsHelp,
	})

	ParsedIncomingMessage = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      ParsedIncomingMessageName,
		Help:      ParsedIncomingMessageHelp,
	})

	CheckSchemaVersion = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      CheckSchemaVersionName,
		Help:      CheckSchemaVersionHelp,
	})

	MarshalReport = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      MarshalReportName,
		Help:      MarshalReportHelp,
	})

	ShrinkReport = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      ShrinkReportName,
		Help:      ShrinkReportHelp,
	})

	CheckLastCheckedTimestamp = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      CheckLastCheckedTimestampName,
		Help:      CheckLastCheckedTimestampHelp,
	})

	StoredMessages = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      StoredMessagesName,
		Help:      StoredMessagesHelp,
	})

	StoredBytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      StoredBytesName,
		Help:      StoredBytesHelp,
	})
}
