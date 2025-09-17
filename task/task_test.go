package task

import (
	"errors"
	"sync"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFetcher struct {
	mock.Mock
}

func (m *mockFetcher) DirectURL(characterID string) string {
	args := m.Called(characterID)
	return args.String(0)
}

func (m *mockFetcher) IsSourceUp(c *req.Client) bool {
	args := m.Called(c)
	return args.Bool(0)
}

func (m *mockFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.CardInfo, models.JsonResponse, error) {
	args := m.Called(c, normalizedURL, characterID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.JsonResponse), args.Error(2)
	}
	return args.Get(0).(*models.CardInfo), args.Get(1).(models.JsonResponse), args.Error(2)
}

func (m *mockFetcher) FetchCharacterCard(c *req.Client, metadata *models.CardInfo, response models.JsonResponse) (*png.CharacterCard, error) {
	args := m.Called(c, metadata, response)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*png.CharacterCard), args.Error(1)
}

func (m *mockFetcher) SourceID() source.ID {
	args := m.Called()
	return args.Get(0).(source.ID)
}

func (m *mockFetcher) CharacterID(url string, matchedURL string) string {
	args := m.Called(url, matchedURL)
	return args.String(0)
}

func (m *mockFetcher) NormalizeURL(characterID string) string {
	args := m.Called(characterID)
	return args.String(0)
}

func (m *mockFetcher) MainURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFetcher) SourceURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFetcher) BaseURLs() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockFetcher) Extends(f fetcher.SourceHandler) {
	m.Called(f)
}

func TestNew(t *testing.T) {
	mockF := new(mockFetcher)
	url, matchedURL := "http://example.com/char/123", "http://example.com/char/123"
	charID, normURL := "123", "http://normalized.com/123"

	mockF.On("CharacterID", url, matchedURL).Return(charID).Once()
	mockF.On("NormalizeURL", charID).Return(normURL).Once()

	taskInstance := New(req.C(), mockF, url, matchedURL)

	assert.NotNil(t, taskInstance)
	internalTask, ok := taskInstance.(*task)
	assert.True(t, ok)

	assert.Equal(t, url, internalTask.originalURL)
	assert.Equal(t, normURL, internalTask.normalizedURL)
	assert.Equal(t, charID, internalTask.characterID)
	assert.Equal(t, mockF, internalTask.fetcher)

	mockF.AssertExpectations(t)
}

func TestTask_FetchMetadata(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockF := new(mockFetcher)
		expectedMetadata := &models.CardInfo{Title: "Test Card"}
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(expectedMetadata, models.EmptyJsonResponse, nil).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}

		meta1, err1 := taskInstance.FetchMetadata()
		meta2, err2 := taskInstance.FetchMetadata()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, expectedMetadata, meta1)
		assert.Equal(t, expectedMetadata, meta2)
		assert.NotSame(t, expectedMetadata, meta1, "Should return a clone")
		mockF.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockF := new(mockFetcher)
		expectedErr := errors.New("metadata fetch failed")
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil, models.EmptyJsonResponse, expectedErr).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}
		meta, err := taskInstance.FetchMetadata()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, meta)
		mockF.AssertExpectations(t)
	})
}

func TestTask_FetchCard(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockF := new(mockFetcher)
		metadata := &models.CardInfo{Title: "Test Card"}
		expectedCard := &png.CharacterCard{}
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(metadata, models.EmptyJsonResponse, nil).Once()
		mockF.On("FetchCharacterCard", mock.Anything, metadata, models.EmptyJsonResponse).Return(expectedCard, nil).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}

		card1, err1 := taskInstance.FetchCharacterCard()
		card2, err2 := taskInstance.FetchCharacterCard()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, expectedCard, card1)
		assert.Equal(t, expectedCard, card2)
		mockF.AssertExpectations(t)
	})

	t.Run("CardInfo fetch fails", func(t *testing.T) {
		mockF := new(mockFetcher)
		expectedErr := errors.New("metadata fetch failed")
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil, models.EmptyJsonResponse, expectedErr).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}
		card, err := taskInstance.FetchCharacterCard()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, card)
		mockF.AssertExpectations(t)
		mockF.AssertNotCalled(t, "FetchCharacterCard", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Card fetch fails", func(t *testing.T) {
		mockF := new(mockFetcher)
		metadata := &models.CardInfo{Title: "Test Card"}
		expectedErr := errors.New("card fetch failed")
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(metadata, models.EmptyJsonResponse, nil).Once()
		mockF.On("FetchCharacterCard", mock.Anything, metadata, models.EmptyJsonResponse).Return(nil, expectedErr).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}
		card, err := taskInstance.FetchCharacterCard()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, card)
		mockF.AssertExpectations(t)
	})
}

func TestTask_FetchAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockF := new(mockFetcher)
		metadata := &models.CardInfo{Title: "Test Card"}
		card := &png.CharacterCard{}
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(metadata, models.EmptyJsonResponse, nil).Once()
		mockF.On("FetchCharacterCard", mock.Anything, metadata, models.EmptyJsonResponse).Return(card, nil).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}
		metaResult, cardResult, err := taskInstance.FetchAll()

		assert.NoError(t, err)
		assert.Equal(t, metadata, metaResult)
		assert.NotSame(t, metadata, metaResult, "Should return a clone of metadata")
		assert.Equal(t, card, cardResult)
		mockF.AssertExpectations(t)
	})

	t.Run("Error propagates from FetchCharacterCard", func(t *testing.T) {
		mockF := new(mockFetcher)
		expectedErr := errors.New("any fetch failed")
		mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(nil, models.EmptyJsonResponse, expectedErr).Once()

		taskInstance := &task{fetcher: mockF, client: req.C()}
		meta, card, err := taskInstance.FetchAll()

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, meta)
		assert.Nil(t, card)
		mockF.AssertExpectations(t)
	})
}

func TestTask_Concurrency(t *testing.T) {
	mockF := new(mockFetcher)
	metadata := &models.CardInfo{Title: "Concurrent Card"}
	card := &png.CharacterCard{}

	mockF.On("FetchCardInfo", mock.Anything, mock.Anything, mock.Anything).Return(metadata, models.EmptyJsonResponse, nil).Once()
	mockF.On("FetchCharacterCard", mock.Anything, metadata, models.EmptyJsonResponse).Return(card, nil).Once()

	taskInstance := &task{fetcher: mockF, client: req.C()}
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

	mockF.AssertExpectations(t)
}
