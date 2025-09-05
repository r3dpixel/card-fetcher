package models

import (
	"slices"
	"strings"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/timestamp"
)

type Metadata struct {
	Source         source.ID
	CardURL        string
	DirectURL      string
	PlatformID     string
	CharacterID    string
	CardName       string
	CharacterName  string
	Creator        string
	Tagline        string
	CreateTime     timestamp.Nano
	UpdateTime     timestamp.Nano
	BookUpdateTime timestamp.Nano
	Tags           []Tag
}

func (m *Metadata) LatestUpdateTime() timestamp.Nano {
	return max(m.UpdateTime, m.BookUpdateTime)
}

func (m *Metadata) IsConsistentWith(card *character.Sheet) bool {
	if card == nil {
		return m == nil
	}
	metadataTags := TagsToNames(m.Tags)

	return string(m.Source) == card.Data.SourceID &&
		m.CharacterID == card.Data.CharacterID &&
		m.PlatformID == card.Data.PlatformID &&
		m.DirectURL == card.Data.DirectLink &&
		m.CardName == card.Data.CardName &&
		m.CharacterName == card.Data.CharacterName &&
		m.Creator == card.Data.Creator &&
		strings.HasPrefix(card.Data.CreatorNotes, m.Tagline) &&
		timestamp.Convert[timestamp.Seconds](m.CreateTime) == card.Data.CreationDate &&
		timestamp.Convert[timestamp.Seconds](m.LatestUpdateTime()) == card.Data.ModificationDate &&
		slices.Equal(metadataTags, card.Data.Tags)
}

func (m *Metadata) Clone() *Metadata {
	return &Metadata{
		Source:         m.Source,
		CardURL:        m.CardURL,
		DirectURL:      m.DirectURL,
		PlatformID:     m.PlatformID,
		CharacterID:    m.CharacterID,
		CardName:       m.CardName,
		CharacterName:  m.CharacterName,
		Creator:        m.Creator,
		Tagline:        m.Tagline,
		CreateTime:     m.CreateTime,
		UpdateTime:     m.UpdateTime,
		BookUpdateTime: m.BookUpdateTime,
		Tags:           slices.Clone(m.Tags),
	}
}
