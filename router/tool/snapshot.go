package main

import (
	"maps"
	"slices"

	"github.com/r3dpixel/card-fetcher/factory"
	"github.com/r3dpixel/card-fetcher/router"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/rs/zerolog/log"
)

func main() {
	sonicx.Config = sonicx.StableSort

	factoryOpts := factory.Options{
		ClientOptions:             reqx.Options{},
		PygmalionIdentityProvider: cred.NewManager("pygmalion", cred.Env),
	}

	r := router.New(factoryOpts)
	resourceURLs := router.GetResourceURLs()

	sourceIDs := slices.Collect(maps.Keys(resourceURLs))
	r.RegisterFetchers(sourceIDs...)

	for _, url := range resourceURLs {
		t, ok := r.TaskOf(url)
		if !ok {
			log.Warn().Msg("Task not found: " + url)
			continue
		}
		saveSnapshot(t)
	}
}

func saveSnapshot(t task.Task) {
	_, card, err := t.FetchAll()
	if err != nil {
		log.Err(err).Msg("Failed to fetch " + t.NormalizedURL())
		return
	}

	rawCard, err := card.Encode()
	if err != nil {
		log.Err(err).Msg("Failed to encode " + t.NormalizedURL())
		return
	}

	err = rawCard.ToFile(router.GetResourceCardPath(t.SourceID()))
	if err != nil {
		log.Err(err).Msg("Failed to save " + t.NormalizedURL())
		return
	}

	err = card.Sheet.ToFile(router.GetResourceJsonPath(t.SourceID()), character.JsonPretty)
	if err != nil {
		log.Err(err).Msg("Failed to save JSON " + t.NormalizedURL())
		return
	}
}
