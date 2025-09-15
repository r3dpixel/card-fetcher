package fetcher

import (
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const (
	wyvernSourceURL string = "app.wyvern.chat"
	wyvernDirectURL string = "app.wyvern.chat/characters/"

	wyvernMainURL string = "wyvern.chat/characters/"               // Main CardURL for WyvernChat
	wyvernApiURL  string = "https://api.wyvern.chat/characters/%s" // API CardURL for WyvernChat

	wyvernSystemPromptField string = "pre_history_instructions" // System Prompt Field for WyvernChat
	wyvernTaglineField      string = "tagline"                  // Tagline Field for WyvernChat
	wyvernDateFormat        string = time.RFC3339Nano           // Date Format for WyvernChat

	wyvernBookExtensionsField string = "extensions" // Extensions field for WyvernChat Book
)

type wyvernChatFetcher struct {
	BaseFetcher
}

// NewWyvernChatFetcher - Create a new WyvernChat source
func NewWyvernChatFetcher(client *req.Client) Fetcher {
	impl := &wyvernChatFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.WyvernChat,
			sourceURL: wyvernSourceURL,
			directURL: wyvernDirectURL,
			mainURL:   wyvernMainURL,
			baseURLs:  []string{wyvernMainURL},
		},
	}

	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *wyvernChatFetcher) FetchMetadata(normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Create the API url for the card
	metadataUrl := fmt.Sprintf(wyvernApiURL, characterID)

	// Retrieve the metadata (log error is response is invalid)
	response, err := s.client.R().Get(metadataUrl)
	// Check if response is a valid JSON
	if !reqx.IsResponseOk(response, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// TaskOf the JSON string response
	metadataResponse := gjson.Parse(response.String())

	// Retrieve the real card name
	cardName := metadataResponse.Get(character.NameField).String()
	// Retrieve the character name
	name := metadataResponse.Get("chat_name").String()

	// Retrieve creator
	creator := metadataResponse.Get("creator.displayName").String()
	// Tagline for WyvernChat is an actual tagline
	tagline := strings.TrimSpace(metadataResponse.Get(wyvernTaglineField).String())
	// Parse tags
	tags := models.TagsFromJsonArray(metadataResponse.Get(character.TagsField), gjsonx.Stringifier)

	// Extract the update time and created time
	updateTime := s.fromDate(wyvernDateFormat, metadataResponse.Get("updated_at").String(), normalizedURL)
	createTime := s.fromDate(wyvernDateFormat, metadataResponse.Get("created_at").String(), normalizedURL)

	bookUpdateTime := timestamp.Nano(0)
	metadataResponse.Get("lorebooks.#.updated_at").ForEach(func(key, value gjson.Result) bool {
		bookUpdateTime = max(bookUpdateTime, s.fromDate(wyvernDateFormat, value.String(), normalizedURL))
		return true
	})

	metadata := &models.Metadata{
		Source:         s.sourceID,
		CardURL:        normalizedURL,
		PlatformID:     strings.TrimPrefix(characterID, symbols.Underscore),
		CharacterID:    characterID,
		CardName:       cardName,
		CharacterName:  name,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: bookUpdateTime,
		Tags:           tags,
	}

	fullResponse := models.JsonResponse{
		Metadata: metadataResponse,
	}
	return metadata, fullResponse, nil
}

// FetchPngCard - Retrieve card for given url
func (s *wyvernChatFetcher) FetchCharacterCard(metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata
	avatarURL := metadataResponse.Get("avatar").String()
	// Download avatar and transform to PNG
	rawCard, err := png.FromURL(s.client, avatarURL).DeepScan().Get()
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
	sheet.Data.Description = metadataResponse.Get(character.DescriptionField).String()
	// Assign the personality field
	sheet.Data.Personality = metadataResponse.Get(character.PersonalityField).String()
	// Assign the character scenario field
	sheet.Data.Scenario = metadataResponse.Get(character.ScenarioField).String()
	// Assign the first message
	sheet.Data.FirstMessage = metadataResponse.Get(character.FirstMessageField).String()
	// Assign the example dialogs
	sheet.Data.MessageExamples = metadataResponse.Get(character.MessageExamplesField).String()
	// Assemble the creator notes from tagline and original creator notes
	creatorNotes := metadataResponse.Get(character.CreatorNotesField).String()
	// Assign the assembled creator notes
	sheet.Data.CreatorNotes = stringsx.JoinNonBlank(character.CreatorNotesSeparator,
		metadata.Tagline, creatorNotes,
	)
	// Assign the system prompt field
	sheet.Data.SystemPrompt = metadataResponse.Get(wyvernSystemPromptField).String()
	// Assign the post history instruction field
	sheet.Data.PostHistoryInstructions = metadataResponse.Get(character.PostHistoryInstructionsField).String()

	// Assign the greetings
	alternateGreetings := make([]string, 0)
	// Append alternate greetings to list
	for _, greetingResult := range metadataResponse.Get(character.AlternateGreetingsField).Array() {
		alternateGreetings = append(alternateGreetings, greetingResult.String())
	}
	// Set the alternate greetings
	sheet.Data.AlternateGreetings = alternateGreetings

	// Add depth prompt if it exists (equivalent to author notes)
	prompt := metadataResponse.Get("character_note").String()
	// Add the depth prompt at depth level 4
	sheet.Data.DepthPrompt = &character.DepthPrompt{
		Prompt: prompt,
		Depth:  character.DefaultDepthPromptLevel,
	}

	// MergeTags all books (WyvernChat allows linking multiple books to a character)
	// Currently, WyvernChat does not allow book download (merging in the embedded book is the only option)
	bookMerger := character.NewBookMerger()

	// TaskOf books gJson array
	books := metadataResponse.Get("lorebooks")
	if !books.Exists() || books.Type == gjson.Null {
		return characterCard, nil
	}

	books.ForEach(func(_, value gjson.Result) bool {
		book := (*character.Book)(nil)
		bookErr := json.Unmarshal([]byte(value.String()), &book)

		if bookErr != nil || book == nil {
			log.Error().
				Err(bookErr).
				Str(trace.SOURCE, string(s.sourceID)).
				Str(trace.URL, metadata.CardURL).
				Msg("Could not parse book character")
			return true
		}

		bookMerger.AppendBook(book)
		return true
	})

	// Attach the character book to the sheet
	sheet.Data.CharacterBook = bookMerger.Build()

	// Return the parsed PNG sheet
	return characterCard, nil
}
