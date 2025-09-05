package fetcher_test

import (
	"testing"
	"time"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/stretchr/testify/assert"
)

const (
	TestCardSpec                      = character.SpecV2
	TestCardVersion                   = character.V2
	TestCardName                      = "Test"
	TestCardChatName                  = "ChatName"
	TestCardDescription               = "Description"
	TestCardPersonality               = "Personality"
	TestCardScenario                  = "Scenario"
	TestCardFirstMessage              = "FirstMessage"
	TestCardMessageExamples           = "ExampleDialog"
	TestCardTagline                   = "Tagline"
	TestCardCreatorNotes              = "CreatorNotes"
	TestCardSystemPrompt              = "SystemPrompt"
	TestCardPostHistoryInstructions   = "PostHistoryInstructions"
	TestCardAlternateGreetings        = 1
	TestCardAlternateGreeting         = "AlternateGreeting"
	TestCardLoreBookName              = "LoreBook"
	TestCardLoreBookDescription       = "LoreBookDescription"
	TestCardLoreBookEntries           = 1
	TestCardLoreBookEntryName         = "LoreBookEntry"
	TestCardLoreBookEntryComment      = "LoreEntryComment"
	TestCardLoreBookEntryPrimaryKey   = "primary"
	TestCardLoreBookEntrySecondaryKey = "secondary"
	TestCardLoreBookEntryContent      = "LoreEntryContent"
	TestCardLorebookEntryKeys         = 1
	TestCardTags                      = 1
	TestCardTag                       = "Female"
	TestCardDepthPromptContent        = "CharacterNote"
	TestCardDepthPromptLevel          = 4
)

var testClient = reqx.NewRetryClient(reqx.ClientOptions{
	RetryCount:    4,
	MinBackoff:    10 * time.Millisecond,
	MaxBackoff:    500 * time.Millisecond,
	Impersonation: reqx.Chrome,
})

func assertCommonFields(
	t *testing.T, metadata *models.Metadata,
	jsonCard *character.Sheet,
	source source.ID,
	normalizedUrl string,
	creator string,
	description string,
	personality string,
	systemPrompt string,
	characterVersion string,
	creatorNotes string,
	spec character.Spec,
	version character.Version,
) {
	assert.Equal(t, source, metadata.Source)
	assert.Equal(t, normalizedUrl, metadata.CardURL)
	assert.Equal(t, TestCardTags, len(metadata.Tags))
	assert.Equal(t, TestCardTag, metadata.Tags[0].Name)
	assert.Equal(t, creator, metadata.Creator)
	assert.Equal(t, TestCardName, metadata.CardName)

	assert.Equal(t, spec, jsonCard.Spec)
	assert.Equal(t, version, jsonCard.Version)
	assert.Equal(t, TestCardName, jsonCard.Data.CardName)
	assert.Equal(t, TestCardChatName, jsonCard.Data.CharacterName)
	assert.Equal(t, description, jsonCard.Data.Description)
	assert.Equal(t, personality, jsonCard.Data.Personality)
	assert.Equal(t, TestCardScenario, jsonCard.Data.Scenario)
	assert.Equal(t, TestCardFirstMessage, jsonCard.Data.FirstMessage)
	assert.Equal(t, TestCardMessageExamples, jsonCard.Data.MessageExamples)
	assert.Equal(t, creatorNotes, jsonCard.Data.CreatorNotes)
	assert.Equal(t, systemPrompt, jsonCard.Data.SystemPrompt)
	assert.Equal(t, TestCardPostHistoryInstructions, jsonCard.Data.PostHistoryInstructions)
	assert.Equal(t, TestCardAlternateGreetings, len(jsonCard.Data.AlternateGreetings))
	assert.Equal(t, TestCardAlternateGreeting, jsonCard.Data.AlternateGreetings[0])
	assert.Equal(t, TestCardTags, len(jsonCard.Data.Tags))
	assert.Equal(t, TestCardTag, jsonCard.Data.Tags[0])
	assert.Equal(t, characterVersion, jsonCard.Data.CharacterVersion)
	assert.Equal(t, creator, jsonCard.Data.Creator)
}

func assertCharacterLoreBookCommonFields(t *testing.T, jsonCard *character.Sheet) {
	assert.Equal(t, TestCardLoreBookName, *jsonCard.Data.CharacterBook.Name)
	assert.Equal(t, TestCardLoreBookEntries, len(jsonCard.Data.CharacterBook.Entries))
	assert.Equal(t, TestCardLoreBookDescription, *jsonCard.Data.CharacterBook.Description)
	assert.Equal(t, TestCardLoreBookEntryName, *jsonCard.Data.CharacterBook.Entries[0].Name)
	assert.Equal(t, TestCardLoreBookEntryContent, jsonCard.Data.CharacterBook.Entries[0].Content)
	assert.Equal(t, TestCardLorebookEntryKeys, len(jsonCard.Data.CharacterBook.Entries[0].Keys))
	assert.Equal(t, TestCardLoreBookEntryPrimaryKey, jsonCard.Data.CharacterBook.Entries[0].Keys[0])
	assert.Equal(t, TestCardLorebookEntryKeys, len(jsonCard.Data.CharacterBook.Entries[0].SecondaryKeys))
	assert.Equal(t, TestCardLoreBookEntrySecondaryKey, jsonCard.Data.CharacterBook.Entries[0].SecondaryKeys[0])
}

func assertDepthPrompt(t *testing.T, jsonCard *character.Sheet) {
	depthPrompt := jsonCard.Data.DepthPrompt
	assert.Equal(t, TestCardDepthPromptContent, depthPrompt.Prompt)
	assert.Equal(t, TestCardDepthPromptLevel, depthPrompt.Depth)
}

func assertImage(t *testing.T, context *png.CharacterCard) {
	rawContext, err := context.Encode()
	assert.NoError(t, err)
	image, err := rawContext.Image()
	assert.NoError(t, err)
	assert.NotNil(t, image)
	_, err = rawContext.ToBytes()
	assert.NoError(t, err)
}

func assertConsistency(t *testing.T, metadata *models.Metadata, context *png.CharacterCard) {
	stringTags := models.TagsToNames(metadata.Tags)
	assert.Equal(t, stringTags, context.Sheet.Data.Tags)
	assert.Equal(t, metadata.CharacterName, context.Sheet.Data.CharacterName)
	assert.Equal(t, metadata.Creator, context.Sheet.Data.Creator)
	assert.Equal(t, timestamp.Convert[timestamp.Seconds](metadata.CreateTime), context.Sheet.Data.CreationDate)
	assert.Equal(t, timestamp.Convert[timestamp.Seconds](metadata.LatestUpdateTime()), context.Sheet.Data.ModificationDate)
	assert.Equal(t, metadata.CardName, context.Sheet.Data.CardName)
	assert.NotEmpty(t, context.Sheet.Data.Nickname)
}

func assertSourceIsUp(t *testing.T, fetcher fetcher.Fetcher) {
	assert.True(t, fetcher.IsSourceUp(testClient))
}
