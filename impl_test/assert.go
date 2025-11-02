package fetcher_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/router"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRouter = router.EnvConfigured()

type CharacterAssertion struct {
	t        *testing.T
	metadata *models.Metadata
	card     *png.CharacterCard
	sheet    *character.Sheet
	err      error
}

func FetchAndAssert(t *testing.T, url string) *CharacterAssertion {
	fetcherTask, ok := testRouter.TaskOf(url)
	assert.True(t, ok, "Failed to find fetcher for url: %s", url)

	metadata, metadataErr := fetcherTask.FetchMetadata()
	if metadataErr != nil {
		return &CharacterAssertion{
			t:        t,
			metadata: nil,
			card:     nil,
			sheet:    nil,
			err:      metadataErr,
		}
	}

	card, cardErr := fetcherTask.FetchCharacterCard()
	if cardErr != nil {
		return &CharacterAssertion{
			t:        t,
			metadata: metadata,
			card:     nil,
			sheet:    nil,
			err:      cardErr,
		}
	}

	return &CharacterAssertion{
		t:        t,
		metadata: metadata,
		card:     card,
		sheet:    card.Sheet,
		err:      nil,
	}
}

func (ca *CharacterAssertion) AssertPygmalionCredentials() *CharacterAssertion {
	requiredEnvVars := []string{
		"PYGMALION_USERNAME",
		"PYGMALION_PASSWORD",
	}

	for _, envVar := range requiredEnvVars {
		if stringsx.IsBlank(os.Getenv(envVar)) {
			assert.Fail(ca.t, "Missing required environment variable: %s", envVar)
		}
	}

	return ca
}

func (ca *CharacterAssertion) AssetJannyAICookie() *CharacterAssertion {
	requiredEnvVars := []string{
		"JANNY_CF_COOKIE",
		"JANNY_USER_AGENT",
	}

	for _, envVar := range requiredEnvVars {
		if stringsx.IsBlank(os.Getenv(envVar)) {
			assert.Fail(ca.t, "Missing required environment variable: %s", envVar)
		}
	}

	return ca
}

func (ca *CharacterAssertion) AssertNoErr() *CharacterAssertion {
	require.NotNil(ca.t, ca.metadata, "metadata cannot be nil")
	require.NotNil(ca.t, ca.card, "card cannot be nil")
	require.NotNil(ca.t, ca.card.Sheet, "card.Sheet cannot be nil")
	require.Nil(ca.t, ca.err, "err should be nil")
	return ca
}

func (ca *CharacterAssertion) AssertErr() *CharacterAssertion {
	require.NotNil(ca.t, ca.err, "err must be present")
	return ca
}

func (ca *CharacterAssertion) Source(expected source.ID) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Source)
	return ca
}

func (ca *CharacterAssertion) NormalizedURL(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.NormalizedURL)
	return ca
}

func (ca *CharacterAssertion) DirectURL(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.DirectURL)
	return ca
}

func (ca *CharacterAssertion) CharacterPlatformID(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.CardInfo.PlatformID)
	return ca
}

func (ca *CharacterAssertion) CharacterID(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.CharacterID)
	return ca
}

func (ca *CharacterAssertion) Name(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Name)
	return ca
}

func (ca *CharacterAssertion) Title(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Title)
	return ca
}

func (ca *CharacterAssertion) Tagline(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Tagline)
	return ca
}

func (ca *CharacterAssertion) CreateTime(expected int64) *CharacterAssertion {
	assert.Equal(ca.t, expected, int64(ca.metadata.CreateTime))
	return ca
}

func (ca *CharacterAssertion) UpdateTime(expected int64) *CharacterAssertion {
	assert.Equal(ca.t, expected, int64(ca.metadata.UpdateTime))
	return ca
}

func (ca *CharacterAssertion) Nickname(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Nickname)
	return ca
}

func (ca *CharacterAssertion) Username(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.Username)
	return ca
}

func (ca *CharacterAssertion) CreatorPlatformID(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.CreatorInfo.PlatformID)
	return ca
}

func (ca *CharacterAssertion) BookUpdateTime(expected int64) *CharacterAssertion {
	assert.Equal(ca.t, expected, int64(ca.metadata.BookUpdateTime))
	return ca
}

func (ca *CharacterAssertion) IsForked(expected bool) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.metadata.IsForked)
	return ca
}

func (ca *CharacterAssertion) SheetName(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Name))
	return ca
}

func (ca *CharacterAssertion) SheetDescription(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Description))
	return ca
}

func (ca *CharacterAssertion) SheetPersonality(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Personality))
	return ca
}

func (ca *CharacterAssertion) SheetScenario(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Scenario))
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessage(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.FirstMessage))
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamples(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.MessageExamples))
	return ca
}

func (ca *CharacterAssertion) SheetSystemPrompt(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.SystemPrompt))
	return ca
}

func (ca *CharacterAssertion) SheetPostHistoryInstructions(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.PostHistoryInstructions))
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotes(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.CreatorNotes))
	return ca
}

func (ca *CharacterAssertion) SheetCreator(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Creator))
	return ca
}

func (ca *CharacterAssertion) SheetTitle(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Title))
	return ca
}

func (ca *CharacterAssertion) SheetSpec(expected character.Spec) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.sheet.Spec)
	return ca
}

func (ca *CharacterAssertion) SheetVersion(expected character.Version) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.sheet.Version)
	return ca
}

func (ca *CharacterAssertion) CreationDate(expected int64) *CharacterAssertion {
	assert.Equal(ca.t, expected, int64(ca.sheet.Content.CreationDate))
	return ca
}

func (ca *CharacterAssertion) ModificationDate(expected int64) *CharacterAssertion {
	assert.Equal(ca.t, expected, int64(ca.sheet.Content.ModificationDate))
	return ca
}

func (ca *CharacterAssertion) TagCount(expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.metadata.Tags, expected)
	return ca
}

func (ca *CharacterAssertion) SheetTagCount(expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.sheet.Tags, expected)
	return ca
}

func (ca *CharacterAssertion) SheetAlternateGreetingsCount(expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.sheet.AlternateGreetings, expected)
	return ca
}

func (ca *CharacterAssertion) SheetDescriptionContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.Description), substring, "Description should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetPersonalityContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.Personality), substring, "Personality should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetScenarioContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.Scenario), substring, "Scenario should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessageContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.FirstMessage), substring, "FirstMessage should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamplesContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.MessageExamples), substring, "MessageExamples should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetSystemPromptContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.SystemPrompt), substring, "SystemPrompt should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetPostHistoryInstructionsContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.PostHistoryInstructions), substring, "PostHistoryInstructions should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotesContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.CreatorNotes), substring, "CreatorNotes should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) TaglineContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, ca.metadata.Tagline, substring, "Tagline should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetBookNameContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.CharacterBook.Name), substring, "Book name should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) SheetBookDescriptionContains(substring string) *CharacterAssertion {
	assert.Contains(ca.t, string(ca.sheet.CharacterBook.Description), substring, "Book description should contain: %s", substring)
	return ca
}

func (ca *CharacterAssertion) Tag(tag string) *CharacterAssertion {
	found := false
	for _, t := range ca.metadata.Tags {
		if t.Name == tag {
			found = true
			break
		}
	}
	assert.True(ca.t, found, "Expected metadata to have tag: %s", tag)
	return ca
}

func (ca *CharacterAssertion) SheetTag(tag string) *CharacterAssertion {
	found := false
	for _, t := range ca.sheet.Tags {
		if t == tag {
			found = true
			break
		}
	}
	assert.True(ca.t, found, "Expected sheet to have tag: %s", tag)
	return ca
}

func (ca *CharacterAssertion) TagContains(expectedTags ...string) *CharacterAssertion {
	metadataTags := models.TagsToNames(ca.metadata.Tags)
	tagMap := make(map[string]bool, len(metadataTags))
	for _, tag := range metadataTags {
		tagMap[tag] = true
	}
	for _, expectedTag := range expectedTags {
		assert.True(ca.t, tagMap[expectedTag], "Metadata tags should contain: %s", expectedTag)
	}
	return ca
}

func (ca *CharacterAssertion) SheetTagContains(expectedTags ...string) *CharacterAssertion {
	tagMap := make(map[string]bool, len(ca.sheet.Tags))
	for _, tag := range ca.sheet.Tags {
		tagMap[tag] = true
	}
	for _, expectedTag := range expectedTags {
		assert.True(ca.t, tagMap[expectedTag], "Sheet tags should contain: %s", expectedTag)
	}
	return ca
}

func (ca *CharacterAssertion) TagNames(expectedTags ...string) *CharacterAssertion {
	metadataTags := models.TagsToNames(ca.metadata.Tags)
	assert.ElementsMatch(ca.t, expectedTags, metadataTags, "Metadata tag names mismatch")
	return ca
}

func (ca *CharacterAssertion) SheetTagNames(expectedTags ...string) *CharacterAssertion {
	sheetTags := slices.Clone(ca.sheet.Tags)
	assert.ElementsMatch(ca.t, expectedTags, []string(sheetTags), "Sheet tag names mismatch")
	return ca
}

func (ca *CharacterAssertion) SheetDescriptionPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.Description), prefix), "Description should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetDescriptionSuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.Description), suffix), "Description should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) SheetPersonalityPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.Personality), prefix), "Personality should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetPersonalitySuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.Personality), suffix), "Personality should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) SheetScenarioPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.Scenario), prefix), "Scenario should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetScenarioSuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.Scenario), suffix), "Scenario should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessagePrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.FirstMessage), prefix), "FirstMessage should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessageSuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.FirstMessage), suffix), "FirstMessage should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamplesPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.MessageExamples), prefix), "MessageExamples should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamplesSuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.MessageExamples), suffix), "MessageExamples should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotesPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.CreatorNotes), prefix), "CreatorNotes should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) TaglinePrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(ca.metadata.Tagline, prefix), "Tagline should start with: %s", prefix)

	return ca
}

func (ca *CharacterAssertion) SheetDepthPromptPromptPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(ca.sheet.DepthPrompt.Prompt, prefix))
	return ca
}

func (ca *CharacterAssertion) SheetBookDescriptionPrefix(prefix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasPrefix(string(ca.sheet.CharacterBook.Description), prefix), "Book description should start with: %s", prefix)
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotesSuffix(suffix string) *CharacterAssertion {
	assert.True(ca.t, strings.HasSuffix(string(ca.sheet.CreatorNotes), suffix), "CreatorNotes should end with: %s", suffix)
	return ca
}

func (ca *CharacterAssertion) ValidTimestamps() *CharacterAssertion {
	assert.Positive(ca.t, ca.metadata.CreateTime, "CreateTime should be positive")
	assert.Positive(ca.t, ca.metadata.UpdateTime, "UpdateTime should be positive")
	assert.GreaterOrEqual(ca.t, ca.metadata.UpdateTime, ca.metadata.CreateTime, "UpdateTime should be >= CreateTime")
	return ca
}

func (ca *CharacterAssertion) SheetAlternateGreetingsNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.sheet.AlternateGreetings, "Should have alternate greetings")
	return ca
}

func (ca *CharacterAssertion) TagsNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Tags, "Should have metadata tags")
	return ca
}

func (ca *CharacterAssertion) SheetTagsNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.sheet.Tags, "Should have sheet tags")
	return ca
}

func (ca *CharacterAssertion) SheetHasCharacterBook() *CharacterAssertion {
	assert.NotNil(ca.t, ca.sheet.CharacterBook, "Should have character book")
	return ca
}

func (ca *CharacterAssertion) SheetNoCharacterBook() *CharacterAssertion {
	assert.Nil(ca.t, ca.sheet.CharacterBook, "Should not have character book")
	return ca
}

func (ca *CharacterAssertion) MinTags(min int) *CharacterAssertion {
	assert.GreaterOrEqual(ca.t, len(ca.metadata.Tags), min, "Should have at least %d metadata tags", min)
	return ca
}

func (ca *CharacterAssertion) SheetMinTags(min int) *CharacterAssertion {
	assert.GreaterOrEqual(ca.t, len(ca.sheet.Tags), min, "Should have at least %d sheet tags", min)
	return ca
}

func (ca *CharacterAssertion) SheetMinAlternateGreetings(min int) *CharacterAssertion {
	assert.GreaterOrEqual(ca.t, len(ca.sheet.AlternateGreetings), min, "Should have at least %d alternate greetings", min)
	return ca
}

func (ca *CharacterAssertion) SourceNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.metadata.Source), "Source should not be empty")
	return ca
}

func (ca *CharacterAssertion) NormalizedURLNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.NormalizedURL, "NormalizedURL should not be empty")
	return ca
}

func (ca *CharacterAssertion) DirectURLNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.DirectURL, "DirectURL should not be empty")
	return ca
}

func (ca *CharacterAssertion) CharacterPlatformIDNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.CardInfo.PlatformID, "CharacterPlatformID should not be empty")
	return ca
}

func (ca *CharacterAssertion) CharacterIDNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.CharacterID, "CharacterID should not be empty")
	return ca
}

func (ca *CharacterAssertion) MetadataNameNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Name, "Metadata Name should not be empty")
	return ca
}

func (ca *CharacterAssertion) MetadataTitleNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Title, "Metadata Title should not be empty")
	return ca
}

func (ca *CharacterAssertion) TaglineNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Tagline, "Tagline should not be empty")
	return ca
}

func (ca *CharacterAssertion) NicknameNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Nickname, "Nickname should not be empty")
	return ca
}

func (ca *CharacterAssertion) UsernameNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.Username, "Username should not be empty")
	return ca
}

func (ca *CharacterAssertion) CreatorPlatformIDNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, ca.metadata.CreatorInfo.PlatformID, "CreatorPlatformID should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetNameNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Name), "Name should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetDescriptionNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Description), "Description should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetPersonalityNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Personality), "Personality should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetScenarioNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Scenario), "Scenario should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessageNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.FirstMessage), "FirstMessage should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamplesNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.MessageExamples), "MessageExamples should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetSystemPromptNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.SystemPrompt), "SystemPrompt should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetPostHistoryInstructionsNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.PostHistoryInstructions), "PostHistoryInstructions should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotesNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.CreatorNotes), "CreatorNotes should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Creator), "Creator should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetTitleNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.Title), "Title should not be empty")
	return ca
}

func (ca *CharacterAssertion) SourceEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.metadata.Source), "Source should be empty")
	return ca
}

func (ca *CharacterAssertion) NormalizedURLEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.NormalizedURL, "NormalizedURL should be empty")
	return ca
}

func (ca *CharacterAssertion) DirectURLEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.DirectURL, "DirectURL should be empty")
	return ca
}

func (ca *CharacterAssertion) CharacterPlatformIDEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.CardInfo.PlatformID, "CharacterPlatformID should be empty")
	return ca
}

func (ca *CharacterAssertion) CharacterIDEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.CharacterID, "CharacterID should be empty")
	return ca
}

func (ca *CharacterAssertion) MetadataNameEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Name, "Metadata Name should be empty")
	return ca
}

func (ca *CharacterAssertion) MetadataTitleEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Title, "Metadata Title should be empty")
	return ca
}

func (ca *CharacterAssertion) TaglineEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Tagline, "Tagline should be empty")
	return ca
}

func (ca *CharacterAssertion) NicknameEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Nickname, "Nickname should be empty")
	return ca
}

func (ca *CharacterAssertion) UsernameEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Username, "Username should be empty")
	return ca
}

func (ca *CharacterAssertion) CreatorPlatformIDEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.CreatorInfo.PlatformID, "CreatorPlatformID should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetNameEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Name), "Name should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetDescriptionEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Description), "Description should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetPersonalityEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Personality), "Personality should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetScenarioEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Scenario), "Scenario should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetFirstMessageEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.FirstMessage), "FirstMessage should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetMessageExamplesEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.MessageExamples), "MessageExamples should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetSystemPromptEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.SystemPrompt), "SystemPrompt should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetPostHistoryInstructionsEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.PostHistoryInstructions), "PostHistoryInstructions should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetCreatorNotesEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.CreatorNotes), "CreatorNotes should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetCreatorEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Creator), "Creator should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetTitleEmpty() *CharacterAssertion {
	assert.Empty(ca.t, string(ca.sheet.Title), "Title should be empty")
	return ca
}

func (ca *CharacterAssertion) TagsEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.metadata.Tags, "Metadata tags should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetTagsEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.sheet.Tags, "Sheet tags should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetAlternateGreetingsEmpty() *CharacterAssertion {
	assert.Empty(ca.t, ca.sheet.AlternateGreetings, "AlternateGreetings should be empty")
	return ca
}

func (ca *CharacterAssertion) SheetBookName(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.CharacterBook.Name), "Book name mismatch")
	return ca
}

func (ca *CharacterAssertion) SheetBookDescription(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.CharacterBook.Description), "Book description mismatch")
	return ca
}

func (ca *CharacterAssertion) SheetBookEntryCount(expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.sheet.CharacterBook.Entries, expected, "Book entry count mismatch")
	return ca
}

func (ca *CharacterAssertion) SheetBookEntryKeyCount(index int, expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.sheet.CharacterBook.Entries[index].Keys, expected, "Book entry[%d] key count mismatch", index)
	return ca
}

func (ca *CharacterAssertion) SheetBookEntrySecondaryKeyCount(index int, expected int) *CharacterAssertion {
	assert.Len(ca.t, ca.sheet.CharacterBook.Entries[index].SecondaryKeys, expected, "Book entry[%d] secondary key count mismatch", index)
	return ca
}

func (ca *CharacterAssertion) SheetBookNameNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.CharacterBook.Name), "Book name should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetBookDescriptionNotEmpty() *CharacterAssertion {
	assert.NotEmpty(ca.t, string(ca.sheet.CharacterBook.Description), "Book description should not be empty")
	return ca
}

func (ca *CharacterAssertion) SheetBookEntryName(index int, expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.CharacterBook.Entries[index].Name), "Book entry[%d] name mismatch", index)
	return ca
}

func (ca *CharacterAssertion) SheetBookEntryContent(index int, expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.CharacterBook.Entries[index].Content), "Book entry[%d] content mismatch", index)
	return ca
}

func (ca *CharacterAssertion) SheetBookEntryPrimaryKey(entryIndex int, expected string) *CharacterAssertion {
	require.NotEmpty(ca.t, ca.sheet.CharacterBook.Entries[entryIndex].Keys, "Entry[%d] should have keys", entryIndex)

	found := false
	for _, key := range ca.sheet.CharacterBook.Entries[entryIndex].Keys {
		if key == expected {
			found = true
			break
		}
	}
	assert.True(ca.t, found, "Book entry[%d] should contain primary key: %s", entryIndex, expected)
	return ca
}

func (ca *CharacterAssertion) SheetBookEntrySecondaryKey(entryIndex int, expected string) *CharacterAssertion {
	found := false
	for _, key := range ca.sheet.CharacterBook.Entries[entryIndex].SecondaryKeys {
		if key == expected {
			found = true
			break
		}
	}
	assert.True(ca.t, found, "Book entry[%d] should contain secondary key: %s", entryIndex, expected)
	return ca
}

func (ca *CharacterAssertion) SheetDepthPromptPrompt(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.sheet.DepthPrompt.Prompt)
	return ca
}

func (ca *CharacterAssertion) SheetDepthPromptDepth(expected int) *CharacterAssertion {
	assert.Equal(ca.t, expected, ca.sheet.DepthPrompt.Depth)
	return ca
}

func (ca *CharacterAssertion) SheetNickname(expected string) *CharacterAssertion {
	assert.Equal(ca.t, expected, string(ca.sheet.Nickname))
	return ca
}

func (ca *CharacterAssertion) SheetRevision(revision character.Revision) *CharacterAssertion {
	stamp := character.Stamps[revision]
	assert.Equal(ca.t, revision, stamp.Revision, "Revision %d should be %d", revision, stamp.Revision)
	return ca.SheetVersion(stamp.Version).SheetSpec(stamp.Spec)
}

func (ca *CharacterAssertion) Consistent() *CharacterAssertion {
	assert.True(ca.t, ca.metadata.IsConsistentWith(ca.sheet), "Metadata should be consistent with character sheet")
	return ca.ValidTimestamps()
}

func (ca *CharacterAssertion) AssertImage() *CharacterAssertion {
	rawContext, err := ca.card.Encode()
	assert.NoError(ca.t, err, "Failed to encode character card")

	image, err := rawContext.Image()
	assert.NoError(ca.t, err, "Failed to get image from raw context")
	assert.NotNil(ca.t, image, "Image should not be nil")

	_, err = rawContext.ToBytes()
	assert.NoError(ca.t, err, "Failed to get bytes from raw context")

	return ca
}

func (ca *CharacterAssertion) Metadata() *models.Metadata {
	return ca.metadata
}

func (ca *CharacterAssertion) Card() *png.CharacterCard {
	return ca.card
}

func (ca *CharacterAssertion) Sheet() *character.Sheet {
	return ca.sheet
}

func (ca *CharacterAssertion) NoError() *CharacterAssertion {
	assert.NoError(ca.t, ca.err, "Card fetch should not have error")
	return ca
}

func (ca *CharacterAssertion) HasError() *CharacterAssertion {
	assert.Error(ca.t, ca.err, "Card fetch should have error")
	return ca
}
