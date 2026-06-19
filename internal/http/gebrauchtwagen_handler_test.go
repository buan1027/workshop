package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buan1027/workshop/internal/domain"
)

type fakeGebrauchtwagenRepository struct {
	items []domain.Gebrauchtwagen
}

func (f *fakeGebrauchtwagenRepository) List(_ context.Context, search domain.SearchParams) (domain.Page, error) {
	return domain.Page{Data: f.items, Total: len(f.items), Page: search.Page, Size: search.Size}, nil
}

func (f *fakeGebrauchtwagenRepository) FindByID(_ context.Context, id int) (domain.Gebrauchtwagen, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.Gebrauchtwagen{}, domain.ErrNotFound
}

func (f *fakeGebrauchtwagenRepository) FindDetailByID(ctx context.Context, id int) (domain.GebrauchtwagenDetail, error) {
	item, err := f.FindByID(ctx, id)
	if err != nil {
		return domain.GebrauchtwagenDetail{}, err
	}
	return domain.GebrauchtwagenDetail{
		Gebrauchtwagen: item,
		Standort:       &domain.Standort{PLZ: "76131", Ort: "Karlsruhe"},
		Schaeden: []domain.Schaden{{
			Bezeichnung:        "Kratzer",
			Beschreibung:       "Kleiner Lackkratzer",
			Feststellungsdatum: "2024-11-10",
		}},
	}, nil
}

func (f *fakeGebrauchtwagenRepository) Create(_ context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	item := domain.Gebrauchtwagen{
		ID:             len(f.items) + 1,
		Marke:          input.Marke,
		Modell:         input.Modell,
		Fahrzeugklasse: input.Fahrzeugklasse,
		Kraftstoffart:  input.Kraftstoffart,
		Schadenfrei:    input.Schadenfrei,
		Kilometerstand: input.Kilometerstand,
		Version:        1,
	}
	f.items = append(f.items, item)
	return item, nil
}

func (f *fakeGebrauchtwagenRepository) Update(_ context.Context, id int, expectedVersion int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	for index, item := range f.items {
		if item.ID == id {
			if item.Version != expectedVersion {
				return domain.Gebrauchtwagen{}, domain.ErrVersionConflict
			}
			item.Marke = input.Marke
			item.Modell = input.Modell
			item.Fahrzeugklasse = input.Fahrzeugklasse
			item.Kraftstoffart = input.Kraftstoffart
			item.Schadenfrei = input.Schadenfrei
			item.Kilometerstand = input.Kilometerstand
			item.Version++
			f.items[index] = item
			return item, nil
		}
	}
	return domain.Gebrauchtwagen{}, domain.ErrNotFound
}

func (f *fakeGebrauchtwagenRepository) Delete(_ context.Context, id int) error {
	for index, item := range f.items {
		if item.ID == id {
			f.items = append(f.items[:index], f.items[index+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

func TestLiveness(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/liveness", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestOptionsReturnsCORSHeaders(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/gebrauchtwagen", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if allowOrigin := response.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Fatalf("expected wildcard CORS origin, got %q", allowOrigin)
	}
}

func TestCreateGebrauchtwagen(t *testing.T) {
	repo := &fakeGebrauchtwagenRepository{}
	router := NewRouter(Dependencies{Repository: repo})
	body := `{"marke":"VW","modell":"Golf","fahrzeugklasse":"KOMPAKTKLASSE","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":12000}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/gebrauchtwagen/", strings.NewReader(body))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", response.Code, response.Body.String())
	}
	if location := response.Header().Get("Location"); location != "/api/gebrauchtwagen/1" {
		t.Fatalf("expected Location /api/gebrauchtwagen/1, got %q", location)
	}
	if etag := response.Header().Get("ETag"); etag != `"1"` {
		t.Fatalf("expected ETag \"1\", got %q", etag)
	}
	if !strings.Contains(response.Body.String(), `"id":1`) {
		t.Fatalf("expected created item in response body, got %s", response.Body.String())
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one created item, got %d", len(repo.items))
	}
}

func TestCreateGebrauchtwagenRejectsInvalidInput(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{}})
	body := `{"marke":"","modell":"","fahrzeugklasse":"FALSCH","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":-1}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/gebrauchtwagen/", bytes.NewBufferString(body))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.Code)
	}
}

func TestListRejectsUnknownQueryParameter(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/gebrauchtwagen?farbe=rot", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", response.Code)
	}
}

func TestDetailReturnsETag(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{items: []domain.Gebrauchtwagen{{
		ID: 1, Marke: "VW", Modell: "Golf", Fahrzeugklasse: "KOMPAKTKLASSE", Kraftstoffart: "BENZIN", Schadenfrei: true, Kilometerstand: 12000, Version: 3,
	}}}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/gebrauchtwagen/1", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if etag := response.Header().Get("ETag"); etag != `"3"` {
		t.Fatalf("expected ETag \"3\", got %q", etag)
	}
	if !strings.Contains(response.Body.String(), `"standort"`) {
		t.Fatalf("expected detail response to contain relation data, got %s", response.Body.String())
	}
}

func TestUpdateRequiresIfMatch(t *testing.T) {
	router := NewRouter(Dependencies{Repository: &fakeGebrauchtwagenRepository{items: []domain.Gebrauchtwagen{{ID: 1, Version: 1}}}})
	body := `{"marke":"VW","modell":"Golf","fahrzeugklasse":"KOMPAKTKLASSE","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":12000}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/gebrauchtwagen/1", strings.NewReader(body))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusPreconditionRequired {
		t.Fatalf("expected status 428, got %d", response.Code)
	}
}

func TestProblemDetailsJSONShape(t *testing.T) {
	response := httptest.NewRecorder()

	writeProblem(response, http.StatusNotFound, "nicht gefunden")

	var problem map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatalf("problem response is not json: %v", err)
	}
	if problem["title"] != "Not Found" {
		t.Fatalf("unexpected title: %v", problem["title"])
	}
}

func TestFakeRepositoryVersionConflict(t *testing.T) {
	repo := &fakeGebrauchtwagenRepository{items: []domain.Gebrauchtwagen{{ID: 1, Version: 2}}}

	_, err := repo.Update(context.Background(), 1, 1, domain.GebrauchtwagenWrite{})

	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
}
