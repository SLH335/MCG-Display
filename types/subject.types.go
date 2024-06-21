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
	Kunst
	Latein
	Mathematik
	Musik
	PB
	Physik
	Seminarkurs
	Spanisch
	Sport
	Technik
	WAT
)

func (subject Subject) String() string {
	return []string{
		"Biologie",
		"Chemie",
		"Deutsch",
		"Englisch",
		"Französisch",
		"Geographie",
		"Geschichte",
		"Informatik",
		"Kunst",
		"Latein",
		"Mathematik",
		"Musik",
		"PB",
		"Physik",
		"Seminarkurs",
		"Spanisch",
		"Sport",
		"Technik",
		"WAT",
	}[subject]
}

func (subject Subject) Short() string {
	return []string{
		"BI",
		"CH",
		"DE",
		"EN",
		"FR",
		"EK",
		"GE",
		"IF",
		"KU",
		"LA",
		"MA",
		"MU",
		"PB",
		"PH",
		"SK",
		"SN",
		"SP",
		"TE",
		"AL",
	}[subject]
}
