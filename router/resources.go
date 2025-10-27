package router

import (
	"embed"
	"maps"
	"path/filepath"

	"github.com/r3dpixel/card-fetcher/source"
	"github.com/r3dpixel/card-parser/character"
	"github.com/r3dpixel/card-parser/png"
)

//go:embed snapshots
var embeddedResources embed.FS

var resourceURLs = map[source.ID]string{
	source.CharacterTavern: "character-tavern.com/character/animatedspell/Violete%20V4",
	source.ChubAI:          "chub.ai/characters/Enkob/maxine-flirty-warrior-of-the-explorers-guild-d1eb387be026",
	source.NyaiMe:          "nyai.me/ai/bots/Beatrix_lj",
	source.PepHop:          "pephop.ai/characters/eaa39561-1d4d-4d10-a87c-77f038f8b211",
	source.Pygmalion:       "pygmalion.chat/character/d47f2f4e-0263-49f8-b872-d8fd7588dbb5",
	source.WyvernChat:      "wyvern.chat/characters/_MBnL8cfMUVNFVTBe4GNm4",
	source.JannyAI:         "jannyai.com/characters/421439ad-de63-4448-bc9b-c2c75cedb0af",
	//source.RisuAI:          "-",
	//source.AICharacterCard: "-",
}

const (
	cardExtension = ".card"
	jsonExtension = ".json"
)

func GetResourceURLs() map[source.ID]string {
	return maps.Clone(resourceURLs)
}

func GetResourceURL(sourceID source.ID) (string, bool) {
	url, ok := resourceURLs[sourceID]
	return url, ok
}

func GetResourcePath(sourceID source.ID) string {
	return filepath.Join("snapshots", string(sourceID))
}

func GetResourceCardPath(sourceID source.ID) string {
	return GetResourcePath(sourceID) + cardExtension
}

func GetResourceJsonPath(sourceID source.ID) string {
	return GetResourcePath(sourceID) + jsonExtension
}

func GetResourceCard(sourceID source.ID) (*png.CharacterCard, error) {
	cardPath := GetResourceCardPath(sourceID)
	data, err := embeddedResources.ReadFile(cardPath)
	if err != nil {
		return nil, err
	}

	rawCard, err := png.FromBytes(data).First().Get()
	if err != nil {
		return nil, err
	}

	return rawCard.Decode()
}

func GetResourceJson(sourceID source.ID) (*character.Sheet, error) {
	jsonPath := GetResourceJsonPath(sourceID)
	data, err := embeddedResources.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	sheet, err := character.FromBytes(data)
	if err != nil {
		return nil, err
	}

	return sheet, nil
}
