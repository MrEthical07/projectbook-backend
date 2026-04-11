package permissions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
)

type resolverFakeStore struct {
	rowValues   []any
	query       string
	queryArgs   []any
	queryRowErr error
	executeErr  error
}

func (s *resolverFakeStore) Kind() storage.Kind {
	return storage.KindRelational
}

func (s *resolverFakeStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return fmt.Errorf("nil tx callback")
	}
	return fn(ctx)
}

func (s *resolverFakeStore) Execute(ctx context.Context, op storage.RelationalOperation) error {
	if op == nil {
		return fmt.Errorf("nil operation")
	}
	if s.executeErr != nil {
		return s.executeErr
	}
	return op.ExecuteRelational(ctx, &resolverFakeExecutor{store: s})
}

type resolverFakeExecutor struct {
	store *resolverFakeStore
}

func (e *resolverFakeExecutor) Exec(_ context.Context, _ string, _ ...any) error {
	return nil
}

func (e *resolverFakeExecutor) Query(_ context.Context, _ string, _ func(storage.RowScanner) error, _ ...any) error {
	return nil
}

func (e *resolverFakeExecutor) QueryRow(_ context.Context, query string, scan func(storage.RowScanner) error, args ...any) error {
	e.store.query = query
	e.store.queryArgs = append([]any(nil), args...)
	if e.store.queryRowErr != nil {
		return e.store.queryRowErr
	}
	if scan == nil {
		return nil
	}
	return scan(resolverFakeRow{values: e.store.rowValues})
}

type resolverFakeRow struct {
	values []any
}

func (r resolverFakeRow) Scan(dest ...any) error {
	if len(dest) != len(r.values) {
		return fmt.Errorf("scan destination mismatch: got=%d want=%d", len(dest), len(r.values))
	}

	for idx := range dest {
		dv := reflect.ValueOf(dest[idx])
		if dv.Kind() != reflect.Ptr || dv.IsNil() {
			return fmt.Errorf("destination at index %d is not a non-nil pointer", idx)
		}
		target := dv.Elem()

		sv := reflect.ValueOf(r.values[idx])
		if !sv.IsValid() {
			target.SetZero()
			continue
		}
		if sv.Type().AssignableTo(target.Type()) {
			target.Set(sv)
			continue
		}
		if sv.Type().ConvertibleTo(target.Type()) {
			target.Set(sv.Convert(target.Type()))
			continue
		}

		return fmt.Errorf("cannot assign %s to %s at index %d", sv.Type(), target.Type(), idx)
	}

	return nil
}

func TestRelationalResolverResolveBySlugReturnsCanonicalProjectID(t *testing.T) {
	store := &resolverFakeStore{rowValues: []any{
		"f0123d85-4c8f-4f93-b6bd-b7a6546e2284",
		"Admin",
		false,
		int64(4),
		int64(12),
		true,
		int64(100),
		int64(120),
	}}

	resolver, err := NewRelationalResolver(store, 0)
	if err != nil {
		t.Fatalf("NewRelationalResolver() error = %v", err)
	}

	resolved, err := resolver.Resolve(context.Background(), "user-1", "atlas-2026")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolved.ProjectID != "f0123d85-4c8f-4f93-b6bd-b7a6546e2284" {
		t.Fatalf("resolved project id = %q", resolved.ProjectID)
	}
	if resolved.Mask != 12 {
		t.Fatalf("resolved mask = %d want 12", resolved.Mask)
	}
	if resolved.Role != "Admin" {
		t.Fatalf("resolved role = %q want Admin", resolved.Role)
	}
	if resolved.UpdatedAtUnix != 120 {
		t.Fatalf("resolved updatedAt = %d want 120", resolved.UpdatedAtUnix)
	}
	if len(store.queryArgs) != 2 {
		t.Fatalf("query arg count = %d want 2", len(store.queryArgs))
	}
	if store.queryArgs[0] != "atlas-2026" {
		t.Fatalf("project arg = %v want atlas-2026", store.queryArgs[0])
	}
}

func TestRelationalResolverResolveCustomMaskUsesMemberMask(t *testing.T) {
	store := &resolverFakeStore{rowValues: []any{
		"f0123d85-4c8f-4f93-b6bd-b7a6546e2284",
		"Editor",
		true,
		int64(99),
		int64(3),
		true,
		int64(220),
		int64(210),
	}}

	resolver, err := NewRelationalResolver(store, 0)
	if err != nil {
		t.Fatalf("NewRelationalResolver() error = %v", err)
	}

	resolved, err := resolver.Resolve(context.Background(), "user-1", "f0123d85-4c8f-4f93-b6bd-b7a6546e2284")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolved.Mask != 99 {
		t.Fatalf("resolved mask = %d want 99", resolved.Mask)
	}
	if resolved.IsCustom != true {
		t.Fatalf("resolved isCustom = %v want true", resolved.IsCustom)
	}
	if resolved.UpdatedAtUnix != 220 {
		t.Fatalf("resolved updatedAt = %d want 220", resolved.UpdatedAtUnix)
	}
}

func TestRelationalResolverResolveMembershipNotFound(t *testing.T) {
	store := &resolverFakeStore{queryRowErr: pgx.ErrNoRows}

	resolver, err := NewRelationalResolver(store, 0)
	if err != nil {
		t.Fatalf("NewRelationalResolver() error = %v", err)
	}

	_, err = resolver.Resolve(context.Background(), "user-1", "atlas-2026")
	if err == nil {
		t.Fatal("expected membership not found error")
	}
	if !errors.Is(err, ErrMembershipNotFound) {
		t.Fatalf("expected ErrMembershipNotFound, got %v", err)
	}
}
