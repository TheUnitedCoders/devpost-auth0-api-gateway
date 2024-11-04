package sdk

//go:generate go run github.com/abice/go-enum

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// HandlerFunc ...
type HandlerFunc func(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error)

// HTTPMethod ...
// ENUM(unspecified, get, put, post, delete, patch)
type HTTPMethod uint8

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

func (r *RateLimiterDescription) validate() error {
	if r.Rate == 0 {
		return errors.New("rate cannot be 0")
	}

	if r.Burst == 0 {
		return errors.New("burst cannot be 0")
	}

	if r.Period <= 0 {
		return errors.New("period cannot be less or equal than 0")
	}

	return nil
}

// HandlerSettings ...
type HandlerSettings struct {
	AuditEnabled           bool
	RateLimiterDescription *RateLimiterDescription
	RequiredAuthentication bool
	RequiredPermissions    []string
}

func (s *HandlerSettings) validate() error {
	if s.RateLimiterDescription != nil {
		if err := s.RateLimiterDescription.validate(); err != nil {
			return fmt.Errorf("invalid rate limiter description: %w", err)
		}
	}

	return nil
}

// Handler ...
type Handler struct {
	Method string
	HandlerSettings
	AllowedHTTPMethods []HTTPMethod
	ProcessFunc        HandlerFunc
}

func (h *Handler) validate() error {
	if h.Method == "" {
		return errors.New("method is required")
	}

	if err := h.HandlerSettings.validate(); err != nil {
		return fmt.Errorf("invalid handler settings: %w", err)
	}

	if h.ProcessFunc == nil {
		return errors.New("process function is required")
	}

	return nil
}

// SubjectInformation ...
type SubjectInformation struct {
	ID          string
	Permissions []string
}

// ProcessRequest ...
type ProcessRequest struct {
	HTTPMethod         HTTPMethod
	Path               string
	Query              url.Values
	Body               []byte
	Headers            http.Header
	SubjectInformation *SubjectInformation
}

// ProcessResponse ...
type ProcessResponse struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
}
