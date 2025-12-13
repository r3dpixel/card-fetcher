package impl

import (
	"fmt"
	"path"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
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

// JannyAIBuilder builder for JannyAI fetcher
type JannyAIBuilder struct{}

// Build creates a new JannyAI fetcher
func (b JannyAIBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewJannyAIFetcher(client)
}

// jannyAIFetcher JannyAI fetcher implementation
type jannyAIFetcher struct {
	BaseFetcher
}

// NewJannyAIFetcher creates a new JannyAI fetcher
func NewJannyAIFetcher(client *reqx.Client) fetcher.Fetcher {
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
	return f.client.R().Get(url)
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
	_, err := f.client.R().Get(url)
	return err
}
