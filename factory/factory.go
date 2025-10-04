package factory

import (
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/impl"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/reqx"
)

type Options struct {
	ClientOptions             reqx.Options
	PygmalionIdentityProvider cred.IdentityReader
}

type Factory interface {
	FetcherOf(sourceID source.ID) fetcher.Fetcher
}
type factory struct {
	client                    *reqx.Client
	pygmalionIdentityProvider cred.IdentityReader
}

func New(opts Options) Factory {
	return &factory{
		client:                    reqx.NewClient(opts.ClientOptions),
		pygmalionIdentityProvider: opts.PygmalionIdentityProvider,
	}
}

func (f *factory) FetcherOf(sourceID source.ID) fetcher.Fetcher {
	switch sourceID {
	case source.CharacterTavern:
		return impl.NewCharacterTavernFetcher(f.client)
	case source.ChubAI:
		return impl.NewChubAIFetcher(f.client)
	case source.NyaiMe:
		return impl.NewNyaiMeFetcher(f.client)
	case source.PepHop:
		return impl.NewPephopFetcher(f.client)
	case source.Pygmalion:
		return impl.NewPygmalionFetcher(f.client, f.pygmalionIdentityProvider)
	case source.WyvernChat:
		return impl.WyvernChatHandler(f.client)
	default:
		return nil
	}
}
