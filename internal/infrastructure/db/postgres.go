package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MrEthical07/superapi/internal/core/config"
	coredb "github.com/MrEthical07/superapi/internal/core/db"
)

// NewPostgresPool creates a pgx pool and performs fail-fast startup ping checks.
func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	resolved := cfg
	resolved.URL = resolvePostgresURL(cfg.URL)
	if strings.TrimSpace(resolved.URL) == "" {
		return nil, fmt.Errorf("postgres url cannot be empty")
	}

	return coredb.NewPool(ctx, resolved)
}

// CheckPostgresHealth performs a bounded Postgres ping for readiness probes.
func CheckPostgresHealth(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	return coredb.CheckHealth(ctx, pool, timeout)
}

func resolvePostgresURL(configURL string) string {
	if trimmed := strings.TrimSpace(configURL); trimmed != "" {
		return trimmed
	}
	if envValue := strings.TrimSpace(os.Getenv("DATABASE_URL")); envValue != "" {
		return envValue
	}
	return strings.TrimSpace(os.Getenv("POSTGRES_URL"))
}
