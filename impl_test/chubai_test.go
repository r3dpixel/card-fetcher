package fetcher_test

//
//import (
//	"testing"
//
//	"github.com/r3dpixel/card-fetcher/fetcher"
//	"github.com/r3dpixel/card-fetcher/source"
//	"github.com/r3dpixel/card-fetcher/task"
//	"github.com/r3dpixel/card-parser/character"
//	"github.com/r3dpixel/toolkit/stringsx"
//	"github.com/stretchr/testify/assert"
//)
//
//var testChubAIFetcher = processor.New(fetcher.NewChubAIFetcher())
//
//func TestChubAIFetcher(t *testing.T) {
//	assertSourceIsUp(t, testChubAIFetcher)
//}
//
//func TestChubAIImport(t *testing.T) {
//	t.Parallel()
//
//	const creator = "Anonymous"
//	url := "https://chub.ai/characters/Anonymous/test-f26406a9718a"
//	fetcherTask := task.New(testClient, testChubAIFetcher, url, testChubAIFetcher.MainURL())
//	metadata, err := fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//	card, err := fetcherTask.FetchCharacterCard()
//	assert.NoError(t, err)
//	jsonSheet := card.Sheet
//
//	metadata, err = fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "chub.ai/characters/Anonymous/test-f26406a9718a", metadata.CardURL)
//	assert.Equal(t, "chub.ai/characters/Anonymous/test-f26406a9718a", metadata.DirectURL)
//	assert.Equal(t, "Anonymous/test-f26406a9718a", metadata.CharacterID)
//	assert.Equal(t, TestCardLoreBookEntryName, *jsonSheet.Data.CharacterBook.Entries[0].Comment)
//	assert.Equal(t, "3186564", metadata.PlatformID)
//
//	assertConsistency(t, metadata, card)
//	assertCommonFields(t,
//		metadata,
//		jsonSheet,
//		source.ChubAI,
//		testChubAIFetcher.NormalizeURL(metadata.CharacterID),
//		creator,
//		TestCardDescription,
//		stringsx.Empty,
//		TestCardSystemPrompt,
//		"main",
//		TestCardTagline+character.CreatorNotesSeparator+TestCardCreatorNotes,
//		character.SpecV2,
//		character.V2,
//	)
//	assertCharacterLoreBookCommonFields(t, jsonSheet)
//	assertImage(t, card)
//}
//
//func TestChubAIImport_OneMultiLoreBook(t *testing.T) {
//	t.Parallel()
//
//	const creator = "statuotw"
//	url := "https://chub.ai/characters/statuotw/konako-the-adventurer-5c21dcc2"
//	fetcherTask := task.New(testClient, testChubAIFetcher, url, testChubAIFetcher.MainURL())
//	metadata, err := fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//	card, err := fetcherTask.FetchCharacterCard()
//	assert.NoError(t, err)
//	jsonSheet := card.Sheet
//
//	metadata, err = fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "The Fantasy World of Adolion", jsonSheet.Data.CharacterBook.GetName())
//	assert.Empty(t, jsonSheet.Data.CharacterBook.GetDescription())
//	assert.NotNil(t, jsonSheet.Data.CharacterBook.Description)
//	assert.Len(t, jsonSheet.Data.CharacterBook.Entries, 212)
//	assert.Equal(t, "chub.ai/characters/statuotw/konako-the-adventurer-5c21dcc2", metadata.CardURL)
//	assert.Equal(t, "statuotw/konako-the-adventurer-5c21dcc2", metadata.CharacterID)
//	assert.Equal(t, "2072277", metadata.PlatformID)
//
//	assertConsistency(t, metadata, card)
//
//	assertImage(t, card)
//}
//
//func TestChubAIImport_AuxiliaryLoreBook(t *testing.T) {
//	t.Parallel()
//
//	url := "https://chub.ai/characters/Enkob/maxine-flirty-warrior-of-the-explorers-guild-d1eb387be026"
//	fetcherTask := task.New(testClient, testChubAIFetcher, url, testChubAIFetcher.MainURL())
//	metadata, err := fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//	card, err := fetcherTask.FetchCharacterCard()
//	assert.NoError(t, err)
//	jsonSheet := card.Sheet
//
//	metadata, err = fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "Maxine -- The fantasy world of Runa (Base Lore)", jsonSheet.Data.CharacterBook.GetName())
//	assert.NotNil(t, jsonSheet.Data.CharacterBook.Description)
//	assert.Len(t, jsonSheet.Data.CharacterBook.Entries, 129)
//
//	assertConsistency(t, metadata, card)
//
//	assertImage(t, card)
//}
//
//func TestChubAI_MissingImage(t *testing.T) {
//	t.Parallel()
//
//	url := "https://chub.ai/characters/Sugondees/saeko-490390c0ebee"
//	fetcherTask := task.New(testClient, testChubAIFetcher, url, testChubAIFetcher.MainURL())
//	metadata, err := fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//	card, err := fetcherTask.FetchCharacterCard()
//	assert.NoError(t, err)
//
//	metadata, err = fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//
//	assertConsistency(t, metadata, card)
//
//	assertImage(t, card)
//}
