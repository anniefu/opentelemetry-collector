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
	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
)

// Config defines configuration for Resource processor.
type Config struct {
	configmodels.ProcessorSettings `mapstructure:",squash"`
	Action                         Action       `mapstructure:"action"`
	Metrics                        MetricFilter `mapstructure:"metrics"`
	Traces                         TraceFilter  `mapstructure:"traces"`
}

// Action is the enum to specify what happens to metrics and
// traces that match the filter.
type Action string

const (
	// INCLUDE means metrics or traces matching the filter will be
	// included in further processing, all other data will be dropped.
	INCLUDE Action = "include"
	// EXCLUDE means metrics or traces matching the filter will be excluded
	// from further processing and dropped, all other data will continue to be processed.
	EXCLUDE Action = "exclude"
)

// Filter is an re2 regex string, see https://github.com/google/re2/wiki/Syntax.
type Filter string

// MetricFilter filters by Metric properties.
type MetricFilter struct {
	// NameFilters filters by the Name specified in the Metric's MetricDescriptor.
	NameFilters []Filter `mapstructure:"names"`
}

// TraceFilter filters by Span properties.
type TraceFilter struct {
	// NameFilters filters the Name specified in the Span.
	NameFilters []Filter `mapstructure:"names"`
}
