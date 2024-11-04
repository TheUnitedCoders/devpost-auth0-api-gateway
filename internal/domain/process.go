package domain

import (
	"errors"
	"net/http"
)

// ProcessRequest ...
type ProcessRequest struct {
	Service    string
	HTTPMethod HTTPMethod
	APIMethod  string
	Path       string
	Query      string
	Body       []byte
	Headers    http.Header
	RemoteAddr string
}

var (
	errEmptyService    = errors.New("service might be not empty")
	errEmptyAPIMethod  = errors.New("API method might be not empty")
	errEmptyHTTPMethod = errors.New("HTTP method might be not empty")
)

// Validate ...
func (r ProcessRequest) Validate() error {
	if r.Service == "" {
		return errEmptyService
	}

	if r.APIMethod == "" {
		return errEmptyAPIMethod
	}

	if r.HTTPMethod == 0 {
		return errEmptyHTTPMethod
	}

	return nil
}
