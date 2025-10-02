package models

import (
	"testing"
	"time"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/stretchr/testify/assert"
)

func createConsistentPair() (*Metadata, *character.Sheet) {
	now := time.Now()
	createTime := now.Add(-2 * time.Hour)
	updateTime := now.Add(-1 * time.Hour)
	bookUpdateTime := now

	metadata := &Metadata{
		Source: source.ID("test-source"),
		CardInfo: CardInfo{
			NormalizedURL: "https://example.com/card",
			DirectURL:     "https://example.com/direct",
			PlatformID:    "platform-123",
			CharacterID:   "char-456",
			Name:          "Test Character",
			Title:         "Test Card",
			Tagline:       "This is a test tagline.",
			CreateTime:    timestamp.Nano(createTime.UnixNano()),
			UpdateTime:    timestamp.Nano(updateTime.UnixNano()),
			Tags: []Tag{
				{Slug: "fantasy", Name: "Fantasy"},
				{Slug: "adventure", Name: "Adventure"},
			},
		},
		CreatorInfo: CreatorInfo{
			Nickname:   "Test CreatorInfo",
			Username:   "testcreator",
			PlatformID: "creator-123",
		},
		BookUpdateTime: timestamp.Nano(bookUpdateTime.UnixNano()),
	}

	sheet := &character.Sheet{
		Content: character.Content{
			SourceID:         "test-source",
			CharacterID:      "char-456",
			PlatformID:       "platform-123",
			DirectLink:       "https://example.com/direct",
			Title:            "Test Card",
			Name:             "Test Character",
			FirstMessage:     "This is a first message.\nAnd some more notes.",
			Description:      "This is a description.\nAnd some more notes.",
			Nickname:         "Test Character",
			Creator:          "Test CreatorInfo",
			CreatorNotes:     "This is a test tagline.\nAnd some more notes.",
			CreationDate:     timestamp.Seconds(createTime.Unix()),
			ModificationDate: timestamp.Seconds(bookUpdateTime.Unix()), // Uses the latest of the update times
			Tags:             []string{"Fantasy", "Adventure"},
			CharacterBook:    character.DefaultBook(),
		},
	}
	return metadata, sheet
}

func TestMetadata_IsConsistentWith(t *testing.T) {

	t.Run("should return true for consistent data", func(t *testing.T) {
		metadata, card := createConsistentPair()
		assert.True(t, metadata.IsConsistentWith(card), "Expected metadata to be consistent with the card")
	})

	t.Run("should correctly use the latest update time", func(t *testing.T) {
		metadata, card := createConsistentPair()
		// Manually set UpdateTime to be the most recent
		latestTime := time.Now().Add(5 * time.Minute)
		metadata.UpdateTime = timestamp.Nano(latestTime.UnixNano())
		card.Content.ModificationDate = timestamp.Seconds(latestTime.Unix())

		assert.True(t, metadata.IsConsistentWith(card), "Expected consistency when UpdateTime is the latest")
	})

	t.Run("should handle nil inputs", func(t *testing.T) {
		metadata, _ := createConsistentPair()
		var nilMetadata *Metadata
		var nilCard *character.Sheet

		assert.False(t, metadata.IsConsistentWith(nilCard), "Non-nil metadata should be inconsistent with a nil card")
		assert.True(t, nilMetadata.IsConsistentWith(nilCard), "Nil metadata should be consistent with a nil card")
	})

	t.Run("should return false for inconsistent data", func(t *testing.T) {
		testCases := []struct {
			name    string
			mutator func(r *Metadata, c *character.Sheet) // Function to make the data inconsistent
		}{
			// Malformed metadata conditions
			{
				name:    "malformed metadata - empty source",
				mutator: func(r *Metadata, c *character.Sheet) { r.Source = "" },
			},
			{
				name:    "malformed metadata - empty normalized URL",
				mutator: func(r *Metadata, c *character.Sheet) { r.NormalizedURL = "" },
			},
			{
				name:    "malformed metadata - empty direct URL",
				mutator: func(r *Metadata, c *character.Sheet) { r.DirectURL = "" },
			},
			{
				name:    "malformed metadata - empty CardInfo platform ID",
				mutator: func(r *Metadata, c *character.Sheet) { r.CardInfo.PlatformID = "" },
			},
			{
				name:    "malformed metadata - empty character ID",
				mutator: func(r *Metadata, c *character.Sheet) { r.CharacterID = "" },
			},
			{
				name:    "malformed metadata - empty name",
				mutator: func(r *Metadata, c *character.Sheet) { r.Name = "" },
			},
			{
				name:    "malformed metadata - empty title",
				mutator: func(r *Metadata, c *character.Sheet) { r.Title = "" },
			},
			{
				name:    "malformed metadata - zero create time",
				mutator: func(r *Metadata, c *character.Sheet) { r.CreateTime = 0 },
			},
			{
				name:    "malformed metadata - zero update time",
				mutator: func(r *Metadata, c *character.Sheet) { r.UpdateTime = 0 },
			},
			{
				name:    "malformed metadata - empty nickname",
				mutator: func(r *Metadata, c *character.Sheet) { r.Nickname = "" },
			},
			{
				name:    "malformed metadata - empty username",
				mutator: func(r *Metadata, c *character.Sheet) { r.Username = "" },
			},
			{
				name:    "malformed metadata - empty CreatorInfo platform ID",
				mutator: func(r *Metadata, c *character.Sheet) { r.CreatorInfo.PlatformID = "" },
			},
			// Field mismatch conditions
			{
				name:    "mismatched source ID",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.SourceID = "different-source" },
			},
			{
				name:    "mismatched character ID",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.CharacterID = "char-different" },
			},
			{
				name:    "mismatched platform ID",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.PlatformID = "different-platform" },
			},
			{
				name:    "mismatched direct URL",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.DirectLink = "https://different.com/direct" },
			},
			{
				name:    "mismatched title",
				mutator: func(r *Metadata, c *character.Sheet) { r.Title = "A Different Title" },
			},
			{
				name:    "mismatched character name",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.Name = "Different Character" },
			},
			{
				name:    "mismatched creator nickname",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.Creator = "Different Creator" },
			},
			{
				name:    "tagline is not a prefix of creator notes",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.CreatorNotes = "Notes do not start with tagline" },
			},
			{
				name: "mismatched creation time",
				mutator: func(r *Metadata, c *character.Sheet) {
					// Set a different creation time (1 hour difference)
					newTime := time.Now().Add(-3 * time.Hour)
					r.CreateTime = timestamp.Nano(newTime.UnixNano())
				},
			},
			{
				name: "mismatched modification time",
				mutator: func(r *Metadata, c *character.Sheet) {
					// Set a different modification time
					newTime := time.Now().Add(-5 * time.Hour)
					c.Content.ModificationDate = timestamp.Seconds(newTime.Unix())
				},
			},
			{
				name: "character book exists but BookUpdateTime is zero",
				mutator: func(r *Metadata, c *character.Sheet) {
					c.CharacterBook = &character.Book{} // Add a character book
					r.BookUpdateTime = 0                // But set BookUpdateTime to zero
				},
			},
			{
				name:    "tags have different length",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.Tags = []string{"Fantasy"} },
			},
			{
				name: "tags have different content",
				mutator: func(r *Metadata, c *character.Sheet) {
					r.Tags = []Tag{{Slug: "sci-fi", Name: "Sci-Fi"}, {Slug: "adventure", Name: "Adventure"}}
				},
			},
			{
				name:    "tags have different order",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.Tags = []string{"Adventure", "Fantasy"} },
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange: Get a consistent pair and then make it inconsistent
				metadata, card := createConsistentPair()
				tc.mutator(metadata, card)

				// Act & Assert
				assert.False(t, metadata.IsConsistentWith(card), "Expected IsConsistentWith to return false for: %s", tc.name)
			})
		}
	})

	t.Run("should handle character book conditions correctly", func(t *testing.T) {
		t.Run("consistent when no character book and BookUpdateTime is zero", func(t *testing.T) {
			metadata, card := createConsistentPair()
			card.CharacterBook = nil
			metadata.BookUpdateTime = 0
			// Need to adjust modification time since LatestUpdateTime() will now return UpdateTime instead of BookUpdateTime
			card.Content.ModificationDate = timestamp.Convert[timestamp.Seconds](metadata.UpdateTime)

			assert.True(t, metadata.IsConsistentWith(card), "Should be consistent when no character book and BookUpdateTime is zero")
		})

		t.Run("consistent when no character book but BookUpdateTime is zero", func(t *testing.T) {
			metadata, card := createConsistentPair()
			card.CharacterBook = nil

			metadata.BookUpdateTime = 0
			card.ModificationDate = timestamp.Convert[timestamp.Seconds](metadata.UpdateTime)

			assert.True(t, metadata.IsConsistentWith(card), "Should be consistent - condition allows zero BookUpdateTime when no character book")
		})

		t.Run("inconsistent when no character book but BookUpdateTime is non-zero", func(t *testing.T) {
			metadata, card := createConsistentPair()
			card.CharacterBook = nil

			assert.False(t, metadata.IsConsistentWith(card), "Should be consistent - condition allows zero BookUpdateTime when no character book")
		})
	})
}

func TestMetadata_IsMalformed(t *testing.T) {
	t.Run("should return false for well-formed metadata", func(t *testing.T) {
		metadata := &Metadata{
			Source: source.ID("test-source"),
			CardInfo: CardInfo{
				NormalizedURL: "https://example.com/card",
				DirectURL:     "https://example.com/direct",
				PlatformID:    "platform-123",
				CharacterID:   "char-456",
				Name:          "Test Character",
				Title:         "Test Card",
				Tagline:       "This is a test tagline.",
				CreateTime:    1234567890,
				UpdateTime:    1234567890,
			},
			CreatorInfo: CreatorInfo{
				Nickname:   "Test Creator",
				Username:   "testcreator",
				PlatformID: "creator-123",
			},
		}

		assert.False(t, metadata.IsMalformed(), "Well-formed metadata should not be malformed")
	})

	t.Run("should return true for malformed metadata", func(t *testing.T) {
		testCases := []struct {
			name    string
			mutator func(*Metadata) // Function to make the metadata malformed
		}{
			{
				name:    "empty source",
				mutator: func(m *Metadata) { m.Source = "" },
			},
			{
				name:    "blank source with spaces",
				mutator: func(m *Metadata) { m.Source = "   " },
			},
			{
				name:    "empty normalized URL",
				mutator: func(m *Metadata) { m.NormalizedURL = "" },
			},
			{
				name:    "blank normalized URL with spaces",
				mutator: func(m *Metadata) { m.NormalizedURL = "   " },
			},
			{
				name:    "empty direct URL",
				mutator: func(m *Metadata) { m.DirectURL = "" },
			},
			{
				name:    "blank direct URL with spaces",
				mutator: func(m *Metadata) { m.DirectURL = "   " },
			},
			{
				name:    "empty CardInfo PlatformID",
				mutator: func(m *Metadata) { m.CardInfo.PlatformID = "" },
			},
			{
				name:    "blank CardInfo PlatformID with spaces",
				mutator: func(m *Metadata) { m.CardInfo.PlatformID = "   " },
			},
			{
				name:    "empty CharacterID",
				mutator: func(m *Metadata) { m.CharacterID = "" },
			},
			{
				name:    "blank CharacterID with spaces",
				mutator: func(m *Metadata) { m.CharacterID = "   " },
			},
			{
				name:    "empty Name",
				mutator: func(m *Metadata) { m.Name = "" },
			},
			{
				name:    "blank Name with spaces",
				mutator: func(m *Metadata) { m.Name = "   " },
			},
			{
				name:    "empty Title",
				mutator: func(m *Metadata) { m.Title = "" },
			},
			{
				name:    "blank Title with spaces",
				mutator: func(m *Metadata) { m.Title = "   " },
			},
			{
				name:    "zero CreateTime",
				mutator: func(m *Metadata) { m.CreateTime = 0 },
			},
			{
				name:    "zero UpdateTime",
				mutator: func(m *Metadata) { m.UpdateTime = 0 },
			},
			{
				name:    "empty Nickname",
				mutator: func(m *Metadata) { m.Nickname = "" },
			},
			{
				name:    "blank Nickname with spaces",
				mutator: func(m *Metadata) { m.Nickname = "   " },
			},
			{
				name:    "empty Username",
				mutator: func(m *Metadata) { m.Username = "" },
			},
			{
				name:    "blank Username with spaces",
				mutator: func(m *Metadata) { m.Username = "   " },
			},
			{
				name:    "empty CreatorInfo PlatformID",
				mutator: func(m *Metadata) { m.CreatorInfo.PlatformID = "" },
			},
			{
				name:    "blank CreatorInfo PlatformID with spaces",
				mutator: func(m *Metadata) { m.CreatorInfo.PlatformID = "   " },
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create well-formed metadata
				metadata := &Metadata{
					Source: source.ID("test-source"),
					CardInfo: CardInfo{
						NormalizedURL: "https://example.com/card",
						DirectURL:     "https://example.com/direct",
						PlatformID:    "platform-123",
						CharacterID:   "char-456",
						Name:          "Test Character",
						Title:         "Test Card",
						Tagline:       "This is a test tagline.",
						CreateTime:    1234567890,
						UpdateTime:    1234567890,
					},
					CreatorInfo: CreatorInfo{
						Nickname:   "Test Creator",
						Username:   "testcreator",
						PlatformID: "creator-123",
					},
				}

				// Apply mutation to make it malformed
				tc.mutator(metadata)

				assert.True(t, metadata.IsMalformed(), "Expected metadata to be malformed for case: %s", tc.name)
			})
		}
	})
}

func TestMetadata_LatestUpdateTime(t *testing.T) {
	t.Run("should return UpdateTime when it is later than BookUpdateTime", func(t *testing.T) {
		now := time.Now()
		updateTime := now
		bookUpdateTime := now.Add(-1 * time.Hour) // BookUpdateTime is earlier

		metadata := &Metadata{
			CardInfo: CardInfo{
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(bookUpdateTime.UnixNano()),
		}

		expected := timestamp.Nano(updateTime.UnixNano())
		assert.Equal(t, expected, metadata.LatestUpdateTime(), "Should return UpdateTime when it is the latest")
	})

	t.Run("should return BookUpdateTime when it is later than UpdateTime", func(t *testing.T) {
		now := time.Now()
		updateTime := now.Add(-1 * time.Hour) // UpdateTime is earlier
		bookUpdateTime := now

		metadata := &Metadata{
			CardInfo: CardInfo{
				UpdateTime: timestamp.Nano(updateTime.UnixNano()),
			},
			BookUpdateTime: timestamp.Nano(bookUpdateTime.UnixNano()),
		}

		expected := timestamp.Nano(bookUpdateTime.UnixNano())
		assert.Equal(t, expected, metadata.LatestUpdateTime(), "Should return BookUpdateTime when it is the latest")
	})

	t.Run("should return same time when UpdateTime and BookUpdateTime are equal", func(t *testing.T) {
		now := time.Now()
		sameTime := timestamp.Nano(now.UnixNano())

		metadata := &Metadata{
			CardInfo: CardInfo{
				UpdateTime: sameTime,
			},
			BookUpdateTime: sameTime,
		}

		assert.Equal(t, sameTime, metadata.LatestUpdateTime(), "Should return the same time when both are equal")
	})

	t.Run("should handle zero values correctly", func(t *testing.T) {
		now := time.Now()
		updateTime := timestamp.Nano(now.UnixNano())

		metadata := &Metadata{
			CardInfo: CardInfo{
				UpdateTime: updateTime,
			},
			BookUpdateTime: 0,
		}

		assert.Equal(t, updateTime, metadata.LatestUpdateTime(), "Should return UpdateTime when BookUpdateTime is zero")
	})
}

func TestMetadata_Clone(t *testing.T) {
	original := &Metadata{
		Source: source.ID("test-source"),
		CardInfo: CardInfo{
			NormalizedURL: "https://example.com/card",
			DirectURL:     "https://direct.example.com/card",
			PlatformID:    "platform-123",
			CharacterID:   "char-456",
			Title:         "Test Card",
			Name:          "Test Character",
			Tagline:       "This is a test tagline.",
			CreateTime:    timestamp.Nano(time.Now().Add(-24 * time.Hour).UnixNano()),
			UpdateTime:    timestamp.Nano(time.Now().Add(-12 * time.Hour).UnixNano()),
			Tags: []Tag{
				{Slug: "tag-1", Name: "Tag One"},
				{Slug: "tag-2", Name: "Tag Two"},
			},
		},
		CreatorInfo: CreatorInfo{
			Nickname:   "Test CreatorInfo",
			Username:   "testcreator",
			PlatformID: "creator-123",
		},
		BookUpdateTime: timestamp.Nano(time.Now().UnixNano()),
	}

	clone := original.Clone()

	t.Run("Clone is a new instance", func(t *testing.T) {
		assert.NotSame(t, original, clone, "Clone should be a new instance, not a pointer to the original")
	})

	t.Run("All fields are equal", func(t *testing.T) {
		assert.Equal(t, original, clone, "Cloned metadata should be equal to the original")
	})

	t.Run("Tags slice is a separate instance", func(t *testing.T) {
		if assert.NotEmpty(t, clone.Tags, "Tags should not be empty") {
			clone.Tags[0].Name = "Modified Tag"
			assert.NotEqual(t, original.Tags[0].Name, clone.Tags[0].Name, "Modifying the clone's tag should not affect the original's tag")
		}

		clone.Tags = append(clone.Tags, Tag{Slug: "tag-3", Name: "Tag Three"})
		assert.NotEqual(t, len(original.Tags), len(clone.Tags), "Appending to the clone's Tags slice should not affect the original slice")
	})
}
