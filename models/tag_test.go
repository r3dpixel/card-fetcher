package models

import (
	"cmp"
	"slices"
	"testing"

	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeSlug(t *testing.T) {
	testCases := []struct {
		name     string
		input    Slug
		expected Slug
	}{
		{"Empty string", "", ""},
		{"Simple lowercase", "tag", "tag"},
		{"With uppercase", "TaG", "tag"},
		{"With spaces", " my tag ", "mytag"},
		{"With symbols", "tag-one_two", "tagonetwo"},
		{"With non-ASCII", "ta★g1", "tag1"},
		{"Complex string", "Rand st★rin★g,/- sh★ou★ld be cap", "randstringshouldbecap"},
		{"Non-alphabet KJC characters ", "恧恨恩恪", "恧恨恩恪"},
		{"Mixed alphanumeric and symbols", "Tag-1/2 3", "tag123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, SanitizeSlug(tc.input))
		})
	}
}

func TestSanitizeName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"First letter capitalized", "Db One", "Db One"},
		{"Second letter capitalized", "DB One", "DB One"},
		{"All capitalized", "RPG", "RPG"},
		{"Simple lowercase", "tag", "Tag"},
		{"Already titled", "Tag Name", "Tag Name"},
		{"With extra spaces", "  my tag  ", "My Tag"},
		{"With non-ASCII", "ta ★g1", "Ta  G1"},
		{"With symbols for capitalization", "a-b/c,d", "A-B/C,D"},
		{"Complex string", "Rand st★rin★g,/- sh★ou★ld be cap", "Rand St Rin G,/- Sh Ou Ld Be Cap"},
		{"Trim space string", "   Rand st★rin★g,/- sh★ou★ld be cap   ", "Rand St Rin G,/- Sh Ou Ld Be Cap"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, SanitizeName(tc.input))
		})
	}
}

func TestKJC(t *testing.T) {
	chinese := Slug(`恧恨恩恪恫恬恭恮息恰恱恲恳恴恵恶恷恸恹恺恻恼恽恾恿悀悁悂悃悄悅悆悇悈悉悊悋悌悍悎悏悐悑悒悓悔悕悖悗悘悙悚悛悜悝悞悟悠悡悢患悤悥悦悧您悩悪悫悬悭悮悯悰悱悲悳悴悵悶悷悸悹悺悻悼悽悾悿惀惁惂惃惄情惆惇惈惉惊惋惌惍惎`)
	japanese1 := Slug(`ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶヷヸヹヺ`)
	japanese2 := Slug(`ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔ`)
	korean := Slug(`ᄀᄁᄂᄃᄄᄅᄆᄇᄈᄉᄊᄋᄌᄍᄎᄏᄐᄑ햬양약얀야앵액애앞앙압암알안악아어억언얼엄업엉에여역연열염엽영예용욕요왼외왜왕왈완와옹옴올온옥오우욱운울움웅워원월위유육윤율융윷잎잉입임일인익이의응읍음을은으`)

	assert.Equal(t, SanitizeSlug(chinese), chinese)
	assert.Equal(t, SanitizeSlug(japanese1), japanese1)
	assert.Equal(t, SanitizeSlug(japanese2), japanese2)
	assert.Equal(t, SanitizeSlug(korean), korean)

}

func TestTagsToNames(t *testing.T) {
	testCases := []struct {
		name     string
		input    []Tag
		expected []string
	}{
		{"Nil slice", nil, []string{}},
		{"Empty slice", []Tag{}, []string{}},
		{"Populated slice", []Tag{
			{Slug: "tag1", Name: "Tag One"},
			{Slug: "tag2", Name: "Tag Two"},
		}, []string{"Tag One", "Tag Two"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, TagsToNames(tc.input))
		})
	}
}

func TestTagsToSlugs(t *testing.T) {
	testCases := []struct {
		name     string
		input    []Tag
		expected []string
	}{
		{"Nil slice", nil, []string{}},
		{"Empty slice", []Tag{}, []string{}},
		{"Populated slice", []Tag{
			{Slug: "tag1", Name: "Tag One"},
			{Slug: "tag2", Name: "Tag Two"},
		}, []string{"tag1", "tag2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, TagsToSlugs(tc.input))
		})
	}
}

func TestTagsFromMap(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[Slug]string
		expected []Tag
	}{
		{"Nil map", nil, []Tag{}},
		{"Empty map", map[Slug]string{}, []Tag{}},
		{"Populated map", map[Slug]string{
			"tag1": "Tag One",
			"tag2": "Tag Two",
		}, []Tag{
			{Slug: "tag1", Name: "Tag One"},
			{Slug: "tag2", Name: "Tag Two"},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TagsFromMap(tc.input)
			// Sort both slices to ensure consistent comparison
			slices.SortFunc(result, func(a, b Tag) int { return cmp.Compare(a.Slug, b.Slug) })
			slices.SortFunc(tc.expected, func(a, b Tag) int { return cmp.Compare(a.Slug, b.Slug) })
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTagsFromJsonArray(t *testing.T) {
	testCases := []struct {
		name      string
		jsonInput string
		extractor func(node *sonicx.Wrap) string
		expected  []Tag
	}{
		{
			name:      "Simple string array",
			jsonInput: `["tag1", "T A G 2", "★tag3"]`,
			extractor: sonicx.WrapString,
			expected: []Tag{
				{Slug: "tag1", Name: "Tag1"},
				{Slug: "tag2", Name: "T A G 2"},
				{Slug: "tag3", Name: "Tag3"},
			},
		},
		{
			name:      "Array of objects",
			jsonInput: `[{"name": "tag1"}, {"name": "tag2"}]`,
			extractor: func(r *sonicx.Wrap) string { return r.Get("name").String() },
			expected: []Tag{
				{Slug: "tag1", Name: "Tag1"},
				{Slug: "tag2", Name: "Tag2"},
			},
		},
		{
			name:      "Empty array",
			jsonInput: `[]`,
			extractor: sonicx.WrapString,
			expected:  nil,
		},
		{
			name:      "Array with blank items to be filtered",
			jsonInput: `["tag1", "  ", "tag2"]`,
			extractor: sonicx.WrapString,
			expected: []Tag{
				{Slug: "tag1", Name: "Tag1"},
				{Slug: "tag2", Name: "Tag2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonArray, _ := sonicx.GetFromString(tc.jsonInput)
			result := TagsFromJsonArray(jsonArray, tc.extractor)
			slices.SortFunc(tc.expected, func(a, b Tag) int { return cmp.Compare(a.Name, b.Name) })
			assert.Equal(t, tc.expected, result)
		})
	}
}
