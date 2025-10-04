package task

import (
	"sync"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
)

type Task interface {
	SourceID() source.ID
	NormalizedURL() string
	OriginalURL() string
	FetchMetadata() (*models.Metadata, error)
	FetchCharacterCard() (*png.CharacterCard, error)
	FetchAll() (*models.Metadata, *png.CharacterCard, error)
}

type task struct {
	fetchMetadata      func() (*models.Metadata, error)
	fetchCharacterCard func() (*png.CharacterCard, error)

	sourceID      source.ID
	originalURL   string
	normalizedURL string
	characterID   string
}

func New(f fetcher.Fetcher, url string, matchedURL string) Task {
	characterID := f.CharacterID(url, matchedURL)
	normalizedURL := f.NormalizeURL(characterID)

	binderFlow := sync.OnceValues(func() (*fetcher.Binder, error) {
		return executeBinderFlow(f, characterID, normalizedURL)
	})

	metadataFlow := sync.OnceValues(func() (*models.Metadata, error) {
		return executeMetadataFlow(f, binderFlow)
	})

	characterCardFlow := sync.OnceValues(func() (*png.CharacterCard, error) {
		return executeCharacterCardFlow(f, binderFlow, metadataFlow)
	})

	return &task{
		fetchMetadata:      metadataFlow,
		fetchCharacterCard: characterCardFlow,

		sourceID:      f.SourceID(),
		originalURL:   url,
		normalizedURL: normalizedURL,
		characterID:   characterID,
	}
}

func (t *task) SourceID() source.ID {
	return t.sourceID
}

func (t *task) OriginalURL() string {
	return t.originalURL
}

func (t *task) NormalizedURL() string {
	return t.normalizedURL
}

func (t *task) FetchMetadata() (*models.Metadata, error) {
	return t.fetchMetadata()
}

func (t *task) FetchCharacterCard() (*png.CharacterCard, error) {
	return t.fetchCharacterCard()
}

func (t *task) FetchAll() (*models.Metadata, *png.CharacterCard, error) {
	metadata, err := t.fetchMetadata()
	if err != nil {
		return nil, nil, err
	}
	characterCard, err := t.fetchCharacterCard()
	if err != nil {
		return nil, nil, err
	}
	return metadata, characterCard, nil
}

func executeBinderFlow(
	f fetcher.Fetcher,
	characterID,
	normalizedURL string,
) (*fetcher.Binder, error) {
	response, err := reqx.String(f.FetchMetadataResponse(characterID))

	metadataResponse, err := sonicx.GetFromString(response)
	if err != nil {
		return nil, err
	}
	metadataBinder, err := f.CreateBinder(characterID, metadataResponse)
	if err != nil {
		return nil, err
	}
	bookBinder, err := f.FetchBookResponses(metadataBinder)
	if err != nil {
		return nil, err
	}
	return &fetcher.Binder{MetadataBinder: *metadataBinder, BookBinder: *bookBinder}, nil
}

func executeMetadataFlow(
	f fetcher.Fetcher,
	binderFlow func() (*fetcher.Binder, error),
) (*models.Metadata, error) {
	binder, err := binderFlow()
	if err != nil {
		return nil, err
	}
	cardInfo, err := f.FetchCardInfo(&binder.MetadataBinder)
	if err != nil {
		return nil, err
	}
	creatorInfo, err := f.FetchCreatorInfo(&binder.MetadataBinder)
	if err != nil {
		return nil, err
	}
	metadata := &models.Metadata{
		Source:         f.SourceID(),
		CardInfo:       *cardInfo,
		CreatorInfo:    *creatorInfo,
		BookUpdateTime: binder.UpdateTime,
	}
	fetcher.PatchMetadata(metadata)
	return metadata, nil
}

func executeCharacterCardFlow(
	f fetcher.Fetcher,
	binderFlow func() (*fetcher.Binder, error),
	metadataFlow func() (*models.Metadata, error),
) (*png.CharacterCard, error) {
	binder, err := binderFlow()
	if err != nil {
		return nil, err
	}
	metadata, err := metadataFlow()
	if err != nil {
		return nil, err
	}
	characterCard, err := f.FetchCharacterCard(binder)
	if err != nil {
		return nil, err
	}
	fetcher.PatchSheet(characterCard.Sheet, metadata)
	return characterCard, nil
}
