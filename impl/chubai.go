package impl

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/slicesx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/structx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

const (
	chubSourceURL          string = "chub.ai"
	chubMainURL            string = "chub.ai/characters/"                                     // Main NormalizedURL for ChubAI
	chubAlternateURL       string = "characterhub.org/characters/"                            // Mirror NormalizedURL for ChubAI
	chubApiURL             string = "https://api.chub.ai/api/characters/%s?full=true"         // Public API for retrieving metadata
	chubApiCardDownloadURL string = "https://avatars.charhub.io/avatars/%s/chara_card_v2.png" // Download NormalizedURL for retrieving card
	chubApiBookURL         string = "https://api.chub.ai/api/lorebooks/%s?full=true"          // Public API for retrieving books
	chubApiUsersURL        string = "https://api.chub.ai/api/users/%s"

	chubAiTaglineFieldName string = "node.tagline"   // Tagline field name for ChubAI
	chubAiDateFormat       string = time.RFC3339Nano // Date Format for ChubAI API

	chubCharaPath = "chara_char_v2" + png.Extension
	chubCardPath  = "chara_card_v2" + png.Extension
)

var (
	bookRegexp = regexp.MustCompile(`lorebooks/([^"\s<>()]+)`)
)

type chubAIFetcher struct {
	BaseHandler
}

// NewChubAIFetcher - Create a new ChubAI source
func NewChubAIFetcher(client *reqx.Client) fetcher.Fetcher {
	impl := &chubAIFetcher{
		BaseHandler: BaseHandler{
			client:    client,
			sourceID:  source.ChubAI,
			sourceURL: chubSourceURL,
			directURL: chubMainURL,
			mainURL:   chubMainURL,
			baseURLs:  []string{chubMainURL, chubAlternateURL},
		},
	}
	impl.Extends(impl)
	return impl
}

func (s *chubAIFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(chubApiURL, characterID)
	return s.client.R().Get(metadataURL)
}

func (s *chubAIFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return s.BaseHandler.CreateBinder(metadataResponse.GetByPath("node", "fullPath").String(), metadataResponse)
}

func (s *chubAIFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	node := metadataBinder.Get("node")
	definitionNode := node.Get("definition")

	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     s.DirectURL(metadataBinder.CharacterID),
		PlatformID:    node.Get("id").String(),
		CharacterID:   metadataBinder.CharacterID,
		Name:          definitionNode.Get("name").String(),
		Title:         node.Get("name").String(),
		Tagline:       node.Get("tagline").String(),
		CreateTime:    s.fromDate(chubAiDateFormat, node.Get("createdAt").String(), metadataBinder.NormalizedURL),
		UpdateTime:    s.fromDate(chubAiDateFormat, node.Get("lastActivityAt").String(), metadataBinder.NormalizedURL),
		Tags:          models.TagsFromJsonArray(node.Get("topics"), sonicx.WrapString),
	}, nil
}

func (s *chubAIFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	displayName := strings.Split(metadataBinder.CharacterID, `/`)[0]

	response, err := reqx.String(s.client.R().Get(fmt.Sprintf(chubApiUsersURL, displayName)))
	if err != nil {
		return nil, err
	}

	wrap, err := sonicx.GetFromString(response)
	if err != nil {
		return nil, err
	}

	return &models.CreatorInfo{
		Nickname:   wrap.Get("username").String(),
		Username:   wrap.Get("name").String(),
		PlatformID: wrap.Get("id").String(),
	}, nil
}

func (s *chubAIFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	bookIDs := sonicx.ArrayToMap(
		metadataBinder.GetByPath("node", "related_lorebooks"),
		func(token string) bool {
			intToken, tokenErr := cast.ToIntE(token)
			return tokenErr != nil || (tokenErr == nil && intToken >= 0)
		},
		sonicx.WrapString,
	)

	linkedBookResponses, linkedBookUpdateTime := s.retrieveLinkedBooks(metadataBinder, bookIDs)
	auxBookResponses, auxBookUpdateTime := s.retrieveAuxBooks(metadataBinder, bookIDs)
	linkedBookResponses = append(linkedBookResponses, auxBookResponses...)

	return &fetcher.BookBinder{
		Responses:  linkedBookResponses,
		UpdateTime: max(linkedBookUpdateTime, auxBookUpdateTime),
	}, nil
}

func (s *chubAIFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	node := binder.Get("node")
	chubCardURL := node.Get("max_res_url").String()
	backupURL := node.Get("avatar_url").String()

	characterCard, err := s.retrieveCardData(chubCardURL, backupURL)
	if err != nil {
		return nil, err
	}
	definitionNode := node.Get("definition")

	if err := s.updateFieldsWithFallback(characterCard, binder, definitionNode); err != nil {
		return nil, err
	}

	merger := character.NewBookMerger()

	embeddedBook := character.DefaultBook()
	if err = sonicx.Config.UnmarshalFromString(definitionNode.Get("embedded_lorebook").Raw(), &embeddedBook); err == nil && embeddedBook != nil {
		if stringsx.IsBlank(string(embeddedBook.Name)) {
			embeddedBook.Name = character.BookNamePlaceholder
		}
		merger.AppendBook(embeddedBook)
	}

	for _, bookResponse := range binder.Responses {
		book := character.DefaultBook()
		tagline := bookResponse.GetByPath("node", "tagline").String()
		bookDefinition := bookResponse.GetByPath("node", "definition")
		chubDescription := bookDefinition.Get("description").String()
		chubName := bookDefinition.Get("name").String()
		var descriptionTokens []string
		if err = sonicx.Config.UnmarshalFromString(bookDefinition.Get("embedded_lorebook").Raw(), &book); err == nil && book != nil {
			book.Name.SetIf(chubName)
			switch {
			case len(book.Entries) == 0:
				descriptionTokens = []string{string(book.Description)}
			case chubDescription == string(book.Description):
				descriptionTokens = []string{tagline, chubDescription}
			default:
				descriptionTokens = []string{tagline, chubDescription, string(book.Description)}
			}
			book.Description = property.String(stringsx.JoinNonBlank(character.CreatorNotesSeparator, descriptionTokens...))
			merger.AppendBook(book)
		}
	}

	characterCard.CharacterBook = merger.Build()

	return characterCard, nil
}

func (s *chubAIFetcher) updateFieldsWithFallback(characterCard *png.CharacterCard, binder *fetcher.Binder, definitionNode fetcher.JsonResponse) error {
	characterCard.Description.SetIf(definitionNode.Get("personality").String())
	characterCard.Personality.SetIf(definitionNode.Get("tavern_personality").String())
	characterCard.Scenario.SetIf(definitionNode.Get("scenario").String())
	characterCard.FirstMessage.SetIf(definitionNode.Get("first_message").String())
	characterCard.MessageExamples.SetIf(definitionNode.Get("example_dialogs").String())
	characterCard.CreatorNotes.SetIf(definitionNode.Get("description").String())
	characterCard.SystemPrompt.SetIf(definitionNode.Get("system_prompt").String())
	characterCard.PostHistoryInstructions.SetIf(definitionNode.Get("post_history_instructions").String())
	var alternateGreetings property.StringArray

	err := sonicx.Config.UnmarshalFromString(definitionNode.Get("alternate_greetings").Raw(), &alternateGreetings)
	if err != nil {
		return err
	}
	characterCard.AlternateGreetings = slicesx.MergeStable(alternateGreetings, characterCard.AlternateGreetings)
	return nil
}

func (s *chubAIFetcher) retrieveCardData(cardURL string, backupURL string) (*png.CharacterCard, error) {
	rawCard, err := png.FromURL(s.client, cardURL).LastVersion().Get()
	if err != nil {
		rawCard, err = png.FromURL(s.client, s.fixAvatarURL(cardURL)).LastVersion().Get()
	}
	if err != nil {
		rawCard, err = png.FromURL(s.client, backupURL).LastVersion().Get()
	}
	if err != nil {
		return nil, err
	}

	return rawCard.Decode()
}

func (s *chubAIFetcher) retrieveLinkedBooks(metadataBinder *fetcher.MetadataBinder, bookIDs *orderedmap.OrderedMap[string, struct{}]) ([]fetcher.JsonResponse, timestamp.Nano) {
	var bookResponses []fetcher.JsonResponse
	maxBookUpdateTime := timestamp.Nano(0)
	for bookID := range bookIDs.Keys() {
		if parsedResponse, bookUpdateTime, found := s.retrieveBookData(metadataBinder, bookID); found {
			bookResponses = append(bookResponses, parsedResponse)
			maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
		}
	}

	return bookResponses, maxBookUpdateTime
}

func (s *chubAIFetcher) retrieveAuxBooks(metadataBinder *fetcher.MetadataBinder, bookIDs *orderedmap.OrderedMap[string, struct{}]) ([]fetcher.JsonResponse, timestamp.Nano) {
	var bookResponses []fetcher.JsonResponse
	maxBookUpdateTime := timestamp.Nano(0)
	auxSources := metadataBinder.GetByPath("node", "description").String() + symbols.Space + metadataBinder.GetByPath("node", "tagline").String()
	bookURLs := bookRegexp.FindAllStringSubmatch(auxSources, -1)
	for _, bookURLMatches := range bookURLs {
		if len(bookURLMatches) <= 1 {
			continue
		}
		bookPath := bookURLMatches[1]
		for stringsx.IsNotBlank(bookPath) {
			parsedResponse, bookUpdateTime, found := s.retrieveBookData(metadataBinder, bookPath)
			if found {
				bookID := strings.TrimSpace(parsedResponse.GetByPath("node", "id").String())
				if !bookIDs.Has(bookID) {
					bookIDs.Set(bookID, structx.Empty)
					bookResponses = append(bookResponses, parsedResponse)
					maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
				}
				break
			}
			lastSlash := max(strings.LastIndex(bookPath, symbols.Slash), 0)
			bookPath = bookPath[:lastSlash]
		}
	}

	return bookResponses, maxBookUpdateTime
}

func (s *chubAIFetcher) retrieveBookData(metadataBinder *fetcher.MetadataBinder, bookID string) (fetcher.JsonResponse, timestamp.Nano, bool) {
	response, err := reqx.String(
		s.client.R().
			SetContentType(reqx.JsonApplicationContentType).
			Get(fmt.Sprintf(chubApiBookURL, bookID)),
	)

	if err != nil {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, metadataBinder.DirectURL).
			Str("bookID", bookID).
			Msg("Lorebook unlinked/deleted")
		return sonicx.Empty, 0, false
	}

	wrap, err := sonicx.GetFromString(response)
	if err != nil {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, metadataBinder.DirectURL).
			Str("bookID", bookID).
			Msg("Could not parse book")
		return sonicx.Empty, 0, false
	}

	updateTime := s.fromDate(chubAiDateFormat, wrap.GetByPath("node", "lastActivityAt").String(), metadataBinder.DirectURL)
	return wrap, updateTime, true
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

// fixAvatarURL - corrects the chub avatar NormalizedURL in case it has the wrong path (replaces chara_char_v2 with chara_card_v2)
func (s *chubAIFetcher) fixAvatarURL(avatarURL string) string {
	avatarURL = strings.TrimSuffix(avatarURL, chubCharaPath)
	avatarURL = avatarURL + chubCardPath
	return avatarURL
}
