package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type SubscriptionToken struct {
	Id             int    `json:"id" db:"id"`
	SubscriptionId int    `json:"subscriptionId" db:"subscription_id"`
	Token          string `json:"token" db:"token"`
}

type GetSubscriptionTokensFilter struct {
	Token string
}

type GetSubscriptionTokensResponse struct {
	Data     []SubscriptionToken `json:"data"`
	Metadata GetMetadata         `json:"metadata"`
}

type CreateSubscriptionTokenPayload struct {
	SubscriptionId int `json:"subscriptionId"`
}

func (r *Repository) GetSubscriptionTokens(ctx context.Context, filter GetSubscriptionTokensFilter) (response *GetSubscriptionTokensResponse, err error) {
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
		"subscription_id",
		"token",
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(cols...).
		From(`"subscription_token"`).
		Limit(perPageMax)

	if filter.Token != "" {
		psql = psql.Where(sq.Eq{"token": filter.Token})
	}

	var subscriptionTokens []SubscriptionToken
	{
		sqlStmt, sqlArgs, err := psql.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		rows, err := tx.Query(ctx, sqlStmt, sqlArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %s args: %v | %w", sqlStmt, sqlArgs, err)
		}
		if err := pgxscan.ScanAll(&subscriptionTokens, rows); err != nil {
			return nil, fmt.Errorf("failed to scan rows | %w", err)
		}
	}

	return &GetSubscriptionTokensResponse{
		Data: subscriptionTokens,
		Metadata: GetMetadata{
			Page:  0,
			Total: 1,
		},
	}, nil
}

func (r *Repository) CreateSubscriptionToken(ctx context.Context, payload CreateSubscriptionTokenPayload) (subscription *SubscriptionToken, err error) {
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

	cols := []string{"subscription_id", "token"}
	vals := []interface{}{payload.SubscriptionId, generateToken()}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sqlStmt, sqlArgs, err := psql.Insert(`"subscription_token"`).
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	var newSubscriptionToken SubscriptionToken
	if err := tx.QueryRow(ctx, sqlStmt, sqlArgs...).Scan(&newSubscriptionToken.Id); err != nil {
		return nil, fmt.Errorf("failed to execute: %s args: %v | %w", sqlStmt, sqlArgs, err)
	}

	return &newSubscriptionToken, nil
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
