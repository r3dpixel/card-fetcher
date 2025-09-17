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
//var testNyaiMeFetcher = processor.New(fetcher.NewNyaiMeFetcher())
//
//func TestNyaiMeFetcher(t *testing.T) {
//	assertSourceIsUp(t, testNyaiMeFetcher)
//}
//
//func TestNyaiMeImport(t *testing.T) {
//
//	creator := "cutie2"
//	url := "https://nyai.me/ai/bots/Test_aru"
//	fetcherTask := task.New(testClient, testNyaiMeFetcher, url, testNyaiMeFetcher.MainURL())
//	metadata, err := fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//	characterCard, err := fetcherTask.FetchCharacterCard()
//	assert.NoError(t, err)
//	if characterCard == nil {
//		assert.Fail(t, "Card is nil")
//	}
//	jsonSheet := characterCard.Sheet
//
//	metadata, err = fetcherTask.FetchMetadata()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "nyai.me/ai/bots/Test_aru", metadata.CardURL)
//	assert.Equal(t, "nyai.me/ai/bots/Test_aru", metadata.DirectURL)
//	assert.Equal(t, "Test_aru", metadata.CharacterID)
//	assert.Equal(t, TestCardLoreBookEntryName, *jsonSheet.Data.CharacterBook.Entries[0].Comment)
//	assert.Equal(t, "1165", metadata.PlatformID)
//
//	assertConsistency(t, metadata, characterCard)
//	assertCommonFields(t,
//		metadata,
//		jsonSheet,
//		source.NyaiMe,
//		testNyaiMeFetcher.NormalizeURL(metadata.CharacterID),
//		creator,
//		TestCardDescription,
//		stringsx.Empty,
//		TestCardSystemPrompt,
//		"main",
//		"Test"+character.CreatorNotesSeparator+"Test"+character.CreatorNotesSeparator+TestCardCreatorNotes,
//		character.SpecV2,
//		character.V2,
//	)
//	assertCharacterLoreBookCommonFields(t, jsonSheet)
//	assertImage(t, characterCard)
//}
