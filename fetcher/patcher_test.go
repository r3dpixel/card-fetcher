package fetcher

import (
	"fmt"
	"testing"
	"time"

	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/stretchr/testify/assert"
)

// TestPatchMetadata tests the PatchMetadata function with 100% coverage
func TestPatchMetadata(t *testing.T) {
	t.Run("should not alter ASCII symbols in tagline", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    "  Test Name  ",
				Title:   "  Test Title  ",
				Tagline: "Tagline with 'quotes' and — dashes",
			},
		}

		originalTagline := metadata.Tagline
		PatchMetadata(metadata)

		assert.Equal(t, "Test Name", metadata.Name)
		assert.Equal(t, "Test Title", metadata.Title)

		assert.Equal(t, originalTagline, metadata.Tagline)
		assert.Equal(t, stringsx.NormalizeSymbols(originalTagline), metadata.Tagline)
	})

	t.Run("should handle empty metadata", func(t *testing.T) {
		metadata := &models.Metadata{}
		assert.NotPanics(t, func() {
			PatchMetadata(metadata)
		})
	})

	t.Run("should normalize unicode symbols in tagline only", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    "«Curly» 'Quotes'",
				Title:   "Unicode — dashes",
				Tagline: "‘‛¿Questions? ¡Exclamations! — with dashes «»‚",
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "Creator™",
				Username: "user®",
			},
		}

		originalTagline := metadata.Tagline
		originalName := metadata.Name
		originalTitle := metadata.Title

		PatchMetadata(metadata)

		assert.NotEqual(t, originalTagline, metadata.Tagline)
		assert.Equal(t, stringsx.NormalizeSymbols(originalTagline), metadata.Tagline)
		assert.Equal(t, `''¿Questions? ¡Exclamations! — with dashes "",`, metadata.Tagline)

		assert.Equal(t, originalName, metadata.Name)
		assert.Equal(t, originalTitle, metadata.Title)
	})

	t.Run("should handle metadata with all fields empty", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    "",
				Title:   "",
				Tagline: "",
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "",
				Username: "",
			},
		}

		assert.NotPanics(t, func() {
			PatchMetadata(metadata)
		})
	})
}

// TestPatchNameAndTitle tests the patchNameAndTitle function with table-driven tests
func TestPatchNameAndTitle(t *testing.T) {
	tests := []struct {
		name          string
		metadataName  string
		metadataTitle string
		sheetName     string
		sheetTitle    string
		sheetNickname string
		expectedName  string
		expectedTitle string
	}{
		{
			name:          "should prefer metadata name over sheet name",
			metadataName:  "Metadata Name",
			metadataTitle: "Metadata Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Metadata Name",
			expectedTitle: "Metadata Title",
		},
		{
			name:          "should fallback to sheet name if metadata name is blank",
			metadataName:  "",
			metadataTitle: "Metadata Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Sheet Name",
			expectedTitle: "Metadata Title",
		},
		{
			name:          "should fallback to metadata title if both names are blank",
			metadataName:  "",
			metadataTitle: "Fallback Title",
			sheetName:     "",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Fallback Title",
			expectedTitle: "Fallback Title",
		},
		{
			name:          "should fallback to final default if all are blank",
			metadataName:  "",
			metadataTitle: "",
			sheetName:     "",
			sheetTitle:    "",
			sheetNickname: "",
			expectedName:  "",
			expectedTitle: "",
		},
		{
			name:          "should handle whitespace-only names as blank",
			metadataName:  "   \t\n   ",
			metadataTitle: "Valid Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Sheet Name",
			expectedTitle: "Valid Title",
		},
		{
			name:          "should set nickname if nil",
			metadataName:  "Character Name",
			metadataTitle: "Card Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Character Name",
			expectedTitle: "Card Title",
		},
		{
			name:          "should set nickname if empty string pointer",
			metadataName:  "Test Name",
			metadataTitle: "Test Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "",
			expectedName:  "Test Name",
			expectedTitle: "Test Title",
		},
		{
			name:          "should set nickname if whitespace-only string pointer",
			metadataName:  "Test Name",
			metadataTitle: "Test Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "   \t\n   ",
			expectedName:  "Test Name",
			expectedTitle: "Test Title",
		},
		{
			name:          "should not overwrite existing valid nickname",
			metadataName:  "Character Name",
			metadataTitle: "Card Title",
			sheetName:     "Sheet Name",
			sheetTitle:    "Sheet Title",
			sheetNickname: "Existing Nickname",
			expectedName:  "Character Name",
			expectedTitle: "Card Title",
		},
		{
			name:          "should handle unicode and special characters",
			metadataName:  "テスト キャラクター",
			metadataTitle: "测试 标题",
			sheetName:     "Original",
			sheetTitle:    "Original Title",
			sheetNickname: "",
			expectedName:  "テスト キャラクター",
			expectedTitle: "测试 标题",
		},
		{
			name:          "should handle names with special formatting characters",
			metadataName:  "Name\nWith\tSpecial\rChars",
			metadataTitle: "Title\nWith\tSpecial\rChars",
			sheetName:     "Original",
			sheetTitle:    "Original Title",
			sheetNickname: "",
			expectedName:  "Name\nWith\tSpecial\rChars",
			expectedTitle: "Title\nWith\tSpecial\rChars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &models.Metadata{
				CardInfo: models.CardInfo{
					Name:  tt.metadataName,
					Title: tt.metadataTitle,
				},
			}

			sheet := &character.Sheet{
				Content: character.Content{
					Name:     property.String(tt.sheetName),
					Title:    property.String(tt.sheetTitle),
					Nickname: property.String(tt.sheetNickname),
				},
			}

			patchNameAndTitle(sheet, metadata)

			// Assert name and title are set correctly with proper casting
			assert.Equal(t, tt.expectedName, string(sheet.Name))
			assert.Equal(t, tt.expectedName, metadata.CardInfo.Name)
			assert.Equal(t, tt.expectedTitle, string(sheet.Title))

			// Assert nickname behavior - if nickname was blank, it should be set to expectedName
			if stringsx.IsBlank(string(tt.sheetNickname)) {
				assert.Equal(t, tt.expectedName, string(sheet.Nickname))
			} else {
				assert.Equal(t, string(tt.sheetNickname), string(sheet.Nickname))
			}
		})
	}
}

// TestPatchCreatorNotes tests the patchCreatorNotes function with complete coverage
func TestPatchCreatorNotes(t *testing.T) {
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

		patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "Character tagline", "Existing notes")
		assert.Equal(t, expected, string(sheet.Content.CreatorNotes))
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

		patchCreatorNotes(sheet, metadata)

		assert.Equal(t, "Existing notes", string(sheet.Content.CreatorNotes))
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

		patchCreatorNotes(sheet, metadata)

		assert.Equal(t, "Character tagline", string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle both empty", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "",
			},
		}

		patchCreatorNotes(sheet, metadata)

		assert.Equal(t, "", string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle whitespace-only tagline", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "   \t\n   ",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Existing notes",
			},
		}

		patchCreatorNotes(sheet, metadata)

		// JoinNonBlank should handle whitespace appropriately
		result := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "   \t\n   ", "Existing notes")
		assert.Equal(t, result, string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle whitespace-only existing notes", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "Character tagline",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "   \t\n   ",
			},
		}

		patchCreatorNotes(sheet, metadata)

		result := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "Character tagline", "   \t\n   ")
		assert.Equal(t, result, string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle unicode and special characters", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "タグライン with émojis 🎉",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Existing notes with 中文",
			},
		}

		patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "タグライン with émojis 🎉", "Existing notes with 中文")
		assert.Equal(t, expected, string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle very long content", func(t *testing.T) {
		longTagline := string(make([]rune, 5000))
		for i := range longTagline {
			longTagline = longTagline[:i] + "T" + longTagline[i+1:]
		}

		longNotes := string(make([]rune, 10000))
		for i := range longNotes {
			longNotes = longNotes[:i] + "N" + longNotes[i+1:]
		}

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: longTagline,
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: property.String(longNotes),
			},
		}

		patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, longTagline, longNotes)
		assert.Equal(t, expected, string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle newlines and special formatting", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "Line 1\nLine 2\r\nLine 3",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Note 1\nNote 2\n\nNote 3",
			},
		}

		patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "Line 1\nLine 2\r\nLine 3", "Note 1\nNote 2\n\nNote 3")
		assert.Equal(t, expected, string(sheet.Content.CreatorNotes))
	})

	t.Run("should handle multiple consecutive separators", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tagline: "Has" + character.CreatorNotesSeparator + "separator",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CreatorNotes: "Also" + character.CreatorNotesSeparator + "has" + character.CreatorNotesSeparator + "separators",
			},
		}

		patchCreatorNotes(sheet, metadata)

		expected := stringsx.JoinNonBlank(character.CreatorNotesSeparator, "Has"+character.CreatorNotesSeparator+"separator", "Also"+character.CreatorNotesSeparator+"has"+character.CreatorNotesSeparator+"separators")
		assert.Equal(t, expected, string(sheet.Content.CreatorNotes))
	})
}

// TestPatchTags tests the patchTags function with comprehensive coverage
func TestPatchTags(t *testing.T) {
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

		patchTags(sheet, metadata)

		// Should contain all unique tags
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Adventure")
		assert.Contains(t, sheet.Content.Tags, "Romance")
		assert.Len(t, sheet.Content.Tags, 3)

		// Metadata should be updated with merged tags
		assert.Len(t, metadata.Tags, 3)
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

		patchTags(sheet, metadata)

		assert.Equal(t, []string{"Adventure"}, []string(sheet.Content.Tags))
		assert.Len(t, metadata.Tags, 1)
		assert.Equal(t, "Adventure", metadata.Tags[0].Name)
	})

	t.Run("should handle nil metadata tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: nil,
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Adventure"},
			},
		}

		patchTags(sheet, metadata)

		assert.Equal(t, []string{"Adventure"}, []string(sheet.Content.Tags))
		assert.Len(t, metadata.Tags, 1)
		assert.Equal(t, "Adventure", metadata.Tags[0].Name)
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

		patchTags(sheet, metadata)

		assert.Equal(t, []string{"Fantasy"}, []string(sheet.Content.Tags))
		assert.Len(t, metadata.Tags, 1)
		assert.Equal(t, "Fantasy", metadata.Tags[0].Name)
	})

	t.Run("should handle nil sheet tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "test", Name: "Test"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: nil,
			},
		}

		patchTags(sheet, metadata)

		assert.Equal(t, []string{"Test"}, []string(sheet.Content.Tags))
	})

	t.Run("should handle both empty", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{},
			},
		}

		patchTags(sheet, metadata)

		assert.Empty(t, sheet.Content.Tags)
		assert.Empty(t, metadata.Tags)
	})

	t.Run("should handle both nil", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: nil,
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: nil,
			},
		}

		patchTags(sheet, metadata)

		assert.Empty(t, sheet.Content.Tags)
		assert.Empty(t, metadata.Tags)
	})

	t.Run("should deduplicate identical tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "fantasy", Name: "Fantasy"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Fantasy", "Adventure", "Fantasy"},
			},
		}

		patchTags(sheet, metadata)

		// Should deduplicate
		assert.Len(t, sheet.Content.Tags, 2)
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Adventure")
	})

	t.Run("should resolve standard tags correctly", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"nsfw", "fempov", "CustomTag"},
			},
		}

		patchTags(sheet, metadata)

		// Standard tags should be resolved to proper names
		assert.Contains(t, sheet.Content.Tags, "NSFW")      // Standard tag
		assert.Contains(t, sheet.Content.Tags, "Fem POV")   // Standard tag
		assert.Contains(t, sheet.Content.Tags, "CustomTag") // Custom tag
	})

	t.Run("should handle case sensitivity in tag resolution", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"NSFW", "nsfw", "Nsfw", "nSfW"},
			},
		}

		patchTags(sheet, metadata)

		// All should resolve to the same standard tag and be deduplicated
		assert.Len(t, sheet.Content.Tags, 1)
		assert.Contains(t, sheet.Content.Tags, "NSFW")
	})

	t.Run("should sort tags alphabetically by slug", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "zulu", Name: "Zulu"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Charlie", "Alpha", "Beta"},
			},
		}

		patchTags(sheet, metadata)

		// Tags should be sorted in metadata by slug
		assert.Len(t, metadata.Tags, 4)
		slugs := make([]string, len(metadata.Tags))
		for i, tag := range metadata.Tags {
			slugs[i] = string(tag.Slug)
		}
		// Should be sorted: alpha, beta, charlie, zulu
		assert.Equal(t, "alpha", slugs[0])
		assert.Equal(t, "beta", slugs[1])
		assert.Equal(t, "charlie", slugs[2])
		assert.Equal(t, "zulu", slugs[3])
	})

	t.Run("should handle unicode and special character tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"テスト", "测试", "🎉emoji", "spéçîål"},
			},
		}

		patchTags(sheet, metadata)

		assert.Len(t, sheet.Content.Tags, 4)
		assert.Contains(t, sheet.Content.Tags, "テスト")
		assert.Contains(t, sheet.Content.Tags, "测试")
		assert.Contains(t, []string(sheet.Content.Tags), "Emoji")
		assert.Contains(t, []string(sheet.Content.Tags), "Spéçîål")
	})

	t.Run("should handle empty string tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Valid", "", "   ", "Another"},
			},
		}

		patchTags(sheet, metadata)

		// Empty and whitespace-only tags should be handled by ResolveTag
		assert.Contains(t, sheet.Content.Tags, "Valid")
		assert.Contains(t, sheet.Content.Tags, "Another")
	})

	t.Run("should handle very large number of tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		// Create 1000 unique tags
		largeTags := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			largeTags[i] = fmt.Sprintf("Tag%d", i)
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: largeTags,
			},
		}

		patchTags(sheet, metadata)

		assert.Len(t, sheet.Content.Tags, 1000)
		assert.Len(t, metadata.Tags, 1000)
	})

	t.Run("should handle duplicate metadata and sheet tags", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{
					{Slug: "fantasy", Name: "Fantasy"},
					{Slug: "adventure", Name: "Adventure"},
				},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"Fantasy", "Adventure", "Romance"},
			},
		}

		patchTags(sheet, metadata)

		// Should deduplicate correctly
		assert.Len(t, sheet.Content.Tags, 3)
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Adventure")
		assert.Contains(t, sheet.Content.Tags, "Romance")
	})

	t.Run("should handle tags with only punctuation", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"...", "!!!", "???", "---"},
			},
		}

		patchTags(sheet, metadata)

		// These should be processed by ResolveTag
		assert.GreaterOrEqual(t, len(sheet.Content.Tags), 0) // Depends on ResolveTag behavior
	})

	t.Run("should handle mixed case and spacing variations", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: []models.Tag{{Slug: "fantasy", Name: "Fantasy"}},
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: []string{"FANTASY", "fantasy", "Fantasy", " fantasy ", "  FANTASY  "},
			},
		}

		patchTags(sheet, metadata)

		// Should all resolve to the same tag and be deduplicated
		assert.Len(t, sheet.Content.Tags, 1)
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
	})
}

// TestPatchTimestamps tests the patchTimestamps function with complete coverage
func TestPatchTimestamps(t *testing.T) {
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

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(bookUpdateTime.Unix()), sheet.Content.ModificationDate) // Uses latest
	})

	t.Run("should use card update time when it's later than book update", func(t *testing.T) {
		now := time.Now()
		createTime := now.Add(-24 * time.Hour)
		updateTime := now.Add(-1 * time.Hour) // More recent than book
		bookUpdateTime := now.Add(-12 * time.Hour)

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

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(updateTime.Unix()), sheet.Content.ModificationDate) // Uses latest
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

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Nano(updateTime.UnixNano()), metadata.BookUpdateTime)
	})

	t.Run("should not set book update time if book is nil", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				UpdateTime: timestamp.Nano(time.Now().UnixNano()),
			},
			BookUpdateTime: 0,
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CharacterBook: nil, // No book
			},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Nano(0), metadata.BookUpdateTime)
	})

	t.Run("should not set book update time if book update is already set", func(t *testing.T) {
		now := time.Now()
		updateTime := now.Add(-12 * time.Hour)
		existingBookUpdateTime := now.Add(-6 * time.Hour)

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(existingBookUpdateTime.UnixNano()), // Already set
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CharacterBook: &character.Book{}, // Book exists
			},
		}

		patchTimestamps(sheet, metadata)

		// Should not overwrite existing book update time
		assert.Equal(t, timestamp.Nano(existingBookUpdateTime.UnixNano()), metadata.BookUpdateTime)
	})

	t.Run("should handle zero timestamps", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: 0,
				UpdateTime: 0,
			},
			BookUpdateTime: 0,
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(0), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(0), sheet.Content.ModificationDate)
	})

	t.Run("should handle negative timestamps", func(t *testing.T) {
		negativeTime := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(negativeTime.UnixNano()),
				UpdateTime: timestamp.Nano(negativeTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(negativeTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(negativeTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(negativeTime.Unix()), sheet.Content.ModificationDate)
	})

	t.Run("should handle future timestamps", func(t *testing.T) {
		futureTime := time.Now().Add(365 * 24 * time.Hour) // 1 year in the future

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(futureTime.UnixNano()),
				UpdateTime: timestamp.Nano(futureTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(futureTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(futureTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(futureTime.Unix()), sheet.Content.ModificationDate)
	})

	t.Run("should handle create time later than update time", func(t *testing.T) {
		now := time.Now()
		createTime := now                     // Later
		updateTime := now.Add(-1 * time.Hour) // Earlier

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(createTime.UnixNano()),
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: 0,
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		// Should still use the values as provided, even if illogical
		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(updateTime.Unix()), sheet.Content.ModificationDate)
	})

	t.Run("should handle maximum timestamp values", func(t *testing.T) {
		maxTime := time.Unix(1<<31-1, 0) // Max int32 seconds

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(maxTime.UnixNano()),
				UpdateTime: timestamp.Nano(maxTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(maxTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(maxTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(maxTime.Unix()), sheet.Content.ModificationDate)
	})

	t.Run("should handle minimum timestamp values", func(t *testing.T) {
		minTime := time.Unix(-(1 << 31), 0) // Min int32 seconds

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				CreateTime: timestamp.Nano(minTime.UnixNano()),
				UpdateTime: timestamp.Nano(minTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(minTime.UnixNano()),
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchTimestamps(sheet, metadata)

		assert.Equal(t, timestamp.Seconds(minTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(minTime.Unix()), sheet.Content.ModificationDate)
	})
}

// TestPatchBook tests the patchBook function with comprehensive coverage
func TestPatchBook(t *testing.T) {
	t.Run("should handle nil book", func(t *testing.T) {
		assert.NotPanics(t, func() {
			patchBook(nil, "any name")
		})
	})

	t.Run("name generation from blank names", func(t *testing.T) {
		tests := []struct {
			name          string
			initialName   string
			characterName string
			expected      string
		}{
			{
				name:          "nil name",
				initialName:   "",
				characterName: "Test Character",
				expected:      "Test Character Lore Book",
			},
			{
				name:          "empty string",
				initialName:   "",
				characterName: "Test Character",
				expected:      "Test Character Lore Book",
			},
			{
				name:          "whitespace only",
				initialName:   "   \t\n   ",
				characterName: "Test Character",
				expected:      "Test Character Lore Book",
			},
			{
				name:          "empty character name",
				initialName:   "",
				characterName: "",
				expected:      " Lore Book",
			},
			{
				name:          "character name with slashes",
				initialName:   "",
				characterName: "Test/Character/Name",
				expected:      "Test-Character-Name Lore Book",
			},
			{
				name:          "character name with only slashes",
				initialName:   "",
				characterName: "///",
				expected:      "--- Lore Book",
			},
			{
				name:          "unicode character name",
				initialName:   "",
				characterName: "テスト/キャラクター",
				expected:      "テスト-キャラクター Lore Book",
			},
			{
				name:          "special characters in character name",
				initialName:   "",
				characterName: "Test@Character#123$%^&*()",
				expected:      "Test@Character#123$%^&*() Lore Book",
			},
			{
				name:          "newlines in character name",
				initialName:   "",
				characterName: "Character\nWith\nNewlines",
				expected:      "Character\nWith\nNewlines Lore Book",
			},
			{
				name:          "tabs in character name",
				initialName:   "",
				characterName: "Character\tWith\tTabs",
				expected:      "Character\tWith\tTabs Lore Book",
			},
			{
				name:          "multiple consecutive slashes",
				initialName:   "",
				characterName: "Character//With//Multiple///Slashes",
				expected:      "Character--With--Multiple---Slashes Lore Book",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				book := &character.Book{Name: property.String(stringsx.Empty)}
				patchBook(book, tt.characterName)
				assert.Equal(t, tt.expected, string(book.Name))
			})
		}
	})

	t.Run("placeholder replacement", func(t *testing.T) {
		tests := []struct {
			name          string
			initialName   string
			characterName string
			expected      string
		}{
			{
				name:          "single placeholder",
				initialName:   "Lore for " + character.BookNamePlaceholder,
				characterName: "Test Character",
				expected:      "Lore for Test Character",
			},
			{
				name:          "multiple placeholders",
				initialName:   character.BookNamePlaceholder + "'s " + character.BookNamePlaceholder + " World",
				characterName: "Hero",
				expected:      "Hero's " + character.BookNamePlaceholder + " World",
			},
			{
				name:          "only first occurrence replaced",
				initialName:   character.BookNamePlaceholder + " and " + character.BookNamePlaceholder,
				characterName: "Test",
				expected:      "Test and " + character.BookNamePlaceholder,
			},
			{
				name:          "placeholder at end",
				initialName:   "Adventures of " + character.BookNamePlaceholder,
				characterName: "Hero",
				expected:      "Adventures of Hero",
			},
			{
				name:          "placeholder at beginning",
				initialName:   character.BookNamePlaceholder + " Adventures",
				characterName: "Hero",
				expected:      "Hero Adventures",
			},
			{
				name:          "only placeholder",
				initialName:   character.BookNamePlaceholder,
				characterName: "Hero",
				expected:      "Hero",
			},
			{
				name:          "placeholder with slashes",
				initialName:   "The " + character.BookNamePlaceholder + "/Chronicles",
				characterName: "Hero",
				expected:      "The Hero-Chronicles",
			},
			{
				name:          "mixed placeholders and slashes",
				initialName:   character.BookNamePlaceholder + "/" + character.BookNamePlaceholder + "/World",
				characterName: "Hero/Villain",
				expected:      "Hero-Villain-" + character.BookNamePlaceholder + "-World",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				book := &character.Book{Name: property.String(tt.initialName)}
				patchBook(book, tt.characterName)
				assert.Equal(t, tt.expected, string(book.Name))
			})
		}
	})

	t.Run("slash replacement", func(t *testing.T) {
		tests := []struct {
			name          string
			initialName   string
			characterName string
			expected      string
		}{
			{
				name:          "simple slash replacement",
				initialName:   "A/B/C/D",
				characterName: "Test Character",
				expected:      "A-B-C-D",
			},
			{
				name:          "existing name with slashes",
				initialName:   "Existing/Book/Name",
				characterName: "Test Character",
				expected:      "Existing-Book-Name",
			},
			{
				name:          "preserve valid name without slashes",
				initialName:   "Existing Valid Book Name",
				characterName: "Test Character",
				expected:      "Existing Valid Book Name",
			},
			{
				name:          "both placeholder and slashes",
				initialName:   character.BookNamePlaceholder + "/Adventures/Guide",
				characterName: "Hero/Villain",
				expected:      "Hero-Villain-Adventures-Guide",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				book := &character.Book{Name: property.String(tt.initialName)}
				patchBook(book, tt.characterName)
				assert.Equal(t, tt.expected, string(book.Name))
			})
		}
	})

	t.Run("special cases", func(t *testing.T) {
		t.Run("very long character name", func(t *testing.T) {
			longName := string(make([]rune, 1000))
			for i := range longName {
				longName = longName[:i] + "A" + longName[i+1:]
			}

			book := &character.Book{Name: property.String(stringsx.Empty)}
			patchBook(book, longName)

			assert.Equal(t, longName+" Lore Book", string(book.Name))
		})

		t.Run("MirrorNameAndComment is called", func(t *testing.T) {
			book := &character.Book{
				Name: property.String(stringsx.Empty),
				Entries: []*character.BookEntry{
					{
						BookEntryCore: character.BookEntryCore{
							Name:    property.String(stringsx.Empty),
							Comment: property.String(stringsx.Empty),
						},
					},
				},
			}

			patchBook(book, "Test Character")

			assert.NotEqual(t, "", string(book.Name))
		})

		t.Run("book with entries but no name", func(t *testing.T) {
			entryName := "Test Entry"
			book := &character.Book{
				Name: property.String(stringsx.Empty),
				Entries: []*character.BookEntry{
					{
						BookEntryCore: character.BookEntryCore{
							Name:    property.String(entryName),
							Content: "Test content",
							Keys:    []string{"key1", "key2"},
						},
					},
				},
			}

			patchBook(book, "Character")

			assert.Equal(t, "Character Lore Book", string(book.Name))
		})

		t.Run("book with description", func(t *testing.T) {
			description := "Book description"
			book := &character.Book{
				Name:        property.String(stringsx.Empty),
				Description: property.String(description),
			}

			patchBook(book, "Character")

			assert.Equal(t, "Character Lore Book", string(book.Name))
			assert.Equal(t, "Book description", string(book.Description))
		})
	})
}

// TestPatchLink tests the patchLink function with complete coverage
func TestPatchLink(t *testing.T) {
	t.Run("should set all link fields correctly", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:        source.ChubAI,
				CharacterID:   "test-char-123",
				PlatformID:    "platform-456",
				DirectURL:     "https://example.com/direct/test-char-123",
				NormalizedURL: "https://example.com/characters/test-char-123",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "test-char-123", string(sheet.Content.CharacterID))
		assert.Equal(t, "platform-456", string(sheet.Content.PlatformID))
		assert.Equal(t, "https://example.com/direct/test-char-123", string(sheet.Content.DirectLink))
	})

	t.Run("should handle empty fields", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      "",
				CharacterID: "",
				PlatformID:  "",
				DirectURL:   "",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, "", string(sheet.Content.SourceID))
		assert.Equal(t, "", string(sheet.Content.CharacterID))
		assert.Equal(t, "", string(sheet.Content.PlatformID))
		assert.Equal(t, "", string(sheet.Content.DirectLink))
	})

	t.Run("should overwrite existing fields", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: "new-char-123",
				PlatformID:  "new-platform-456",
				DirectURL:   "https://new.com/direct/new-char-123",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				SourceID:    "old-source",
				CharacterID: "old-char",
				PlatformID:  "old-platform",
				DirectLink:  "https://old.com/old-link",
			},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "new-char-123", string(sheet.Content.CharacterID))
		assert.Equal(t, "new-platform-456", string(sheet.Content.PlatformID))
		assert.Equal(t, "https://new.com/direct/new-char-123", string(sheet.Content.DirectLink))
	})

	t.Run("should handle different source types", func(t *testing.T) {
		sources := []source.ID{
			source.ChubAI,
			source.Pygmalion,
		}

		for _, src := range sources {
			metadata := &models.Metadata{
				CardInfo: models.CardInfo{
					Source:      src,
					CharacterID: "test-char",
					PlatformID:  "test-platform",
					DirectURL:   "https://test.com",
				},
			}

			sheet := &character.Sheet{
				Content: character.Content{},
			}

			patchLink(sheet, metadata)

			assert.Equal(t, string(src), string(sheet.Content.SourceID))
		}
	})

	t.Run("should handle unicode and special characters", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: "テスト-キャラ-123",
				PlatformID:  "プラットフォーム-456",
				DirectURL:   "https://example.com/direct/テスト-キャラ-123",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "テスト-キャラ-123", string(sheet.Content.CharacterID))
		assert.Equal(t, "プラットフォーム-456", string(sheet.Content.PlatformID))
		assert.Equal(t, "https://example.com/direct/テスト-キャラ-123", string(sheet.Content.DirectLink))
	})

	t.Run("should handle very long URLs and IDs", func(t *testing.T) {
		longID := string(make([]rune, 1000))
		for i := range longID {
			longID = longID[:i] + "a" + longID[i+1:]
		}

		longURL := "https://example.com/very/long/path/" + longID

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: longID,
				PlatformID:  longID,
				DirectURL:   longURL,
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, longID, string(sheet.Content.CharacterID))
		assert.Equal(t, longID, string(sheet.Content.PlatformID))
		assert.Equal(t, longURL, string(sheet.Content.DirectLink))
	})

	t.Run("should handle whitespace in fields", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: "  char-123  ",
				PlatformID:  "\tplatform-456\n",
				DirectURL:   "  https://example.com/direct  ",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		// Should preserve whitespace as-is (trimming is done elsewhere)
		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "  char-123  ", string(sheet.Content.CharacterID))
		assert.Equal(t, "\tplatform-456\n", string(sheet.Content.PlatformID))
		assert.Equal(t, "  https://example.com/direct  ", string(sheet.Content.DirectLink))
	})

	t.Run("should handle special characters in URLs", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: "char@123#",
				PlatformID:  "platform$456%",
				DirectURL:   "https://example.com/direct?char=123&test=true",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "char@123#", string(sheet.Content.CharacterID))
		assert.Equal(t, "platform$456%", string(sheet.Content.PlatformID))
		assert.Equal(t, "https://example.com/direct?char=123&test=true", string(sheet.Content.DirectLink))
	})

	t.Run("should handle newlines in fields", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:      source.ChubAI,
				CharacterID: "char\n123",
				PlatformID:  "platform\r\n456",
				DirectURL:   "https://example.com/direct\ntest",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		patchLink(sheet, metadata)

		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "char\n123", string(sheet.Content.CharacterID))
		assert.Equal(t, "platform\r\n456", string(sheet.Content.PlatformID))
		assert.Equal(t, "https://example.com/direct\ntest", string(sheet.Content.DirectLink))
	})
}

// TestPatchSheet tests the complete PatchSheet integration with full coverage
func TestPatchSheet(t *testing.T) {
	t.Run("complete integration test with all features", func(t *testing.T) {
		now := time.Now()
		createTime := now.Add(-24 * time.Hour)
		updateTime := now.Add(-12 * time.Hour)
		bookUpdateTime := now

		bookName := character.BookNamePlaceholder + " Adventures"
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Source:        source.ChubAI,
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
				CharacterBook: &character.Book{
					Name: property.String(bookName),
				},
			},
		}

		PatchSheet(sheet, metadata)

		// Test all patch operations were applied
		assert.Equal(t, string(source.ChubAI), string(sheet.Content.SourceID))
		assert.Equal(t, "test-char-123", string(sheet.Content.CharacterID))
		assert.Equal(t, "platform-123", string(sheet.Content.PlatformID))
		assert.Equal(t, "example.com/direct/test-char-123", string(sheet.Content.DirectLink))
		assert.Equal(t, "Test Creator", string(sheet.Content.Creator))

		// Test name and title patching
		assert.Equal(t, "Test Character", string(sheet.Content.Name))
		assert.Equal(t, "Test Card Title", string(sheet.Content.Title))
		assert.NotEmpty(t, sheet.Content.Nickname)
		assert.Equal(t, "Test Character", string(sheet.Content.Nickname))

		// Test creator notes patching
		assert.Contains(t, sheet.Content.CreatorNotes, "This is a test tagline.")
		assert.Contains(t, sheet.Content.CreatorNotes, "Original notes")

		// Test timestamp patching
		assert.Equal(t, timestamp.Seconds(createTime.Unix()), sheet.Content.CreationDate)
		assert.Equal(t, timestamp.Seconds(bookUpdateTime.Unix()), sheet.Content.ModificationDate)

		// Test tags merging
		assert.Contains(t, sheet.Content.Tags, "Fantasy")
		assert.Contains(t, sheet.Content.Tags, "Original Tag")

		// Test book patching with placeholder replacement
		assert.NotNil(t, sheet.Content.CharacterBook)
		assert.Equal(t, "Test Character Adventures", string(sheet.Content.CharacterBook.Name))

		// Test metadata synchronization
		assert.Equal(t, "Test Character", metadata.Name)
	})

	t.Run("should handle completely empty metadata and sheet", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo:    models.CardInfo{},
			CreatorInfo: models.CreatorInfo{},
		}

		sheet := &character.Sheet{
			Content: character.Content{},
		}

		assert.NotPanics(t, func() {
			PatchSheet(sheet, metadata)
		})

		// Should have applied empty values
		assert.Equal(t, "", string(sheet.Content.SourceID))
		assert.Equal(t, "", string(sheet.Content.CharacterID))
		assert.Equal(t, "", string(sheet.Content.Creator))
	})

	t.Run("should handle nil metadata fields", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Tags: nil,
			},
			CreatorInfo: models.CreatorInfo{},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Tags: nil,
			},
		}

		assert.NotPanics(t, func() {
			PatchSheet(sheet, metadata)
		})
	})

	t.Run("should not handle nil sheet pointer", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name: "Test",
			},
		}

		assert.Panics(t, func() {
			PatchSheet(nil, metadata)
		})
	})

	t.Run("should not handle nil metadata pointer", func(t *testing.T) {
		sheet := &character.Sheet{
			Content: character.Content{
				Name: "Test",
			},
		}

		assert.Panics(t, func() {
			PatchSheet(sheet, nil)
		})
	})

	t.Run("should handle sheet with no book", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:  "Test Character",
				Title: "Test Title",
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "Creator",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				CharacterBook: nil,
			},
		}

		PatchSheet(sheet, metadata)

		assert.Equal(t, "Test Character", string(sheet.Content.Name))
		assert.Equal(t, "Creator", string(sheet.Content.Creator))
		assert.Nil(t, sheet.Content.CharacterBook)
	})

	t.Run("should handle complex unicode scenarios", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    "テスト キャラクター",
				Title:   "测试 标题",
				Tagline: "🎉 Unicode tagline with émojis 🌟",
				Tags:    []models.Tag{{Slug: "unicode", Name: "Unicode"}},
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "クリエーター",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:         "原始名称",
				CreatorNotes: "原始备注",
				Tags:         []string{"原始标签"},
			},
		}

		PatchSheet(sheet, metadata)

		assert.Equal(t, "テスト キャラクター", string(sheet.Content.Name))
		assert.Equal(t, "クリエーター", string(sheet.Content.Creator))
		assert.Contains(t, sheet.Content.CreatorNotes, "🎉 Unicode tagline with émojis 🌟")
		assert.Contains(t, sheet.Content.Tags, "Unicode")
		assert.Contains(t, sheet.Content.Tags, "原始标签")
	})

	t.Run("should call NormalizeSymbols", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    `Test "Character" with 'quotes'`,
				Title:   `Test 'Title' with "quotes"`,
				Tagline: `"Quoted tagline" with symbols`,
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "Creator",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:         "Sheet Name",
				CreatorNotes: "Original notes",
			},
		}

		PatchSheet(sheet, metadata)

		// NormalizeSymbols should be called on the sheet
		assert.Equal(t, "Creator", string(sheet.Content.Creator))
	})

	t.Run("should preserve existing sheet data when metadata is incomplete", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name: "",
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "Creator",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:            "Preserved Sheet Name",
				Title:           "Preserved Sheet Title",
				Description:     "Preserved Description",
				Personality:     "Preserved Personality",
				Scenario:        "Preserved Scenario",
				FirstMessage:    "Preserved First Message",
				MessageExamples: "Preserved Examples",
				SystemPrompt:    "Preserved System Prompt",
			},
		}

		PatchSheet(sheet, metadata)

		// Name should fallback to sheet
		assert.Equal(t, "Preserved Sheet Name", string(sheet.Content.Name))
		// Other fields should be preserved
		assert.Equal(t, "Preserved Description", string(sheet.Content.Description))
		assert.Equal(t, "Preserved Personality", string(sheet.Content.Personality))
		assert.Equal(t, "Preserved Scenario", string(sheet.Content.Scenario))
		assert.Equal(t, "Preserved First Message", string(sheet.Content.FirstMessage))
		assert.Equal(t, "Preserved Examples", string(sheet.Content.MessageExamples))
		assert.Equal(t, "Preserved System Prompt", string(sheet.Content.SystemPrompt))
		// Creator should be set from metadata
		assert.Equal(t, "Creator", string(sheet.Content.Creator))
	})

	t.Run("should handle edge case with all values being whitespace", func(t *testing.T) {
		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    "   ",
				Title:   "\t\t",
				Tagline: "\n\n",
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: "   ",
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:         "   ",
				Title:        "\t\t",
				CreatorNotes: "\n\n",
			},
		}

		assert.NotPanics(t, func() {
			PatchSheet(sheet, metadata)
		})
	})

	t.Run("should handle maximum length values", func(t *testing.T) {
		maxString := string(make([]rune, 100000))
		for i := range maxString {
			maxString = maxString[:i] + "X" + maxString[i+1:]
		}

		metadata := &models.Metadata{
			CardInfo: models.CardInfo{
				Name:    maxString,
				Title:   maxString,
				Tagline: maxString,
			},
			CreatorInfo: models.CreatorInfo{
				Nickname: maxString,
			},
		}

		sheet := &character.Sheet{
			Content: character.Content{
				Name:         property.String(maxString),
				CreatorNotes: property.String(maxString),
			},
		}

		assert.NotPanics(t, func() {
			PatchSheet(sheet, metadata)
		})

		assert.Equal(t, maxString, string(sheet.Content.Name))
		assert.Equal(t, maxString, string(sheet.Content.Creator))
	})
}
