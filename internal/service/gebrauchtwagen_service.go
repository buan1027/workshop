package service

import (
	"context"

	"github.com/buan1027/workshop/internal/domain"
	"github.com/buan1027/workshop/internal/repository"
)

type GebrauchtwagenService interface {
	List(ctx context.Context, search domain.SearchParams) (domain.Page, error)
	FindDetailByID(ctx context.Context, id int) (domain.GebrauchtwagenDetail, error)
	Create(ctx context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error)
	Update(ctx context.Context, id int, expectedVersion int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error)
	Delete(ctx context.Context, id int) error
}

type ValidationError struct {
	Problems []string
}

func (e ValidationError) Error() string {
	return "validation failed"
}

type gebrauchwagenService struct {
	repository repository.GebrauchtwagenRepository
}

func NewGebrauchtwagenService(repository repository.GebrauchtwagenRepository) GebrauchtwagenService {
	return gebrauchwagenService{repository: repository}
}

func (s gebrauchwagenService) List(ctx context.Context, search domain.SearchParams) (domain.Page, error) {
	return s.repository.List(ctx, search)
}

func (s gebrauchwagenService) FindDetailByID(ctx context.Context, id int) (domain.GebrauchtwagenDetail, error) {
	return s.repository.FindDetailByID(ctx, id)
}

func (s gebrauchwagenService) Create(ctx context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	if problems := domain.ValidateWrite(&input); len(problems) > 0 {
		return domain.Gebrauchtwagen{}, ValidationError{Problems: problems}
	}

	return s.repository.Create(ctx, input)
}

func (s gebrauchwagenService) Update(ctx context.Context, id int, expectedVersion int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	if problems := domain.ValidateWrite(&input); len(problems) > 0 {
		return domain.Gebrauchtwagen{}, ValidationError{Problems: problems}
	}

	return s.repository.Update(ctx, id, expectedVersion, input)
}

func (s gebrauchwagenService) Delete(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}
