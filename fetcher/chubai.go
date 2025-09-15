package fetcher

import (
	"encoding/json/v2"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/properties"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/structx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const (
	chubSourceURL          string = "chub.ai"
	chubMainURL            string = "chub.ai/characters/"                                     // Main CardURL for ChubAI
	chubAlternateURL       string = "characterhub.org/characters/"                            // Mirror CardURL for ChubAI
	chubApiURL             string = "https://api.chub.ai/api/characters/%s?full=true"         // Public API for retrieving metadata
	chubApiCardDownloadURL string = "https://avatars.charhub.io/avatars/%s/chara_card_v2.png" // Download CardURL for retrieving card
	chubApiBookURL         string = "https://api.chub.ai/api/lorebooks/%s?full=true"          // Public API for retrieving books

	chubAiTaglineFieldName string = "node.tagline"   // Tagline field name for ChubAI
	chubAiDateFormat       string = time.RFC3339Nano // Date Format for ChubAI API

	chubCharaPath = "chara_char_v2" + png.Extension
	chubCardPath  = "chara_card_v2" + png.Extension
)

var (
	bookRegexp = regexp.MustCompile(`lorebooks/([^"\s<>()]+)`)
)

type chubAIFetcher struct {
	BaseFetcher
}

// NewChubAIFetcher - Create a new ChubAI source
func NewChubAIFetcher(client *req.Client) Fetcher {
	impl := &chubAIFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.ChubAI,
			sourceURL: chubSourceURL,
			directURL: chubMainURL,
			mainURL:   chubMainURL,
			baseURLs:  []string{chubMainURL, chubAlternateURL},
		},
	}
	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *chubAIFetcher) FetchMetadata(normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Create the API url for retrieving the metadata
	metadataURL := fmt.Sprintf(chubApiURL, characterID)

	// Retrieve the metadata (log error is response is invalid)
	response, err := s.client.R().Get(metadataURL)
	// Check if the response is a valid JSON
	if !reqx.IsResponseOk(response, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// TaskOf the JSON string response
	metadataResponse := gjson.Parse(response.String())

	// Retrieve the updated characterID (ChubAI allows creators to change username)
	characterID = metadataResponse.Get("node.fullPath").String()
	normalizedURL = s.NormalizeURL(characterID)

	// Retrieve the real card name
	cardName := metadataResponse.Get("node.name").String()
	// Retrieve the character name
	name := metadataResponse.Get("node.definition.name").String()

	// For ChubAI characterID is "creator/CardName"
	creator := strings.Split(characterID, `/`)[0]
	// Tagline for ChubAI is an actual tagline
	tagline := metadataResponse.Get(chubAiTaglineFieldName).String()
	// Create and parse card specific tags
	tags := models.TagsFromJsonArray(metadataResponse.Get("node.topics"), gjsonx.Stringifier)

	// Extract the update time and created time
	updateTime := s.fromDate(chubAiDateFormat, metadataResponse.Get("node.lastActivityAt").String(), normalizedURL)
	createTime := s.fromDate(chubAiDateFormat, metadataResponse.Get("node.createdAt").String(), normalizedURL)

	// ChubAI "sometimes" doesn't use the latest version of the linked books,
	// which means fixing it by manually merging books...
	bookIDs := gjsonx.ArrayToMap(
		metadataResponse.Get("node.related_lorebooks"),
		func(token string) bool {
			intToken, tokenErr := strconv.Atoi(token)
			return tokenErr != nil || (tokenErr == nil && intToken >= 0)
		},
		gjsonx.Stringifier,
	)

	linkedBookResponses, linkedBookUpdateTime := s.retrieveLinkedBooks(metadataURL, bookIDs)
	auxBookResponses, auxBookUpdateTime := s.retrieveAuxBooks(metadataURL, metadataResponse, bookIDs)

	metadata := &models.Metadata{
		Source:         s.sourceID,
		CardURL:        normalizedURL,
		PlatformID:     metadataResponse.Get("node.id").String(),
		CharacterID:    characterID,
		CardName:       cardName,
		CharacterName:  name,
		Tags:           tags,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: max(linkedBookUpdateTime, auxBookUpdateTime),
	}
	fullResponse := models.JsonResponse{
		Metadata:         metadataResponse,
		BookResponses:    linkedBookResponses,
		AuxBookResponses: auxBookResponses,
	}
	return metadata, fullResponse, nil
}

// FetchPngCard - Retrieve card for given url
func (s *chubAIFetcher) FetchCharacterCard(metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata
	chubCardURL := metadataResponse.Get("node.max_res_url").String()
	backupURL := metadataResponse.Get("node.avatar_url").String()

	characterCard, err := s.retrieveCardData(chubCardURL, backupURL)
	if err != nil {
		return nil, err
	}
	if characterCard.Sheet == nil {
		characterCard.Sheet = character.EmptySheet(character.RevisionV2)
	}
	sheet := characterCard.Sheet

	s.updateFieldsWithFallback(&sheet.Data, response.Metadata, metadata.CardURL)

	// Assemble CreatorNotes using any creator notes in the downloaded card,
	// and any description/tagline from the JSON response
	// Append the tagline to the original creator notes
	sheet.Data.CreatorNotes = stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		metadata.Tagline, sheet.Data.CreatorNotes,
	)

	// If there are no related books, no processing needed
	if response.BookCount() == 0 {
		// Return the parsed PNG card
		return characterCard, nil
	}

	merger := character.NewBookMerger()

	embeddedBook := s.parseBookGJson(
		metadataResponse,
		metadata.CardURL,
		false,
	)

	if embeddedBook != nil {
		if stringsx.IsBlankPtr(embeddedBook.Name) {
			embeddedBook.Name = new(string)
			*embeddedBook.Name = character.BookNamePlaceholder
		}
		merger.AppendBook(embeddedBook)
	}
	for _, bookResponse := range response.BookResponses {
		book := s.parseBookGJson(bookResponse, metadata.CardURL, true)
		if book != nil {
			merger.AppendBook(book)
		}
	}
	for _, bookResponse := range response.AuxBookResponses {
		book := s.parseBookGJson(bookResponse, metadata.CardURL, true)
		if book != nil && (embeddedBook == nil || *embeddedBook.Name != *book.Name) {
			merger.AppendBook(book)
		}
	}

	sheet.Data.CharacterBook = merger.Build()

	return characterCard, nil
}

func (s *chubAIFetcher) updateFieldsWithFallback(data *character.Data, metadataResponse gjson.Result, url string) {
	stringsx.UpdateIfExists(&data.Description, metadataResponse.Get("node.definition.personality").String())
	stringsx.UpdateIfExists(&data.Personality, metadataResponse.Get("node.definition.tavern_personality").String())
	stringsx.UpdateIfExists(&data.Scenario, metadataResponse.Get("node.definition.scenario").String())
	stringsx.UpdateIfExists(&data.FirstMessage, metadataResponse.Get("node.definition.first_message").String())
	stringsx.UpdateIfExists(&data.MessageExamples, metadataResponse.Get("node.definition.example_dialogs").String())
	stringsx.UpdateIfExists(&data.CreatorNotes, metadataResponse.Get("node.definition.description").String())
	stringsx.UpdateIfExists(&data.SystemPrompt, metadataResponse.Get("node.definition.system_prompt").String())
	stringsx.UpdateIfExists(&data.PostHistoryInstructions, metadataResponse.Get("node.definition.post_history_instructions").String())
	var alternateGreetings properties.StringArray
	err := json.Unmarshal([]byte(metadataResponse.Get("node.definition.alternate_greetings").String()), &alternateGreetings)
	if err != nil {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, url).
			Msg("failed to unmarshal alternate greetings")
	}
	if len(alternateGreetings) > 0 {
		data.AlternateGreetings = alternateGreetings
	}
}

func (s *chubAIFetcher) retrieveCardData(cardURL string, backupURL string) (*png.CharacterCard, error) {
	rawCard, err := png.FromURL(s.client, cardURL).DeepScan().Get()
	if err != nil {
		rawCard, err = png.FromURL(s.client, s.fixAvatarURL(cardURL)).DeepScan().Get()
	}
	if err != nil {
		rawCard, err = png.FromURL(s.client, backupURL).DeepScan().Get()
	}
	if err != nil {
		return nil, err
	}

	return rawCard.Decode()
}

func (s *chubAIFetcher) retrieveLinkedBooks(metadataURL string, bookIDs *orderedmap.OrderedMap[string, struct{}]) ([]gjson.Result, timestamp.Nano) {
	var bookResponses []gjson.Result
	maxBookUpdateTime := timestamp.Nano(0)
	for bookID := range bookIDs.Keys() {
		bookGJson, bookUpdateTime, found := s.retrieveBookData(bookID, metadataURL)
		if found {
			bookResponses = append(bookResponses, bookGJson)
			maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
		}
	}

	return bookResponses, maxBookUpdateTime
}

func (s *chubAIFetcher) retrieveAuxBooks(
	metadataURL string,
	metadataResponse gjson.Result,
	bookIDs *orderedmap.OrderedMap[string, struct{}],
) ([]gjson.Result, timestamp.Nano) {
	var auxBookResponses []gjson.Result
	maxBookUpdateTime := timestamp.Nano(0)
	auxSources := metadataResponse.Get("node.description").String() + symbols.Space + metadataResponse.Get("node.tagline").String()
	bookURLs := bookRegexp.FindAllStringSubmatch(auxSources, -1)
	for _, bookURLMatches := range bookURLs {
		if len(bookURLMatches) <= 1 {
			continue
		}
		bookPath := bookURLMatches[1]
		for len(bookPath) > 0 {
			bookGJson, bookUpdateTime, found := s.retrieveBookData(bookPath, metadataURL)
			if found {
				bookID := strings.TrimSpace(bookGJson.Get("node.id").String())
				if !bookIDs.Has(bookID) {
					bookIDs.Set(bookID, structx.Empty)
					auxBookResponses = append(auxBookResponses, bookGJson)
					maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
				}
				break
			}

			lastSlash := strings.LastIndex(bookPath, symbols.Slash)
			lastSlash = max(lastSlash, 0)
			bookPath = bookPath[:lastSlash]
		}
	}

	return auxBookResponses, maxBookUpdateTime
}

func (s *chubAIFetcher) retrieveBookData(bookID string, url string) (gjson.Result, timestamp.Nano, bool) {
	// Retrieve the book data
	response, err := s.client.R().
		SetContentType(reqx.JsonApplicationContentType).
		Get(fmt.Sprintf(chubApiBookURL, bookID))
	if !reqx.IsResponseOk(response, err) {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, url).
			Str("bookID", bookID).
			Msg("Lorebook unlinked/deleted")
		return gjsonx.Empty, 0, false
	}
	gJson := gjson.Parse(response.String())
	updateTime := s.fromDate(chubAiDateFormat, gJson.Get("node.lastActivityAt").String(), url)
	return gJson, updateTime, true
}

func (s *chubAIFetcher) parseBookGJson(fullResponse gjson.Result, url string, overrideName bool) *character.Book {
	content := fullResponse.Get("node.definition.embedded_lorebook")
	if !content.Exists() || content.Type == gjson.Null {
		return nil
	}
	rawJson := content.String()
	if stringsx.IsBlank(rawJson) {
		return nil
	}

	book := (*character.Book)(nil)
	err := json.Unmarshal([]byte(rawJson), &book)
	if err != nil {
		log.Error().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, url).
			Msg("Could not parse book")
	}

	if overrideName {
		book.Name = new(string)
		*book.Name = fullResponse.Get("node.name").String()
	}

	return book
}

// getChubIdentifier - Return the chub specific identifier (based on characterID)
func (s *chubAIFetcher) getChubIdentifier(characterID string) string {
	// Find the last index of the '-' character
	dashIndex := strings.LastIndex(characterID, symbols.Dash)
	// If there is no '-', there is no identifier
	if dashIndex == -1 {
		return stringsx.Empty
	}
	// If the last '-' is the last character, there is no identifier
	if dashIndex == len(characterID)-1 {
		return stringsx.Empty
	}
	// Substring of the content before the last '-'
	path := strings.ToLower(characterID[0:dashIndex])
	// Substring of the content after the last '-'
	identifier := characterID[dashIndex+1:]
	// If somehow the identifier is contained inside the existing path, there is no actual identifier
	if strings.Contains(path, identifier) {
		return stringsx.Empty
	}
	// Return the identifier
	return identifier
}

// fixAvatarURL - corrects the chub avatar CardURL in case it has the wrong path (replaces chara_char_v2 with chara_card_v2)
func (s *chubAIFetcher) fixAvatarURL(avatarURL string) string {
	avatarURL = strings.TrimSuffix(avatarURL, chubCharaPath)
	avatarURL = avatarURL + chubCardPath
	return avatarURL
}
