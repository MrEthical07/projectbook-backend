package cache

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func BenchmarkBuildReadKey(b *testing.B) {
	mgr, _ := newTestManager(b)

	req := httptest.NewRequest("GET", "/api/v1/projects/p1/tasks/t1?limit=10&cursor=abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", "p1")
	rctx.URLParams.Add("task_id", "t1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	cfg := CacheReadConfig{
		TTL: 30 * time.Second,
		TagSpecs: []CacheTagSpec{
			{Name: "project"},
			{Name: "task"},
		},
		VaryBy: CacheVaryBy{
			ProjectID:   true,
			PathParams:  []string{"project_id", "task_id"},
			QueryParams: []string{"limit", "cursor"},
		},
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := mgr.BuildReadKey(ctx, req, "/api/v1/projects/{project_id}/tasks/{task_id}", cfg); err != nil {
			b.Fatalf("BuildReadKey error: %v", err)
		}
	}
}

func BenchmarkBuildReadKeyWithTemplate(b *testing.B) {
	mgr, _ := newTestManager(b)

	req := httptest.NewRequest("GET", "/api/v1/projects/p1/tasks/t1?limit=10&cursor=abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", "p1")
	rctx.URLParams.Add("task_id", "t1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	cfg := CacheReadConfig{
		TTL: 30 * time.Second,
		TagSpecs: []CacheTagSpec{
			{Name: "project"},
			{Name: "task"},
		},
		VaryBy: CacheVaryBy{
			ProjectID:   true,
			PathParams:  []string{"project_id", "task_id"},
			QueryParams: []string{"limit", "cursor"},
		},
	}
	template := PrepareReadKeyTemplate(cfg)
	routePart := NormalizeRoute("/api/v1/projects/{project_id}/tasks/{task_id}")

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := mgr.BuildReadKeyWithTemplate(ctx, req, routePart, template); err != nil {
			b.Fatalf("BuildReadKeyWithTemplate error: %v", err)
		}
	}
}
