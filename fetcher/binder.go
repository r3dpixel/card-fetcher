package fetcher

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/r3dpixel/toolkit/timestamp"
)

// Binder is a container for all the data fetched from a single source
type Binder struct {
	MetadataBinder
	BookBinder
}

// MetadataBinder is a container for all the metadata fetched from a single source
type MetadataBinder struct {
	CharacterID   string
	NormalizedURL string
	DirectURL     string
	Document      *goquery.Document
	JsonResponse
}

// BookBinder is a container for all the book data fetched from a single source
type BookBinder struct {
	Responses  []JsonResponse
	UpdateTime timestamp.Nano
}

// EmptyBinder singleton
var EmptyBinder = Binder{}

// EmptyMetadataBinder singleton
var EmptyMetadataBinder = MetadataBinder{}

// EmptyBookBinder singleton
var EmptyBookBinder = BookBinder{}

// JsonString is a generic JSON response
type JsonString string
