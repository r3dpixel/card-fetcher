package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
)

func TestCharacterTavernImport(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://character-tavern.com/character/redpixel/test").
		AssertNoErr().
		Source(source.CharacterTavern).
		NormalizedURL("character-tavern.com/character/redpixel/test").
		DirectURL("character-tavern.com/character/redpixel/test").
		CharacterPlatformID("CT_bf4c92dee1f7690c28589360ee3f1380").
		CharacterID("redpixel/test").
		Name("ChatName").
		Title("Test").
		Tagline("TaglineDescription").
		CreateTime(1737895009100000000).
		UpdateTime(1737944149051000000).
		TagNames("Female").
		Nickname("redpixel").
		Username("redpixel").
		CreatorPlatformID("obu6tvoquhr22npm").
		BookUpdateTime(0).
		SheetRevision(character.RevisionV3).
		SheetDescription("Description AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").
		SheetPersonality("Personality").
		SheetScenario("Scenario").
		SheetFirstMessage("FirstMessage").
		SheetMessageExamples("ExampleDialog").
		SheetCreatorNotes("TaglineDescription\n\nCreatorNotes").
		SheetSystemPrompt(stringsx.Empty).
		SheetPostHistoryInstructions("PostHistoryInstructions").
		SheetAlternateGreetingsCount(1).
		SheetTagNames("Female").
		SheetNoCharacterBook().
		SheetDepthPromptPrompt(stringsx.Empty).
		SheetDepthPromptDepth(0).
		SheetNickname("ChatName").
		SheetCreator("redpixel").
		Consistent().
		AssertImage()
}

func TestCharacterTavernImport_CCV3(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://character-tavern.com/character/tidyup/beth_homeless_on_her_birthday_").
		AssertNoErr().
		Source(source.CharacterTavern).
		SheetRevision(character.RevisionV3).
		Consistent().
		AssertImage()
}

func TestCharacterTavernImportFail(t *testing.T) {
	FetchAndAssert(t, "character-tavern.com/character/brian007/lara_s").
		AssertErr()
}

func TestCharacterTavernImportNotes(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://character-tavern.com/character/animatedspell/Violete%20V4").
		AssertNoErr().
		Source(source.CharacterTavern).
		SheetCreator("animatedspell").
		SheetCreatorNotes("Your shapeshifting monster waifu—sweet as sin, twice as wicked. A devoted lover by day, an Encyclopedia-inspired hentai experiment by night.\n\nViolete is a bot with lots of kinky tag, if you don't like any of her suggestion during chat, just gently asked her she will comply (tentacle, Insect,monster, monster girl, etc)").
		Consistent().
		AssertImage()
}

func TestCharacterTavernPngData(t *testing.T) {
	FetchAndAssert(t, "https://character-tavern.com/character/wicked_ali/Veronica").
		AssertNoErr().
		Source(source.CharacterTavern).
		Consistent().
		AssertImage()
}
