package fetcher

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/tidwall/gjson"
)

type SourceHandler interface {
	SourceID() source.ID
	SourceURL() string
	MainURL() string
	BaseURLs() []string
	CharacterID(url string, matchedURL string) string
	DirectURL(characterID string) string
	NormalizeURL(characterID string) string

	FetchMetadataResponse(characterID string) (*req.Response, error)
	ParseMetadataResponse(response *req.Response) (gjson.Result, error)
	CreateBinder(characterID string, normalizedURL string, metadataResponse gjson.Result) (*MetadataBinder, error)
	FetchCardInfo(metadataBinder *MetadataBinder) (*models.CardInfo, error)
	FetchCreatorInfo(metadataBinder *MetadataBinder) (*models.CreatorInfo, error)
	FetchBookResponses(metadataBinder *MetadataBinder) (*BookBinder, error)
	FetchCharacterCard(binder *Binder) (*png.CharacterCard, error)

	IsSourceUp() bool
}
