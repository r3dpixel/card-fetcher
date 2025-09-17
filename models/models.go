package models

import (
	"slices"
	"strings"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/timestamp"
)

type Metadata struct {
	CardInfo
	CreatorInfo
	BookUpdateTime timestamp.Nano
}

type CardInfo struct {
	Source        source.ID
	NormalizedURL string
	DirectURL     string
	PlatformID    string
	CharacterID   string
	Name          string
	Title         string
	Tagline       string
	CreateTime    timestamp.Nano
	UpdateTime    timestamp.Nano
	Tags          []Tag
}

type CreatorInfo struct {
	Nickname   string
	Username   string
	PlatformID string
}

func (m *Metadata) LatestUpdateTime() timestamp.Nano {
	return max(m.CardInfo.UpdateTime, m.BookUpdateTime)
}

func (m *Metadata) IsConsistentWith(card *character.Sheet) bool {
	if card == nil {
		return m == nil
	}
	metadataTags := TagsToNames(m.CardInfo.Tags)

	return string(m.Source) == card.Content.SourceID &&
		m.CharacterID == card.Content.CharacterID &&
		m.CardInfo.PlatformID == card.Content.PlatformID &&
		m.DirectURL == card.Content.DirectLink &&
		m.Title == card.Content.Title &&
		m.Name == card.Content.Name &&
		strings.HasPrefix(card.Content.CreatorNotes, m.Tagline) &&
		timestamp.Convert[timestamp.Seconds](m.CreateTime) == card.Content.CreationDate &&
		timestamp.Convert[timestamp.Seconds](m.LatestUpdateTime()) == card.Content.ModificationDate &&
		slices.Equal(metadataTags, card.Content.Tags)
}

func (m *Metadata) Clone() *Metadata {
	tags := slices.Clone(m.Tags)

	clone := *m
	clone.Tags = tags

	return &clone
}
