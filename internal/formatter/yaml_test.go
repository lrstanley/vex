// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"testing"
)

func TestToYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		mask     bool
		indent   int
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			mask:     false,
			indent:   2,
			expected: "null",
		},
		{
			name:     "simple string",
			input:    "test",
			mask:     false,
			indent:   2,
			expected: "test",
		},
		{
			name:     "simple number",
			input:    123,
			mask:     false,
			indent:   2,
			expected: "123",
		},
		{
			name:     "simple boolean",
			input:    true,
			mask:     false,
			indent:   2,
			expected: "true",
		},
		{
			name: "simple map",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			mask:   false,
			indent: 2,
			expected: `name: test
value: 123`,
		},
		{
			name: "nested map",
			input: map[string]any{
				"user": map[string]any{
					"name": "john",
					"age":  30,
				},
				"active": true,
			},
			mask:   false,
			indent: 2,
			expected: `active: true
user:
  age: 30
  name: john`,
		},
		{
			name: "slice",
			input: []any{
				"item1",
				123,
				true,
			},
			mask:   false,
			indent: 2,
			expected: `- item1
- 123
- true`,
		},
		{
			name: "struct with yaml tags",
			input: struct {
				Name  string `yaml:"name"`
				Age   int    `yaml:"age"`
				Email string `yaml:"-"`
			}{
				Name:  "test",
				Age:   25,
				Email: "test@example.com",
			},
			mask:   false,
			indent: 2,
			expected: `name: test
age: 25`,
		},
		{
			name:     "simple string masked",
			input:    "test",
			mask:     true,
			indent:   2,
			expected: "\"***\"",
		},
		{
			name:     "simple number masked",
			input:    123,
			mask:     true,
			indent:   2,
			expected: "\"***\"",
		},
		{
			name: "simple map masked",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			mask:   true,
			indent: 2,
			expected: `name: "***"
value: "***"`,
		},
		{
			name: "nested map masked",
			input: map[string]any{
				"user": map[string]any{
					"name": "john",
					"age":  30,
				},
				"active": true,
			},
			mask:   true,
			indent: 2,
			expected: `active: "***"
user:
  age: "***"
  name: "***"`,
		},
		{
			name: "slice masked",
			input: []any{
				"item1",
				123,
				true,
			},
			mask:   true,
			indent: 2,
			expected: `- "***"
- "***"
- "***"`,
		},
		{
			name: "with indentation",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			mask:   false,
			indent: 2,
			expected: `name: test
value: 123`,
		},
		{
			name: "with indentation and masking",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			mask:   true,
			indent: 4,
			expected: `name: "***"
value: "***"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ToYAML(tt.input, tt.mask, tt.indent)
			if result != tt.expected {
				t.Errorf("ToYAML() = %v, want %v", result, tt.expected)
			}
		})
	}
}
