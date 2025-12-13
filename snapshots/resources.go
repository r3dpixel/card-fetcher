package snapshots

import (
	"embed"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
)

// embeddedResources contains the resources used to test the router.
//
//go:embed cards
var embeddedResources embed.FS

// resourceURLs contains the URLs of the resources used to test the router.
var resourceURLs = map[source.ID][]string{
	source.CharacterTavern: {"character-tavern.com/character/animatedspell/Violete%20V4"},
	source.ChubAI: {
		"chub.ai/characters/Boy_Next_Door/mina-the-girl-next-door-421a72119482",
		"chub.ai/characters/Xenton05/vesperine-2e090284ce72",
		"chub.ai/characters/Decent_Coast/your-older-orc-step-sister-61c8107a3862",
	},
	source.NyaiMe:     {"nyai.me/ai/bots/Beatrix_lj"},
	source.PepHop:     {"pephop.ai/characters/eaa39561-1d4d-4d10-a87c-77f038f8b211"},
	source.Pygmalion:  {"pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5"},
	source.WyvernChat: {"wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4"},
	source.JannyAI:    {"jannyai.com/characters/421439ad-de63-4448-bc9b-c2c75cedb0af"},
	source.AICC: {
		"aicharactercards.com/charactercards/confession/linux4life/nora",
		"aicharactercards.com/charactercards/free-use/tempuser213525312/evelyn",
	},
	//source.RisuAI:          {"-"},
}

// File extensions used for the resources
const (
	cardExtension = ".card"
	jsonExtension = ".json"
)

// GetResourceMap returns a copy of the resource map
func GetResourceMap() map[source.ID][]string {
	resourceMap := make(map[source.ID][]string, len(resourceURLs))
	for sourceID, urls := range resourceURLs {
		resourceMap[sourceID] = slices.Clone(urls)
	}
	return resourceMap
}

// GetResourceURLs returns the URLs of the resources for the given source
func GetResourceURLs(sourceID source.ID) ([]string, bool) {
	urls, ok := resourceURLs[sourceID]
	return slices.Clone(urls), ok
}

// GetResourcePath returns the path of the resource for the given source and index
func GetResourcePath(sourceID source.ID, index int) string {
	return filepath.Join("cards", string(sourceID)+"_"+strconv.Itoa(index))
}

// GetResourceCardPath returns the path of the character card resource for the given source and index
func GetResourceCardPath(sourceID source.ID, index int) string {
	return GetResourcePath(sourceID, index) + cardExtension
}

// GetResourceJsonPath returns the path of the character sheet resource for the given source and index
func GetResourceJsonPath(sourceID source.ID, index int) string {
	return GetResourcePath(sourceID, index) + jsonExtension
}

// GetResourceCard returns the character card resource for the given source and index
func GetResourceCard(sourceID source.ID, index int) (*png.CharacterCard, error) {
	// Create the character card path
	cardPath := GetResourceCardPath(sourceID, index)
	// Read the character card
	data, err := embeddedResources.ReadFile(cardPath)
	if err != nil {
		return nil, err
	}

	// Decode the character card from the PNG
	rawCard, err := png.FromBytes(data).First().Get()
	if err != nil {
		return nil, err
	}

	// Return the decoded character card
	return rawCard.Decode()
}

// GetResourceCards returns the character card resource for the given source and index
func GetResourceCards(sourceID source.ID) ([]*png.CharacterCard, error) {
	// Get the URLs of the resources
	urls, _ := GetResourceURLs(sourceID)
	// Create the character cards slice
	cards := make([]*png.CharacterCard, len(urls))

	// Iterate over the URLs and decode the character cards (the order of the URLs corresponds to the order of the cards)
	for index := range urls {
		// Get the character card resource
		cardPath := GetResourceCardPath(sourceID, index)
		// Read the character card
		data, err := embeddedResources.ReadFile(cardPath)
		if err != nil {
			return nil, err
		}

		// Decode the character card from the PNG
		rawCard, err := png.FromBytes(data).First().Get()
		if err != nil {
			return nil, err
		}

		// Decode the character card
		characterCard, err := rawCard.Decode()
		if err != nil {
			return nil, err
		}

		// Add the decoded character card to the slice
		cards[index] = characterCard
	}

	// Return the slice of decoded character cards
	return cards, nil
}

// GetResourceJson returns the character sheet resource for the given source and index
func GetResourceJson(sourceID source.ID, index int) (*character.Sheet, error) {
	// Create the character sheet JSON path
	jsonPath := GetResourceJsonPath(sourceID, index)
	// Read the character sheet JSON
	data, err := embeddedResources.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	// Decode the character sheet from the JSON
	sheet, err := character.FromBytes(data)
	if err != nil {
		return nil, err
	}

	// Return the decoded character sheet
	return sheet, nil
}
