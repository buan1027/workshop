package domain

import "testing"

func TestValidateWriteAcceptsValidInputAndTrimsStrings(t *testing.T) {
	input := GebrauchtwagenWrite{
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

	if len(problems) != 5 {
		t.Fatalf("expected 5 validation problems, got %d: %v", len(problems), problems)
	}
}
