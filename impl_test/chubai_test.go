package impl_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
)

func TestChubAI_Import(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://chub.ai/characters/Anonymous/test-f26406a9718a").
		AssertNoErr().
		Source(source.ChubAI).
		NormalizedURL("chub.ai/characters/Anonymous/test-f26406a9718a").
		DirectURL("chub.ai/characters/Anonymous/test-f26406a9718a").
		CharacterPlatformID("3186564").
		CharacterID("Anonymous/test-f26406a9718a").
		Name("ChatName").
		Title("Test").
		Tagline("Tagline").
		CreateTime(1737810077000000000).
		UpdateTime(1747541869000000000).
		IsForked(false).
		TagNames("ChubAI", "Female").
		Nickname("Anonymous").
		Username("anonymous").
		CreatorPlatformID("90").
		BookUpdateTime(1747541869000000000).
		SheetRevision(character.RevisionV2).
		SheetDescription("Description").
		SheetPersonality("").
		SheetScenario("Scenario").
		SheetFirstMessage("FirstMessage").
		SheetMessageExamples("ExampleDialog").
		SheetCreatorNotes("Tagline\n\nCreatorNotes").
		SheetSystemPrompt("SystemPrompt").
		SheetPostHistoryInstructions("PostHistoryInstructions").
		SheetAlternateGreetingsCount(1).
		SheetTagNames("ChubAI", "Female").
		SheetHasCharacterBook().
		SheetBookName("LoreBook").
		SheetBookDescription("LoreBookDescription").
		SheetBookEntryCount(1).
		SheetBookEntryKeyCount(0, 1).
		SheetBookEntrySecondaryKeyCount(0, 1).
		SheetBookEntryName(0, "LoreBookEntry").
		SheetBookEntryContent(0, "LoreEntryContent").
		SheetDepthPromptPrompt("CharacterNote").
		SheetDepthPromptDepth(0).
		SheetNickname("ChatName").
		SheetCreator("Anonymous").
		Consistent().
		AssertImage()
}

func TestChubAI_Import_OneMultiLoreBook(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://chub.ai/characters/statuotw/konako-the-adventurer-5c21dcc2").
		AssertNoErr().
		NormalizedURL("chub.ai/characters/statuotw/konako-the-adventurer-5c21dcc2").
		CharacterID("statuotw/konako-the-adventurer-5c21dcc2").
		CharacterPlatformID("2072277").
		SheetBookName("The Fantasy World of Adolion").
		SheetBookDescriptionNotEmpty().
		SheetBookEntryCount(212).
		IsForked(false).
		Consistent().
		AssertImage()
}

func TestChubAI_Import_AuxiliaryLoreBook(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://chub.ai/characters/Enkob/maxine-flirty-warrior-of-the-explorers-guild-d1eb387be026").
		AssertNoErr().
		SheetBookName("Maxine -- The fantasy world of Runa (Base Lore)").
		SheetBookEntryCount(129).
		IsForked(false).
		Consistent().
		AssertImage()
}

func TestChubAI_MissingImage(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://chub.ai/characters/Sugondees/saeko-490390c0ebee").
		AssertNoErr().
		IsForked(false).
		Consistent().
		AssertImage()
}

func TestChubAI_BrokenLorebook(t *testing.T) {
	FetchAndAssert(t, "https://chub.ai/characters/Decent_Coast/your-older-orc-step-sister-61c8107a3862").
		AssertNoErr().
		SheetBookEntryCount(57).
		IsForked(false).
		Consistent().
		AssertImage()
}

func TestChubAI_Forked(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://chub.ai/characters/Celestial_Coomer/fried-your-little-neighbor-was-left-in-your-care-0617b5c66b90").
		AssertNoErr().
		IsForked(true).
		Consistent().
		AssertImage()
}
