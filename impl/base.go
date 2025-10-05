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

const ()

// BaseHandler - Embeddable struct for creating a new source
type BaseHandler struct {
	fetcher.Fetcher
	serviceLabel string
	client       *reqx.Client
	sourceID     source.ID
	sourceURL    string
	directURL    string
	mainURL      string
	baseURLs     []string
}

func (s *BaseHandler) Extends(top fetcher.Fetcher) {
	s.Fetcher = top
	s.serviceLabel = fmt.Sprintf("%s::%s", s.Fetcher.SourceID(), uuid.New())
}

func (s *BaseHandler) SourceID() source.ID {
	return s.sourceID
}

func (s *BaseHandler) SourceURL() string {
	return s.sourceURL
}

func (s *BaseHandler) MainURL() string {
	return s.mainURL
}

func (s *BaseHandler) BaseURLs() []string {
	return s.baseURLs
}

func (s *BaseHandler) CharacterID(url string, matchedURL string) string {
	tokens := strings.Split(url, matchedURL)
	return tokens[len(tokens)-1]
}

func (s *BaseHandler) DirectURL(characterID string) string {
	return path.Join(s.directURL, characterID)
}

func (s *BaseHandler) NormalizeURL(characterID string) string {
	return path.Join(s.Fetcher.MainURL(), characterID)
}

func (s *BaseHandler) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return &fetcher.MetadataBinder{
		CharacterID:   characterID,
		NormalizedURL: s.Fetcher.NormalizeURL(characterID),
		DirectURL:     s.Fetcher.DirectURL(characterID),
		JsonResponse:  metadataResponse,
	}, nil
}

func (s *BaseHandler) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

func (s *BaseHandler) IsSourceUp() bool {
	_, err := s.client.R().Get("https://" + s.sourceURL)
	return err == nil
}

func (s *BaseHandler) Close() {}
