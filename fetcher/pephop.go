package fetcher

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/tidwall/gjson"
)

const (
	pephopUuidLength int = 36 // PepHop Slug length

	pephopURL       string = "pephop.ai"
	pephopBaseURL   string = "pephop.ai/characters/"                                          // Main CardURL for PepHop
	pephopApiURL    string = "https://api.eosai.chat/characters/%s"                           // API CardURL for PepHop
	pephopAvatarURL string = "https://sp.eosai.chat//storage/v1/object/public/bot-avatars/%s" // Avatar Download CardURL for PepHop

	pephopFirstMessageField    string = "first_message"   // First Message Field for PepHop
	pephopMessageExamplesField string = "example_dialogs" // Message Examples Field for PepHop
	pepHopDateFormat           string = time.RFC3339Nano  // Date Format for PepHop
)

type pephopFetcher struct {
	BaseFetcher
}

// NewPephopFetcher - Create a new ChubAI source
func NewPephopFetcher() Fetcher {
	impl := &pephopFetcher{
		BaseFetcher: BaseFetcher{
			sourceID:  source.PepHop,
			sourceURL: pephopURL,
			directURL: pephopBaseURL,
			baseURLs:  []string{pephopBaseURL},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *pephopFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Create the API url for retrieving the metadata
	metadataURL := fmt.Sprintf(pephopApiURL, characterID)

	// Retrieve the metadata (log error is response is invalid)
	response, err := c.R().Get(metadataURL)
	// Check if the response is a valid JSON
	if !reqx.IsResponseOk(response, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// TaskOf the JSON string response
	metadataResponse := gjson.Parse(response.String())

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

	metadata := &models.Metadata{
		Source:         s.sourceID,
		CardURL:        normalizedURL,
		PlatformID:     characterID,
		CharacterID:    characterID,
		CardName:       cardName,
		CharacterName:  cardName,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: 0,
		Tags:           tags,
	}

	fullResponse := models.JsonResponse{
		Metadata: metadataResponse,
	}
	return metadata, fullResponse, nil
}

// FetchPngCard - Retrieve card for given url
func (s *pephopFetcher) FetchCharacterCard(c *req.Client, metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata
	// Download avatar and transform to PNG
	pepHopAvatarURL := fmt.Sprintf(pephopAvatarURL, metadataResponse.Get("avatar").String())
	rawCard, err := png.FromURL(c, pepHopAvatarURL).DeepScan().Get()
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
	creatorNotesIntroduction := metadataResponse.Get("introduction.characterIntroduction").String()
	// Assign the assembled creator notes
	sheet.Data.CreatorNotes = stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		metadata.Tagline, creatorNotesIntroduction,
	)

	// Return the parsed PNG sheet
	return characterCard, nil
}

// CharacterID - returns the characterID for pephop source
// For PepHop the suffix must be trimmed to leave just the real Slug
func (s *pephopFetcher) CharacterID(url string, matchedURL string) string {
	return s.BaseFetcher.CharacterID(url, matchedURL)[0:pephopUuidLength]
}
