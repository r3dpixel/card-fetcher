package impl

import (
	"fmt"
	"path"
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
	"github.com/rs/zerolog/log"
)

const (
	wyvernDomain     string = "wyvern.chat"                           // Domain for WyvernChat
	wyvernPath       string = "characters/"                           // Path for WyvernChat
	wyvernApiURL     string = "https://api.wyvern.chat/characters/%s" // API URL for WyvernChat
	wyvernDateFormat string = time.RFC3339Nano                        // Date Format for WyvernChat
)

// WyvernChatBuilder builder for WyvernChat fetcher
type WyvernChatBuilder struct{}

// Build creates a new WyvernChat fetcher
func (b WyvernChatBuilder) Build(client *reqx.Client) fetcher.Fetcher {
	return NewWyvernChatFetcher(client)
}

// wyvernChatFetcher WyvernChat fetcher implementation
type wyvernChatFetcher struct {
	BaseFetcher
}

// NewWyvernChatFetcher creates a new WyvernChat fetcher
func NewWyvernChatFetcher(client *reqx.Client) fetcher.Fetcher {
	mainURL := path.Join(wyvernDomain, wyvernPath)
	impl := &wyvernChatFetcher{
		BaseFetcher: BaseFetcher{
			client:    client,
			sourceID:  source.WyvernChat,
			sourceURL: "app." + wyvernDomain,
			directURL: "app." + mainURL,
			mainURL:   mainURL,
			baseURLs: []fetcher.BaseURL{
				{Domain: wyvernDomain, Path: wyvernPath},
			},
		},
	}
	impl.Extends(impl)
	return impl
}

// FetchMetadataResponse fetches the metadata response from the source for the given characterID
func (f *wyvernChatFetcher) FetchMetadataResponse(characterID string) (*req.Response, error) {
	metadataUrl := fmt.Sprintf(wyvernApiURL, characterID)
	return f.client.R().Get(metadataUrl)
}

// CreateBinder creates a MetadataBinder from the metadata response
func (f *wyvernChatFetcher) CreateBinder(characterID string, metadataResponse string) (*fetcher.MetadataBinder, error) {
	return f.CreateBinderFromJSON(characterID, metadataResponse, "id")
}

// FetchCardInfo fetches the card info from the source
func (f *wyvernChatFetcher) FetchCardInfo(metadataBinder *fetcher.MetadataBinder) (*models.CardInfo, error) {
	return &models.CardInfo{
		NormalizedURL: metadataBinder.NormalizedURL,
		DirectURL:     f.DirectURL(metadataBinder.CharacterID),
		PlatformID:    strings.TrimPrefix(metadataBinder.CharacterID, symbols.Underscore),
		CharacterID:   metadataBinder.CharacterID,
		Name:          metadataBinder.Get("chat_name").String(),
		Title:         metadataBinder.Get("name").String(),
		Tagline:       metadataBinder.Get("tagline").String(),
		CreateTime:    timestamp.ParseF(wyvernDateFormat, metadataBinder.Get("created_at").String(), trace.URL, metadataBinder.NormalizedURL),
		UpdateTime:    timestamp.ParseF(wyvernDateFormat, metadataBinder.Get("updated_at").String(), trace.URL, metadataBinder.NormalizedURL),
		IsForked:      stringsx.IsNotBlank(metadataBinder.GetByPath("forked_from", "id").RefString()),
		Tags:          models.TagsFromJsonArray(metadataBinder.Get("tags"), sonicx.WrapString),
	}, nil
}

// FetchCreatorInfo fetches the creator info from the source
func (f *wyvernChatFetcher) FetchCreatorInfo(metadataBinder *fetcher.MetadataBinder) (*models.CreatorInfo, error) {
	// Extract the creator node
	creatorNode := metadataBinder.Get("creator")
	// Extract the displayName
	displayName := creatorNode.Get("displayName").String()

	// Return the creator info
	return &models.CreatorInfo{
		Nickname:   displayName,
		Username:   displayName,
		PlatformID: creatorNode.Get("id").String(),
	}, nil
}

// FetchBookResponses fetches the book responses from the source
func (f *wyvernChatFetcher) FetchBookResponses(metadataBinder *fetcher.MetadataBinder) (*fetcher.BookBinder, error) {
	// Initialize the book update time as ZERO
	bookUpdateTime := timestamp.Nano(0)
	// Extract the lorebooks node
	lorebooksNode := metadataBinder.Get("lorebooks")
	// Extract the lorebooks array
	array, _ := lorebooksNode.ArrayUseNode()
	// Iterate over the lorebooks array
	for _, lorebookNode := range array {
		// Update the book update time with the latest book update time
		bookUpdateTime = max(
			bookUpdateTime,
			timestamp.ParseF(wyvernDateFormat, sonicx.Of(lorebookNode).Get("updated_at").String(), trace.URL, metadataBinder.NormalizedURL),
		)
	}
	// Return the book binder
	return &fetcher.BookBinder{
		UpdateTime: bookUpdateTime,
	}, nil
}

// FetchCharacterCard fetches the character card from the source
func (f *wyvernChatFetcher) FetchCharacterCard(binder *fetcher.Binder) (*png.CharacterCard, error) {
	// Extract the avatar URL
	avatarURL := binder.Get("avatar").String()

	// Fetch the character avatar from the API
	rawCard, err := png.FromURL(f.client, avatarURL).LastVersion().Get()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.FetchAvatarErr)
	}

	// Decode the character card
	characterCard, err := rawCard.Decode()
	if err != nil {
		return nil, fetcher.NewError(err, fetcher.DecodeErr)
	}

	// Parse the character sheet
	var wSheet wyvernSheet
	if err := sonicx.Config.UnmarshalFromString(binder.Raw(), &wSheet); err != nil {
		return nil, fetcher.NewError(err, fetcher.MalformedCardDataErr)
	}

	// Update the character sheet fields
	wSheet.fillIn(characterCard.Sheet)

	// Return the character card
	return characterCard, nil
}

// wyvernSheet WyvernChat sheet structure
type wyvernSheet struct {
	Description             property.String      `json:"description"`
	Personality             property.String      `json:"personality"`
	MessageExamples         property.String      `json:"mes_example"`
	CreatorNotes            property.String      `json:"creator_notes"`
	PostHistoryInstructions property.String      `json:"post_history_instructions"`
	PreHistoryInstructions  property.String      `json:"pre_history_instructions"`
	FirstMessage            property.String      `json:"first_mes"`
	AlternateGreetings      property.StringArray `json:"alternate_greetings"`
	Scenario                property.String      `json:"scenario"`
	CharacterNote           property.String      `json:"character_note"`
	SharedInfo              property.String      `json:"shared_info"`

	SecretFields []any             `json:"secretFields"`
	Secrets      map[string]any    `json:"secrets"`
	Fields       map[string]any    `json:"fields"`
	Scripts      map[string]any    `json:"scripts"`
	Commands     []any             `json:"commands"`
	Personas     []any             `json:"personas"`
	Lexicon      []wyvernBookEntry `json:"lexicon"`
	Lorebooks    wyvernLoreBooks   `json:"lorebooks"`
}

// wyvernLoreBooks WyvernChat lorebooks structure
type wyvernLoreBooks []wyvernBook

// UnmarshalJSON custom unmarshal for wyvernLoreBooks
func (w *wyvernLoreBooks) UnmarshalJSON(data []byte) error {
	// arr holds the unmarshalled
	var arr []wyvernBook

	// Unmarshal the data into the array
	if err := sonicx.Config.Unmarshal(data, &arr); err != nil {
		// Log the error and return nil
		dataLen := len(data)
		if dataLen > 256 {
			dataLen = 256
		}
		log.Error().Err(err).Str("data", string(data[:dataLen])).Msg("could not parse lorebooks")
		*w = nil
		return nil
	}
	// Set the unmarshalled array to the pointer
	*w = arr

	// Return nil
	return nil
}

// fillIn fills the character sheet fields from the wyvernSheet structure
func (w *wyvernSheet) fillIn(sheet *character.Sheet) {
	// Set the sheet fields
	sheet.Description = w.Description
	sheet.Personality = w.Personality
	sheet.MessageExamples = w.MessageExamples
	sheet.CreatorNotes = w.CreatorNotes
	sheet.PostHistoryInstructions = w.PostHistoryInstructions
	sheet.SystemPrompt = w.PreHistoryInstructions
	sheet.FirstMessage = w.FirstMessage
	sheet.AlternateGreetings = w.AlternateGreetings
	sheet.Scenario = w.Scenario
	sheet.Content.DepthPrompt.Prompt = string(w.CharacterNote)
	sheet.Content.DepthPrompt.Depth = character.DefaultDepth

	// Extract the shared info and add it to the creator notes and group greeting messages
	sharedInfo := strings.TrimSpace(string(w.SharedInfo))
	if stringsx.IsNotBlank(sharedInfo) {
		sheet.GroupGreetings = property.StringArray{sharedInfo}
		sheet.CreatorNotes = property.String(stringsx.JoinNonBlank(character.CreatorNotesSeparator, string(sheet.CreatorNotes), sharedInfo))
	}

	// Add the secret fields, secrets, scripts, commands and personas to the extensions, key prefixed with "wyvern_"
	sheet.Extensions = make(map[string]any)
	if len(w.SecretFields) > 0 {
		sheet.Extensions["wyvern_secret_fields"] = w.SecretFields
	}
	if len(w.Secrets) > 0 {
		sheet.Extensions["wyvern_secrets"] = w.Secrets
	}
	if len(w.Scripts) > 0 {
		sheet.Extensions["wyvern_fields"] = w.Fields
	}
	if len(w.Scripts) > 0 {
		sheet.Extensions["wyvern_scripts"] = w.Scripts
	}
	if len(w.Commands) > 0 {
		sheet.Extensions["wyvern_commands"] = w.Commands
	}
	if len(w.Personas) > 0 {
		sheet.Extensions["wyvern_personas"] = w.Personas
	}

	// Merge the lorebooks and lexicon into a single book
	bookMerger := character.NewBookMerger()
	// Iterate over the lorebooks, convert them to books and append them to the book merger
	for _, book := range w.Lorebooks {
		bookMerger.AppendBook(book.convert())
	}
	// Iterate over the lexicon entries, convert them to entries and append them to the book merger
	for index := range w.Lexicon {
		bookMerger.AppendEntry(w.Lexicon[index].convert())
	}
	// Set the character book to the book merger result
	sheet.Content.CharacterBook = bookMerger.Build()
}

// wyvernBook WyvernChat book structure
type wyvernBook struct {
	Name              property.String   `json:"name"`
	Description       property.String   `json:"description"`
	ScanDepth         property.Integer  `json:"scan_depth"`
	TokenBudget       property.Integer  `json:"token_budget"`
	RecursiveScanning property.Bool     `json:"recursive_scanning"`
	Extensions        map[string]any    `json:"extensions"`
	Entries           []wyvernBookEntry `json:"entries"`
}

// convert converts the wyvernBook structure to a character.Book
func (w *wyvernBook) convert() *character.Book {
	// Create a new character book
	book := character.DefaultBook()
	// Set the book fields
	book.Name = w.Name
	book.Description = w.Description
	book.ScanDepth = w.ScanDepth
	book.TokenBudget = w.TokenBudget
	book.RecursiveScanning = w.RecursiveScanning
	book.Extensions = w.Extensions
	book.Entries = make([]*character.BookEntry, len(w.Entries))

	// Convert the entries
	for index := range w.Entries {
		book.Entries[index] = w.Entries[index].convert()
	}

	// Return the book
	return book
}

// wyvernBookEntry WyvernChat book entry structure
type wyvernBookEntry struct {
	ID               property.Union           `json:"entry_id"`
	Keys             property.StringArray     `json:"keys"`
	Content          property.String          `json:"content"`
	Extensions       map[string]any           `json:"extensions"`
	Enabled          property.Bool            `json:"enabled"`
	CaseSensitive    property.Bool            `json:"case_sensitive"`
	InsertionOrder   property.Integer         `json:"insertion_order"`
	Name             property.String          `json:"name"`
	Priority         property.Integer         `json:"priority"`
	Comment          property.String          `json:"comment"`
	SecondaryKeys    property.StringArray     `json:"secondary_keys"`
	Constant         property.Bool            `json:"constant"`
	Position         *property.LorePosition   `json:"position"`
	ScanPersona      property.Bool            `json:"scan_persona"`
	MatchWholeWords  property.Bool            `json:"whole_words_only"`
	SelectiveLogic   *property.SelectiveLogic `json:"key_logic"`
	Delay            property.Integer         `json:"delay"`
	Sticky           property.Integer         `json:"sticky"`
	Cooldown         property.Integer         `json:"cooldown"`
	ActivationChance *property.Float          `json:"activation_chance"`
	CustomFields     map[string]any           `json:"custom_fields"`
}

// convert converts the wyvernBookEntry structure to a character.BookEntry
func (w *wyvernBookEntry) convert() *character.BookEntry {
	// Create a new character book entry
	entry := character.DefaultBookEntry()

	// Set the entry fields
	entry.ID = w.ID
	entry.Keys = w.Keys
	entry.Content = w.Content
	entry.RawExtensions = w.Extensions
	entry.Enabled = w.Enabled
	entry.Extensions.CaseSensitive = w.CaseSensitive
	entry.Extensions.Depth = w.InsertionOrder
	entry.Name = w.Name
	entry.InsertionOrder = w.Priority
	entry.Comment = w.Comment
	entry.SecondaryKeys = w.SecondaryKeys
	entry.Constant = w.Constant
	entry.Extensions.LorePosition.SetIfPropertyPtr(w.Position)
	entry.Extensions.MatchWholeWords = w.MatchWholeWords
	entry.Extensions.SelectiveLogic.SetIfPropertyPtr(w.SelectiveLogic)
	entry.Extensions.Delay = w.Delay
	entry.Extensions.Sticky = w.Sticky
	entry.Extensions.Cooldown = w.Cooldown
	entry.Extensions.Probability.SetIfPropertyPtr(w.ActivationChance)

	// Add the custom fields to the raw extensions, key prefixed with "wyvern_"
	if len(w.CustomFields) > 0 {
		if entry.RawExtensions == nil {
			entry.RawExtensions = make(map[string]any)
		}
		entry.RawExtensions["wyvern_custom_fields"] = w.CustomFields
	}

	// Return the entry
	return entry
}
