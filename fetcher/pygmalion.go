package fetcher

import (
	"encoding/json/v2"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/properties"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const (
	pygmalionAuthUsernameField = "username"
	pygmalionAuthPasswordField = "password"

	// Pygmalion Headers
	pygmalionHost    = "auth.pygmalion.chat"     // Header for Pygmalion requests
	pygmalionOrigin  = "https://pygmalion.chat"  // Header for Pygmalion requests
	pygmalionReferer = "https://pygmalion.chat/" // Header for Pygmalion requests

	pygmalionSourceURL     = "pygmalion.chat"
	pygmalionBaseURL       = "pygmalion.chat/character/"                                                           // Main CardURL for Pygmalion
	pygmalionApiURL        = "https://server.pygmalion.chat/galatea.v1.PublicCharacterService/Character"           // API CardURL for Pygmalion
	pygmalionAuthURL       = "https://auth.pygmalion.chat/session"                                                 // Authentication CardURL for Pygmalion
	pygmalionCardExportURL = "https://server.pygmalion.chat/api/export/character/%s/v2"                            // Avatar Download CardURL for Pygmalion (contains chara metadata - PNG V2)
	pygmalionLinkedBookURL = "https://server.pygmalion.chat/galatea.v1.UserLorebookService/LorebooksByCharacterId" // Book Download CardURL for Pygmalion
)

type pygmalionFetcher struct {
	BaseFetcher
	identityReader cred.IdentityReader
	authManager    *reqx.AuthManager
	headers        map[string]string
}

// NewPygmalionFetcher - Create a new ChubAI source
func NewPygmalionFetcher(client *req.Client, identityReader cred.IdentityReader) Fetcher {
	impl := &pygmalionFetcher{
		identityReader: identityReader,
		headers: map[string]string{
			"Referer": pygmalionReferer,
			"Origin":  pygmalionOrigin,
			"Host":    pygmalionHost,
		},
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.Pygmalion,
			sourceURL: pygmalionSourceURL,
			directURL: pygmalionBaseURL,
			mainURL:   pygmalionBaseURL,
			baseURLs:  []string{pygmalionBaseURL},
		},
	}
	impl.authManager = reqx.NewAuthManager(impl.refreshBearerToken)
	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *pygmalionFetcher) FetchMetadata(normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Send POST request for metadata (check if response is valid JSON, log error)
	metadataRequestBody := map[string]string{
		"characterMetaId": characterID,
	}
	requestBodyBytes, _ := json.Marshal(metadataRequestBody)
	jsonResponse, err := s.client.R().
		SetContentType(reqx.JsonApplicationContentType).
		SetBody(requestBodyBytes).
		Post(pygmalionApiURL)
	// Check if the response is a valid JSON
	if !reqx.IsResponseOk(jsonResponse, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// TaskOf the JSON string response
	metadataResponse := gjson.Parse(jsonResponse.String())

	// Retrieve the real card name
	cardName := metadataResponse.Get("character.displayName").String()
	// Retrieve the character name
	name := metadataResponse.Get("character.personality.name").String()

	// Retrieve creator
	creator := metadataResponse.Get("character.owner.displayName").String()
	// Tagline for Pygmalion is creator notes, which is extracted here from the metadata directly
	tagline := metadataResponse.Get("character.description").String()
	// Retrieve the card tags
	tags := models.TagsFromJsonArray(metadataResponse.Get("character.tags"), gjsonx.Stringifier)

	// Extract the update time and created time (format is in seconds, converted to timestamp.Milli)
	updateTime := timestamp.Convert[timestamp.Nano](timestamp.Seconds(metadataResponse.Get("character.updatedAt").Int()))
	createTime := timestamp.Convert[timestamp.Nano](timestamp.Seconds(metadataResponse.Get("character.createdAt").Int()))

	bookResponse := s.retrieveBookData(characterID)
	bookUpdateTime := timestamp.Nano(0)
	bookResponse.Get("lorebooks.#.updatedAt").ForEach(func(key, value gjson.Result) bool {
		bookUpdateTime = max(bookUpdateTime, timestamp.Convert[timestamp.Nano](timestamp.Seconds(value.Int())))
		return true
	})

	// Return the metadata
	metadata := &models.Metadata{
		Source:         s.sourceID,
		CardURL:        normalizedURL,
		PlatformID:     characterID,
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
		Metadata:      metadataResponse,
		BookResponses: []gjson.Result{bookResponse},
	}
	return metadata, fullResponse, nil
}

// FetchPngCard - Retrieve card for given url
func (s *pygmalionFetcher) FetchCharacterCard(metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata
	characterCard, err := s.retrieveCardData(metadata, metadataResponse)
	if err != nil {
		return nil, err
	}

	// Assign the assembled merged book
	characterCard.Sheet.Data.CharacterBook = s.getMergedBook(response)

	// Return the parsed PNG card
	return characterCard, nil
}

func (s *pygmalionFetcher) retrieveCardData(metadata *models.Metadata, gJsonResponse gjson.Result) (*png.CharacterCard, error) {
	// Download avatar and transform to PNG
	avatarUrl := gJsonResponse.Get("character.avatarUrl").String()
	rawCard, err := png.FromURL(s.client, avatarUrl).DeepScan().Get()
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	// TaskOf the characterCard card (from Pygmalion export)
	exportUrl := fmt.Sprintf(pygmalionCardExportURL, metadata.CharacterID)
	response, err := s.client.R().
		SetContentType(reqx.JsonApplicationContentType).
		Get(exportUrl)
	// Check if the response is a valid JSON (error is treated upstream)
	if !reqx.IsResponseOk(response, err) {
		return nil, err
	}
	// Unmarshal Pygmalion export into the characterCard
	jsonCard := response.Bytes()
	// Optimization to remove the prefix `{character:` and suffix `}` from the byte response without processing
	characterCard.Sheet, err = character.FromBytes(jsonCard[13 : len(jsonCard)-1])
	// If the card is nil, then the export failed (error is treated upstream)
	if err != nil {
		return nil, err
	}

	// Return the parsed PNG card
	return characterCard, nil
}

func (s *pygmalionFetcher) getMergedBook(response models.JsonResponse) *character.Book {
	// MergeTags all books (Pygmalion allows linking multiple Books to a character)
	// Merging in an embedded book is the only option
	bookMerger := character.NewBookMerger()

	for _, bookResponse := range response.BookResponses {
		// Retrieve books gJson array
		books := bookResponse.Get("lorebooks").Array()
		// If there are no books return nil
		if len(books) == 0 {
			continue
		}

		for bookIndex := range books {
			// Parse book
			book := books[bookIndex].Map()
			// Extract name and description
			bookMerger.AppendNameAndDescription(book["name"].String(), book["description"].String())

			// Parse the book entries
			entries := book["entries"].Array()
			for entryIndex := range entries {
				// Parse the JSON string for the current entry
				entryJson := entries[entryIndex].Map()
				entry := character.BookEntry{}
				// Parse the entry keywords
				entry.Keys = gjsonx.ArrayToSlice(entryJson["keywords"], stringsx.IsNotBlank, gjsonx.Stringifier)
				// Parse the entry content
				entry.Content = entryJson["content"].String()
				// Parse the entry enabled state
				entry.Enabled = entryJson["enabled"].Bool()
				// Parse the entry name (which is under field 'title' in pygmalion)
				entry.Name = new(string)
				*entry.Name = entryJson["title"].String()
				// Parse the entry priority
				entry.Priority = new(int)
				*entry.Priority = int(entryJson["priority"].Int())
				// Parse the entry selective state
				entry.Selective = new(bool)
				*entry.Selective = entryJson["selective"].Bool()
				// Parse the entry constant state (which is under the field 'alwaysPresent' in pygmalion)
				entry.Constant = new(bool)
				*entry.Constant = entryJson["alwaysPresent"].Bool()
				// Parse the entry secondary keywords (which are under the field 'andKeywords' in pygmalion)
				entry.SecondaryKeys = gjsonx.ArrayToSlice(entryJson["andKeywords"], stringsx.IsNotBlank, gjsonx.Stringifier)
				// Parse the sentry elective logic
				//entry.SelectiveLogic = new(properties.SelectiveLogic)
				_ = json.Unmarshal([]byte(entryJson["selectiveLogic"].String()), &entry.SelectiveLogic)
				// Parse the entry lore position
				// which for some reason, in pygmalion it is a bool denoting if the entry if 'before_char'
				// The absence of this value is treated as false
				if beforeDescriptionPosition := entryJson["beforeDescription"].Bool(); beforeDescriptionPosition {
					entry.LorePosition = properties.BeforeCharPosition
				} else {
					entry.LorePosition = properties.AfterCharPosition
				}
				// Append the current entry into the merged book
				bookMerger.AppendEntry(&entry)
			}
		}

	}

	// Return the assembled book
	return bookMerger.Build()
}

func (s *pygmalionFetcher) retrieveBookData(characterID string) gjson.Result {
	// Send GET request for the book (check if response is valid JSON, log error)
	bookRequestBody := map[string]string{
		"characterId": characterID,
	}
	requestBodyBytes, _ := json.Marshal(bookRequestBody)
	// Send the POST request for the metadata
	// Retrieve bearer token
	response, err := s.authManager.Do(s.client, func(c *req.Client, token string) (*req.Response, error) {
		return c.R().
			SetBearerAuthToken(token).
			SetContentType(reqx.JsonApplicationContentType).
			SetBody(requestBodyBytes).
			Post(pygmalionLinkedBookURL)
	})

	// Check if the response is a valid JSON
	if !reqx.IsResponseOk(response, err) {
		log.Error().Err(err).
			Str(trace.SOURCE, string(s.sourceID)).
			Str(trace.URL, pygmalionBaseURL+characterID).
			Msg("Could not parse book character")
		return gjsonx.Empty
	}

	// Return the response as a GJsonResponse
	return gjson.ParseBytes(response.Bytes())
}

func (s *pygmalionFetcher) refreshBearerToken(c *req.Client) (string, error) {
	identity, err := s.identityReader.Get()
	if err != nil {
		return stringsx.Empty, trace.Err().
			Wrap(err).
			Field(trace.SOURCE, string(s.sourceID)).
			Msg("Failed to get credentials")
	}

	credentialsMap := map[string]string{
		pygmalionAuthUsernameField: identity.User,
		pygmalionAuthPasswordField: identity.Secret,
	}

	response, err := c.R().
		SetContentType("application/x-www-form-urlencoded").
		SetHeaders(s.headers).
		SetFormData(credentialsMap).
		Post(pygmalionAuthURL)

	if !reqx.IsResponseOk(response, err) {
		return stringsx.Empty, trace.Err().
			Wrap(err).
			Field(trace.SOURCE, string(s.sourceID)).
			Field("username", identity.User).
			Msg("Failed to refresh bearer token")
	}

	return gjson.Get(response.String(), "result.id_token").String(), nil
}
