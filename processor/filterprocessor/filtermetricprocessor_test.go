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
	"testing"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	ptest "github.com/open-telemetry/opentelemetry-collector/processor/processortest"
)

var (
	validFilters = []string{
		"exact_match_string",
		".*contains.*",
		".*/suffix",
		"prefix/.*",
		"(a|b)",
	}
)

func TestFilterMetricProcessor(t *testing.T) {
	inMetricNames := []string{
		"exact_match_string",
		"not_exact_string_match",
		"test_contains_match",
		"test/match/suffix",
		"exact_match_string",
		"random",
		"test/match/suffixwrong",
		"prefix/test/match",
		"a",
		"wrongprefix/test/match",
		"not_exact_string_match",
		"c",
	}

	tests := []struct {
		name  string
		cfg   *Config
		inMN  []string // input Metric names
		outMN []string // output Metric names
	}{
		{
			name: "includeFilter",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: INCLUDE,
				Metrics: MetricFilter{
					NameFilters: validFilters,
				},
			},
			inMN: inMetricNames,
			outMN: []string{
				"exact_match_string",
				"test_contains_match",
				"test/match/suffix",
				"exact_match_string",
				"prefix/test/match",
				"a",
			},
		}, {
			name: "excludeFilter",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: EXCLUDE,
				Metrics: MetricFilter{
					NameFilters: validFilters,
				},
			},
			inMN: inMetricNames,
			outMN: []string{
				"not_exact_string_match",
				"random",
				"test/match/suffixwrong",
				"wrongprefix/test/match",
				"not_exact_string_match",
				"c",
			},
		}, {
			name: "emptyFilterInclude",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: INCLUDE,
			},
			inMN:  inMetricNames,
			outMN: inMetricNames,
		}, {
			name: "emptyFilterExclude",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: EXCLUDE,
			},
			inMN:  inMetricNames,
			outMN: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := &ptest.CachingMetricsConsumer{}
			fmp, err := newFilterMetricProcessor(next, test.cfg)
			assert.NotNil(t, fmp)
			assert.Nil(t, err)

			md := consumerdata.MetricsData{
				Metrics: make([]*metricspb.Metric, len(test.inMN)),
			}

			for idx, in := range test.inMN {
				md.Metrics[idx] = &metricspb.Metric{
					MetricDescriptor: &metricspb.MetricDescriptor{
						Name: in,
					},
				}
			}

			cErr := fmp.ConsumeMetricsData(context.Background(), md)
			assert.Nil(t, cErr)

			for idx, out := range next.Data.Metrics {
				assert.Equal(t, test.outMN[idx], out.MetricDescriptor.Name)
			}
		})
	}
}
