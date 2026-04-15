package app

import (
	"context"
	"fmt"
	"strings"

	goauth "github.com/MrEthical07/goAuth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/MrEthical07/superapi/internal/core/auth"
	corecache "github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/config"
	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	"github.com/MrEthical07/superapi/internal/core/metrics"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/readiness"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/MrEthical07/superapi/internal/core/tracing"
	infracache "github.com/MrEthical07/superapi/internal/infrastructure/cache"
	infradb "github.com/MrEthical07/superapi/internal/infrastructure/db"
	infraemail "github.com/MrEthical07/superapi/internal/infrastructure/email"
	infstore "github.com/MrEthical07/superapi/internal/infrastructure/store"
)

// START HERE:
// - This file wires process dependencies (Postgres, Redis, auth, cache, tracing).
// - Module routes should consume these via app.DependencyBinder or modulekit.Runtime.
//
// WARNING:
// This is core infrastructure code.
// Avoid modifying dependency ordering unless you understand startup and readiness behavior.

// Dependencies stores initialized process-level services shared with modules.
type Dependencies struct {
	// Postgres is the optional pgx pool initialized from config.
	Postgres *pgxpool.Pool
	// Store is the primary store surface used by modules.
	Store storage.Store
	// RelationalStore is the relational execution store.
	RelationalStore storage.RelationalStore
	// DocumentStore is the document execution store.
	DocumentStore storage.DocumentStore
	// Redis is the optional Redis client used by auth/cache/ratelimit.
	Redis *redis.Client
	// Mongo is the optional MongoDB client used by document store.
	Mongo *mongo.Client
	// Readiness aggregates health checks for readiness responses.
	Readiness *readiness.Service
	// Metrics is the Prometheus instrumentation service.
	Metrics *metrics.Service
	// Tracing is the OpenTelemetry lifecycle service.
	Tracing *tracing.Service
	// AuthEngine is the optional goAuth engine.
	AuthEngine *goauth.Engine
	// AuthMode is the normalized auth mode used by auth policies.
	AuthMode auth.Mode
	// RateLimit is the resolved rate-limit config snapshot.
	RateLimit config.RateLimitConfig
	// Cache is the resolved cache config snapshot.
	Cache config.CacheConfig
	// Tuning is the release-managed tuning profile snapshot.
	Tuning config.TuningConfig
	// Limiter is the optional route rate limiter.
	Limiter ratelimit.Limiter
	// CacheMgr is the optional response cache manager.
	CacheMgr *corecache.Manager
	// PermissionsResolver resolves effective project permission masks.
	PermissionsResolver permissions.Resolver
	// EmailSender sends transactional emails.
	EmailSender coreemail.Sender
	// WebAppBaseURL is used to compose frontend links in outgoing auth emails.
	WebAppBaseURL string
	authClose     func()
}

// DependencyBinder allows modules to receive initialized Dependencies.
type DependencyBinder interface {
	BindDependencies(*Dependencies)
}

func initDependencies(ctx context.Context, cfg *config.Config) (*Dependencies, error) {
	deps := &Dependencies{
		Readiness:     readiness.NewService(),
		RateLimit:     cfg.RateLimit,
		Cache:         cfg.Cache,
		Tuning:        cfg.Tuning,
		DocumentStore: storage.NoopDocumentStore{},
		EmailSender:   coreemail.NoopSender{},
		WebAppBaseURL: strings.TrimSpace(cfg.Email.WebAppBaseURL),
	}

	if !cfg.Postgres.Enabled {
		return nil, fmt.Errorf("postgres must be enabled for api startup")
	}
	if !cfg.Redis.Enabled {
		return nil, fmt.Errorf("redis must be enabled for api startup")
	}
	if !cfg.Mongo.Enabled {
		return nil, fmt.Errorf("mongo must be enabled for api startup")
	}

	if cfg.Postgres.Enabled {
		pool, err := infradb.NewPostgresPool(ctx, cfg.Postgres)
		if err != nil {
			return nil, fmt.Errorf("init postgres: %w", err)
		}

		infraStore, err := infstore.NewPostgresStore(pool)
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("init relational store: %w", err)
		}

		deps.Postgres = pool
		deps.RelationalStore = infraStore.Relational()
		deps.Store = infraStore.Relational()
		deps.Readiness.Add("postgres", true, cfg.Postgres.HealthCheckTimeout, func(checkCtx context.Context) error {
			return infradb.CheckPostgresHealth(checkCtx, pool, cfg.Postgres.HealthCheckTimeout)
		})
	} else {
		deps.Readiness.Add("postgres", false, cfg.Postgres.HealthCheckTimeout, nil)
	}

	if cfg.Redis.Enabled {
		client, err := infracache.NewRedisClient(ctx, cfg.Redis)
		if err != nil {
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init redis: %w", err)
		}
		deps.Redis = client
		deps.Readiness.Add("redis", true, cfg.Redis.HealthCheckTimeout, func(checkCtx context.Context) error {
			return infracache.CheckRedisHealth(checkCtx, client, cfg.Redis.HealthCheckTimeout)
		})
	} else {
		deps.Readiness.Add("redis", false, cfg.Redis.HealthCheckTimeout, nil)
	}

	metricsSvc, err := metrics.New(cfg.Metrics, deps.Postgres)
	if err != nil {
		if deps.Redis != nil {
			_ = deps.Redis.Close()
		}
		if deps.Postgres != nil {
			deps.Postgres.Close()
		}
		return nil, fmt.Errorf("init metrics: %w", err)
	}
	deps.Metrics = metricsSvc

	authMode, err := auth.ParseMode(cfg.Auth.Mode)
	if err != nil {
		if deps.Redis != nil {
			_ = deps.Redis.Close()
		}
		if deps.Postgres != nil {
			deps.Postgres.Close()
		}
		return nil, fmt.Errorf("init auth mode: %w", err)
	}
	deps.AuthMode = authMode
	deps.AuthEngine = nil

	if cfg.Auth.Enabled {
		if deps.RelationalStore == nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init auth provider: relational store unavailable")
		}

		userRepo := auth.NewRelationalUserRepository(deps.RelationalStore)
		if userRepo == nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init auth provider: user repository unavailable")
		}

		engine, closeFn, err := auth.NewGoAuthEngine(deps.Redis, authMode, auth.NewStoreUserProvider(userRepo))
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init auth provider: %w", err)
		}
		deps.AuthEngine = engine
		deps.authClose = closeFn
	}

	if cfg.Email.Enabled {
		transactionalSender := coreemail.NormalizeSenderIdentity(coreemail.SenderIdentity{
			Name:  cfg.Email.TransactionalSenderName,
			Email: cfg.Email.TransactionalSenderEmail,
		})
		verificationSender := resolveFlowSenderIdentity(
			cfg.Email.VerificationSenderName,
			cfg.Email.VerificationSenderEmail,
			transactionalSender,
		)
		passwordResetSender := resolveFlowSenderIdentity(
			cfg.Email.PasswordResetSenderName,
			cfg.Email.PasswordResetSenderEmail,
			transactionalSender,
		)
		passwordChangeSender := resolveFlowSenderIdentity(
			cfg.Email.PasswordChangeSenderName,
			cfg.Email.PasswordChangeSenderEmail,
			transactionalSender,
		)

		sender, err := infraemail.NewResendSender(cfg.Email.ResendAPIKey, infraemail.SenderProfiles{
			Transactional:  transactionalSender,
			Verification:   verificationSender,
			PasswordReset:  passwordResetSender,
			PasswordChange: passwordChangeSender,
		})
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init resend sender: %w", err)
		}
		deps.EmailSender = sender
	}

	if cfg.RateLimit.Enabled {
		limiter, err := ratelimit.NewRedisLimiter(deps.Redis, ratelimit.Config{
			Env:      cfg.Env,
			FailOpen: cfg.RateLimit.FailOpen,
			Observe: func(route string, outcome ratelimit.Outcome) {
				if deps.Metrics == nil {
					return
				}
				deps.Metrics.ObserveRateLimit(route, string(outcome))
			},
		})
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init rate limiter: %w", err)
		}
		deps.Limiter = limiter
	}

	if cfg.Cache.Enabled {
		cacheMgr, err := corecache.NewManager(deps.Redis, corecache.ManagerConfig{
			Env:                cfg.Env,
			FailOpen:           cfg.Cache.FailOpen,
			DefaultMaxBytes:    cfg.Cache.DefaultMaxBytes,
			TagVersionCacheTTL: cfg.Cache.TagVersionCacheTTL,
			Observe: func(route, outcome string) {
				if deps.Metrics == nil {
					return
				}
				deps.Metrics.ObserveCache(route, outcome)
			},
		})
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init cache manager: %w", err)
		}
		deps.CacheMgr = cacheMgr
	}

	if cfg.Permissions.Enabled {
		if deps.RelationalStore == nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init permissions resolver: relational store unavailable")
		}

		fallbackResolver, err := permissions.NewRelationalResolver(deps.RelationalStore, cfg.Permissions.DBQueryTimeout)
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init permissions fallback resolver: %w", err)
		}

		hybridResolver, err := permissions.NewHybridResolver(deps.Redis, fallbackResolver, permissions.Config{
			RedisTTL:        cfg.Permissions.RedisTTL,
			BackfillTimeout: cfg.Permissions.BackfillTimeout,
		})
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init permissions hybrid resolver: %w", err)
		}

		deps.PermissionsResolver = hybridResolver
		var permissionTagInvalidator permissions.TagInvalidator
		if deps.CacheMgr != nil {
			permissionTagInvalidator = deps.CacheMgr
		}

		lifecycle, err := permissions.NewLifecycle(deps.RelationalStore, deps.Redis, permissionTagInvalidator)
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init permissions lifecycle: %w", err)
		}

		if err := lifecycle.RunStartup(ctx); err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("run permissions lifecycle: %w", err)
		}
	}

	if cfg.Mongo.Enabled {
		mongoClient, err := infradb.NewMongoClient(ctx, cfg.Mongo)
		if err != nil {
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init mongo: %w", err)
		}

		mongoDatabase, err := infradb.NewMongoDatabase(mongoClient, cfg.Mongo.Database)
		if err != nil {
			_ = mongoClient.Disconnect(context.Background())
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init mongo database: %w", err)
		}

		if cfg.Mongo.BootstrapEnabled {
			bootstrapCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.BootstrapTimeout)
			err = infradb.BootstrapMongoProjectBookCollections(bootstrapCtx, mongoDatabase)
			cancel()
			if err != nil {
				_ = mongoClient.Disconnect(context.Background())
				if deps.Redis != nil {
					_ = deps.Redis.Close()
				}
				if deps.Postgres != nil {
					deps.Postgres.Close()
				}
				return nil, fmt.Errorf("mongo bootstrap: %w", err)
			}
		}

		documentStore, err := storage.NewMongoDocumentStore(mongoClient, mongoDatabase)
		if err != nil {
			_ = mongoClient.Disconnect(context.Background())
			if deps.Redis != nil {
				_ = deps.Redis.Close()
			}
			if deps.Postgres != nil {
				deps.Postgres.Close()
			}
			return nil, fmt.Errorf("init mongo document store: %w", err)
		}

		deps.Mongo = mongoClient
		deps.DocumentStore = documentStore
		deps.Readiness.Add("mongo", true, cfg.Mongo.HealthCheckTimeout, func(checkCtx context.Context) error {
			return infradb.CheckMongoHealth(checkCtx, mongoClient, cfg.Mongo.HealthCheckTimeout)
		})
	} else {
		deps.Readiness.Add("mongo", false, cfg.Mongo.HealthCheckTimeout, nil)
	}

	tracingSvc, err := tracing.New(ctx, cfg.Tracing, cfg.Env)
	if err != nil {
		if deps.Mongo != nil {
			_ = deps.Mongo.Disconnect(context.Background())
		}
		if deps.Redis != nil {
			_ = deps.Redis.Close()
		}
		if deps.Postgres != nil {
			deps.Postgres.Close()
		}
		return nil, fmt.Errorf("init tracing: %w", err)
	}
	deps.Tracing = tracingSvc

	return deps, nil
}

func resolveFlowSenderIdentity(name, email string, fallback coreemail.SenderIdentity) coreemail.SenderIdentity {
	resolved := coreemail.NormalizeSenderIdentity(coreemail.SenderIdentity{
		Name:  name,
		Email: email,
	})
	if resolved.Email == "" {
		return fallback
	}
	if resolved.Name == "" {
		resolved.Name = fallback.Name
	}
	return resolved
}

func (a *App) closeDependencies() {
	if a == nil || a.deps == nil {
		return
	}

	if a.deps.Mongo != nil {
		if err := a.deps.Mongo.Disconnect(context.Background()); err != nil {
			a.log.Error().Err(err).Msg("mongo close error")
		}
	}

	if a.deps.Redis != nil {
		if err := a.deps.Redis.Close(); err != nil {
			a.log.Error().Err(err).Msg("redis close error")
		}
	}
	if a.deps.Postgres != nil {
		a.deps.Postgres.Close()
	}
	if a.deps.Tracing != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := a.deps.Tracing.Shutdown(shutdownCtx); err != nil {
			a.log.Error().Err(err).Msg("tracing shutdown error")
		}
	}
	if a.deps.authClose != nil {
		a.deps.authClose()
	}
}
