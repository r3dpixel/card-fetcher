package impl

import (
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/toolkit/cred"
)

// BuilderOptions options for builders
type BuilderOptions struct {
	PygmalionIdentityReader cred.IdentityReader
	ChromePath              func() string
}

// DefaultBuilders returns a list of default fetchers using the provided options
func DefaultBuilders(opts BuilderOptions) []fetcher.Builder {
	return []fetcher.Builder{
		CharacterTavernBuilder{},
		ChubAIBuilder{},
		NyaiMeBuilder{},
		PephopBuilder{},
		PygmalionBuilder{IdentityReader: opts.PygmalionIdentityReader},
		WyvernChatBuilder{},
		JannyAIBuilder{ChromePath: opts.ChromePath},
		AiccBuilder{},
	}
}
