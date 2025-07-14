package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// BasicConfigManager implements ConfigManager using Claude's .claude.json files
type BasicConfigManager struct {
	configDir string
	mu        sync.RWMutex
}

// claudeConfig represents the structure of Claude's configuration file
type claudeConfig struct {
	OAuthAccount *oauthAccount              `json:"oauthAccount,omitempty"`
	Other        map[string]json.RawMessage `json:"-"`
}

// oauthAccount represents the OAuth account section in Claude config
type oauthAccount struct {
	EmailAddress string `json:"emailAddress"`
	AccountUUID  string `json:"accountUuid"`
}

// NewBasicConfigManager creates a new basic config manager
func NewBasicConfigManager(configDir string) ports.ConfigManager {
	return &BasicConfigManager{
		configDir: configDir,
	}
}

// GetCurrentAccount reads the current account from Claude config
func (m *BasicConfigManager) GetCurrentAccount(ctx context.Context) (*domain.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configPath := filepath.Join(m.configDir, ".claude.json")
	
	// If config doesn't exist, return nil (no current account)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config map[string]json.RawMessage
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Check if oauthAccount exists
	oauthData, exists := config["oauthAccount"]
	if !exists {
		return nil, nil // No OAuth account configured
	}

	var oauth oauthAccount
	if err := json.Unmarshal(oauthData, &oauth); err != nil {
		return nil, fmt.Errorf("failed to parse oauthAccount: %w", err)
	}

	// Validate required fields
	if oauth.EmailAddress == "" || oauth.AccountUUID == "" {
		return nil, nil // Incomplete account data
	}

	// Create domain account
	// Note: We don't have an alias from Claude config, so we use empty string
	account, err := domain.NewAccount(oauth.EmailAddress, "", oauth.AccountUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create account from config: %w", err)
	}

	return account, nil
}

// SetCurrentAccount updates Claude config with the new account
func (m *BasicConfigManager) SetCurrentAccount(ctx context.Context, account *domain.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	configPath := filepath.Join(m.configDir, ".claude.json")
	
	// Read existing config or create new one
	var config map[string]json.RawMessage
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing config: %w", err)
	} else {
		config = make(map[string]json.RawMessage)
	}

	// Create OAuth account data
	oauth := oauthAccount{
		EmailAddress: string(account.Email()),
		AccountUUID:  account.UUID(),
	}

	oauthData, err := json.Marshal(oauth)
	if err != nil {
		return fmt.Errorf("failed to marshal oauth account: %w", err)
	}

	// Update config with new OAuth account
	config["oauthAccount"] = oauthData

	// Write updated config
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}