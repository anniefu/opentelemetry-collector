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

	"github.com/stretchr/testify/assert"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"

	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	ptest "github.com/open-telemetry/opentelemetry-collector/processor/processortest"
)

func TestFilterTraceProcessor(t *testing.T) {
	inTraceNames := []string{
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
		inTN  []string // input trace names
		outTN []string // output trace names
	}{
		{
			name: "includeFilter",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: INCLUDE,
				Traces: TraceFilter{
					NameFilters: validFilters,
				},
			},
			inTN: inTraceNames,
			outTN: []string{
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
				Traces: TraceFilter{
					NameFilters: validFilters,
				},
			},
			inTN: inTraceNames,
			outTN: []string{
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
			inTN:  inTraceNames,
			outTN: inTraceNames,
		}, {
			name: "emptyFilterExclude",
			cfg: &Config{
				ProcessorSettings: configmodels.ProcessorSettings{
					TypeVal: typeStr,
					NameVal: typeStr,
				},
				Action: EXCLUDE,
			},
			inTN:  inTraceNames,
			outTN: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := &ptest.CachingTraceConsumer{}
			fmp, err := newFilterTraceProcessor(next, test.cfg)
			assert.NotNil(t, fmp)
			assert.Nil(t, err)

			td := consumerdata.TraceData{
				Spans: make([]*tracepb.Span, len(test.inTN)),
			}

			for idx, in := range test.inTN {
				td.Spans[idx] = &tracepb.Span{
					Name: &tracepb.TruncatableString{
						Value: in,
					},
				}
			}

			cErr := fmp.ConsumeTraceData(context.Background(), td)
			assert.Nil(t, cErr)

			for idx, out := range next.Data.Spans {
				assert.Equal(t, test.outTN[idx], out.Name.GetValue())
			}
		})
	}
}
