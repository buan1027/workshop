package domain

import (
	"errors"
	"strings"
)

type Gebrauchtwagen struct {
	ID             int    `json:"id"`
	Marke          string `json:"marke"`
	Modell         string `json:"modell"`
	Fahrzeugklasse string `json:"fahrzeugklasse"`
	Kraftstoffart  string `json:"kraftstoffart"`
	Schadenfrei    bool   `json:"schadenfrei"`
	Kilometerstand int    `json:"kilometerstand"`
	Version        int    `json:"version"`
}

type GebrauchtwagenDetail struct {
	Gebrauchtwagen
	Standort          *Standort          `json:"standort,omitempty"`
	Hauptuntersuchung *Hauptuntersuchung `json:"hauptuntersuchung,omitempty"`
	Schaeden          []Schaden          `json:"schaeden"`
}

type Standort struct {
	PLZ string `json:"plz"`
	Ort string `json:"ort"`
}

type Hauptuntersuchung struct {
	Pruefdatum        string `json:"pruefdatum"`
	GueltigBis        string `json:"gueltigBis"`
	Prueforganisation string `json:"prueforganisation"`
	Status            string `json:"status"`
}

type Schaden struct {
	Bezeichnung        string `json:"bezeichnung"`
	Beschreibung       string `json:"beschreibung"`
	Feststellungsdatum string `json:"feststellungsdatum"`
}

type GebrauchtwagenWrite struct {
	Marke          string `json:"marke"`
	Modell         string `json:"modell"`
	Fahrzeugklasse string `json:"fahrzeugklasse"`
	Kraftstoffart  string `json:"kraftstoffart"`
	Schadenfrei    bool   `json:"schadenfrei"`
	Kilometerstand int    `json:"kilometerstand"`
}

type SearchParams struct {
	Marke          string
	Modell         string
	Fahrzeugklasse string
	Kraftstoffart  string
	Schadenfrei    *bool
	Page           int
	Size           int
}

type Page struct {
	Data  []Gebrauchtwagen `json:"data"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

var (
	ErrNotFound        = errors.New("gebrauchtwagen not found")
	ErrVersionConflict = errors.New("version conflict")
)

var fahrzeugklassen = map[string]bool{
	"KLEINWAGEN": true, "KOMPAKTKLASSE": true, "MITTELKLASSE": true, "OBERKLASSE": true,
	"SUV": true, "KOMBI": true, "CABRIO": true, "TRANSPORTER": true,
}

var kraftstoffarten = map[string]bool{
	"BENZIN": true, "DIESEL": true, "ELEKTRO": true, "HYBRID": true, "ERDGAS": true, "WASSERSTOFF": true,
}

func ValidateWrite(input *GebrauchtwagenWrite) []string {
	var problems []string
	input.Marke = strings.TrimSpace(input.Marke)
	input.Modell = strings.TrimSpace(input.Modell)
	input.Fahrzeugklasse = strings.TrimSpace(input.Fahrzeugklasse)
	input.Kraftstoffart = strings.TrimSpace(input.Kraftstoffart)

	if input.Marke == "" {
		problems = append(problems, "marke darf nicht leer sein")
	}
	if input.Modell == "" {
		problems = append(problems, "modell darf nicht leer sein")
	}
	if !fahrzeugklassen[input.Fahrzeugklasse] {
		problems = append(problems, "fahrzeugklasse ist ungueltig")
	}
	if !kraftstoffarten[input.Kraftstoffart] {
		problems = append(problems, "kraftstoffart ist ungueltig")
	}
	if input.Kilometerstand < 0 {
		problems = append(problems, "kilometerstand muss groesser oder gleich 0 sein")
	}

	return problems
}

func IsValidFahrzeugklasse(value string) bool {
	return value == "" || fahrzeugklassen[value]
}

func IsValidKraftstoffart(value string) bool {
	return value == "" || kraftstoffarten[value]
}
