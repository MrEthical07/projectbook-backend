package policy

import "github.com/MrEthical07/superapi/internal/core/ratelimit"

// RateLimitOptional applies RateLimit when limiter is available; otherwise returns Noop.
func RateLimitOptional(limiter ratelimit.Limiter, rule ratelimit.Rule) Policy {
	if limiter == nil {
		return Noop()
	}
	return RateLimit(limiter, rule)
}
