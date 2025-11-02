package main

import (
	"github.com/r3dpixel/card-fetcher/router"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/toolkit/jsonx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/rs/zerolog/log"
)

func main() {
	sonicx.Config = sonicx.StableSort

	r := router.EnvConfigured()
	resourceURLs := router.GetResourceURLs()

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

	err = card.Sheet.ToFile(router.GetResourceJsonPath(t.SourceID()), jsonx.Options{
		Pretty: true,
		Indent: "  ",
	})
	if err != nil {
		log.Err(err).Msg("Failed to save JSON " + t.NormalizedURL())
		return
	}
}
