package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/postprocessor"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
)

var testCharacterTavernFetcher = postprocessor.New(fetcher.NewCharacterTavernFetcher())

func TestCharacterTavernFetcher(t *testing.T) {
	assertSourceIsUp(t, testCharacterTavernFetcher)
}

func TestCharacterTavernImport(t *testing.T) {
	t.Parallel()

	const creator = "redpixel"
	const description = "Description AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	url := "https://character-tavern.com/character/redpixel/test"
	fetcherTask := task.New(testClient, testCharacterTavernFetcher, url, testCharacterTavernFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	assert.Equal(t, "character-tavern.com/character/redpixel/test", metadata.CardURL)
	assert.Equal(t, "character-tavern.com/character/redpixel/test", metadata.DirectURL)
	assert.Equal(t, "redpixel/test", metadata.CharacterID)
	assert.Equal(t, "CT_bf4c92dee1f7690c28589360ee3f1380", metadata.PlatformID)

	assertConsistency(t, metadata, card)
	assertCommonFields(t,
		metadata,
		jsonSheet,
		source.CharacterTavern,
		testCharacterTavernFetcher.NormalizeURL(metadata.CharacterID),
		creator,
		description,
		TestCardPersonality,
		stringsx.Empty,
		"main",
		TestCardTagline+"Description"+character.CreatorNotesSeparator+TestCardCreatorNotes,
		character.SpecV3,
		character.V3,
	)
	assertImage(t, card)
}

func TestCharacterTavernImport_CCV3(t *testing.T) {
	url := "https://character-tavern.com/character/tidyup/beth_homeless_on_her_birthday_"
	fetcherTask := task.New(testClient, testCharacterTavernFetcher, url, testCharacterTavernFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	assertConsistency(t, metadata, card)
}

func TestCharacterTavernImportFail(t *testing.T) {
	url := "character-tavern.com/character/brian007/lara_s"
	fetcherTask := task.New(testClient, testCharacterTavernFetcher, url, testCharacterTavernFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.Error(t, err)
	assert.Nil(t, metadata)
}

//func TestCharacterTavernImportNotes(t *testing.T) {
//	t.Parallel()
//
//	const creator = "animatedspell"
//	url := "https://character-tavern.com/character/animatedspell/Violete%20V4"
//	testWyvernFetcher := NewCharacterTavernFetcher()
//	metadata, responseString := testWyvernFetcher.FetchMetadata(url)
//	card := testWyvernFetcher.FetchCharacterCard(metadata, responseString)
//	jsonCard := card.Card
//	println(jsonCard)
//}
