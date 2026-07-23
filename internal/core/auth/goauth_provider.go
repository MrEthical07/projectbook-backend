package auth

import (
	"fmt"
	"os"
	"strings"

	goauth "github.com/MrEthical07/goAuth"
	"github.com/redis/go-redis/v9"
)

type providerCloser interface {
	Close()
}

const goAuthAuthOnlyPermission = "projectbook.auth_only"

var goAuthAuthOnlyRoles = map[string][]string{
	"user":  {goAuthAuthOnlyPermission},
	"admin": {goAuthAuthOnlyPermission},
}

// NewGoAuthEngine builds a goAuth engine backed by Redis and SQLC user provider.
//
// Usage:
//
//	engine, shutdown, err := auth.NewGoAuthEngine(redisClient, mode, userProvider)
//
// Notes:
// - redisClient must be non-nil
// - shutdown should be called during application shutdown
// - AUTH_TEST_SHARED_SECRET enables deterministic local signer behavior
func NewGoAuthEngine(redisClient redis.UniversalClient, mode Mode, userProvider goauth.UserProvider) (*goauth.Engine, func(), error) {
	if redisClient == nil {
		return nil, nil, fmt.Errorf("goAuth provider requires redis client")
	}

	cfg := projectBookGoAuthConfig(mode)

	// Optional deterministic signer for local perf tests across multiple processes.
	if sharedSecret := strings.TrimSpace(os.Getenv("AUTH_TEST_SHARED_SECRET")); sharedSecret != "" {
		cfg.JWT.SigningMethod = "hs256"
		cfg.JWT.PrivateKey = []byte(sharedSecret)
		cfg.JWT.PublicKey = []byte(sharedSecret)
	}

	engine, err := goauth.New().
		WithConfig(cfg).
		WithRedis(redisClient).
		WithPermissions([]string{goAuthAuthOnlyPermission}).
		WithRoles(goAuthAuthOnlyRoles).
		WithUserProvider(userProvider).
		Build()
	if err != nil {
		return nil, nil, fmt.Errorf("build goAuth engine: %w", err)
	}

	shutdown := func() {
		if closer, ok := any(engine).(providerCloser); ok {
			closer.Close()
		}
	}

	return engine, shutdown, nil
}
