package source

import (
	"github.com/r3dpixel/toolkit/slicesx"
)

// ID type for source identifiers
type ID string

// Enum for all included sources
const (
	Local           ID = "Local"
	CharacterTavern ID = "CharacterTavern"
	ChubAI          ID = "ChubAI"
	NyaiMe          ID = "NyaiMe"
	PepHop          ID = "PepHop"
	Pygmalion       ID = "Pygmalion"
	WyvernChat      ID = "WyvernChat"
	JannyAI         ID = "JannyAI"
	AICC            ID = "AICC"
	//RisuAI          ID = "RisuAI"
)

// Values returns all source identifiers
func (ID) Values() []string {
	return []string{
		string(Local),
		string(CharacterTavern),
		string(ChubAI),
		string(NyaiMe),
		string(PepHop),
		string(Pygmalion),
		string(WyvernChat),
		string(JannyAI),
		string(AICC),
		//string(RisuAI),
	}
}

// All returns all source identifiers
func All() []ID {
	return slicesx.Map(ID("").Values(), func(id string) ID {
		return ID(id)
	})
}
