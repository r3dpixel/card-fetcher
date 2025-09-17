package fetcher

import (
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/tidwall/gjson"
)

type Binder struct {
	MetadataBinder
	BookBinder
}

type MetadataBinder struct {
	CharacterID   string
	NormalizedURL string
	gjson.Result
}

type BookBinder struct {
	Responses  []gjson.Result
	UpdateTime timestamp.Nano
}

var EmptyBinder = Binder{}
var EmptyMetadataBinder = MetadataBinder{}
var EmptyBookBinder = BookBinder{}
