package router

import (
	"slices"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/factory"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/rs/zerolog/log"
)

type IntegrationStatus string

const (
	SourceDown               IntegrationStatus = "SOURCE DOWN"
	MissingRemoteResource    IntegrationStatus = "MISSING REMOTE RESOURCE"
	MismatchedRemoteResource IntegrationStatus = "MISMATCHED REMOTE RESOURCE"
	MissingLocalResource     IntegrationStatus = "MISSING LOCAL RESOURCE"
	IntegrationFailure       IntegrationStatus = "INTEGRATION FAILURE"
	IntegrationSuccess       IntegrationStatus = "INTEGRATION SUCCESS"
)

type TaskBucket struct {
	Tasks       map[string]task.Task
	ValidURLs   []string
	InvalidURLs []string
}

type TaskSlice struct {
	Tasks       []task.Task
	ValidURLs   []string
	InvalidURLs []string
}

type Options struct {
	FactoryOptions factory.Options
	ClientOptions  reqx.Options
}

type Router struct {
	client    *req.Client
	factory   factory.Factory
	fetcherMu sync.RWMutex
	fetchers  []fetcher.Fetcher
}

func New(opts Options) *Router {
	return &Router{
		client:  reqx.NewRetryClient(opts.ClientOptions),
		factory: factory.New(opts.FactoryOptions),
	}
}

func (r *Router) RegisterFetcher(sourceID source.ID) {
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	f := r.factory.FetcherOf(sourceID)
	if f == nil {
		log.Warn().Msgf("Count not find fetcher for source ID %s", sourceID)
		return
	}
	r.fetchers = append(r.fetchers, f)
}

func (r *Router) RegisterFetchers(sourceIDs ...source.ID) {
	for _, sourceID := range sourceIDs {
		r.RegisterFetcher(sourceID)
	}
}

func (r *Router) Sources() []source.ID {
	r.fetcherMu.RLock()
	defer r.fetcherMu.RUnlock()

	sources := make([]source.ID, len(r.fetchers))
	for index, f := range r.fetchers {
		sources[index] = f.SourceID()
	}

	return sources
}

func (r *Router) Fetchers() []fetcher.Fetcher {
	r.fetcherMu.RLock()
	defer r.fetcherMu.RUnlock()

	return slices.Clone(r.fetchers)
}

func (r *Router) CheckIntegrations() map[source.ID]IntegrationStatus {
	type result struct {
		id     source.ID
		status IntegrationStatus
	}
	var wg sync.WaitGroup
	resultsChan := make(chan result, len(r.fetchers))

	r.fetcherMu.RLock()
	for _, f := range r.fetchers {
		wg.Add(1)
		go func(fetcher fetcher.Fetcher) {
			defer wg.Done()
			status := r.checkIntegration(fetcher)
			resultsChan <- result{id: fetcher.SourceID(), status: status}
		}(f)
	}
	wg.Wait()
	close(resultsChan)
	r.fetcherMu.RUnlock()

	resultMap := make(map[source.ID]IntegrationStatus)
	for res := range resultsChan {
		resultMap[res.id] = res.status
	}

	return resultMap
}

func (r *Router) TaskMapOf(urls ...string) TaskBucket {
	var container TaskBucket
	container.Tasks = make(map[string]task.Task)
	for _, url := range urls {
		if fetcherTask, ok := r.TaskOf(url); ok {
			normalizedURL := fetcherTask.NormalizedURL()
			container.Tasks[normalizedURL] = fetcherTask
			container.ValidURLs = append(container.ValidURLs, normalizedURL)
		} else {
			container.InvalidURLs = append(container.InvalidURLs, url)
		}
	}
	return container
}

func (r *Router) TaskSliceOf(urls ...string) TaskSlice {
	var container TaskSlice
	for _, url := range urls {
		if fetcherTask, ok := r.TaskOf(url); ok {
			container.Tasks = append(container.Tasks, fetcherTask)
			container.ValidURLs = append(container.ValidURLs, fetcherTask.NormalizedURL())
		} else {
			container.InvalidURLs = append(container.InvalidURLs, url)
		}
	}
	return container
}

func (r *Router) TaskOf(url string) (task.Task, bool) {
	for _, f := range r.fetchers {
		if fetcherTask, ok := r.tryFetcher(f, url); ok {
			return fetcherTask, ok
		}
	}
	return nil, false
}

func (r *Router) tryFetcher(fetcher fetcher.Fetcher, url string) (task.Task, bool) {
	if matchedURL, found := stringsx.ContainsAny(url, fetcher.BaseURLs()...); found {
		return task.New(fetcher, url, matchedURL), true
	}
	return nil, false
}

func (r *Router) checkIntegration(f fetcher.Fetcher) IntegrationStatus {
	localCard, err := GetResourceCard(f.SourceID())
	if err != nil {
		return MissingLocalResource
	}

	isSourceUp := f.IsSourceUp()
	if !isSourceUp {
		return SourceDown
	}

	resourceURL, ok := GetResourceURL(f.SourceID())
	if !ok {
		return MismatchedRemoteResource
	}
	fetcherTask, ok := r.tryFetcher(f, resourceURL)
	if !ok {
		return MismatchedRemoteResource
	}

	metadata, err := fetcherTask.FetchMetadata()
	if err != nil {
		return MissingRemoteResource
	}
	characterCard, err := fetcherTask.FetchCharacterCard()
	if err != nil {
		return MissingRemoteResource
	}

	metadata, _ = fetcherTask.FetchMetadata()
	if characterCard.IsMalformed() {
		return IntegrationFailure
	}
	if !metadata.IsConsistentWith(characterCard.Sheet) {
		return IntegrationFailure
	}

	if !cmp.Equal(localCard.Sheet, characterCard.Sheet, cmpopts.EquateEmpty()) {
		return IntegrationFailure
	}

	return IntegrationSuccess
}
