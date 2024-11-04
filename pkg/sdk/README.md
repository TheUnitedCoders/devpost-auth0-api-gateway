# SDK [![Go Reference](https://pkg.go.dev/badge/github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/sdk.svg)](https://pkg.go.dev/github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/sdk)

This module will help you to quickly integrate with the API gateway.

## Installation
Install SDK to your Go project.
```shell
go get github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/sdk
```

## Example of usage
1. SDK Initialization. Create a context for working with the SDK and initialize it with the main parameters.
```go
s, err := sdk.New(sdk.NewOptions{
    ServerAddress: ":8001",
    Auth0Domain:   "<AUTH0_DOMAIN>",
    Auth0Audience: "<AUTH0_AUDIENCE>",
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
```

2. Registering Handlers
```go
err := s.RegisterHandler(sdk.Handler{
    Method: "handler",
    HandlerSettings: sdk.HandlerSettings{
        RateLimiterDescription: &sdk.RateLimiterDescription{
            By:    sdk.RateLimitDescriptionBySubjectId, // Limit by user ID
            Rate:  1, // Number of requests per minute
            Burst: 1, // Burst value
            Period: time.Minute, // Time period
        },
        RequiredPermissions: []string{"read:greeting"}, // Required access permissions
    },
    AllowedHTTPMethods: []sdk.HTTPMethod{sdk.HTTPMethodGet},
    ProcessFunc:        greetingProcess,
})

if err != nil {
    slog.Error("Failed to register greeting handler", slog.String("error", err.Error()))
    return
}
```

3. Define the request processing function that will execute the processing logic and return a response:
```go
func greetingProcess(ctx context.Context, req *sdk.ProcessRequest) (*sdk.ProcessResponse, error) {
    name := req.Query.Get("name")
    if name == "" {
        name = "unknown :("
    }

    return &sdk.ProcessResponse{
        Body: []byte(fmt.Sprintf(`{"msg": "hello from greeting-service to %s with auth0 ID %s"}`, 
            name, req.SubjectInformation.ID)),
        StatusCode: http.StatusOK,
    }, nil
}
```

4. Running the Service
```go
if err = s.Run(ctx); err != nil {
    slog.Error("Failed to run SDK", slog.String("error", err.Error()))
}
```

The full example of usage can be found [here](example/example.go).
Also, full documentation available [here](https://pkg.go.dev/github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/sdk). 
