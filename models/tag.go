package models

import (
	"cmp"
	"slices"
	"strings"

	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/tidwall/gjson"
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

// MergeTags - given a list of db tags and a list of string tags merge the two lists and return the results
func MergeTags(tags []Tag, stringTags []string) ([]Tag, []string) {
	capacity := len(tags) + len(stringTags)
	tagMap := make(map[Slug]string, capacity)

	for _, tag := range tags {
		if slug := SanitizeSlug(tag.Slug); stringsx.IsNotBlank(slug) {
			tagMap[slug] = ResolveStandardTag(slug, tag.Name)
		}
	}
	for _, tagName := range stringTags {
		if slug := SanitizeSlug(tagName); stringsx.IsNotBlank(slug) {
			tagMap[slug] = ResolveStandardTag(slug, tagName)
		}
	}

	mergedTags := make([]Tag, 0, len(tagMap))
	for slug, name := range tagMap {
		mergedTags = append(mergedTags, Tag{Slug: slug, Name: name})
	}

	slices.SortFunc(mergedTags, func(a, b Tag) int {
		return cmp.Compare(a.Slug, b.Slug)
	})

	stringNames := make([]string, len(mergedTags))
	for i, tag := range mergedTags {
		stringNames[i] = tag.Name
	}

	return mergedTags, stringNames
}

// SanitizeSlug - Sanitizes the given tag to be used as a slug (removes non-ASCII, '-', '_', whitespace and lowers all characters)
func SanitizeSlug(slug Slug) Slug {
	// Remove non-ASCII, symbols, and whitespace, and lower all characters
	return strings.ToLower(stringsx.Remove(slug, symbols.SymbolsWhiteSpaceRegExp))
}

// SanitizeName - Sanitizes the given tag to be used as a name (removes non-ASCII, trims trailing spaces, and titles)
func SanitizeName(name string) string {
	// Remove non-ASCII
	sanitizedTagName := strings.TrimSpace(stringsx.Remove(name, symbols.NonAsciiSymbolsRegExp))

	// Mark symbols with spaces so that the title works properly
	for _, symbol := range symbols.CapitalizeAfter {
		marker := symbols.Space + symbol + symbols.Space
		sanitizedTagName = strings.ReplaceAll(sanitizedTagName, symbol, marker)
	}

	// Return the given name as title
	sanitizedTagName = stringsx.ToTitle(sanitizedTagName)

	// Revert marks
	for _, symbol := range symbols.CapitalizeAfter {
		marker := symbols.Space + symbol + symbols.Space
		sanitizedTagName = strings.ReplaceAll(sanitizedTagName, marker, symbol)
	}

	// Return sanitized name
	return sanitizedTagName
}

// TagsFromJsonArray - given an array GJson result, return a list of tags, extracted from the array
func TagsFromJsonArray(array gjson.Result, extractor func(gjson.Result) string) []Tag {
	tags := gjsonx.ArrayToSlice(
		array,
		func(tag Tag) bool {
			return stringsx.IsNotBlank(tag.Slug)
		},
		func(result gjson.Result) Tag {
			stringTag := extractor(result)
			slug := SanitizeSlug(stringTag)
			return Tag{
				Slug: slug,
				Name: SanitizeName(stringTag),
			}
		},
	)

	slices.SortFunc(tags, func(a, b Tag) int {
		return cmp.Compare(a.Name, b.Name)
	})

	// Return the slice of tags extracted
	return tags
}
