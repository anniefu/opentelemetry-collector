// Copyright 2020 OpenTelemetry Authors
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

package filterprocessor

import (
	"context"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

	"github.com/open-telemetry/opentelemetry-collector/component"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-collector/processor"
)

type filterMetricProcessor struct {
	processor        *filterProcessor
	capabilities     processor.Capabilities
	next             consumer.MetricsConsumer
	metricNameToKeep map[string]bool
}

func newFilterMetricProcessor(next consumer.MetricsConsumer, cfg *Config) (*filterMetricProcessor, error) {
	fp, err := newFilterProcessor(cfg)
	if err != nil {
		return nil, err
	}

	aErr := fp.addFilters(cfg.Metrics.NameFilters)
	if aErr != nil {
		return nil, err
	}

	return &filterMetricProcessor{
		processor: fp,
		next:      next,
	}, nil
}

// GetCapabilities returns the Capabilities assocciated with the resource processor.
func (fmp *filterMetricProcessor) GetCapabilities() processor.Capabilities {
	return fmp.capabilities
}

// Start is invoked during service startup.
func (*filterMetricProcessor) Start(host component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (*filterMetricProcessor) Shutdown() error {
	return nil
}

// ConsumeMetricsData implements the MetricsProcessor interface
func (fmp *filterMetricProcessor) ConsumeMetricsData(ctx context.Context, md consumerdata.MetricsData) error {
	return fmp.next.ConsumeMetricsData(ctx, consumerdata.MetricsData{
		Node:     md.Node,
		Resource: md.Resource,
		Metrics:  fmp.filterMetrics(md.Metrics),
	})
}

// filterSpans filters the given spans based off the filterTraceProcessor's filters.
func (fmp *filterMetricProcessor) filterMetrics(metrics []*metricspb.Metric) []*metricspb.Metric {
	keep := []*metricspb.Metric{}
	for _, m := range metrics {

		if fmp.keepMetric(m) {
			keep = append(keep, m)
		}
	}

	return keep
}

// keepSpan determines whether or not a span should be kept based off the filterTraceProcessor's filters.
func (fmp *filterMetricProcessor) keepMetric(metric *metricspb.Metric) bool {
	p := fmp.processor
	cfg := p.cfg

	name := metric.GetMetricDescriptor().GetName()
	if prevResult, ok := fmp.metricNameToKeep[name]; ok {
		return prevResult
	}

	nameMatch := p.stringMatchesFilters(name, cfg.Traces.NameFilters)
	fmp.metricNameToKeep[name] = nameMatch && cfg.Action == INCLUDE
	return fmp.metricNameToKeep[name]
}
