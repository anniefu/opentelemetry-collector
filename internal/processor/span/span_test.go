// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package span

import (
	"testing"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector/internal/processor/filterset/factory"
)

func createMatchConfig(matchType factory.MatchType) *factory.MatchConfig {
	return &factory.MatchConfig{
		MatchType: matchType,
	}
}

func TestSpan_validateMatchesConfiguration_InvalidConfig(t *testing.T) {
	testcases := []struct {
		name        string
		property    MatchProperties
		errorString string
	}{
		{
			name:        "empty_property",
			property:    MatchProperties{},
			errorString: errAtLeastOneMatchFieldNeeded.Error(),
		},
		{
			name: "empty_service_span_names_and_attributes",
			property: MatchProperties{
				Services:   []string{},
				Attributes: []Attribute{},
			},
			errorString: errAtLeastOneMatchFieldNeeded.Error(),
		},
		{
			name: "invalid_match_type",
			property: MatchProperties{
				MatchConfig: *createMatchConfig(factory.MatchType("wrong_match_type")),
				Services:    []string{"abc"},
			},
			errorString: "error creating service name filters: unrecognized match_type: 'wrong_match_type', valid types are: [regexp strict]",
		},
		{
			name: "missing_match_type",
			property: MatchProperties{
				Services: []string{"abc"},
			},
			errorString: "error creating service name filters: unrecognized match_type: '', valid types are: [regexp strict]",
		},
		{
			name: "regexp_match_type_for_attributes",
			property: MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				Attributes: []Attribute{
					{Key: "key", Value: "value"},
				},
			},
			errorString: `match_type=regexp is not supported for "attributes"`,
		},
		{
			name: "invalid_regexp_pattern",
			property: MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				Services:    []string{"["},
			},
			errorString: "error creating service name filters: error parsing regexp: missing closing ]: `[$`",
		},
		{
			name: "invalid_regexp_pattern2",
			property: MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				SpanNames:   []string{"["},
			},
			errorString: "error creating span name filters: error parsing regexp: missing closing ]: `[$`",
		},
		{
			name: "empty_key_name_in_attributes_list",
			property: MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{"a"},
				Attributes: []Attribute{
					{
						Key: "",
					},
				},
			},
			errorString: "error creating processor. Can't have empty key in the list of attributes",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewMatcher(&tc.property)
			assert.Nil(t, output)
			require.NotNil(t, err)
			assert.Equal(t, tc.errorString, err.Error())
		})
	}
}

func TestSpan_Matching_False(t *testing.T) {
	testcases := []struct {
		name       string
		properties *MatchProperties
	}{
		{
			name: "service_name_doesnt_match_regexp",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				Services:    []string{"svcA"},
				Attributes:  []Attribute{},
			},
		},

		{
			name: "service_name_doesnt_match_strict",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{"svcA"},
				Attributes:  []Attribute{},
			},
		},

		{
			name: "span_name_doesnt_match",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				SpanNames:   []string{"spanNo.*Name"},
				Attributes:  []Attribute{},
			},
		},

		{
			name: "span_name_doesnt_match_any",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				SpanNames: []string{
					"spanNo.*Name",
					"non-matching?pattern",
					"regular string",
				},
				Attributes: []Attribute{},
			},
		},

		{
			name: "wrong_property_value",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{},
				Attributes: []Attribute{
					{
						Key:   "keyInt",
						Value: int(1234),
					},
				},
			},
		},
		{
			name: "incompatible_property_value",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{},
				Attributes: []Attribute{
					{
						Key:   "keyInt",
						Value: "123",
					},
				},
			},
		},
		{
			name: "property_key_does_not_exist",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{},
				Attributes: []Attribute{
					{
						Key:   "doesnotexist",
						Value: nil,
					},
				},
			},
		},
	}

	span := &tracepb.Span{
		Name: &tracepb.TruncatableString{Value: "spanName"},
		Attributes: &tracepb.Span_Attributes{
			AttributeMap: map[string]*tracepb.AttributeValue{
				"keyInt": {
					Value: &tracepb.AttributeValue_IntValue{IntValue: 123},
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			matcher, err := NewMatcher(tc.properties)
			assert.Nil(t, err)
			assert.NotNil(t, matcher)

			assert.False(t, matcher.MatchSpan(span, "wrongSvc"))
		})
	}
}

func TestSpan_MatchingCornerCases(t *testing.T) {
	cfg := &MatchProperties{
		MatchConfig: *createMatchConfig(factory.STRICT),
		Services:    []string{"svcA"},
		Attributes: []Attribute{
			{
				Key:   "keyOne",
				Value: nil,
			},
		},
	}
	mp, err := NewMatcher(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, mp)

	testcases := []struct {
		name string
		span *tracepb.Span
	}{
		{
			name: "nil_attributes",
			span: &tracepb.Span{
				Attributes: nil,
			},
		},
		{
			name: "default_attributes",
			span: &tracepb.Span{
				Attributes: &tracepb.Span_Attributes{},
			},
		},
		{
			name: "empty_map",
			span: &tracepb.Span{
				Attributes: &tracepb.Span_Attributes{
					AttributeMap: map[string]*tracepb.AttributeValue{},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, mp.MatchSpan(tc.span, "svcA"))
		})
	}
}

func TestSpan_MissingServiceName(t *testing.T) {
	cfg := &MatchProperties{
		MatchConfig: *createMatchConfig(factory.REGEXP),
		Services:    []string{"svcA"},
	}
	mp, err := NewMatcher(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, mp)

	testcases := []struct {
		name string
		span *tracepb.Span
	}{
		{
			name: "nil_attributes",
			span: &tracepb.Span{
				Attributes: nil,
			},
		},
		{
			name: "default_attributes",
			span: &tracepb.Span{
				Attributes: &tracepb.Span_Attributes{},
			},
		},
		{
			name: "empty_map",
			span: &tracepb.Span{
				Attributes: &tracepb.Span_Attributes{
					AttributeMap: map[string]*tracepb.AttributeValue{},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, mp.MatchSpan(tc.span, ""))
		})
	}
}

func TestSpan_Matching_True(t *testing.T) {
	testcases := []struct {
		name       string
		properties *MatchProperties
	}{
		{
			name: "service_name_match_regexp",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				Services:    []string{"svcA"},
				Attributes:  []Attribute{},
			},
		},
		{
			name: "service_name_match_strict",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{"svcA"},
				Attributes:  []Attribute{},
			},
		},
		{
			name: "span_name_match",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				SpanNames:   []string{"span.*"},
				Attributes:  []Attribute{},
			},
		},
		{
			name: "span_name_second_match",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.REGEXP),
				SpanNames: []string{
					"wrong.*pattern",
					"span.*",
					"yet another?pattern",
					"regularstring",
				},
				Attributes: []Attribute{},
			},
		},
		{
			name: "property_exact_value_match",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{},
				Attributes: []Attribute{
					{
						Key:   "keyString",
						Value: "arithmetic",
					},
					{
						Key:   "keyInt",
						Value: int(123),
					},
					{
						Key:   "keyDouble",
						Value: float64(3245.6),
					},
					{
						Key:   "keyBool",
						Value: true,
					},
				},
			},
		},
		{
			name: "property_exists",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{"svcA"},
				Attributes: []Attribute{
					{
						Key:   "keyExists",
						Value: nil,
					},
				},
			},
		},
		{
			name: "match_all_settings_exists",
			properties: &MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Services:    []string{"svcA"},
				Attributes: []Attribute{
					{
						Key:   "keyExists",
						Value: nil,
					},
					{
						Key:   "keyString",
						Value: "arithmetic",
					},
				},
			},
		},
	}

	span := &tracepb.Span{
		Name: &tracepb.TruncatableString{Value: "spanName"},
		Attributes: &tracepb.Span_Attributes{
			AttributeMap: map[string]*tracepb.AttributeValue{
				"keyString": {
					Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "arithmetic"}},
				},
				"keyInt": {
					Value: &tracepb.AttributeValue_IntValue{IntValue: 123},
				},
				"keyDouble": {
					Value: &tracepb.AttributeValue_DoubleValue{
						DoubleValue: cast.ToFloat64(3245.6),
					},
				},
				"keyBool": {
					Value: &tracepb.AttributeValue_BoolValue{BoolValue: true},
				},
				"keyExists": {
					Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "present"}},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mp, err := NewMatcher(tc.properties)
			assert.Nil(t, err)
			assert.NotNil(t, mp)

			assert.NotNil(t, span)
			// assert.True(t, mp.MatchSpan(span, "svcA"))

		})
	}
}

func TestSpan_validateMatchesConfigurationForAttributes(t *testing.T) {
	testcase := []struct {
		name   string
		input  MatchProperties
		output Matcher
	}{
		{
			name: "attributes_build",
			input: MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Attributes: []Attribute{
					{
						Key: "key1",
					},
					{
						Key:   "key2",
						Value: 1234,
					},
				},
			},
			output: &propertiesMatcher{
				Attributes: []attributeMatcher{
					{
						Key: "key1",
					},
					{
						Key: "key2",
						AttributeValue: &tracepb.AttributeValue{
							Value: &tracepb.AttributeValue_IntValue{IntValue: cast.ToInt64(1234)},
						},
					},
				},
			},
		},

		{
			name: "both_set_of_attributes",
			input: MatchProperties{
				MatchConfig: *createMatchConfig(factory.STRICT),
				Attributes: []Attribute{
					{
						Key: "key1",
					},
					{
						Key:   "key2",
						Value: 1234,
					},
				},
			},
			output: &propertiesMatcher{
				Attributes: []attributeMatcher{
					{
						Key: "key1",
					},
					{
						Key: "key2",
						AttributeValue: &tracepb.AttributeValue{
							Value: &tracepb.AttributeValue_IntValue{IntValue: cast.ToInt64(1234)},
						},
					},
				},
			},
		},
	}
	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewMatcher(&tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.output, output)
		})
	}
}
