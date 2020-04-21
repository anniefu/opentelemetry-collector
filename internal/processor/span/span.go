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
	"errors"
	"fmt"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"

	"github.com/open-telemetry/opentelemetry-collector/internal/processor"
	"github.com/open-telemetry/opentelemetry-collector/internal/processor/filterset"
	"github.com/open-telemetry/opentelemetry-collector/internal/processor/filterset/factory"
)

var (
	// TODO Add processor type invoking the NewMatcher in error text.
	errAtLeastOneMatchFieldNeeded = errors.New(
		`error creating processor. At least one ` +
			`of "services", "span_names" or "attributes" field must be specified"`)
)

// TODO: Modify Matcher to invoke both the include and exclude properties so
// calling processors will always have the same logic.
// Matcher is an interface that allows matching a span against a configuration
// of a match.
type Matcher interface {
	MatchSpan(span *tracepb.Span, serviceName string) bool
}

type propertiesMatcher struct {
	// Service names to compare to.
	serviceFilters filterset.FilterSet

	// Span names to compare to.
	nameFilters filterset.FilterSet

	// The attribute values are stored in the internal format.
	Attributes attributesMatcher
}

type attributesMatcher []attributeMatcher

// attributeMatcher is a attribute key/value pair to match to.
type attributeMatcher struct {
	Key            string
	AttributeValue *tracepb.AttributeValue
}

func NewMatcher(config *MatchProperties) (Matcher, error) {
	if config == nil {
		return nil, nil
	}

	if len(config.Services) == 0 && len(config.SpanNames) == 0 && len(config.Attributes) == 0 {
		return nil, errAtLeastOneMatchFieldNeeded
	}

	var err error

	var am attributesMatcher
	if len(config.Attributes) > 0 {
		am, err = newAttributesMatcher(config)
		if err != nil {
			return nil, err
		}
	}

	f := factory.Factory{}

	var serviceFS filterset.FilterSet
	serviceFS = nil
	if len(config.Services) > 0 {
		serviceFS, err = f.CreateFilterSet(config.Services, &config.MatchConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating service name filters: %v", err)
		}
	}

	var nameFS filterset.FilterSet
	nameFS = nil
	if len(config.SpanNames) > 0 {
		nameFS, err = f.CreateFilterSet(config.SpanNames, &config.MatchConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating span name filters: %v", err)
		}
	}

	return &propertiesMatcher{
		serviceFilters: serviceFS,
		nameFilters:    nameFS,
		Attributes:     am,
	}, nil
}

func newAttributesMatcher(config *MatchProperties) (attributesMatcher, error) {
	// attribute matching is only supported with strict matching
	if config.MatchConfig.MatchType != factory.STRICT {
		return nil, fmt.Errorf(
			"%s=%s is not supported for %q",
			MatchTypeFieldName, factory.REGEXP, AttributesFieldName,
		)
	}

	// Convert attribute values from config representation to in-memory representation.
	var rawAttributes []attributeMatcher
	for _, attribute := range config.Attributes {

		if attribute.Key == "" {
			return nil, errors.New("error creating processor. Can't have empty key in the list of attributes")
		}

		entry := attributeMatcher{
			Key: attribute.Key,
		}
		if attribute.Value != nil {
			val, err := processor.AttributeValue(attribute.Value)
			if err != nil {
				return nil, err
			}
			entry.AttributeValue = val
		}

		rawAttributes = append(rawAttributes, entry)
	}
	return rawAttributes, nil
}

// MatchSpan matches a span and service to a set of properties.
// There are 3 sets of properties to match against.
// The service name is checked first, if specified. Then span names are matched, if specified.
// The attributes are checked last, if specified.
// At least one of services, span names or attributes must be specified. It is supported
// to have more than one of these specified, and all specified must evaluate
// to true for a match to occur.
func (mp *propertiesMatcher) MatchSpan(span *tracepb.Span, serviceName string) bool {
	// If a set of properties was not in the config, all spans are considered to match on that property
	if mp.serviceFilters != nil && !mp.serviceFilters.Matches(serviceName) {
		return false
	}

	if mp.nameFilters != nil && !mp.nameFilters.Matches(span.Name.Value) {
		return false
	}

	// Service name and span name matched. Now match attributes.
	return mp.Attributes.match(span)
}

// match attributes specification against a span.
func (ma attributesMatcher) match(span *tracepb.Span) bool {
	// If there are no attributes to match against, the span matches.
	if len(ma) == 0 {
		return true
	}

	// At this point, it is expected of the span to have attributes because of
	// len(ma) != 0. This means for spans with no attributes, it does not match.
	if span.Attributes == nil || len(span.Attributes.AttributeMap) == 0 {
		return false
	}

	// Check that all expected properties are set.
	for _, property := range ma {
		val, exist := span.Attributes.AttributeMap[property.Key]
		if !exist {
			return false
		}

		// This is for the case of checking that the key existed.
		if property.AttributeValue == nil {
			continue
		}

		var isMatch bool
		switch attribValue := val.Value.(type) {
		case *tracepb.AttributeValue_StringValue:
			if sv, ok := property.AttributeValue.GetValue().(*tracepb.AttributeValue_StringValue); ok {
				isMatch = attribValue.StringValue.GetValue() == sv.StringValue.GetValue()
			}
		case *tracepb.AttributeValue_IntValue:
			if iv, ok := property.AttributeValue.GetValue().(*tracepb.AttributeValue_IntValue); ok {
				isMatch = attribValue.IntValue == iv.IntValue
			}
		case *tracepb.AttributeValue_BoolValue:
			if bv, ok := property.AttributeValue.GetValue().(*tracepb.AttributeValue_BoolValue); ok {
				isMatch = attribValue.BoolValue == bv.BoolValue
			}
		case *tracepb.AttributeValue_DoubleValue:
			if dv, ok := property.AttributeValue.GetValue().(*tracepb.AttributeValue_DoubleValue); ok {
				isMatch = attribValue.DoubleValue == dv.DoubleValue
			}
		}
		if !isMatch {
			return false
		}
	}
	return true
}
