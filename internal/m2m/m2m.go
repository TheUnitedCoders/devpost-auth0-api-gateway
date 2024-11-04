package m2m

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/clients/auth0"
)

type tokenIssuer interface {
	Token(ctx context.Context, audience string) (*auth0.Token, error)
}

// Source of M2M token.
type Source interface {
	Token() string
}

type impl struct {
	tokenIssuer  tokenIssuer
	currentToken atomic.Pointer[auth0.Token]
	audience     string
}

// Create M2M Source for provided audience.
func Create(ctx context.Context, issuer tokenIssuer, audience string) (Source, error) {
	src := &impl{
		tokenIssuer: issuer,
		audience:    audience,
	}

	if err := src.updater(ctx); err != nil {
		return nil, fmt.Errorf("creating m2m source: %w", err)
	}

	return src, nil
}

func (s *impl) Token() string {
	loadedToken := s.currentToken.Load()
	if loadedToken == nil {
		return ""
	}

	return loadedToken.AccessToken
}

func (s *impl) updater(ctx context.Context) error {
	if err := s.update(ctx); err != nil {
		return fmt.Errorf("first updating m2m source: %w", err)
	}

	go func() {
		timer := time.NewTimer(calculateNextUpdate(s.currentToken.Load().ExpiresIn))
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := s.update(ctx); err != nil {
					slog.Error("Failed to update m2m source", slog.String("err", err.Error()))
					timer.Reset(time.Minute) // let's retry after one minute
					return
				}

				timer.Reset(calculateNextUpdate(s.currentToken.Load().ExpiresIn))
			}
		}
	}()

	return nil
}

func (s *impl) update(ctx context.Context) error {
	newToken, err := s.tokenIssuer.Token(ctx, s.audience)
	if err != nil {
		return fmt.Errorf("updating m2m token: %w", err)
	}

	s.currentToken.Store(newToken)
	return nil
}

func calculateNextUpdate(expiresIn uint) time.Duration {
	return time.Duration(expiresIn)*time.Second - time.Minute*10
}
