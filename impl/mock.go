package impl

import (
	"path"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
)

// MockBuilder builder for mock fetchers
type MockBuilder struct {
	MockConfig
	MockData
}

// Build creates a new mock fetcher using the configured options
func (b MockBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewMockFetcher(b.MockConfig, b.MockData)
}

// MockConfig contains the configuration for the mock fetcher
type MockConfig struct {
	MockSourceID           source.ID
	MockDomain             string
	MockPath               string
	MockDirectURL          string
	MockAdditionalBaseURLs []fetcher.BaseURL
	IsUp                   bool
}

// MockData contains the mock data to be returned by the fetcher
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

// mockFetcher is a fetcher that returns the mock data
type mockFetcher struct {
	BaseFetcher
	MockData MockData
}

// NewMockFetcher creates a new mock fetcher
func NewMockFetcher(config MockConfig, mockData MockData) fetcher.Fetcher {
	mainURL := path.Join(config.MockDomain, config.MockPath)
	f := &mockFetcher{
		BaseFetcher: BaseFetcher{
			client:    nil,
			sourceID:  config.MockSourceID,
			sourceURL: config.MockDomain,
			directURL: config.MockDirectURL,
			mainURL:   mainURL,
			baseURLs: append(
				[]fetcher.BaseURL{{Domain: config.MockDomain, Path: config.MockPath}},
				config.MockAdditionalBaseURLs...,
			),
		},
		MockData: mockData,
	}
	f.Extends(f)
	return f
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *mockFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	return f.MockData.Response, f.MockData.ResponseError
}

// FetchCardInfo fetches the card info from the source
func (f *mockFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	return f.MockData.CardInfo, f.MockData.CardInfoErr
}

// FetchCreatorInfo fetches the creator info from the source
func (f *mockFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	return f.MockData.CreatorInfo, f.MockData.CreatorErr
}

// FetchCharacterCard fetches the character card from the source
func (f *mockFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	return f.MockData.CharacterCard, f.MockData.CharacterCardErr
}
