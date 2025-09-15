package fetcher

import (
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
)

type Fetcher interface {
	SourceURL() string
	MainURL() string
	BaseURLs() []string
	SourceID() source.ID
	NormalizeURL(characterID string) string
	DirectURL(characterID string) string
	CharacterID(url string, matchedURL string) string
	FetchMetadata(normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error)
	FetchCharacterCard(metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error)
	IsSourceUp() bool
}
