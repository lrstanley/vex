package fuzzy

import (
	"strings"
	"testing"
)

func TestFindRankedSlice(t *testing.T) {
	t.Parallel()

	type testItem struct {
		name string
		tags []string
	}

	tests := []struct {
		name     string
		filter   string
		items    []testItem
		expected []string
	}{
		{
			name:   "empty filter returns all items",
			filter: "",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
				{name: "carrot", tags: []string{"vegetable"}},
			},
			expected: []string{"apple", "banana", "carrot"},
		},
		{
			name:   "exact match gets highest priority",
			filter: "apple",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "pineapple", tags: []string{"fruit"}},
				{name: "crabapple", tags: []string{"fruit"}},
			},
			expected: []string{"apple", "pineapple", "crabapple"},
		},
		{
			name:   "substring match",
			filter: "app",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "pineapple", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
			},
			expected: []string{"apple", "pineapple"},
		},
		{
			name:   "case insensitive matching",
			filter: "APP",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "PINEAPPLE", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
			},
			expected: []string{"apple", "PINEAPPLE"},
		},
		{
			name:   "fuzzy matching with word boundaries",
			filter: "ap",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "pineapple", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
			},
			expected: []string{"apple", "pineapple"},
		},
		{
			name:   "no matches returns empty slice",
			filter: "xyz",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
			},
			expected: []string{},
		},
		{
			name:   "multiple searchable strings per item",
			filter: "fruit",
			items: []testItem{
				{name: "apple", tags: []string{"fruit", "red"}},
				{name: "banana", tags: []string{"fruit", "yellow"}},
				{name: "carrot", tags: []string{"vegetable", "orange"}},
			},
			expected: []string{"apple", "banana"},
		},
		{
			name:   "consecutive character bonus",
			filter: "ap",
			items: []testItem{
				{name: "apple", tags: []string{"fruit"}},
				{name: "a_p_p_l_e", tags: []string{"fruit"}},
				{name: "banana", tags: []string{"fruit"}},
			},
			expected: []string{"apple", "a_p_p_l_e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FindRankedRow(tt.filter, tt.items, func(item testItem) []string {
				return append([]string{item.name}, item.tags...)
			})

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].name != expected {
					t.Errorf("at index %d: expected %q, got %q", i, expected, result[i].name)
				}
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		query    string
		expected int
	}{
		{
			name:     "empty query returns 0",
			text:     "hello world",
			query:    "",
			expected: 0,
		},
		{
			name:     "exact match gets highest score",
			text:     "hello",
			query:    "hello",
			expected: 10000,
		},
		{
			name:     "exact match case insensitive",
			text:     "Hello",
			query:    "hello",
			expected: 325, // Not exact match due to case difference, so fuzzy matching applies
		},
		{
			name:     "substring match at beginning",
			text:     "hello world",
			query:    "hello",
			expected: 1000,
		},
		{
			name:     "substring match in middle",
			text:     "hello world",
			query:    "world",
			expected: 1000 - 6, // 1000 - position
		},
		{
			name:     "substring match at end",
			text:     "hello world",
			query:    "world",
			expected: 1000 - 6, // 1000 - position
		},
		{
			name:     "fuzzy match with consecutive characters",
			text:     "hello",
			query:    "hl",
			expected: 155, // word boundary bonus: 50 + 50, consecutive bonus: 1*10 + 1*10, length bonus: 100-5
		},
		{
			name:     "fuzzy match with word boundary bonus",
			text:     "hello world",
			query:    "hw",
			expected: 199, // word boundary bonus: 50 + 50, length bonus: 100-11
		},
		{
			name:     "fuzzy match with uppercase bonus",
			text:     "Hello World",
			query:    "hw",
			expected: 259, // word boundary bonus: 50 + 50, uppercase bonus: 30 + 30, length bonus: 100-11
		},
		{
			name:     "fuzzy match with mixed bonuses",
			text:     "HelloWorld",
			query:    "hw",
			expected: 210, // word boundary: 50, uppercase: 30 + 30, length bonus: 100-10
		},
		{
			name:     "no match returns 0",
			text:     "hello world",
			query:    "xyz",
			expected: 0,
		},
		{
			name:     "partial match returns 0",
			text:     "hello world",
			query:    "hellox",
			expected: 0,
		},
		{
			name:     "shorter text gets bonus",
			text:     "hi",
			query:    "hi",
			expected: 10000, // exact match (length bonus is not applied for exact matches)
		},
		{
			name:     "longer text gets penalty",
			text:     "very long text here",
			query:    "text",
			expected: 990, // substring bonus: 1000 - 10 (position), length bonus: 100 - 19
		},
		{
			name:     "consecutive matches get increasing bonus",
			text:     "hello",
			query:    "hel",
			expected: 1000, // substring match: "hel" is a substring of "hello"
		},
		{
			name:     "word boundary with non-letter character",
			text:     "hello-world",
			query:    "hw",
			expected: 199, // word boundary bonus: 50 + 50, length bonus: 100 - 11
		},
		{
			name:     "word boundary with underscore",
			text:     "hello_world",
			query:    "hw",
			expected: 199, // word boundary bonus: 50 + 50, length bonus: 100 - 11
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := calculateScore(tt.text, tt.query)
			if result != tt.expected {
				t.Errorf("calculateScore(%q, %q) = %d, expected %d", tt.text, tt.query, result, tt.expected)
			}
		})
	}
}

func TestFindRankedSliceWithComplexData(t *testing.T) {
	t.Parallel()

	type fileInfo struct {
		path    string
		name    string
		content string
	}

	items := []fileInfo{
		{path: "/usr/bin/python", name: "python", content: "python interpreter"},
		{path: "/usr/bin/python3", name: "python3", content: "python3 interpreter"},
		{path: "/usr/local/bin/pip", name: "pip", content: "python package installer"},
		{path: "/usr/bin/bash", name: "bash", content: "bourne again shell"},
		{path: "/usr/bin/zsh", name: "zsh", content: "z shell"},
	}

	tests := []struct {
		name     string
		filter   string
		expected []string
	}{
		{
			name:     "search by name",
			filter:   "python",
			expected: []string{"python", "python3", "pip"},
		},
		{
			name:     "search by path",
			filter:   "bin",
			expected: []string{"python", "python3", "bash", "zsh", "pip"},
		},
		{
			name:     "search by content",
			filter:   "interpreter",
			expected: []string{"python", "python3"},
		},
		{
			name:     "fuzzy search",
			filter:   "py",
			expected: []string{"python", "python3", "pip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FindRankedRow(tt.filter, items, func(item fileInfo) []string {
				return []string{item.path, item.name, item.content}
			})

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].name != expected {
					t.Errorf("at index %d: expected %q, got %q", i, expected, result[i].name)
				}
			}
		})
	}
}

func TestFindRankedSliceEdgeCases(t *testing.T) {
	t.Parallel()

	type simpleItem struct {
		value string
	}

	tests := []struct {
		name     string
		filter   string
		items    []simpleItem
		expected []string
	}{
		{
			name:     "empty items slice",
			filter:   "test",
			items:    []simpleItem{},
			expected: []string{},
		},
		{
			name:     "single character filter",
			filter:   "a",
			items:    []simpleItem{{value: "apple"}, {value: "banana"}, {value: "cherry"}},
			expected: []string{"apple", "banana"},
		},
		{
			name:     "special characters in filter",
			filter:   "test-file",
			items:    []simpleItem{{value: "test-file.txt"}, {value: "test_file.txt"}, {value: "testfile.txt"}},
			expected: []string{"test-file.txt"}, // Only exact substring match works
		},
		{
			name:     "unicode characters",
			filter:   "café",
			items:    []simpleItem{{value: "café"}, {value: "cafe"}, {value: "restaurant"}},
			expected: []string{"café"}, // Only exact match works for unicode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FindRankedRow(tt.filter, tt.items, func(item simpleItem) []string {
				return []string{item.value}
			})

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].value != expected {
					t.Errorf("at index %d: expected %q, got %q", i, expected, result[i].value)
				}
			}
		})
	}
}

func BenchmarkFindRankedSlice(b *testing.B) {
	type testItem struct {
		name string
		tags []string
	}

	items := make([]testItem, 1000)
	for i := range 1000 {
		items[i] = testItem{
			name: strings.Repeat("item", i%10+1) + string(rune('a'+i%26)),
			tags: []string{"tag1", "tag2", "tag3"},
		}
	}

	for b.Loop() {
		FindRankedRow("item", items, func(item testItem) []string {
			return append([]string{item.name}, item.tags...)
		})
	}
}
