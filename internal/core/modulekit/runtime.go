package modulekit

import (
	goauth "github.com/MrEthical07/goAuth"
	"github.com/redis/go-redis/v9"

	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/auth"
	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/config"
	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

// Runtime gives modules a single injected surface for optional infrastructure
// dependencies used during route registration and handler/service wiring.
type Runtime struct {
	deps *app.Dependencies
}

// New creates a module runtime wrapper around app dependencies.
func New(deps *app.Dependencies) Runtime {
	return Runtime{deps: deps}
}

// Dependencies returns raw dependency bag for advanced module wiring.
func (r Runtime) Dependencies() *app.Dependencies {
	return r.deps
}

// Store returns the primary configured storage backend.
func (r Runtime) Store() storage.Store {
	if r.deps == nil {
		return nil
	}
	return r.deps.Store
}

// RelationalStore returns the relational storage backend.
func (r Runtime) RelationalStore() storage.RelationalStore {
	if r.deps == nil {
		return nil
	}
	return r.deps.RelationalStore
}

// DocumentStore returns the document storage backend.
func (r Runtime) DocumentStore() storage.DocumentStore {
	if r.deps == nil {
		return nil
	}
	return r.deps.DocumentStore
}

// Redis returns configured redis client when Redis is enabled.
func (r Runtime) Redis() *redis.Client {
	if r.deps == nil {
		return nil
	}
	return r.deps.Redis
}

// RateLimitConfig returns resolved rate-limit configuration snapshot.
func (r Runtime) RateLimitConfig() config.RateLimitConfig {
	if r.deps == nil {
		return config.RateLimitConfig{}
	}
	return r.deps.RateLimit
}

// CacheConfig returns resolved cache configuration snapshot.
func (r Runtime) CacheConfig() config.CacheConfig {
	if r.deps == nil {
		return config.CacheConfig{}
	}
	return r.deps.Cache
}

// TuningConfig returns release-managed tuning defaults for modules.
func (r Runtime) TuningConfig() config.TuningConfig {
	if r.deps == nil {
		return config.TuningConfig{}
	}
	return r.deps.Tuning
}

// AuthEngine returns configured goAuth engine when auth is enabled.
func (r Runtime) AuthEngine() *goauth.Engine {
	if r.deps == nil {
		return nil
	}
	return r.deps.AuthEngine
}

// AuthMode returns resolved auth mode, optionally overridden per route.
func (r Runtime) AuthMode(overrides ...auth.Mode) auth.Mode {
	if len(overrides) > 0 && overrides[0] != "" {
		return overrides[0]
	}
	if r.deps == nil || r.deps.AuthMode == "" {
		return auth.ModeHybrid
	}
	return r.deps.AuthMode
}

// Limiter returns configured route limiter when rate limiting is enabled.
func (r Runtime) Limiter() ratelimit.Limiter {
	if r.deps == nil {
		return nil
	}
	return r.deps.Limiter
}

// CacheManager returns configured cache manager when caching is enabled.
func (r Runtime) CacheManager() *cache.Manager {
	if r.deps == nil {
		return nil
	}
	return r.deps.CacheMgr
}

// PermissionResolver returns configured project permission resolver.
func (r Runtime) PermissionResolver() permissions.Resolver {
	if r.deps == nil {
		return nil
	}
	return r.deps.PermissionsResolver
}

// EmailSender returns configured transactional sender.
func (r Runtime) EmailSender() coreemail.Sender {
	if r.deps == nil {
		return nil
	}
	return r.deps.EmailSender
}

// WebAppBaseURL returns the configured frontend base URL for auth email links.
func (r Runtime) WebAppBaseURL() string {
	if r.deps == nil {
		return ""
	}
	return r.deps.WebAppBaseURL
}
