package fetcher

import (
	"github.com/r3dpixel/toolkit/timestamp"
)

type Binder struct {
	MetadataBinder
	BookBinder
}

type MetadataBinder struct {
	CharacterID   string
	NormalizedURL string
	DirectURL     string
	JsonResponse
}

type BookBinder struct {
	Responses  []JsonResponse
	UpdateTime timestamp.Nano
}

var EmptyBinder = Binder{}
var EmptyMetadataBinder = MetadataBinder{}
var EmptyBookBinder = BookBinder{}
