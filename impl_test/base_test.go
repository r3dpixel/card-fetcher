package fetcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSources(t *testing.T) {
	for _, fetcher := range testRouter.Fetchers() {
		assert.True(t, fetcher.IsSourceUp())
	}
}
