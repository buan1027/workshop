package repository

import (
	"context"

	"github.com/buan1027/workshop/internal/domain"
)

type GebrauchtwagenRepository interface {
	List(ctx context.Context, search domain.SearchParams) (domain.Page, error)
	FindByID(ctx context.Context, id int) (domain.Gebrauchtwagen, error)
	Create(ctx context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error)
	Update(ctx context.Context, id int, expectedVersion int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error)
	Delete(ctx context.Context, id int) error
}
