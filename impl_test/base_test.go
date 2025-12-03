package impl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSources(t *testing.T) {
	for _, fetcher := range afTestRouter.Fetchers() {
		assert.NoError(t, fetcher.IsSourceUp())
	}
}
