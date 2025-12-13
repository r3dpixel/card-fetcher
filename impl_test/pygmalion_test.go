package impl_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
)

func TestPygmalionImport(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5").
		AssertPygmalionCredentials().
		AssertNoErr().
		Source(source.Pygmalion).
		NormalizedURL("pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5").
		DirectURL("pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5").
		CharacterPlatformID("d47f2f4e-0263-49f8-b872-d8fd7588dbb5").
		CharacterID("d47f2f4e-0263-49f8-b872-d8fd7588dbb5").
		Name("Veliona").
		Title("Veliona").
		TaglinePrefix("Seele Vollerei's Alter Ego (Scenario A). Made by Bronya Rand and TheWandering514. ").
		CreateTime(1706328712000000000).
		UpdateTime(1722520534000000000).
		IsForked(false).
		IsForked(false).
		TagCount(5).
		TagContains("Pygmalion", "Ali:Chat").
		Nickname("Bronya Rand").
		Username("bronya_rand").
		CreatorPlatformID("7f0482b0-cb7b-432f-9837-647e73ea19c4").
		BookUpdateTime(1737777979000000000).
		SheetRevision(character.RevisionV2).
		SheetDescriptionPrefix("{{user}}: Brief introduction?").
		SheetPersonality("").
		SheetScenario("").
		SheetFirstMessagePrefix("*A young woman with long black-red hair is seen with a hand brought up to her chin").
		SheetMessageExamples("").
		SheetCreatorNotesPrefix("Seele Vollerei's Alter Ego (Scenario A). Made by Bronya Rand and TheWandering514.").
		SheetSystemPrompt("").
		SheetPostHistoryInstructions("").
		SheetAlternateGreetingsCount(0).
		SheetTagCount(5).
		SheetTagContains("Pygmalion", "Ali:Chat").
		SheetHasCharacterBook().
		SheetBookNameContains("Honkai Impact 3rd Lorebook").
		SheetBookNameContains(character.BookNameSeparator).
		SheetBookNameContains("Veliona-WI").
		SheetBookEntryCount(37).
		SheetBookDescriptionContains("A small collection of entries of Honkai Impact 3rd for Bronya Rand's Bronya Zaychik and Veliona bots.").
		SheetBookDescriptionContains(character.BookDescriptionSeparator).
		SheetBookDescriptionContains("A lorebook for Bronya Rand's and The Wandering514's Veliona (Scenarios A and B).").
		SheetBookEntryName(0, "").
		SheetBookEntryContent(0, "[ Honkai: lead by the Will of Honkai, forms(Herrschers, Honkai Beasts, Honkai Sickness, Honkai Energy) ]").
		SheetBookEntryKeyCount(0, 1).
		SheetBookEntryPrimaryKey(0, "Honkai").
		SheetBookEntrySecondaryKeyCount(0, 0).
		SheetDepthPromptPrompt("").
		SheetDepthPromptDepth(0).
		SheetNickname("Veliona").
		SheetCreator("Bronya Rand").
		Consistent().
		AssertImage()
}

func TestPygmalionDepthPrompt(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://pygmalion.chat/character/f5e311e1-3815-45f5-8bdc-4d728150cf38").
		AssertPygmalionCredentials().
		AssertNoErr().
		Source(source.Pygmalion).
		SheetDepthPromptPromptPrefix("[ Writing Style: {{char}} speaks sharply, sometimes provocatively, but never outright cruel. Her rivalry with {{user}} is clear, but there are moments where she struggles to maintain her composure. ]").
		SheetDepthPromptDepth(character.DefaultDepth).
		IsForked(false).
		Consistent().
		AssertImage()
}
