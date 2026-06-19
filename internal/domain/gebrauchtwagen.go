package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Gebrauchtwagen struct {
	ID             int    `json:"id"`
	FIN            string `json:"fin"`
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
	FIN               string                  `json:"fin"`
	Marke             string                  `json:"marke"`
	Modell            string                  `json:"modell"`
	Fahrzeugklasse    string                  `json:"fahrzeugklasse"`
	Kraftstoffart     string                  `json:"kraftstoffart"`
	Schadenfrei       bool                    `json:"schadenfrei"`
	Kilometerstand    int                     `json:"kilometerstand"`
	Standort          *StandortWrite          `json:"standort,omitempty"`
	Hauptuntersuchung *HauptuntersuchungWrite `json:"hauptuntersuchung,omitempty"`
	Schaeden          []SchadenWrite          `json:"schaeden,omitempty"`
}

type StandortWrite struct {
	PLZ string `json:"plz"`
	Ort string `json:"ort"`
}

type HauptuntersuchungWrite struct {
	Pruefdatum        string `json:"pruefdatum"`
	GueltigBis        string `json:"gueltigBis"`
	Prueforganisation string `json:"prueforganisation"`
	Status            string `json:"status"`
}

type SchadenWrite struct {
	Bezeichnung        string `json:"bezeichnung"`
	Beschreibung       string `json:"beschreibung"`
	Feststellungsdatum string `json:"feststellungsdatum"`
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

var huStatus = map[string]bool{
	"BESTANDEN": true, "NICHT_BESTANDEN": true, "AUSSTEHEND": true,
}

func ValidateWrite(input *GebrauchtwagenWrite) []string {
	var problems []string
	input.FIN = strings.TrimSpace(input.FIN)
	input.Marke = strings.TrimSpace(input.Marke)
	input.Modell = strings.TrimSpace(input.Modell)
	input.Fahrzeugklasse = strings.TrimSpace(input.Fahrzeugklasse)
	input.Kraftstoffart = strings.TrimSpace(input.Kraftstoffart)

	if len(input.FIN) != 17 {
		problems = append(problems, "fin muss genau 17 Zeichen lang sein")
	}
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

	problems = append(problems, validateStandort(input.Standort)...)
	problems = append(problems, validateHauptuntersuchung(input.Hauptuntersuchung)...)
	for index := range input.Schaeden {
		problems = append(problems, validateSchaden(index, &input.Schaeden[index])...)
	}

	return problems
}

func IsValidFahrzeugklasse(value string) bool {
	return value == "" || fahrzeugklassen[value]
}

func IsValidKraftstoffart(value string) bool {
	return value == "" || kraftstoffarten[value]
}

func validateStandort(input *StandortWrite) []string {
	if input == nil {
		return nil
	}

	var problems []string
	input.PLZ = strings.TrimSpace(input.PLZ)
	input.Ort = strings.TrimSpace(input.Ort)

	if input.PLZ == "" {
		problems = append(problems, "standort.plz darf nicht leer sein")
	}
	if input.Ort == "" {
		problems = append(problems, "standort.ort darf nicht leer sein")
	}

	return problems
}

func validateHauptuntersuchung(input *HauptuntersuchungWrite) []string {
	if input == nil {
		return nil
	}

	var problems []string
	input.Pruefdatum = strings.TrimSpace(input.Pruefdatum)
	input.GueltigBis = strings.TrimSpace(input.GueltigBis)
	input.Prueforganisation = strings.TrimSpace(input.Prueforganisation)
	input.Status = strings.TrimSpace(input.Status)

	if !isISODate(input.Pruefdatum) {
		problems = append(problems, "hauptuntersuchung.pruefdatum muss ein Datum im Format YYYY-MM-DD sein")
	}
	if !isISODate(input.GueltigBis) {
		problems = append(problems, "hauptuntersuchung.gueltigBis muss ein Datum im Format YYYY-MM-DD sein")
	}
	if input.Prueforganisation == "" {
		problems = append(problems, "hauptuntersuchung.prueforganisation darf nicht leer sein")
	}
	if !huStatus[input.Status] {
		problems = append(problems, "hauptuntersuchung.status ist ungueltig")
	}

	return problems
}

func validateSchaden(index int, input *SchadenWrite) []string {
	var problems []string
	input.Bezeichnung = strings.TrimSpace(input.Bezeichnung)
	input.Beschreibung = strings.TrimSpace(input.Beschreibung)
	input.Feststellungsdatum = strings.TrimSpace(input.Feststellungsdatum)
	prefix := "schaeden." + strconv.Itoa(index)

	if input.Bezeichnung == "" {
		problems = append(problems, prefix+".bezeichnung darf nicht leer sein")
	}
	if input.Beschreibung == "" {
		problems = append(problems, prefix+".beschreibung darf nicht leer sein")
	}
	if !isISODate(input.Feststellungsdatum) {
		problems = append(problems, prefix+".feststellungsdatum muss ein Datum im Format YYYY-MM-DD sein")
	}

	return problems
}

func isISODate(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
