package provider

import (
	mapset "github.com/deckarep/golang-set/v2"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/utils/slice"
	provider "github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/pb/contract/v1"
)

func descriptionFromProto(desc *provider.DescriptionResponse) *domain.ProviderDescription {
	descriptionByMethod := make(map[string]*domain.ProviderDescriptionMethod, len(desc.GetMethods()))
	for _, method := range desc.GetMethods() {
		descriptionByMethod[method.GetMethod()] = descriptionMethodFromProto(method)
	}

	return &domain.ProviderDescription{
		AuditEnabled:          desc.GetAuditEnabled(),
		RateLimiter:           rateLimiterFromProto(desc.GetRateLimiter()),
		RequireAuthentication: desc.GetRequiredAuthentication(),
		RequiredPermissions:   desc.GetRequiredPermissions(),
		DescriptionByMethod:   descriptionByMethod,
	}
}

func descriptionMethodFromProto(desc *provider.DescriptionMethod) *domain.ProviderDescriptionMethod {
	return &domain.ProviderDescriptionMethod{
		Method:                 desc.GetMethod(),
		AuditEnabled:           desc.GetAuditEnabled(),
		RateLimiter:            rateLimiterFromProto(desc.GetRateLimiter()),
		RequiredAuthentication: desc.GetRequiredAuthentication(),
		RequiredPermissions:    desc.GetRequiredPermissions(),
		AllowedHTTPMethods:     mapset.NewThreadUnsafeSet(slice.ConvertFunc(desc.GetAllowedHttpMethods(), httpMethodFromProto)...),
	}
}

func httpMethodFromProto(method provider.HttpMethod) domain.HTTPMethod {
	switch method {
	case provider.HttpMethod_HTTP_METHOD_GET:
		return domain.HTTPMethodGet
	case provider.HttpMethod_HTTP_METHOD_PUT:
		return domain.HTTPMethodPut
	case provider.HttpMethod_HTTP_METHOD_POST:
		return domain.HTTPMethodPost
	case provider.HttpMethod_HTTP_METHOD_DELETE:
		return domain.HTTPMethodDelete
	case provider.HttpMethod_HTTP_METHOD_PATCH:
		return domain.HTTPMethodPatch
	}

	return domain.HTTPMethodUnspecified
}

func httpMethodToProto(method domain.HTTPMethod) provider.HttpMethod {
	switch method {
	case domain.HTTPMethodGet:
		return provider.HttpMethod_HTTP_METHOD_GET
	case domain.HTTPMethodPut:
		return provider.HttpMethod_HTTP_METHOD_PUT
	case domain.HTTPMethodPost:
		return provider.HttpMethod_HTTP_METHOD_POST
	case domain.HTTPMethodDelete:
		return provider.HttpMethod_HTTP_METHOD_DELETE
	case domain.HTTPMethodPatch:
		return provider.HttpMethod_HTTP_METHOD_PATCH
	}

	return provider.HttpMethod_HTTP_METHOD_UNSPECIFIED
}

func headersToProto(headers map[string][]string) map[string]*provider.HeaderValue {
	result := make(map[string]*provider.HeaderValue, len(headers))

	for name, values := range headers {
		result[name] = &provider.HeaderValue{
			Values: values,
		}
	}

	return result
}

func headersFromProto(headers map[string]*provider.HeaderValue) map[string][]string {
	result := make(map[string][]string, len(headers))

	for name, values := range headers {
		result[name] = values.GetValues()
	}

	return result
}

func subjectInformationToProto(info *domain.SubjectInformation) *provider.SubjectInformation {
	return &provider.SubjectInformation{
		Id:          info.ID,
		Permissions: info.Permissions.ToSlice(),
	}
}

func rateLimiterFromProto(rateLimiter *provider.RateLimiter) *domain.RateLimiterDescription {
	if rateLimiter == nil {
		return nil
	}

	return &domain.RateLimiterDescription{
		By:     rateLimitByFromProto(rateLimiter.GetBy()),
		Rate:   rateLimiter.GetLimit(),
		Burst:  rateLimiter.GetBurst(),
		Period: rateLimiter.GetPeriod().AsDuration(),
	}
}

func rateLimitByFromProto(by provider.RateLimitBy) domain.RateLimitDescriptionBy {
	switch by {
	case provider.RateLimitBy_RATE_LIMIT_BY_IP:
		return domain.RateLimitDescriptionByIp
	case provider.RateLimitBy_RATE_LIMIT_BY_SUBJECT_ID:
		return domain.RateLimitDescriptionBySubjectId
	}

	return domain.RateLimitDescriptionByIp
}
