package fetcher

import (
	"cmp"
	"slices"
	"strings"

	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
)

func PatchMetadata(metadata *models.Metadata) {
	creatorNicknameBlank := stringsx.IsBlank(metadata.Nickname)
	creatorUsernameBlank := stringsx.IsBlank(metadata.Username)
	switch {
	case creatorNicknameBlank && !creatorUsernameBlank:
		metadata.Nickname = metadata.Username
	case !creatorNicknameBlank && creatorUsernameBlank:
		metadata.Username = metadata.Nickname
	case creatorNicknameBlank && creatorUsernameBlank:
		metadata.Nickname = character.AnonymousCreator
		metadata.Username = character.AnonymousCreator
		metadata.CreatorInfo.PlatformID = stringsx.Empty
	}
	metadata.NormalizeSymbols()

}

func PatchSheet(sheet *character.Sheet, metadata *models.Metadata) {
	patchNameAndTitle(sheet, metadata)
	patchCreatorNotes(sheet, metadata)
	patchTags(sheet, metadata)
	patchTimestamps(sheet, metadata)
	patchBookName(sheet.CharacterBook, metadata.Name)
	patchLink(sheet, metadata)

	sheet.Creator = property.String(metadata.Nickname)
	sheet.NormalizeSymbols()
}

func patchNameAndTitle(card *character.Sheet, metadata *models.Metadata) {
	card.Title = property.String(metadata.Title)

	var nameSource string
	switch {
	case stringsx.IsNotBlank(metadata.Name):
		// First option (from card info)
		nameSource = metadata.Name
	case stringsx.IsNotBlank(string(card.Name)):
		// Second option (from sheet)
		nameSource = string(card.Name)
	default:
		// Fallback (from card info title)
		nameSource = metadata.Title
	}

	// Synchronize the name in both card and metadata
	card.Name = property.String(nameSource)
	metadata.Name = nameSource

	// Assign the new V3 nickname field
	if stringsx.IsBlank(string(card.Nickname)) {
		card.Nickname = property.String(nameSource)
	}
}

func patchCreatorNotes(card *character.Sheet, metadata *models.Metadata) {
	card.CreatorNotes = property.String(stringsx.JoinNonBlank(character.CreatorNotesSeparator, metadata.Tagline, string(card.CreatorNotes)))
}

func patchTags(card *character.Sheet, metadata *models.Metadata) {
	capacity := len(metadata.Tags) + len(card.Tags)
	mapping := make(map[models.Slug]string, capacity)

	for _, stringTag := range card.Tags {
		cardTag := models.ResolveTag(stringTag)
		mapping[cardTag.Slug] = cardTag.Name
	}

	for _, metadataTag := range metadata.Tags {
		mapping[metadataTag.Slug] = metadataTag.Name
	}

	mergedTags := models.TagsFromMap(mapping)

	slices.SortFunc(mergedTags, func(a, b models.Tag) int {
		return cmp.Compare(a.Slug, b.Slug)
	})

	// Push-back cards from the PNG to the meta-data and vice versa
	metadata.Tags = mergedTags
	card.Tags = models.TagsToNames(mergedTags)
}

func patchTimestamps(card *character.Sheet, metadata *models.Metadata) {
	// Assign the new V3 fields (modification and creation date)
	card.ModificationDate = timestamp.ConvertToSeconds(metadata.LatestUpdateTime())
	card.CreationDate = timestamp.ConvertToSeconds(metadata.CreateTime)
	if metadata.BookUpdateTime == 0 && card.CharacterBook != nil {
		metadata.BookUpdateTime = metadata.UpdateTime
	}
}

func patchBookName(book *character.Book, characterName string) {
	if book == nil {
		return
	}
	if stringsx.IsBlank(string(book.Name)) {
		book.Name = property.String(characterName + " Lore Book")
	} else {
		book.Name = property.String(strings.Replace(string(book.Name), character.BookNamePlaceholder, characterName, 1))
	}
	book.Name = property.String(strings.Replace(string(book.Name), "/", "-", -1))
}

func patchLink(sheet *character.Sheet, metadata *models.Metadata) {
	sheet.SourceID = property.String(metadata.Source)
	sheet.CharacterID = property.String(metadata.CharacterID)
	sheet.PlatformID = property.String(metadata.CardInfo.PlatformID)
	sheet.DirectLink = property.String(metadata.DirectURL)
}
