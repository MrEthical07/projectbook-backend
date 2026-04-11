package permissions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/rbac"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type relationalExecCall struct {
	query string
	args  []any
}

type fakeRelationalStore struct {
	execCalls      []relationalExecCall
	queryRows      map[string][][]any
	executeErr     error
	execErr        error
	queryErr       error
	queryRowErr    error
	queryRowResult []any
}

func (s *fakeRelationalStore) Kind() storage.Kind {
	return storage.KindRelational
}

func (s *fakeRelationalStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return fmt.Errorf("nil tx callback")
	}
	return fn(ctx)
}

func (s *fakeRelationalStore) Execute(ctx context.Context, op storage.RelationalOperation) error {
	if op == nil {
		return fmt.Errorf("nil operation")
	}
	if s.executeErr != nil {
		return s.executeErr
	}
	return op.ExecuteRelational(ctx, &fakeRelationalExecutor{store: s})
}

type fakeRelationalExecutor struct {
	store *fakeRelationalStore
}

func (e *fakeRelationalExecutor) Exec(_ context.Context, query string, args ...any) error {
	e.store.execCalls = append(e.store.execCalls, relationalExecCall{query: query, args: append([]any(nil), args...)})
	if e.store.execErr != nil {
		return e.store.execErr
	}
	return nil
}

func (e *fakeRelationalExecutor) QueryRow(_ context.Context, _ string, scan func(storage.RowScanner) error, _ ...any) error {
	if e.store.queryRowErr != nil {
		return e.store.queryRowErr
	}
	if scan == nil {
		return nil
	}
	return scan(fakeRowScanner{values: e.store.queryRowResult})
}

func (e *fakeRelationalExecutor) Query(_ context.Context, query string, scan func(storage.RowScanner) error, _ ...any) error {
	if e.store.queryErr != nil {
		return e.store.queryErr
	}
	rows := e.store.queryRows[query]
	for _, values := range rows {
		if err := scan(fakeRowScanner{values: values}); err != nil {
			return err
		}
	}
	return nil
}

type fakeRowScanner struct {
	values []any
}

func (r fakeRowScanner) Scan(dest ...any) error {
	if len(dest) != len(r.values) {
		return fmt.Errorf("scan destination mismatch: got=%d want=%d", len(dest), len(r.values))
	}
	for idx := range dest {
		switch d := dest[idx].(type) {
		case *string:
			value, ok := r.values[idx].(string)
			if !ok {
				return fmt.Errorf("value at index %d is not a string", idx)
			}
			*d = value
		default:
			return fmt.Errorf("unsupported scan destination type at index %d", idx)
		}
	}
	return nil
}

type fakeTagInvalidator struct {
	tags [][]string
	err  error
}

func (f *fakeTagInvalidator) BumpTags(_ context.Context, tags []string) error {
	f.tags = append(f.tags, append([]string(nil), tags...))
	return f.err
}

func TestNewLifecycleRejectsNilStore(t *testing.T) {
	if _, err := NewLifecycle(nil, nil, nil); err == nil {
		t.Fatal("expected error for nil relational store")
	}
}

func TestLifecycleSeedRoleMasks(t *testing.T) {
	store := &fakeRelationalStore{}
	lifecycle, err := NewLifecycle(store, nil, nil)
	if err != nil {
		t.Fatalf("NewLifecycle() error = %v", err)
	}

	if err := lifecycle.SeedRoleMasks(context.Background()); err != nil {
		t.Fatalf("SeedRoleMasks() error = %v", err)
	}

	roles := rbac.CanonicalRoles()
	if len(store.execCalls) != len(roles) {
		t.Fatalf("exec call count = %d want %d", len(store.execCalls), len(roles))
	}

	for idx, role := range roles {
		call := store.execCalls[idx]
		if call.query != querySeedRolePermissionMask {
			t.Fatalf("unexpected seed query at index %d", idx)
		}
		if len(call.args) != 2 {
			t.Fatalf("seed args len = %d want 2", len(call.args))
		}

		roleArg, ok := call.args[0].(string)
		if !ok {
			t.Fatalf("seed role argument at index %d is not string", idx)
		}
		if roleArg != role {
			t.Fatalf("seed role argument = %q want %q", roleArg, role)
		}

		maskArg, ok := call.args[1].(int64)
		if !ok {
			t.Fatalf("seed mask argument at index %d is not int64", idx)
		}

		mask, _ := rbac.DefaultRoleMask(role)
		if maskArg != int64(mask) {
			t.Fatalf("seed mask argument = %d want %d", maskArg, mask)
		}
	}
}

func TestLifecycleResyncNonCustomMemberMasks(t *testing.T) {
	store := &fakeRelationalStore{
		queryRows: map[string][][]any{
			queryResyncNonCustomMemberMasks: {
				{" user-1 ", " project-1 "},
				{"user-1", "project-1"},
				{"", "project-2"},
				{"user-2", ""},
				{"user-2", "project-2"},
			},
		},
	}
	lifecycle, err := NewLifecycle(store, nil, nil)
	if err != nil {
		t.Fatalf("NewLifecycle() error = %v", err)
	}

	affected, err := lifecycle.ResyncNonCustomMemberMasks(context.Background())
	if err != nil {
		t.Fatalf("ResyncNonCustomMemberMasks() error = %v", err)
	}

	if len(affected) != 2 {
		t.Fatalf("affected scope count = %d want 2", len(affected))
	}

	if affected[0].UserID != "user-1" || affected[0].ProjectID != "project-1" {
		t.Fatalf("unexpected first affected scope: %#v", affected[0])
	}
	if affected[1].UserID != "user-2" || affected[1].ProjectID != "project-2" {
		t.Fatalf("unexpected second affected scope: %#v", affected[1])
	}
}

func TestLifecycleRunStartupInvalidatesTags(t *testing.T) {
	store := &fakeRelationalStore{
		queryRows: map[string][][]any{
			queryResyncNonCustomMemberMasks: {
				{"user-1", "project-1"},
				{"user-2", "project-1"},
				{"user-1", "project-1"},
			},
		},
	}
	invalidator := &fakeTagInvalidator{}

	lifecycle, err := NewLifecycle(store, nil, invalidator)
	if err != nil {
		t.Fatalf("NewLifecycle() error = %v", err)
	}

	if err := lifecycle.RunStartup(context.Background()); err != nil {
		t.Fatalf("RunStartup() error = %v", err)
	}

	if len(invalidator.tags) != 1 {
		t.Fatalf("invalidator call count = %d want 1", len(invalidator.tags))
	}

	expected := map[string]struct{}{
		"permissions:user=user-1":                   {},
		"permissions:user=user-2":                   {},
		"permissions:project=project-1":             {},
		"permissions:project_user=project-1:user-1": {},
		"permissions:project_user=project-1:user-2": {},
	}

	actualTags := invalidator.tags[0]
	if len(actualTags) != len(expected) {
		t.Fatalf("tag count = %d want %d", len(actualTags), len(expected))
	}
	for _, tag := range actualTags {
		if _, ok := expected[tag]; !ok {
			t.Fatalf("unexpected invalidation tag: %s", tag)
		}
		delete(expected, tag)
	}
	if len(expected) != 0 {
		t.Fatal("missing expected invalidation tags")
	}
}

func TestLifecycleRunStartupPropagatesInvalidatorError(t *testing.T) {
	store := &fakeRelationalStore{
		queryRows: map[string][][]any{
			queryResyncNonCustomMemberMasks: {
				{"user-1", "project-1"},
			},
		},
	}
	invalidator := &fakeTagInvalidator{err: errors.New("cache unavailable")}

	lifecycle, err := NewLifecycle(store, nil, invalidator)
	if err != nil {
		t.Fatalf("NewLifecycle() error = %v", err)
	}

	err = lifecycle.RunStartup(context.Background())
	if err == nil {
		t.Fatal("expected run startup invalidator error")
	}
	if !strings.Contains(err.Error(), "invalidate permission tags") {
		t.Fatalf("expected wrapped invalidation error, got: %v", err)
	}
}
