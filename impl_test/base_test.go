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

var testClient = reqx.NewRetryClient(reqx.Options{
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
	assert.Equal(t, normalizedUrl, metadata.NormalizedURL)
	assert.Equal(t, TestCardTags, len(metadata.Tags))
	assert.Equal(t, TestCardTag, metadata.Tags[0].Name)
	assert.Equal(t, creator, metadata.CreatorInfo.Nickname)
	assert.Equal(t, TestCardName, metadata.Title)

	assert.Equal(t, spec, jsonCard.Spec)
	assert.Equal(t, version, jsonCard.Version)
	assert.Equal(t, TestCardName, jsonCard.Content.Title)
	assert.Equal(t, TestCardChatName, jsonCard.Content.Name)
	assert.Equal(t, description, jsonCard.Content.Description)
	assert.Equal(t, personality, jsonCard.Content.Personality)
	assert.Equal(t, TestCardScenario, jsonCard.Content.Scenario)
	assert.Equal(t, TestCardFirstMessage, jsonCard.Content.FirstMessage)
	assert.Equal(t, TestCardMessageExamples, jsonCard.Content.MessageExamples)
	assert.Equal(t, creatorNotes, jsonCard.Content.CreatorNotes)
	assert.Equal(t, systemPrompt, jsonCard.Content.SystemPrompt)
	assert.Equal(t, TestCardPostHistoryInstructions, jsonCard.Content.PostHistoryInstructions)
	assert.Equal(t, TestCardAlternateGreetings, len(jsonCard.Content.AlternateGreetings))
	assert.Equal(t, TestCardAlternateGreeting, jsonCard.Content.AlternateGreetings[0])
	assert.Equal(t, TestCardTags, len(jsonCard.Content.Tags))
	assert.Equal(t, TestCardTag, jsonCard.Content.Tags[0])
	assert.Equal(t, characterVersion, jsonCard.Content.CharacterVersion)
	assert.Equal(t, creator, jsonCard.Content.Creator)
}

func assertCharacterLoreBookCommonFields(t *testing.T, jsonCard *character.Sheet) {
	assert.Equal(t, TestCardLoreBookName, *jsonCard.Content.CharacterBook.Name)
	assert.Equal(t, TestCardLoreBookEntries, len(jsonCard.Content.CharacterBook.Entries))
	assert.Equal(t, TestCardLoreBookDescription, *jsonCard.Content.CharacterBook.Description)
	assert.Equal(t, TestCardLoreBookEntryName, *jsonCard.Content.CharacterBook.Entries[0].Name)
	assert.Equal(t, TestCardLoreBookEntryContent, jsonCard.Content.CharacterBook.Entries[0].Content)
	assert.Equal(t, TestCardLorebookEntryKeys, len(jsonCard.Content.CharacterBook.Entries[0].Keys))
	assert.Equal(t, TestCardLoreBookEntryPrimaryKey, jsonCard.Content.CharacterBook.Entries[0].Keys[0])
	assert.Equal(t, TestCardLorebookEntryKeys, len(jsonCard.Content.CharacterBook.Entries[0].SecondaryKeys))
	assert.Equal(t, TestCardLoreBookEntrySecondaryKey, jsonCard.Content.CharacterBook.Entries[0].SecondaryKeys[0])
}

func assertDepthPrompt(t *testing.T, jsonCard *character.Sheet) {
	depthPrompt := jsonCard.Content.DepthPrompt
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
	assert.Equal(t, stringTags, context.Sheet.Content.Tags)
	assert.Equal(t, metadata.Name, context.Sheet.Content.Name)
	assert.Equal(t, metadata.CreatorInfo.Nickname, context.Sheet.Content.Creator)
	assert.Equal(t, timestamp.Convert[timestamp.Seconds](metadata.CreateTime), context.Sheet.Content.CreationDate)
	assert.Equal(t, timestamp.Convert[timestamp.Seconds](metadata.LatestUpdateTime()), context.Sheet.Content.ModificationDate)
	assert.Equal(t, metadata.Title, context.Sheet.Content.Title)
	assert.NotEmpty(t, context.Sheet.Content.Nickname)
}

func assertSourceIsUp(t *testing.T, fetcher fetcher.SourceHandler) {
	assert.True(t, fetcher.IsSourceUp())
}
