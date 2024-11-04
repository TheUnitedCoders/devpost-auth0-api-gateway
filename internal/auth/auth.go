package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	mapset "github.com/deckarep/golang-set/v2"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
)

// Auth ...
type Auth struct {
	validator *validator.Validator
}

type customClaims struct {
	Permissions []string `json:"permissions"`
}

// Validate is just a func to make compatibility with validator.CustomClaims.
func (customClaims) Validate(_ context.Context) error {
	return nil
}

func (c customClaims) UniquePermissions() mapset.Set[string] {
	return mapset.NewThreadUnsafeSet(c.Permissions...)
}

// New returns new Auth.
func New(domain, audience string) (*Auth, error) { //revive:disable:import-shadowing
	iu, err := url.Parse(domain)
	if err != nil {
		return nil, fmt.Errorf("issuer url parse: %w", err)
	}

	if !strings.HasSuffix(domain, "/") {
		domain += "/"
	}

	provider := jwks.NewCachingProvider(iu, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		domain,
		[]string{audience},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &customClaims{}
		}),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &Auth{
		validator: jwtValidator,
	}, nil
}

var (
	errInvalidClaims = errors.New("invalid claims")
)

// ParseToken ...
func (a *Auth) ParseToken(ctx context.Context, token string) (*domain.SubjectInformation, error) {
	data, err := a.validator.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	claims, ok := data.(*validator.ValidatedClaims)
	if !ok {
		return nil, errInvalidClaims
	}

	return claimsToSubjectInformationAdapter(claims)
}

func claimsToSubjectInformationAdapter(claims *validator.ValidatedClaims) (*domain.SubjectInformation, error) {
	var permissions mapset.Set[string]
	cClaims, ok := claims.CustomClaims.(*customClaims)
	if ok {
		permissions = cClaims.UniquePermissions()
	}

	subjectInfo := &domain.SubjectInformation{
		ID:          claims.RegisteredClaims.Subject,
		Permissions: permissions,
	}

	if err := subjectInfo.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate SubjectInformation: %w", err)
	}

	return subjectInfo, nil
}
