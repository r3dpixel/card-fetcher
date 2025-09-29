package models

import (
	"cmp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bytedance/sonic/ast"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
)

type Slug = string

// Tag - DB Model for storing tags
type Tag struct {
	Slug Slug
	Name string
}

// TagsToNames - get a list of string tags from a list of db tags
func TagsToNames(tags []Tag) []string {
	stringTags := make([]string, len(tags))
	for index, tag := range tags {
		stringTags[index] = tag.Name
	}
	return stringTags
}

func TagsToSlugs(tags []Tag) []Slug {
	slugs := make([]string, len(tags))
	for index, tag := range tags {
		slugs[index] = tag.Slug
	}
	return slugs
}

// TagsFromMap - transforms a map into a list of Tags (no sanitization will be applied)
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

// TagsFromJsonArray - given an array GJson result, return a list of tags, extracted from the array
func TagsFromJsonArray(array *ast.Node, extractor func(result *ast.Node) string) []Tag {
	tags := sonicx.ArrayToSlice(
		array,
		func(tag Tag) bool {
			return stringsx.IsNotBlank(tag.Slug)
		},
		func(result *ast.Node) Tag {
			return ResolveTag(extractor(result))
		},
	)

	slices.SortFunc(tags, func(a, b Tag) int {
		return cmp.Compare(a.Name, b.Name)
	})

	// Return the slice of tags extracted
	return tags
}

// SanitizeSlug - Sanitizes the given tag to be used as a slug (removes non-ASCII, '-', '_', whitespace and lowers all characters)
func SanitizeSlug(slug Slug) Slug {
	// Remove non-ASCII, symbols, and whitespace, and lower all characters
	return strings.ToLower(stringsx.Remove(slug, symbols.SymbolsWhiteSpaceRegExp))
}

// SanitizeName - Sanitizes the given tag to be used as a name (removes non-ASCII, trims trailing spaces, and titles)

var capitalizeAfterSet = initCapitalizeAfterSet()

func SanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == stringsx.Empty {
		return name
	}
	return strings.TrimSpace(processAsciiRemainder(name))
}

func processAsciiRemainder(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	capitalizeNext := true
	for i := 0; i < len(input); i++ {
		b := input[i]
		if b >= utf8.RuneSelf {
			return processUnicodeRemainder(input[i:], capitalizeNext, &result)
		}

		if isASCIILetter(b) && capitalizeNext {
			capitalizeNext = false
			b = toASCIIUpper(b)
		}
		if stringsx.IsAsciiSpace(b) || capitalizeAfterSet[b] == 1 {
			capitalizeNext = true
		}

		result.WriteByte(b)
	}

	return result.String()
}

func processUnicodeRemainder(input string, capitalizeNext bool, result *strings.Builder) string {
	for _, r := range input {
		isLetter := unicode.IsLetter(r)
		if r >= utf8.RuneSelf && !isLetter && !unicode.IsNumber(r) {
			capitalizeNext = true
			result.WriteByte(symbols.SpaceByte)
			continue
		}

		if isLetter && capitalizeNext {
			capitalizeNext = false
			r = unicode.ToUpper(r)
		}
		if unicode.IsSpace(r) || (r < utf8.RuneSelf && capitalizeAfterSet[uint8(r)] == 1) {
			capitalizeNext = true
		}
		result.WriteRune(r)
	}

	return result.String()
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func toASCIIUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func initCapitalizeAfterSet() [256]uint8 {
	var symbolSet [256]uint8
	for _, symbol := range symbols.CapitalizeAfter {
		symbolSet[symbol[0]] = 1
	}
	return symbolSet
}
