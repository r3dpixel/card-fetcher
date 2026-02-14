package impl

import (
	"fmt"
	"path"
	"slices"
	"strings"

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
	"github.com/r3dpixel/toolkit/timestamp"
	"github.com/r3dpixel/toolkit/trace"
	"github.com/rs/zerolog/log"
)

const (
	pygmalionAuthUsernameField = "username" // Field for Pygmalion username
	pygmalionAuthPasswordField = "password" // Field for Pygmalion password

	// Pygmalion Headers
	pygmalionHost    = "auth.pygmalion.chat"     // Header for Pygmalion requests
	pygmalionOrigin  = "https://pygmalion.chat"  // Header for Pygmalion requests
	pygmalionReferer = "https://pygmalion.chat/" // Header for Pygmalion requests

	pygmalionDomain        = "pygmalion.chat"                                                                      // Domain for Pygmalion
	pygmalionPath          = "character/"                                                                          // Path for Pygmalion
	pygmalionApiURL        = "https://server.pygmalion.chat/galatea.v1.PublicCharacterService/Character"           // API URL for Pygmalion
	pygmalionAuthURL       = "https://auth.pygmalion.chat/session"                                                 // Authentication URL for Pygmalion
	pygmalionCardExportURL = "https://server.pygmalion.chat/api/export/character/%s/v2"                            // Avatar Download URL for Pygmalion (contains chara metadata - PNG V2)
	pygmalionLinkedBookURL = "https://server.pygmalion.chat/galatea.v1.UserLorebookService/LorebooksByCharacterId" // Book Download URL for Pygmalion
)

// PygmalionOpts options for PygmalionBuilder
type PygmalionOpts struct {
	IdentityReader cred.IdentityReader
}

// PygmalionBuilder builder for Pygmalion fetcher
type PygmalionBuilder PygmalionOpts

// Build creates a Pygmalion fetcher using the provided options
func (b PygmalionBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewPygmalionFetcher(client, PygmalionOpts(b))
}

// pygmalionFetcher Pygmalion fetcher implementation
type pygmalionFetcher struct {
	BaseFetcher
	headers map[string]string
}

// NewPygmalionFetcher create a new Pygmalion fetcher
func NewPygmalionFetcher(client *reqx.Client, opts PygmalionOpts) fetcher.Fetcher {
	mainURL := path.Join(pygmalionDomain, pygmalionPath)
	impl := &pygmalionFetcher{
		headers: map[string]string{
			"Referer": pygmalionReferer,
			"Origin":  pygmalionOrigin,
			"Host":    pygmalionHost,
		},
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.Pygmalion,
			sourceURL: pephopDomain,
			directURL: mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: pygmalionDomain, Path: pygmalionPath},
			},
		},
	}
	impl.Extends(impl)
	impl.client.RegisterAuth(impl.serviceLabel, opts.IdentityReader, impl.refreshBearerToken)

	return impl
}

// Close closes the fetcher
func (f *pygmalionFetcher) Close() {
	f.client.UnregisterAuth(f.serviceLabel)
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *pygmalionFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	// Create request body
	requestBodyBytes, _ := sonicx.Config.Marshal(
		map[string]string{
			"characterMetaId": characterID,
		},
	)
	// Send request
	return f.client.R().
		SetContentType(reqx.JsonApplicationContentType).
		SetBody(requestBodyBytes).
		Post(pygmalionApiURL)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *pygmalionFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "character", "id")
}

// FetchCardInfo fetches the card info from the source
func (f *pygmalionFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	// Extract the character node
	characterNode := metadataBinder.Get("character")

	// Return the card info
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    metadataBinder.CharacterID,
		CharacterID:   metadataBinder.CharacterID,
		Title:         characterNode.Get("displayName").String(),
		Name:          characterNode.GetByPath("personality", "name").String(),
		Tagline:       characterNode.Get("description").String(),
		CreateTime:    timestamp.ConvertToNano(timestamp.Seconds(characterNode.Get("createdAt").Integer64())),
		UpdateTime:    timestamp.ConvertToNano(timestamp.Seconds(characterNode.Get("updatedAt").Integer64())),
		IsForked:      false,
		Tags:          models.TagsFromJsonArray(characterNode.Get("tags"), sonicx.WrapString),
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *pygmalionFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	// Extract the owner node
	ownerNode := metadataBinder.GetByPath("character", "owner")

	// Return the creator info
	return &models.CreatorInfo{
		Nickname:   ownerNode.Get("displayName").String(),
		Username:   ownerNode.Get("username").String(),
		PlatformID: ownerNode.Get("id").String(),
	}, nil
}

// FetchBookResponses fetches the book responses from the source
func (f *pygmalionFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	// Fetch book responses
	bookResponses, err := f.fetchBookResponses(metadataBinder.CharacterID)
	if err != nil {
		return nil, err
	}
	// Extract the lorebooks node
	lorebooksNode := bookResponses.Get("lorebooks")
	if lorebooksNode.TypeSafe() != ast.V_ARRAY {
		return &fetcher.EmptyBookBinder, nil
	}
	// Extract the book array
	bookArray, err := lorebooksNode.ArrayUseNode()
	if err != nil {
		return nil, err
	}
	// If the array is empty, return an empty binder
	if len(bookArray) == 0 {
		return &fetcher.EmptyBookBinder, nil
	}

	// Parse the book responses
	parsedResponses := make([]fetcher.JsonResponse, len(bookArray))
	bookUpdateTime := timestamp.Nano(0)
	for index, bookResult := range bookArray {
		// Parse the book response
		bookResponse := sonicx.Of(bookResult)
		// Extract the updatedAt field
		updatedAt := bookResponse.Get("updatedAt").Integer64()
		// Save the parsed response
		parsedResponses[index] = bookResponse
		// Update the book update time
		bookUpdateTime = max(bookUpdateTime, timestamp.ConvertToNano(timestamp.Seconds(updatedAt)))
	}

	// Return the binder
	return &fetcher.BookBinder{
		Responses:  parsedResponses,
		UpdateTime: bookUpdateTime,
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *pygmalionFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Fetch the character card
	characterCard, err := f.fetchCharacterCard(binder)
	if err != nil {
		return nil, err
	}

	// Parse the book responses
	characterCard.Sheet.CharacterBook = f.parseBookResponses(binder)

	// Return the character card
	return characterCard, nil
}

// fetchCharacterCard fetches the character card from the source
func (f *pygmalionFetcher) fetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Fetch the avatar
	avatarUrl := binder.GetByPath("character", "avatarUrl").String()
	rawCard, err := png.FromURL(f.client, avatarUrl).LastVersion().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Fetch the JSON sheet from the API
	bytes, err := reqx.Bytes(
		f.client.R().
			SetContentType(reqx.JsonApplicationContentType).
			Get(fmt.Sprintf(pygmalionCardExportURL, binder.CharacterID)),
	)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchCardDataErr)
	}

	// Optimization to remove the prefix `{character:` and suffix `}` from the byte response without processing
	characterCard.Sheet, err = character.FromBytes(bytes[13 : len(bytes)-1])
	// If the card is nil, then the export failed (error is treated upstream)
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedCardDataErr)
	}

	// Set empty creator notes (description == tagline, and the patcher will set the creator notes to the tagline)
	characterCard.Sheet.CreatorNotes = property.String("")

	// Return the parsed PNG card
	return characterCard, nil
}

// parseBookResponses merges the book responses into a single book
func (f *pygmalionFetcher) parseBookResponses(binder *fetcher.Binder) *character.Book {
	// Create the book merger
	bookMerger := character.NewBookMerger()

	// Parse the book responses
	var books []*character.Book
	for _, bookResponse := range binder.Responses {
		// Parse the book
		var pygBook pygmalionBook
		// Unmarshall the book into the pygmalionBook struct
		if err := sonicx.Config.UnmarshalFromString(bookResponse.Raw(), &pygBook); err != nil {
			// Log the error and continue
			log.Warn().Err(err).
				Str(trace.SOURCE, string(f.sourceID)).
				Str(trace.URL, binder.DirectURL).
				Msg("Could not parse book")
			continue
		}
		// Convert the pygmalionBook into a character.Book and append it to the books slice
		books = append(books, pygBook.convert())
	}

	// Sort the books by name
	slices.SortFunc(books, func(a, b *character.Book) int {
		return strings.Compare(string(a.Name), string(b.Name))
	})

	// Merge the books
	for _, book := range books {
		bookMerger.AppendBook(book)
	}

	// Return the merged book
	return bookMerger.Build()
}

// fetchBookResponses fetches the book responses from the source
func (f *pygmalionFetcher) fetchBookResponses(characterID string) (fetcher.JsonResponse, error) {
	// Create request body
	requestBodyBytes, _ := sonicx.Config.Marshal(
		map[string]string{
			"characterId": characterID,
		},
	)

	// Send request
	httpResponse, err := f.client.AR(f.serviceLabel).
		SetContentType(reqx.JsonApplicationContentType).
		SetBody(requestBodyBytes).
		Post(pygmalionLinkedBookURL)
	// If the response is 400, return an empty binder (request not authorized)
	if httpResponse != nil && httpResponse.StatusCode == 400 {
		return sonicx.Empty, fetcher.NewError(err, fetcher.InvalidCredentialsErr)
	}

	// Get the response as a string
	response, err := reqx.String(httpResponse, err)
	if err != nil {
		return sonicx.Empty, fetcher.NewError(err, fetcher.FetchBookDataErr)
	}

	// Parse the response JSON
	wrap, err := sonic.GetFromString(response)
	if err != nil {
		return sonicx.Empty, fetcher.NewError(err, fetcher.MalformedBookDataErr)
	}

	// Return the parsed response
	return sonicx.Of(wrap), nil
}

// refreshBearerToken refreshes the bearer token using the provided identity
func (f *pygmalionFetcher) refreshBearerToken(c *reqx.Client, identity cred.Identity) (string, error) {
	// Create credentials map
	credentialsMap := map[string]string{
		pygmalionAuthUsernameField: identity.User,
		pygmalionAuthPasswordField: identity.Secret,
	}

	// Send request
	response, err := reqx.String(
		c.R().
			SetContentType("application/x-www-form-urlencoded").
			SetHeaders(f.headers).
			SetFormData(credentialsMap).
			Post(pygmalionAuthURL),
	)
	if err != nil {
		return "", err
	}

	// Extract the token
	tokenWrap, err := sonicx.GetFromString(response, "result", "id_token")
	if err != nil {
		return "", err
	}

	// Return the token string
	return tokenWrap.String(), nil
}

// pygmalionBook struct for parsing Pygmalion book responses
type pygmalionBook struct {
	Name        property.String      `json:"name"`
	Description property.String      `json:"description"`
	Entries     []pygmalionBookEntry `json:"entries"`
}

// convert converts the pygmalionBook into a character.Book
func (p *pygmalionBook) convert() *character.Book {
	// Create the book
	b := character.DefaultBook()
	// Set the book fields
	b.Name = p.Name
	b.Description = p.Description
	b.Entries = make([]*character.BookEntry, len(p.Entries))

	// Convert the entries
	for index := range p.Entries {
		b.Entries[index] = p.Entries[index].convert()
	}

	// Return the book
	return b
}

// pygmalionBookEntry struct for parsing Pygmalion book entries
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

// convert converts the pygmalionBookEntry into a character.BookEntry
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
