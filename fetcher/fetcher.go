package fetcher

import (
	"strings"

	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
)

type Fetcher interface {
	SourceHandler
	PatchSheet(sheet *character.Sheet, metadata *models.Metadata)
}

// fetcher - applies any post-processing to the extracted metadata and png card
type fetcher struct {
	SourceHandler
}

func New(handler SourceHandler) Fetcher {
	p := &fetcher{SourceHandler: handler}
	return p
}

func (f *fetcher) FetchCardInfo(metadataBinder *MetadataBinder) (*models.CardInfo, error) {
	cardInfo, err := f.SourceHandler.FetchCardInfo(metadataBinder)
	if err != nil {
		return nil, err
	}
	cardInfo.Name = strings.TrimSpace(cardInfo.Name)
	cardInfo.Title = strings.TrimSpace(cardInfo.Title)
	cardInfo.Tagline = stringsx.FixQuotes(cardInfo.Tagline)
	cardInfo.DirectURL = f.DirectURL(cardInfo.CharacterID)
	return cardInfo, nil
}

func (f *fetcher) PatchSheet(sheet *character.Sheet, metadata *models.Metadata) {
	f.patchNameAndTitle(sheet, metadata)
	f.patchCreatorNotes(sheet, metadata)
	f.patchTags(sheet, metadata)
	f.patchTimestamps(sheet, metadata)
	f.patchBook(sheet.Content.CharacterBook, metadata.CardInfo.Name)

	// Assign the creator to the sheet
	sheet.Content.Creator = metadata.CreatorInfo.Nickname
	// Fix quotes
	sheet.FixQuotes()

	sheet.Content.SourceID = string(metadata.CardInfo.Source)
	sheet.Content.CharacterID = metadata.CardInfo.CharacterID
	sheet.Content.PlatformID = metadata.CardInfo.PlatformID
	sheet.Content.DirectLink = metadata.CardInfo.DirectURL
}

func (f *fetcher) patchNameAndTitle(card *character.Sheet, metadata *models.Metadata) {
	// Assign the real title
	card.Content.Title = metadata.CardInfo.Title

	var nameSource string
	switch {
	case stringsx.IsNotBlank(metadata.CardInfo.Name):
		// First option (from card info)
		nameSource = metadata.CardInfo.Name
	case stringsx.IsNotBlank(card.Content.Name):
		// Second option (from sheet)
		nameSource = card.Content.Name
	default:
		// Fallback (from card info title)
		nameSource = metadata.CardInfo.Title
	}

	// Synchronize the name in both card and metadata
	card.Content.Name = nameSource
	metadata.CardInfo.Name = nameSource

	// Assign the new V3 nickname field
	if stringsx.IsBlankPtr(card.Content.Nickname) {
		card.Content.Nickname = new(string)
		*card.Content.Nickname = nameSource
	}
}

func (f *fetcher) patchCreatorNotes(card *character.Sheet, metadata *models.Metadata) {
	card.Content.CreatorNotes = stringsx.JoinNonBlank(character.CreatorNotesSeparator, metadata.CardInfo.Tagline, card.Content.CreatorNotes)
}

func (f *fetcher) patchTags(card *character.Sheet, metadata *models.Metadata) {
	// Push-back cards from the PNG to the meta-data and vice versa
	tags, stringTags := models.MergeTags(metadata.CardInfo.Tags, card.Content.Tags)
	metadata.CardInfo.Tags = tags
	card.Content.Tags = stringTags
}

func (f *fetcher) patchTimestamps(card *character.Sheet, metadata *models.Metadata) {
	// Assign the new V3 fields (modification and creation date)
	card.Content.ModificationDate = timestamp.Convert[timestamp.Seconds](metadata.LatestUpdateTime())
	card.Content.CreationDate = timestamp.Convert[timestamp.Seconds](metadata.CardInfo.CreateTime)
	if metadata.BookUpdateTime == 0 && card.Content.CharacterBook != nil {
		metadata.BookUpdateTime = metadata.CardInfo.UpdateTime
	}
}

func (f *fetcher) patchBook(book *character.Book, characterName string) {
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
