package validator

import (
	"regexp"
	"testing"
)

func TestIn(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    string
		list     []string
		expected bool
	}{
		{
			name:     "value in list",
			value:    "foo",
			list:     []string{"bar", "foo", "baz"},
			expected: true,
		},
		{
			name:     "value not in list",
			value:    "qux",
			list:     []string{"bar", "foo", "baz"},
			expected: false,
		},
		{
			name:     "empty list",
			value:    "foo",
			list:     []string{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := In(tc.value, tc.list...)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    string
		pattern  string
		expected bool
	}{
		{
			name:     "valid email",
			value:    "test@example.com",
			pattern:  EmailRX.String(),
			expected: true,
		},
		{
			name:     "invalid email",
			value:    "testexample.com",
			pattern:  EmailRX.String(),
			expected: false,
		},
		{
			name:     "empty pattern",
			value:    "test@example.com",
			pattern:  "",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rx := regexp.MustCompile(tc.pattern)
			result := Matches(tc.value, rx)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		values   []string
		expected bool
	}{
		{
			name:     "no duplicates",
			values:   []string{"foo", "bar", "baz"},
			expected: true,
		},
		{
			name:     "duplicates",
			values:   []string{"foo", "bar", "baz", "foo"},
			expected: false,
		},
		{
			name:     "empty list",
			values:   []string{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Unique(tc.values)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}
