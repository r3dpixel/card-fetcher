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

// PatchMetadata ensures that the nickname and username fields are not blank
func PatchMetadata(metadata *models.Metadata) {
	// Check if the creator nickname is blank
	creatorNicknameBlank := stringsx.IsBlank(metadata.Nickname)
	// Check if the creator username is blank
	creatorUsernameBlank := stringsx.IsBlank(metadata.Username)
	switch {
	// If the creator nickname is blank, and the creator username is not, set the nickname to the username
	case creatorNicknameBlank && !creatorUsernameBlank:
		metadata.Nickname = metadata.Username
	// If the creator nickname is not blank, and the creator username is blank, set the username to the nickname
	case !creatorNicknameBlank && creatorUsernameBlank:
		metadata.Username = metadata.Nickname
	// If both the creator nickname and username are blank, set them to the anonymous creator
	case creatorNicknameBlank && creatorUsernameBlank:
		metadata.Nickname = character.AnonymousCreator
		metadata.Username = character.AnonymousCreator
		metadata.CreatorInfo.PlatformID = ""
	}
	// Normalize symbols in metadata
	metadata.NormalizeSymbols()

}

// PatchSheet ensures that the sheet is consistent with the metadata
func PatchSheet(sheet *character.Sheet, metadata *models.Metadata) {
	// Patch name and title
	patchNameAndTitle(sheet, metadata)
	// Patch creator notes
	patchCreatorNotes(sheet, metadata)
	// Patch tags
	patchTags(sheet, metadata)
	// Patch timestamps
	patchTimestamps(sheet, metadata)
	// Patch book name
	patchBookName(sheet.CharacterBook, metadata.Name)
	// Patch meta fields
	patchMetaFields(sheet, metadata)

	// Set greetings count
	metadata.GreetingsCount = len(sheet.AlternateGreetings)

	// Set has book flag
	metadata.HasBook = sheet.CharacterBook != nil

	// Set the creator of the sheet to the nickname
	sheet.Creator = property.String(metadata.Nickname)

	// Fix user templates
	sheet.FixUserCharTemplates()

	// Normalize symbols in sheet
	sheet.NormalizeSymbols()
}

// patchNameAndTitle ensures that the name and title fields are consistent with the metadata
func patchNameAndTitle(sheet *character.Sheet, metadata *models.Metadata) {
	// Set the title to the metadata title
	sheet.Title = property.String(metadata.Title)

	// Select the name source
	var nameSource string
	switch {
	// If the metadata name is not blank, use it
	case stringsx.IsNotBlank(metadata.Name):
		// First option (from metadata)
		nameSource = metadata.Name
	// If the sheet name is not blank, use it
	case stringsx.IsNotBlank(string(sheet.Name)):
		// Second option (from the sheet)
		nameSource = string(sheet.Name)
	// Otherwise, use the metadata title
	default:
		// Fallback (from sheet info title)
		nameSource = metadata.Title
	}

	// Synchronize the name in both sheet and metadata
	sheet.Name = property.String(nameSource)
	metadata.Name = nameSource

	// Assign the new V3 nickname field
	if stringsx.IsBlank(string(sheet.Nickname)) {
		sheet.Nickname = property.String(nameSource)
	}
}

// patchCreatorNotes ensures that the creator notes field is consistent with the metadata
func patchCreatorNotes(sheet *character.Sheet, metadata *models.Metadata) {
	// Join the tagline with the existing notes, using the separator
	sheet.CreatorNotes = property.String(
		stringsx.JoinNonBlank(
			character.CreatorNotesSeparator,
			metadata.Tagline,
			string(sheet.CreatorNotes),
		),
	)
}

// patchTags ensures that the tags field is consistent with the metadata
func patchTags(sheet *character.Sheet, metadata *models.Metadata) {
	// Create a map of tags from the sheet and the metadata
	capacity := len(metadata.Tags) + len(sheet.Tags)
	mapping := make(map[models.Slug]string, capacity)

	// Iterate over the sheet tags
	for _, stringTag := range sheet.Tags {
		// Resolve the tag slug and name
		cardTag := models.ResolveTag(stringTag)
		// Add the tag to the map
		mapping[cardTag.Slug] = cardTag.Name
	}

	// Iterate over the metadata tags
	for _, metadataTag := range metadata.Tags {
		// Add the tag to the map
		mapping[metadataTag.Slug] = metadataTag.Name
	}

	// Add the source tag to the map
	sourceTag := models.ResolveTag(string(metadata.Source))
	if _, ok := mapping[sourceTag.Slug]; !ok {
		mapping[sourceTag.Slug] = sourceTag.Name
	}

	// Create a slice of tags from the map
	mergedTags := models.TagsFromMap(mapping)

	// Sort the tags by slug
	slices.SortFunc(mergedTags, func(a, b models.Tag) int {
		return cmp.Compare(a.Slug, b.Slug)
	})

	// Synchronize the tags in both sheet and metadata
	metadata.Tags = mergedTags
	sheet.Tags = models.TagsToNames(mergedTags)
}

// patchTimestamps ensures that the timestamp fields are consistent with the metadata
func patchTimestamps(sheet *character.Sheet, metadata *models.Metadata) {
	// Assign the new V3 fields (modification and creation date), converting nanoseconds to seconds
	sheet.ModificationDate = timestamp.ConvertToSeconds(metadata.LatestUpdateTime())
	sheet.CreationDate = timestamp.ConvertToSeconds(metadata.CreateTime)

	// If the book update time is zero and the character book is not nil, set the book update time to the latest update time
	if metadata.BookUpdateTime == 0 && sheet.CharacterBook != nil {
		metadata.BookUpdateTime = metadata.UpdateTime
	}
}

// patchBookName ensures that the book name is consistent with the metadata
func patchBookName(book *character.Book, characterName string) {
	// Return if the book is nil
	if book == nil {
		return
	}
	// If the book name is blank, set it to the character name with the Lore Book suffix
	if stringsx.IsBlank(string(book.Name)) {
		book.Name = property.String(characterName + " Lore Book")
	} else {
		// Replace the placeholder with the character name
		book.Name = property.String(strings.Replace(string(book.Name), character.BookNamePlaceholder, characterName, 1))
	}
	// Replace slashes with dashes in the book name
	book.Name = property.String(strings.Replace(string(book.Name), "/", "-", -1))
}

// patchMetaFields ensures that the meta-fields are consistent with the metadata
func patchMetaFields(sheet *character.Sheet, metadata *models.Metadata) {
	sheet.SourceID = property.String(metadata.Source)
	sheet.CharacterID = property.String(metadata.CharacterID)
	sheet.PlatformID = property.String(metadata.CardInfo.PlatformID)
	sheet.DirectLink = property.String(metadata.DirectURL)
}
