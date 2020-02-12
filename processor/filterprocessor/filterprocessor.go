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
	"regexp"

	"github.com/open-telemetry/opentelemetry-collector/processor"
)

// filterProcessor manages state and functionality that is common to filtering both
// traces and metrics.
type filterProcessor struct {
	capabilities processor.Capabilities
	regexes      map[Filter]*regexp.Regexp
	cfg          *Config
}

// stringMatchesFilters returns true if the given string matches any of the given filters.
func (fp *filterProcessor) stringMatchesFilters(toMatch string, filters []Filter) bool {
	for _, f := range filters {
		if r, ok := fp.regexes[f]; ok {
			if r.MatchString(toMatch) {
				return true
			}
		} else {
			// reject matches to any input filters that aren't known to the filterProcessor
			return false
		}
	}

	return false
}

// compileFilters compiles all the given filters and stores them as regexes.
func (fp *filterProcessor) compileFilters(filters []Filter) error {
	for _, f := range filters {
		if re, err := regexp.Compile(string(f)); err == nil {
			fp.regexes[f] = re
		} else {
			return err
		}
	}

	return nil
}

// addFilters adds sets of filters to the filterProcessor's stored regexes.
func (fp *filterProcessor) addFilters(filterSets ...[]Filter) error {
	for _, filters := range filterSets {
		if err := fp.compileFilters(filters); err != nil {
			return err
		}
	}

	return nil
}

func newFilterProcessor(cfg *Config) (*filterProcessor, error) {
	return &filterProcessor{
		regexes:      map[Filter]*regexp.Regexp{},
		capabilities: processor.Capabilities{MutatesConsumedData: true},
		cfg:          cfg,
	}, nil
}
