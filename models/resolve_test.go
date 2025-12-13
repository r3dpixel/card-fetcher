package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveTag(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Tag
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: Tag{Slug: "", Name: ""},
		},
		{
			name:     "Standard tag - exact match lowercase",
			input:    "nsfw",
			expected: Tag{Slug: "nsfw", Name: "NSFW"},
		},
		{
			name:     "Standard tag - exact match with different case",
			input:    "NSFW",
			expected: Tag{Slug: "nsfw", Name: "NSFW"},
		},
		{
			name:     "Standard tag - with spaces and symbols",
			input:    " N S F W ",
			expected: Tag{Slug: "nsfw", Name: "NSFW"},
		},
		{
			name:     "Standard tag - complex case",
			input:    "Fox Girl",
			expected: Tag{Slug: "foxgirl", Name: "Fox Girl"},
		},
		{
			name:     "Standard tag - with symbols",
			input:    "AI-Assistant",
			expected: Tag{Slug: "aiassistant", Name: "AI Assistant"},
		},
		{
			name:     "Standard tag - quoted NTR edge case",
			input:    `"N-TR"`,
			expected: Tag{Slug: `ntr`, Name: `NTR`},
		},
		{
			name:     "Non-standard tag - simple case",
			input:    "custom tag",
			expected: Tag{Slug: "customtag", Name: "Custom Tag"},
		},
		{
			name:     "Non-standard tag - with symbols",
			input:    "my-custom_tag★test",
			expected: Tag{Slug: "mycustomtagtest", Name: "My-Custom_Tag Test"},
		},
		{
			name:     "Non-standard tag - with numbers",
			input:    "tag123",
			expected: Tag{Slug: "tag123", Name: "Tag123"},
		},
		{
			name:     "Non-standard tag - with numbers and non-ASCII",
			input:    "ta★g123",
			expected: Tag{Slug: "tag123", Name: "Ta G123"},
		},
		{
			name:     "Non-standard tag - mixed case and symbols",
			input:    "CUSTOM/Tag-Name",
			expected: Tag{Slug: "customtagname", Name: "CUSTOM/Tag-Name"},
		},
		{
			name:     "Non-standard tag - with extra spaces",
			input:    "  spaced   tag  ",
			expected: Tag{Slug: "spacedtag", Name: "Spaced   Tag"},
		},
		{
			name:     "Non-standard tag - CJK characters",
			input:    "恧恨恩恪",
			expected: Tag{Slug: "恧恨恩恪", Name: "恧恨恩恪"},
		},
		{
			name:     "Non-standard tag - Japanese hiragana",
			input:    "ひらがな",
			expected: Tag{Slug: "ひらがな", Name: "ひらがな"},
		},
		{
			name:     "Non-standard tag - Japanese katakana",
			input:    "カタカナ",
			expected: Tag{Slug: "カタカナ", Name: "カタカナ"},
		},
		{
			name:     "Non-standard tag - Korean",
			input:    "한국어",
			expected: Tag{Slug: "한국어", Name: "한국어"},
		},
		{
			name:     "Standard tag - case insensitive lookup",
			input:    "VTuber",
			expected: Tag{Slug: "vtuber", Name: "VTuber"},
		},
		{
			name:     "Standard tag - complex standardized tag",
			input:    "Can be any POV but made with Male POV in mind",
			expected: Tag{Slug: "canbeanypovbutmadewithmalepovinmind", Name: "Can Be Any POV But Made With Male POV In Mind"},
		},
		{
			name:     "Standard tag - BDSM",
			input:    "bdsm",
			expected: Tag{Slug: "bdsm", Name: "BDSM"},
		},
		{
			name:     "Standard tag - MILF",
			input:    "MILF",
			expected: Tag{Slug: "milf", Name: "MILF"},
		},
		{
			name:     "Standard tag - with underscores",
			input:    "well_intentioned_extremist",
			expected: Tag{Slug: "wellintentionedextremist", Name: "Well Intentioned Extremist"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResolveTag(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
