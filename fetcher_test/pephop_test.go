package fetcher_test

import (
	"strings"
	"testing"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/postprocessor"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
)

var testPephopFetcher = postprocessor.New(fetcher.NewPephopFetcher())

func TestPephopFetcher(t *testing.T) {
	assertSourceIsUp(t, testPephopFetcher)
}

func TestPephopImport(t *testing.T) {
	t.Parallel()

	const creator = "TechWhiz"
	const cardName = "Jessie"
	url := "https://pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee_character-jessie"
	fetcherTask := task.New(testClient, testPephopFetcher, url, testPephopFetcher.MainURL())

	metadata, err := fetcherTask.FetchMetadata()
	assert.NoError(t, err)
	card, err := fetcherTask.FetchCharacterCard()
	assert.NoError(t, err)
	jsonSheet := card.Sheet

	metadata, err = fetcherTask.FetchMetadata()
	assert.NoError(t, err)

	assertConsistency(t, metadata, card)
	assert.Equal(t, "pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee", metadata.CardURL)
	assert.Equal(t, "pephop.ai/characters/75882045-96ef-41eb-bb23-ca1a3fe67aee", metadata.DirectURL)
	assert.Equal(t, "75882045-96ef-41eb-bb23-ca1a3fe67aee", metadata.CharacterID)
	assert.Equal(t, metadata.CharacterID, metadata.PlatformID)

	assert.Equal(t, testPephopFetcher.SourceID(), metadata.Source)
	assert.Equal(t, testPephopFetcher.NormalizeURL(metadata.CharacterID), metadata.CardURL)
	assert.Equal(t, 4, len(metadata.Tags))
	assert.Equal(t, "Dominant", metadata.Tags[0].Name)
	assert.Equal(t, "Female", metadata.Tags[1].Name)
	assert.Equal(t, creator, metadata.Creator)
	assert.Equal(t, cardName, metadata.CardName)

	assert.Equal(t, TestCardSpec, jsonSheet.Spec)
	assert.Equal(t, TestCardVersion, jsonSheet.Version)
	assert.Equal(t, cardName, jsonSheet.Data.CardName)
	assert.Equal(t, cardName, jsonSheet.Data.CharacterName)
	assert.True(t, strings.HasPrefix(jsonSheet.Data.Description, "Jessie Hayes"))
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.Personality)
	assert.True(t, strings.HasPrefix(jsonSheet.Data.Scenario, "{{char}} must tutor {{user}}"))
	assert.True(t, strings.HasPrefix(jsonSheet.Data.FirstMessage, `*You didn't even realize you had the worst grades`))
	assert.True(t, strings.HasPrefix(jsonSheet.Data.MessageExamples, `{{user}}: "Stop yelling."`))
	assert.True(t, strings.HasPrefix(jsonSheet.Data.CreatorNotes, "Fiery tomboy with top grades"))
	assert.True(t, strings.HasSuffix(jsonSheet.Data.CreatorNotes, "Jessie's tutoring package?"))
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.SystemPrompt)
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.PostHistoryInstructions)
	assert.Equal(t, 0, len(jsonSheet.Data.AlternateGreetings))
	assert.Equal(t, 4, len(jsonSheet.Data.Tags))
	assert.Equal(t, "Dominant", jsonSheet.Data.Tags[0])
	assert.Equal(t, "Female", jsonSheet.Data.Tags[1])
	assert.Equal(t, stringsx.Empty, jsonSheet.Data.CharacterVersion)
	assert.Equal(t, creator, jsonSheet.Data.Creator)
}
