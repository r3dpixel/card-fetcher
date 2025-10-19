package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
)

func TestNyaiMeImport(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://nyai.me/ai/bots/Test_aru").
		AssertNoErr().
		Source(source.NyaiMe).
		NormalizedURL("nyai.me/ai/bots/Test_aru").
		DirectURL("nyai.me/ai/bots/Test_aru").
		CharacterPlatformID("1165").
		CharacterID("Test_aru").
		Name("ChatName").
		Title("Test").
		Tagline("Test").
		CreateTime(1737826389292081000).
		UpdateTime(1737826389292081000).
		IsForked(false).
		TagNames("Female").
		Nickname("cutie2").
		Username("cutie2").
		CreatorPlatformID("cutie2").
		BookUpdateTime(1737826389292081000).
		SheetRevision(character.RevisionV2).
		SheetDescription("Description").
		SheetPersonality(stringsx.Empty).
		SheetScenario("Scenario").
		SheetFirstMessage("FirstMessage").
		SheetMessageExamples("ExampleDialog").
		SheetCreatorNotes("Test"+character.CreatorNotesSeparator+"Test"+character.CreatorNotesSeparator+"CreatorNotes").
		SheetSystemPrompt("SystemPrompt").
		SheetPostHistoryInstructions("PostHistoryInstructions").
		SheetAlternateGreetingsCount(1).
		SheetTagNames("Female").
		SheetCreator("cutie2").
		SheetHasCharacterBook().
		SheetBookEntryCount(1).
		SheetBookEntryKeyCount(0, 1).
		SheetBookEntryName(0, "LoreBookEntry").
		SheetBookEntryContent(0, "LoreEntryContent").
		SheetDepthPromptPrompt("CharacterNote").
		SheetDepthPromptDepth(0).
		SheetNickname("ChatName").
		SheetCreator("cutie2").
		Consistent().
		AssertImage()
}
