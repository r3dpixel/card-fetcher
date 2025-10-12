package impl

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/imroc/req/v3"
	"github.com/r3dpixel/card-fetcher/fetcher"
	"github.com/r3dpixel/card-fetcher/models"
	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
	"github.com/r3dpixel/card-parser/property"
	"github.com/r3dpixel/toolkit/cred"
	"github.com/r3dpixel/toolkit/reqx"
	"github.com/r3dpixel/toolkit/sonicx"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
)

const (
	pygmalionAuthUsernameField = "username"
	pygmalionAuthPasswordField = "password"

	// Pygmalion Headers
	pygmalionHost    = "auth.pygmalion.chat"     // Header for Pygmalion requests
	pygmalionOrigin  = "https://pygmalion.chat"  // Header for Pygmalion requests
	pygmalionReferer = "https://pygmalion.chat/" // Header for Pygmalion requests

	pygmalionSourceURL     = "pygmalion.chat"
	pygmalionBaseURL       = "pygmalion.chat/character/"                                                           // Main NormalizedURL for Pygmalion
	pygmalionApiURL        = "https://server.pygmalion.chat/galatea.v1.PublicCharacterService/Character"           // API NormalizedURL for Pygmalion
	pygmalionAuthURL       = "https://auth.pygmalion.chat/session"                                                 // Authentication NormalizedURL for Pygmalion
	pygmalionCardExportURL = "https://server.pygmalion.chat/api/export/character/%s/v2"                            // Avatar Download NormalizedURL for Pygmalion (contains chara metadata - PNG V2)
	pygmalionLinkedBookURL = "https://server.pygmalion.chat/galatea.v1.UserLorebookService/LorebooksByCharacterId" // Book Download NormalizedURL for Pygmalion
)

type PygmalionBuilder struct {
	IdentityReader cred.IdentityReader
}

func (b PygmalionBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewPygmalionFetcher(client, b.IdentityReader)
}

type pygmalionFetcher struct {
	BaseFetcher
	headers map[string]string
}

// NewPygmalionFetcher - Create a new ChubAI source
func NewPygmalionFetcher(client *reqx.Client, identityReader cred.IdentityReader) fetcher.Fetcher {
	impl := &pygmalionFetcher{
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
	impl.Extends(impl)
	impl.client.RegisterAuth(impl.serviceLabel, identityReader, impl.refreshBearerToken)

	return impl
}

func (s *pygmalionFetcher) Close() {
	s.client.UnregisterAuth(s.serviceLabel)
}

func (s *pygmalionFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	requestBodyBytes, _ := sonicx.Config.Marshal(
		map[string]string{
			"characterMetaId": characterID,
		},
	)
	return s.client.R().
		SetContentType(reqx.JsonApplicationContentType).
		SetBody(requestBodyBytes).
		Post(pygmalionApiURL)
}

func (s *pygmalionFetcher) CreateBinder(characterID string, metadataResponse fetcher.JsonResponse) (*fetcher.MetadataBinder, error) {
	newCharacterID := metadataResponse.GetByPath("character", "id").String()
	return s.BaseFetcher.CreateBinder(newCharacterID, metadataResponse)
}

func (s *pygmalionFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	characterNode := metadataBinder.Get("character")

	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     s.DirectURL(metadataBinder.CharacterID),
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Title:         characterNode.Get("displayName").String(),
		Name:          characterNode.GetByPath("personality", "name").String(),
		Tagline:       characterNode.Get("description").String(),
		CreateTime:    timestamp.Convert[timestamp.Nano](timestamp.Seconds(characterNode.Get("createdAt").Integer64())),
		UpdateTime:    timestamp.Convert[timestamp.Nano](timestamp.Seconds(characterNode.Get("updatedAt").Integer64())),
		Tags:          models.TagsFromJsonArray(characterNode.Get("tags"), sonicx.WrapString),
	}, nil
}

func (s *pygmalionFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	ownerNode := metadataBinder.GetByPath("character", "owner")

	return &models.CreatorInfo{
		Nickname:   ownerNode.Get("displayName").String(),
		Username:   ownerNode.Get("username").String(),
		PlatformID: ownerNode.Get("id").String(),
	}, nil
}

func (s *pygmalionFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	bookResponses, err := s.fetchBookResponses(metadataBinder.CharacterID)
	if err != nil {
		return nil, err
	}
	lorebooksNode := bookResponses.Get("lorebooks")
	if lorebooksNode.TypeSafe() != ast.V_ARRAY {
		return &fetcher.EmptyBookBinder, nil
	}
	bookArray, err := lorebooksNode.ArrayUseNode()
	if err != nil {
		return nil, err
	}
	if len(bookArray) == 0 {
		return &fetcher.EmptyBookBinder, nil
	}

	parsedResponses := make([]fetcher.JsonResponse, len(bookArray))
	bookUpdateTime := timestamp.Nano(0)
	for index, bookResult := range bookArray {
		bookResponse := sonicx.Of(bookResult)
		updatedAt := bookResponse.Get("updatedAt").Integer64()
		parsedResponses[index] = bookResponse
		bookUpdateTime = max(bookUpdateTime, timestamp.Convert[timestamp.Nano](timestamp.Seconds(updatedAt)))
	}

	return &fetcher.BookBinder{
		Responses:  parsedResponses,
		UpdateTime: bookUpdateTime,
	}, nil
}

func (s *pygmalionFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	characterCard, err := s.fetchCharacterCard(binder)
	if err != nil {
		return nil, err
	}

	characterCard.Sheet.CharacterBook = s.parseBookResponses(binder)

	return characterCard, nil
}

func (s *pygmalionFetcher) fetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Download avatar and transform to PNG
	avatarUrl := binder.GetByPath("character", "avatarUrl").String()
	rawCard, err := png.FromURL(s.client, avatarUrl).LastVersion().Get()
	if err != nil {
		return nil, err
	}
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, err
	}

	bytes, err := reqx.Bytes(
		s.client.R().
			SetContentType(reqx.JsonApplicationContentType).
			Get(fmt.Sprintf(pygmalionCardExportURL, binder.CharacterID)),
	)
	if err != nil {
		return nil, err
	}

	// Optimization to remove the prefix `{character:` and suffix `}` from the byte response without processing
	characterCard.Sheet, err = character.FromBytes(bytes[13 : len(bytes)-1])
	// If the card is nil, then the export failed (error is treated upstream)
	if err != nil {
		return nil, err
	}
	characterCard.Sheet.CreatorNotes = property.String(stringsx.Empty)

	// Return the parsed PNG card
	return characterCard, nil
}

func (s *pygmalionFetcher) parseBookResponses(binder *fetcher.Binder) *character.Book {

	bookMerger := character.NewBookMerger()

	for _, bookResponse := range binder.Responses {
		var pygBook pygmalionBook
		err := sonicx.Config.UnmarshalFromString(bookResponse.Raw(), &pygBook)
		if err != nil {
			println(err)
			log.Warn().Err(err).
				Str(trace.SOURCE, string(s.sourceID)).
				Str(trace.URL, binder.DirectURL).
				Msg("Could not parse book")
			continue
		}
		bookMerger.AppendBook(pygBook.convert())
	}

	return bookMerger.Build()
}

func (s *pygmalionFetcher) fetchBookResponses(characterID string) (fetcher.JsonResponse, error) {
	requestBodyBytes, _ := sonicx.Config.Marshal(
		map[string]string{
			"characterId": characterID,
		},
	)

	response, err := reqx.String(
		s.client.AR(s.serviceLabel).
			SetContentType(reqx.JsonApplicationContentType).
			SetBody(requestBodyBytes).
			Post(pygmalionLinkedBookURL),
	)
	if err != nil {
		return sonicx.Empty, err
	}

	wrap, err := sonic.GetFromString(response)
	if err != nil {
		return sonicx.Empty, err
	}

	return sonicx.Of(wrap), nil
}

func (s *pygmalionFetcher) refreshBearerToken(c *reqx.Client, identity cred.Identity) (string, error) {
	credentialsMap := map[string]string{
		pygmalionAuthUsernameField: identity.User,
		pygmalionAuthPasswordField: identity.Secret,
	}

	response, err := reqx.String(
		c.R().
			SetContentType("application/x-www-form-urlencoded").
			SetHeaders(s.headers).
			SetFormData(credentialsMap).
			Post(pygmalionAuthURL),
	)
	if err != nil {
		return stringsx.Empty, err
	}

	wrap, err := sonicx.GetFromString(response, "result", "id_token")
	if err != nil {
		return stringsx.Empty, err
	}

	return wrap.String(), nil
}

type pygmalionBook struct {
	Name        property.String      `json:"name"`
	Description property.String      `json:"description"`
	Entries     []pygmalionBookEntry `json:"entries"`
}

func (p *pygmalionBook) convert() *character.Book {
	b := character.DefaultBook()
	b.Name = p.Name
	b.Description = p.Description
	b.Entries = make([]*character.BookEntry, len(p.Entries))

	for index := range p.Entries {
		b.Entries[index] = p.Entries[index].convert()
	}

	return b
}

type pygmalionBookEntry struct {
	ID             property.Union          `json:"id"`
	Title          property.String         `json:"title"`
	Content        property.String         `json:"content"`
	Priority       property.Integer        `json:"priority"`
	Keys           property.StringArray    `json:"keywords"`
	SecondaryKeys  property.StringArray    `json:"andKeywords"`
	Constant       property.Bool           `json:"alwaysPresent"`
	LorePosition   *property.LorePosition  `json:"position"`
	Role           *property.Role          `json:"role"`
	Enabled        property.Bool           `json:"enabled"`
	Selective      property.Bool           `json:"selective"`
	SelectiveLogic property.SelectiveLogic `json:"selectiveLogic"`
	Sticky         property.Integer        `json:"sticky"`
	Cooldown       property.Integer        `json:"cooldown"`
	Delay          property.Integer        `json:"delay"`
	Depth          property.Integer        `json:"depth"`
}

func (p *pygmalionBookEntry) convert() *character.BookEntry {
	entry := character.DefaultBookEntry()
	entry.ID = p.ID
	entry.Name = p.Title
	entry.Comment = p.Title
	entry.Content = p.Content
	entry.InsertionOrder = p.Priority
	entry.Keys = p.Keys
	entry.SecondaryKeys = p.SecondaryKeys
	entry.Constant = p.Constant
	entry.Extensions.LorePosition.SetIfPropertyPtr(p.LorePosition)
	entry.Extensions.Role.SetIfPropertyPtr(p.Role)
	entry.Enabled = p.Enabled
	entry.Selective = p.Selective
	entry.Extensions.SelectiveLogic = p.SelectiveLogic
	entry.Extensions.Sticky = p.Sticky
	entry.Extensions.Cooldown = p.Cooldown
	entry.Extensions.Delay = p.Delay
	entry.Extensions.Depth = p.Depth
	return entry
}
