package fetcher

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
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/tidwall/gjson"
)

const (
	// NyaiMeHeaders
	nyaiMeReferer string = "https://nyai.me/" // Header for NyaiMe download request
	nyaiMeOrigin  string = "https://nyai.me"  // Header for NyaiMe download request

	nyaiStartingRune rune   = 'a' // Starting rune for NyaiMe identifier conversion to PostID (base26 conversion)
	nyaiMeURL        string = "nyai.me"
	nyaiMeBaseURL    string = "nyai.me/ai/bots/"     // Main CardURL for NyaiMe
	nyaiMeApiURL     string = "https://api.nyai.me/" // API CardURL for NyaiMe

	nyaiMeShortDescriptionField string = "Post.ShortDescription" // The field name of the short description for NyaiMe
	nyaiMeDateFormat            string = time.RFC3339Nano        // Date Format for NyaiMe

)

type nyaiMeFetcher struct {
	BaseFetcher
	headers map[string]string
}

// NewNyaiMeFetcher - Create a new NyaiMe source
func NewNyaiMeFetcher() Fetcher {
	impl := &nyaiMeFetcher{
		BaseFetcher: BaseFetcher{
			sourceID:  source.NyaiMe,
			sourceURL: nyaiMeURL,
			directURL: nyaiMeBaseURL,
			baseURLs:  []string{nyaiMeBaseURL},
		},
		headers: map[string]string{
			"Referer":     nyaiMeReferer,
			"Origin":      nyaiMeOrigin,
			"RequestType": "PstGetPostPage",
			"IsGuest":     "1",
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadata - Retrieve metadata for given url
func (s *nyaiMeFetcher) FetchMetadata(c *req.Client, normalizedURL string, characterID string) (*models.Metadata, models.JsonResponse, error) {
	// Retrieve NyaiMe identifier
	identifier := s.getIdentifier(characterID)
	// Compute PostID (base26 conversion of the identifier)
	postID := s.getPostID(identifier)

	// Retrieve the metadata (log error is response is invalid)
	jsonResponse, err := c.R().
		SetHeaders(s.headers).
		SetBodyString(s.downloadRequestBody(postID)).
		Post(nyaiMeApiURL)
	// Check if the response is a valid JSON
	if !reqx.IsResponseOk(jsonResponse, err) {
		return nil, models.EmptyJsonResponse, s.fetchMetadataErr(normalizedURL, err)
	}
	// TaskOf the JSON string response
	metadataResponse := gjson.Parse(jsonResponse.String())

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

	metadata := &models.Metadata{
		Source:         s.sourceID,
		CardURL:        normalizedURL,
		PlatformID:     strconv.Itoa(postID),
		CharacterID:    characterID,
		CardName:       cardName,
		CharacterName:  name,
		Creator:        creator,
		Tagline:        tagline,
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		BookUpdateTime: 0,
		Tags:           tags,
	}

	// Return metadata
	fullResponse := models.JsonResponse{
		Metadata: metadataResponse,
	}
	return metadata, fullResponse, nil
}

// FetchPngCard - Retrieve card for given url
func (s *nyaiMeFetcher) FetchCharacterCard(c *req.Client, metadata *models.Metadata, response models.JsonResponse) (*png.CharacterCard, error) {
	metadataResponse := response.Metadata

	// Retrieve png sheet CardURL
	nyaiMeCardURL := metadataResponse.Get("Post.ImageURL").String()
	// Download PNG sheet
	rawCard, err := png.FromURL(c, nyaiMeCardURL).DeepScan().Get()
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
		metadata.Tagline, introduction, sheet.Data.CreatorNotes,
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
