package impl

import (
	"fmt"
	"path"

	"github.com/google/uuid"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
)

// BaseFetcher embeddable struct for creating a new fetcher
type BaseFetcher struct {
	fetcher.Fetcher
	serviceLabel string
	client       *reqx.Client
	sourceID     source.ID
	sourceURL    string
	directURL    string
	mainURL      string
	baseURLs     []fetcher.BaseURL
}

// Extends extends the fetcher with the given fetcher
func (f *BaseFetcher) Extends(top fetcher.Fetcher) {
	// Set the internal fetcher reference to the top fetcher
	f.Fetcher = top
	// Set the service label
	f.serviceLabel = fmt.Sprintf("%s::%s", f.Fetcher.SourceID(), uuid.New())
}

// SourceID returns the source ID of the fetcher
func (f *BaseFetcher) SourceID() source.ID {
	return f.sourceID
}

// SourceURL returns the source URL of the fetcher
func (f *BaseFetcher) SourceURL() string {
	return f.sourceURL
}

// MainURL returns the main URL of the fetcher
func (f *BaseFetcher) MainURL() string {
	return f.mainURL
}

// BaseURLs returns the base URLs of the fetcher
func (f *BaseFetcher) BaseURLs() []fetcher.BaseURL {
	return f.baseURLs
}

// CharacterID hook for overriding the character ID
// NO-OP by default
func (f *BaseFetcher) CharacterID(rawCharacterID string) string {
	return rawCharacterID
}

// DirectURL returns the direct URL for a character
func (f *BaseFetcher) DirectURL(characterID string) string {
	return path.Join(f.directURL, characterID)
}

// NormalizeURL returns the normalized URL for a character
func (f *BaseFetcher) NormalizeURL(characterID string) string {
	// Uses the internal fetcher MainURL in case of override
	return path.Join(f.Fetcher.MainURL(), characterID)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *BaseFetcher) CreateBinder(characterID string, response string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, response)
}

// FetchBookResponses fetches the book responses from the source (no-op for convenience, override if needed)
func (f *BaseFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

// IsSourceUp checks if the source is up
func (f *BaseFetcher) IsSourceUp() error {
	_, err := f.client.R().Get("https://" + f.sourceURL)
	return err
}

// Close closes the fetcher (no-op for convenience, override if needed)
func (f *BaseFetcher) Close() {}

// CreateBinderFromJSON parses the JSON metadata response from the source, and creates a MetadataBinder
// Optionally, the path to the character ID can be specified, for overriding the character ID
func (f *BaseFetcher) CreateBinderFromJSON(characterID string, response string, pathCharacterID ...any) (*fetcher.MetadataBinder, error) {
	// Parse metadata response
	jsonResponse, err := sonicx.GetFromString(response)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedMetadataErr)
	}

	// Override the character ID if needed
	if len(pathCharacterID) > 0 {
		// Get the character ID from the JSON response and check if it's not blank
		if id := jsonResponse.GetByPath(pathCharacterID...).String(); stringsx.IsNotBlank(id) {
			// Override the character ID with the one from the JSON response
			characterID = id
		}
	}

	// Return the binder
	return &fetcher.MetadataBinder{
		CharacterID:   characterID,
		NormalizedURL: f.Fetcher.NormalizeURL(characterID),
		DirectURL:     f.Fetcher.DirectURL(characterID),
		JsonResponse:  jsonResponse,
	}, nil
}
