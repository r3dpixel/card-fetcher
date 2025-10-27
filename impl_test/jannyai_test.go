package fetcher_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
)

func TestJannyAIImport(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://jannyai.com/characters/421439ad-de63-4448-bc9b-c2c75cedb0af_character-amber-hawthorn").
		AssertNoErr().
		Source(source.JannyAI).
		NormalizedURL("jannyai.com/characters/421439ad-de63-4448-bc9b-c2c75cedb0af").
		DirectURL("jannyai.com/characters/421439ad-de63-4448-bc9b-c2c75cedb0af").
		CharacterPlatformID("421439ad-de63-4448-bc9b-c2c75cedb0af").
		CharacterID("421439ad-de63-4448-bc9b-c2c75cedb0af").
		Name("Amber Hawthorn").
		Title("Amber Hawthorn").
		TaglinePrefix("").
		CreateTime(1727063067584588000).
		UpdateTime(1761523200000000000).
		IsForked(false).
		TagCount(7).
		TagContains("Female", "Dominant").
		Nickname("Katrealynne").
		Username("Katrealynne").
		CreatorPlatformID("2613e569-12f0-4302-9037-b9c5b7a65ce1").
		BookUpdateTime(0).
		SheetRevision(character.RevisionV2).
		SheetDescriptionContains("Character: {{char}}\nAge: 27\nGender: Female\nAppearance: red eyes, blonde hair,").
		SheetPersonality(stringsx.Empty).
		SheetScenarioContains("In this role play, {{char}} must fully embrace her role as {{user}}'s bully, hurling insults and scathing remarks while doing everything in her power to humiliate {{user}}..").
		SheetFirstMessageContains("*Amber lounged against the wall in the dimly lit kitchen, her fingers casually tracing the rim of her glass").
		SheetMessageExamplesContains("\"You pathetic little slut, on your knees! Crawl over here and worship my boots.\"\n\"Look at you, pathetically begging for my attention like a dog in heat.").
		SheetCreatorNotesContains(`<p style="text-align: center">Your bully.</p><hr><p style="text-align: center"><span style="color: #f41111">Dead Dove tag for the following: bullying, degradation`).
		SheetSystemPrompt(stringsx.Empty).
		SheetPostHistoryInstructions(stringsx.Empty).
		SheetAlternateGreetingsCount(0).
		SheetTagCount(7).
		SheetTagContains("Dominant", "Female").
		SheetNoCharacterBook().
		SheetDepthPromptPrompt(stringsx.Empty).
		SheetDepthPromptDepth(0).
		SheetNickname("Amber Hawthorn").
		SheetCreator("Katrealynne").
		Consistent().
		AssertImage()
}
