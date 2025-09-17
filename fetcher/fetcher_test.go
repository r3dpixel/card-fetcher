package fetcher

import (
	"testing"
	"time"

	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/stretchr/testify/assert"
)

func TestFetcher_PatchSheet(t *testing.T) {
	// Create a minimal fetcher for testing
	f := &fetcher{}

	t.Run("should patch all basic fields", func(t *testing.T) {
		now := time.Now()
		createTime := now.Add(-24 * time.Hour)
		updateTime := now.Add(-12 * time.Hour)
		bookUpdateTime := now

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:        source.ID("test"),
				NormalizedURL: "example.com/characters/test-char-123",
				DirectURL:     "example.com/direct/test-char-123",
				PlatformID:    "platform-123",
				CharacterID:   "test-char-123",
				Name:          "Test Character",
				Title:         "Test Card Title",
				Tagline:       "This is a test tagline.",
				CreateTime:    timestamp.Nano(createTime.UnixNano()),
				UpdateTime:    timestamp.Nano(updateTime.UnixNano()),
				Tags:          []models.Tag{{Slug: "fantasy", Name: "Fantasy"}},
			},
			CreatorInfo: models.CreatorInfo{
				Nickname:   "Test Creator",
				Username:   "testcreator",
				PlatformID: "creator-123",
			},
			BookUpdateTime: timestamp.Nano(bookUpdateTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:         "Original Name",
				Title:        "Original Title",
				CreatorNotes: "Original notes",
				Tags:         []string{"Original Tag"},
			},
		}

		f.PatchSheet(sheet, metadata)

		// Test basic field assignments
		assert.Equal(t, "test", sheet.Content.SourceID)
		assert.Equal(t, "test-char-123", sheet.Content.CharacterID)
		assert.Equal(t, "platform-123", sheet.Content.PlatformID)
		assert.Equal(t, "example.com/direct/test-char-123", sheet.Content.DirectLink)
		assert.Equal(t, "Test Creator", sheet.Content.Creator)

		// Test name and title patching
		assert.Equal(t, "Test Character", sheet.Content.Name)
		assert.Equal(t, "Test Card Title", sheet.Content.Title)
		assert.NotNil(t, sheet.Content.Nickname)
		assert.Equal(t, "Test Character", *sheet.Content.Nickname)

		// Test creator notes patching (tagline should be prepended)
		assert.Contains(t, sheet.Content.CreatorNotes, "This is a test tagline.")
		assert.Contains(t, sheet.Content.CreatorNotes, "Original notes")

		// Test timestamp patching
		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(bookUpdateTime.Unix()), sheet.Content.ModificationDate) // Uses latest update time

		// Test tags merging
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Original Tag")
	})
}

func TestFetcher_PatchNameAndTitle(t *testing.T) {
	f := &fetcher{}

	t.Run("should prefer metadata name over sheet name", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "Metadata Name",
				Title: "Metadata Title",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:  "Sheet Name",
				Title: "Sheet Title",
			},
		}

		f.patchNameAndTitle(sheet, metadata)

		assert.Equal(t, "Metadata Name", sheet.Content.Name)
		assert.Equal(t, "Metadata Title", sheet.Content.Title)
		assert.Equal(t, "Metadata Name", metadata.CardInfo.Name) // Should sync back
	})

	t.Run("should fallback to sheet name if metadata name is blank", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "",
				Title: "Metadata Title",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:  "Sheet Name",
				Title: "Sheet Title",
			},
		}

		f.patchNameAndTitle(sheet, metadata)

		assert.Equal(t, "Sheet Name", sheet.Content.Name)
		assert.Equal(t, "Sheet Name", metadata.CardInfo.Name) // Should sync back
	})

	t.Run("should fallback to title if both names are blank", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "",
				Title: "Fallback Title",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:  "",
				Title: "Sheet Title",
			},
		}

		f.patchNameAndTitle(sheet, metadata)

		assert.Equal(t, "Fallback Title", sheet.Content.Name)
		assert.Equal(t, "Fallback Title", metadata.CardInfo.Name) // Should sync back
	})

	t.Run("should set nickname if empty", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "Character Name",
				Title: "Card Title",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Nickname: nil,
			},
		}

		f.patchNameAndTitle(sheet, metadata)

		assert.NotNil(t, sheet.Content.Nickname)
		assert.Equal(t, "Character Name", *sheet.Content.Nickname)
	})

	t.Run("should not overwrite existing nickname", func(t *testing.T) {
		existingNickname := "Existing Nickname"
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "Character Name",
				Title: "Card Title",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Nickname: &existingNickname,
			},
		}

		f.patchNameAndTitle(sheet, metadata)

		assert.Equal(t, "Existing Nickname", *sheet.Content.Nickname)
	})
}

func TestFetcher_PatchCreatorNotes(t *testing.T) {
	f := &fetcher{}

	t.Run("should join tagline with existing notes", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "Character tagline",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Existing notes",
			},
		}

		f.patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "Character tagline", "Existing notes")
		assert.Equal(t, expected, sheet.Content.CreatorNotes)
	})

	t.Run("should handle empty tagline", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Existing notes",
			},
		}

		f.patchCreatorNotes(sheet, metadata)

		assert.Equal(t, "Existing notes", sheet.Content.CreatorNotes)
	})

	t.Run("should handle empty existing notes", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "Character tagline",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "",
			},
		}

		f.patchCreatorNotes(sheet, metadata)

		assert.Equal(t, "Character tagline", sheet.Content.CreatorNotes)
	})
}

func TestFetcher_PatchTags(t *testing.T) {
	f := &fetcher{}

	t.Run("should merge tags from metadata and sheet", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "fantasy", Name: "Fantasy"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Adventure", "Romance"},
			},
		}

		f.patchTags(sheet, metadata)

		// Should contain all unique tags
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Adventure")
		assert.Contains(t, sheet.Content.Tags, "Romance")
	})

	t.Run("should handle empty metadata tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Adventure"},
			},
		}

		f.patchTags(sheet, metadata)

		assert.Equal(t, []string{"Adventure"}, sheet.Content.Tags)
	})

	t.Run("should handle empty sheet tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "fantasy", Name: "Fantasy"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{},
			},
		}

		f.patchTags(sheet, metadata)

		assert.Equal(t, []string{"Fantasy"}, sheet.Content.Tags)
	})
}

func TestFetcher_PatchTimestamps(t *testing.T) {
	f := &fetcher{}

	t.Run("should set timestamps correctly", func(t *testing.T) {
		now := time.Now()
		createTime := now.Add(-24 * time.Hour)
		updateTime := now.Add(-12 * time.Hour)
		bookUpdateTime := now

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(createTime.UnixNano()),
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(bookUpdateTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		f.patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(bookUpdateTime.Unix()), sheet.Content.ModificationDate) // Uses latest
	})

	t.Run("should set book update time if zero and book exists", func(t *testing.T) {
		now := time.Now()
		updateTime := now.Add(-12 * time.Hour)

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: 0, // Zero
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CharacterBook: &character.Book{}, // Book exists
			},
		}

		f.patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Nano(updateTime.UnixNano()), metadata.BookUpdateTime)
	})

	t.Run("should not set book update time if book is nil", func(t *testing.T) {
		metadata := &models.Metadata{
			BookUpdateTime: 0,
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CharacterBook: nil, // No book
			},
		}

		f.patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Nano(0), metadata.BookUpdateTime)
	})
}

func TestFetcher_PatchBook(t *testing.T) {
	f := &fetcher{}

	t.Run("should handle nil book", func(t *testing.T) {
		assert.NotPanics(t, func() {
			f.patchBook(nil, "any name")
		})
	})

	t.Run("should set name if blank", func(t *testing.T) {
		book := &character.Book{Name: nil}
		f.patchBook(book, "Test Character")

		assert.NotNil(t, book.Name)
		assert.Equal(t, "Test Character Lore Book", *book.Name)
	})

	t.Run("should set name if empty string", func(t *testing.T) {
		emptyName := ""
		book := &character.Book{Name: &emptyName}
		f.patchBook(book, "Test Character")

		assert.Equal(t, "Test Character Lore Book", *book.Name)
	})

	t.Run("should replace placeholder", func(t *testing.T) {
		bookName := "Lore for " + character.BookNamePlaceholder
		book := &character.Book{Name: &bookName}
		f.patchBook(book, "Test Character")

		assert.Equal(t, "Lore for Test Character", *book.Name)
	})

	t.Run("should replace slashes with hyphens", func(t *testing.T) {
		bookName := "A/B/C"
		book := &character.Book{Name: &bookName}
		f.patchBook(book, "Test Character")

		assert.Equal(t, "A-B-C", *book.Name)
	})

	t.Run("should replace slashes in generated name", func(t *testing.T) {
		book := &character.Book{Name: nil}
		f.patchBook(book, "Test/Character")

		assert.Equal(t, "Test-Character Lore Book", *book.Name)
	})
}
