package fetcher_test

import (
	"strings"
	"testing"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/postprocessor"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
)

const (
	testPygmalionService string = "pygmalion"
)

var testPygmalionCredManager = cred.NewManager(testPygmalionService, cred.Env)
var testPygmalionFetcher = postprocessor.New(fetcher.NewPygmalionFetcher(testPygmalionCredManager))

func TestPygmalionFetcher(t *testing.T) {
	assertSourceIsUp(t, testPygmalionFetcher)
}

func TestPygmalionImport(t *testing.T) {
	t.Parallel()
	const creator = "Bronya Rand"
	url := "https://pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5"
	fetcherTask := task.New(testClient, testPygmalionFetcher, url, testPygmalionFetcher.MainURL())
	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	assertConsistency(t, metadata, card)
	assert.NotNil(t, jsonSheet)
	assert.Equal(t, "pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5", metadata.CardURL)
	assert.Equal(t, "pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5", metadata.DirectURL)
	assert.Equal(t, "d47f2f4e-0263-49f8-b872-d8fd7588dbb5", metadata.CharacterID)
	assert.Equal(t, metadata.CharacterID, metadata.PlatformID)
	assert.Nil(t, jsonSheet.Data.CharacterBook.Entries[0].Comment)

	assert.Equal(t, source.Pygmalion, metadata.Source)
	assert.Equal(t, testPygmalionFetcher.NormalizeURL(metadata.CharacterID), metadata.CardURL)
	assert.Equal(t, 4, len(metadata.Tags))
	assert.Equal(t, "Ali:Chat", metadata.Tags[0].Name)
	assert.Equal(t, creator, metadata.Creator)
	assert.Equal(t, "Veliona", metadata.CardName)

	assert.Equal(t, TestCardSpec, jsonSheet.Spec)
	assert.Equal(t, TestCardVersion, jsonSheet.Version)
	assert.Equal(t, "Veliona", jsonSheet.Data.CardName)
	assert.Equal(t, "Veliona", jsonSheet.Data.CharacterName)
	assert.True(t, strings.HasPrefix(jsonSheet.Data.Description,
		"{{user}}: Brief introduction?"))
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.Personality)
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.Scenario)
	assert.True(t, strings.HasPrefix(jsonSheet.Data.FirstMessage,
		"*A young woman with long black-red hair is seen with a hand brought up to her chin"))
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.MessageExamples)
	assert.True(t, strings.HasPrefix(jsonSheet.Data.CreatorNotes,
		"Seele Vollerei's Alter Ego (Scenario A). Made by Bronya Rand and TheWandering514."))
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.SystemPrompt)
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.PostHistoryInstructions)
	assert.Equal(t, 0, len(jsonSheet.Data.AlternateGreetings))
	assert.Equal(t, 4, len(jsonSheet.Data.Tags))
	assert.Equal(t, "Ali:Chat", jsonSheet.Data.Tags[0])
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.CharacterVersion)
	assert.Equal(t, creator, jsonSheet.Data.Creator)

	assert.Equal(t, "Honkai Impact 3rd Lorebook"+character.BookNameSeparator+"Veliona-WI", *jsonSheet.Data.CharacterBook.Name)
	assert.Equal(t, 37, len(jsonSheet.Data.CharacterBook.Entries))
	assert.Equal(t, "A small collection of entries of Honkai Impact 3rd for Bronya Rand's Bronya Zaychik and Veliona bots."+
		character.BookDescriptionSeparator+"A lorebook for Bronya Rand's and The Wandering514's Veliona (Scenarios A and B).",
		*jsonSheet.Data.CharacterBook.Description)
	assert.Equal(t, stringsx.Empty, *jsonSheet.Data.CharacterBook.Entries[0].Name)
	assert.Equal(t,
		"[ Honkai: lead by the Will of Honkai, forms(Herrschers, Honkai Beasts, Honkai Sickness, Honkai Energy) ]",
		jsonSheet.Data.CharacterBook.Entries[0].Content)
	assert.Equal(t, 1, len(jsonSheet.Data.CharacterBook.Entries[0].Keys))
	assert.Equal(t, "Honkai", jsonSheet.Data.CharacterBook.Entries[0].Keys[0])
	assert.Equal(t, 0, len(jsonSheet.Data.CharacterBook.Entries[0].SecondaryKeys))
}

func TestPygmalionDepthPrompt(t *testing.T) {
	t.Parallel()

	url := "https://pygmalion.chat/character/f5e311e1-3815-45f5-8bdc-4d728150cf38"
	fetcherTask := task.New(testClient, testPygmalionFetcher, url, testPygmalionFetcher.MainURL())
	_, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	depthPrompt := jsonSheet.Data.DepthPrompt

	assert.True(t, strings.HasPrefix(
		depthPrompt.Prompt,
		"[ Writing Style: {{char}} speaks sharply, sometimes provocatively, but never outright cruel. Her rivalry with {{user}} is clear, but there are moments where she struggles to maintain her composure. ]",
	))
	assert.Equal(t, character.DefaultDepthPromptLevel, depthPrompt.Depth)
}
