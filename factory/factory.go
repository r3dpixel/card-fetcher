package factory

import (
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/postprocessor"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/cred"
)

type FactoryOptions struct {
	Client                    *req.Client
	PygmalionIdentityProvider cred.IdentityReader
}

type Factory interface {
	FetcherOf(sourceID source.ID) fetcher.Fetcher
}
type factory struct {
	client                    *req.Client
	pygmalionIdentityProvider cred.IdentityReader
}

func New(opts FactoryOptions) Factory {
	return &factory{
		client:                    opts.Client,
		pygmalionIdentityProvider: opts.PygmalionIdentityProvider,
	}
}

func (f *factory) implementationOf(sourceID source.ID) fetcher.Fetcher {
	switch sourceID {
	case source.CharacterTavern:
		return fetcher.NewCharacterTavernFetcher(f.client)
	case source.ChubAI:
		return fetcher.NewChubAIFetcher(f.client)
	case source.NyaiMe:
		return fetcher.NewNyaiMeFetcher(f.client)
	case source.PepHop:
		return fetcher.NewPephopFetcher(f.client)
	case source.Pygmalion:
		return fetcher.NewPygmalionFetcher(f.client, f.pygmalionIdentityProvider)
	case source.WyvernChat:
		return fetcher.NewWyvernChatFetcher(f.client)
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
