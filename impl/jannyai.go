package impl

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	jannyAIUuidLength int    = 36 // JannyAI UUID length
	jannyAIDateFormat string = "2006-01-02 15:04:05.999999-07"

	jannyAISourceURL string = "jannyai.com"
	jannyAIMainURL   string = "jannyai.com/characters/"
	jannyAIApiURL    string = "https://api.jannyai.com/api/v1/characters/%s"
	jannyAIAvatarURL string = "https://image.jannyai.com/bot-avatars/%s"
)

type JannyCookies struct {
	CloudflareClearance string
	UserAgent           string
}

type JannyAIOpts struct {
	CookieProvider func() JannyCookies
}

type JannyAIBuilder JannyAIOpts

func (b JannyAIBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewJannyAIFetcher(client, JannyAIOpts(b))
}

type jannyAIFetcher struct {
	BaseFetcher
	cookieProvider func() JannyCookies
	commonCookie   http.Cookie
}

func NewJannyAIFetcher(client *reqx.Client, opts JannyAIOpts) fetcher.Fetcher {
	impl := &jannyAIFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.JannyAI,
			sourceURL: jannyAISourceURL,
			directURL: jannyAIMainURL,
			mainURL:   jannyAIMainURL,
			baseURLs:  []string{jannyAIMainURL},
		},
		cookieProvider: opts.CookieProvider,
	}
	impl.Extends(impl)
	return impl
}

func (f *jannyAIFetcher) CharacterID(url string, matchedURL string) string {
	return f.BaseFetcher.CharacterID(url, matchedURL)[0:jannyAIUuidLength]
}

func (f *jannyAIFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	cookies := f.cookieProvider()
	return f.client.R().
		SetHeader("User-Agent", cookies.UserAgent).
		SetCookies(
			&http.Cookie{
				Name:  "cf_clearance",
				Value: cookies.CloudflareClearance,
			},
		).
		Get(fmt.Sprintf(jannyAIApiURL, characterID))
}

func (f *jannyAIFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return f.BaseFetcher.CreateBinder(metadataResponse.Get("id").String(), metadataResponse)
}

func (f *jannyAIFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	name := metadataBinder.Get("name").String()
	createTime := timestamp.ParseF(jannyAIDateFormat, metadataBinder.Get("createdAt").String(), trace.URL, metadataBinder.NormalizedURL)
	tags := models.TagsFromJsonArray(metadataBinder.Get("tags"), func(result *sonicx.Wrap) string {
		return result.Get("name").String()
	})

	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     metadataBinder.DirectURL,
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Name:          name,
		Title:         name,
		Tagline:       stringsx.Empty,
		CreateTime:    createTime,
		UpdateTime:    timestamp.Nano(time.Now().Truncate(24 * time.Hour).UnixNano()),
		IsForked:      false,
		Tags:          tags,
	}, nil
}

func (f *jannyAIFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	username := metadataBinder.Get("creatorName").String()
	return &models.CreatorInfo{
		Nickname:   username,
		Username:   username,
		PlatformID: metadataBinder.Get("creatorId").String(),
	}, nil
}

func (f *jannyAIFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Download avatar and transform to PNG
	jannyAIAvatarURL := fmt.Sprintf(jannyAIAvatarURL, binder.Get("avatar").String())
	rawCard, err := png.FromURL(f.client, jannyAIAvatarURL).LastVersion().Get()
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	characterCard.Description = property.String(binder.Get("personality").String())
	characterCard.Scenario = property.String(binder.Get("scenario").String())
	characterCard.FirstMessage = property.String(binder.Get("firstMessage").String())
	characterCard.MessageExamples = property.String(binder.Get("exampleDialogs").String())
	characterCard.CreatorNotes = property.String(binder.Get("description").String())
	return characterCard, nil
}

func (f *jannyAIFetcher) IsSourceUp() bool {
	cookies := f.cookieProvider()
	_, err := f.client.R().
		SetHeader("User-Agent", cookies.UserAgent).
		SetCookies(
			&http.Cookie{
				Name:  "cf_clearance",
				Value: cookies.CloudflareClearance,
			},
		).
		Get("https://" + f.sourceURL + "/collections")
	return err == nil
}
