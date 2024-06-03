package model

// Todo model
type Bewertung struct {
	Vorname       string
	Nachname      string
	ID            int
	HvPunkte      float64
	HvProzent     float64
	HvNote        int
	LvPunkte      float64
	LvProzent     float64
	LvNote        int
	GesamtProzent float64
	GesamtNote    int
	Gewertet      bool
}

type MaxPunkte struct {
	HvMax        float64
	LvMax        float64
	HvGewichtung float64
	LvGewichtung float64
}
