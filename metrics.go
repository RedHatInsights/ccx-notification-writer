// Copyright 2021 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// File metrics contains all metrics that needs to be exposed to Prometheus and
// indirectly to Grafana.

import (
	"github.com/RedHatInsights/insights-operator-utils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics names
const (
	ConsumedMessagesName = "consumed_messages"
	ConsumingErrorsName  = "consuming_errors"
)

// Metrics helps
const (
	ConsumedMessagesHelp = "The total number of messages consumed from Kafka"
	ConsumingErrorsHelp  = "The total number of errors during consuming messages from Kafka"
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

// AddMetricsWithNamespace register the desired metrics using a given namespace
func AddMetricsWithNamespace(namespace string) {
	metrics.AddAPIMetricsWithNamespace(namespace)

	prometheus.Unregister(ConsumedMessages)
	prometheus.Unregister(ConsumingErrors)

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
}
