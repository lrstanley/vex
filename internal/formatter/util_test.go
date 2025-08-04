// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"testing"
)

func TestIsFlatValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: false,
		},
		{
			name: "flat values only strings",
			input: map[string]any{
				"name": "test",
				"type": "string",
			},
			expected: true,
		},
		{
			name: "flat values only numbers",
			input: map[string]any{
				"int":     123,
				"int8":    int8(123),
				"int16":   int16(123),
				"int32":   int32(123),
				"int64":   int64(123),
				"uint":    uint(123),
				"uint8":   uint8(123),
				"uint16":  uint16(123),
				"uint32":  uint32(123),
				"uint64":  uint64(123),
				"uintptr": uintptr(123),
				"float32": float32(123.45),
				"float64": 123.45,
			},
			expected: true,
		},
		{
			name: "flat values only complex numbers",
			input: map[string]any{
				"complex64":  complex64(1 + 2i),
				"complex128": complex128(1 + 2i),
			},
			expected: true,
		},
		{
			name: "flat values only booleans",
			input: map[string]any{
				"true":  true,
				"false": false,
			},
			expected: true,
		},
		{
			name: "flat values only mixed types",
			input: map[string]any{
				"string":  "test",
				"int":     123,
				"float":   123.45,
				"bool":    true,
				"complex": complex128(1 + 2i),
			},
			expected: true,
		},
		{
			name: "nil values",
			input: map[string]any{
				"nil1": nil,
				"nil2": nil,
			},
			expected: true,
		},
		{
			name: "mixed flat and nil values",
			input: map[string]any{
				"string": "test",
				"nil":    nil,
				"int":    123,
			},
			expected: true,
		},
		{
			name: "nested map not flat",
			input: map[string]any{
				"user": map[string]any{
					"name": "john",
					"age":  30,
				},
				"active": true,
			},
			expected: false,
		},
		{
			name: "slice not flat",
			input: map[string]any{
				"items": []any{"item1", "item2"},
				"name":  "test",
			},
			expected: false,
		},
		{
			name: "nested slice not flat",
			input: map[string]any{
				"data": []any{
					"item1",
					[]any{"nested1", "nested2"},
				},
				"count": 5,
			},
			expected: false,
		},
		{
			name: "map with any keys not flat",
			input: map[string]any{
				"config": map[any]any{
					"key1": "value1",
					"key2": "value2",
				},
				"enabled": true,
			},
			expected: false,
		},
		{
			name: "slice of maps not flat",
			input: map[string]any{
				"users": []map[string]any{
					{"name": "john", "age": 30},
					{"name": "jane", "age": 25},
				},
				"total": 2,
			},
			expected: false,
		},
		{
			name: "struct not flat",
			input: map[string]any{
				"person": struct {
					Name string
					Age  int
				}{"john", 30},
				"active": true,
			},
			expected: false,
		},
		{
			name: "interface containing flat value flat",
			input: map[string]any{
				"value": any("test"),
				"count": any(123),
			},
			expected: true,
		},
		{
			name: "interface containing nested structure not flat",
			input: map[string]any{
				"data": any([]any{"item1", "item2"}),
				"name": "test",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IsFlatValue(tt.input)
			if result != tt.expected {
				t.Errorf("IsFlatValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}
