package source

import (
	"github.com/r3dpixel/toolkit/slicesx"
	"github.com/r3dpixel/toolkit/stringsx"
)

type ID string

// Enum for all included sources
const (
	//File            ID = "File"
	CharacterTavern ID = "CharacterTavern"
	ChubAI          ID = "ChubAI"
	NyaiMe          ID = "NyaiMe"
	PepHop          ID = "PepHop"
	Pygmalion       ID = "Pygmalion"
	WyvernChat      ID = "WyvernChat"
	//RisuAI          ID = "RisuAI"
	//AICharacterCard ID = "AICharacterCard"
)

func (ID) Values() []string {
	return []string{
		//string(File),
		string(CharacterTavern),
		string(ChubAI),
		string(NyaiMe),
		string(PepHop),
		string(Pygmalion),
		string(WyvernChat),
		//string(RisuAI),
		//string(AICharacterCard),
	}
}

func All() []ID {
	return slicesx.Map(ID(stringsx.Empty).Values(), func(id string) ID {
		return ID(id)
	})
}
