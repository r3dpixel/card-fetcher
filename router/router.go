package router

import (
	"cmp"
	"slices"
	"sync"

	gcmp "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
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

var cmpOptions = []gcmp.Option{
	cmpopts.EquateEmpty(),
	cmpopts.SortSlices(comparator[string]),
	cmpopts.SortSlices(comparator[int]),
	cmpopts.SortSlices(comparator[float64]),
	cmpopts.SortSlices(comparator[property.String]),
	cmpopts.SortSlices(comparator[property.Integer]),
	cmpopts.SortSlices(comparator[property.Float]),
}

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

type Router struct {
	client    *reqx.Client
	fetcherMu sync.RWMutex
	fetchers  []fetcher.Fetcher
}

func New(opts reqx.Options) *Router {
	return &Router{
		client: reqx.NewClient(opts),
	}
}

func (r *Router) RegisterFetcher(fetcher fetcher.Fetcher) {
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	r.fetchers = append(r.fetchers, fetcher)
}

func (r *Router) RegisterFetchers(fetchers ...fetcher.Fetcher) {
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	r.fetchers = append(r.fetchers, fetchers...)
}

func (r *Router) RegisterBuilder(builder fetcher.Builder) {
	r.RegisterFetcher(builder.Build(r.client))
}

func (r *Router) RegisterBuilders(builders ...fetcher.Builder) {
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	for _, builder := range builders {
		r.fetchers = append(r.fetchers, builder.Build(r.client))
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

	if !gcmp.Equal(localCard.Sheet, characterCard.Sheet, cmpOptions...) {
		return IntegrationFailure
	}

	return IntegrationSuccess
}

func comparator[T cmp.Ordered](a, b T) bool {
	return a < b
}
