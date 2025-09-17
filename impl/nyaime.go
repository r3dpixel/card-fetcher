package impl

import (
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/toolkit/gjsonx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/tidwall/gjson"
)

const (
	// NyaiMeHeaders
	nyaiMeReferer string = "https://nyai.me/" // Header for NyaiMe download request
	nyaiMeOrigin  string = "https://nyai.me"  // Header for NyaiMe download request

	nyaiStartingRune rune   = 'a' // Starting rune for NyaiMe identifier conversion to PostID (base26 conversion)
	nyaiMeSourceURL  string = "nyai.me"
	nyaiMeBaseURL    string = "nyai.me/ai/bots/"     // Main NormalizedURL for NyaiMe
	nyaiMeApiURL     string = "https://api.nyai.me/" // API NormalizedURL for NyaiMe

	nyaiMeShortDescriptionField string = "Post.ShortDescription" // The field name of the short description for NyaiMe
	nyaiMeDateFormat            string = time.RFC3339Nano        // Date Format for NyaiMe

)

type nyaiMeFetcher struct {
	BaseHandler
	headers map[string]string
}

// NewNyaiMeFetcher - Create a new NyaiMe source
func NewNyaiMeFetcher(client *req.Client) SourceHandler {
	impl := &nyaiMeFetcher{
		BaseHandler: BaseHandler{
			client:    client,
			sourceID:  source.NyaiMe,
			sourceURL: nyaiMeSourceURL,
			directURL: nyaiMeBaseURL,
			mainURL:   nyaiMeBaseURL,
			baseURLs:  []string{nyaiMeBaseURL},
		},
		headers: map[string]string{
			"Referer":     nyaiMeReferer,
			"Origin":      nyaiMeOrigin,
			"RequestType": "PstGetPostPage",
			"IsGuest":     "1",
		},
	}
	return impl
}

func (s *nyaiMeFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	// Retrieve NyaiMe identifier
	identifier := s.getIdentifier(characterID)
	// Compute PostID (base26 conversion of the identifier)
	postID := s.getPostID(identifier)

	// Retrieve the metadata (log error is response is invalid)
	return s.client.R().
		SetHeaders(s.headers).
		SetBodyString(s.downloadRequestBody(postID)).
		Post(nyaiMeApiURL)
}

func (s *nyaiMeFetcher) ExtractMetadata(normalizedURL string, characterID string, metadataResponse gjson.Result) (*models.CardInfo, error) {
	postID := metadataResponse.Get("Post.ID").String()
	// Retrieve the real card name
	cardName := metadataResponse.Get("Post.Title").String()
	// Retrieve the character name
	metadataDefinitions, _ := strconv.Unquote(metadataResponse.Get("Post.AdditionalDefinitions.BotJSONBase64").String())
	name := gjson.Parse(metadataDefinitions).Get("data.name").String()

	// Retrieve creator
	creator := metadataResponse.Get("Post.UserName").String()
	// Tagline for NyaiMe is the short description
	tagline := metadataResponse.Get(nyaiMeShortDescriptionField).String()
	// Create a tag list and parse tags
	tags := models.TagsFromJsonArray(metadataResponse.Get("Post.Tags"), func(result gjson.Result) string {
		return gjsonx.Stringifier(result.Get("Name"))
	})

	// Extract the update time and created time
	updateTime := s.fromDate(nyaiMeDateFormat, metadataResponse.Get("Post.EditedDate").String(), normalizedURL)
	createTime := s.fromDate(nyaiMeDateFormat, metadataResponse.Get("Post.Date").String(), normalizedURL)

	metadata := &models.CardInfo{
		Source:         s.sourceID,
		NormalizedURL:  normalizedURL,
		PlatformID:     postID,
		CharacterID:    characterID,
		Title:          cardName,
		Name:           name,
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
func (s *nyaiMeFetcher) FetchCharacterCard(normalizedURL string, characterID string, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata

	// Retrieve png sheet NormalizedURL
	nyaiMeCardURL := metadataResponse.Get("Post.ImageURL").String()
	// Download PNG sheet
	rawCard, err := png.FromURL(s.client, nyaiMeCardURL).DeepScan().Get()
	// If the characterCard or the sheet is nil, then the export failed
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	// TaskOf the characterCard sheet
	sheet := characterCard.Sheet

	// Assemble CreatorNotes using any creator notes in the downloaded sheet,
	// and any short description (tagline) / introduction from the json response
	// Tagline for NyaiMe is the short description
	// Retrieve the introduction
	introduction := metadataResponse.Get("Post.Content").String()
	// Assign the assembled creator notes
	sheet.Data.CreatorNotes = stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		introduction, sheet.Data.CreatorNotes,
	)

	// Return the parsed PNG sheet
	return characterCard, nil
}

// downloadRequestBody - create the body for the POST download request (based on characterID)
func (s *nyaiMeFetcher) downloadRequestBody(postId int) string {
	return `{"PostID": ` + strconv.Itoa(postId) + `}`
}

// getPostID - convert NyaiMe identifier to base26
func (s *nyaiMeFetcher) getPostID(identifier string) int {
	postId := 0
	for _, char := range identifier {
		postId = postId*26 + int(char-nyaiStartingRune+1)
	}
	return postId
}

// getIdentifier - parse NyaiMe url and retrieve unique identifier
func (s *nyaiMeFetcher) getIdentifier(url string) string {
	tokens := strings.Split(url, symbols.Underscore)
	return tokens[len(tokens)-1]
}
