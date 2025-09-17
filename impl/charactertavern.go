package impl

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/structx"
	"github.com/tidwall/gjson"
)

const (
	characterTavernSourceURL  string = "character-tavern.com"
	characterTavernBaseURL    string = "character-tavern.com/character/"                                    // Main NormalizedURL for CharacterTavern
	characterTavernApiURL     string = "https://character-tavern.com/api/character/%s"                      // API NormalizedURL for CharacterTavern
	characterTavernAvatarURL  string = "https://cards.character-tavern.com/cdn-cgi/image/format=png/%s.png" // Avatar Download NormalizedURL for CharacterTavern
	characterTavernChunkRegex string = `({[\s\S]*})[\s\S]*{"type":"chunk"[\s\S]*`                           // Regex used to extract the relevant chunk of metadata from the API NormalizedURL json response

	characterTavernTaglineField string = "tagline"        // Tagline field name for CharacterTavern
	characterTavernDateFormat   string = time.RFC3339Nano // Date Format for CharacterTavern

)

type characterTavernFetcher struct {
	BaseHandler
}

// NewCharacterTavernFetcher - Create a new CharacterTavern source
func NewCharacterTavernFetcher(client *req.Client) fetcher.SourceHandler {
	impl := &characterTavernFetcher{
		BaseHandler: BaseHandler{
			client:    client,
			sourceID:  source.CharacterTavern,
			sourceURL: characterTavernSourceURL,
			directURL: characterTavernBaseURL,
			mainURL:   characterTavernBaseURL,
			baseURLs:  []string{characterTavernBaseURL},
		},
	}
	return impl
}

func (s *characterTavernFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(characterTavernApiURL, characterID)

	return s.client.R().Get(metadataURL)
}

// FetchMetadata - Retrieve metadata for given url
func (s *characterTavernFetcher) ExtractMetadata(normalizedURL string, characterID string, metadataResponse gjson.Result) (*models.CardInfo, error) {
	// Retrieve the platform Slug
	platformID := metadataResponse.Get("id").String()

	// Retrieve the real card name
	cardName := metadataResponse.Get("name").String()
	// Retrieve the character name
	name := metadataResponse.Get("inChatName").String()
	// Retrieve creator
	creator := strings.Split(metadataResponse.Get("path").String(), `/`)[0]
	// Retrieve tagline
	tagline := metadataResponse.Get("tagline").String()
	// Extract the update time and created time
	updateTime := s.fromDate(characterTavernDateFormat, metadataResponse.Get("lastUpdatedAt").String(), normalizedURL)
	createTime := s.fromDate(characterTavernDateFormat, metadataResponse.Get("createdAt").String(), normalizedURL)

	if stringsx.IsBlank(platformID) {
		return nil, s.missingPlatformIdErr(normalizedURL, nil)
	}

	metadata := &models.CardInfo{
		Source:         source.CharacterTavern,
		NormalizedURL:  normalizedURL,
		PlatformID:     platformID,
		CharacterID:    characterID,
		Title:          cardName,
		Name:           name,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: 0,
		Tags:           s.getJsonTags(metadataResponse),
	}

	return metadata, nil
}

// FetchCharacterCard - Retrieve card for given url
func (s *characterTavernFetcher) FetchCharacterCard(normalizedURL string, characterID string, response models.JsonResponse) (*png.CharacterCard, error) {
	// TaskOf the character avatar png and preserve any still existing metadata
	rawCard, err := png.FromURL(s.client, fmt.Sprintf(characterTavernAvatarURL, characterID)).DeepScan().Get()
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

	metadataResponse := response.Metadata

	sheet := characterCard.Sheet
	stringsx.UpdateIfExists(&sheet.Data.Description, metadataResponse.Get("definition_character_description").String())
	stringsx.UpdateIfExists(&sheet.Data.Personality, metadataResponse.Get("definition_personality").String())
	stringsx.UpdateIfExists(&sheet.Data.Scenario, metadataResponse.Get("definition_scenario").String())
	stringsx.UpdateIfExists(&sheet.Data.FirstMessage, metadataResponse.Get("definition_first_message").String())
	stringsx.UpdateIfExists(&sheet.Data.MessageExamples, metadataResponse.Get("definition_example_messages").String())
	stringsx.UpdateIfExists(&sheet.Data.CreatorNotes, metadataResponse.Get(character.DescriptionField).String())
	stringsx.UpdateIfExists(&sheet.Data.SystemPrompt, metadataResponse.Get("definition_system_prompt").String())
	stringsx.UpdateIfExists(&sheet.Data.PostHistoryInstructions, metadataResponse.Get("definition_post_history_prompt").String())

	// Assign the greetings
	greetings := s.getJsonAlternateGreetings(metadataResponse)
	if len(sheet.Data.AlternateGreetings) > 0 {
		// If greetings exist in the png metadata, merge all the greetings with deduplication
		greetingsList := make([]string, 0)
		greetingsMap := make(map[string]struct{})
		for _, greeting := range greetings {
			greetingsMap[greeting] = structx.Empty
		}
		for _, greeting := range sheet.Data.AlternateGreetings {
			greetingsMap[greeting] = structx.Empty
		}
		for greeting := range greetingsMap {
			greetingsList = append(greetingsList, greeting)
		}
		sheet.Data.AlternateGreetings = greetingsList
	} else {
		// Assign the greetings, since none exist in the metadata
		sheet.Data.AlternateGreetings = greetings
	}

	// Return the parsed PNG sheet
	return characterCard, nil
}

func (s *characterTavernFetcher) getCharacterJsonString(responseString string) gjson.Result {
	// Select only the current character JSON data from the response
	chunkSelectionRegex, _ := regexp.Compile(characterTavernChunkRegex)
	// Apply the chunk selection regex
	match := chunkSelectionRegex.FindString(responseString)
	// If no chunk match is found, return an empty string
	data := gjsonx.Empty
	if stringsx.IsBlank(match) {
		return data
	}
	// TaskOf the node data holding the relevant JSON character information
	nodes := gjson.Get(match, "nodes.#.data").Array()
	maxLength := -1
	// From all the data children choose the one with the maximum length
	for _, node := range nodes {
		currentData := node.String()
		currentLength := len(currentData)
		if currentLength > maxLength {
			data = node
			maxLength = currentLength
		}
	}
	// Return the relevant data
	return data
}

// Extract the tags from the JSON data
func (s *characterTavernFetcher) getJsonTags(gJsonResponse gjson.Result) []models.Tag {
	return models.TagsFromJsonArray(s.getJsonField(character.TagsField, gJsonResponse), func(result gjson.Result) string {
		return gjsonx.Stringifier(gJsonResponse.Get(fmt.Sprintf("%d", result.Int())))
	})
}

// Extract the greetings from the JSON data
func (s *characterTavernFetcher) getJsonAlternateGreetings(gJsonResponse gjson.Result) []string {
	// Initialize a greeting list
	greetings := make([]string, 0)
	// TaskOf the raw greetings field (which is an array of indices corresponding to the greetings)
	greetingsField := s.getJsonField("alternativeGreetings", gJsonResponse).Array()
	// Extract each greeting and add it to the list
	for _, greetingIndex := range greetingsField {
		greeting := gJsonResponse.Get(fmt.Sprintf("%d", greetingIndex.Int())).String()
		greetings = append(greetings, greeting)
	}
	// Return the greetings
	return greetings
}

// TaskOf the raw value of a JSON field
// CharacterTavern holds a JSON Array of field
// The first two elements are maps from field name to indices
// The rest of the elements are the values of the fields, corresponding to the index from the indices maps
// Example (an element with index 1 has the map { name: 4 } --> element with index 4 holds the name of the character)
func (s *characterTavernFetcher) getJsonField(fieldName string, gJsonResponse gjson.Result) gjson.Result {
	// Check the second map first since it holds all the information except greetings, tags, author
	filedIndex := gJsonResponse.Get(fmt.Sprintf("1.%s", fieldName)).Int()
	if filedIndex <= 1 {
		// Check the first map for any remaining information
		filedIndex = gJsonResponse.Get(fmt.Sprintf("0.%s", fieldName)).Int()
	}
	// Return the real value of the field (element with the index equal to the fieldIndex in the array)
	return gJsonResponse.Get(fmt.Sprintf("%d", filedIndex))
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
	return s.BaseHandler.CharacterID(sanitizedURL, matchedURL)
}
