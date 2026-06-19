package service

import (
	"context"
	"errors"
	"testing"

	"github.com/buan1027/workshop/internal/domain"
)

type fakeRepository struct {
	created *domain.GebrauchtwagenWrite
}

func (f *fakeRepository) List(_ context.Context, search domain.SearchParams) (domain.Page, error) {
	return domain.Page{Page: search.Page, Size: search.Size}, nil
}

func (f *fakeRepository) FindByID(_ context.Context, _ int) (domain.Gebrauchtwagen, error) {
	return domain.Gebrauchtwagen{}, domain.ErrNotFound
}

func (f *fakeRepository) FindDetailByID(_ context.Context, _ int) (domain.GebrauchtwagenDetail, error) {
	return domain.GebrauchtwagenDetail{}, domain.ErrNotFound
}

func (f *fakeRepository) Create(_ context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	f.created = &input
	return domain.Gebrauchtwagen{ID: 1, Version: 1}, nil
}

func (f *fakeRepository) Update(_ context.Context, _ int, _ int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	return domain.Gebrauchtwagen{Marke: input.Marke, Version: 2}, nil
}

func (f *fakeRepository) Delete(_ context.Context, _ int) error {
	return nil
}

func TestCreateValidatesBeforeRepositoryCall(t *testing.T) {
	repo := &fakeRepository{}
	service := NewGebrauchtwagenService(repo)

	_, err := service.Create(context.Background(), domain.GebrauchtwagenWrite{})

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if repo.created != nil {
		t.Fatal("expected repository not to be called for invalid input")
	}
}

func TestCreateDelegatesValidInput(t *testing.T) {
	repo := &fakeRepository{}
	service := NewGebrauchtwagenService(repo)

	created, err := service.Create(context.Background(), domain.GebrauchtwagenWrite{
		FIN:            "WVWZZZ1JZXW000001",
		Marke:          "VW",
		Modell:         "Golf",
		Fahrzeugklasse: "KOMPAKTKLASSE",
		Kraftstoffart:  "BENZIN",
		Schadenfrei:    true,
		Kilometerstand: 12000,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected created id 1, got %d", created.ID)
	}
	if repo.created == nil {
		t.Fatal("expected repository to be called")
	}
}
