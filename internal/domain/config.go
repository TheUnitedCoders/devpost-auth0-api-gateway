package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	defaultPublicListenAddress = ":7070"
	defaultAdminListenAddress  = ":7071"

	defaultDescriptionSyncPeriod   = time.Minute
	defaultServiceOperationTimeout = time.Minute
)

// Config ...
type Config struct {
	PublicListenAddress   string           `json:"public_listen_address"`
	AdminListenAddress    string           `json:"admin_listen_address"`
	Auth0Domain           string           `json:"auth0_domain"`
	Auth0Audience         string           `json:"auth0_audience"`
	Auth0ClientID         string           `json:"auth0_client_id"`
	Auth0ClientSecret     string           `json:"auth0_client_secret"`
	RedisAddress          string           `json:"redis_address"`
	RedisPassword         string           `json:"redis_password"`
	DescriptionSyncPeriod time.Duration    `json:"description_sync_period"`
	Services              []*ConfigService `json:"services"`
}

// ConfigService ...
type ConfigService struct {
	Name             string        `json:"name"`
	Address          string        `json:"address"`
	M2MAudience      string        `json:"m2m_audience"`
	OperationTimeout time.Duration `json:"timeout"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	if len(c.PublicListenAddress) == 0 {
		c.PublicListenAddress = defaultPublicListenAddress
	}

	if len(c.AdminListenAddress) == 0 {
		c.AdminListenAddress = defaultAdminListenAddress
	}

	if c.DescriptionSyncPeriod <= 0 {
		c.DescriptionSyncPeriod = defaultDescriptionSyncPeriod
	}

	if c.RedisAddress == "" {
		c.RedisAddress = "localhost:6379"
	}

	for _, s := range c.Services {
		s.SetDefaults()
	}
}

// Validate ...
func (c *Config) Validate() error {
	if c.PublicListenAddress == "" {
		return errors.New("field PublicListenAddress is required")
	}

	if c.AdminListenAddress == "" {
		return errors.New("field AdminListenAddress is required")
	}

	if c.Auth0Domain == "" {
		return errors.New("field Auth0Domain is required")
	}

	if c.Auth0Audience == "" {
		return errors.New("field Auth0Audience is required")
	}

	if c.Auth0ClientID == "" {
		return errors.New("field Auth0ClientID is required")
	}

	if c.Auth0ClientSecret == "" {
		return errors.New("field Auth0ClientSecret is required")
	}

	if c.DescriptionSyncPeriod <= 0 {
		return errors.New("field DescriptionSyncPeriod must be greater than zero")
	}

	if c.RedisAddress == "" {
		return errors.New("field RedisAddress is required")
	}

	for index, s := range c.Services {
		if err := s.Validate(); err != nil {
			name := s.Name
			if name == "" {
				name = fmt.Sprintf("with index %d", index)
			}

			return fmt.Errorf("service %s is invalid: %w", name, err)
		}
	}

	return nil
}

// SetDefaults ...
func (cs *ConfigService) SetDefaults() {
	if cs.OperationTimeout <= 0 {
		cs.OperationTimeout = defaultServiceOperationTimeout
	}
}

// Validate ...
func (cs *ConfigService) Validate() error {
	if len(cs.Name) == 0 {
		return errors.New("field Name is required")
	}

	if len(cs.Address) == 0 {
		return errors.New("field Address is required")
	}

	if cs.OperationTimeout <= 0 {
		return errors.New("field OperationTimeout must be greater than zero")
	}

	return nil
}
