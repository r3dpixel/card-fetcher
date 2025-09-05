package factory

import (
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/postprocessor"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/cred"
)

type FactoryOptions struct {
	PygmalionIdentityProvider cred.IdentityReader
}

type Factory interface {
	FetcherOf(sourceID source.ID) fetcher.Fetcher
}
type factory struct {
	pygmalionIdentityProvider cred.IdentityReader
}

func New(opts FactoryOptions) Factory {
	return &factory{
		pygmalionIdentityProvider: opts.PygmalionIdentityProvider,
	}
}

func (f *factory) implementationOf(sourceID source.ID) fetcher.Fetcher {
	switch sourceID {
	case source.CharacterTavern:
		return fetcher.NewCharacterTavernFetcher()
	case source.ChubAI:
		return fetcher.NewChubAIFetcher()
	case source.NyaiMe:
		return fetcher.NewNyaiMeFetcher()
	case source.PepHop:
		return fetcher.NewPephopFetcher()
	case source.Pygmalion:
		return fetcher.NewPygmalionFetcher(f.pygmalionIdentityProvider)
	case source.WyvernChat:
		return fetcher.NewWyvernChatFetcher()
	default:
		return nil
	}
}

func (f *factory) FetcherOf(sourceID source.ID) fetcher.Fetcher {
	simpleFetcher := f.implementationOf(sourceID)
	if simpleFetcher == nil {
		return nil
	}
	return postprocessor.New(simpleFetcher)
}
