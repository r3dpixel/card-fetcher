package impl

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
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
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	jannyAIUuidLength int    = 36                              // JannyAI UUID length
	jannyAIDateFormat string = "2006-01-02 15:04:05.999999-07" // JannyAI date format

	jannyAIDomain          string = "jannyai.com"                                  // Domain for JannyAI
	jannyAIPath            string = "characters/"                                  // Path for JannyAI
	jannyAIApiURL          string = "https://api.jannyai.com/api/v1/characters/%s" // API URL for JannyAI
	jannyAIAvatarURL       string = "https://image.jannyai.com/bot-avatars/%s"     // Avatar URL for JannyAI
	jannyAIPlaceholderSize int    = 512                                            // Placeholder avatar size for JannyAI
)

// JannyAIInterceptor handles cookie extraction using chromedp for JannyAI.
type JannyAIInterceptor struct {
	chromePath func() string

	mu        sync.Mutex
	cookies   []*http.Cookie
	userAgent string
}

// jannyAIRecoveryResult holds the extracted cookies and user agent from Chrome.
type jannyAIRecoveryResult struct {
	cookies   []*http.Cookie
	userAgent string
}

// JannyAIOpts JannyAI fetcher options
type JannyAIOpts struct {
	ChromePath func() string
}

// JannyAIBuilder builder for JannyAI fetcher
type JannyAIBuilder JannyAIOpts

// Build creates a new JannyAI fetcher
func (b JannyAIBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewJannyAIFetcher(client, JannyAIOpts(b))
}

// jannyAIFetcher JannyAI fetcher implementation
type jannyAIFetcher struct {
	BaseFetcher
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
	}
	impl.Extends(impl)
	client.RegisterInterceptor(impl.serviceLabel, NewJannyAIInterceptor(opts.ChromePath))
	return impl
}

// CharacterID returns the character ID from a URL
func (f *jannyAIFetcher) CharacterID(rawCharacterID string) string {
	if len(rawCharacterID) < jannyAIUuidLength {
		return ""
	}
	return rawCharacterID[0:jannyAIUuidLength]
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *jannyAIFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	url := fmt.Sprintf(jannyAIApiURL, characterID)
	return f.client.IR(f.serviceLabel).Get(url)
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
		Tagline:       "",
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
	_, err := f.client.IR(f.serviceLabel).Get(url)
	return err
}

// Close closes the fetcher
func (f *jannyAIFetcher) Close() {
	f.client.UnregisterInterceptor(f.serviceLabel)
}

// NewJannyAIInterceptor creates a new JannyAI interceptor.
func NewJannyAIInterceptor(chromePath func() string) *JannyAIInterceptor {
	return &JannyAIInterceptor{chromePath: chromePath}
}

// ShouldIntercept returns true if the response indicates a Cloudflare challenge.
func (i *JannyAIInterceptor) ShouldIntercept(resp *req.Response, _ error) bool {
	if resp == nil || resp.Response == nil {
		return false
	}
	if resp.StatusCode == 403 {
		return true
	}
	body := resp.String()
	return strings.Contains(body, "cf-browser-verification") ||
		strings.Contains(body, "Just a moment")
}

// Recover uses chromedp to pass the Cloudflare challenge and extract cookies.
func (i *JannyAIInterceptor) Recover(_ *reqx.Client, r *req.Response) error {
	// Get chrome path from closure
	var chromePath string
	if i.chromePath != nil {
		chromePath = i.chromePath()
	}

	// Run Chrome and extract cookies/user agent
	result, err := chromex.RunChrome(chromex.Options{
		Path:    chromePath,
		Timeout: chromex.DefaultTimeout,
		Flags: []chromedp.ExecAllocatorOption{
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("disable-infobars", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("headless", false),
		},
	}, func(ctx context.Context) (jannyAIRecoveryResult, error) {
		// Navigate to the URL
		if err := chromedp.Run(ctx, chromedp.Navigate(r.Request.RawURL)); err != nil {
			return jannyAIRecoveryResult{}, fmt.Errorf("navigation failed: %w", err)
		}

		// Wait for Cloudflare challenge to pass
		for range 60 {
			var html string
			if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &html)); err != nil {
				time.Sleep(time.Second)
				continue
			}
			if !strings.Contains(html, "Just a moment") &&
				!strings.Contains(html, "cf-browser-verification") &&
				len(html) > 500 {
				break
			}
			time.Sleep(time.Second)
		}

		// Extract cookies
		var networkCookies []*network.Cookie
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			networkCookies, err = network.GetCookies().Do(ctx)
			return err
		})); err != nil {
			return jannyAIRecoveryResult{}, fmt.Errorf("failed to get cookies: %w", err)
		}

		cookies := make([]*http.Cookie, 0, len(networkCookies))
		for _, cookie := range networkCookies {
			if strings.Contains(cookie.Domain, jannyAIDomain) ||
				strings.Contains(jannyAIDomain, cookie.Domain) {
				cookies = append(cookies, &http.Cookie{
					Name:   cookie.Name,
					Value:  cookie.Value,
					Domain: cookie.Domain,
					Path:   cookie.Path,
				})
			}
		}

		// Extract user agent
		var userAgent string
		if err := chromedp.Run(ctx, chromedp.Evaluate(`navigator.userAgent`, &userAgent)); err != nil {
			return jannyAIRecoveryResult{}, fmt.Errorf("failed to get user agent: %w", err)
		}

		return jannyAIRecoveryResult{cookies: cookies, userAgent: userAgent}, nil
	})
	if err != nil {
		return err
	}

	// Store the result
	i.mu.Lock()
	i.cookies = result.cookies
	i.userAgent = result.userAgent
	i.mu.Unlock()

	return nil
}

// Apply sets the cookies and user agent on the request.
func (i *JannyAIInterceptor) Apply(r *req.Request) *req.Request {
	i.mu.Lock()
	defer i.mu.Unlock()

	if len(i.cookies) > 0 {
		r.SetCookies(i.cookies...)
	}
	if i.userAgent != "" {
		r.SetHeader("User-Agent", i.userAgent)
	}
	return r
}

// MaxRetries returns the maximum number of recovery attempts.
func (i *JannyAIInterceptor) MaxRetries() int {
	return 3
}
