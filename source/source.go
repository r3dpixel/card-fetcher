package source

type ID string

// Enum for all included sources
const (
	File            ID = "File"
	CharacterTavern ID = "CharacterTavern"
	ChubAI          ID = "ChubAI"
	NyaiMe          ID = "NyaiMe"
	PepHop          ID = "PepHop"
	Pygmalion       ID = "Pygmalion"
	WyvernChat      ID = "WyvernChat"
	RisuAI          ID = "RisuAI"
	AICharacterCard ID = "AICharacterCard"
)

func (ID) Values() []string {
	return []string{
		string(File),
		string(CharacterTavern),
		string(ChubAI),
		string(NyaiMe),
		string(PepHop),
		string(Pygmalion),
		string(WyvernChat),
		string(RisuAI),
		string(AICharacterCard),
	}
}
