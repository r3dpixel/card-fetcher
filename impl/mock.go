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

func (f *mockFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	return f.MockData.Response, f.MockData.ResponseError
}

func (f *mockFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	return f.MockData.CardInfo, f.MockData.CardInfoErr
}

func (f *mockFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	return f.MockData.CreatorInfo, f.MockData.CreatorErr
}

func (f *mockFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	return f.MockData.CharacterCard, f.MockData.CharacterCardErr
}
