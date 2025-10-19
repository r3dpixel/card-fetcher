package impl

import (
	"fmt"
	"net/url"
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
	characterTavernSourceURL    string = "character-tavern.com"
	characterTavernBaseURL      string = "character-tavern.com/character/"                                     // Main URL for CharacterTavern
	characterTavernApiURL       string = "https://character-tavern.com/api/character/%s"                       // API URL for CharacterTavern
	characterTavernTagsURL      string = "https://character-tavern.com/api/character/%s/tags"                  // Tags URL for CharacterTavern
	characterTavernGreetingsURL string = "https://character-tavern.com/api/character/%s/alternative-greetings" // Greetings URL for CharacterTavern
	characterTavernAvatarURL    string = "https://cards.character-tavern.com/cdn-cgi/image/format=png/%s.png"  // Avatar Download URL for CharacterTavern

	characterTavernTaglineField string = "tagline"        // Tagline field name for CharacterTavern
	characterTavernDateFormat   string = time.RFC3339Nano // Date Format for CharacterTavern

)

type CharacterTavernBuilder struct{}

func (b CharacterTavernBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewCharacterTavernFetcher(client)
}

type characterTavernFetcher struct {
	BaseFetcher
}

// NewCharacterTavernFetcher - Create a new CharacterTavern source
func NewCharacterTavernFetcher(client *reqx.Client) fetcher.Fetcher {
	impl := &characterTavernFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.CharacterTavern,
			sourceURL: characterTavernSourceURL,
			directURL: characterTavernBaseURL,
			mainURL:   characterTavernBaseURL,
			baseURLs:  []string{characterTavernBaseURL},
		},
	}
	impl.Extends(impl)
	return impl
}

func (s *characterTavernFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(characterTavernApiURL, characterID)

	return s.client.R().Get(metadataURL)
}

func (s *characterTavernFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return s.BaseFetcher.CreateBinder(metadataResponse.GetByPath("card", "path").String(), metadataResponse)
}

func (s *characterTavernFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	cardNode := metadataBinder.Get("card")
	platformID := cardNode.Get("id").String()
	tagResponse, err := reqx.String(s.client.R().Get(fmt.Sprintf(characterTavernTagsURL, platformID)))
	if err != nil {
		return nil, err
	}

	tagNode, err := sonicx.GetFromString(tagResponse)
	if err != nil {
		return nil, err
	}
	resolvedTags := models.TagsFromJsonArray(tagNode, sonicx.WrapString)

	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     s.DirectURL(metadataBinder.CharacterID),
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

func (s *characterTavernFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	displayName := strings.Split(metadataBinder.GetByPath("card", "path").String(), `/`)[0]
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: metadataBinder.Get("ownerCTId").String(),
	}, nil
}

// FetchCharacterCard - Retrieve card for given url
func (s *characterTavernFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	rawCard, err := png.FromURL(s.client, fmt.Sprintf(characterTavernAvatarURL, binder.CharacterID)).LastLongest().Get()
	if err != nil {
		return nil, err
	}

	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	cardNode := binder.Get("card")

	sheet := characterCard.Sheet
	sheet.Description.SetIf(cardNode.Get("definition_character_description").String())
	sheet.Personality.SetIf(cardNode.Get("definition_personality").String())
	sheet.Scenario.SetIf(cardNode.Get("definition_scenario").String())
	sheet.FirstMessage.SetIf(cardNode.Get("definition_first_message").String())
	sheet.MessageExamples.SetIf(cardNode.Get("definition_example_messages").String())
	sheet.CreatorNotes.SetIf(cardNode.Get("description").String())
	sheet.SystemPrompt.SetIf(cardNode.Get("definition_system_prompt").String())
	sheet.PostHistoryInstructions.SetIf(cardNode.Get("definition_post_history_prompt").String())

	platformID := cardNode.Get("id").String()
	greetingsResponse, err := reqx.String(s.client.R().Get(fmt.Sprintf(characterTavernGreetingsURL, platformID)))
	if err != nil {
		return nil, err
	}

	var greetings property.StringArray
	if err := sonicx.Config.UnmarshalFromString(greetingsResponse, &greetings); err != nil {
		return nil, err
	}

	sheet.AlternateGreetings = slicesx.MergeStable(greetings, sheet.AlternateGreetings)

	// Return the parsed PNG sheet
	return characterCard, nil
}

// CharacterID - override the GetCharacterID behavior to account for allowed spaces in the NormalizedURL
// CharacterTavern allows spaces in the NormalizedURL (why???)
func (s *characterTavernFetcher) CharacterID(cardURL string, matchedURL string) string {
	// Unescape NormalizedURL if needed
	sanitizedURL, err := url.QueryUnescape(cardURL)
	if err != nil {
		sanitizedURL = cardURL
	}

	// Extract characterID
	return s.BaseFetcher.CharacterID(sanitizedURL, matchedURL)
}
