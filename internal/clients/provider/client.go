package provider

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/m2m"
	provider "github.com/TheUnitedCoders/devpost-auth0-api-gateway/pkg/pb/contract/v1"
)

// Client to provider.
type Client interface {
	Description(ctx context.Context) (*domain.ProviderDescription, error)
	Process(ctx context.Context, req *domain.ProviderProcessRequest) (*domain.ProviderProcessResponse, error)
}

type impl struct {
	timeout        time.Duration
	client         provider.ProviderServiceClient
	m2mTokenSource m2m.Source
}

// NewOptions ...
type NewOptions struct {
	Name             string
	Address          string
	M2MTokenSource   m2m.Source
	OperationTimeout time.Duration
}

// New returns new Client.
func New(opts NewOptions) (Client, error) {
	conn, err := grpc.NewClient(
		opts.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			clMetrics.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &impl{
		timeout:        opts.OperationTimeout,
		client:         provider.NewProviderServiceClient(conn),
		m2mTokenSource: opts.M2MTokenSource,
	}, nil
}

func (i *impl) Description(ctx context.Context) (*domain.ProviderDescription, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	ctx = i.addM2MToken(ctx)

	resp, err := i.client.Description(ctx, &provider.DescriptionRequest{})
	if err != nil {
		return nil, fmt.Errorf("could not get provider description: %w", err)
	}

	return descriptionFromProto(resp), nil
}

func (i *impl) Process(ctx context.Context, req *domain.ProviderProcessRequest) (*domain.ProviderProcessResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	ctx = i.addM2MToken(ctx)

	resp, err := i.client.Process(ctx, &provider.ProcessRequest{
		ApiMethod:          req.APIMethod,
		HttpMethod:         httpMethodToProto(req.HTTPMethod),
		Path:               req.Path,
		Query:              req.Query,
		Body:               req.Body,
		Headers:            headersToProto(req.Headers),
		SubjectInformation: subjectInformationToProto(req.SubjectInformation),
	})
	if err != nil {
		return nil, fmt.Errorf("could not process: %w", err)
	}

	return &domain.ProviderProcessResponse{
		Body:       resp.GetBody(),
		StatusCode: resp.GetStatusCode(),
		Headers:    headersFromProto(resp.GetHeaders()),
	}, nil
}

const m2mTokenMetadataKey = "x-m2m-token"

func (i *impl) addM2MToken(ctx context.Context) context.Context {
	if i.m2mTokenSource == nil {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, m2mTokenMetadataKey, i.m2mTokenSource.Token())
}
