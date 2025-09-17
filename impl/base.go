package impl

import (
	"strings"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const ()

// BaseHandler - Embeddable struct for creating a new source
type BaseHandler struct {
	client    *req.Client
	sourceID  source.ID
	sourceURL string
	directURL string
	mainURL   string
	baseURLs  []string
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
	return s.directURL + characterID
}

func (s *BaseHandler) NormalizeURL(characterID string) string {
	return s.mainURL + characterID
}

func (s *BaseHandler) ParseMetadataResponse(response *req.Response) (gjson.Result, error) {
	bytes, err := response.ToBytes()
	if err != nil {
		return gjsonx.Empty, err
	}
	if gjson.ValidBytes(bytes) {
		return gjsonx.Empty, fetcher.InvalidJsonResponse
	}
	return gjson.ParseBytes(bytes), nil
}

func (s *BaseHandler) CreateBinder(characterID string, normalizedURL string, metadataResponse gjson.Result) (*fetcher.MetadataBinder, error) {
	return &fetcher.MetadataBinder{CharacterID: characterID, NormalizedURL: normalizedURL, Result: metadataResponse}, nil
}

func (s *BaseHandler) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

func (s *BaseHandler) IsSourceUp() bool {
	response, err := s.client.R().Get("https://" + s.sourceURL)
	return reqx.IsResponseErrOk(response, err)
}

func (s *BaseHandler) fromDate(format string, date string, url string) timestamp.Nano {
	t, err := timestamp.Parse[timestamp.Nano](format, date)
	if err != nil {
		log.Error().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, url).
			Msg("Could not parse timestamp")
	}
	return t
}
