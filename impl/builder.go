package impl

import (
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/toolkit/cred"
)

type BuilderOptions struct {
	PygmalionIdentityReader cred.IdentityReader
	JannyAICookieProvider   func() JannyCookies
}

func DefaultBuilders(opts BuilderOptions) []fetcher.Builder {
	return []fetcher.Builder{
		CharacterTavernBuilder{},
		ChubAIBuilder{},
		NyaiMeBuilder{},
		PephopBuilder{},
		PygmalionBuilder{IdentityReader: opts.PygmalionIdentityReader},
		WyvernChatBuilder{},
		JannyAIBuilder{cookieProvider: opts.JannyAICookieProvider},
	}
}
