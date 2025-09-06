package fetcher

import (
	"strings"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
)

const ()

// BaseFetcher - Embeddable struct for creating a new source
type BaseFetcher struct {
	Fetcher
	sourceID  source.ID
	sourceURL string
	directURL string
	baseURLs  []string
}

func (s *BaseFetcher) SourceURL() string {
	return s.sourceURL
}

func (s *BaseFetcher) MainURL() string {
	return s.baseURLs[0]
}

func (s *BaseFetcher) BaseURLs() []string {
	return s.baseURLs
}

func (s *BaseFetcher) SourceID() source.ID {
	return s.sourceID
}

func (s *BaseFetcher) NormalizeURL(characterID string) string {
	return s.Fetcher.MainURL() + characterID
}

func (s *BaseFetcher) DirectURL(characterID string) string {
	return s.directURL + characterID
}

func (s *BaseFetcher) CharacterID(url string, matchedURL string) string {
	tokens := strings.Split(url, matchedURL)
	return tokens[len(tokens)-1]
}

func (s *BaseFetcher) Extends(f Fetcher) {
	s.Fetcher = f
}

func (s *BaseFetcher) fromDate(format string, date string, url string) timestamp.Nano {
	t, err := timestamp.Parse[timestamp.Nano](format, date)
	if err != nil {
		log.Error().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, url).
			Msg("Could not parse timestamp")
	}
	return t
}

func (s *BaseFetcher) fetchMetadataErr(url string, cause error) error {
	return trace.Err().
		Wrap(cause).
		Field(trace.SOURCE, s.sourceID).
		Field(trace.SERVICE, "fetcher").
		Field(trace.ACTIVITY, "fetch metadata").
		Field(trace.URL, url).
		Msg("Failed to fetch metadata")
}

func (s *BaseFetcher) missingPlatformIdErr(url string, cause error) error {
	return trace.Err().
		Wrap(cause).
		Field(trace.SOURCE, s.sourceID).
		Field(trace.SERVICE, "fetcher").
		Field(trace.ACTIVITY, "fetch metadata").
		Field(trace.URL, url).
		Msg("Missing platform ID")
}

func (s *BaseFetcher) IsSourceUp(c *req.Client) bool {
	response, err := c.R().Get("https://" + s.sourceURL)
	return reqx.IsResponseOk(response, err)
}
