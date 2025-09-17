package fetcher_test

import (
	"strings"
	"testing"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/impl"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
)

var testWyvernFetcher = fetcher.New(impl.WyvernChatHandler(testClient))

func TestWyvernFetcher(t *testing.T) {
	assertSourceIsUp(t, testWyvernFetcher)
}

func TestWyvernChatImport(t *testing.T) {
	t.Parallel()

	const creator = "WindWave"
	url := "https://app.wyvern.chat/characters/_wA7KG6rCcfHTFrpD3XLJ8"
	fetcherTask := task.New(testWyvernFetcher, url, testWyvernFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	assert.Equal(t, "wyvern.chat/characters/_wA7KG6rCcfHTFrpD3XLJ8", metadata.CardInfo.NormalizedURL)
	assert.Equal(t, "_wA7KG6rCcfHTFrpD3XLJ8", metadata.CharacterID)
	assert.Equal(t, "wA7KG6rCcfHTFrpD3XLJ8", metadata.CardInfo.PlatformID)
	assert.Equal(t, TestCardLoreBookEntryComment, *jsonSheet.Content.CharacterBook.Entries[0].Comment)

	assertConsistency(t, metadata, card)
	assertCommonFields(t,
		metadata,
		jsonSheet,
		source.WyvernChat,
		testWyvernFetcher.NormalizeURL(metadata.CharacterID),
		creator,
		TestCardDescription,
		TestCardPersonality,
		TestCardSystemPrompt,
		stringsx.Empty,
		TestCardTagline+character.CreatorNotesSeparator+TestCardCreatorNotes,
		character.SpecV2,
		character.V2,
	)
	assertCharacterLoreBookCommonFields(t, jsonSheet)
	assertDepthPrompt(t, jsonSheet)
	assertImage(t, card)
}

func TestWyvernChatImport_MultiLoreBook(t *testing.T) {
	t.Parallel()

	const creator = "Ultimate"
	url := "https://app.wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4"
	fetcherTask := task.New(testWyvernFetcher, url, testWyvernFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	assert.Equal(t, jsonSheet.Content.CharacterBook.GetName(), "Arvath Dungeon Lore -- Arcaea Lore")
	assert.Contains(t, jsonSheet.Content.CharacterBook.GetDescription(), "----------------------")
	assert.True(t, strings.HasPrefix(jsonSheet.Content.CharacterBook.GetDescription(), "Information for Arvath Dungeon, a dungeon with five levels set in the Arcaea region of my fantasy world."))
	assert.Len(t, jsonSheet.Content.CharacterBook.Entries, 38)
	assert.Equal(t, "wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4", metadata.NormalizedURL)
	assert.Equal(t, "app.wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4", metadata.DirectURL)
	assert.Equal(t, "_MBnL8cfMUVNFVTBe4GNm4", metadata.CharacterID)
	assert.Equal(t, "MBnL8cfMUVNFVTBe4GNm4", metadata.CardInfo.PlatformID)

	assertConsistency(t, metadata, card)

	assertImage(t, card)
}
