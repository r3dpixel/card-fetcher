package impl

import (
	"path"
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

	nyaiStartingRune rune   = 'a'                    // Starting rune for NyaiMe identifier conversion to PostID (base26 conversion)
	nyaiMeDomain     string = "nyai.me"              // Domain for NyaiMe
	nyaiMePath       string = "ai/bots/"             // Path for NyaiMe
	nyaiMeApiURL     string = "https://api.nyai.me/" // API URL for NyaiMe

	nyaiMeDateFormat string = time.RFC3339Nano // Date Format for NyaiMe
)

// NyaiMeBuilder builder for NyaiMe fetcher
type NyaiMeBuilder struct{}

// Build creates a new NyaiMe fetcher
func (b NyaiMeBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewNyaiMeFetcher(client)
}

// nyaiMeFetcher NyaiMe fetcher implementation
type nyaiMeFetcher struct {
	BaseFetcher
	headers map[string]string
}

// NewNyaiMeFetcher create a new NyaiMe fetcher
func NewNyaiMeFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(nyaiMeDomain, nyaiMePath)
	impl := &nyaiMeFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.NyaiMe,
			sourceURL: nyaiMeDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: nyaiMeDomain, Path: nyaiMePath},
			},
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

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *nyaiMeFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	// Retrieve NyaiMe identifier
	identifier := f.getIdentifier(characterID)
	// Compute PostID (base26 conversion of the identifier)
	postID := f.getPostID(identifier)

	// Retrieve the metadata (log error is response is invalid)
	return f.client.R().
		SetHeaders(f.headers).
		SetBodyString(f.downloadRequestBody(postID)).
		Post(nyaiMeApiURL)
}

// FetchCardInfo fetches the card info from the source
func (f *nyaiMeFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the post node
	postNode := metadataBinder.Get("Post")

	// Extract the name
	var name string

	// Try to extract the name from the AdditionalDefinitions
	raw := stringsx.Unquote(postNode.GetByPath("AdditionalDefinitions", 0, "BotJSONBase64").Raw())
	if additionalDefinitions, err := sonicx.GetFromString(raw); err == nil {
		name = additionalDefinitions.GetByPath("data", "name").String()
	}

	// Extract tags
	tags := models.TagsFromJsonArray(
		postNode.Get("Tags"),
		func(result *sonicx.Wrap) string {
			return result.Get("Name").String()
		},
	)

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    postNode.Get("ID").String(),
		CharacterID:   metadataBinder.CharacterID,
		Name:          name,
		Title:         postNode.Get("Title").String(),
		Tagline:       postNode.Get("ShortDescription").String(),
		CreateTime:    timestamp.ParseF(nyaiMeDateFormat, postNode.Get("Date").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(nyaiMeDateFormat, postNode.Get("EditedDate").String(), trace.URL, metadataBinder.NormalizedURL),
		IsForked:      false,
		Tags:          tags,
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *nyaiMeFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	// Extract the displayName
	displayName := metadataBinder.GetByPath("Post", "UserName").String()

	// Return the creator info
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: displayName,
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *nyaiMeFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Extract the post node
	postNode := binder.Get("Post")
	// Retrieve the avatar URL
	nyaiMeCardURL := postNode.Get("ImageURL").String()
	// Fetch the character card from the API
	rawCard, err := png.FromURL(f.client, nyaiMeCardURL).LastVersion().Get()
	// If the characterCard or the sheet is nil, then the export failed
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}
	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Update the character sheet creator notes
	introduction := postNode.Get("Content").String()
	characterCard.CreatorNotes = property.String(stringsx.JoinNonBlank(
		character.CreatorNotesSeparator,
		introduction, string(characterCard.CreatorNotes),
	))

	// Return the character card
	return characterCard, nil
}

// downloadRequestBody - create the body for the POST download request (based on characterID)
func (f *nyaiMeFetcher) downloadRequestBody(postId int) string {
	return `{"PostID": ` + strconv.Itoa(postId) + `}`
}

// getPostID - convert NyaiMe identifier to base26
func (f *nyaiMeFetcher) getPostID(identifier string) int {
	postId := 0
	for _, char := range identifier {
		postId = postId*26 + int(char-nyaiStartingRune+1)
	}
	return postId
}

// getIdentifier - parse NyaiMe url and retrieve unique identifier
func (f *nyaiMeFetcher) getIdentifier(url string) string {
	tokens := strings.Split(url, symbols.Underscore)
	return tokens[len(tokens)-1]
}
