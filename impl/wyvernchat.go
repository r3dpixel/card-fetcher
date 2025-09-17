package impl

import (
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const (
	wyvernSourceURL string = "app.wyvern.chat"
	wyvernDirectURL string = "app.wyvern.chat/characters/"

	wyvernMainURL string = "wyvern.chat/characters/"               // Main NormalizedURL for WyvernChat
	wyvernApiURL  string = "https://api.wyvern.chat/characters/%s" // API NormalizedURL for WyvernChat

	wyvernTaglineField string = "tagline"        // Tagline Field for WyvernChat
	wyvernDateFormat   string = time.RFC3339Nano // Date Format for WyvernChat

	wyvernBookExtensionsField string = "extensions" // Extensions field for WyvernChat Book
)

type wyvernChatFetcher struct {
	BaseHandler
}

// WyvernChatHandler - Create a new WyvernChat source
func WyvernChatHandler(client *req.Client) fetcher.SourceHandler {
	impl := &wyvernChatFetcher{
		BaseHandler: BaseHandler{
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

func (s *wyvernChatFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataUrl := fmt.Sprintf(wyvernApiURL, characterID)
	return s.client.R().Get(metadataUrl)
}

func (s *wyvernChatFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Retrieve the real card name
	cardName := metadataBinder.Get(character.NameField).String()
	// Retrieve the character name
	name := metadataBinder.Get("chat_name").String()

	// Tagline for WyvernChat is an actual tagline
	tagline := strings.TrimSpace(metadataBinder.Get(wyvernTaglineField).String())
	// Parse tags
	tags := models.TagsFromJsonArray(metadataBinder.Get(character.TagsField), gjsonx.Stringifier)

	// Extract the update time and created time
	updateTime := s.fromDate(wyvernDateFormat, metadataBinder.Get("updated_at").String(), metadataBinder.NormalizedURL)
	createTime := s.fromDate(wyvernDateFormat, metadataBinder.Get("created_at").String(), metadataBinder.NormalizedURL)

	metadata := &models.CardInfo{
		Source:        s.sourceID,
		NormalizedURL: metadataBinder.NormalizedURL,
		PlatformID:    strings.TrimPrefix(metadataBinder.CharacterID, symbols.Underscore),
		CharacterID:   metadataBinder.CharacterID,
		Title:         cardName,
		Name:          name,
		Tagline:       tagline,
		CreateTime:    createTime,
		UpdateTime:    updateTime,
		Tags:          tags,
	}

	return metadata, nil
}

func (s *wyvernChatFetcher) CreateBinder(characterID string, normalizedURL string, metadataResponse gjson.Result) (*fetcher.MetadataBinder, error) {
	updatedCharacterID := metadataResponse.Get("id").String()
	return s.BaseHandler.CreateBinder(
		updatedCharacterID,
		s.NormalizeURL(updatedCharacterID),
		metadataResponse,
	)
}

func (s *wyvernChatFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	displayName := metadataBinder.Get("creator.displayName").String()
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: metadataBinder.Get("creator.id").String(),
	}, nil
}

func (s *wyvernChatFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	bookUpdateTime := timestamp.Nano(0)
	metadataBinder.Get("lorebooks.#.updated_at").ForEach(func(key, value gjson.Result) bool {
		bookUpdateTime = max(bookUpdateTime, s.fromDate(wyvernDateFormat, value.String(), metadataBinder.NormalizedURL))
		return true
	})
	return &fetcher.BookBinder{
		UpdateTime: bookUpdateTime,
	}, nil
}

func (s *wyvernChatFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	avatarURL := binder.Get("avatar").String()

	rawCard, err := png.FromURL(s.client, avatarURL).DeepScan().Get()
	if err != nil {
		return nil, err
	}

	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}
	sheet := characterCard.Sheet

	sheet.Content.Description = binder.Get(character.DescriptionField).String()
	sheet.Content.Personality = binder.Get(character.PersonalityField).String()
	sheet.Content.Scenario = binder.Get(character.ScenarioField).String()
	sheet.Content.FirstMessage = binder.Get(character.FirstMessageField).String()
	sheet.Content.MessageExamples = binder.Get(character.MessageExamplesField).String()
	sheet.Content.CreatorNotes = binder.Get(character.CreatorNotesField).String()
	sheet.Content.SystemPrompt = binder.Get("pre_history_instructions").String()
	sheet.Content.PostHistoryInstructions = binder.Get(character.PostHistoryInstructionsField).String()

	alternateGreetings := make([]string, 0)
	for _, greetingResult := range binder.Get(character.AlternateGreetingsField).Array() {
		alternateGreetings = append(alternateGreetings, greetingResult.String())
	}
	sheet.Content.AlternateGreetings = alternateGreetings

	prompt := binder.Get("character_note").String()
	sheet.Content.DepthPrompt = &character.DepthPrompt{
		Prompt: prompt,
		Depth:  character.DefaultDepthPromptLevel,
	}

	bookMerger := character.NewBookMerger()

	books := binder.Get("lorebooks")
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
				Str(trace.URL, binder.NormalizedURL).
				Msg("Could not parse book character")
			return true
		}

		bookMerger.AppendBook(book)
		return true
	})

	sheet.Content.CharacterBook = bookMerger.Build()

	return characterCard, nil
}
