package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Subscription struct {
	Id         int     `json:"id" db:"id"`
	Email      string  `json:"email" db:"email"`
	TargetDate string  `json:"targetDate" db:"target_date"`
	FacilityId int     `json:"facilityId" db:"facility_id"`
	Status     *string `json:"status" db:"status"`
}

type GetSubscriptionsFilter struct {
	FacilityIds string
	Status      string
	Page        string
}

type GetSubscriptionsResponse struct {
	Data     []Subscription `json:"data"`
	Metadata GetMetadata    `json:"metadata"`
}

type CreateSubscriptionPayload struct {
	Email      string  `json:"email"`
	TargetDate string  `json:"targetDate"`
	FacilityId int     `json:"facilityId"`
	Status     *string `json:"status"`
}

func (r *Repository) GetSubscriptions(ctx context.Context, filter GetSubscriptionsFilter) (response *GetSubscriptionsResponse, err error) {
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
		"email",
		"target_date",
		"facility_id",
		"status",
	}

	countSql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("COUNT(*)").
		From(`"subscription"`)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"subscription"`).
		Limit(perPageMax)

	if filter.Status != "" {
		countSql = countSql.Where(sq.Eq{"status": filter.Status})
		psql = psql.Where(sq.Eq{"status": filter.Status})
	}

	if filter.FacilityIds != "" {
		facilityIds := strings.Split(filter.FacilityIds, ",")
		countSql = countSql.Where(sq.Eq{"facility_id": facilityIds})
		psql = psql.Where(sq.Eq{"facility_id": facilityIds})
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

	var subscriptions []Subscription
	{
		sqlStmt, sqlArgs, err := psql.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		rows, err := tx.Query(ctx, sqlStmt, sqlArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		if err := pgxscan.ScanAll(&subscriptions, rows); err != nil {
			return nil, fmt.Errorf("failed to scan rows | %w", err)
		}
	}

	return &GetSubscriptionsResponse{
		Data: subscriptions,
		Metadata: GetMetadata{
			Page:  offset,
			Total: totalCnt,
		},
	}, nil
}

func (r *Repository) CreateSubscription(ctx context.Context, payload CreateSubscriptionPayload) (subscription *Subscription, err error) {
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

	cols := []string{"email", "target_date", "facility_id", "status"}
	vals := []interface{}{payload.Email, payload.TargetDate, payload.FacilityId, payload.Status}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sqlStmt, sqlArgs, err := psql.Insert(`"subscription"`).
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	var newSubscription Subscription
	if err := tx.QueryRow(ctx, sqlStmt, sqlArgs...).Scan(&newSubscription.Id); err != nil {
		return nil, fmt.Errorf("failed to execute: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	return &newSubscription, nil
}
