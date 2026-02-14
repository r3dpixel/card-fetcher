package main

import (
	"path"

	"github.com/r3dpixel/card-fetcher/router"
	"github.com/r3dpixel/card-fetcher/snapshots"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/toolkit/jsonx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/rs/zerolog/log"
)

const snapshotDir = "snapshots"

// main is the entry point of the tool
func main() {
	// Use stable sort for snapshots
	sonicx.Config = sonicx.StableSort

	// Configure the router
	r := router.EnvConfigured(nil)
	// Get the resource map
	resourceMap := snapshots.GetResourceMap()

	// Save snapshots
	for sourceID, urls := range resourceMap {
		// Iterate over the URLs
		for i, url := range urls {
			// Get the task
			t, ok := r.TaskOf(url)
			if !ok {
				// Log the missing task
				log.Warn().Msg("Task not found: " + url)
				continue
			}
			// Save the snapshot
			saveSnapshot(t, sourceID, i)
		}
	}
}

// saveSnapshot saves the given task snapshot to the filesystem
func saveSnapshot(t task.Task, sourceID source.ID, index int) {
	// Fetch the card
	_, card, err := t.FetchAll()
	if err != nil {
		log.Err(err).Msg("Failed to fetch " + t.NormalizedURL())
		return
	}

	// Encode the card
	rawCard, err := card.Encode()
	if err != nil {
		log.Err(err).Msg("Failed to encode " + t.NormalizedURL())
		return
	}

	// Save the card to the filesystem
	err = rawCard.ToFile(path.Join(snapshotDir, snapshots.GetResourceCardPath(sourceID, index)))
	if err != nil {
		log.Err(err).Msg("Failed to save " + t.NormalizedURL())
		return
	}

	// Save the sheet to the filesystem
	err = card.Sheet.ToFile(path.Join(snapshotDir, snapshots.GetResourceJsonPath(sourceID, index)), jsonx.Options{
		Pretty: true,
		Indent: "  ",
	})
	if err != nil {
		log.Err(err).Msg("Failed to save JSON " + t.NormalizedURL())
		return
	}
}
