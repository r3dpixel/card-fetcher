package models

import (
	"slices"
	"strings"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
)

type Metadata struct {
	Source source.ID
	CardInfo
	CreatorInfo
	BookUpdateTime timestamp.Nano
}

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

func (c *CardInfo) NormalizeSymbols() {
	c.Name = strings.TrimSpace(c.Name)
	c.Title = strings.TrimSpace(c.Title)
	c.Tagline = stringsx.NormalizeSymbols(c.Tagline)
}

type CreatorInfo struct {
	Nickname   string
	Username   string
	PlatformID string
}

func (m *Metadata) LatestUpdateTime() timestamp.Nano {
	return max(m.CardInfo.UpdateTime, m.BookUpdateTime)
}

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
		stringsx.IsNotBlank(m.CreatorInfo.PlatformID)
}

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
		((card.CharacterBook == nil && m.BookUpdateTime == 0) || (card.CharacterBook != nil && m.BookUpdateTime != 0)) &&
		slices.Equal(metadataTags, card.Tags)
}

func (m *Metadata) Clone() *Metadata {
	tags := slices.Clone(m.Tags)

	clone := *m
	clone.Tags = tags

	return &clone
}
