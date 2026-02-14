package models

import (
	"cmp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
)

// capitalizeAfterSet is a set of symbols that should be capitalized after
var capitalizeAfterSet = initCapitalizeAfterSet()

// Slug type representing a tag slug
type Slug = string

// Tag struct representing a tag
type Tag struct {
	Slug Slug
	Name string
}

// TagsToNames get a slice of tag names from a slice of tags
func TagsToNames(tags []Tag) []string {
	// Create a slice of names from the tags
	names := make([]string, len(tags))
	// Iterate over the tags and populate the slice
	for index, tag := range tags {
		names[index] = tag.Name
	}
	// Return the slice
	return names
}

// TagsToSlugs get a slice of tag slugs from a slice of tags
func TagsToSlugs(tags []Tag) []Slug {
	// Create a slice of slugs from the tags
	slugs := make([]string, len(tags))
	// Iterate over the tags and populate the slice
	for index, tag := range tags {
		slugs[index] = tag.Slug
	}
	// Return the slice
	return slugs
}

// TagsFromMap transforms a map into a list of Tags (no sanitization will be applied)
func TagsFromMap(tags map[Slug]string) []Tag {
	// Slice to hold the list of tags
	result := make([]Tag, len(tags))

	// Process each map entry and transform into a tag
	index := 0
	for slug, name := range tags {
		result[index] = Tag{
			Slug: slug,
			Name: name,
		}
		index++
	}

	// Return result
	return result
}

// TagsFromJsonArray extracts tags from a JSON array
func TagsFromJsonArray(array *sonicx.Wrap, extractor func(result *sonicx.Wrap) string) []Tag {
	tags := sonicx.ArrayToSlice(
		array,
		func(tag Tag) bool {
			return stringsx.IsNotBlank(tag.Slug)
		},
		func(result *sonicx.Wrap) Tag {
			return ResolveTag(extractor(result))
		},
	)

	slices.SortFunc(tags, func(a, b Tag) int {
		return cmp.Compare(a.Name, b.Name)
	})

	// Return the slice of tags extracted
	return tags
}

// SanitizeSlug sanitizes the given tag to be used as a slug (removes non-ASCII, '-', '_', whitespace and lowers all characters)
func SanitizeSlug(slug Slug) Slug {
	// Remove non-ASCII, symbols, and whitespace, and lower all characters
	return strings.ToLower(stringsx.Remove(slug, symbols.SymbolsWhiteSpaceRegExp))
}

// SanitizeName sanitizes the given tag to be used as a name (removes non-ASCII, trims trailing spaces, and titles)
func SanitizeName(name string) string {
	// Trim trailing spaces
	name = strings.TrimSpace(name)
	// If the name is empty, return it immediately
	if name == "" {
		return name
	}
	// Process the name
	return strings.TrimSpace(processAsciiRemainder(name))
}

// processAsciiRemainder sanitize the remainder of the string assuming it is ASCII (fast path)
// switches to Unicode processing after the first Unicode character
func processAsciiRemainder(input string) string {
	// Create a result buffer
	var result strings.Builder
	// Grow the buffer to the length of the input string
	result.Grow(len(input))

	// Flag indicating whether the next character should be capitalized
	capitalizeNext := true

	// Iterate over the string and process each character
	for i := 0; i < len(input); i++ {
		// Get the current rune
		b := input[i]
		// If the rune is a Unicode character, switch to Unicode processing
		if b >= utf8.RuneSelf {
			// Process the remainder of the string as Unicode
			return processUnicodeRemainder(input[i:], capitalizeNext, &result)
		}

		// Process the ASCII character
		if isASCIILetter(b) && capitalizeNext {
			// Reset the flag
			capitalizeNext = false
			// Capitalize the character
			b = toASCIIUpper(b)
		}
		// If the character is a space or should be capitalized after, set the flag
		if stringsx.IsAsciiSpace(b) || capitalizeAfterSet[b] == 1 {
			// Set the flag
			capitalizeNext = true
		}

		// Write the character to the result
		result.WriteByte(b)
	}

	// Return the result string
	return result.String()
}

// processAsciiRemainder sanitize the remainder of the string assuming it is Unicode (slow path)
func processUnicodeRemainder(input string, capitalizeNext bool, result *strings.Builder) string {
	// Iterate over the string and process each rune
	for _, r := range input {
		// Check if the rune is a letter
		isLetter := unicode.IsLetter(r)
		// Check if the capitalization should be changed
		if r >= utf8.RuneSelf && !isLetter && !unicode.IsNumber(r) {
			// Set the flag
			capitalizeNext = true
			// Replace the rune with a space (not letter or number, or ASCII character)
			result.WriteByte(symbols.SpaceByte)
			continue
		}

		// If the rune is a letter and should be capitalized, capitalize it
		if isLetter && capitalizeNext {
			// Reset the flag
			capitalizeNext = false
			// Capitalize the rune
			r = unicode.ToUpper(r)
		}

		// If the rune is a space or should be capitalized after, set the flag
		if unicode.IsSpace(r) || (r < utf8.RuneSelf && capitalizeAfterSet[uint8(r)] == 1) {
			// Set the flag
			capitalizeNext = true
		}

		// Write the rune to the result
		result.WriteRune(r)
	}

	// Return the result string
	return result.String()
}

// isASCIILetter returns true if the given byte is an ASCII letter
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// toASCIIUpper converts a lowercase ASCII letter to uppercase
func toASCIIUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

// initCapitalizeAfterSet initializes a set of symbols that should be capitalized after
func initCapitalizeAfterSet() [256]uint8 {
	var symbolSet [256]uint8
	for _, symbol := range symbols.CapitalizeAfter {
		symbolSet[symbol[0]] = 1
	}
	return symbolSet
}
