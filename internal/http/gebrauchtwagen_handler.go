package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/buan1027/workshop/internal/domain"
	"github.com/buan1027/workshop/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	defaultPageSize = 5
	maxPageSize     = 50
)

type GebrauchtwagenHandler struct {
	Repository repository.GebrauchtwagenRepository
	AdminToken string
}

func (h GebrauchtwagenHandler) List(w http.ResponseWriter, r *http.Request) {
	search, problems := parseSearch(r)
	if len(problems) > 0 {
		writeProblem(w, http.StatusUnprocessableEntity, problems)
		return
	}

	page, err := h.Repository.List(r.Context(), search)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "gebrauchtwagen konnten nicht gelesen werden")
		return
	}

	if _, ok := r.URL.Query()["count-only"]; ok {
		writeJSON(w, http.StatusOK, map[string]int{"count": page.Total})
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func (h GebrauchtwagenHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	item, err := h.Repository.FindDetailByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Kein Gebrauchtwagen mit id=%d gefunden", id))
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "gebrauchtwagen konnte nicht gelesen werden")
		return
	}

	etag := createETag(item.Version)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	writeJSON(w, http.StatusOK, item)
}

func (h GebrauchtwagenHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.requireWriteAccess(w, r) {
		return
	}

	input, ok := decodeWriteBody(w, r)
	if !ok {
		return
	}

	created, err := h.Repository.Create(r.Context(), input)
	if repository.IsUniqueViolation(err) {
		writeProblem(w, http.StatusUnprocessableEntity, "fin ist bereits vorhanden")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "gebrauchtwagen konnte nicht erstellt werden")
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%d", strings.TrimRight(r.URL.Path, "/"), created.ID))
	w.WriteHeader(http.StatusCreated)
}

func (h GebrauchtwagenHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !h.requireWriteAccess(w, r) {
		return
	}

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	version, ok := parseIfMatch(w, r)
	if !ok {
		return
	}

	input, ok := decodeWriteBody(w, r)
	if !ok {
		return
	}

	updated, err := h.Repository.Update(r.Context(), id, version, input)
	if errors.Is(err, domain.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Kein Gebrauchtwagen mit id=%d gefunden", id))
		return
	}
	if errors.Is(err, domain.ErrVersionConflict) {
		writeProblem(w, http.StatusPreconditionFailed, fmt.Sprintf("Version %d ist nicht mehr aktuell", version))
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "gebrauchtwagen konnte nicht aktualisiert werden")
		return
	}

	w.Header().Set("ETag", createETag(updated.Version))
	w.WriteHeader(http.StatusNoContent)
}

func (h GebrauchtwagenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.requireWriteAccess(w, r) {
		return
	}

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	err := h.Repository.Delete(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Kein Gebrauchtwagen mit id=%d gefunden", id))
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "gebrauchtwagen konnte nicht geloescht werden")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h GebrauchtwagenHandler) requireWriteAccess(w http.ResponseWriter, r *http.Request) bool {
	if h.AdminToken == "" {
		return true
	}

	if r.Header.Get("Authorization") != "Bearer "+h.AdminToken {
		writeProblem(w, http.StatusUnauthorized, "gueltiger Bearer Token erforderlich")
		return false
	}

	return true
}

func decodeWriteBody(w http.ResponseWriter, r *http.Request) (domain.GebrauchtwagenWrite, bool) {
	defer r.Body.Close()

	var input domain.GebrauchtwagenWrite
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeProblem(w, http.StatusBadRequest, "ungueltiger JSON-Request")
		return domain.GebrauchtwagenWrite{}, false
	}

	if problems := domain.ValidateWrite(&input); len(problems) > 0 {
		writeProblem(w, http.StatusUnprocessableEntity, problems)
		return domain.GebrauchtwagenWrite{}, false
	}

	return input, true
}

func parseSearch(r *http.Request) (domain.SearchParams, []string) {
	query := r.URL.Query()
	search := domain.SearchParams{
		Marke:          strings.TrimSpace(query.Get("marke")),
		Modell:         strings.TrimSpace(query.Get("modell")),
		Fahrzeugklasse: strings.TrimSpace(query.Get("fahrzeugklasse")),
		Kraftstoffart:  strings.TrimSpace(query.Get("kraftstoffart")),
		Page:           defaultPage(query.Get("page")),
		Size:           defaultSize(query.Get("size")),
	}

	var problems []string
	if !domain.IsValidFahrzeugklasse(search.Fahrzeugklasse) {
		problems = append(problems, "fahrzeugklasse ist ungueltig")
	}
	if !domain.IsValidKraftstoffart(search.Kraftstoffart) {
		problems = append(problems, "kraftstoffart ist ungueltig")
	}
	if search.Page < 1 {
		problems = append(problems, "page muss groesser oder gleich 1 sein")
	}
	if search.Size < 1 || search.Size > maxPageSize {
		problems = append(problems, fmt.Sprintf("size muss zwischen 1 und %d liegen", maxPageSize))
	}
	if raw := query.Get("schadenfrei"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			problems = append(problems, "schadenfrei muss true oder false sein")
		} else {
			search.Schadenfrei = &value
		}
	}

	return search, problems
}

func defaultPage(raw string) int {
	if raw == "" {
		return 1
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}

func defaultSize(raw string) int {
	if raw == "" {
		return defaultPageSize
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "id muss eine positive ganze Zahl sein")
		return 0, false
	}
	return id, true
}

func parseIfMatch(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := strings.TrimSpace(r.Header.Get("If-Match"))
	if raw == "" {
		writeProblem(w, http.StatusPreconditionRequired, "Header \"If-Match\" fehlt oder ist ungueltig")
		return 0, false
	}

	raw = strings.Trim(raw, `"`)
	version, err := strconv.Atoi(raw)
	if err != nil || version <= 0 {
		writeProblem(w, http.StatusPreconditionRequired, "Header \"If-Match\" fehlt oder ist ungueltig")
		return 0, false
	}

	return version, true
}

func createETag(version int) string {
	return fmt.Sprintf(`"%d"`, version)
}
