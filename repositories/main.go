package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CtxKey string

const (
	TxnKey CtxKey = "txnKey"
)

type GetMetadata struct {
	Page  int `json:"page"`
	Total int `json:"total"`
}

type IRepository interface {
	BeginTxn(ctx context.Context) (context.Context, error)
	CommitTxn(ctx context.Context) error
	RollbackTxn(ctx context.Context) error

	GetFacilities(ctx context.Context, filter GetFacilitiesFilter) (*GetFacilitiesResponse, error)
	GetFacility(ctx context.Context, id string) (*Facility, error)
	GetSubscriptions(ctx context.Context, filter GetSubscriptionsFilter) (*GetSubscriptionsResponse, error)
	CreateSubscription(ctx context.Context, payload CreateSubscriptionPayload) (*Subscription, error)
	UpdateSubscription(ctx context.Context, id string, payload UpdateSubscriptionPayload) (*Subscription, error)
	GetSubscriptionTokens(ctx context.Context, filter GetSubscriptionTokensFilter) (*GetSubscriptionTokensResponse, error)
	CreateSubscriptionToken(ctx context.Context, payload CreateSubscriptionTokenPayload) (*SubscriptionToken, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) (*Repository, error) {
	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
}

func (r Repository) BeginTxn(ctx context.Context) (context.Context, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, TxnKey, tx), nil
}

func (r Repository) CommitTxn(ctx context.Context) error {
	tx := ctx.Value(TxnKey)
	if tx == nil {
		return fmt.Errorf("failed to get txn from ctx")
	}

	return tx.(pgx.Tx).Commit(ctx)
}

func (r Repository) RollbackTxn(ctx context.Context) error {
	tx := ctx.Value(TxnKey)
	if tx == nil {
		return fmt.Errorf("failed to get txn from ctx")
	}

	return tx.(pgx.Tx).Rollback(ctx)
}
