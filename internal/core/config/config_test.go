package config

import (
	"strings"
	"testing"
	"time"
)

func TestLintRejectsInvalidMiddlewareBoolEnv(t *testing.T) {
	t.Setenv("HTTP_MIDDLEWARE_REQUEST_ID_ENABLED", "not-a-bool")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid middleware bool env")
	}
}

func TestLintRejectsNegativeMiddlewareBodyLimit(t *testing.T) {
	t.Setenv("HTTP_MIDDLEWARE_MAX_BODY_BYTES", "-1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for negative max body bytes")
	}
}

func TestLintRejectsEnabledPostgresWithoutURL(t *testing.T) {
	t.Setenv("POSTGRES_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Postgres.URL = ""

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for enabled postgres without URL")
	}
}

func TestLoadSupportsDatabaseURLAlias(t *testing.T) {
	t.Setenv("APP_PROFILE", "dev")
	t.Setenv("DATABASE_URL", "postgres://alias-user:alias-pass@localhost:5432/aliasdb?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.Postgres.URL, "postgres://alias-user:alias-pass@localhost:5432/aliasdb?sslmode=disable"; got != want {
		t.Fatalf("Postgres.URL=%q want=%q", got, want)
	}
}

func TestLoadSupportsRedisURLAlias(t *testing.T) {
	t.Setenv("APP_PROFILE", "dev")
	t.Setenv("REDIS_URL", "redis://:secret@localhost:6380/2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.Redis.Addr, "redis://:secret@localhost:6380/2"; got != want {
		t.Fatalf("Redis.Addr=%q want=%q", got, want)
	}
}

func TestLintRejectsEnabledRedisWithoutAddr(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Redis.Addr = ""

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for enabled redis without addr")
	}
}

func TestLintRejectsInvalidRedisPoolSize(t *testing.T) {
	t.Setenv("REDIS_POOL_SIZE", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for redis pool size")
	}
}

func TestLintRejectsInvalidMetricsPath(t *testing.T) {
	t.Setenv("METRICS_PATH", "metrics")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for metrics path")
	}
}

func TestLintRejectsInvalidAccessLogSampleRate(t *testing.T) {
	t.Setenv("HTTP_MIDDLEWARE_ACCESS_LOG_SAMPLE_RATE", "1.2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for access log sample rate")
	}
}

func TestLintRejectsInvalidAccessLogExcludePath(t *testing.T) {
	t.Setenv("HTTP_MIDDLEWARE_ACCESS_LOG_EXCLUDE_PATHS", "healthz,/readyz")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for access log exclude path")
	}
}

func TestLintRejectsInvalidTracingExcludePath(t *testing.T) {
	t.Setenv("HTTP_MIDDLEWARE_TRACING_EXCLUDE_PATHS", "healthz,/readyz")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for tracing exclude path")
	}
}

func TestLintRejectsInvalidMetricsExcludePath(t *testing.T) {
	t.Setenv("METRICS_EXCLUDE_PATHS", "healthz,/readyz")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for metrics exclude path")
	}
}

func TestLintRejectsNegativeCacheTagVersionCacheTTL(t *testing.T) {
	t.Setenv("CACHE_TAG_VERSION_CACHE_TTL", "-1s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for negative cache tag version cache ttl")
	}
}

func TestLintRejectsMiddlewareTimeoutExceedingWriteTimeout(t *testing.T) {
	t.Setenv("HTTP_WRITE_TIMEOUT", "100ms")
	t.Setenv("HTTP_MIDDLEWARE_REQUEST_TIMEOUT", "200ms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for middleware timeout > write timeout")
	}
}

func TestTracingDefaultsToDisabled(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Tracing.Enabled {
		t.Fatalf("expected tracing to be disabled by default")
	}
}

func TestLoadDefaultsEnableCoreDependencies(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.HTTP.Middleware.RequestTimeout, 10*time.Second; got != want {
		t.Fatalf("RequestTimeout=%s want=%s", got, want)
	}
	if !cfg.HTTP.Middleware.CORS.Enabled {
		t.Fatalf("expected CORS enabled by default")
	}
	if !cfg.Auth.Enabled {
		t.Fatalf("expected auth enabled by default")
	}
	if !cfg.RateLimit.Enabled {
		t.Fatalf("expected ratelimit enabled by default")
	}
	if !cfg.Cache.Enabled {
		t.Fatalf("expected cache enabled by default")
	}
	if !cfg.Permissions.Enabled {
		t.Fatalf("expected permissions enabled by default")
	}
	if !cfg.Postgres.Enabled {
		t.Fatalf("expected postgres enabled by default")
	}
	if !cfg.Redis.Enabled {
		t.Fatalf("expected redis enabled by default")
	}
	if !cfg.Mongo.Enabled {
		t.Fatalf("expected mongo enabled by default")
	}
	if got, want := cfg.Postgres.URL, "postgres://superapi:superapi@127.0.0.1:5432/superapi?sslmode=disable"; got != want {
		t.Fatalf("Postgres.URL=%q want=%q", got, want)
	}
	if got, want := cfg.Mongo.URL, "mongodb://127.0.0.1:27017"; got != want {
		t.Fatalf("Mongo.URL=%q want=%q", got, want)
	}
}

func TestDefaultsSetGlobalMaxBodyLimit(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.HTTP.Middleware.MaxBodyBytes, int64(1<<20); got != want {
		t.Fatalf("MaxBodyBytes=%d want=%d", got, want)
	}
}

func TestSecurityHeadersDefaultByEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	cfgProd, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfgProd.HTTP.Middleware.SecurityHeadersEnabled {
		t.Fatalf("expected security headers enabled by default in prod")
	}

	t.Setenv("APP_ENV", "dev")
	cfgDev, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfgDev.HTTP.Middleware.SecurityHeadersEnabled {
		t.Fatalf("expected security headers disabled by default in dev")
	}
}

func TestTracingInsecureDefaultByEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	cfgProd, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfgProd.Tracing.Insecure {
		t.Fatalf("expected tracing insecure=false by default in prod")
	}

	t.Setenv("APP_ENV", "dev")
	cfgDev, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfgDev.Tracing.Insecure {
		t.Fatalf("expected tracing insecure=true by default in dev")
	}
}

func TestLintRejectsMetricsWithoutAuthTokenInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_AUTH_TOKEN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for missing metrics auth token in prod")
	}
}

func TestLoadUsesFailClosedDefaultsInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RateLimit.FailOpen {
		t.Fatalf("expected ratelimit fail-open disabled by default in prod")
	}
	if cfg.Cache.FailOpen {
		t.Fatalf("expected cache fail-open disabled by default in prod")
	}
}

func TestLintRejectsRateLimitFailOpenInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("RATELIMIT_ENABLED", "true")
	t.Setenv("RATELIMIT_FAIL_OPEN", "true")
	t.Setenv("REDIS_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for prod ratelimit fail-open")
	}
}

func TestLintRejectsCacheFailOpenInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("CACHE_FAIL_OPEN", "true")
	t.Setenv("REDIS_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for prod cache fail-open")
	}
}

func TestLintAllowsMetricsTokenInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", strings.Repeat("s", 40))
	t.Setenv("WEB_APP_BASE_URL", "https://app.projectbook.dev")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_AUTH_TOKEN", strings.Repeat("m", 32))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err != nil {
		t.Fatalf("expected lint success, got: %v", err)
	}
}

func TestLintRejectsMissingPermissionContextSecretInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("WEB_APP_BASE_URL", "https://app.projectbook.dev")
	t.Setenv("METRICS_AUTH_TOKEN", strings.Repeat("m", 32))
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for missing permission context secret in prod")
	}
}

func TestLintRejectsFallbackPermissionContextSecretInProd(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("WEB_APP_BASE_URL", "https://app.projectbook.dev")
	t.Setenv("METRICS_AUTH_TOKEN", strings.Repeat("m", 32))
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", "projectbook-dev-permission-context-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for fallback permission context secret in prod")
	}
}

func TestLintRejectsLocalhostWebAppBaseURLInProdProfile(t *testing.T) {
	t.Setenv("APP_PROFILE", "prod")
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", strings.Repeat("s", 40))
	t.Setenv("METRICS_AUTH_TOKEN", strings.Repeat("m", 32))
	t.Setenv("WEB_APP_BASE_URL", "http://localhost:5173")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for localhost web app base url in production profile")
	}
}

func TestLintRejectsWeakMetricsAuthTokenInProdProfile(t *testing.T) {
	t.Setenv("APP_PROFILE", "prod")
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", strings.Repeat("s", 40))
	t.Setenv("WEB_APP_BASE_URL", "https://app.projectbook.dev")
	t.Setenv("METRICS_AUTH_TOKEN", "change-me")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for weak metrics token in production profile")
	}
}

func TestLintRejectsProdProfilePlaceholderDefaults(t *testing.T) {
	t.Setenv("APP_PROFILE", "prod")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for insecure production profile defaults")
	}
}

func TestSensitiveFallbackWarningsIncludeMissingSecret(t *testing.T) {
	t.Setenv("PROJECTBOOK_PERMISSION_CONTEXT_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	warnings := SensitiveFallbackWarnings(cfg)
	found := false
	for _, warning := range warnings {
		if strings.Contains(warning, "PROJECTBOOK_PERMISSION_CONTEXT_SECRET") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected warning for missing permission context secret")
	}
}

func TestLintRejectsInvalidTracingSamplerWhenEnabled(t *testing.T) {
	t.Setenv("TRACING_ENABLED", "true")
	t.Setenv("TRACING_SAMPLER", "invalid")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid tracing sampler")
	}
}

func TestLintRejectsInvalidTracingSampleRatioWhenEnabled(t *testing.T) {
	t.Setenv("TRACING_ENABLED", "true")
	t.Setenv("TRACING_SAMPLE_RATIO", "1.5")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid tracing sample ratio")
	}
}

func TestLintRejectsInvalidAuthMode(t *testing.T) {
	t.Setenv("AUTH_MODE", "invalid")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid auth mode")
	}
}

func TestLintRejectsAuthEnabledWithoutRedis(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("REDIS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for auth enabled without redis")
	}
}

func TestLintRejectsAuthEnabledWithoutPostgres(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("POSTGRES_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for auth enabled without postgres")
	}
}

func TestLintRejectsRateLimitEnabledWithoutRedis(t *testing.T) {
	t.Setenv("RATELIMIT_ENABLED", "true")
	t.Setenv("REDIS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for ratelimit enabled without redis")
	}
}

func TestLintRejectsCacheEnabledWithoutRedis(t *testing.T) {
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("REDIS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for cache enabled without redis")
	}
}

func TestLintRejectsInvalidCacheDefaultMaxBytes(t *testing.T) {
	t.Setenv("CACHE_DEFAULT_MAX_BYTES", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid cache default max bytes")
	}
}

func TestLintRejectsInvalidRateLimitDefaults(t *testing.T) {
	t.Setenv("RATELIMIT_DEFAULT_LIMIT", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Redis.Enabled = true

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for invalid ratelimit default limit")
	}
}

func TestLoadAppliesMinimalProfileDefaults(t *testing.T) {
	t.Setenv("APP_PROFILE", "minimal")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.Enabled {
		t.Fatalf("expected auth disabled in minimal profile")
	}
	if cfg.Cache.Enabled {
		t.Fatalf("expected cache disabled in minimal profile")
	}
	if cfg.RateLimit.Enabled {
		t.Fatalf("expected ratelimit disabled in minimal profile")
	}
	if !cfg.Postgres.Enabled {
		t.Fatalf("expected postgres enabled in minimal profile")
	}
	if !cfg.Redis.Enabled {
		t.Fatalf("expected redis enabled in minimal profile")
	}
	if !cfg.Mongo.Enabled {
		t.Fatalf("expected mongo enabled in minimal profile")
	}
}

func TestLoadAppliesDevProfileDefaults(t *testing.T) {
	t.Setenv("APP_PROFILE", "dev")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Auth.Enabled {
		t.Fatalf("expected auth enabled in dev profile")
	}
	if cfg.Auth.Mode != "jwt_only" {
		t.Fatalf("auth mode=%q want=%q", cfg.Auth.Mode, "jwt_only")
	}
	if !cfg.Cache.Enabled {
		t.Fatalf("expected cache enabled in dev profile")
	}
	if !cfg.RateLimit.Enabled {
		t.Fatalf("expected ratelimit enabled in dev profile")
	}
}

func TestLoadEnvOverridesProfileDefaults(t *testing.T) {
	t.Setenv("APP_PROFILE", "dev")
	t.Setenv("AUTH_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.Enabled {
		t.Fatalf("expected env override AUTH_ENABLED=false to win over profile")
	}
}

func TestLoadRejectsInvalidProfile(t *testing.T) {
	t.Setenv("APP_PROFILE", "unknown")

	if _, err := Load(); err == nil {
		t.Fatalf("expected load error for invalid profile")
	}
}

func TestLintRejectsEnabledMongoWithoutURL(t *testing.T) {
	t.Setenv("MONGO_ENABLED", "true")
	t.Setenv("MONGO_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Mongo.URL = ""

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for enabled mongo without url")
	}
}

func TestLintRejectsDisabledPostgres(t *testing.T) {
	t.Setenv("POSTGRES_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for disabled postgres")
	}
}

func TestLintRejectsDisabledRedis(t *testing.T) {
	t.Setenv("REDIS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for disabled redis")
	}
}

func TestLintRejectsDisabledMongo(t *testing.T) {
	t.Setenv("MONGO_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for disabled mongo")
	}
}

func TestLintRejectsInvalidMongoPoolBounds(t *testing.T) {
	t.Setenv("MONGO_ENABLED", "true")
	t.Setenv("MONGO_URL", "mongodb://localhost:27017")
	t.Setenv("MONGO_MAX_POOL_SIZE", "5")
	t.Setenv("MONGO_MIN_POOL_SIZE", "6")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Lint(); err == nil {
		t.Fatalf("expected lint error for mongo min pool size exceeding max")
	}
}

func TestLoadMongoDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Mongo.Database != "projectbook" {
		t.Fatalf("mongo database=%q want=%q", cfg.Mongo.Database, "projectbook")
	}
	if cfg.Mongo.MaxPoolSize != 50 {
		t.Fatalf("mongo max pool size=%d want=%d", cfg.Mongo.MaxPoolSize, 50)
	}
	if cfg.Mongo.ConnectTimeout <= 0 {
		t.Fatalf("mongo connect timeout must be > 0")
	}
}
