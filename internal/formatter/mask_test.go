// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"reflect"
	"testing"
)

func TestMaskValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "simple string",
			input:    "test",
			expected: MaskReplacementValue,
		},
		{
			name:     "simple number",
			input:    123,
			expected: MaskReplacementValue,
		},
		{
			name:     "simple boolean",
			input:    true,
			expected: MaskReplacementValue,
		},
		{
			name:     "simple float",
			input:    123.45,
			expected: MaskReplacementValue,
		},
		{
			name: "simple map",
			input: map[string]any{
				"name":  "test",
				"value": 123,
			},
			expected: map[string]any{
				"name":  MaskReplacementValue,
				"value": MaskReplacementValue,
			},
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
			expected: map[string]any{
				"user": map[string]any{
					"name": MaskReplacementValue,
					"age":  MaskReplacementValue,
				},
				"active": MaskReplacementValue,
			},
		},
		{
			name: "slice",
			input: []any{
				"item1",
				123,
				true,
			},
			expected: []any{
				MaskReplacementValue,
				MaskReplacementValue,
				MaskReplacementValue,
			},
		},
		{
			name: "nested slice",
			input: []any{
				"item1",
				[]any{"nested1", "nested2"},
				123,
			},
			expected: []any{
				MaskReplacementValue,
				[]any{MaskReplacementValue, MaskReplacementValue},
				MaskReplacementValue,
			},
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
			expected: map[string]any{
				"name": MaskReplacementValue,
				"age":  MaskReplacementValue,
			},
		},
		{
			name: "struct without json tags",
			input: struct {
				Name  string
				Age   int
				Email string
			}{
				Name:  "test",
				Age:   25,
				Email: "test@example.com",
			},
			expected: map[string]any{
				"Name":  MaskReplacementValue,
				"Age":   MaskReplacementValue,
				"Email": MaskReplacementValue,
			},
		},
		{
			name: "pointer to string",
			input: func() *string {
				s := "test"
				return &s
			}(),
			expected: MaskReplacementValue,
		},
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "pointer to struct",
			input: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "test",
				Age:  25,
			},
			expected: map[string]any{
				"name": MaskReplacementValue,
				"age":  MaskReplacementValue,
			},
		},
		{
			name: "map with non-string keys",
			input: map[int]string{
				1: "one",
				2: "two",
			},
			expected: map[string]any{
				"1": MaskReplacementValue,
				"2": MaskReplacementValue,
			},
		},
		{
			name:  "array",
			input: [3]string{"one", "two", "three"},
			expected: []any{
				MaskReplacementValue,
				MaskReplacementValue,
				MaskReplacementValue,
			},
		},
		{
			name: "complex nested structure",
			input: map[string]any{
				"users": []any{
					map[string]any{
						"name": "john",
						"age":  30,
						"address": map[string]any{
							"street": "123 main st",
							"city":   "boston",
						},
					},
					map[string]any{
						"name": "jane",
						"age":  25,
					},
				},
				"active": true,
				"count":  2,
			},
			expected: map[string]any{
				"users": []any{
					map[string]any{
						"name": MaskReplacementValue,
						"age":  MaskReplacementValue,
						"address": map[string]any{
							"street": MaskReplacementValue,
							"city":   MaskReplacementValue,
						},
					},
					map[string]any{
						"name": MaskReplacementValue,
						"age":  MaskReplacementValue,
					},
				},
				"active": MaskReplacementValue,
				"count":  MaskReplacementValue,
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: MaskReplacementValue,
		},
		{
			name:     "zero value",
			input:    0,
			expected: MaskReplacementValue,
		},
		{
			name:     "false boolean",
			input:    false,
			expected: MaskReplacementValue,
		},
		{
			name:     "empty slice",
			input:    []any{},
			expected: []any{},
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "empty struct",
			input:    struct{}{},
			expected: map[string]any{},
		},
		{
			name: "struct with only ignored fields",
			input: struct {
				Email string `json:"-"`
				Token string `json:"-"`
			}{
				Email: "test@example.com",
				Token: "secret",
			},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MaskValue(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MaskValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}
