package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	corestorage "github.com/MrEthical07/superapi/internal/core/storage"
)

var errRelationalStoreNotConfigured = errors.New("relational store is not configured")

// Store provides infrastructure-level store wiring helpers.
type Store struct {
	relational corestorage.RelationalStore
}

// NewPostgresStore builds a relational store backed by Postgres.
func NewPostgresStore(pool *pgxpool.Pool) (*Store, error) {
	relational, err := corestorage.NewPostgresRelationalStore(pool)
	if err != nil {
		return nil, err
	}
	return &Store{relational: relational}, nil
}

// Relational exposes the relational execution store for repositories.
func (s *Store) Relational() corestorage.RelationalStore {
	if s == nil {
		return nil
	}
	return s.relational
}

// WithTx applies a write transaction boundary at the infrastructure layer.
func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if s == nil || s.relational == nil {
		return errRelationalStoreNotConfigured
	}
	return s.relational.WithTx(ctx, fn)
}
