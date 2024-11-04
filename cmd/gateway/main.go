package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/audit"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/auth"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/clients/auth0"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/clients/provider"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/clients/redis"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/config"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/handlers/admin"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/handlers/gateway"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/m2m"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/processor"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/ratelimit"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/utils/server"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/utils/store"
)

var (
	configPath = flag.String("config-path", "./config.json", "Path to config")
)

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.FromFile(*configPath)
	if err != nil {
		slog.Error("failed to load config file", slog.String("err", err.Error()))
		return
	}

	redisClient, err := redis.New(ctx, cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		slog.Error("failed to create redis client", slog.String("err", err.Error()))
		return
	}

	auth0Client := auth0.New(auth0.NewOptions{
		Domain:       cfg.Auth0Domain,
		ClientID:     cfg.Auth0ClientID,
		ClientSecret: cfg.Auth0ClientSecret,
	})

	clientStore, err := initClientStore(ctx, auth0Client, cfg.Services)
	if err != nil {
		slog.Error("failed to initialize client store", slog.String("err", err.Error()))
		return
	}

	descriptionStore := store.New[string, *domain.ProviderDescription](nil)
	descriptionStoreSync(ctx, cfg.DescriptionSyncPeriod, clientStore, descriptionStore)

	tokenParser, err := auth.New(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		slog.Error("failed to initialize token parser", slog.String("err", err.Error()))
		return
	}

	processorSvc := processor.New(
		descriptionStore,
		clientStore,
		tokenParser,
		audit.NewLogAuditor(slog.With("kind", "auditor")),
		ratelimit.NewRedis(redisClient),
	)

	processorSvc = processor.WithMetricsMiddleware(processorSvc)

	publicServer := server.New(cfg.PublicListenAddress, gateway.Handler(processorSvc), slog.With("kind", "public"))
	adminServer := server.New(cfg.AdminListenAddress, admin.Handler(), slog.With("kind", "admin"))

	eg, eCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return publicServer.Run(eCtx)
	})
	eg.Go(func() error {
		return adminServer.Run(eCtx)
	})

	if err = eg.Wait(); err != nil {
		slog.Error("An un-recoverable error occurred, exiting", slog.String("err", err.Error()))
	}
}

func initClientStore(ctx context.Context, auth0Client *auth0.Client, services []*domain.ConfigService) (*store.Store[string, provider.Client], error) {
	clients := make(map[string]provider.Client)

	for _, service := range services {
		var m2mTokenSource m2m.Source
		if service.M2MAudience != "" {
			src, err := m2m.Create(ctx, auth0Client, service.M2MAudience)
			if err != nil {
				return nil, fmt.Errorf("failed to create m2m token source for provider %s: %w", service.Name, err)
			}

			m2mTokenSource = src
		}

		providerClient, err := provider.New(provider.NewOptions{
			Name:             service.Name,
			Address:          service.Address,
			M2MTokenSource:   m2mTokenSource,
			OperationTimeout: service.OperationTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create client to provider %s: %w", service.Name, err)
		}

		clients[service.Name] = providerClient
	}

	return store.New[string, provider.Client](clients), nil
}

func descriptionStoreSync(ctx context.Context, period time.Duration, clientStore *store.Store[string, provider.Client], descriptionStore *store.Store[string, *domain.ProviderDescription]) {
	syncFunc := func() error {
		var syncErr error

		for name, client := range clientStore.Data() {
			description, err := client.Description(ctx)
			if err != nil {
				// we try to sync all descriptions, but skip errors.
				syncErr = errors.Join(syncErr, fmt.Errorf("could not get description for %s: %w", name, err))
				continue
			}

			descriptionStore.Set(name, description)
		}

		return syncErr
	}

	if err := syncFunc(); err != nil {
		slog.Error("failed to sync some description store entities", slog.String("err", err.Error()))
	}

	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := syncFunc(); err != nil {
					slog.Error("failed to sync some description store entities", slog.String("err", err.Error()))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
