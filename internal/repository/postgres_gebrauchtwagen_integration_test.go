package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/buan1027/workshop/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresGebrauchtwagenRepositoryCRUD(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	repo := NewPostgresGebrauchtwagenRepository(pool)
	input := domain.GebrauchtwagenWrite{
		FIN:            "TSTZZZ1JZXW000001",
		Marke:          "Testmarke",
		Modell:         "Testmodell",
		Fahrzeugklasse: "KOMPAKTKLASSE",
		Kraftstoffart:  "BENZIN",
		Schadenfrei:    true,
		Kilometerstand: 12345,
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("create gebrauchtwagen: %v", err)
	}
	t.Cleanup(func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Delete(deleteCtx, created.ID)
	})

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find created gebrauchtwagen: %v", err)
	}
	if found.FIN != input.FIN || found.Marke != input.Marke || found.Version != 1 {
		t.Fatalf("unexpected created gebrauchtwagen: %+v", found)
	}

	input.Kilometerstand = 13000
	updated, err := repo.Update(ctx, created.ID, found.Version, input)
	if err != nil {
		t.Fatalf("update gebrauchtwagen: %v", err)
	}
	if updated.Version != found.Version+1 {
		t.Fatalf("expected version increment, got %d after %d", updated.Version, found.Version)
	}

	page, err := repo.List(ctx, domain.SearchParams{Marke: "Testmarke", Page: 1, Size: 5})
	if err != nil {
		t.Fatalf("list gebrauchtwagen: %v", err)
	}
	if page.Total < 1 {
		t.Fatalf("expected at least one matching gebrauchtwagen, got %d", page.Total)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete gebrauchtwagen: %v", err)
	}
	if _, err := repo.FindByID(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestPostgresGebrauchtwagenRepositoryCreatesRelationData(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	repo := NewPostgresGebrauchtwagenRepository(pool)
	input := domain.GebrauchtwagenWrite{
		FIN:            "RELZZZ1JZXW000001",
		Marke:          "Relation",
		Modell:         "Test",
		Fahrzeugklasse: "KOMPAKTKLASSE",
		Kraftstoffart:  "BENZIN",
		Schadenfrei:    false,
		Kilometerstand: 20000,
		Standort:       &domain.StandortWrite{PLZ: "76131", Ort: "Karlsruhe"},
		Hauptuntersuchung: &domain.HauptuntersuchungWrite{
			Pruefdatum:        "2025-06-01",
			GueltigBis:        "2027-06-01",
			Prueforganisation: "TUEV",
			Status:            "BESTANDEN",
		},
		Schaeden: []domain.SchadenWrite{{
			Bezeichnung:        "Kratzer",
			Beschreibung:       "Kleiner Lackkratzer",
			Feststellungsdatum: "2024-11-10",
		}},
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("create gebrauchtwagen with relations: %v", err)
	}
	t.Cleanup(func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Delete(deleteCtx, created.ID)
	})

	detail, err := repo.FindDetailByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find created detail: %v", err)
	}
	if detail.Standort == nil || detail.Standort.Ort != "Karlsruhe" {
		t.Fatalf("expected Standort Karlsruhe, got %+v", detail.Standort)
	}
	if detail.Hauptuntersuchung == nil || detail.Hauptuntersuchung.Status != "BESTANDEN" {
		t.Fatalf("expected HU BESTANDEN, got %+v", detail.Hauptuntersuchung)
	}
	if len(detail.Schaeden) != 1 {
		t.Fatalf("expected one schaden, got %d", len(detail.Schaeden))
	}
}

func TestPostgresGebrauchtwagenRepositoryFindsDetailRelations(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	repo := NewPostgresGebrauchtwagenRepository(pool)
	detail, err := repo.FindDetailByID(ctx, 1)
	if err != nil {
		t.Fatalf("find detail: %v", err)
	}

	if detail.Standort == nil || detail.Standort.Ort != "Karlsruhe" {
		t.Fatalf("expected Standort Karlsruhe, got %+v", detail.Standort)
	}
	if detail.Hauptuntersuchung == nil || detail.Hauptuntersuchung.Status != "BESTANDEN" {
		t.Fatalf("expected bestandene Hauptuntersuchung, got %+v", detail.Hauptuntersuchung)
	}
	if detail.Schaeden == nil {
		t.Fatal("expected schaeden slice to be initialized")
	}
}
