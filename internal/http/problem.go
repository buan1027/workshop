package httpapi

import (
	"encoding/json"
	"net/http"
)

type problemDetails struct {
	Title      string `json:"title"`
	StatusCode int    `json:"statusCode"`
	Detail     any    `json:"detail"`
}

func writeProblem(w http.ResponseWriter, status int, detail any) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problemDetails{
		Title:      http.StatusText(status),
		StatusCode: status,
		Detail:     detail,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
