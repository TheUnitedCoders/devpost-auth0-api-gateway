package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/processor"
)

const (
	maxBodySize = 1 << 20 // 10mb
)

// Handler is a api-gateway handler.
func Handler(processor processor.Processor) http.HandlerFunc { //revive:disable:import-shadowing
	return func(w http.ResponseWriter, r *http.Request) {
		splitedPath := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 3)
		if len(splitedPath) <= 2 {
			writeJSONError(w, http.StatusBadRequest, "Invalid path")
			return
		}

		serviceName := splitedPath[0]
		method := splitedPath[1]

		var path string
		if len(splitedPath) == 3 {
			path = splitedPath[2]
		}

		body, err := readBody(w, r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Failed to read body: "+err.Error())
			return
		}

		resp := processor.Process(r.Context(), &domain.ProcessRequest{
			Service:    serviceName,
			HTTPMethod: httpMethodToDomain(r.Method),
			APIMethod:  method,
			Path:       path,
			Query:      r.URL.RawQuery,
			Body:       body,
			Headers:    r.Header,
			RemoteAddr: r.RemoteAddr,
		})
		for name, values := range resp.Headers {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(int(resp.StatusCode))
		_, _ = w.Write(resp.Body) //nolint:errcheck
	}
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Body != nil {
		return nil, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	defer r.Body.Close() // nolint:errcheck
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type jsonError struct {
	ErrorMsg string `json:"error_msg,omitempty"`
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	e := jsonError{
		ErrorMsg: message,
	}
	data, _ := json.Marshal(e) //nolint:errcheck
	_, _ = w.Write(data)       //nolint:errcheck
}

func httpMethodToDomain(method string) domain.HTTPMethod {
	switch method {
	case http.MethodGet:
		return domain.HTTPMethodGet
	case http.MethodPut:
		return domain.HTTPMethodPut
	case http.MethodPost:
		return domain.HTTPMethodPost
	case http.MethodDelete:
		return domain.HTTPMethodDelete
	case http.MethodPatch:
		return domain.HTTPMethodPatch
	}

	return domain.HTTPMethodUnspecified
}
