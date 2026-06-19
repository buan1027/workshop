package domain

import "testing"

func TestValidateWriteAcceptsValidInputAndTrimsStrings(t *testing.T) {
	input := GebrauchtwagenWrite{
		FIN:            "WVWZZZ1JZXW000001",
		Marke:          " VW ",
		Modell:         " Golf ",
		Fahrzeugklasse: "KOMPAKTKLASSE",
		Kraftstoffart:  "BENZIN",
		Schadenfrei:    true,
		Kilometerstand: 12000,
	}

	problems := ValidateWrite(&input)

	if len(problems) != 0 {
		t.Fatalf("expected no validation problems, got %v", problems)
	}
	if input.Marke != "VW" || input.Modell != "Golf" {
		t.Fatalf("expected strings to be trimmed, got marke=%q modell=%q", input.Marke, input.Modell)
	}
}

func TestValidateWriteRejectsInvalidInput(t *testing.T) {
	input := GebrauchtwagenWrite{
		Marke:          " ",
		Modell:         "",
		Fahrzeugklasse: "RAUMSCHIFF",
		Kraftstoffart:  "DAMPF",
		Kilometerstand: -1,
	}

	problems := ValidateWrite(&input)

	if len(problems) != 6 {
		t.Fatalf("expected 6 validation problems, got %d: %v", len(problems), problems)
	}
}

func TestValidateWriteAcceptsOptionalRelations(t *testing.T) {
	input := GebrauchtwagenWrite{
		FIN:            "WVWZZZ1JZXW000001",
		Marke:          "VW",
		Modell:         "Golf",
		Fahrzeugklasse: "KOMPAKTKLASSE",
		Kraftstoffart:  "BENZIN",
		Schadenfrei:    true,
		Kilometerstand: 12000,
		Standort:       &StandortWrite{PLZ: "76131", Ort: "Karlsruhe"},
		Hauptuntersuchung: &HauptuntersuchungWrite{
			Pruefdatum:        "2025-06-01",
			GueltigBis:        "2027-06-01",
			Prueforganisation: "TUEV",
			Status:            "BESTANDEN",
		},
		Schaeden: []SchadenWrite{{
			Bezeichnung:        "Kratzer",
			Beschreibung:       "Kleiner Lackkratzer",
			Feststellungsdatum: "2024-11-10",
		}},
	}

	problems := ValidateWrite(&input)

	if len(problems) != 0 {
		t.Fatalf("expected no validation problems, got %v", problems)
	}
}
