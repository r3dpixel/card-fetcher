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
		CardInfo: CardInfo{
			Source:        source.ID("test-source"),
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
			Creator:          "Test CreatorInfo",
			CreatorNotes:     "This is a test tagline.\nAnd some more notes.",
			CreationDate:     timestamp.Seconds(createTime.Unix()),
			ModificationDate: timestamp.Seconds(bookUpdateTime.Unix()), // Uses the latest of the update times
			Tags:             []string{"Fantasy", "Adventure"},
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
			{
				name:    "mismatched character ID",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.CharacterID = "char-different" },
			},
			{
				name:    "mismatched card name",
				mutator: func(r *Metadata, c *character.Sheet) { r.Title = "A Different Name" },
			},
			{
				name:    "tagline is not a prefix of creator notes",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.CreatorNotes = "Notes do not start with tagline" },
			},
			{
				name:    "mismatched creation time",
				mutator: func(r *Metadata, c *character.Sheet) { r.CreateTime = 0 },
			},
			{
				name:    "mismatched modification time",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.ModificationDate = 0 },
			},
			{
				name:    "tags have different length",
				mutator: func(r *Metadata, c *character.Sheet) { c.Content.Tags = []string{"Fantasy"} },
			},
			{
				name:    "tags have different content",
				mutator: func(r *Metadata, c *character.Sheet) { r.Tags = []Tag{{Name: "Sci-Fi"}, {Name: "Adventure"}} },
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
}

func TestMetadata_Clone(t *testing.T) {
	original := &Metadata{
		CardInfo: CardInfo{
			Source:        source.ID("test-source"),
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
