package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/audit"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/clients/provider"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/ratelimit"
)

const (
	authorizationHeader = "Authorization"
	forwardedForHeader  = "X-Forwarded-For"
)

type descriptionStore interface {
	Get(service string) (*domain.ProviderDescription, bool)
}

type clientStore interface {
	Get(service string) (provider.Client, bool)
}

type tokenParser interface {
	ParseToken(ctx context.Context, token string) (*domain.SubjectInformation, error)
}

// Processor is a api-gateway entrypoint.
type Processor interface {
	Process(ctx context.Context, request *domain.ProcessRequest) *domain.ProviderProcessResponse
}

type impl struct {
	descriptionStore descriptionStore
	clientStore      clientStore
	tokenParser      tokenParser

	auditor     audit.Auditor
	reteLimiter ratelimit.Limiter
}

// New returns new Processor.
func New(descriptionStore descriptionStore, clientStore clientStore, tokenParser tokenParser, auditor audit.Auditor, reteLimiter ratelimit.Limiter) Processor {
	return &impl{
		descriptionStore: descriptionStore,
		clientStore:      clientStore,
		tokenParser:      tokenParser,

		auditor:     auditor,
		reteLimiter: reteLimiter,
	}
}

// Process ...
func (p *impl) Process(ctx context.Context, request *domain.ProcessRequest) *domain.ProviderProcessResponse {
	if err := request.Validate(); err != nil {
		return newErrorResponse(http.StatusBadRequest, fmt.Sprintf("failed to validate process request: %s", err), nil)
	}

	description, ok := p.descriptionStore.Get(request.Service)
	if !ok {
		return newErrorResponse(http.StatusNotFound, fmt.Sprintf("description for service %s not found", request.Service), nil)
	}

	methodDescription, exists := description.DescriptionByMethod[request.APIMethod]
	if !exists {
		return newErrorResponse(http.StatusNotFound, fmt.Sprintf("description for method %s of service %s not found", request.APIMethod, request.Service), nil)
	}

	if !methodDescription.AllowedHTTPMethods.ContainsOne(request.HTTPMethod) {
		return newErrorResponse(http.StatusMethodNotAllowed, fmt.Sprintf("http method %s not allowed", request.HTTPMethod.String()), nil)
	}

	client, exists := p.clientStore.Get(request.Service)
	if !exists {
		return newErrorResponse(http.StatusNotFound, fmt.Sprintf("client for service %s not found", request.Service), nil)
	}

	requiredPermissions := description.Permissions(request.APIMethod)
	needAuthentication := description.NeedAuthentication(request.APIMethod)

	needAuthentication = needAuthentication || len(requiredPermissions) != 0 // handle case when we don't need auth, but permissions was passed

	subjectInformation, err := p.tokenParser.ParseToken(ctx, getAuthorizationHeaderValue(request.Headers))
	if needAuthentication && err != nil {
		return newErrorResponse(http.StatusUnauthorized, fmt.Sprintf("failed to auntificate: %s", err), nil)
	}

	if needAuthentication {
		for _, permission := range requiredPermissions {
			if subjectInformation.Permissions.ContainsOne(permission) {
				continue
			}

			return newErrorResponse(http.StatusForbidden, fmt.Sprintf("subject don't have required permission %s", permission), nil)
		}
	}

	auditFields := audit.Fields{
		Service: request.Service,
		Method:  request.APIMethod,
		Subject: subjectInformation,
		Result:  audit.ResultOk,
	}

	defer func() {
		if !description.NeedAudit(request.APIMethod) {
			return
		}

		p.auditor.Write(ctx, auditFields)
	}()

	rateLimiterDescription, isServiceRateLimiter := description.SelectRateLimiter(request.APIMethod)
	if rateLimiterDescription != nil {
		var entity string

		switch rateLimiterDescription.By {
		case domain.RateLimitDescriptionByIp:
			entity = getRealIP(request.RemoteAddr, request.Headers)
		case domain.RateLimitDescriptionBySubjectId:
			if subjectInformation != nil {
				entity = subjectInformation.ID
			}
		}

		isAllowed, retryAfter, err := p.reteLimiter.Allow(
			ctx,
			ratelimit.Key{
				Service:          request.Service,
				IsServiceLimiter: isServiceRateLimiter,
				Method:           request.APIMethod,
				Entity:           entity,
			},
			ratelimit.Limit{
				Rate:   rateLimiterDescription.Rate,
				Burst:  rateLimiterDescription.Burst,
				Period: rateLimiterDescription.Period,
			},
		)
		if err != nil {
			return newErrorResponse(http.StatusInternalServerError, fmt.Sprintf("internal server error with ratelimiter: %s", err), nil)
		}

		if !isAllowed {
			return newErrorResponse(
				http.StatusTooManyRequests,
				fmt.Sprintf("rate limit exceeded, retry after %s", retryAfter.String()),
				map[string][]string{
					"Retry-After": {strconv.Itoa(int(math.Ceil(retryAfter.Seconds())))},
				},
			)
		}
	}

	processRequest := &domain.ProviderProcessRequest{
		APIMethod:          request.APIMethod,
		HTTPMethod:         request.HTTPMethod,
		Path:               request.Path,
		Query:              request.Query,
		Body:               request.Body,
		Headers:            request.Headers,
		SubjectInformation: subjectInformation,
	}

	processRequest.Preprocess()

	processResp, err := client.Process(ctx, processRequest)
	if err != nil {
		auditFields.Result = audit.ResultError
		return newErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to process request: %s", err), nil)
	}

	processResp.SetDefaults()

	return processResp
}

func getAuthorizationHeaderValue(headers http.Header) string {
	return strings.TrimPrefix(headers.Get(authorizationHeader), "Bearer ")
}

func getRealIP(remoteAddr string, headers http.Header) string {
	realIP := headers.Get(forwardedForHeader)
	if realIP == "" {
		realIP, _, _ = net.SplitHostPort(remoteAddr) //nolint:errcheck
	}

	return realIP
}

type jsonError struct {
	ErrorMsg string `json:"error_msg,omitempty"`
}

func newErrorResponse(code int, msg string, headers http.Header) *domain.ProviderProcessResponse {
	data, _ := json.Marshal(jsonError{ //nolint:errcheck
		ErrorMsg: msg,
	})

	if headers == nil {
		headers = make(http.Header)
	}

	headers.Set("Content-Type", "application/json")

	return &domain.ProviderProcessResponse{
		Body:       data,
		StatusCode: uint32(code),
		Headers:    headers,
	}
}
