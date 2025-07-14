package json

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestBasicConfigManager_GetCurrentAccount(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configManager := NewBasicConfigManager(tmpDir)
	ctx := context.Background()

	// Test when no config exists
	account, err := configManager.GetCurrentAccount(ctx)
	if err != nil {
		t.Fatalf("Expected no error when no config exists, got: %v", err)
	}
	if account != nil {
		t.Fatal("Expected nil account when no config exists")
	}

	// Create a test config file with proper oauthAccount structure
	claudeConfig := map[string]interface{}{
		"oauthAccount": map[string]interface{}{
			"emailAddress": "test@example.com",
			"accountUuid":  "test-uuid-12345",
		},
		"other_setting": "should-be-preserved",
	}

	configData, _ := json.MarshalIndent(claudeConfig, "", "  ")
	configPath := filepath.Join(tmpDir, ".claude.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test reading existing config
	account, err = configManager.GetCurrentAccount(ctx)
	if err != nil {
		t.Fatalf("Failed to get current account: %v", err)
	}

	if account == nil {
		t.Fatal("Expected account to be found")
	}

	if string(account.Email()) != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", string(account.Email()))
	}
}

func TestBasicConfigManager_SetCurrentAccount(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configManager := NewBasicConfigManager(tmpDir)
	ctx := context.Background()

	// Create test account
	account, err := domain.NewAccount(
		"newuser@example.com",
		"new-alias",
		"new-uuid-12345",
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Test setting account
	err = configManager.SetCurrentAccount(ctx, account)
	if err != nil {
		t.Fatalf("Failed to set current account: %v", err)
	}

	// Verify config file was created/updated
	configPath := filepath.Join(tmpDir, ".claude.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal(".claude.json file was not created")
	}

	// Verify config content
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Check oauthAccount structure
	oauthAccount, ok := config["oauthAccount"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected oauthAccount to be present")
	}

	if oauthAccount["emailAddress"] != "newuser@example.com" {
		t.Errorf("Expected email 'newuser@example.com', got '%v'", oauthAccount["emailAddress"])
	}
}

func TestBasicConfigManager_UpdateExistingConfig(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing config with additional fields and oauthAccount
	existingConfig := map[string]interface{}{
		"oauthAccount": map[string]interface{}{
			"emailAddress": "old@example.com",
			"accountUuid":  "old-uuid",
		},
		"organization":   "old-org",
		"api_key":       "old-key",
		"other_setting": "should-be-preserved",
	}

	configData, _ := json.MarshalIndent(existingConfig, "", "  ")
	configPath := filepath.Join(tmpDir, ".claude.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	configManager := NewBasicConfigManager(tmpDir)
	ctx := context.Background()

	// Create new account
	account, err := domain.NewAccount(
		"new@example.com",
		"new-alias",
		"new-uuid",
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Update config
	err = configManager.SetCurrentAccount(ctx, account)
	if err != nil {
		t.Fatalf("Failed to set current account: %v", err)
	}

	// Verify other settings are preserved
	configData, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read updated config: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal updated config: %v", err)
	}

	// Check oauthAccount was updated
	oauthAccount, ok := config["oauthAccount"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected oauthAccount to be present")
	}

	if oauthAccount["emailAddress"] != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%v'", oauthAccount["emailAddress"])
	}

	if config["other_setting"] != "should-be-preserved" {
		t.Errorf("Expected other_setting to be preserved, got '%v'", config["other_setting"])
	}
}