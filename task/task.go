package task

import (
	"sync"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
)

type Task interface {
	SourceID() source.ID
	OriginalURL() string
	NormalizedURL() string
	FetchMetadata() (*models.Metadata, error)
	FetchCharacterCard() (*png.CharacterCard, error)
	FetchAll() (*models.Metadata, *png.CharacterCard, error)
}

type task struct {
	fetchMetadataOnce sync.Once
	fetchCardOnce     sync.Once

	client        *req.Client
	fetcher       fetcher.Fetcher
	originalURL   string
	normalizedURL string
	characterID   string

	response    models.JsonResponse
	metadata    *models.Metadata
	metadataErr error
	card        *png.CharacterCard
	cardErr     error
}

func New(client *req.Client, fetcher fetcher.Fetcher, url string, matchedURL string) Task {
	characterID := fetcher.CharacterID(url, matchedURL)
	normalizedURL := fetcher.NormalizeURL(characterID)
	return &task{
		fetcher:       fetcher,
		client:        client,
		originalURL:   url,
		normalizedURL: normalizedURL,
		characterID:   characterID,
	}
}

func (t *task) SourceID() source.ID {
	return t.fetcher.SourceID()
}

func (t *task) OriginalURL() string {
	return t.originalURL
}

func (t *task) NormalizedURL() string {
	return t.normalizedURL
}

func (t *task) internalFetchMetadata() {
	t.fetchMetadataOnce.Do(func() {
		t.metadata, t.response, t.metadataErr = t.fetcher.FetchMetadata(t.client, t.normalizedURL, t.characterID)
	})
}

func (t *task) FetchMetadata() (*models.Metadata, error) {
	t.internalFetchMetadata()
	if t.metadataErr != nil {
		return nil, t.metadataErr
	}
	return t.metadata.Clone(), t.metadataErr
}

func (t *task) FetchCharacterCard() (*png.CharacterCard, error) {
	t.internalFetchMetadata()
	if t.metadataErr != nil {
		return nil, t.metadataErr
	}

	t.fetchCardOnce.Do(func() {
		t.card, t.cardErr = t.fetcher.FetchCharacterCard(t.client, t.metadata, t.response)
	})

	return t.card, t.cardErr
}

func (t *task) FetchAll() (*models.Metadata, *png.CharacterCard, error) {
	card, err := t.FetchCharacterCard()
	if err != nil {
		return nil, nil, err
	}
	return t.metadata.Clone(), card, nil
}
