package repositories

import (
	"context"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

const (
	perPageMax    = 25
	defaultRadius = 80000 // 80km
)

type Facility struct {
	Id          int      `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Description *string  `json:"description" db:"description"`
	Latitude    *float32 `json:"latitude" db:"latitude"`
	Longitude   *float32 `json:"longitude" db:"longitude"`
	FacilityId  string   `json:"facilityId" db:"facility_id"`
	PrimaryImg  *string  `json:"primaryImg"`
}

type FacilityMedia struct {
	Id         int     `json:"id" db:"id"`
	Title      *string `json:"title" db:"title"`
	Url        *string `json:"url" db:"url"`
	IsPrimary  bool    `json:"isPrimary" db:"is_primary"`
	FacilityId string  `json:"facilityId" db:"facility_id"`
}

type GetFacilitiesFilter struct {
	Lat  string
	Lng  string
	Sort string
	Page string
}

type GetMetadata struct {
	Page  int `json:"page"`
	Total int `json:"total"`
}

type GetFacilitiesResponse struct {
	Data     []Facility  `json:"data"`
	Metadata GetMetadata `json:"metadata"`
}

func (r *Repository) GetFacilities(ctx context.Context, filter GetFacilitiesFilter) (response *GetFacilitiesResponse, err error) {
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

	countSql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("COUNT(*)").
		From(`"facility"`)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"facility"`).
		Limit(perPageMax)

	switch filter.Sort {
	case "za":
		psql = psql.OrderBy("name DESC")
	case "new":
		psql = psql.OrderBy("id DESC")
	default:
		psql = psql.OrderBy("name")
	}

	if filter.Lat != "" && filter.Lng != "" {
		countSql = countSql.Where("ST_DWithin(geom, ST_MakePoint(?, ?)::geography, ?)", filter.Lng, filter.Lat, defaultRadius)
		psql = psql.Where("ST_DWithin(geom, ST_MakePoint(?, ?)::geography, ?)", filter.Lng, filter.Lat, defaultRadius)
	}

	offset := 0
	if filter.Page != "" {
		offset, err = strconv.Atoi(filter.Page)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query | %w", err)
		}
		psql = psql.Offset(uint64(offset-1) * perPageMax)
	}

	var totalCnt int
	{
		sqlStmt, sqlArgs, err := countSql.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		rows, err := tx.Query(ctx, sqlStmt, sqlArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		if err := pgxscan.ScanOne(&totalCnt, rows); err != nil {
			return nil, fmt.Errorf("failed to scan rows | %w", err)
		}
	}

	var facilities []Facility
	{
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
	}

	if err := r.setFacilityMedias(ctx, facilities); err != nil {
		return nil, fmt.Errorf("failed to set facility medias | %w", err)
	}

	return &GetFacilitiesResponse{
		Data: facilities,
		Metadata: GetMetadata{
			Page:  offset,
			Total: totalCnt,
		},
	}, nil
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

func (r *Repository) setFacilityMedias(ctx context.Context, facilities []Facility) (err error) {
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

	facilityIds := make([]string, len(facilities))
	for idx := range facilities {
		facilityIds = append(facilityIds, facilities[idx].FacilityId)
	}

	cols := []string{
		"title",
		"url",
		"is_primary",
		"facility_id",
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"facility_media"`).
		Where(sq.Eq{"is_primary": true}).
		Where(sq.Eq{"facility_id": facilityIds})

	sqlStmt, sqlArgs, err := psql.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	rows, err := tx.Query(ctx, sqlStmt, sqlArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	var facilityMedias []FacilityMedia
	if err := pgxscan.ScanAll(&facilityMedias, rows); err != nil {
		return fmt.Errorf("failed to scan rows | %w", err)
	}

	facilityMediasMap := make(map[string]string, len(facilities))
	for idx := range facilityMedias {
		facilityMediasMap[facilityMedias[idx].FacilityId] = *facilityMedias[idx].Url
	}

	for idx := range facilities {
		facility := &facilities[idx]
		primaryImg := facilityMediasMap[facility.FacilityId]
		facility.PrimaryImg = &primaryImg
	}

	return nil
}
