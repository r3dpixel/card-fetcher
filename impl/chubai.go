package impl

import (
	"fmt"
	"path"
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
	chubDomain      string = "chub.ai"                                         // ChubAI domain
	chubPath        string = "characters/"                                     // Path to characters on ChubAI
	chubApiURL      string = "https://api.chub.ai/api/characters/%s?full=true" // Public API for retrieving metadata
	chubApiBookURL  string = "https://api.chub.ai/api/lorebooks/%s?full=true"  // Public API for retrieving books
	chubApiUsersURL string = "https://api.chub.ai/api/users/%s"

	chubAiDateFormat string = time.RFC3339Nano // Date Format for ChubAI API

	// ChubAI avatar paths
	chubCharaPath = "chara_char_v2" + png.Extension
	chubCardPath  = "chara_card_v2" + png.Extension
)

var (
	// Regexp for extracting book URLs from the character description
	bookRegexp = regexp.MustCompile(`lorebooks/([^"\s<>()]+)`)
)

// ChubAIBuilder builder for ChubAI fetcher
type ChubAIBuilder struct{}

// Build creates a new ChubAI fetcher
func (b ChubAIBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewChubAIFetcher(client)
}

// ChubAIFetcher ChubAI fetcher implementation
type chubAIFetcher struct {
	BaseFetcher
}

// NewChubAIFetcher creates a new ChubAI fetcher
func NewChubAIFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(chubDomain, chubPath)
	impl := &chubAIFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.ChubAI,
			sourceURL: chubDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: chubDomain, Path: chubPath},
				{Domain: "characterhub.org", Path: chubPath},
			},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *chubAIFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(chubApiURL, characterID)
	return f.client.R().Get(metadataURL)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *chubAIFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "node", "fullPath")
}

// FetchCardInfo fetches the card info from the source
func (f *chubAIFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the root node
	node := metadataBinder.Get("node")
	// Extract the definition node
	definitionNode := node.Get("definition")

	// Check if the character is forked (iterate through labels to find "forked")
	forked := sonicx.ArrayToSlice(
		node.Get("labels"),
		func(s string) bool {
			return strings.ToLower(s) == "forked"
		},
		func(wrap *sonicx.Wrap) string {
			return wrap.Get("title").String()
		},
	)

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    node.Get("id").String(),
		CharacterID:   metadataBinder.CharacterID,
		Name:          definitionNode.Get("name").String(),
		Title:         node.Get("name").String(),
		Tagline:       node.Get("tagline").String(),
		CreateTime:    timestamp.ParseF(chubAiDateFormat, node.Get("createdAt").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(chubAiDateFormat, node.Get("lastActivityAt").String(), trace.URL, metadataBinder.NormalizedURL),
		IsForked:      len(forked) > 0,
		Tags:          models.TagsFromJsonArray(node.Get("topics"), sonicx.WrapString),
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *chubAIFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	// Extract the displayName from the CharacterID
	displayName := strings.Split(metadataBinder.CharacterID, `/`)[0]

	// Fetch the creator data from the API
	response, err := reqx.String(f.client.R().Get(fmt.Sprintf(chubApiUsersURL, displayName)))
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchMetadataErr)
	}

	// Parse the response
	wrap, err := sonicx.GetFromString(response)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedMetadataErr)
	}

	// Return the creator info
	return &models.CreatorInfo{
		Nickname:   wrap.Get("username").String(),
		Username:   wrap.Get("name").String(),
		PlatformID: wrap.Get("id").String(),
	}, nil
}

// FetchBookResponses fetches the book responses from the source
func (f *chubAIFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	// Extract the book IDs from the metadataBinder
	bookIDs := sonicx.ArrayToMap(
		metadataBinder.GetByPath("node", "related_lorebooks"),
		func(token string) bool {
			// Check if the token is a valid integer
			intToken, tokenErr := cast.ToIntE(token)
			// Return true if the token is valid or if there was an error (not an integer)
			// This filters out any IDs that are negative integers
			return tokenErr != nil || (tokenErr == nil && intToken >= 0)
		},
		sonicx.WrapString,
	)

	// Fetch the book responses from the API
	linkedBookResponses, linkedBookUpdateTime := f.retrieveLinkedBooks(metadataBinder, bookIDs)
	// Fetch the aux book responses from the description and tagline
	auxBookResponses, auxBookUpdateTime := f.retrieveAuxBooks(metadataBinder, bookIDs)
	// Merge the book responses
	linkedBookResponses = append(linkedBookResponses, auxBookResponses...)

	// Return the book binder
	return &fetcher.BookBinder{
		Responses:  linkedBookResponses,
		UpdateTime: max(linkedBookUpdateTime, auxBookUpdateTime),
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *chubAIFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Extract the root node
	node := binder.Get("node")
	// Extract the character card URL
	chubCardURL := node.Get("max_res_url").String()
	// Extract the backup URL
	backupURL := node.Get("avatar_url").String()

	// Fetch the character card from the API (in order of preference: max_res_url, fixed max_res_url, avatar_url)
	rawCard, err := png.FromURL(f.client, chubCardURL, f.fixAvatarURL(chubCardURL), backupURL).LastVersion().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Update the character card with the definition data
	definitionNode := node.Get("definition")
	if err := f.updateFieldsWithFallback(characterCard, definitionNode); err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedCardDataErr)
	}

	// Merge the books into the character card (including the embedded book)
	characterCard.CharacterBook = f.mergeBooks(definitionNode.Get("embedded_lorebook").Raw(), binder)

	// Return the character card
	return characterCard, nil
}

// updateFieldsWithFallback updates the fields of the character card with the data from the definition node,
// falling back to the existing values if the definition values are blank
func (f *chubAIFetcher) updateFieldsWithFallback(characterCard *png.CharacterCard, definitionNode fetcher.JsonResponse) error {
	characterCard.Description.SetIf(definitionNode.Get("personality").String())
	characterCard.Personality.SetIf(definitionNode.Get("tavern_personality").String())
	characterCard.Scenario.SetIf(definitionNode.Get("scenario").String())
	characterCard.FirstMessage.SetIf(definitionNode.Get("first_message").String())
	characterCard.MessageExamples.SetIf(definitionNode.Get("example_dialogs").String())
	characterCard.CreatorNotes.SetIf(definitionNode.Get("description").String())
	characterCard.SystemPrompt.SetIf(definitionNode.Get("system_prompt").String())
	characterCard.PostHistoryInstructions.SetIf(definitionNode.Get("post_history_instructions").String())
	var alternateGreetings property.StringArray

	// Parse the alternate_greetings field
	err := sonicx.Config.UnmarshalFromString(definitionNode.Get("alternate_greetings").Raw(), &alternateGreetings)
	if err != nil {
		return err
	}

	// Merge the alternate greetings with the existing values, deduplicating them
	characterCard.AlternateGreetings = slicesx.DeduplicateStable(alternateGreetings, characterCard.AlternateGreetings)

	// Return nil (success)
	return nil
}

// mergeBooks merges all books from the binder (including the embedded book) into a single Book object
func (f *chubAIFetcher) mergeBooks(embeddedBookRaw string, binder *fetcher.Binder) *character.Book {
	// Create a new BookMerger
	merger := character.NewBookMerger()

	// Parse and merge the embedded book
	if embeddedBook, err := f.parseEmbeddedBook(embeddedBookRaw); err == nil {
		merger.AppendBook(embeddedBook)
	} else {
		log.Warn().
			Err(err).
			Str(trace.SOURCE, string(f.sourceID)).
			Str(trace.URL, binder.DirectURL).
			Msg("Could not parse embedded book")
	}

	// Iterate through all the book responses
	for _, bookResponse := range binder.Responses {
		// Parse and merge the linked/auxiliary book
		if book, bookName, err := f.parseBookResponse(bookResponse); err == nil {
			merger.AppendBook(book)
		} else {
			log.Warn().
				Err(err).
				Str(trace.SOURCE, string(f.sourceID)).
				Str(trace.URL, binder.DirectURL).
				Str("bookName", bookName).
				Msg("Could not parse linked book")
		}
	}

	// Return the merged book
	return merger.Build()
}

// ParseEmbeddedBook parses the embedded book from the raw string
func (f *chubAIFetcher) parseEmbeddedBook(embeddedBookRaw string) (*character.Book, error) {
	// Initialize the embedded book with the default values
	embeddedBook := character.DefaultBook()
	// Parse the embedded book from the raw string
	if err := sonicx.Config.UnmarshalFromString(embeddedBookRaw, &embeddedBook); err != nil {
		// Return the error if it occurred
		return nil, err
	}
	// No embedded book found
	if embeddedBook == nil {
		return nil, nil
	}

	// Set the name if it is blank
	if stringsx.IsBlank(string(embeddedBook.Name)) {
		// Use the placeholder name (will be replaced with the character name in the patcher phase)
		embeddedBook.Name = character.BookNamePlaceholder
	}

	// Return the embedded book
	return embeddedBook, nil
}

// parseBookResponse parses the book response from the API
func (f *chubAIFetcher) parseBookResponse(bookResponse fetcher.JsonResponse) (*character.Book, string, error) {
	// Initialize the book with the default values
	book := character.DefaultBook()

	// Extract book tagline
	tagline := bookResponse.GetByPath("node", "tagline").String()
	// Extract book definition
	bookDefinition := bookResponse.GetByPath("node", "definition")
	// Extract book description
	chubDescription := bookDefinition.Get("description").String()
	// Extract book name
	chubName := bookDefinition.Get("name").String()

	// Parse the book definition
	if err := sonicx.Config.UnmarshalFromString(bookDefinition.Get("embedded_lorebook").Raw(), &book); err != nil {
		// Return the error if it occurred
		return nil, chubName, err
	}

	// No book found
	if book == nil {
		return nil, chubName, nil
	}

	// Set the name if it's not blank
	book.Name.SetIf(chubName)

	// Compose the description
	var descriptionTokens []string
	switch {
	// If the book has no entries, set the description only (a book with no entries will be process by the merger, by moving description to entries)
	case len(book.Entries) == 0:
		descriptionTokens = []string{string(book.Description)}
	// If the chub description is the same as the parsed description, set tagline and parsed description
	case chubDescription == string(book.Description):
		descriptionTokens = []string{tagline, chubDescription}
	// Otherwise, set tagline, chub description and parsed description
	default:
		descriptionTokens = []string{tagline, chubDescription, string(book.Description)}
	}

	// Merge the description tokens
	book.Description = property.String(stringsx.JoinNonBlank(character.CreatorNotesSeparator, descriptionTokens...))

	// Return the book and the book name
	return book, chubName, nil
}

// retrieveLinkedBooks retrieves the linked books from the API
func (f *chubAIFetcher) retrieveLinkedBooks(metadataBinder *fetcher.MetadataBinder, bookIDs *orderedmap.OrderedMap[string, struct{}]) ([]fetcher.JsonResponse, timestamp.Nano) {
	// bookResponse will contain the book responses from the API
	var bookResponses []fetcher.JsonResponse
	// maxBookUpdateTime will contain the maximum update time of the book responses
	maxBookUpdateTime := timestamp.Nano(0)
	// Iterate through all the book IDs
	for bookID := range bookIDs.Keys() {
		// Retrieve the book data from the API
		if parsedResponse, bookUpdateTime, found := f.retrieveBookData(metadataBinder, bookID); found {
			// If found, add the book response to the bookResponses
			bookResponses = append(bookResponses, parsedResponse)
			// Update the maxBookUpdateTime
			maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
		}
	}

	// Return the book responses and the maxBookUpdateTime
	return bookResponses, maxBookUpdateTime
}

// retrieveAuxBooks retrieves the auxiliary books from the API
func (f *chubAIFetcher) retrieveAuxBooks(metadataBinder *fetcher.MetadataBinder, bookIDs *orderedmap.OrderedMap[string, struct{}]) ([]fetcher.JsonResponse, timestamp.Nano) {
	// bookResponses will contain the book responses from the API
	var bookResponses []fetcher.JsonResponse
	// maxBookUpdateTime will contain the maximum update time of the book responses
	maxBookUpdateTime := timestamp.Nano(0)
	// Merge the description and tagline into a single string for regex searching
	auxSources := metadataBinder.GetByPath("node", "description").String() + symbols.Space + metadataBinder.GetByPath("node", "tagline").String()
	// Find all book URLs in the description and tagline
	bookURLs := bookRegexp.FindAllStringSubmatch(auxSources, -1)

	// Iterate through all the matches
	for _, bookURLMatches := range bookURLs {
		// If the match has only one element, skip it (the regex contains a capturing group, which if present len(matches) > 1)
		if len(bookURLMatches) <= 1 {
			continue
		}
		// Get the capture group
		bookPath := bookURLMatches[1]

		// While the book path is not blank, retrieve the book data
		for stringsx.IsNotBlank(bookPath) {
			// Retrieve the book data from the API
			parsedResponse, bookUpdateTime, found := f.retrieveBookData(metadataBinder, bookPath)
			if found {
				// Extract the book ID
				bookID := strings.TrimSpace(parsedResponse.GetByPath("node", "id").String())
				// If the book ID is not already in the bookIDs, add it to the bookResponses and update the maxBookUpdateTime
				if !bookIDs.Has(bookID) {
					// Add the book ID to the bookIDs
					bookIDs.Set(bookID, structx.Empty)
					// Add the book response to the bookResponses
					bookResponses = append(bookResponses, parsedResponse)
					// Update the maxBookUpdateTime
					maxBookUpdateTime = max(maxBookUpdateTime, bookUpdateTime)
				}
				// If found, break the loop
				break
			}
			// Remove the part of the path
			// This searches for partial URL in case of: /lorebooks/123456789/main, /lorebooks/123456789/v2, ...
			lastSlash := max(strings.LastIndex(bookPath, symbols.Slash), 0)
			bookPath = bookPath[:lastSlash]
		}
	}

	// Return the book responses and the maxBookUpdateTime
	return bookResponses, maxBookUpdateTime
}

// retrieveBookData retrieves the book data from the API
func (f *chubAIFetcher) retrieveBookData(metadataBinder *fetcher.MetadataBinder, bookID string) (fetcher.JsonResponse, timestamp.Nano, bool) {
	// Retrieve the book data from the API
	response, err := reqx.String(
		f.client.R().
			SetContentType(reqx.JsonApplicationContentType).
			Get(fmt.Sprintf(chubApiBookURL, bookID)),
	)

	// Log the error and return empty book data if the book was not found
	if err != nil {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(f.sourceID)).
			Str(trace.URL, metadataBinder.DirectURL).
			Str("bookID", bookID).
			Msg("Lorebook unlinked/deleted")
		return sonicx.Empty, 0, false
	}

	// Parse the book data
	wrap, err := sonicx.GetFromString(response)
	if err != nil {
		log.Warn().Err(err).
			Str(trace.SOURCE, string(f.sourceID)).
			Str(trace.URL, metadataBinder.DirectURL).
			Str("bookID", bookID).
			Msg("Could not parse book")
		return sonicx.Empty, 0, false
	}

	// Return the book data and the update time
	updateTime := timestamp.ParseF(chubAiDateFormat, wrap.GetByPath("node", "lastActivityAt").String(), trace.URL, metadataBinder.DirectURL)
	return wrap, updateTime, true
}

// fixAvatarURL - corrects the chub avatar NormalizedURL in case it has the wrong path (replaces chara_char_v2 with chara_card_v2)
func (f *chubAIFetcher) fixAvatarURL(avatarURL string) string {
	avatarURL = strings.TrimSuffix(avatarURL, chubCharaPath)
	avatarURL = avatarURL + chubCardPath
	return avatarURL
}
