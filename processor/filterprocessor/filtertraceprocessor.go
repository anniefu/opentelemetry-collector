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

	v1 "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"

	"github.com/open-telemetry/opentelemetry-collector/component"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-collector/processor"
)

type filterTraceProcessor struct {
	processor      *filterProcessor
	next           consumer.TraceConsumer
	spanNameToKeep map[string]bool // caches whether trace should be kept by Span name
}

func newFilterTraceProcessor(next consumer.TraceConsumer, cfg *Config) (*filterTraceProcessor, error) {
	fp, err := newFilterProcessor(cfg)
	if err != nil {
		return nil, err
	}

	aErr := fp.addFilters(cfg.Traces.NameFilters)
	if aErr != nil {
		return nil, err
	}

	return &filterTraceProcessor{
		processor: fp,
		next:      next,
	}, nil
}

// ConsumeTraceData implements the TraceProcessor interface
func (ftp *filterTraceProcessor) ConsumeTraceData(ctx context.Context, td consumerdata.TraceData) error {
	return ftp.next.ConsumeTraceData(ctx, consumerdata.TraceData{
		Node:         td.Node,
		Resource:     td.Resource,
		Spans:        ftp.filterSpans(td.Spans),
		SourceFormat: td.SourceFormat,
	})
}

// GetCapabilities returns the Capabilities assocciated with the resource processor.
func (ftp *filterTraceProcessor) GetCapabilities() processor.Capabilities {
	return ftp.processor.capabilities
}

// Start is invoked during service startup.
func (*filterTraceProcessor) Start(host component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (*filterTraceProcessor) Shutdown() error {
	return nil
}

// filterSpans filters the given spans based off the filterTraceProcessor's filters.
func (ftp *filterTraceProcessor) filterSpans(spans []*v1.Span) []*v1.Span {
	keep := []*v1.Span{}
	for _, s := range spans {

		if ftp.keepSpan(s) {
			keep = append(keep, s)
		}
	}

	return keep
}

// keepSpan determines whether or not a span should be kept based off the filterTraceProcessor's filters.
func (ftp *filterTraceProcessor) keepSpan(span *v1.Span) bool {
	p := ftp.processor
	cfg := p.cfg

	name := span.GetName().GetValue()
	if prevResult, ok := ftp.spanNameToKeep[name]; ok {
		return prevResult
	}

	nameMatch := p.stringMatchesFilters(name, cfg.Traces.NameFilters)
	ftp.spanNameToKeep[name] = nameMatch && cfg.Action == INCLUDE
	return ftp.spanNameToKeep[name]
}