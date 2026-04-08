package store

import (
	"context"
	"testing"
)

func TestWithTxReturnsErrorWhenStoreNotConfigured(t *testing.T) {
	var s *Store

	err := s.WithTx(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error when store is not configured")
	}
}

func TestRelationalReturnsNilWhenStoreNotConfigured(t *testing.T) {
	var s *Store

	if got := s.Relational(); got != nil {
		t.Fatalf("Relational() expected nil for unconfigured store")
	}
}
