package router

import (
	"context"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/impl"
	"github.com/r3dpixel/card-fetcher/snapshots"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-fetcher/task"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/iterx"
	"github.com/r3dpixel/toolkit/lexer"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/scheduler"
	"github.com/r3dpixel/toolkit/symbols"
)

const (
	defaultMaxParallelism = 4
)

// IntegrationStatus represents the status of a fetcher integration test
type IntegrationStatus string

// Integration statuses
const (
	MissingFetcher           IntegrationStatus = "MISSING FETCHER"
	SourceDown               IntegrationStatus = "SOURCE DOWN"
	InvalidCredentials       IntegrationStatus = "INVALID CREDENTIALS"
	MissingRemoteResource    IntegrationStatus = "MISSING REMOTE RESOURCE"
	MismatchedRemoteResource IntegrationStatus = "MISMATCHED REMOTE RESOURCE"
	MissingLocalResource     IntegrationStatus = "MISSING LOCAL RESOURCE"
	IntegrationFailure       IntegrationStatus = "INTEGRATION FAILURE"
	IntegrationSuccess       IntegrationStatus = "INTEGRATION SUCCESS"
)

// TaskBucket represents a collection of tasks for a given set of URLs (indexed by normalized URL)
type TaskBucket struct {
	Tasks       map[string]task.Task
	ValidURLs   []string
	InvalidURLs []string
}

// TaskSlice represents a collection of tasks
type TaskSlice struct {
	Tasks       []task.Task
	ValidURLs   []string
	InvalidURLs []string
}

// lexResult represents a result from the lexer trie
type lexResult struct {
	fetcher fetcher.Fetcher
	baseURL fetcher.BaseURL
}

// integrationCheckParams represents the parameters for checking the integration of a resource URL
type integrationCheckParams struct {
	sourceID    source.ID
	index       int
	resourceURL string
	resultCh    chan<- IntegrationStatus
}

// Router routes URLs to fetchers
type Router struct {
	client    *reqx.Client
	lex       *lexer.Lexer[rune, lexResult]
	fetchers  map[source.ID]fetcher.Fetcher
	fetcherMu sync.RWMutex
}

// EnvConfigured creates a new router with default builders configured for environment variables
func EnvConfigured(chromePath func() string) *Router {
	// Create a new router with default options
	r := New(
		reqx.Options{
			RetryCount:        4,
			MinBackoff:        10 * time.Millisecond,
			MaxBackoff:        500 * time.Millisecond,
			DisableKeepAlives: true,
			Impersonation:     reqx.Chrome,
			Timeout:           20 * time.Second,
		},
	)

	// Create builders with environment variables (PYGMALION_USERNAME, PYGMALION_PASSWORD)
	builders := impl.DefaultBuilders(impl.BuilderOptions{
		PygmalionIdentityReader: cred.NewManager("pygmalion", cred.Env),
		ChromePath:              chromePath,
	})

	// Register builders with the router
	r.RegisterBuilders(builders...)

	// Return the router
	return r
}

// New creates a new router with the given options
func New(opts reqx.Options) *Router {
	return &Router{
		client:   reqx.NewClient(opts),
		lex:      lexer.New[rune, lexResult](),
		fetchers: make(map[source.ID]fetcher.Fetcher),
	}
}

// RegisterFetcher registers a fetcher with the router
func (r *Router) RegisterFetcher(fetcher fetcher.Fetcher) {
	// Lock fetchers map
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	// Register fetcher
	r.unsafeRegisterFetcher(fetcher)
}

// RegisterFetchers registers multiple fetchers with the router
func (r *Router) RegisterFetchers(fetchers ...fetcher.Fetcher) {
	// Lock fetchers map
	r.fetcherMu.Lock()
	defer r.fetcherMu.Unlock()

	// Register fetchers
	for _, f := range fetchers {
		r.unsafeRegisterFetcher(f)
	}
}

// RegisterBuilder registers a builder with the router
func (r *Router) RegisterBuilder(builder fetcher.Builder) {
	r.RegisterFetcher(builder.Build(r.client))
}

// RegisterBuilders registers multiple builders with the router
func (r *Router) RegisterBuilders(builders ...fetcher.Builder) {
	// Create fetchers from builders
	fetchers := make([]fetcher.Fetcher, 0, len(builders))
	for _, builder := range builders {
		fetchers = append(fetchers, builder.Build(r.client))
	}
	// Register fetchers
	r.RegisterFetchers(fetchers...)
}

// unsafeRegisterFetcher registers a fetcher with the router without locking the fetchers map
func (r *Router) unsafeRegisterFetcher(fetcher fetcher.Fetcher) {
	// Register fetcher
	r.fetchers[fetcher.SourceID()] = fetcher
	// Register base URLs (we reverse the base URL and start searching from the first `/` backwards)
	for _, baseURL := range fetcher.BaseURLs() {
		r.lex.InsertIter(iterx.RunesReverse(baseURL.Domain), lexResult{fetcher, baseURL})
	}
}

// Sources returns the list of sources registered with the router
func (r *Router) Sources() []source.ID {
	// Lock fetchers map for reading
	r.fetcherMu.RLock()
	defer r.fetcherMu.RUnlock()

	// Return the list of sources
	return slices.Collect(maps.Keys(r.fetchers))
}

// Fetchers returns the list of fetchers registered with the router
func (r *Router) Fetchers() []fetcher.Fetcher {
	// Lock fetchers map for reading
	r.fetcherMu.RLock()
	defer r.fetcherMu.RUnlock()

	// Return the list of fetchers
	return slices.Collect(maps.Values(r.fetchers))
}

// TaskMapOf creates a map of tasks for the given URLs (indexed by normalized URL)
func (r *Router) TaskMapOf(urls ...string) TaskBucket {
	// Create a map of tasks
	var container TaskBucket
	container.Tasks = make(map[string]task.Task)

	// Iterate over the URLs and create tasks
	for _, url := range urls {
		// Try to create a task for the URL
		if fetcherTask, ok := r.TaskOf(url); ok {
			// Get the normalized URL
			normalizedURL := fetcherTask.NormalizedURL()
			// Add the task to the map
			container.Tasks[normalizedURL] = fetcherTask
			// Add the normalized URL to the list of valid URLs
			container.ValidURLs = append(container.ValidURLs, normalizedURL)
		} else {
			// Add the URL to the list of invalid URLs
			container.InvalidURLs = append(container.InvalidURLs, url)
		}
	}
	// Return the map
	return container
}

// TaskSliceOf creates a slice of tasks for the given URLs
func (r *Router) TaskSliceOf(urls ...string) TaskSlice {
	// Create a slice of tasks
	var container TaskSlice
	// Iterate over the URLs and create tasks
	for _, url := range urls {
		// Try to create a task for the URL
		if fetcherTask, ok := r.TaskOf(url); ok {
			// Add the task to the slice
			container.Tasks = append(container.Tasks, fetcherTask)
			// Add the normalized URL to the list of valid URLs
			container.ValidURLs = append(container.ValidURLs, fetcherTask.NormalizedURL())
		} else {
			// Add the URL to the list of invalid URLs
			container.InvalidURLs = append(container.InvalidURLs, url)
		}
	}
	// Return the slice
	return container
}

// TaskOf tries to create a task for the given URL using the fetcher with the shortest base URL
func (r *Router) TaskOf(url string) (task.Task, bool) {
	// Remove trailing slash
	if url[len(url)-1] == symbols.SlashByte {
		url = url[:len(url)-1]
	}

	// Find the start of the domain (https://example.com/path -> example.com, ://)
	// If the scheme exists, we skip the first 3 characters (://)
	// If the scheme does not exist, the last index returns -1 => domainStart = 0
	domainStart := strings.LastIndexByte(url[:min(10, len(url))], symbols.ColonByte)
	if domainStart >= 0 {
		domainStart += 3
	} else {
		domainStart = 0
	}

	// Find the end of the domain (https://example.com/path -> /path)
	// The next slash is the end of the domain
	domainEnd := domainStart + strings.IndexByte(url[domainStart:], symbols.SlashByte)

	// If no domain was found, return nil and false
	if domainEnd <= 0 {
		return nil, false
	}

	// Use the lexer trie to match all domains in O(N) time
	// All domains end at the first slash, so we search in reverse, starting from the end of the domain
	match, _, found := r.lex.FirstMatch(iterx.RunesReverse(url[domainStart:domainEnd]))

	// If no fetcher was found, return nil and false
	if !found {
		return nil, false
	}

	// Move the domainEnd back to the start of the domain, not the slash
	domainEnd++

	// The rest of the URL must have the path prefix of the baseURL
	if !strings.HasPrefix(url[domainEnd:], match.baseURL.Path) {
		return nil, false
	}

	// Return nil and false if no base URL matched the URL
	return task.New(match.fetcher, url, url[domainEnd+len(match.baseURL.Path):]), true
}

// CheckIntegration checks the integration status of a given source
func (r *Router) CheckIntegration(sourceID source.ID) IntegrationStatus {
	// Get fetcher
	r.fetcherMu.RLock()
	f, ok := r.fetchers[sourceID]
	r.fetcherMu.RUnlock()

	// If the fetcher is not found, return MISSING FETCHER
	if !ok {
		return MissingFetcher
	}

	// Check if the source is up
	if err := f.IsSourceUp(); err != nil {
		// Return INVALID CREDENTIALS if there was a credential error
		if fetcher.GetErrCode(err) == fetcher.InvalidCredentialsErr {
			return InvalidCredentials
		}
		// Otherwise, return SOURCE DOWN
		return SourceDown
	}

	// Get resource URLs
	resourceURLs, ok := snapshots.GetResourceURLs(sourceID)
	if !ok || len(resourceURLs) == 0 {
		return MismatchedRemoteResource
	}

	// Check each resource URL in parallel using scheduler pool
	resultCh := make(chan IntegrationStatus, len(resourceURLs))

	// If there is only one resource URL, check it synchronously
	if len(resourceURLs) == 1 {
		r.checkResourceIntegration(integrationCheckParams{
			sourceID:    sourceID,
			index:       0,
			resourceURL: resourceURLs[0],
			resultCh:    resultCh,
		})
		close(resultCh)
		return <-resultCh
	}

	// Set the maximum number of goroutines
	parallelism := len(resourceURLs)
	if parallelism > defaultMaxParallelism {
		parallelism = defaultMaxParallelism
	}

	// Create a pool to check each resource URL
	pool := scheduler.NewPool(scheduler.Options[integrationCheckParams]{
		Handler: func(_ context.Context, p integrationCheckParams) {
			r.checkResourceIntegration(p)
		},
		Parallelism: parallelism,
	})

	// Submit tasks to the pool
	for index, resourceURL := range resourceURLs {
		pool.Submit(integrationCheckParams{
			sourceID:    sourceID,
			index:       index,
			resourceURL: resourceURL,
			resultCh:    resultCh,
		})
	}

	// Close the pool and wait for all tasks to finish
	pool.Close()
	close(resultCh)

	// Return first failure found, or success if all passed
	for status := range resultCh {
		if status != IntegrationSuccess {
			return status
		}
	}

	// Return INTEGRATION SUCCESS
	return IntegrationSuccess
}

// checkResourceIntegration checks the integration of a single resource URL
func (r *Router) checkResourceIntegration(p integrationCheckParams) {
	// Try to create a task for the resource URL
	resourceTask, ok := r.TaskOf(p.resourceURL)
	if !ok || resourceTask.SourceID() != p.sourceID {
		p.resultCh <- MismatchedRemoteResource
		return
	}

	// Get the local card corresponding to the resource URL
	localCard, err := snapshots.GetResourceCard(p.sourceID, p.index)
	if err != nil {
		p.resultCh <- MissingLocalResource
		return
	}

	// Check the integration of the task and the local card
	if status := r.checkTaskIntegration(resourceTask, localCard); status != IntegrationSuccess {
		p.resultCh <- status
		return
	}

	// If the integration check passed, return INTEGRATION SUCCESS
	p.resultCh <- IntegrationSuccess
}

// checkTaskIntegration checks the integration status of a given task and local character card
func (r *Router) checkTaskIntegration(task task.Task, localCard *png.CharacterCard) IntegrationStatus {
	// Fetch metadata and character card
	metadata, characterCard, err := task.FetchAll()
	if err != nil {
		// Return INVALID CREDENTIALS if there was a credential error
		if fetcher.GetErrCode(err) == fetcher.InvalidCredentialsErr {
			return InvalidCredentials
		}
		// Otherwise, return MISSING REMOTE RESOURCE
		return MissingRemoteResource
	}

	// If the character card is not valid, or not consistent with the metadata, return INTEGRATION FAILURE
	if !characterCard.Integrity() || !metadata.IsConsistentWith(characterCard.Sheet) {
		return IntegrationFailure
	}

	// Set the local card modification date to the remote card modification date (modification date was already validated above)
	localCard.Sheet.ModificationDate = characterCard.Sheet.ModificationDate
	// If the local card is not equal to the remote card, return INTEGRATION FAILURE
	if !localCard.Sheet.DeepEquals(characterCard.Sheet) {
		return IntegrationFailure
	}

	// Return INTEGRATION SUCCESS
	return IntegrationSuccess
}
