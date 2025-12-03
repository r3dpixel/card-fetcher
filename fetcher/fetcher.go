package fetcher

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
)

// BaseURL represents a base URL of a fetcher
type BaseURL struct {
	Domain string
	Path   string
}

// Builder builds a Fetcher
type Builder interface {
	Build(client *reqx.Client) Fetcher
}

// JsonResponse type alias for *sonicx.Wrap
type JsonResponse = *sonicx.Wrap

// Fetcher interface for all fetchers
type Fetcher interface {
	// Extends allows a fetcher to extend another fetcher
	Extends(top Fetcher)
	// SourceID returns the source ID of the fetcher
	SourceID() source.ID

	// SourceURL returns the source URL of the fetcher
	SourceURL() string
	// MainURL returns the main URL of the fetcher
	MainURL() string
	// BaseURLs returns the base URLs of the fetcher
	BaseURLs() []BaseURL
	// CharacterID returns the character ID from a URL
	CharacterID(rawCharacterID string) string
	// DirectURL returns the direct URL for a character
	DirectURL(characterID string) string
	// NormalizeURL returns the normalized URL for a character
	NormalizeURL(characterID string) string

	// FetchMetadataResponse fetches the metadata response from the source for the given characterID
	FetchMetadataResponse(characterID string) (*req.Response, error)
	// CreateBinder creates a MetadataBinder from the metadata response
	CreateBinder(characterID string, response string) (*MetadataBinder, error)
	// FetchCardInfo fetches the card info from the source
	FetchCardInfo(metadataBinder *MetadataBinder) (*models.CardInfo, error)
	// FetchCreatorInfo fetches the creator info from the source
	FetchCreatorInfo(metadataBinder *MetadataBinder) (*models.CreatorInfo, error)
	// FetchBookResponses fetches the book responses from the source
	FetchBookResponses(metadataBinder *MetadataBinder) (*BookBinder, error)
	// FetchCharacterCard fetches the character card from the source
	FetchCharacterCard(binder *Binder) (*png.CharacterCard, error)
	// Close closes the fetcher
	Close()

	// IsSourceUp checks if the source is up
	IsSourceUp() error
}
