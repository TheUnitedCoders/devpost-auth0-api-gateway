package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/sdk"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s, err := sdk.New(sdk.NewOptions{
		ServerAddress: ":8001",
		Auth0Domain:   "<AUTH0 DOMAIN>",
		Auth0Audience: "<AUTH0 AUDIENCE>",
		M2MValidation: true,
		GlobalHandlerSettings: sdk.HandlerSettings{
			AuditEnabled:           true,
			RequiredAuthentication: true,
		},
	})
	if err != nil {
		slog.Error("Failed to init SDK", slog.String("error", err.Error()))
		return
	}

	err = s.RegisterHandler(sdk.Handler{
		Method: "handler",
		HandlerSettings: sdk.HandlerSettings{
			RateLimiterDescription: &sdk.RateLimiterDescription{
				By:     sdk.RateLimitDescriptionBySubjectId,
				Rate:   10,
				Burst:  20,
				Period: time.Minute,
			},
			RequiredPermissions: []string{"read:test"},
		},
		AllowedHTTPMethods: []sdk.HTTPMethod{sdk.HTTPMethodGet},
		ProcessFunc:        greetingProcess,
	})
	if err != nil {
		slog.Error("Failed to register greeting handler", slog.String("error", err.Error()))
		return
	}

	if err = s.Run(ctx); err != nil {
		slog.Error("Failed run SDK", slog.String("error", err.Error()))
	}
}

func greetingProcess(_ context.Context, req *sdk.ProcessRequest) (*sdk.ProcessResponse, error) {
	name := req.Query.Get("name")
	if name == "" {
		name = "unknown :("
	}

	return &sdk.ProcessResponse{
		Body:       []byte(fmt.Sprintf(`{"msg": "hello from greeting-service to %s with auth0 ID %s"}`, name, req.SubjectInformation.ID)),
		StatusCode: http.StatusOK,
	}, nil
}
