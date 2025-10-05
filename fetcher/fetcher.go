package fetcher

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
)

type Builder func(*reqx.Client) Fetcher

type ConfigBuilder[T any] func(*reqx.Client, T) Fetcher

type JsonResponse = *sonicx.Wrap

type Fetcher interface {
	Extends(top Fetcher)
	SourceID() source.ID

	SourceURL() string
	MainURL() string
	BaseURLs() []string
	CharacterID(url string, matchedURL string) string
	DirectURL(characterID string) string
	NormalizeURL(characterID string) string

	FetchMetadataResponse(characterID string) (*req.Response, error)
	CreateBinder(characterID string, metadataResponse JsonResponse) (*MetadataBinder, error)
	FetchCardInfo(metadataBinder *MetadataBinder) (*models.CardInfo, error)
	FetchCreatorInfo(metadataBinder *MetadataBinder) (*models.CreatorInfo, error)
	FetchBookResponses(metadataBinder *MetadataBinder) (*BookBinder, error)
	FetchCharacterCard(binder *Binder) (*png.CharacterCard, error)
	Close()

	IsSourceUp() bool
}

func BuilderOf[T any](config T, builder ConfigBuilder[T]) Builder {
	return func(client *reqx.Client) Fetcher {
		return builder(client, config)
	}
}
