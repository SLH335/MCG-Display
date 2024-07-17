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
	LER
	Mathematik
	Musik
	PB
	Physik
	Recht
	ReligionEv
	ReligionKa
	Seminarkurs
	Spanisch
	Sport
	Technik
	WAT
	EmptySubject
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
		"LER",
		"Mathematik",
		"Musik",
		"PB",
		"Physik",
		"Recht",
		"ev. Religion",
		"kat. Religion",
		"Seminarkurs",
		"Spanisch",
		"Sport",
		"Technik",
		"WAT",
		"",
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
		"LE",
		"MA",
		"MU",
		"PB",
		"PH",
		"RL",
		"RE",
		"RK",
		"SK",
		"SN",
		"SP",
		"TE",
		"LE",
		"",
	}[subject]
}

func (subject Subject) ShortAlt() string {
	return []string{
		"Bio",
		"Che",
		"Deu",
		"Eng",
		"Fra",
		"Geo",
		"Ges",
		"Inf",
		"Kun",
		"Lat",
		"LER",
		"Mat",
		"Mus",
		"PB",
		"Phy",
		"Rec",
		"evR",
		"kaR",
		"SK",
		"Spa",
		"Spo",
		"Tec",
		"WAT",
		"",
	}[subject]
}

func (subject Subject) Variants() (variants []string) {
	if subject == PB {
		variants = append(variants, "Politische Bildung")
		variants = append(variants, "Polit. Bildung")
	}
	variants = append(variants, subject.String())
	if subject == Mathematik {
		variants = append(variants, "Mathe")
	}
	variants = append(variants, subject.ShortAlt())
	variants = append(variants, subject.Short())
	return variants
}
