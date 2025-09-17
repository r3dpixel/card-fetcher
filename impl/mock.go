package impl

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
)

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

type mockHandler struct {
	BaseHandler
	MockData MockData
}

func MockHandler(config MockConfig, mockData MockData) fetcher.SourceHandler {
	return &mockHandler{
		BaseHandler: BaseHandler{
			client:    nil,
			sourceID:  config.MockSourceID,
			sourceURL: config.MockSourceURL,
			directURL: config.MockMainURL,
			baseURLs:  append([]string{config.MockMainURL}, config.MockAlternateURLs...),
		},
		MockData: mockData,
	}
}

func (m *mockHandler) FetchMetadataResponse(characterID string) (*req.Response, error) {
	return m.MockData.Response, m.MockData.ResponseError
}

func (m *mockHandler) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	return m.MockData.CardInfo, m.MockData.CardInfoErr
}

func (m *mockHandler) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	return m.MockData.CreatorInfo, m.MockData.CreatorErr
}

func (m *mockHandler) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	return m.MockData.CharacterCard, m.MockData.CharacterCardErr
}
