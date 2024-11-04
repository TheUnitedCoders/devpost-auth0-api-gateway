package sdk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/auth"
	provider "github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/pb/contract/v1"
)

// SDK that helps integrate with api-gateway.
type SDK struct {
	serverAddress         string
	tokenParser           tokenParser
	m2mValidation         bool
	globalHandlerSettings HandlerSettings
	handlers              map[string]Handler
	serverCloseTimeout    time.Duration
	logger                *slog.Logger
	run                   atomic.Bool
}

// NewOptions ...
type NewOptions struct {
	ServerAddress         string
	Auth0Domain           string
	Auth0Audience         string
	M2MValidation         bool
	GlobalHandlerSettings HandlerSettings
	Handlers              []Handler
	ServerCloseTimeout    time.Duration
	Logger                *slog.Logger
}

func (opts *NewOptions) setDefault() {
	if opts.ServerAddress == "" {
		opts.ServerAddress = ":8081"
	}

	if opts.ServerCloseTimeout <= 0 {
		opts.ServerCloseTimeout = time.Second * 10
	}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
}

func (opts *NewOptions) validate() error {
	if opts.ServerAddress == "" {
		return errors.New("server address is required")
	}

	if opts.M2MValidation {
		if opts.Auth0Domain == "" {
			return errors.New("auth0 domain is required")
		}

		if opts.Auth0Audience == "" {
			return errors.New("auth0 audience is required")
		}
	}

	if err := opts.GlobalHandlerSettings.validate(); err != nil {
		return fmt.Errorf("failed to validate handlers settings: %w", err)
	}

	for _, handler := range opts.Handlers {
		if err := handler.validate(); err != nil {
			return fmt.Errorf("failed to validate handler: %w", err)
		}
	}

	if opts.ServerCloseTimeout <= 0 {
		return fmt.Errorf("server close-timeout is required")
	}

	if opts.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	return nil
}

// New returns new SDK with provided options.
func New(opts NewOptions) (*SDK, error) {
	opts.setDefault()

	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	var tParser tokenParser
	if opts.Auth0Domain != "" || opts.Auth0Audience != "" {
		tParserImpl, err := auth.New(opts.Auth0Domain, opts.Auth0Audience)
		if err != nil {
			return nil, fmt.Errorf("failed to create auth: %w", err)
		}

		tParser = tParserImpl
	}

	s := &SDK{
		serverAddress:         opts.ServerAddress,
		tokenParser:           tParser,
		m2mValidation:         opts.M2MValidation,
		globalHandlerSettings: opts.GlobalHandlerSettings,
		handlers:              make(map[string]Handler),
		serverCloseTimeout:    opts.ServerCloseTimeout,
		logger:                opts.Logger,
	}

	for _, handler := range opts.Handlers {
		if err := s.RegisterHandler(handler); err != nil {
			return nil, fmt.Errorf("failed to register handler for method %s: %w", handler.Method, err)
		}
	}

	return s, nil
}

var (
	errAlreadyRunning = errors.New("already running")
)

// RegisterHandler in SDK.
// All registrations must be done before Run.
func (s *SDK) RegisterHandler(h Handler) error {
	if s.run.Load() {
		return errAlreadyRunning
	}

	if err := h.validate(); err != nil {
		return fmt.Errorf("failed to validate handler: %w", err)
	}

	if _, exists := s.handlers[h.Method]; exists {
		return errors.New("handler already registered")
	}

	s.handlers[h.Method] = h

	return nil
}

// Run gRPC server.
func (s *SDK) Run(ctx context.Context) error {
	if s.run.Swap(true) {
		return errAlreadyRunning
	}

	lis, err := net.Listen("tcp", s.serverAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	srv := grpc.NewServer()

	provider.RegisterProviderServiceServer(
		srv,
		&server{
			tokenParser:           s.tokenParser,
			globalHandlerSettings: s.globalHandlerSettings,
			handlers:              s.handlers,
		},
	)
	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		s.logger.Info("stopping gRPC server")

		sCtx, sCancel := context.WithTimeout(context.Background(), s.serverCloseTimeout)
		defer sCancel()

		closeChan := make(chan struct{})
		go func() {
			defer close(closeChan)
			srv.GracefulStop()
		}()

		for {
			select {
			case <-closeChan:
				s.logger.Info("gRPC server gracefully stopped")
				return
			case <-sCtx.Done():
				s.logger.Info("failed to gracefully stop gRPC, use immediately closer")
				srv.Stop()
				return
			}
		}
	}()

	s.logger.Info("starting gRPC server", slog.String("addr", lis.Addr().String()))

	if err = srv.Serve(lis); err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}

		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
