package impl_test

import (
	"testing"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
)

func TestAiccImport_Tags(t *testing.T) {
	t.Parallel()
	FetchAndAssert(t, "https://aicharactercards.com/charactercards/confession/linux4life/nora/").
		AssertNoErr().
		Source(source.AICC).
		NormalizedURL("aicharactercards.com/charactercards/confession/linux4life/nora").
		DirectURL("aicharactercards.com/charactercards/confession/linux4life/nora").
		CharacterPlatformID("10844").
		CharacterID("confession/linux4life/nora").
		Name("Nora").
		Title("Nora").
		TaglinePrefix("Nora is a 34-year-old reclusive novelist celebrated for her ").
		CreateTime(1764349932000000000).
		UpdateTime(1764349932000000000).
		IsForked(false).
		TagCount(9).
		TagContains("AICC", "Female", "Introvert").
		Nickname("Linux4life").
		Username("Linux4life").
		CreatorPlatformID("Linux4life").
		BookUpdateTime(0).
		SheetRevision(character.RevisionV2).
		SheetDescriptionContains("She wears a simple white blouse with lace-trimmed collar,").
		SheetDescriptionContains("Expressive blue eyes, framed by thin wire-rimmed glasses perched").
		SheetPersonalityPrefix("{{char}} is a 34-year-old reclusive novelist").
		SheetScenarioPrefix("In the heart of a snow-swept Victorian manor on a bitter winter's eve, {{char}}'s private study glows").
		SheetFirstMessagePrefix("*The fire crackles in the hearth, its golden flames licking the logs with a hunger that mirrors the slow").
		SheetMessageExamplesPrefix("<START>\r\n{{user}}: *I take  your hand gently, pulling you closer* \"Nora, let me be the one to show you what your words only whisper.").
		SheetCreatorNotesContains("Nora is a 34-year-old reclusive novelist celebrated for her").
		SheetSystemPrompt("").
		SheetPostHistoryInstructions("").
		SheetAlternateGreetingsCount(1).
		SheetTagCount(9).
		TagContains("AICC", "Female", "Introvert").
		SheetNoCharacterBook().
		SheetDepthPromptPrompt("").
		SheetDepthPromptDepth(0).
		SheetNickname("Nora").
		SheetCreator("Linux4life").
		Consistent().
		AssertImage()
}
