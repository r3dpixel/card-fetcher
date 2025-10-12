package impl

import (
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
)

const (
	pephopUuidLength int = 36 // PepHop Slug length

	pephopSourceURL string = "pephop.ai"
	pephopBaseURL   string = "pephop.ai/characters/"                                          // Main NormalizedURL for PepHop
	pephopApiURL    string = "https://api.eosai.chat/characters/%s"                           // API NormalizedURL for PepHop
	pephopAvatarURL string = "https://sp.eosai.chat//storage/v1/object/public/bot-avatars/%s" // Avatar Download NormalizedURL for PepHop

	pepHopDateFormat string = time.RFC3339Nano // Date Format for PepHop
)

type PephopBuilder struct{}

func (b PephopBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewPephopFetcher(client)
}

type pephopFetcher struct {
	BaseFetcher
}

// NewPephopFetcher - Create a new ChubAI source
func NewPephopFetcher(client *reqx.Client) fetcher.Fetcher {
	impl := &pephopFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.PepHop,
			sourceURL: pephopSourceURL,
			directURL: pephopBaseURL,
			mainURL:   pephopBaseURL,
			baseURLs:  []string{pephopBaseURL},
		},
	}
	impl.Extends(impl)
	return impl
}

func (s *pephopFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(pephopApiURL, characterID)
	return s.client.R().Get(metadataURL)
}

func (s *pephopFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	return s.BaseFetcher.CreateBinder(metadataResponse.Get("id").String(), metadataResponse)
}

func (s *pephopFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	cardName := metadataBinder.Get("name").String()

	tags := models.TagsFromJsonArray(
		metadataBinder.Get("tags"),
		func(result *sonicx.Wrap) string {
			return result.Get("name").String()
		},
	)

	metadata := &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     s.DirectURL(metadataBinder.CharacterID),
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Name:          cardName,
		Title:         cardName,
		Tagline:       metadataBinder.Get("description").String(),
		CreateTime:    timestamp.ParseF[timestamp.Nano](pepHopDateFormat, metadataBinder.Get("created_at").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF[timestamp.Nano](pepHopDateFormat, metadataBinder.Get("updated_at").String(), trace.URL, metadataBinder.NormalizedURL),
		Tags:          tags,
	}

	return metadata, nil
}

func (s *pephopFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	displayName := metadataBinder.Get("creator_name").String()
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: metadataBinder.Get("creator_id").String(),
	}, nil
}

func FetchBookResponses(*fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	return &fetcher.EmptyBookBinder, nil
}

func (s *pephopFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Download avatar and transform to PNG
	pepHopAvatarURL := fmt.Sprintf(pephopAvatarURL, binder.Get("avatar").String())
	rawCard, err := png.FromURL(s.client, pepHopAvatarURL).LastVersion().Get()
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	// Assign the character description field
	characterCard.Description = property.String(binder.Get("personality").String())
	// Personality field is not used on PepHop
	// Assign the character scenario field
	characterCard.Scenario = property.String(binder.Get("scenario").String())
	// Assign the first message
	characterCard.FirstMessage = property.String(binder.Get("first_message").String())
	// Assign the example dialogs
	characterCard.MessageExamples = property.String(binder.Get("example_dialogs").String())
	// Assemble CreatorNotes using description/introduction from the json response
	// Tagline for PepHop is the original creator notes
	// Retrieve the character introduction
	// Assign the assembled creator notes
	characterCard.CreatorNotes = property.String(binder.GetByPath("introduction", "characterIntroduction").String())

	// Return the parsed PNG sheet
	return characterCard, nil
}

// CharacterID - returns the characterID for pephop source
// For PepHop the suffix must be trimmed to leave just the real Slug
func (s *pephopFetcher) CharacterID(url string, matchedURL string) string {
	return s.BaseFetcher.CharacterID(url, matchedURL)[0:pephopUuidLength]
}
