package impl

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/slicesx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	characterTavernDomain       string = "character-tavern.com"                                                // Source URL for CharacterTavern
	characterTavernPath         string = "character/"                                                          // Path for CharacterTavern
	characterTavernApiURL       string = "https://character-tavern.com/api/character/%s"                       // API URL for CharacterTavern
	characterTavernTagsURL      string = "https://character-tavern.com/api/character/%s/tags"                  // Tags URL for CharacterTavern
	characterTavernGreetingsURL string = "https://character-tavern.com/api/character/%s/alternative-greetings" // Greetings URL for CharacterTavern
	characterTavernAvatarURL    string = "https://cards.character-tavern.com/%s.png?action=download"           // Avatar Download URL for CharacterTavern

	characterTavernDateFormat string = time.RFC3339Nano // Date Format for CharacterTavern

)

// CharacterTavernBuilder builder for CharacterTavern fetcher
type CharacterTavernBuilder struct{}

// Build creates a new CharacterTavern fetcher
func (b CharacterTavernBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewCharacterTavernFetcher(client)
}

// characterTavernFetcher CharacterTavern fetcher implementation
type characterTavernFetcher struct {
	BaseFetcher
}

// NewCharacterTavernFetcher creates a new CharacterTavern fetcher
func NewCharacterTavernFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(characterTavernDomain, characterTavernPath)
	impl := &characterTavernFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.CharacterTavern,
			sourceURL: characterTavernDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: characterTavernDomain, Path: characterTavernPath},
			},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *characterTavernFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(characterTavernApiURL, characterID)
	return f.client.R().Get(metadataURL)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *characterTavernFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "card", "path")
}

// FetchCardInfo fetches the card info from the source
func (f *characterTavernFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Fetch the card node
	cardNode := metadataBinder.Get("card")
	// Extract platformID
	platformID := cardNode.Get("id").String()
	// Fetch tags
	tagResponse, err := reqx.String(f.client.R().Get(fmt.Sprintf(characterTavernTagsURL, platformID)))
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchMetadataErr)
	}
	// Parse tags
	tagNode, err := sonicx.GetFromString(tagResponse)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedMetadataErr)
	}
	// Resolve tags
	resolvedTags := models.TagsFromJsonArray(tagNode, sonicx.WrapString)

	// Return the parsed card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    platformID,
		CharacterID:   metadataBinder.CharacterID,
		Name:          cardNode.Get("inChatName").String(),
		Title:         cardNode.Get("name").String(),
		Tagline:       cardNode.Get("tagline").String(),
		CreateTime:    timestamp.ParseF(characterTavernDateFormat, cardNode.Get("createdAt").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(characterTavernDateFormat, cardNode.Get("lastUpdatedAt").String(), trace.URL, metadataBinder.NormalizedURL),
		IsForked:      false,
		Tags:          resolvedTags,
	}, nil
}

// FetchCreatorInfo retrieves the creator info for the given metadata binder
func (f *characterTavernFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	// Extract the display name from the path
	displayName := strings.Split(metadataBinder.GetByPath("card", "path").String(), `/`)[0]
	// Return the creator info
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: metadataBinder.Get("ownerCTId").String(),
	}, nil
}

// FetchCharacterCard retrieves card for given url
func (f *characterTavernFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Fetch avatar
	rawCard, err := png.FromURL(f.client, fmt.Sprintf(characterTavernAvatarURL, binder.CharacterID)).LastLongest().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Extract card node
	cardNode := binder.Get("card")

	// Update character sheet fields
	sheet := characterCard.Sheet
	sheet.Description.SetIf(cardNode.Get("definition_character_description").String())
	sheet.Personality.SetIf(cardNode.Get("definition_personality").String())
	sheet.Scenario.SetIf(cardNode.Get("definition_scenario").String())
	sheet.FirstMessage.SetIf(cardNode.Get("definition_first_message").String())
	sheet.MessageExamples.SetIf(cardNode.Get("definition_example_messages").String())
	sheet.CreatorNotes.SetIf(cardNode.Get("description").String())
	sheet.SystemPrompt.SetIf(cardNode.Get("definition_system_prompt").String())
	sheet.PostHistoryInstructions.SetIf(cardNode.Get("definition_post_history_prompt").String())

	// Fetch greetings
	greetingsResponse, err := reqx.String(f.client.R().Get(fmt.Sprintf(characterTavernGreetingsURL, cardNode.Get("id").String())))
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Parse greetings
	var greetings property.StringArray
	if err := sonicx.Config.UnmarshalFromString(greetingsResponse, &greetings); err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Update alternate greetings
	sheet.AlternateGreetings = slicesx.DeduplicateStable(greetings, sheet.AlternateGreetings)

	// Return the parsed PNG sheet
	return characterCard, nil
}

// CharacterID overrides the GetCharacterID behavior to account for allowed spaces in the URL
// CharacterTavern allows spaces in the URL (why???)
func (f *characterTavernFetcher) CharacterID(rawCharacterID string) string {
	// Unescape ID if needed
	unescapedID, err := url.QueryUnescape(rawCharacterID)
	if err != nil {
		// Return the raw ID if unescaping fails
		return rawCharacterID
	}

	// Return the unescaped ID
	return unescapedID
}
