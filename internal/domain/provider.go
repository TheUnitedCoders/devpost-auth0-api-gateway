package domain

//go:generate go run github.com/abice/go-enum

import (
	"errors"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/utils/slice"
)

const (
	authorizationHeaderName = "Authorization"
	contentTypeHeaderName   = "Content-Type"
)

// RateLimitDescriptionBy ...
// ENUM(ip, subject_id)
type RateLimitDescriptionBy uint8

// RateLimiterDescription ...
type RateLimiterDescription struct {
	By     RateLimitDescriptionBy
	Rate   uint64
	Burst  uint64
	Period time.Duration
}

// ProviderDescription ...
type ProviderDescription struct {
	AuditEnabled          bool
	RateLimiter           *RateLimiterDescription
	RequireAuthentication bool
	RequiredPermissions   []string
	DescriptionByMethod   map[string]*ProviderDescriptionMethod
}

// ProviderDescriptionMethod ...
type ProviderDescriptionMethod struct {
	Method                 string
	AuditEnabled           bool
	RateLimiter            *RateLimiterDescription
	RequiredAuthentication bool
	RequiredPermissions    []string
	AllowedHTTPMethods     mapset.Set[HTTPMethod]
}

// NeedAudit ...
func (p *ProviderDescription) NeedAudit(method string) bool {
	if p.AuditEnabled {
		return true
	}

	if desc, ok := p.DescriptionByMethod[method]; ok {
		return desc.AuditEnabled
	}

	return false
}

// SelectRateLimiter ...
func (p *ProviderDescription) SelectRateLimiter(method string) (description *RateLimiterDescription, isServiceLimiter bool) {
	if desc, ok := p.DescriptionByMethod[method]; ok && desc.RateLimiter != nil {
		return desc.RateLimiter, false
	}

	return p.RateLimiter, true
}

// NeedAuthentication ...
func (p *ProviderDescription) NeedAuthentication(method string) bool {
	if p.RequireAuthentication {
		return true
	}

	if desc, ok := p.DescriptionByMethod[method]; ok {
		return desc.RequiredAuthentication
	}

	return false
}

// Permissions ...
func (p *ProviderDescription) Permissions(method string) []string {
	if desc, ok := p.DescriptionByMethod[method]; ok {
		return slice.Merge(p.RequiredPermissions, desc.RequiredPermissions)
	}

	return p.RequiredPermissions
}

// SubjectInformation ...
type SubjectInformation struct {
	ID          string
	Permissions mapset.Set[string]
}

var (
	errIDMustBeNotEmpty = errors.New("ID cannot be empty")
)

// Validate ...
func (u *SubjectInformation) Validate() error {
	if u.ID == "" {
		return errIDMustBeNotEmpty
	}

	return nil
}

// ProviderProcessRequest ...
type ProviderProcessRequest struct {
	APIMethod          string
	HTTPMethod         HTTPMethod
	Path               string
	Query              string
	Body               []byte
	Headers            http.Header
	SubjectInformation *SubjectInformation
}

// Preprocess ...
func (p *ProviderProcessRequest) Preprocess() {
	p.Headers.Del(authorizationHeaderName)
}

// ProviderProcessResponse ...
type ProviderProcessResponse struct {
	Body       []byte
	StatusCode uint32
	Headers    http.Header
}

// SetDefaults ...
func (p *ProviderProcessResponse) SetDefaults() {
	if len(p.Body) != 0 {
		if p.Headers == nil {
			p.Headers = make(http.Header)
		}

		if val := p.Headers.Get(contentTypeHeaderName); val == "" {
			p.Headers.Set(contentTypeHeaderName, "application/json")
		}
	}
}
