package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
)

func TestWyvernChatImport(t *testing.T) {
	t.Parallel()

	const creator = "WindWave"
	url := "https://app.wyvern.chat/characters/_wA7KG6rCcfHTFrpD3XLJ8"

	FetchAndAssert(t, url).
		AssertNoErr().
		Source(source.WyvernChat).
		NormalizedURL("wyvern.chat/characters/_wA7KG6rCcfHTFrpD3XLJ8").
		DirectURL("app.wyvern.chat/characters/_wA7KG6rCcfHTFrpD3XLJ8").
		CharacterPlatformID("wA7KG6rCcfHTFrpD3XLJ8").
		CharacterID("_wA7KG6rCcfHTFrpD3XLJ8").
		Name("ChatName").
		Title("Test").
		Tagline("Tagline").
		CreateTime(1737813581781000000).
		UpdateTime(1759046051628000000).
		TagNames("Female", "Prose").
		Nickname("WindWave").
		Username("WindWave").
		CreatorPlatformID("5P3fzFav0TaceZVSJUf5577UImc2").
		BookUpdateTime(1758908855109000000).
		SheetRevision(character.RevisionV2).
		SheetDescription("Description AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").
		SheetPersonality("Personality").
		SheetScenario("Scenario").
		SheetFirstMessage("FirstMessage").
		SheetMessageExamples("ExampleDialog").
		SheetCreatorNotes("Tagline\n\nCreatorNotes\n\nShared Info").
		SheetSystemPrompt("SystemPrompt").
		SheetPostHistoryInstructions("PostHistoryInstructions").
		SheetAlternateGreetingsCount(1).
		SheetHasCharacterBook().
		SheetBookEntryCount(2).
		SheetBookEntryKeyCount(0, 1).
		SheetBookEntryKeyCount(1, 1).
		SheetBookEntryName(0, "LoreBookEntry").
		SheetBookEntryContent(0, "LoreEntryContent").
		SheetBookEntryName(1, "Lexicon Entry Name").
		SheetBookEntryContent(1, "Lexicon Content").
		SheetTagNames("Female", "Prose").
		SheetDepthPromptPrompt("CharacterNote").
		SheetDepthPromptDepth(4).
		SheetNickname("ChatName").
		SheetCreator("WindWave").
		Consistent().
		AssertImage()
}

func TestWyvernChatImport_MultiLoreBook(t *testing.T) {
	t.Parallel()

	const creator = "Ultimate"
	url := "https://app.wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4"

	FetchAndAssert(t, url).
		AssertNoErr().
		SheetBookName("Arvath Dungeon Lore -- Arcaea Lore").
		SheetBookDescriptionContains(character.BookDescriptionSeparator).
		SheetBookDescriptionPrefix("Information for Arvath Dungeon, a dungeon with five levels set in the Arcaea region of my fantasy world.").
		SheetBookEntryCount(38).
		NormalizedURL("wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4").
		DirectURL("app.wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4").
		CharacterID("_MBnL8cfMUVNFVTBe4GNm4").
		CharacterPlatformID("MBnL8cfMUVNFVTBe4GNm4").
		Consistent().
		AssertImage()
}

func TestWyvernChatImport_MalformedLoreBook(t *testing.T) {
	t.Parallel()

	const creator = "Zootopiabest"
	url := "https://app.wyvern.chat/characters/_fgFkgNeGtQEf6k7PQzkXw"

	FetchAndAssert(t, url).
		AssertNoErr().
		SheetNoCharacterBook().
		NormalizedURL("wyvern.chat/characters/_fgFkgNeGtQEf6k7PQzkXw").
		DirectURL("app.wyvern.chat/characters/_fgFkgNeGtQEf6k7PQzkXw").
		CharacterID("_fgFkgNeGtQEf6k7PQzkXw").
		CharacterPlatformID("fgFkgNeGtQEf6k7PQzkXw").
		Consistent().
		AssertImage()
}
