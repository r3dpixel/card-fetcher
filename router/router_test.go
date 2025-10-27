package router

import (
	"testing"
	"time"

	"github.com/r3dpixel/card-fetcher/impl"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name   string
		opts   reqx.Options
		assert func(t *testing.T, r *Router)
	}{
		{
			name: "Default options",
			opts: reqx.Options{},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
		{
			name: "With Retries",
			opts: reqx.Options{RetryCount: 5, MinBackoff: 1 * time.Second, MaxBackoff: 5 * time.Second},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
		{
			name: "With HTTP3",
			opts: reqx.Options{EnableHttp3: true},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
		{
			name: "With Chrome Impersonation",
			opts: reqx.Options{Impersonation: reqx.Chrome},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
		{
			name: "With Firefox Impersonation",
			opts: reqx.Options{Impersonation: reqx.Firefox},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
		{
			name: "With Safari Impersonation",
			opts: reqx.Options{Impersonation: reqx.Safari},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := New(tc.opts)
			tc.assert(t, router)
		})
	}
}

func TestRouter_RegisterFetchers(t *testing.T) {
	router := New(reqx.Options{})

	siteA := source.ID("site-a")
	siteB := source.ID("site-b")

	mockFetcher1 := impl.NewMockFetcher(impl.MockConfig{
		MockSourceID:  siteA,
		MockMainURL:   "site-a.com/",
		MockSourceURL: "site-a.com",
	}, impl.MockData{})

	mockFetcher2 := impl.NewMockFetcher(impl.MockConfig{
		MockSourceID:  siteB,
		MockMainURL:   "site-b.org/",
		MockSourceURL: "site-b.org",
	}, impl.MockData{})

	router.RegisterFetchers(mockFetcher1, mockFetcher2)

	assert.Len(t, router.fetchers, 2)
	assert.Len(t, router.Sources(), 2)
	assert.Contains(t, router.Sources(), siteA)
	assert.Contains(t, router.Sources(), siteB)
}

func TestRouter_TaskDispatching(t *testing.T) {
	siteA := source.ID("site-a")
	siteB := source.ID("site-b")

	mockFetcher1 := impl.NewMockFetcher(impl.MockConfig{
		MockSourceID:      siteA,
		MockSourceURL:     "site-a.com",
		MockDirectURL:     "direct.site-a.com/",
		MockMainURL:       "site-a.com/",
		MockAlternateURLs: []string{},
	}, impl.MockData{})

	mockFetcher2 := impl.NewMockFetcher(impl.MockConfig{
		MockSourceID:      siteB,
		MockSourceURL:     "site-b.com",
		MockDirectURL:     "direct.site-b.com/",
		MockMainURL:       "site-b.com/",
		MockAlternateURLs: []string{},
	}, impl.MockData{})

	router := New(reqx.Options{})
	router.RegisterFetchers(mockFetcher1, mockFetcher2)

	t.Run("TaskOf", func(t *testing.T) {
		t.Run("Finds correct fetcher", func(t *testing.T) {
			taskInstance, ok := router.TaskOf("https://site-a.com/char/1")
			assert.True(t, ok)
			assert.NotNil(t, taskInstance)
			assert.Equal(t, siteA, taskInstance.SourceID())
			assert.Equal(t, "site-a.com/char/1", taskInstance.NormalizedURL())
		})

		t.Run("Fails for unknown URL", func(t *testing.T) {
			_, ok := router.TaskOf("https://unknown.com/char/1")
			assert.False(t, ok)
		})
	})

	t.Run("TaskMapOf", func(t *testing.T) {
		urls := []string{
			"https://site-a.com/char/1",
			"http://www.site-b.com/id/2",
			"https://invalid.com/3",
		}
		bucket := router.TaskMapOf(urls...)

		assert.Len(t, bucket.Tasks, 2)
		assert.Len(t, bucket.ValidURLs, 2)
		assert.Len(t, bucket.InvalidURLs, 1)

		assert.Contains(t, bucket.Tasks, "site-a.com/char/1")
		assert.Contains(t, bucket.Tasks, "site-b.com/id/2")
		assert.Equal(t, "https://invalid.com/3", bucket.InvalidURLs[0])
	})

	t.Run("TaskSliceOf", func(t *testing.T) {
		urls := []string{
			"https://site-a.com/char/1",
			"https://invalid.com/3",
			"http://www.site-b.com/id/2",
		}
		slice := router.TaskSliceOf(urls...)

		assert.Len(t, slice.Tasks, 2)
		assert.Len(t, slice.ValidURLs, 2)
		assert.Len(t, slice.InvalidURLs, 1)

		assert.Equal(t, siteA, slice.Tasks[0].SourceID())
		assert.Equal(t, siteB, slice.Tasks[1].SourceID())
		assert.Equal(t, "https://invalid.com/3", slice.InvalidURLs[0])
	})
}

func TestRouter_Integrations(t *testing.T) {
	r := New(reqx.Options{
		RetryCount:    3,
		Impersonation: reqx.Chrome,
	})

	// Register all default fetchers
	builders := impl.DefaultBuilders(impl.BuilderOptions{PygmalionIdentityReader: cred.NewManager("pygmalion", cred.Env)})
	r.RegisterBuilders(builders...)

	result := r.CheckIntegrations()
	for sourceID, status := range result {
		assert.Equal(t, IntegrationSuccess, status, "Integrations failed for %s", sourceID)
	}
}
