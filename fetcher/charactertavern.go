package fetcher

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/structx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/tidwall/gjson"
)

const (
	characterTavernURL        string = "character-tavern.com"
	characterTavernBaseURL    string = "character-tavern.com/character/"                                    // Main CardURL for CharacterTavern
	characterTavernApiURL     string = "https://character-tavern.com/character/%s/__data.json"              // API CardURL for CharacterTavern
	characterTavernAvatarURL  string = "https://cards.character-tavern.com/cdn-cgi/image/format=png/%s.png" // Avatar Download CardURL for CharacterTavern
	characterTavernChunkRegex string = `({[\s\S]*})[\s\S]*{"type":"chunk"[\s\S]*`                           // Regex used to extract the relevant chunk of metadata from the API CardURL json response

	characterTavernTaglineField string = "tagline"        // Tagline field name for CharacterTavern
	characterTavernDateFormat   string = time.RFC3339Nano // Date Format for CharacterTavern

)

type characterTavernFetcher struct {
	BaseFetcher
}

// NewCharacterTavernFetcher - Create a new CharacterTavern source
func NewCharacterTavernFetcher() Fetcher {
	impl := &characterTavernFetcher{
		BaseFetcher: BaseFetcher{
			sourceID:  source.CharacterTavern,
			sourceURL: characterTavernURL,
			directURL: characterTavernBaseURL,
			baseURLs:  []string{characterTavernBaseURL},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *characterTavernFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Create the API url for the card
	metadataURL := fmt.Sprintf(characterTavernApiURL, characterID)

	// Retrieve the metadata (log error is response is invalid)
	response, err := c.R().Get(metadataURL)
	// Check if response is a valid JSON
	if !reqx.IsResponseOk(response, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// Retrieve creator
	metadataResponse := s.getCharacterJsonString(response.String())

	// Retrieve the platform Slug
	platformID := s.getJsonField("id", metadataResponse).String()

	// Retrieve the real card name
	cardName := s.getJsonField(character.NameField, metadataResponse).String()
	// Retrieve the character name
	name := s.getJsonField("inChatName", metadataResponse).String()
	// Retrieve creator
	creator := s.getJsonFieldFromFirstMap("author", metadataResponse).String()
	// Retrieve tagline
	tagline := s.getJsonField(characterTavernTaglineField, metadataResponse).String()
	// Extract the update time and created time
	updateTime := s.getJsonDate("lastUpdatedAt", metadataResponse, normalizedURL)
	createTime := s.getJsonDate("createdAt", metadataResponse, normalizedURL)

	if stringsx.IsBlank(platformID) {
		return nil, models.EmptyJsonResponse, s.missingPlatformIdErr(normalizedURL, nil)
	}

	metadata := &models.Metadata{
		Source:         source.CharacterTavern,
		CardURL:        normalizedURL,
		PlatformID:     platformID,
		CharacterID:    characterID,
		CardName:       cardName,
		CharacterName:  name,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: 0,
		Tags:           s.getJsonTags(metadataResponse),
	}
	fullResponse := models.JsonResponse{
		Metadata: metadataResponse,
	}
	return metadata, fullResponse, nil
}

// FetchCharacterCard - Retrieve card for given url
func (s *characterTavernFetcher) FetchCharacterCard(c *req.Client, metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	// TaskOf the character avatar png and preserve any still existing metadata
	rawCard, err := png.FromURL(c, fmt.Sprintf(characterTavernAvatarURL, metadata.CharacterID)).DeepScan().Get()
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
	// Assign the character description field
	sheet.Data.Description = s.getJsonField("definition_character_description", metadataResponse).String()
	// Assign the personality field
	sheet.Data.Personality = s.getJsonField("definition_personality", metadataResponse).String()
	// Assign the character scenario field
	sheet.Data.Scenario = s.getJsonField("definition_scenario", metadataResponse).String()
	// Assign the first message
	sheet.Data.FirstMessage = s.getJsonField("definition_first_message", metadataResponse).String()
	// Assign the example dialogs
	sheet.Data.MessageExamples = s.getJsonField("definition_example_messages", metadataResponse).String()
	// Assembled the creator notes from the tagline and introduction
	// Introduction for CharacterTavern is description
	introduction := s.getJsonField(character.DescriptionField, metadataResponse).String()
	// Assign the assembled creator notes
	sheet.Data.CreatorNotes = stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		metadata.Tagline, introduction,
	)

	// Assign any existing system prompt only if it is not empty
	// CharacterTavern does not support system prompt, but if there is one in the png metadata, preserve it
	systemPrompt := s.getJsonField("definition_system_prompt", metadataResponse).String()
	if stringsx.IsNotBlank(systemPrompt) {
		sheet.Data.SystemPrompt = systemPrompt
	}
	// Assign the post history instruction field
	sheet.Data.PostHistoryInstructions = s.getJsonField("definition_post_history_prompt", metadataResponse).String()

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

// Extract a date in timestamp unix format from the json data
func (s *characterTavernFetcher) getJsonDate(fieldName string, gJsonResponse gjson.Result, url string) timestamp.Nano {
	// TaskOf the date field (which is an array holding two elements)
	dateResult := s.getJsonField(fieldName, gJsonResponse)
	// Iterate through the elements
	for _, item := range dateResult.Array() {
		// If the element is indeed a date string in the RFC3339 format, transform it into a timestamp and return
		date := item.String()
		if symbols.RFC3339NanoRegExp.MatchString(date) {
			return s.fromDate(characterTavernDateFormat, date, url)
		}
	}
	return 0
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

func (s *characterTavernFetcher) getJsonFieldFromFirstMap(fieldName string, gJsonResponse gjson.Result) gjson.Result {
	// Check the first map for any information
	filedIndex := gJsonResponse.Get(fmt.Sprintf("0.%s", fieldName)).Int()
	// Return the real value of the field (element with the index equal to the fieldIndex in the array)
	return gJsonResponse.Get(fmt.Sprintf("%d", filedIndex))
}

// CharacterID - override the GetCharacterID behavior to account for allowed spaces in the CardURL
// CharacterTavern allows spaces in the CardURL (why???)
func (s *characterTavernFetcher) CharacterID(cardURL string, matchedURL string) string {
	// Unescape CardURL if needed
	sanitizedURL, err := url.QueryUnescape(cardURL)
	if err != nil {
		sanitizedURL = cardURL
	}

	// Extract characterID
	return s.BaseFetcher.CharacterID(sanitizedURL, matchedURL)
}
