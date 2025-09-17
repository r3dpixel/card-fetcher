package impl

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/tidwall/gjson"
)

const (
	pephopUuidLength int = 36 // PepHop Slug length

	pephopSourceURL string = "pephop.ai"
	pephopBaseURL   string = "pephop.ai/characters/"                                          // Main NormalizedURL for PepHop
	pephopApiURL    string = "https://api.eosai.chat/characters/%s"                           // API NormalizedURL for PepHop
	pephopAvatarURL string = "https://sp.eosai.chat//storage/v1/object/public/bot-avatars/%s" // Avatar Download NormalizedURL for PepHop

	pephopFirstMessageField    string = "first_message"   // First Message Field for PepHop
	pephopMessageExamplesField string = "example_dialogs" // Message Examples Field for PepHop
	pepHopDateFormat           string = time.RFC3339Nano  // Date Format for PepHop
)

type pephopFetcher struct {
	BaseHandler
}

// NewPephopFetcher - Create a new ChubAI source
func NewPephopFetcher(client *req.Client) SourceHandler {
	impl := &pephopFetcher{
		BaseHandler: BaseHandler{
			client:    client,
			sourceID:  source.PepHop,
			sourceURL: pephopSourceURL,
			directURL: pephopBaseURL,
			mainURL:   pephopBaseURL,
			baseURLs:  []string{pephopBaseURL},
		},
	}
	return impl
}

func (s *pephopFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(pephopApiURL, characterID)
	return s.client.R().Get(metadataURL)
}

func (s *pephopFetcher) ExtractMetadata(normalizedURL string, characterID string, metadataResponse gjson.Result) (*models.CardInfo, error) {
	// Retrieve the real card name
	cardName := metadataResponse.Get(character.NameField).String()
	// Retrieve creator
	creator := metadataResponse.Get("creator_name").String()
	if stringsx.IsBlank(creator) {
		creator = character.AnonymousCreator
	}
	// Tagline for PepHop is the original creator notes (for the card creators notes we also append the short summary introduction)
	tagline := metadataResponse.Get(character.DescriptionField).String()
	// Create and parse card specific tags
	tags := models.TagsFromJsonArray(metadataResponse.Get(character.TagsField), func(result gjson.Result) string {
		return gjsonx.Stringifier(result.Get(character.NameField))
	})

	// Extract the update time and created time
	updateTime := s.fromDate(pepHopDateFormat, metadataResponse.Get("updated_at").String(), normalizedURL)
	createTime := s.fromDate(pepHopDateFormat, metadataResponse.Get("created_at").String(), normalizedURL)

	metadata := &models.CardInfo{
		Source:         s.sourceID,
		NormalizedURL:  normalizedURL,
		PlatformID:     characterID,
		CharacterID:    characterID,
		Title:          cardName,
		Name:           cardName,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: 0,
		Tags:           tags,
	}

	return metadata, nil
}

// FetchPngCard - Retrieve card for given url
func (s *pephopFetcher) FetchCharacterCard(normalizedURL string, characterID string, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata
	// Download avatar and transform to PNG
	pepHopAvatarURL := fmt.Sprintf(pephopAvatarURL, metadataResponse.Get("avatar").String())
	rawCard, err := png.FromURL(s.client, pepHopAvatarURL).DeepScan().Get()
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}
	// If the sheet is nil, assign a new sheet (which will be populated from the metadataResponse)
	if characterCard.Sheet == nil {
		characterCard.Sheet = character.EmptySheet(character.RevisionV2)
	}

	// TaskOf the characterCard sheet
	sheet := characterCard.Sheet

	// Assign the character description field
	sheet.Data.Description = metadataResponse.Get(character.PersonalityField).String()
	// Personality field is not used on PepHop
	// Assign the character scenario field
	sheet.Data.Scenario = metadataResponse.Get(character.ScenarioField).String()
	// Assign the first message
	sheet.Data.FirstMessage = metadataResponse.Get(pephopFirstMessageField).String()
	// Assign the example dialogs
	sheet.Data.MessageExamples = metadataResponse.Get(pephopMessageExamplesField).String()
	// Assemble CreatorNotes using description/introduction from the json response
	// Tagline for PepHop is the original creator notes
	// Retrieve the character introduction
	// Assign the assembled creator notes
	sheet.Data.CreatorNotes = metadataResponse.Get("introduction.characterIntroduction").String()

	// Return the parsed PNG sheet
	return characterCard, nil
}

// CharacterID - returns the characterID for pephop source
// For PepHop the suffix must be trimmed to leave just the real Slug
func (s *pephopFetcher) CharacterID(url string, matchedURL string) string {
	return s.BaseHandler.CharacterID(url, matchedURL)[0:pephopUuidLength]
}
