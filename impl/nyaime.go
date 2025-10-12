package impl

import (
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	// NyaiMeHeaders
	nyaiMeReferer string = "https://nyai.me/" // Header for NyaiMe download request
	nyaiMeOrigin  string = "https://nyai.me"  // Header for NyaiMe download request

	nyaiStartingRune rune   = 'a' // Starting rune for NyaiMe identifier conversion to PostID (base26 conversion)
	nyaiMeSourceURL  string = "nyai.me"
	nyaiMeBaseURL    string = "nyai.me/ai/bots/"     // Main NormalizedURL for NyaiMe
	nyaiMeApiURL     string = "https://api.nyai.me/" // API NormalizedURL for NyaiMe

	nyaiMeDateFormat string = time.RFC3339Nano // Date Format for NyaiMe
)

type NyaiMeBuilder struct{}

func (b NyaiMeBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewNyaiMeFetcher(client)
}

type nyaiMeFetcher struct {
	BaseFetcher
	headers map[string]string
}

// NewNyaiMeFetcher - Create a new NyaiMe source
func NewNyaiMeFetcher(client *reqx.Client) fetcher.Fetcher {
	impl := &nyaiMeFetcher{
		BaseFetcher: BaseFetcher{
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
	impl.Extends(impl)
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

func (s *nyaiMeFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	postNode := metadataBinder.Get("Post")

	var name string
	raw := stringsx.Unquote(postNode.GetByPath("AdditionalDefinitions", 0, "BotJSONBase64").Raw())
	if additionalDefinitions, err := sonicx.GetFromString(raw); err == nil {
		name = additionalDefinitions.GetByPath("data", "name").String()
	}

	tags := models.TagsFromJsonArray(
		postNode.Get("Tags"),
		func(result *sonicx.Wrap) string {
			return result.Get("Name").String()
		},
	)

	metadata := &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     s.DirectURL(metadataBinder.CharacterID),
		PlatformID:    postNode.Get("ID").String(),
		CharacterID:   metadataBinder.CharacterID,
		Name:          name,
		Title:         postNode.Get("Title").String(),
		Tagline:       postNode.Get("ShortDescription").String(),
		CreateTime:    timestamp.ParseF[timestamp.Nano](nyaiMeDateFormat, postNode.Get("Date").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF[timestamp.Nano](nyaiMeDateFormat, postNode.Get("EditedDate").String(), trace.URL, metadataBinder.NormalizedURL),
		Tags:          tags,
	}

	return metadata, nil
}

func (s *nyaiMeFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	postNode := metadataBinder.Get("Post")
	displayName := postNode.Get("UserName").String()

	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: displayName,
	}, nil
}

func (s *nyaiMeFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	postNode := binder.Get("Post")
	// Retrieve png sheet NormalizedURL
	nyaiMeCardURL := postNode.Get("ImageURL").String()
	// Download PNG sheet
	rawCard, err := png.FromURL(s.client, nyaiMeCardURL).LastVersion().Get()
	// If the characterCard or the sheet is nil, then the export failed
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	introduction := postNode.Get("Content").String()
	characterCard.CreatorNotes = property.String(stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		introduction, string(characterCard.CreatorNotes),
	))

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
