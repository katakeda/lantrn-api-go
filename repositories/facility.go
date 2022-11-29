package repositories

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Facility struct {
	Id          int      `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Description *string  `json:"description" db:"description"`
	Latitude    *float32 `json:"latitude" db:"latitude"`
	Longitude   *float32 `json:"longitude" db:"longitude"`
	FacilityId  string   `json:"facilityId" db:"facility_id"`
}

type GetFacilitiesFilter struct {
	Lat string
	Lng string
}

func (r *Repository) GetFacilities(ctx context.Context, filter GetFacilitiesFilter) (facilities []Facility, err error) {
	tx, ok := ctx.Value(TxnKey).(pgx.Tx)
	if !ok || tx == nil {
		tx, _ = r.db.Begin(ctx)
		defer func() error {
			if err != nil {
				return tx.Rollback(ctx)
			}
			return tx.Commit(ctx)
		}()
	}

	cols := []string{
		"id",
		"name",
		"description",
		"latitude",
		"longitude",
		"facility_id",
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"facility"`).
		OrderBy("name").
		Limit(50)

	if filter.Lat != "" && filter.Lng != "" {
		psql = psql.Where("ST_DWithin(geom, ST_MakePoint(?, ?)::geography, 80000)", filter.Lng, filter.Lat)
	}

	sqlStmt, sqlArgs, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	rows, err := tx.Query(ctx, sqlStmt, sqlArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	if err := pgxscan.ScanAll(&facilities, rows); err != nil {
		return nil, fmt.Errorf("failed to scan rows | %w", err)
	}

	return facilities, nil
}

func (r *Repository) GetFacility(ctx context.Context, id string) (facility *Facility, err error) {
	tx, ok := ctx.Value(TxnKey).(pgx.Tx)
	if !ok || tx == nil {
		tx, _ = r.db.Begin(ctx)
		defer func() error {
			if err != nil {
				return tx.Rollback(ctx)
			}
			return tx.Commit(ctx)
		}()
	}

	cols := []string{
		"id",
		"name",
		"description",
		"latitude",
		"longitude",
		"facility_id",
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"facility"`).
		Where(sq.Eq{"id": id})

	sqlStmt, sqlArgs, err := psql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	facility = &Facility{}
	if err := tx.QueryRow(ctx, sqlStmt, sqlArgs...).Scan(
		&facility.Id,
		&facility.Name,
		&facility.Description,
		&facility.Latitude,
		&facility.Longitude,
		&facility.FacilityId,
	); err != nil {
		return nil, fmt.Errorf("failed to execute: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	return facility, nil
}
