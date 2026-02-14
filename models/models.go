package models

import (
	"slices"
	"strings"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
)

// anonymousIdentifier is the lowercased version of the anonymous creator name
var anonymousIdentifier = strings.ToLower(character.AnonymousCreator)

// Metadata struct for storing card metadata
type Metadata struct {
	Source source.ID
	CardInfo
	CreatorInfo
	BookUpdateTime timestamp.Nano
	GreetingsCount int
	HasBook        bool
}

// LatestUpdateTime returns the latest update time of the card
func (m *Metadata) LatestUpdateTime() timestamp.Nano {
	return max(m.CardInfo.UpdateTime, m.BookUpdateTime)
}

// Integrity checks if the metadata is valid
func (m *Metadata) Integrity() bool {
	return stringsx.IsNotBlank(string(m.Source)) &&
		stringsx.IsNotBlank(m.NormalizedURL) &&
		stringsx.IsNotBlank(m.DirectURL) &&
		stringsx.IsNotBlank(m.CardInfo.PlatformID) &&
		stringsx.IsNotBlank(m.CharacterID) &&
		stringsx.IsNotBlank(m.Name) &&
		stringsx.IsNotBlank(m.Title) &&
		m.CreateTime > 0 &&
		m.UpdateTime > 0 &&
		m.UpdateTime >= m.CreateTime &&
		stringsx.IsNotBlank(m.Nickname) &&
		stringsx.IsNotBlank(m.Username) &&
		(strings.ToLower(m.Nickname) == anonymousIdentifier || stringsx.IsNotBlank(m.CreatorInfo.PlatformID))
}

// IsConsistentWith checks if the metadata is consistent with the card
func (m *Metadata) IsConsistentWith(card *character.Sheet) bool {
	if card == nil {
		return m == nil
	}
	metadataTags := TagsToNames(m.CardInfo.Tags)

	return m.Integrity() &&
		card.Integrity() &&
		string(m.Source) == string(card.SourceID) &&
		m.CharacterID == string(card.CharacterID) &&
		m.CardInfo.PlatformID == string(card.PlatformID) &&
		m.DirectURL == string(card.DirectLink) &&
		m.Title == string(card.Title) &&
		m.Name == string(card.Name) &&
		m.Nickname == string(card.Creator) &&
		strings.HasPrefix(string(card.CreatorNotes), m.Tagline) &&
		timestamp.ConvertToSeconds(m.CreateTime) == card.CreationDate &&
		timestamp.ConvertToSeconds(m.LatestUpdateTime()) == card.ModificationDate &&
		m.HasBook == (card.CharacterBook != nil) &&
		((!m.HasBook && m.BookUpdateTime == 0) || (m.HasBook && m.BookUpdateTime != 0)) &&
		m.GreetingsCount == len(card.AlternateGreetings) &&
		slices.Equal(metadataTags, card.Tags)
}

// Clone returns a deep copy of the Metadata struct
func (m *Metadata) Clone() *Metadata {
	// Clone tags
	tags := slices.Clone(m.Tags)

	// Clone metadata
	clone := *m

	// Set the cloned tags
	clone.Tags = tags

	// Return the cloned metadata
	return &clone
}

// CardInfo struct for storing card information
type CardInfo struct {
	NormalizedURL string
	DirectURL     string
	PlatformID    string
	CharacterID   string
	Name          string
	Title         string
	Tagline       string
	CreateTime    timestamp.Nano
	UpdateTime    timestamp.Nano
	IsForked      bool
	Tags          []Tag
}

// NormalizeSymbols trims whitespace from Name and Title and normalizes symbols in Tagline
func (c *CardInfo) NormalizeSymbols() {
	c.Name = strings.TrimSpace(c.Name)
	c.Title = strings.TrimSpace(c.Title)
	c.Tagline = stringsx.NormalizeSymbols(c.Tagline)
}

// CreatorInfo struct for storing creator information
type CreatorInfo struct {
	Nickname   string
	Username   string
	PlatformID string
}
