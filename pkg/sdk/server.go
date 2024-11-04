package sdk

import (
	"context"
	"net/http"
	"net/url"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/utils/slice"
	provider "github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/pb/contract/v1"
)

type tokenParser interface {
	ParseToken(ctx context.Context, token string) (*domain.SubjectInformation, error)
}

type server struct {
	provider.ProviderServiceServer
	tokenParser tokenParser

	globalHandlerSettings HandlerSettings
	handlers              map[string]Handler
}

func (s *server) Description(ctx context.Context, _ *provider.DescriptionRequest) (*provider.DescriptionResponse, error) {
	if !s.validateM2M(ctx) {
		return nil, status.Error(codes.Unauthenticated, "Failed to validate M2M token")
	}

	methods := make([]*provider.DescriptionMethod, 0, len(s.handlers))
	for _, method := range s.handlers {
		methods = append(methods, &provider.DescriptionMethod{
			Method:                 method.Method,
			AuditEnabled:           method.AuditEnabled,
			RequiredAuthentication: method.RequiredAuthentication,
			RateLimiter:            rateLimiterToProto(method.RateLimiterDescription),
			RequiredPermissions:    method.RequiredPermissions,
			AllowedHttpMethods:     slice.ConvertFunc(method.AllowedHTTPMethods, httpMethodToProto),
		})
	}

	return &provider.DescriptionResponse{
		AuditEnabled:           s.globalHandlerSettings.AuditEnabled,
		RequiredAuthentication: s.globalHandlerSettings.RequiredAuthentication,
		RateLimiter:            rateLimiterToProto(s.globalHandlerSettings.RateLimiterDescription),
		RequiredPermissions:    s.globalHandlerSettings.RequiredPermissions,
		Methods:                methods,
	}, nil
}

func (s *server) Process(ctx context.Context, req *provider.ProcessRequest) (*provider.ProcessResponse, error) {
	if !s.validateM2M(ctx) {
		return nil, status.Error(codes.Unauthenticated, "Failed to validate M2M token")
	}

	handler, ok := s.handlers[req.GetApiMethod()]
	if !ok {
		return nil, status.Error(codes.NotFound, "API method not found")
	}

	queryValues, _ := url.ParseQuery(req.GetQuery()) //nolint:errcheck

	resp, err := handler.ProcessFunc(ctx, &ProcessRequest{
		HTTPMethod:         httpMethodFromProto(req.GetHttpMethod()),
		Path:               req.GetPath(),
		Query:              queryValues,
		Body:               req.GetBody(),
		Headers:            headersFromProto(req.GetHeaders()),
		SubjectInformation: subjectInformationFromProto(req.GetSubjectInformation()),
	})
	if err != nil {
		return nil, err
	}

	return &provider.ProcessResponse{
		Body:       resp.Body,
		StatusCode: uint32(resp.StatusCode),
		Headers:    headersToProto(resp.Headers),
	}, nil
}

func (s *server) validateM2M(ctx context.Context) bool {
	if s.tokenParser == nil {
		return true
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	val := md.Get("x-m2m-token")
	if len(val) == 0 {
		return false
	}

	if _, err := s.tokenParser.ParseToken(ctx, val[0]); err != nil {
		return false
	}

	return true
}

func headersFromProto(headers map[string]*provider.HeaderValue) map[string][]string {
	result := make(map[string][]string, len(headers))

	for name, vals := range headers {
		result[name] = vals.GetValues()
	}

	return result
}

func headersToProto(headers http.Header) map[string]*provider.HeaderValue {
	result := make(map[string]*provider.HeaderValue, len(headers))

	for name, vals := range headers {
		result[name] = &provider.HeaderValue{
			Values: vals,
		}
	}

	return result
}

func subjectInformationFromProto(info *provider.SubjectInformation) *SubjectInformation {
	if info == nil {
		return nil
	}

	return &SubjectInformation{
		ID:          info.GetId(),
		Permissions: info.GetPermissions(),
	}
}

func httpMethodFromProto(method provider.HttpMethod) HTTPMethod {
	switch method {
	case provider.HttpMethod_HTTP_METHOD_GET:
		return HTTPMethodGet
	case provider.HttpMethod_HTTP_METHOD_PUT:
		return HTTPMethodPut
	case provider.HttpMethod_HTTP_METHOD_POST:
		return HTTPMethodPost
	case provider.HttpMethod_HTTP_METHOD_DELETE:
		return HTTPMethodDelete
	case provider.HttpMethod_HTTP_METHOD_PATCH:
		return HTTPMethodPatch
	}

	return HTTPMethodUnspecified
}

func httpMethodToProto(method HTTPMethod) provider.HttpMethod {
	switch method {
	case HTTPMethodGet:
		return provider.HttpMethod_HTTP_METHOD_GET
	case HTTPMethodPut:
		return provider.HttpMethod_HTTP_METHOD_PUT
	case HTTPMethodPost:
		return provider.HttpMethod_HTTP_METHOD_POST
	case HTTPMethodDelete:
		return provider.HttpMethod_HTTP_METHOD_DELETE
	case HTTPMethodPatch:
		return provider.HttpMethod_HTTP_METHOD_PATCH
	}

	return provider.HttpMethod_HTTP_METHOD_UNSPECIFIED
}

func rateLimiterToProto(rateLimiter *RateLimiterDescription) *provider.RateLimiter {
	if rateLimiter == nil {
		return nil
	}

	var by provider.RateLimitBy
	switch rateLimiter.By {
	case RateLimitDescriptionByIp:
		by = provider.RateLimitBy_RATE_LIMIT_BY_IP
	case RateLimitDescriptionBySubjectId:
		by = provider.RateLimitBy_RATE_LIMIT_BY_SUBJECT_ID
	}

	return &provider.RateLimiter{
		By:     by,
		Limit:  rateLimiter.Rate,
		Burst:  rateLimiter.Burst,
		Period: durationpb.New(rateLimiter.Period),
	}
}
