package repository

import (
	"context"
	"errors"
	"fmt"
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
		SELECT id, fin, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version
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
		SELECT id, fin, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version
		FROM gebrauchtwagen.gebrauchtwagen
		WHERE id = $1`, id)

	item, err := scanGebrauchtwagen(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Gebrauchtwagen{}, domain.ErrNotFound
	}
	return item, err
}

func (r *PostgresGebrauchtwagenRepository) FindDetailByID(ctx context.Context, id int) (domain.GebrauchtwagenDetail, error) {
	item, err := r.FindByID(ctx, id)
	if err != nil {
		return domain.GebrauchtwagenDetail{}, err
	}

	detail := domain.GebrauchtwagenDetail{
		Gebrauchtwagen: item,
		Schaeden:       []domain.Schaden{},
	}

	standort, err := r.findStandort(ctx, id)
	if err != nil {
		return domain.GebrauchtwagenDetail{}, err
	}
	detail.Standort = standort

	hu, err := r.findHauptuntersuchung(ctx, id)
	if err != nil {
		return domain.GebrauchtwagenDetail{}, err
	}
	detail.Hauptuntersuchung = hu

	schaeden, err := r.findSchaeden(ctx, id)
	if err != nil {
		return domain.GebrauchtwagenDetail{}, err
	}
	detail.Schaeden = schaeden

	return detail, nil
}

func (r *PostgresGebrauchtwagenRepository) Create(ctx context.Context, input domain.GebrauchtwagenWrite) (domain.Gebrauchtwagen, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Gebrauchtwagen{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now()
	row := tx.QueryRow(ctx, `
		INSERT INTO gebrauchtwagen.gebrauchtwagen (
			fin, marke, modell, baujahr, erstzulassung, kilometerstand,
			kraftstoffart, fahrzeugklasse, ausstattung, schadenfrei, version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '{}'::jsonb, $9, 1)
		RETURNING id, fin, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version`,
		input.FIN, input.Marke, input.Modell, now.Year(), now, input.Kilometerstand,
		input.Kraftstoffart, input.Fahrzeugklasse, input.Schadenfrei)

	created, err := scanGebrauchtwagen(row)
	if err != nil {
		return domain.Gebrauchtwagen{}, err
	}

	if err := insertRelationData(ctx, tx, created.ID, input); err != nil {
		return domain.Gebrauchtwagen{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Gebrauchtwagen{}, err
	}

	return created, nil
}

func (r *PostgresGebrauchtwagenRepository) findStandort(ctx context.Context, gebrauchtwagenID int) (*domain.Standort, error) {
	var standort domain.Standort
	err := r.db.QueryRow(ctx, `
		SELECT plz, ort
		FROM gebrauchtwagen.standort
		WHERE gebrauchtwagen_id = $1`, gebrauchtwagenID).Scan(&standort.PLZ, &standort.Ort)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &standort, nil
}

func (r *PostgresGebrauchtwagenRepository) findHauptuntersuchung(ctx context.Context, gebrauchtwagenID int) (*domain.Hauptuntersuchung, error) {
	var hu domain.Hauptuntersuchung
	err := r.db.QueryRow(ctx, `
		SELECT pruefdatum::text, gueltig_bis::text, prueforganisation, status::text
		FROM gebrauchtwagen.hauptuntersuchung
		WHERE gebrauchtwagen_id = $1`, gebrauchtwagenID).Scan(&hu.Pruefdatum, &hu.GueltigBis, &hu.Prueforganisation, &hu.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hu, nil
}

func (r *PostgresGebrauchtwagenRepository) findSchaeden(ctx context.Context, gebrauchtwagenID int) ([]domain.Schaden, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bezeichnung, beschreibung, feststellungsdatum::text
		FROM gebrauchtwagen.schaden
		WHERE gebrauchtwagen_id = $1
		ORDER BY id`, gebrauchtwagenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Schaden])
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
		    fin = $9,
		    version = version + 1,
		    aktualisiert = CURRENT_TIMESTAMP
		WHERE id = $1 AND version = $2
		RETURNING id, fin, marke, modell, fahrzeugklasse::text, kraftstoffart::text, schadenfrei, kilometerstand, version`,
		id, expectedVersion, input.Marke, input.Modell, input.Kilometerstand,
		input.Kraftstoffart, input.Fahrzeugklasse, input.Schadenfrei, input.FIN)

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
	err := row.Scan(&item.ID, &item.FIN, &item.Marke, &item.Modell, &item.Fahrzeugklasse, &item.Kraftstoffart, &item.Schadenfrei, &item.Kilometerstand, &item.Version)
	return item, err
}

func insertRelationData(ctx context.Context, tx pgx.Tx, gebrauchtwagenID int, input domain.GebrauchtwagenWrite) error {
	if input.Standort != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO gebrauchtwagen.standort (plz, ort, gebrauchtwagen_id)
			VALUES ($1, $2, $3)`,
			input.Standort.PLZ, input.Standort.Ort, gebrauchtwagenID); err != nil {
			return err
		}
	}

	if input.Hauptuntersuchung != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO gebrauchtwagen.hauptuntersuchung (
				pruefdatum, gueltig_bis, prueforganisation, status, gebrauchtwagen_id
			)
			VALUES ($1, $2, $3, $4, $5)`,
			input.Hauptuntersuchung.Pruefdatum,
			input.Hauptuntersuchung.GueltigBis,
			input.Hauptuntersuchung.Prueforganisation,
			input.Hauptuntersuchung.Status,
			gebrauchtwagenID); err != nil {
			return err
		}
	}

	for _, schaden := range input.Schaeden {
		if _, err := tx.Exec(ctx, `
			INSERT INTO gebrauchtwagen.schaden (
				bezeichnung, beschreibung, feststellungsdatum, gebrauchtwagen_id
			)
			VALUES ($1, $2, $3, $4)`,
			schaden.Bezeichnung,
			schaden.Beschreibung,
			schaden.Feststellungsdatum,
			gebrauchtwagenID); err != nil {
			return err
		}
	}

	return nil
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

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
