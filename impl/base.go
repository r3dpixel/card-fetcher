package impl

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/reqx"
)

type BaseConfig struct {
	ServiceLabel string
	Client       *reqx.Client
	SourceID     source.ID
	SourceURL    string
	DirectURL    string
	MainURL      string
	BaseURLs     []string
}

// BaseFetcher - Embeddable struct for creating a new source
type BaseFetcher struct {
	fetcher.Fetcher
	serviceLabel string
	client       *reqx.Client
	sourceID     source.ID
	sourceURL    string
	directURL    string
	mainURL      string
	baseURLs     []string
}

// NewBaseFetcher Creates a new BaseFetcher
func NewBaseFetcher(config BaseConfig) BaseFetcher {
	return BaseFetcher{
		serviceLabel: config.ServiceLabel,
		client:       config.Client,
		sourceID:     config.SourceID,
		sourceURL:    config.SourceURL,
		directURL:    config.DirectURL,
		mainURL:      config.MainURL,
		baseURLs:     config.BaseURLs,
	}
}

func (f *BaseFetcher) Extends(top fetcher.Fetcher) {
	f.Fetcher = top
	f.serviceLabel = fmt.Sprintf("%s::%s", f.Fetcher.SourceID(), uuid.New())
}

func (f *BaseFetcher) SourceID() source.ID {
	return f.sourceID
}

func (f *BaseFetcher) SourceURL() string {
	return f.sourceURL
}

func (f *BaseFetcher) MainURL() string {
	return f.mainURL
}

func (f *BaseFetcher) BaseURLs() []string {
	return f.baseURLs
}

func (f *BaseFetcher) CharacterID(url string, matchedURL string) string {
	tokens := strings.Split(url, matchedURL)
	return tokens[len(tokens)-1]
}

func (f *BaseFetcher) DirectURL(characterID string) string {
	return path.Join(f.directURL, characterID)
}

func (f *BaseFetcher) NormalizeURL(characterID string) string {
	return path.Join(f.Fetcher.MainURL(), characterID)
}

func (f *BaseFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return &fetcher.MetadataBinder{
		CharacterID:   characterID,
		NormalizedURL: f.Fetcher.NormalizeURL(characterID),
		DirectURL:     f.Fetcher.DirectURL(characterID),
		JsonResponse:  metadataResponse,
	}, nil
}

func (f *BaseFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

func (f *BaseFetcher) IsSourceUp() bool {
	_, err := f.client.R().Get("https://" + f.sourceURL)
	return err == nil
}

func (f *BaseFetcher) Close() {}
