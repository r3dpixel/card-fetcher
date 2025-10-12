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

func (s *BaseFetcher) Extends(top fetcher.Fetcher) {
	s.Fetcher = top
	s.serviceLabel = fmt.Sprintf("%s::%s", s.Fetcher.SourceID(), uuid.New())
}

func (s *BaseFetcher) SourceID() source.ID {
	return s.sourceID
}

func (s *BaseFetcher) SourceURL() string {
	return s.sourceURL
}

func (s *BaseFetcher) MainURL() string {
	return s.mainURL
}

func (s *BaseFetcher) BaseURLs() []string {
	return s.baseURLs
}

func (s *BaseFetcher) CharacterID(url string, matchedURL string) string {
	tokens := strings.Split(url, matchedURL)
	return tokens[len(tokens)-1]
}

func (s *BaseFetcher) DirectURL(characterID string) string {
	return path.Join(s.directURL, characterID)
}

func (s *BaseFetcher) NormalizeURL(characterID string) string {
	return path.Join(s.Fetcher.MainURL(), characterID)
}

func (s *BaseFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return &fetcher.MetadataBinder{
		CharacterID:   characterID,
		NormalizedURL: s.Fetcher.NormalizeURL(characterID),
		DirectURL:     s.Fetcher.DirectURL(characterID),
		JsonResponse:  metadataResponse,
	}, nil
}

func (s *BaseFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

func (s *BaseFetcher) IsSourceUp() bool {
	_, err := s.client.R().Get("https://" + s.sourceURL)
	return err == nil
}

func (s *BaseFetcher) Close() {}
