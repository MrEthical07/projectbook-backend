package policy

import (
	"time"

	"github.com/MrEthical07/superapi/internal/core/cache"
)

// CacheReadOptional applies CacheRead when manager is available; otherwise returns Noop.
func CacheReadOptional(manager *cache.Manager, cfg cache.CacheReadConfig) Policy {
	if manager == nil {
		return Noop()
	}
	return CacheRead(manager, cfg)
}

// CacheInvalidateOptional applies CacheInvalidate when manager is available; otherwise returns Noop.
func CacheInvalidateOptional(manager *cache.Manager, cfg cache.CacheInvalidateConfig) Policy {
	if manager == nil {
		return Noop()
	}
	return CacheInvalidate(manager, cfg)
}

// CacheControlOptional applies private cache-control headers when manager is available; otherwise returns Noop.
func CacheControlOptional(manager *cache.Manager, ttl time.Duration) Policy {
	if manager == nil {
		return Noop()
	}
	return CacheControl(CacheControlConfig{Private: true, MaxAge: ttl})
}
