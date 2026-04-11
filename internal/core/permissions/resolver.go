package permissions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Resolver resolves effective project permission masks for an authenticated user.
type Resolver interface {
	Resolve(ctx context.Context, userID, projectID string) (Resolution, error)
}

// ObserveFunc receives non-fatal asynchronous resolver events.
type ObserveFunc func(event string, err error)

// Config configures hybrid resolver behavior.
type Config struct {
	RedisTTL        time.Duration
	BackfillTimeout time.Duration
	Observe         ObserveFunc
}

// DefaultConfig returns safe resolver defaults.
func DefaultConfig() Config {
	return Config{
		RedisTTL:        6 * time.Hour,
		BackfillTimeout: 500 * time.Millisecond,
	}
}

// Resolution is the resolved permission state for a user in one project.
type Resolution struct {
	UserID        string
	ProjectID     string
	Role          string
	Mask          uint64
	IsCustom      bool
	UpdatedAtUnix int64
}

// Error sentinels used for precise policy HTTP mapping.
var (
	ErrMembershipNotFound = errors.New("permissions membership not found")
	ErrMaskInconsistent   = errors.New("permissions mask inconsistent")
	ErrDependencyFailure  = errors.New("permissions dependency failure")
)

// IsMembershipNotFound reports whether err indicates missing project membership.
func IsMembershipNotFound(err error) bool {
	return errors.Is(err, ErrMembershipNotFound)
}

// IsMaskInconsistent reports whether err indicates inconsistent stored permission state.
func IsMaskInconsistent(err error) bool {
	return errors.Is(err, ErrMaskInconsistent)
}

// IsDependencyFailure reports whether err indicates infrastructure dependency failure.
func IsDependencyFailure(err error) bool {
	return errors.Is(err, ErrDependencyFailure)
}

type userMaskCacheEntry struct {
	ProjectID     string `json:"project_id,omitempty"`
	Mask          uint64 `json:"mask"`
	Role          string `json:"role,omitempty"`
	IsCustom      bool   `json:"is_custom"`
	UpdatedAtUnix int64  `json:"updated_at_unix"`
}

// HybridResolver performs Redis-first reads with DB fallback and async backfill.
type HybridResolver struct {
	redis    redis.UniversalClient
	fallback Resolver
	cfg      Config
}

// NewHybridResolver constructs a resolver with Redis-first reads and fallback resolution.
func NewHybridResolver(redisClient redis.UniversalClient, fallback Resolver, cfg Config) (*HybridResolver, error) {
	if fallback == nil {
		return nil, fmt.Errorf("hybrid resolver requires fallback resolver")
	}
	if cfg.RedisTTL <= 0 {
		cfg.RedisTTL = DefaultConfig().RedisTTL
	}
	if cfg.BackfillTimeout <= 0 {
		cfg.BackfillTimeout = DefaultConfig().BackfillTimeout
	}
	return &HybridResolver{redis: redisClient, fallback: fallback, cfg: cfg}, nil
}

// Resolve resolves project permission mask using Redis first and DB fallback.
func (r *HybridResolver) Resolve(ctx context.Context, userID, projectID string) (Resolution, error) {
	userID = strings.TrimSpace(userID)
	projectID = strings.TrimSpace(projectID)
	if userID == "" || projectID == "" {
		return Resolution{}, fmt.Errorf("%w: missing user or project id", ErrMaskInconsistent)
	}

	if r.redis != nil {
		if cached, found, err := r.resolveFromCache(ctx, userID, projectID); err == nil && found {
			return cached, nil
		}
	}

	resolved, err := r.fallback.Resolve(ctx, userID, projectID)
	if err != nil {
		return Resolution{}, err
	}

	if r.redis != nil {
		r.backfillAsync(userID, projectID, resolved)
	}

	return resolved, nil
}

func (r *HybridResolver) resolveFromCache(ctx context.Context, userID, projectID string) (Resolution, bool, error) {
	payload, err := r.redis.Get(ctx, rbacKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Resolution{}, false, nil
		}
		return Resolution{}, false, err
	}

	cacheMap, err := decodeUserMaskCache(payload)
	if err != nil {
		return Resolution{}, false, nil
	}

	entry, ok := cacheMap[projectID]
	if !ok {
		return Resolution{}, false, nil
	}

	return Resolution{
		UserID:        userID,
		ProjectID:     resolvedProjectID(projectID, entry.ProjectID),
		Role:          entry.Role,
		Mask:          entry.Mask,
		IsCustom:      entry.IsCustom,
		UpdatedAtUnix: entry.UpdatedAtUnix,
	}, true, nil
}

func (r *HybridResolver) backfillAsync(userID, projectID string, resolved Resolution) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.cfg.BackfillTimeout)
		defer cancel()

		if err := r.backfillOnce(ctx, userID, projectID, resolved); err != nil && r.cfg.Observe != nil {
			r.cfg.Observe("permissions_backfill_error", err)
		}
	}()
}

// backfillOnce performs one idempotent stale-safe write attempt and never retries indefinitely.
func (r *HybridResolver) backfillOnce(ctx context.Context, userID, projectID string, resolved Resolution) error {
	key := rbacKey(userID)

	return r.redis.Watch(ctx, func(tx *redis.Tx) error {
		currentPayload, err := tx.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return err
		}

		cacheMap, err := decodeUserMaskCache(currentPayload)
		if err != nil {
			cacheMap = make(map[string]userMaskCacheEntry)
		}

		entry := userMaskCacheEntry{
			ProjectID:     strings.TrimSpace(resolved.ProjectID),
			Mask:          resolved.Mask,
			Role:          resolved.Role,
			IsCustom:      resolved.IsCustom,
			UpdatedAtUnix: resolved.UpdatedAtUnix,
		}

		keys := []string{projectID}
		canonicalKey := strings.TrimSpace(resolved.ProjectID)
		if canonicalKey != "" && !strings.EqualFold(canonicalKey, projectID) {
			keys = append(keys, canonicalKey)
		}

		changed := false
		for _, key := range keys {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				continue
			}
			if existing, ok := cacheMap[trimmedKey]; ok && existing.UpdatedAtUnix > resolved.UpdatedAtUnix {
				continue
			}
			cacheMap[trimmedKey] = entry
			changed = true
		}

		if !changed {
			return nil
		}

		updatedPayload, err := json.Marshal(cacheMap)
		if err != nil {
			return err
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, updatedPayload, r.cfg.RedisTTL)
			return nil
		})
		if errors.Is(err, redis.TxFailedErr) {
			return nil
		}
		return err
	}, key)
}

func decodeUserMaskCache(payload string) (map[string]userMaskCacheEntry, error) {
	if strings.TrimSpace(payload) == "" {
		return make(map[string]userMaskCacheEntry), nil
	}
	cacheMap := make(map[string]userMaskCacheEntry)
	if err := json.Unmarshal([]byte(payload), &cacheMap); err != nil {
		return nil, err
	}
	if cacheMap == nil {
		cacheMap = make(map[string]userMaskCacheEntry)
	}
	return cacheMap, nil
}

func resolvedProjectID(requested, resolved string) string {
	resolved = strings.TrimSpace(resolved)
	if resolved != "" {
		return resolved
	}
	return strings.TrimSpace(requested)
}

// ResolverCacheKey returns the Redis key used for user permission cache maps.
func ResolverCacheKey(userID string) string {
	return rbacKey(userID)
}

// InvalidateResolverUserCache removes cached resolver maps for the provided users.
func InvalidateResolverUserCache(ctx context.Context, redisClient redis.UniversalClient, userIDs []string) error {
	if redisClient == nil || len(userIDs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		key := ResolverCacheKey(userID)
		if strings.TrimSpace(key) == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("invalidate resolver user cache: %w", err)
	}

	return nil
}

func rbacKey(userID string) string {
	return "rbac:" + strings.TrimSpace(userID)
}
