package postprocessor

import (
	"strings"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
)

// postProcessor - applies any post-processing to the extracted metadata and png card
type postProcessor struct {
	fetcher.Fetcher
}

// New - creates a new post processor based on an existing source
func New(fetcher fetcher.Fetcher) fetcher.Fetcher {
	p := &postProcessor{Fetcher: fetcher}
	return p
}

// FetchMetadata - Retrieve metadata for given url using the underlying source (and applies post-processing)
func (processor *postProcessor) FetchMetadata(normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	metadata, gJsonResponse, err := processor.Fetcher.FetchMetadata(normalizedURL, characterID)
	if err != nil {
		return nil, models.EmptyJsonResponse, err
	}
	processor.patchMetadata(metadata)
	return metadata, gJsonResponse, nil
}

// FetchPngCard - Retrieve card for given url using the underlying source (and applies post-processing)
func (processor *postProcessor) FetchCharacterCard(metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	characterCard, err := processor.Fetcher.FetchCharacterCard(metadata, response)
	if err != nil {
		return characterCard, err
	}
	processor.patchCard(metadata, characterCard.Sheet)
	return characterCard, nil
}

// patchMetadata - patches metadata (strips all trailing spaces from card name, creator and tagline)
func (processor *postProcessor) patchMetadata(metadata *models.Metadata) {
	metadata.CardName = strings.TrimSpace(metadata.CardName)
	metadata.CharacterName = strings.TrimSpace(metadata.CharacterName)
	metadata.Creator = strings.TrimSpace(metadata.Creator)
	metadata.Tagline = stringsx.FixQuotes(metadata.Tagline)
	metadata.DirectURL = processor.DirectURL(metadata.CharacterID)
}

// patchCard - patches the card with the relevant information from metadata, making sure the card matches the metadata
func (processor *postProcessor) patchCard(metadata *models.Metadata, card *character.Sheet) {
	characterName := processor.patchCardName(metadata, card)
	processor.patchTags(metadata, card)
	processor.patchTimestamps(metadata, card)
	processor.patchBook(card.Data.CharacterBook, characterName)

	// Assign the creator to the card
	card.Data.Creator = metadata.Creator
	// Fix quotes
	card.FixQuotes()

	card.Data.SourceID = string(metadata.Source)
	card.Data.CharacterID = metadata.CharacterID
	card.Data.PlatformID = metadata.PlatformID
	card.Data.DirectLink = metadata.DirectURL
}

func (processor *postProcessor) patchCardName(metadata *models.Metadata, card *character.Sheet) string {
	// Assign the real card name
	card.Data.CardName = metadata.CardName

	var nameSource string
	switch {
	case stringsx.IsNotBlank(metadata.CharacterName):
		// First option (from metadata - character name)
		nameSource = metadata.CharacterName
	case stringsx.IsNotBlank(card.Data.CharacterName):
		// Second option (from card - character name)
		nameSource = card.Data.CharacterName
	default:
		// Fallback (from metadata - card name)
		nameSource = metadata.CardName
	}

	// Synchronize the name in both card and metadata
	card.Data.CharacterName = nameSource
	metadata.CharacterName = nameSource

	// Assign the new V3 nickname field
	if stringsx.IsBlankPtr(card.Data.Nickname) {
		card.Data.Nickname = new(string)
		*card.Data.Nickname = nameSource
	}

	return nameSource
}

func (processor *postProcessor) patchTags(metadata *models.Metadata, card *character.Sheet) {
	// Push-back cards from the PNG to the meta-data and vice versa
	tags, stringTags := models.MergeTags(metadata.Tags, card.Data.Tags)
	metadata.Tags = tags
	card.Data.Tags = stringTags
}

func (processor *postProcessor) patchTimestamps(metadata *models.Metadata, card *character.Sheet) {
	// Assign the new V3 fields (modification and creation date)
	card.Data.ModificationDate = timestamp.Convert[timestamp.Seconds](metadata.LatestUpdateTime())
	card.Data.CreationDate = timestamp.Convert[timestamp.Seconds](metadata.CreateTime)
	if metadata.BookUpdateTime == 0 && card.Data.CharacterBook != nil {
		metadata.BookUpdateTime = metadata.UpdateTime
	}
}

func (processor *postProcessor) patchBook(book *character.Book, characterName string) {
	if book == nil {
		return
	}
	if stringsx.IsBlankPtr(book.Name) {
		book.Name = new(string)
		*book.Name = characterName + " Lore Book"
	} else {
		*book.Name = strings.Replace(*book.Name, character.BookNamePlaceholder, characterName, 1)
	}
	*book.Name = strings.Replace(*book.Name, "/", "-", -1)
	book.MirrorNameAndComment()
}
