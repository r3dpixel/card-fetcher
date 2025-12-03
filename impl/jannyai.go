package impl

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/chromex"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	jannyAIUuidLength int    = 36                              // JannyAI UUID length
	jannyAIDateFormat string = "2006-01-02 15:04:05.999999-07" // JannyAI date format

	jannyAIDomain          string = "jannyai.com"                                  // Domain for JannyAI
	jannyAIPath            string = "characters/"                                  // Path for JannyAI
	jannyAIBaseApiURL      string = "https://api.jannyai.com"                      // Base API URL for JannyAI
	jannyAIApiURL          string = "https://api.jannyai.com/api/v1/characters/%s" // API URL for JannyAI
	jannyAIAvatarURL       string = "https://image.jannyai.com/bot-avatars/%s"     // Avatar URL for JannyAI
	jannyAIPlaceholderSize int    = 512                                            // Placeholder avatar size for JannyAI
)

// JannyCookies contains cookies required for fetching JannyAI metadata
type JannyCookies struct {
	CloudflareClearance string
	UserAgent           string
}

// JannyChromeConfig contains configuration for fetching JannyAI cookies using Chrome
type JannyChromeConfig struct {
	Path             string
	AutoFetchCookies bool
}

// JannyAIOpts contains options for JannyAI fetcher
type JannyAIOpts struct {
	ChromeConfig   func() JannyChromeConfig
	CookieProvider func() JannyCookies
}

// JannyAIBuilder builder for JannyAI fetcher
type JannyAIBuilder JannyAIOpts

// Build creates a new JannyAI fetcher using the configured options
func (b JannyAIBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewJannyAIFetcher(client, JannyAIOpts(b))
}

// jannyAIFetcher JannyAI fetcher implementation
type jannyAIFetcher struct {
	BaseFetcher
	cookieProvider func() JannyCookies
	chromeConfig   func() JannyChromeConfig
	cookies        JannyCookies
	cookiesMu      sync.RWMutex
	refreshMu      sync.Mutex
}

// NewJannyAIFetcher creates a new JannyAI fetcher
func NewJannyAIFetcher(client *reqx.Client, opts JannyAIOpts) fetcher.Fetcher {
	mainURL := path.Join(jannyAIDomain, jannyAIPath)
	impl := &jannyAIFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.JannyAI,
			sourceURL: jannyAIDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: jannyAIDomain, Path: jannyAIPath},
			},
		},
		cookieProvider: opts.CookieProvider,
		chromeConfig:   opts.ChromeConfig,
	}
	impl.Extends(impl)
	return impl
}

// CharacterID returns the character ID from a URL
func (f *jannyAIFetcher) CharacterID(rawCharacterID string) string {
	if len(rawCharacterID) < jannyAIUuidLength {
		return stringsx.Empty
	}
	return rawCharacterID[0:jannyAIUuidLength]
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *jannyAIFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	url := fmt.Sprintf(jannyAIApiURL, characterID)
	return f.executeRequestWithRefresh(url)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *jannyAIFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "id")
}

// FetchCardInfo fetches the card info from the source
func (f *jannyAIFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the name
	name := metadataBinder.Get("name").String()
	// Extract the creation time
	createTime := timestamp.ParseF(jannyAIDateFormat, metadataBinder.Get("createdAt").String(), trace.URL, metadataBinder.NormalizedURL)
	// Extract tags
	tags := models.TagsFromJsonArray(metadataBinder.Get("tags"), func(result *sonicx.Wrap) string {
		return result.Get("name").String()
	})

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     metadataBinder.DirectURL,
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Name:          name,
		Title:         name,
		Tagline:       stringsx.Empty,
		CreateTime:    createTime,
		UpdateTime:    timestamp.Nano(time.Now().Truncate(24 * time.Hour).UnixNano()),
		IsForked:      false,
		Tags:          tags,
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *jannyAIFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	username := metadataBinder.Get("creatorName").String()
	return &models.CreatorInfo{
		Nickname:   username,
		Username:   username,
		PlatformID: metadataBinder.Get("creatorId").String(),
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *jannyAIFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Fetch the character card from the API
	jannyAIAvatarURL := fmt.Sprintf(jannyAIAvatarURL, binder.Get("avatar").String())
	rawCard, err := png.FromURL(f.client, jannyAIAvatarURL).LastVersion().Get()
	if err != nil {
		// If the API call fails, return a placeholder character card
		rawCard, err = png.PlaceholderCharacterCard(jannyAIPlaceholderSize)
	}
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Update the character card fields
	characterCard.Description = property.String(binder.Get("personality").String())
	characterCard.Scenario = property.String(binder.Get("scenario").String())
	characterCard.FirstMessage = property.String(binder.Get("firstMessage").String())
	characterCard.MessageExamples = property.String(binder.Get("exampleDialogs").String())
	characterCard.CreatorNotes = property.String(binder.Get("description").String())

	// Return the character card
	return characterCard, nil
}

// IsSourceUp checks if the source is up
func (f *jannyAIFetcher) IsSourceUp() error {
	url := "https://" + f.SourceURL() + "/collections"
	_, err := f.executeRequestWithRefresh(url)
	return err
}

// executeRequestWithRefresh executes a request with automatic cookie refresh on 403
// only one goroutine refreshes cookies at a time
func (f *jannyAIFetcher) executeRequestWithRefresh(url string) (*req.Response, error) {
	// Get cached cookies
	cookies := f.getCachedCookies()

	// Execute request with current cookies
	response, err := f.makeRequest(url, cookies)
	// Handle 403 - refresh and retry
	if response != nil && response.StatusCode == 403 {
		return f.handleExpiredCookies(url)
	}
	// Passthrough other errors
	if err != nil {
		return nil, err
	}

	// Return response
	return response, nil
}

// makeRequest executes HTTP request with given cookies
func (f *jannyAIFetcher) makeRequest(url string, cookies JannyCookies) (*req.Response, error) {
	return f.client.R().
		SetHeader("User-Agent", cookies.UserAgent).
		SetCookies(
			&http.Cookie{
				Name:  "cf_clearance",
				Value: cookies.CloudflareClearance,
			},
		).
		Get(url)
}

// handleExpiredCookies refreshes cookies and retries request on 403
func (f *jannyAIFetcher) handleExpiredCookies(url string) (*req.Response, error) {
	f.refreshMu.Lock()

	// Check if another goroutine already refreshed cookies while we were waiting
	cachedCookies := f.getCachedCookies()
	response, err := f.makeRequest(url, cachedCookies)
	if err == nil && (response == nil || response.StatusCode != 403) {
		f.refreshMu.Unlock()
		return response, err
	}

	// Still 403 - we need to refresh
	newCookies, err := f.refreshCookiesLocked()
	if err != nil {
		f.refreshMu.Unlock()
		return nil, fetcher.NewError(err, fetcher.InvalidCredentialsErr)
	}

	// Update cached cookies
	f.setCachedCookies(newCookies)

	// Release lock before retry
	f.refreshMu.Unlock()

	// Retry without lock
	response, err = f.makeRequest(url, newCookies)

	// Check for 403 again
	if response != nil && response.StatusCode == 403 {
		return nil, fetcher.NewError(err, fetcher.InvalidCredentialsErr)
	}

	// Passthrough other errors
	if err != nil {
		return nil, err
	}

	// Return response
	return response, nil
}

// getCachedCookies safely retrieves the cached cookies with read lock
func (f *jannyAIFetcher) getCachedCookies() JannyCookies {
	f.cookiesMu.RLock()
	defer f.cookiesMu.RUnlock()
	return f.cookies
}

// setCachedCookies safely updates the cached cookies with write lock
func (f *jannyAIFetcher) setCachedCookies(cookies JannyCookies) {
	f.cookiesMu.Lock()
	defer f.cookiesMu.Unlock()
	f.cookies = cookies
}

// refreshCookiesLocked refreshes cookies based on configuration (assumes refreshMu is already locked)
func (f *jannyAIFetcher) refreshCookiesLocked() (JannyCookies, error) {
	// If chromeConfig provider is available, try auto-fetching first
	if f.chromeConfig != nil {
		chromeConfig := f.chromeConfig()

		// If auto-fetching cookies is enabled, fetch them directly
		if chromeConfig.AutoFetchCookies {
			return f.fetchJannyCookies(chromeConfig.Path)
		}
	}

	// Fall back to the configured cookie provider
	return f.getCookiesFromProvider()
}

// fetchJannyCookies fetches cookies from JannyAI using Chrome
func (f *jannyAIFetcher) fetchJannyCookies(chromePath string) (JannyCookies, error) {
	// Configure Chrome
	config := chromex.Options{
		Path: chromePath,
	}

	// Run Chrome and wait for cookies to be set
	return chromex.RunChrome(config, func(ctx context.Context) (JannyCookies, error) {
		// Navigate to the character page (don't use WaitReady - we'll poll for cookies instead)
		if err := chromedp.Run(ctx,
			chromedp.Navigate(jannyAIBaseApiURL),
		); err != nil {
			return JannyCookies{}, fmt.Errorf("failed to navigate: %w", err)
		}

		// Temporary cookies storage
		var cookies []*network.Cookie
		var userAgent string

		// Poll for cookies to be set (browser stays alive during polling)
		// This gives Cloudflare's challenge time to complete
		if err := f.waitForCookies(ctx, 30*time.Second, &cookies); err != nil {
			return JannyCookies{}, fmt.Errorf("timeout waiting for cookies: %w", err)
		}

		// Get user agent (browser still active from polling)
		if err := chromedp.Run(ctx, chromedp.Evaluate("navigator.userAgent", &userAgent)); err != nil {
			return JannyCookies{}, fmt.Errorf("failed to get user agent: %w", err)
		}

		// Find cf_clearance cookie
		var cfClearance string
		for _, c := range cookies {
			if c.Name == "cf_clearance" {
				cfClearance = c.Value
				break
			}
		}

		// Validate cf_clearance cookie
		if stringsx.IsBlank(cfClearance) {
			return JannyCookies{}, fmt.Errorf("cf_clearance cookie not found")
		}

		// Return cookies
		return JannyCookies{
			CloudflareClearance: cfClearance,
			UserAgent:           userAgent,
		}, nil
	})
}

// waitForCookies waits for cookies to be set in Chrome
func (f *jannyAIFetcher) waitForCookies(ctx context.Context, timeout time.Duration, cookies *[]*network.Cookie) error {
	// Wait for cookies to be set
	deadline := time.Now().Add(timeout)

	// Check every 100ms for cookies
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Loop until cookies are set or timeout is reached
	for {
		select {
		// Check if timeout was reached
		case <-time.After(time.Until(deadline)):
			// Timeout reached
			return fmt.Errorf("timeout waiting for cookies")
		case <-ticker.C:
			// Check if cookies are set
			var tempCookies []*network.Cookie
			if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				tempCookies, err = network.GetCookies().Do(ctx)
				return err
			})); err == nil && len(tempCookies) > 0 {
				*cookies = tempCookies
				return nil
			}
		}
	}
}

// getCookiesFromProvider retrieves cookies from the configured cookie provider
func (f *jannyAIFetcher) getCookiesFromProvider() (JannyCookies, error) {
	// Otherwise, use the configured cookie provider
	if f.cookieProvider == nil {
		// Cookie provider isn't configured, fail
		return JannyCookies{}, fetcher.NewError(nil, fetcher.MissingCookieProviderErr)
	}

	// Fetch cookies from the cookie provider
	cookies := f.cookieProvider()

	// Validate cookies
	if stringsx.IsBlank(cookies.UserAgent) || stringsx.IsBlank(cookies.CloudflareClearance) {
		return JannyCookies{}, fetcher.NewError(fmt.Errorf("missing jannyai cookies"), fetcher.InvalidCredentialsErr)
	}

	// Return cookies
	return cookies, nil
}
