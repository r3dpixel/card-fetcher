package impl

import (
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/toolkit/cred"
)

type BuilderOptions struct {
	PygmalionIdentityReader cred.IdentityReader
}

func DefaultBuilders(opts BuilderOptions) []fetcher.Builder {
	return []fetcher.Builder{
		NewCharacterTavernFetcher,
		NewChubAIFetcher,
		NewNyaiMeFetcher,
		NewPephopFetcher,
		fetcher.BuilderOf(opts.PygmalionIdentityReader, NewPygmalionFetcher),
		NewWyvernChatFetcher,
	}
}
