package types

type Subject int

const (
	Biologie Subject = iota
	Chemie
	Deutsch
	Englisch
	Franzoesisch
	Geographie
	Geschichte
	Informatik
	Mathematik
	PB
	Physik
	Seminarkurs
	Spanisch
	Sport
	Technik
	WAT
)

func (subject Subject) String() string {
	switch subject {
	case Biologie:
		return "Biologie"
	case Chemie:
		return "Chemie"
	case Deutsch:
		return "Deutsch"
	case Englisch:
		return "Englisch"
	case Geographie:
		return "Geographie"
	case Franzoesisch:
		return "Französisch"
	case Geschichte:
		return "Geschichte"
	case Informatik:
		return "Informatik"
	case Mathematik:
		return "Mathematik"
	case PB:
		return "PB"
	case Physik:
		return "Physik"
	case Seminarkurs:
		return "Seminarkurs"
	case Spanisch:
		return "Spanisch"
	case Sport:
		return "Sport"
	case Technik:
		return "Technik"
	case WAT:
		return "WAT"
	default:
		return "Unbekannt"
	}
}

func (subject Subject) Short() string {
	switch subject {
	case Biologie:
		return "BI"
	case Chemie:
		return "CH"
	case Deutsch:
		return "DE"
	case Englisch:
		return "EN"
	case Geographie:
		return "EK"
	case Geschichte:
		return "GE"
	case Franzoesisch:
		return "FR"
	case Informatik:
		return "IF"
	case Mathematik:
		return "MA"
	case PB:
		return "PB"
	case Physik:
		return "PH"
	case Seminarkurs:
		return "SK"
	case Spanisch:
		return "SN"
	case Sport:
		return "SP"
	case Technik:
		return "TE"
	case WAT:
		return "WAT"
	default:
		return "??"
	}
}
