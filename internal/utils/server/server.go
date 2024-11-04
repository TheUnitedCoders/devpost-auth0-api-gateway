package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Server is a HTTP server.
type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

// New returns new Server.
func New(address string, handler http.Handler, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	return &Server{
		srv: &http.Server{
			Addr:    address,
			Handler: handler,
		},
		logger: logger,
	}
}

// Run HTTP server.
func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.logger.Info("stopping http server")

		sCtx, sCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer sCancel()

		if err := s.srv.Shutdown(sCtx); err != nil { //nolint:contextcheck
			s.logger.Error("failed to shutdown server", slog.String("err", err.Error()))
		}
	}()

	s.logger.Info("starting http server", slog.String("addr", s.srv.Addr))

	if err := s.srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}

	return nil
}
