package policy

import (
	"net/http"

	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
)

// ProjectRead returns a validated policy chain for project-scoped read routes.
//
// Usage:
//
//	policies := policy.ProjectRead(
//	    policy.WithAuthEngine(engine, auth.ModeStrict),
//	    policy.WithLimiter(limiter),
//	    policy.WithCacheManager(cacheMgr),
//	)
func ProjectRead(opts ...PresetOption) []Policy {
	cfg := applyPresetOptions(opts...)
	requireAuthEngine("ProjectRead", cfg)
	requireLimiter("ProjectRead", cfg)
	requireCacheManager("ProjectRead", cfg)

	rule := cfg.rateLimitRule
	rule.Scope = ratelimit.ScopeProject

	policies := []Policy{
		AuthRequired(cfg.authEngine, cfg.authMode),
		ProjectRequired(),
		ProjectMatchFromPath(cfg.projectMatchParam),
		RateLimit(cfg.limiter, rule),
		CacheRead(cfg.cacheManager, cache.CacheReadConfig{
			TTL:                cfg.cacheTTL,
			TagSpecs:           append([]cache.CacheTagSpec(nil), cfg.cacheTagSpecs...),
			AllowAuthenticated: cfg.cacheAllowAuth,
			VaryBy:             cfg.cacheVaryBy,
		}),
	}

	mustValidatePreset("ProjectRead", http.MethodGet, "/api/v1/projects/{project_id}/resource/{id}", policies)
	return policies
}

// ProjectWrite returns a validated policy chain for project-scoped write routes.
func ProjectWrite(opts ...PresetOption) []Policy {
	cfg := applyPresetOptions(opts...)
	requireAuthEngine("ProjectWrite", cfg)
	requireLimiter("ProjectWrite", cfg)
	requireCacheManager("ProjectWrite", cfg)

	rule := cfg.rateLimitRule
	rule.Scope = ratelimit.ScopeProject
	tagSpecs := cfg.invalidateTagCfg
	if !cfg.invalidateTagSet {
		tagSpecs = cfg.cacheTagSpecs
	}

	policies := []Policy{
		AuthRequired(cfg.authEngine, cfg.authMode),
		ProjectRequired(),
		ProjectMatchFromPath(cfg.projectMatchParam),
		RateLimit(cfg.limiter, rule),
		CacheInvalidate(cfg.cacheManager, cache.CacheInvalidateConfig{TagSpecs: append([]cache.CacheTagSpec(nil), tagSpecs...)}),
	}

	mustValidatePreset("ProjectWrite", http.MethodPost, "/api/v1/projects/{project_id}/resource", policies)
	return policies
}

// PublicRead returns a validated policy chain for unauthenticated read routes.
func PublicRead(opts ...PresetOption) []Policy {
	cfg := applyPresetOptions(opts...)
	requireLimiter("PublicRead", cfg)
	requireCacheManager("PublicRead", cfg)

	rule := cfg.rateLimitRule
	rule.Scope = ratelimit.ScopeIP

	policies := []Policy{
		RateLimit(cfg.limiter, rule),
		CacheRead(cfg.cacheManager, cache.CacheReadConfig{
			TTL:      cfg.cacheTTL,
			TagSpecs: append([]cache.CacheTagSpec(nil), cfg.cacheTagSpecs...),
			VaryBy:   cache.CacheVaryBy{Method: true},
		}),
	}

	mustValidatePreset("PublicRead", http.MethodGet, "/api/v1/public/resource", policies)
	return policies
}

func requireAuthEngine(name string, cfg presetConfig) {
	if cfg.authEngine == nil {
		panicInvalidRouteConfigf("%s preset requires WithAuthEngine(engine, mode)", name)
	}
}

func requireLimiter(name string, cfg presetConfig) {
	if cfg.limiter == nil {
		panicInvalidRouteConfigf("%s preset requires WithLimiter(limiter)", name)
	}
}

func requireCacheManager(name string, cfg presetConfig) {
	if cfg.cacheManager == nil {
		panicInvalidRouteConfigf("%s preset requires WithCacheManager(manager)", name)
	}
}

func mustValidatePreset(name, method, pattern string, policies []Policy) {
	metas, err := DescribePolicies(policies...)
	if err != nil {
		panicInvalidRouteConfigf("%s preset policies are invalid: %v", name, err)
	}
	if err := ValidateRouteMetadata(method, pattern, metas); err != nil {
		panicInvalidRouteConfigf("%s preset failed validator: %v", name, err)
	}
}
