package impl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSources(t *testing.T) {
	for _, fetcher := range testRouter.Fetchers() {
		assert.NoError(t, fetcher.IsSourceUp())
	}
}
