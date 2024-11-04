package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client ...
type Client struct {
	domain       string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewOptions ...
type NewOptions struct {
	Domain       string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

func (opts *NewOptions) setDefaults() {
	if !strings.HasSuffix(opts.Domain, "/") {
		opts.Domain += "/"
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}
}

// New returns new Client.
func New(opts NewOptions) *Client {
	opts.setDefaults()

	return &Client{
		domain:       opts.Domain,
		clientID:     opts.ClientID,
		clientSecret: opts.ClientSecret,
		httpClient:   opts.HTTPClient,
	}
}

// Token ...
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint   `json:"expires_in"`
}

type tokenRequest struct {
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Token returns new Token from Auth0 with provided audience.
func (c *Client) Token(ctx context.Context, audience string) (*Token, error) {
	jsonRequest, err := json.Marshal(tokenRequest{
		Audience:     audience,
		GrantType:    "client_credentials",
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.domain+"oauth/token", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck

	readData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got response with unexpected status %d: %s", resp.StatusCode, string(readData))
	}

	var token Token
	if err = json.Unmarshal(readData, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}
