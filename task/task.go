package task

import (
	"sync"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/trace"
)

// Task represents a single fetcher task
type Task interface {
	// SourceID returns the source ID of the fetcher
	SourceID() source.ID
	// NormalizedURL returns the normalized URL of the card
	NormalizedURL() string
	// OriginalURL returns the original URL of the card
	OriginalURL() string
	// FetchMetadata fetches the metadata from the source
	FetchMetadata() (*models.Metadata, error)
	// FetchCharacterCard fetches the character card from the source
	FetchCharacterCard() (*png.CharacterCard, error)
	// FetchAll fetches all the data from the source
	FetchAll() (*models.Metadata, *png.CharacterCard, error)
}

// task represents a single fetcher task
type task struct {
	// fetchMetadata closure (executes the metadata flow)
	fetchMetadata func() (*models.Metadata, error)

	// fetchCharacterCard closure (executes the character card flow)
	fetchCharacterCard func() (*png.CharacterCard, error)

	sourceID      source.ID
	originalURL   string
	normalizedURL string
	characterID   string
}

// New creates a new Task instance with the given fetcher and appropriate URLs
func New(f fetcher.Fetcher, url, rawCharacterID string) Task {
	// Extract characterID
	characterID := f.CharacterID(rawCharacterID)
	// Extract normalizedURL
	normalizedURL := f.NormalizeURL(characterID)

	// Create the binder flow closure (executed once and cached)
	binderFlow := sync.OnceValues(func() (*fetcher.Binder, error) {
		return executeBinderFlow(f, characterID)
	})

	// Create the metadata flow closure (executed once and cached)
	metadataFlow := sync.OnceValues(func() (*models.Metadata, error) {
		return executeMetadataFlow(f, binderFlow)
	})

	// Create the character card flow closure (executed once and cached)
	characterCardFlow := sync.OnceValues(func() (*png.CharacterCard, error) {
		return executeCharacterCardFlow(f, binderFlow, metadataFlow)
	})

	// Return the task instance
	return &task{
		fetchMetadata:      metadataFlow,
		fetchCharacterCard: characterCardFlow,

		sourceID:      f.SourceID(),
		originalURL:   url,
		normalizedURL: normalizedURL,
		characterID:   characterID,
	}
}

// SourceID returns the source ID of the fetcher
func (t *task) SourceID() source.ID {
	return t.sourceID
}

// OriginalURL returns the original URL of the card
func (t *task) OriginalURL() string {
	return t.originalURL
}

// NormalizedURL returns the normalized URL of the card
func (t *task) NormalizedURL() string {
	return t.normalizedURL
}

// FetchMetadata fetches the metadata from the source
func (t *task) FetchMetadata() (*models.Metadata, error) {
	return t.fetchMetadata()
}

// FetchCharacterCard fetches the character card from the source
func (t *task) FetchCharacterCard() (*png.CharacterCard, error) {
	return t.fetchCharacterCard()
}

// FetchAll fetches all the data from the source
func (t *task) FetchAll() (*models.Metadata, *png.CharacterCard, error) {
	// Fetch metadata (using the flow, executed once)
	metadata, err := t.fetchMetadata()
	if err != nil {
		return nil, nil, err
	}
	// Fetch character card (using the flow, executed once)
	characterCard, err := t.fetchCharacterCard()
	if err != nil {
		return nil, nil, err
	}
	// Return all the data
	return metadata, characterCard, nil
}

// executeBinderFlow executes the binder flow
func executeBinderFlow(f fetcher.Fetcher, characterID string) (*fetcher.Binder, error) {
	// Fetch metadata response from source
	response, err := reqx.String(f.FetchMetadataResponse(characterID))
	if err != nil {
		// Decorate error if it's not a fetcher error
		if _, ok := err.(*trace.CodedErr[fetcher.ErrCode]); !ok {
			return nil, fetcher.NewError(err, fetcher.FetchMetadataErr)
		}
		// Otherwise, return the error as is
		return nil, err
	}

	// Create metadata binder
	metadataBinder, err := f.CreateBinder(characterID, response)
	if err != nil {
		return nil, err
	}

	// Fetch book responses
	bookBinder, err := f.FetchBookResponses(metadataBinder)
	if err != nil {
		return nil, err
	}

	// Return binder
	return &fetcher.Binder{MetadataBinder: *metadataBinder, BookBinder: *bookBinder}, nil
}

// executeMetadataFlow executes the metadata flow
func executeMetadataFlow(
	f fetcher.Fetcher,
	binderFlow func() (*fetcher.Binder, error),
) (*models.Metadata, error) {
	// Execute binder flow
	binder, err := binderFlow()
	if err != nil {
		return nil, err
	}

	// Fetch card info
	cardInfo, err := f.FetchCardInfo(&binder.MetadataBinder)
	if err != nil {
		return nil, err
	}

	// Fetch creator info
	creatorInfo, err := f.FetchCreatorInfo(&binder.MetadataBinder)
	if err != nil {
		return nil, err
	}

	// Create metadata
	metadata := &models.Metadata{
		Source:         f.SourceID(),
		CardInfo:       *cardInfo,
		CreatorInfo:    *creatorInfo,
		BookUpdateTime: binder.UpdateTime,
		GreetingsCount: -1,
	}

	// Patch metadata
	fetcher.PatchMetadata(metadata)

	// Return metadata
	return metadata, nil
}

// executeCharacterCardFlow executes the character card flow
func executeCharacterCardFlow(
	f fetcher.Fetcher,
	binderFlow func() (*fetcher.Binder, error),
	metadataFlow func() (*models.Metadata, error),
) (*png.CharacterCard, error) {
	// Execute binder flow
	binder, err := binderFlow()
	if err != nil {
		return nil, err
	}

	// Execute metadata flow
	metadata, err := metadataFlow()
	if err != nil {
		return nil, err
	}

	// Fetch character card
	characterCard, err := f.FetchCharacterCard(binder)
	if err != nil {
		return nil, err
	}

	// Patch sheet in the character card
	fetcher.PatchSheet(characterCard.Sheet, metadata)

	// Return character card
	return characterCard, nil
}
