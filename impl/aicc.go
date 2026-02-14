package impl

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	aiccDomain     string = "aicharactercards.com"                                      // Domain for AICC
	aiccPath       string = "charactercards/"                                           // Path for AICC
	aiccPageURL    string = "https://aicharactercards.com/charactercards/%s/"           // AICC page URL
	aiccImageURL   string = "https://aicharactercards.com/wp-json/pngapi/v1/image/%s"   // AICC image URL
	aiccDetailsURL string = "https://aicharactercards.com/wp-json/pngapi/v1/details/%s" // AICC details URL
	aiccDateFormat string = time.RFC3339                                                // Date Format for AICC
)

// aiccDetails represents the partial response from the details API
type aiccDetails struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Excerpt string `json:"excerpt"`
}

// AiccBuilder builder for AICC fetcher
type AiccBuilder struct{}

// Build creates a new AICC fetcher
func (b AiccBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewAiccFetcher(client)
}

// aiccFetcher AICC fetcher implementation
type aiccFetcher struct {
	BaseFetcher
}

// NewAiccFetcher create a new AICC fetcher
func NewAiccFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(aiccDomain, aiccPath)
	impl := &aiccFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.AICC,
			sourceURL: aiccDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: aiccDomain, Path: aiccPath},
			},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadataResponse fetches the HTML page from the source
func (f *aiccFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	return f.client.R().Get(fmt.Sprintf(aiccPageURL, characterID))
}

// CreateBinder stores the raw HTML in StringResponse (no JSON parsing for AICC)
func (f *aiccFetcher) CreateBinder(characterID string, response string) (*fetcher.MetadataBinder, error) {
	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response))
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedMetadataErr)
	}

	// Fetch details from API
	details, err := f.fetchDetails(characterID)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedMetadataErr)
	}

	// Return binder
	return &fetcher.MetadataBinder{
		CharacterID:   characterID,
		NormalizedURL: f.NormalizeURL(characterID),
		DirectURL:     f.DirectURL(characterID),
		Document:      doc,
		JsonResponse:  details,
	}, nil
}

// FetchCardInfo extracts card info from the details API and HTML
func (f *aiccFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the dates from the HTML
	datePublished, dateModified := f.extractDates(metadataBinder.Document)

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    f.extractPostID(metadataBinder.Document),
		CharacterID:   metadataBinder.CharacterID,
		Name:          f.extractName(metadataBinder.Document),
		Title:         metadataBinder.Get("title").String(),
		Tagline:       metadataBinder.Get("excerpt").String(),
		CreateTime:    timestamp.ParseF(aiccDateFormat, datePublished, trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(aiccDateFormat, dateModified, trace.URL, metadataBinder.NormalizedURL),
		IsForked:      false,
		Tags:          f.extractTags(metadataBinder.Document),
	}, nil
}

// FetchCreatorInfo extracts creator info from the HTML using goquery
func (f *aiccFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	user := metadataBinder.Get("author").String()
	return &models.CreatorInfo{
		Nickname:   user,
		Username:   user,
		PlatformID: user,
	}, nil
}

// FetchCharacterCard downloads the PNG from the source
func (f *aiccFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Extract author/title from characterID (truncated path)
	authorTitle, err := f.extractAuthorTitle(binder.CharacterID)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Download PNG from API
	downloadURL := fmt.Sprintf(aiccImageURL, authorTitle)
	rawCard, err := png.FromURL(f.client, downloadURL).LastVersion().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Return the character card
	return characterCard, nil
}

// fetchDetails fetches metadata from the details API endpoint
func (f *aiccFetcher) fetchDetails(characterID string) (fetcher.JsonResponse, error) {
	// Extract author/title from characterID (truncated path)
	authorTitle, err := f.extractAuthorTitle(characterID)
	if err != nil {
		return nil, err
	}

	// Fetch details from API
	response, err := reqx.String(
		f.client.R().
			Get(fmt.Sprintf(aiccDetailsURL, authorTitle)),
	)
	if err != nil {
		return nil, err
	}

	// Parse response
	return sonicx.GetFromString(response)
}

// extractAuthorTitle extracts author/title from characterID (format: category/author/title)
func (f *aiccFetcher) extractAuthorTitle(characterID string) (string, error) {
	// Extract author/title from characterID (category/author/title)
	_, after, found := strings.Cut(characterID, "/")
	if !found {
		return "", fmt.Errorf("invalid character ID format: %s", characterID)
	}

	// Return author/title
	return after, nil
}

// extractPostID extracts the post ID from HTML
func (f *aiccFetcher) extractPostID(doc *goquery.Document) string {
	// Try the download link data-post-id attribute
	if postID, exists := doc.Find("#downloadLink").Attr("data-post-id"); exists && stringsx.IsNotBlank(postID) {
		return postID
	}

	// Try the hidden input field
	if postID, exists := doc.Find(`input[name="post_id"]`).Attr("value"); exists && stringsx.IsNotBlank(postID) {
		return postID
	}

	// Try card grade element
	if postID, exists := doc.Find(".accb-card-grade").Attr("data-post-id"); exists && stringsx.IsNotBlank(postID) {
		return postID
	}

	// Try chat container (uses data-postid without a hyphen)
	if postID, exists := doc.Find("#aicc-chat-container").Attr("data-postid"); exists && stringsx.IsNotBlank(postID) {
		return postID
	}

	// Return empty string if nothing found
	return ""
}

// extractDates extracts from the JSON-LD metadata the datePublished and dateModified values
func (f *aiccFetcher) extractDates(doc *goquery.Document) (datePublished, dateModified string) {
	// Initialize empty dates
	datePublished, dateModified = "", ""
	// Find all JSON-LD scripts
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		// Parse JSON-LD script @graph path
		graph, err := sonicx.GetFromString(s.Text(), "@graph")
		if err != nil {
			return
		}

		// Extract the slice of nodes
		nodes, err := graph.ArrayUseNode()
		if err != nil {
			return
		}

		// Iterate over nodes and extract datePublished and dateModified
		for _, pointer := range nodes {
			// Extract datePublished and dateModified
			publishedNode := pointer.Get("datePublished")
			modifiedNode := pointer.Get("dateModified")
			// Check if either datePublished or dateModified exists
			if publishedNode.Exists() || modifiedNode.Exists() {
				// Return if any dates exist
				datePublished = sonicx.Of(*publishedNode).String()
				dateModified = sonicx.Of(*modifiedNode).String()
				return
			}
		}
	})
	return
}

// extractTags extracts tags from the character tag section
func (f *aiccFetcher) extractTags(doc *goquery.Document) []models.Tag {
	var tags []models.Tag
	doc.Find(".accb-options-column").Each(func(i int, s *goquery.Selection) {
		// Get inner HTML, remove the strong tag content, extract just the values
		html, _ := s.Html()
		// Remove everything up to and including </strong>
		if idx := strings.Index(html, "</strong>"); idx != -1 {
			html = html[idx+len("</strong>"):]
		}
		// Split by comma and add each tag
		for _, t := range strings.Split(html, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, models.ResolveTag(t))
			}
		}
	})
	return tags
}

// extractName extracts the character name from the Character Name accordion section
func (f *aiccFetcher) extractName(doc *goquery.Document) string {
	var name string
	doc.Find(".brz-accordion__item").Each(func(i int, s *goquery.Selection) {
		// Check if this accordion item has "Character Name" as the title
		title := s.Find(".brz-accordion__nav-title").Text()
		if strings.TrimSpace(title) == "Character Name" {
			// Extract the name from the embed content
			name = strings.TrimSpace(s.Find(".brz-embed-content").Text())
		}
	})
	return name
}
