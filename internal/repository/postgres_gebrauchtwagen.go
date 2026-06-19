package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buan1027/workshop/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGebrauchtwagenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGebrauchtwagenRepository(db *pgxpool.Pool) *PostgresGebrauchtwagenRepository {
	return &PostgresGebrauchtwagenRepository{db: db}
}

func (r *PostgresGebrauchtwagenRepository) List(ctx context.Context, search domain.SearchParams) (domain.Page, error) {
	where, args := buildWhere(search)
	limit := len(args) + 1
	offset := len(args) + 2

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version
		FROM gebrauchtwagen.gebrauchtwagen
		%s
		ORDER BY id
		LIMIT $%d OFFSET $%d`, where, limit, offset), append(args, search.Size, (search.Page-1)*search.Size)...)
	if err != nil {
		return domain.Page{}, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Gebrauchtwagen])
	if err != nil {
		return domain.Page{}, err
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM gebrauchtwagen.gebrauchtwagen "+where, args...).Scan(&total); err != nil {
		return domain.Page{}, err
	}

	return domain.Page{Data: items, Total: total, Page: search.Page, Size: search.Size}, nil
}

func (r *PostgresGebrauchtwagenRepository) FindByID(ctx context.Context, id int) (domain.Gebrauchtwagen, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version
		FROM gebrauchtwagen.gebrauchtwagen
		WHERE id = $1`, id)

	item, err := scanGebrauchtwagen(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Gebrauchtwagen{}, domain.ErrNotFound
	}
	return item, err
}

func (r *PostgresGebrauchtwagenRepository) Create(ctx context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	now := time.Now()
	row := r.db.QueryRow(ctx, `
		INSERT INTO gebrauchtwagen.gebrauchtwagen (
			fin, marke, modell, baujahr, erstzulassung, kilometerstand,
			kraftstoffart, fahrzeugklasse, ausstattung, schadenfrei, version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '{}'::jsonb, $9, 1)
		RETURNING id, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version`,
		createFin(), input.Marke, input.Modell, now.Year(), now, input.Kilometerstand,
		input.Kraftstoffart, input.Fahrzeugklasse, input.Schadenfrei)

	return scanGebrauchtwagen(row)
}

func (r *PostgresGebrauchtwagenRepository) Update(ctx context.Context, id int, expectedVersion int, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE gebrauchtwagen.gebrauchtwagen
		SET marke = $3,
		    modell = $4,
		    kilometerstand = $5,
		    kraftstoffart = $6,
		    fahrzeugklasse = $7,
		    schadenfrei = $8,
		    version = version + 1,
		    aktualisiert = CURRENT_TIMESTAMP
		WHERE id = $1 AND version = $2
		RETURNING id, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version`,
		id, expectedVersion, input.Marke, input.Modell, input.Kilometerstand,
		input.Kraftstoffart, input.Fahrzeugklasse, input.Schadenfrei)

	item, err := scanGebrauchtwagen(row)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Gebrauchtwagen{}, err
	}

	_, findErr := r.FindByID(ctx, id)
	if errors.Is(findErr, domain.ErrNotFound) {
		return domain.Gebrauchtwagen{}, domain.ErrNotFound
	}
	if findErr != nil {
		return domain.Gebrauchtwagen{}, findErr
	}
	return domain.Gebrauchtwagen{}, domain.ErrVersionConflict
}

func (r *PostgresGebrauchtwagenRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM gebrauchtwagen.gebrauchtwagen WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanGebrauchtwagen(row pgx.Row) (domain.Gebrauchtwagen, error) {
	var item domain.Gebrauchtwagen
	err := row.Scan(&item.ID, &item.Marke, &item.Modell, &item.Fahrzeugklasse, &item.Kraftstoffart, &item.Schadenfrei, &item.Kilometerstand, &item.Version)
	return item, err
}

func buildWhere(search domain.SearchParams) (string, []any) {
	var clauses []string
	var args []any
	add := func(sql string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(sql, len(args)))
	}

	if search.Marke != "" {
		add("lower(marke) LIKE '%%' || lower($%d) || '%%'", search.Marke)
	}
	if search.Modell != "" {
		add("lower(modell) LIKE '%%' || lower($%d) || '%%'", search.Modell)
	}
	if search.Fahrzeugklasse != "" {
		add("fahrzeugklasse = $%d::gebrauchtwagen.fahrzeugklasse", search.Fahrzeugklasse)
	}
	if search.Kraftstoffart != "" {
		add("kraftstoffart = $%d::gebrauchtwagen.kraftstoffart", search.Kraftstoffart)
	}
	if search.Schadenfrei != nil {
		add("schadenfrei = $%d", *search.Schadenfrei)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func createFin() string {
	fin := "GW" + strings.ToUpper(strconv.FormatInt(time.Now().UnixNano(), 36))
	if len(fin) > 17 {
		return fin[:17]
	}
	return fin
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
