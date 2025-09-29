package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
)

func TestPephopImport(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee_character-jessie").
		AssertNoErr().
		Source(source.PepHop).
		NormalizedURL("pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee").
		DirectURL("pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee").
		CharacterPlatformID("75882045-96ef-41eb-bb23-ca1a3fe67aee").
		CharacterID("75882045-96ef-41eb-bb23-ca1a3fe67aee").
		Name("Jessie").
		Title("Jessie").
		TaglinePrefix("Fiery tomboy with top grades but almost expelled for smoking! 🚬😎 She's your tutor").
		CreateTime(1689150440946000000).
		UpdateTime(1730380506096000000).
		TagCount(4).
		TagContains("Female", "Dominant").
		Nickname("TechWhiz").
		Username("TechWhiz").
		CreatorPlatformID("5734b8bf-67d0-4cbd-9873-beec12b45b17").
		BookUpdateTime(0).
		SheetRevision(character.RevisionV2).
		SheetDescriptionContains("Jessie Hayes").
		SheetPersonality(stringsx.Empty).
		SheetScenarioContains("{{char}} must tutor {{user}}").
		SheetFirstMessageContains("*You didn't even realize you had the worst grades").
		SheetMessageExamplesContains(`{{user}}: "Stop yelling."`).
		SheetCreatorNotesContains("Fiery tomboy with top grades").SheetCreatorNotesContains("Jessie's tutoring package?").
		SheetSystemPrompt(stringsx.Empty).
		SheetPostHistoryInstructions(stringsx.Empty).
		SheetAlternateGreetingsCount(0).
		SheetTagCount(4).
		SheetTagContains("Dominant", "Female").
		SheetNoCharacterBook().
		SheetDepthPromptPrompt(stringsx.Empty).
		SheetDepthPromptDepth(0).
		SheetNickname("Jessie").
		SheetCreator("TechWhiz").
		Consistent().
		AssertImage()
}
