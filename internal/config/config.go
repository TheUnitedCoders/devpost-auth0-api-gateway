package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
)

// FromFile returns config from file.
func FromFile(path string) (*domain.Config, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg domain.Config
	if err = json.Unmarshal(fileContent, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	cfg.SetDefaults()
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config file: %w", err)
	}

	return &cfg, nil
}
