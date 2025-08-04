// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"testing"
)

func TestToJSONWithMask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "null",
		},
		{
			name:     "simple string",
			input:    "test",
			expected: "\"***\"",
		},
		{
			name:     "simple number",
			input:    123,
			expected: "\"***\"",
		},
		{
			name:     "simple boolean",
			input:    true,
			expected: "\"***\"",
		},
		{
			name: "simple map",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			expected: `{
  "name": "***",
  "value": "***"
}`,
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
			expected: `{
  "active": "***",
  "user": {
    "age": "***",
    "name": "***"
  }
}`,
		},
		{
			name: "slice",
			input: []any{
				"item1",
				123,
				true,
			},
			expected: `[
  "***",
  "***",
  "***"
]`,
		},
		{
			name: "struct with json tags",
			input: struct {
				Name  string `json:"name"`
				Age   int    `json:"age"`
				Email string `json:"-"`
			}{
				Name:  "test",
				Age:   25,
				Email: "test@example.com",
			},
			expected: `{
  "age": "***",
  "name": "***"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ToJSON(tt.input, true, 2)
			if result != tt.expected {
				t.Errorf("JSONMask() = %v, want %v", result, tt.expected)
			}
		})
	}
}
