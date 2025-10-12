package impl

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
)

type MockBuilder struct {
	MockConfig
	MockData
}

func (b MockBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewMockFetcher(b.MockConfig, b.MockData)
}

type MockConfig struct {
	MockSourceID      source.ID
	MockSourceURL     string
	MockDirectURL     string
	MockMainURL       string
	MockAlternateURLs []string
	IsUp              bool
}

type MockData struct {
	Response         *req.Response
	ResponseError    error
	CardInfo         *models.CardInfo
	CardInfoErr      error
	CreatorInfo      *models.CreatorInfo
	CreatorErr       error
	CharacterCard    *png.CharacterCard
	CharacterCardErr error
}

type mockFetcher struct {
	BaseFetcher
	MockData MockData
}

func NewMockFetcher(config MockConfig, mockData MockData) fetcher.Fetcher {
	f := &mockFetcher{
		BaseFetcher: BaseFetcher{
			client:    nil,
			sourceID:  config.MockSourceID,
			sourceURL: config.MockSourceURL,
			directURL: config.MockDirectURL,
			mainURL:   config.MockMainURL,
			baseURLs:  append([]string{config.MockMainURL}, config.MockAlternateURLs...),
		},
		MockData: mockData,
	}
	f.Extends(f)
	return f
}

func (m *mockFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	return m.MockData.Response, m.MockData.ResponseError
}

func (m *mockFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	return m.MockData.CardInfo, m.MockData.CardInfoErr
}

func (m *mockFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	return m.MockData.CreatorInfo, m.MockData.CreatorErr
}

func (m *mockFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	return m.MockData.CharacterCard, m.MockData.CharacterCardErr
}
