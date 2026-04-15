package config

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	// ProfileMinimal enables a lean feature profile while keeping core stores on.
	ProfileMinimal = "minimal"
	// ProfileDev enables a developer-friendly local stack.
	ProfileDev = "dev"
	// ProfileProd enables production-style defaults.
	ProfileProd = "prod"
)

var profileDefaults = map[string]map[string]string{
	ProfileMinimal: {
		"APP_ENV":                    "dev",
		"AUTH_ENABLED":               "false",
		"EMAIL_ENABLED":              "false",
		"WEB_APP_BASE_URL":           "http://localhost:5173",
		"TRANSACTIONAL_SENDER_EMAIL": "no-reply@projectbook.dev",
		"TRANSACTIONAL_SENDER_NAME":  "no-reply",
		"CACHE_ENABLED":              "false",
		"RATELIMIT_ENABLED":          "false",
		"PERMISSIONS_ENABLED":        "false",
		"MONGO_ENABLED":              "true",
		"MONGO_URL":                  "mongodb://127.0.0.1:27017",
		"POSTGRES_ENABLED":           "true",
		"POSTGRES_URL":               "postgres://superapi:superapi@127.0.0.1:5432/superapi?sslmode=disable",
		"REDIS_ENABLED":              "true",
		"REDIS_ADDR":                 "127.0.0.1:6379",
	},
	ProfileDev: {
		"APP_ENV":                      "dev",
		"AUTH_ENABLED":                 "true",
		"AUTH_MODE":                    "jwt_only",
		"EMAIL_ENABLED":                "false",
		"WEB_APP_BASE_URL":             "http://localhost:5173",
		"TRANSACTIONAL_SENDER_EMAIL":   "no-reply@projectbook.dev",
		"TRANSACTIONAL_SENDER_NAME":    "no-reply",
		"VERIFICATION_SENDER_EMAIL":    "",
		"VERIFICATION_SENDER_NAME":     "",
		"PASSWORD_RESET_SENDER_EMAIL":  "",
		"PASSWORD_RESET_SENDER_NAME":   "",
		"PASSWORD_CHANGE_SENDER_EMAIL": "",
		"PASSWORD_CHANGE_SENDER_NAME":  "",
		"CACHE_ENABLED":                "true",
		"CACHE_FAIL_OPEN":              "true",
		"RATELIMIT_ENABLED":            "true",
		"RATELIMIT_FAIL_OPEN":          "true",
		"RATELIMIT_DEFAULT_LIMIT":      "1000",
		"RATELIMIT_DEFAULT_WINDOW":     "1m",
		"PERMISSIONS_ENABLED":          "true",
		"MONGO_ENABLED":                "true",
		"MONGO_URL":                    "mongodb://127.0.0.1:27017",
		"POSTGRES_ENABLED":             "true",
		"POSTGRES_URL":                 "postgres://superapi:superapi@127.0.0.1:5432/superapi?sslmode=disable",
		"REDIS_ENABLED":                "true",
		"REDIS_ADDR":                   "127.0.0.1:6379",
	},
	ProfileProd: {
		"APP_ENV":                    "prod",
		"AUTH_ENABLED":               "true",
		"AUTH_MODE":                  "strict",
		"EMAIL_ENABLED":              "false",
		"WEB_APP_BASE_URL":           "http://localhost:5173",
		"TRANSACTIONAL_SENDER_EMAIL": "no-reply@projectbook.dev",
		"TRANSACTIONAL_SENDER_NAME":  "no-reply",
		"CACHE_ENABLED":              "true",
		"CACHE_FAIL_OPEN":            "false",
		"RATELIMIT_ENABLED":          "true",
		"RATELIMIT_FAIL_OPEN":        "false",
		"RATELIMIT_DEFAULT_LIMIT":    "100",
		"RATELIMIT_DEFAULT_WINDOW":   "1m",
		"PERMISSIONS_ENABLED":        "true",
		"MONGO_ENABLED":              "true",
		"MONGO_URL":                  "mongodb://127.0.0.1:27017",
		"POSTGRES_ENABLED":           "true",
		"POSTGRES_URL":               "postgres://superapi:superapi@127.0.0.1:5432/superapi?sslmode=disable",
		"REDIS_ENABLED":              "true",
		"REDIS_ADDR":                 "127.0.0.1:6379",
		"METRICS_AUTH_TOKEN":         "change-me",
	},
}

var (
	activeProfileMu       sync.RWMutex
	activeProfileDefaults map[string]string
)

func activateProfile(profileName string) (func(), error) {
	defaults, err := resolveProfileDefaults(profileName)
	if err != nil {
		return nil, err
	}

	activeProfileMu.Lock()
	previous := activeProfileDefaults
	activeProfileDefaults = defaults
	activeProfileMu.Unlock()

	return func() {
		activeProfileMu.Lock()
		activeProfileDefaults = previous
		activeProfileMu.Unlock()
	}, nil
}

func resolveProfileDefaults(profileName string) (map[string]string, error) {
	normalized := strings.ToLower(strings.TrimSpace(profileName))
	if normalized == "" {
		return nil, nil
	}
	defaults, ok := profileDefaults[normalized]
	if !ok {
		allowed := make([]string, 0, len(profileDefaults))
		for k := range profileDefaults {
			allowed = append(allowed, k)
		}
		sort.Strings(allowed)
		return nil, fmt.Errorf("invalid APP_PROFILE %q (valid: %s)", profileName, strings.Join(allowed, ", "))
	}

	copied := make(map[string]string, len(defaults))
	for key, value := range defaults {
		copied[key] = value
	}
	return copied, nil
}

func profileDefaultValue(key string) (string, bool) {
	activeProfileMu.RLock()
	defer activeProfileMu.RUnlock()
	if activeProfileDefaults == nil {
		return "", false
	}
	value, ok := activeProfileDefaults[key]
	return value, ok
}
