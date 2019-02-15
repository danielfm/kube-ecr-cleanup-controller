package utils

import (
	"testing"
)

func TestParseCommaSeparatedList(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			// No items
			input:    "",
			expected: []string{},
		},
		{
			// No items
			input:    "  ",
			expected: []string{},
		},
		{
			// No items
			input:    " ,  ",
			expected: []string{},
		},
		{
			// One item
			input:    "item-1",
			expected: []string{"item-1"},
		},
		{
			// One item
			input:    " item-1, ",
			expected: []string{"item-1"},
		},
		{
			// One item
			input:    " , item-1 , ",
			expected: []string{"item-1"},
		},
		{
			// Two items
			input:    "item-1, item-2",
			expected: []string{"item-1", "item-2"},
		},
	}

	for _, testCase := range testCases {
		output := ParseCommaSeparatedList(testCase.input)

		if len(output) != len(testCase.expected) {
			t.Errorf("Number of expected elements expected to be %d, but was %d", len(testCase.expected), len(output))
		}

		for i := range output {
			if *output[i] != testCase.expected[i] {
				t.Errorf("Expected output[%d] to be '%s', but was '%s'", i, testCase.expected[i], *output[i])
			}
		}
	}
}
