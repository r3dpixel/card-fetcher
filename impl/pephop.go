package impl

import (
	"fmt"
	"path"
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
	pephopUuidLength int = 36 // PepHop UUID length

	pephopDomain    string = "pephop.ai"                                                      // Domain for PepHop
	pephopPath      string = "characters/"                                                    // Path for PepHop
	pephopApiURL    string = "https://api.eosai.chat/characters/%s"                           // API URL for PepHop
	pephopAvatarURL string = "https://sp.eosai.chat//storage/v1/object/public/bot-avatars/%s" // Avatar Download URL for PepHop

	pepHopDateFormat string = time.RFC3339Nano // Date Format for PepHop
)

// PephopBuilder builder for PepHop fetcher
type PephopBuilder struct{}

// Build creates a new PepHop fetcher
func (b PephopBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewPephopFetcher(client)
}

// pephopFetcher PepHop fetcher implementation
type pephopFetcher struct {
	BaseFetcher
}

// NewPephopFetcher create a new PepHop fetcher
func NewPephopFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(pephopDomain, pephopPath)
	impl := &pephopFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.PepHop,
			sourceURL: pephopDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: pephopDomain, Path: pephopPath},
			},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *pephopFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataURL := fmt.Sprintf(pephopApiURL, characterID)
	return f.client.R().Get(metadataURL)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *pephopFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "id")
}

// FetchCardInfo fetches the card info from the source
func (f *pephopFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the card name
	cardName := metadataBinder.Get("name").String()

	// Extract tags
	tags := models.TagsFromJsonArray(
		metadataBinder.Get("tags"),
		func(result *sonicx.Wrap) string {
			return result.Get("name").String()
		},
	)

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Name:          cardName,
		Title:         cardName,
		Tagline:       metadataBinder.Get("description").String(),
		CreateTime:    timestamp.ParseF(pepHopDateFormat, metadataBinder.Get("created_at").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(pepHopDateFormat, metadataBinder.Get("updated_at").String(), trace.URL, metadataBinder.NormalizedURL),
		IsForked:      false,
		Tags:          tags,
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *pephopFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	displayName := metadataBinder.Get("creator_name").String()
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: metadataBinder.Get("creator_id").String(),
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *pephopFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Fetch the character avatar from the API
	pepHopAvatarURL := fmt.Sprintf(pephopAvatarURL, binder.Get("avatar").String())
	rawCard, err := png.FromURL(f.client, pepHopAvatarURL).LastVersion().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Update the character card fields
	characterCard.Description = property.String(binder.Get("personality").String())
	characterCard.Scenario = property.String(binder.Get("scenario").String())
	characterCard.FirstMessage = property.String(binder.Get("first_message").String())
	characterCard.MessageExamples = property.String(binder.Get("example_dialogs").String())
	characterCard.CreatorNotes = property.String(binder.GetByPath("introduction", "characterIntroduction").String())

	// Return the character card
	return characterCard, nil
}

// CharacterID returns the characterID for pephop source
// For PepHop the suffix must be trimmed to leave just the real characterID
func (f *pephopFetcher) CharacterID(rawCharacterID string) string {
	return rawCharacterID[0:pephopUuidLength]
}
