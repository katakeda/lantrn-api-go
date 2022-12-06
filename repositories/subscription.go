package repositories

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

type Subscription struct {
	Id         int    `json:"id" db:"id"`
	Email      string `json:"email" db:"email"`
	TargetDate string `json:"targetDate" db:"target_date"`
	FacilityId int    `json:"facilityId" db:"facility_id"`
}

type CreateSubscriptionPayload struct {
	Email      string `json:"email"`
	TargetDate string `json:"targetDate"`
	FacilityId int    `json:"facilityId"`
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

	cols := []string{"email", "target_date", "facility_id"}
	vals := []interface{}{payload.Email, payload.TargetDate, payload.FacilityId}

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
