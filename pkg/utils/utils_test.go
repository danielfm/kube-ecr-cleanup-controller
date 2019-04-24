package utils

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
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

func TestApplyKeepFilters(t *testing.T) {
	noMatch := "no-match"
	alsoNoMatch := "also-no-match"
	keep := "keep"
	tag := "tag"
	tagRegex := "tag$"

	testCases := []struct {
		filters  []*string
		expected int
	}{
		{
			filters:  []*string{&noMatch},
			expected: 2,
		},
		{
			filters:  []*string{&noMatch, &alsoNoMatch},
			expected: 2,
		},
		{
			filters:  []*string{&keep},
			expected: 1,
		},
		{
			filters:  []*string{&tag},
			expected: 0,
		},
		{
			filters:  []*string{&tagRegex},
			expected: 1,
		},
		{
			filters:  []*string{},
			expected: 2,
		},
	}

	tagTest := "v1.0.0-tag-test"
	tagKeep := "v1.0.0-keep-tag"
	images := []*ecr.ImageDetail{
		{
			ImageTags: []*string{&tagTest},
		}, {
			ImageTags: []*string{&tagKeep},
		},
	}

	for _, testCase := range testCases {
		filtered := ApplyKeepFilters(images, testCase.filters)

		if len(filtered) != testCase.expected {
			t.Errorf("Expected filtered list to be '%d' images, but got '%d'", testCase.expected, len(filtered))
		}
	}
}
