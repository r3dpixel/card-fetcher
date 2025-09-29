package task

import (
	"errors"
	"sync"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/impl"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	url, matchedURL := "http://example.com/char/123", "example.com/"
	charID, normURL := "char/123", "normalized.com/char/123"
	sourceID := source.ID("test-source")

	mockConfig := impl.MockConfig{
		MockSourceID:      sourceID,
		MockSourceURL:     "http://example.com",
		MockDirectURL:     "http://direct.example.com",
		MockMainURL:       "normalized.com/",
		MockAlternateURLs: []string{"example.com/"},
		IsUp:              true,
	}

	mockData := impl.MockData{
		Response:      &req.Response{},
		ResponseError: nil,
		CardInfo:      &models.CardInfo{CharacterID: charID, NormalizedURL: normURL},
		CardInfoErr:   nil,
		CreatorInfo:   &models.CreatorInfo{},
		CreatorErr:    nil,
	}

	mockF := impl.NewMockFetcher(mockConfig, mockData)

	taskInstance := New(mockF, url, matchedURL)

	assert.NotNil(t, taskInstance)
	assert.Equal(t, sourceID, taskInstance.SourceID())
	assert.Equal(t, normURL, taskInstance.NormalizedURL())
}

func TestTask_FetchMetadata(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedCardInfo := &models.CardInfo{Title: "Test Card", CharacterID: "123"}
		expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "example.com",
			IsUp:          true,
		}
		response := &req.Response{}
		response.SetBodyString(`{}`)
		mockData := impl.MockData{
			Response:      response,
			ResponseError: nil,
			CardInfo:      expectedCardInfo,
			CardInfoErr:   nil,
			CreatorInfo:   expectedCreatorInfo,
			CreatorErr:    nil,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "example.com/"
		taskInstance := New(mockF, url, matchedURL)

		meta1, err1 := taskInstance.FetchMetadata()
		meta2, err2 := taskInstance.FetchMetadata()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, meta1)
		assert.NotNil(t, meta2)
		assert.Equal(t, expectedCardInfo.Title, meta1.Title)
		assert.Equal(t, expectedCreatorInfo.Nickname, meta1.Nickname)
		assert.Equal(t, meta1.Title, meta2.Title)
		assert.Same(t, meta1, meta2, "Should return same instances")
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("metadata fetch failed")

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "http://example.com",
			IsUp:          true,
		}

		mockData := impl.MockData{
			Response:      &req.Response{},
			ResponseError: expectedErr,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		meta, err := taskInstance.FetchMetadata()

		assert.Error(t, err)
		assert.Nil(t, meta)
	})
}

func TestTask_FetchCard(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedCard := &png.CharacterCard{Sheet: character.DefaultSheet(character.RevisionV2)}
		expectedCardInfo := &models.CardInfo{Title: "Test Card", CharacterID: "123"}
		expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "example.com/",
			IsUp:          true,
		}

		response := &req.Response{}
		response.SetBodyString(`{}`)
		mockData := impl.MockData{
			Response:         response,
			ResponseError:    nil,
			CardInfo:         expectedCardInfo,
			CardInfoErr:      nil,
			CreatorInfo:      expectedCreatorInfo,
			CreatorErr:       nil,
			CharacterCard:    expectedCard,
			CharacterCardErr: nil,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		card1, err1 := taskInstance.FetchCharacterCard()
		card2, err2 := taskInstance.FetchCharacterCard()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, expectedCard, card1)
		assert.Equal(t, expectedCard, card2)
	})

	t.Run("Metadata response fetch fails", func(t *testing.T) {
		expectedErr := errors.New("metadata response fetch failed")

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "http://example.com",
			IsUp:          true,
		}

		mockData := impl.MockData{
			Response:      &req.Response{},
			ResponseError: expectedErr,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		card, err := taskInstance.FetchCharacterCard()

		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("Card fetch fails", func(t *testing.T) {
		expectedErr := errors.New("card fetch failed")
		expectedCardInfo := &models.CardInfo{Title: "Test Card", CharacterID: "123"}
		expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "http://example.com",
			IsUp:          true,
		}

		mockData := impl.MockData{
			Response:         &req.Response{},
			ResponseError:    nil,
			CardInfo:         expectedCardInfo,
			CardInfoErr:      nil,
			CreatorInfo:      expectedCreatorInfo,
			CreatorErr:       nil,
			CharacterCard:    nil,
			CharacterCardErr: expectedErr,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		card, err := taskInstance.FetchCharacterCard()

		assert.Error(t, err)
		assert.Nil(t, card)
	})
}

func TestTask_FetchAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedCard := &png.CharacterCard{Sheet: character.DefaultSheet(character.RevisionV2)}
		expectedCardInfo := &models.CardInfo{Title: "Test Card", CharacterID: "123"}
		expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "example.com/",
			IsUp:          true,
		}

		response := &req.Response{}
		response.SetBodyString(`{}`)
		mockData := impl.MockData{
			Response:         response,
			ResponseError:    nil,
			CardInfo:         expectedCardInfo,
			CardInfoErr:      nil,
			CreatorInfo:      expectedCreatorInfo,
			CreatorErr:       nil,
			CharacterCard:    expectedCard,
			CharacterCardErr: nil,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		metaResult, cardResult, err := taskInstance.FetchAll()

		assert.NoError(t, err)
		assert.NotNil(t, metaResult)
		assert.Equal(t, expectedCardInfo.Title, metaResult.Title)
		assert.Equal(t, expectedCreatorInfo.Nickname, metaResult.Nickname)
		assert.Equal(t, expectedCard, cardResult)
	})

	t.Run("Error propagates from metadata fetch", func(t *testing.T) {
		expectedErr := errors.New("metadata fetch failed")

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "http://example.com",
			IsUp:          true,
		}

		mockData := impl.MockData{
			Response:      &req.Response{},
			ResponseError: expectedErr,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		meta, card, err := taskInstance.FetchAll()

		assert.Error(t, err)
		assert.Nil(t, meta)
		assert.Nil(t, card)
	})

	t.Run("Error propagates from character card fetch", func(t *testing.T) {
		expectedErr := errors.New("character card fetch failed")
		expectedCardInfo := &models.CardInfo{Title: "Test Card", CharacterID: "123"}
		expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

		mockConfig := impl.MockConfig{
			MockSourceID:  source.ID("test-source"),
			MockSourceURL: "http://example.com",
			MockMainURL:   "http://example.com",
			IsUp:          true,
		}

		mockData := impl.MockData{
			Response:         &req.Response{},
			ResponseError:    nil,
			CardInfo:         expectedCardInfo,
			CardInfoErr:      nil,
			CreatorInfo:      expectedCreatorInfo,
			CreatorErr:       nil,
			CharacterCard:    nil,
			CharacterCardErr: expectedErr,
		}

		mockF := impl.NewMockFetcher(mockConfig, mockData)
		url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
		taskInstance := New(mockF, url, matchedURL)

		meta, card, err := taskInstance.FetchAll()

		assert.Error(t, err)
		assert.Nil(t, meta)
		assert.Nil(t, card)
	})
}

func TestTask_Concurrency(t *testing.T) {
	expectedCard := &png.CharacterCard{Sheet: character.DefaultSheet(character.RevisionV2)}
	expectedCardInfo := &models.CardInfo{Title: "Concurrent Card", CharacterID: "123"}
	expectedCreatorInfo := &models.CreatorInfo{Nickname: "TestCreator"}

	mockConfig := impl.MockConfig{
		MockSourceID:  source.ID("test-source"),
		MockSourceURL: "http://example.com",
		MockMainURL:   "http://example.com",
		IsUp:          true,
	}

	response := &req.Response{}
	response.SetBodyString(`{}`)
	mockData := impl.MockData{
		Response:         response,
		ResponseError:    nil,
		CardInfo:         expectedCardInfo,
		CardInfoErr:      nil,
		CreatorInfo:      expectedCreatorInfo,
		CreatorErr:       nil,
		CharacterCard:    expectedCard,
		CharacterCardErr: nil,
	}

	mockF := impl.NewMockFetcher(mockConfig, mockData)
	url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
	taskInstance := New(mockF, url, matchedURL)

	wg := sync.WaitGroup{}
	numGoroutines := 10

	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			meta, err := taskInstance.FetchMetadata()
			assert.NoError(t, err)
			assert.NotNil(t, meta)
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			c, err := taskInstance.FetchCharacterCard()
			assert.NoError(t, err)
			assert.NotNil(t, c)
		}()
	}

	wg.Wait()
}

func TestTask_SourceID(t *testing.T) {
	sourceID := source.ID("test-source-id")

	mockConfig := impl.MockConfig{
		MockSourceID:  sourceID,
		MockSourceURL: "http://example.com",
		MockMainURL:   "http://example.com",
		IsUp:          true,
	}

	mockData := impl.MockData{
		Response:      &req.Response{},
		ResponseError: nil,
		CardInfo:      &models.CardInfo{CharacterID: "123", NormalizedURL: "http://normalized.com/123"},
		CardInfoErr:   nil,
		CreatorInfo:   &models.CreatorInfo{},
		CreatorErr:    nil,
	}

	mockF := impl.NewMockFetcher(mockConfig, mockData)
	url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
	taskInstance := New(mockF, url, matchedURL)

	result := taskInstance.SourceID()
	assert.Equal(t, sourceID, result)
}

func TestTask_NormalizedURL(t *testing.T) {
	normalizedURL := "normalized.com/char/123"

	mockConfig := impl.MockConfig{
		MockSourceID:      source.ID("test-source"),
		MockSourceURL:     "http://example.com",
		MockMainURL:       "normalized.com/",
		MockAlternateURLs: []string{"example.com/"},
		IsUp:              true,
	}

	mockData := impl.MockData{
		Response:      &req.Response{},
		ResponseError: nil,
		CardInfo:      &models.CardInfo{CharacterID: "123", NormalizedURL: normalizedURL},
		CardInfoErr:   nil,
		CreatorInfo:   &models.CreatorInfo{},
		CreatorErr:    nil,
	}

	mockF := impl.NewMockFetcher(mockConfig, mockData)
	url, matchedURL := "http://example.com/char/123", "example.com/"
	taskInstance := New(mockF, url, matchedURL)

	result := taskInstance.NormalizedURL()
	assert.Equal(t, normalizedURL, result)
}
