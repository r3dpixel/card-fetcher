package postprocessor

import (
	"errors"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFetcher struct {
	mock.Mock
}

func (m *mockFetcher) DirectURL(characterID string) string {
	args := m.Called(characterID)
	return args.Get(0).(string)
}

func (m *mockFetcher) IsSourceUp(c *req.Client) bool {
	args := m.Called(c)
	return args.Get(0).(bool)
}

func (m *mockFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	args := m.Called(c, normalizedURL, characterID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.JsonResponse), args.Error(2)
	}
	return args.Get(0).(*models.Metadata), args.Get(1).(models.JsonResponse), args.Error(2)
}

func (m *mockFetcher) FetchCharacterCard(c *req.Client, metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
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

func (m *mockFetcher) Extends(f fetcher.Fetcher) {
	m.Called(f)
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

func (m *mockFetcher) NormalizeURL(characterID string) string {
	args := m.Called(characterID)
	return args.String(0)
}

func (m *mockFetcher) CharacterID(url string, matchedURL string) string {
	args := m.Called(url, matchedURL)
	return args.String(0)
}

func TestFetchMetadata(t *testing.T) {
	client := req.C()
	url := "http://example.com"
	charID := "123"

	t.Run("Success", func(t *testing.T) {
		mockF := new(mockFetcher)
		mockF.On("Extends", mock.Anything).Return()

		inputMetadata := &models.Metadata{
			CardName:      "  Card Name  ",
			CharacterName: " Character Name ",
			Creator:       " Creator ",
			Tagline:       " Tagline  ",
		}
		mockF.On("DirectURL", "").Return("a-dummy-direct-url")
		mockF.On("FetchMetadata", client, url, charID).Return(inputMetadata, models.EmptyJsonResponse, nil)

		processor := New(mockF)
		metadata, _, err := processor.FetchMetadata(client, url, charID)

		assert.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.Equal(t, "Card Name", metadata.CardName)
		assert.Equal(t, "Character Name", metadata.CharacterName)
		assert.Equal(t, "Creator", metadata.Creator)
		assert.Equal(t, "Tagline", metadata.Tagline)
		assert.Equal(t, "a-dummy-direct-url", metadata.DirectURL)
		mockF.AssertExpectations(t)
	})

	t.Run("Error from fetcher", func(t *testing.T) {
		mockF := new(mockFetcher)
		mockF.On("Extends", mock.Anything).Return()
		expectedErr := errors.New("fetch failed")
		mockF.On("FetchMetadata", client, url, charID).Return(nil, models.EmptyJsonResponse, expectedErr)

		processor := New(mockF)
		metadata, _, err := processor.FetchMetadata(client, url, charID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, metadata)
		mockF.AssertExpectations(t)
	})
}

func TestPatchingLogic(t *testing.T) {
	mockF := new(mockFetcher)
	mockF.On("Extends", mock.Anything).Return()
	mockF.On("SourceID").Return(source.ID("test-source"))
	mockF.On("DirectURL", "").Return("some-expected-url")
	p := &postProcessor{Fetcher: mockF}

	t.Run("patchMetadata", func(t *testing.T) {
		metadata := &models.Metadata{
			CardName:      "  Card Name  ",
			CharacterName: " Character Name ",
			Creator:       " Creator ",
			Tagline:       " Tagline  ",
		}
		p.patchMetadata(metadata)
		assert.Equal(t, "Card Name", metadata.CardName)
		assert.Equal(t, "Character Name", metadata.CharacterName)
		assert.Equal(t, "Creator", metadata.Creator)
		assert.Equal(t, "Tagline", metadata.Tagline)
		assert.Equal(t, "some-expected-url", metadata.DirectURL)
	})

	t.Run("patchCardName", func(t *testing.T) {
		t.Run("Prefers metadata.CharacterName", func(t *testing.T) {
			metadata := &models.Metadata{CharacterName: "Meta Name", CardName: "Card Name"}
			card := &character.Sheet{Data: character.Data{CharacterName: "Original Card Name"}}
			p.patchCardName(metadata, card)
			assert.Equal(t, "Meta Name", metadata.CharacterName)
			assert.Equal(t, "Meta Name", card.Data.CharacterName)
			assert.NotNil(t, card.Data.Nickname)
			assert.Equal(t, "Meta Name", *card.Data.Nickname)
		})

		t.Run("Falls back to card.Data.Name", func(t *testing.T) {
			metadata := &models.Metadata{CharacterName: " ", CardName: "Card Name"}
			card := &character.Sheet{Data: character.Data{CharacterName: "Original Card Name"}}
			p.patchCardName(metadata, card)
			assert.Equal(t, "Original Card Name", metadata.CharacterName)
			assert.Equal(t, "Original Card Name", card.Data.CharacterName)
		})

		t.Run("Falls back to metadata.CardName", func(t *testing.T) {
			metadata := &models.Metadata{CharacterName: "", CardName: "Card Name"}
			card := &character.Sheet{Data: character.Data{CharacterName: ""}}
			p.patchCardName(metadata, card)
			assert.Equal(t, "Card Name", metadata.CharacterName)
			assert.Equal(t, "Card Name", card.Data.CharacterName)
		})

		t.Run("Does not overwrite existing nickname", func(t *testing.T) {
			nickname := "Existing Nickname"
			metadata := &models.Metadata{CharacterName: "Meta Name"}
			card := &character.Sheet{Data: character.Data{Nickname: &nickname}}
			p.patchCardName(metadata, card)
			assert.NotNil(t, card.Data.Nickname)
			assert.Equal(t, "Existing Nickname", *card.Data.Nickname)
		})
	})

	t.Run("patchTags", func(t *testing.T) {
		metadata := &models.Metadata{Tags: []models.Tag{{Slug: "meta", Name: "Meta Tag"}}}
		card := &character.Sheet{Data: character.Data{Tags: []string{"card-tag", "meta"}}}
		p.patchTags(metadata, card)

		expectedTags := []models.Tag{{Slug: "cardtag", Name: "Card-Tag"}, {Slug: "meta", Name: "Meta"}}
		expectedStringTags := []string{"Card-Tag", "Meta"}

		assert.Equal(t, expectedTags, metadata.Tags)
		assert.Equal(t, expectedStringTags, card.Data.Tags)
	})

	t.Run("patchTimestamps", func(t *testing.T) {
		createTime := timestamp.Nano(time.Now().Add(-24 * time.Hour).UnixMilli())
		updateTime := timestamp.Nano(time.Now().UnixMilli())

		t.Run("Sets book update time if zero", func(t *testing.T) {
			metadata := &models.Metadata{CreateTime: createTime, UpdateTime: updateTime, BookUpdateTime: 0}
			card := &character.Sheet{Data: character.Data{CharacterBook: &character.Book{}}}
			p.patchTimestamps(metadata, card)
			assert.Equal(t, timestamp.Convert[timestamp.Seconds](createTime), card.Data.CreationDate)
			assert.Equal(t, timestamp.Convert[timestamp.Seconds](updateTime), card.Data.ModificationDate)
			assert.Equal(t, updateTime, metadata.BookUpdateTime)
		})

		t.Run("Does not set book update time if non-zero", func(t *testing.T) {
			bookTime := timestamp.Nano(time.Now().Add(-1 * time.Hour).UnixNano())
			metadata := &models.Metadata{CreateTime: createTime, UpdateTime: updateTime, BookUpdateTime: bookTime}
			card := &character.Sheet{Data: character.Data{CharacterBook: &character.Book{}}}
			p.patchTimestamps(metadata, card)
			assert.Equal(t, bookTime, metadata.BookUpdateTime)
		})

		t.Run("Does not set book update time if book is nil", func(t *testing.T) {
			metadata := &models.Metadata{CreateTime: createTime, UpdateTime: updateTime, BookUpdateTime: 0}
			card := &character.Sheet{Data: character.Data{CharacterBook: nil}}
			p.patchTimestamps(metadata, card)
			assert.Equal(t, timestamp.Nano(0), metadata.BookUpdateTime)
		})
	})

	t.Run("patchBook", func(t *testing.T) {
		t.Run("Handles nil book", func(t *testing.T) {
			assert.NotPanics(t, func() {
				p.patchBook(nil, "any name")
			})
		})

		t.Run("Sets name if blank", func(t *testing.T) {
			book := &character.Book{Name: new(string)}
			*book.Name = "  "
			p.patchBook(book, "Test Character")
			assert.Equal(t, "Test Character Lore Book", *book.Name)
		})

		t.Run("Replaces placeholder", func(t *testing.T) {
			book := &character.Book{Name: new(string)}
			*book.Name = "Lore for " + character.BookNamePlaceholder
			p.patchBook(book, "Test Character")
			assert.Equal(t, "Lore for Test Character", *book.Name)
		})

		t.Run("Replaces slashes", func(t *testing.T) {
			book := &character.Book{Name: new(string)}
			*book.Name = "A/B/C"
			p.patchBook(book, "Test Character")
			assert.Equal(t, "A-B-C", *book.Name)
		})
	})
}
