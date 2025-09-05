package router

import (
	"embed"
	"maps"
	"path/filepath"

	"github.com/r3dpixel/card-fetcher/source"
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
	//source.RisuAI:          "-",
	//source.AICharacterCard: "-",
}

func GetResourceURLs() map[source.ID]string {
	return maps.Clone(resourceURLs)
}

func GetResourceURL(sourceID source.ID) (string, bool) {
	url, ok := resourceURLs[sourceID]
	return url, ok
}

func GetResourcePath(sourceID source.ID) string {
	return filepath.Join("snapshots", string(sourceID)+".card")
}

func GetResourceCard(sourceID source.ID) (*png.CharacterCard, error) {
	path := GetResourcePath(sourceID)
	data, err := embeddedResources.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rawCard, err := png.FromBytes(data).DeepScan().Get()
	if err != nil {
		return nil, err
	}

	return rawCard.Decode()
}
