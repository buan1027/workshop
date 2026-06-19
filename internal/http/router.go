package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/buan1027/workshop/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	DB         *pgxpool.Pool
	Repository repository.GebrauchtwagenRepository
	AdminToken string
	Logger     *slog.Logger
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	health := HealthHandler{DB: deps.DB}
	api := GebrauchtwagenHandler{Repository: deps.Repository, AdminToken: deps.AdminToken}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "workshop-server"})
	})
	r.Get("/health/liveness", health.Liveness)
	r.Get("/health/readiness", health.Readiness)

	r.Route("/api/gebrauchtwagen", func(r chi.Router) {
		r.Get("/", api.List)
		r.Post("/", api.Create)
		r.Get("/{id}", api.Detail)
		r.Put("/{id}", api.Update)
		r.Delete("/{id}", api.Delete)
	})

	return r
}
