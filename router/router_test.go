package router

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/factory"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFetcher struct {
	mock.Mock
}

func (m *mockFetcher) SourceURL() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockFetcher) IsSourceUp(c *req.Client) bool {
	args := m.Called(c)
	return args.Bool(0)
}

func (m *mockFetcher) DirectURL(characterID string) string {
	args := m.Called(characterID)
	return args.Get(0).(string)
}

func (m *mockFetcher) SourceID() source.ID {
	args := m.Called()
	return args.Get(0).(source.ID)
}

func (m *mockFetcher) CharacterID(url string, matchedURL string) string {
	args := m.Called(url, matchedURL)
	return args.String(0)
}

func (m *mockFetcher) NormalizeURL(characterID string) string {
	args := m.Called(characterID)
	return args.String(0)
}

func (m *mockFetcher) BaseURLs() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockFetcher) MainURL() string {
	return ""
}
func (m *mockFetcher) Extends(f fetcher.SourceHandler) {}
func (m *mockFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.CardInfo, models.JsonResponse, error) {
	return nil, models.EmptyJsonResponse, nil
}
func (m *mockFetcher) FetchCharacterCard(c *req.Client, metadata *models.CardInfo, response models.JsonResponse) (*png.CharacterCard, error) {
	return nil, nil
}

// MOCK FACTORY (newly added)
type mockFactory struct {
	mock.Mock
}

func (m *mockFactory) FetcherOf(id source.ID) fetcher.SourceHandler {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(fetcher.SourceHandler)
}

func newTestRouter(f factory.Factory) *Router {
	return &Router{
		client:  reqx.NewRetryClient(reqx.ClientOptions{}),
		factory: f,
	}
}

func TestNew(t *testing.T) {
	testCases := []struct {
		name   string
		opts   reqx.ClientOptions
		assert func(t *testing.T, r *Router)
	}{
		{
			name: "Default options",
			opts: reqx.ClientOptions{},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.client)
			},
		},
		{
			name: "With Retries",
			opts: reqx.ClientOptions{RetryCount: 5, MinBackoff: 1 * time.Second, MaxBackoff: 5 * time.Second},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r.client)
			},
		},
		{
			name: "With HTTP3",
			opts: reqx.ClientOptions{EnableHttp3: true},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r.client)
			},
		},
		{
			name: "With Chrome Impersonation",
			opts: reqx.ClientOptions{Impersonation: reqx.Chrome},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r.client)
			},
		},
		{
			name: "With Firefox Impersonation",
			opts: reqx.ClientOptions{Impersonation: reqx.Firefox},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r.client)
			},
		},
		{
			name: "With Safari Impersonation",
			opts: reqx.ClientOptions{Impersonation: reqx.Safari},
			assert: func(t *testing.T, r *Router) {
				assert.NotNil(t, r.client)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Pass a nil factory, as New() initializes its own.
			router := New(Options{ClientOptions: tc.opts, FactoryOptions: factory.Options{}})
			tc.assert(t, router)
		})
	}
}

func TestRouter_RegisterFetchers(t *testing.T) {
	// This test is adapted to use the mock factory.
	factory := new(mockFactory)
	router := newTestRouter(factory)

	mockF1 := new(mockFetcher)
	mockF2 := new(mockFetcher)
	siteA := source.ID("site-a")
	siteB := source.ID("site-b")
	mockF1.On("SourceID").Return(siteA)
	mockF2.On("SourceID").Return(siteB)

	// Configure the mock factory to return the mock fetchers.
	factory.On("FetcherOf", siteA).Return(mockF1)
	factory.On("FetcherOf", siteB).Return(mockF2)
	// Test the case where a fetcher is not found.
	factory.On("FetcherOf", source.ID("non-existent")).Return(nil)

	router.RegisterFetcher(siteA)
	assert.Len(t, router.fetchers, 1)

	router.RegisterFetchers(siteB, source.ID("non-existent"))
	// Should only register the valid fetcher (siteB).
	assert.Len(t, router.fetchers, 2)
	assert.Len(t, router.Sources(), 2)
}

func TestRouter_TaskDispatching(t *testing.T) {
	siteA := source.ID("site-a")
	siteB := source.ID("site-b")

	mockF1 := new(mockFetcher)
	mockF1.On("BaseURLs").Return([]string{"site-a.com/"})
	mockF1.On("CharacterID", "https://site-a.com/char/1", "site-a.com/").Return("1")
	mockF1.On("NormalizeURL", "1").Return("site-a.com/normalized/1")
	mockF1.On("SourceID").Return(siteA)

	mockF2 := new(mockFetcher)
	mockF2.On("BaseURLs").Return([]string{"site-b.org/"})
	mockF2.On("CharacterID", "http://www.site-b.org/id/2", "site-b.org/").Return("2")
	mockF2.On("NormalizeURL", "2").Return("site-b.org/normalized/2")
	mockF2.On("SourceID").Return(siteB)

	factory := new(mockFactory)
	factory.On("FetcherOf", siteA).Return(mockF1)
	factory.On("FetcherOf", siteB).Return(mockF2)

	router := newTestRouter(factory)
	router.RegisterFetchers(siteA, siteB)

	t.Run("TaskOf", func(t *testing.T) {
		t.Run("Finds correct fetcher", func(t *testing.T) {
			taskInstance, ok := router.TaskOf("https://site-a.com/char/1")
			assert.True(t, ok)
			assert.NotNil(t, taskInstance)
			assert.Equal(t, siteA, taskInstance.SourceID())
			assert.Equal(t, "site-a.com/normalized/1", taskInstance.NormalizedURL())
		})

		t.Run("Fails for unknown NormalizedURL", func(t *testing.T) {
			_, ok := router.TaskOf("https://unknown.com/char/1")
			assert.False(t, ok)
		})
	})

	t.Run("TaskMapOf", func(t *testing.T) {
		urls := []string{
			"https://site-a.com/char/1",
			"http://www.site-b.org/id/2",
			"https://invalid.com/3",
		}
		bucket := router.TaskMapOf(urls...)

		assert.Len(t, bucket.Tasks, 2)
		assert.Len(t, bucket.ValidURLs, 2)
		assert.Len(t, bucket.InvalidURLs, 1)

		assert.Contains(t, bucket.Tasks, "site-a.com/normalized/1")
		assert.Contains(t, bucket.Tasks, "site-b.org/normalized/2")
		assert.Equal(t, "https://invalid.com/3", bucket.InvalidURLs[0])
	})

	t.Run("TaskSliceOf", func(t *testing.T) {
		urls := []string{
			"https://site-a.com/char/1",
			"https://invalid.com/3",
			"http://www.site-b.org/id/2",
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
	r := New(Options{
		FactoryOptions: factory.Options{
			PygmalionIdentityProvider: cred.NewManager("pygmalion", cred.Env),
		},
		ClientOptions: reqx.ClientOptions{
			RetryCount:    3,
			Impersonation: reqx.Chrome,
		},
	})
	sourceIDs := slices.Collect(maps.Keys(GetResourceURLs()))
	r.RegisterFetchers(sourceIDs...)
	result := r.CheckIntegrations()
	for sourceID, status := range result {
		assert.Equal(t, status, IntegrationSuccess, "Integrations failed for %s", sourceID)
	}
}
