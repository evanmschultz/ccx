package ports_test

import (
	"context"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// mockConfigManager is a test implementation of ConfigManager
type mockConfigManager struct {
	currentAccount *domain.Account
	err            error
}

func newMockConfigManager() *mockConfigManager {
	return &mockConfigManager{}
}

func (m *mockConfigManager) GetCurrentAccount(_ context.Context) (*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.currentAccount, nil
}

func (m *mockConfigManager) SetCurrentAccount(_ context.Context, account *domain.Account) error {
	if m.err != nil {
		return m.err
	}
	m.currentAccount = account
	return nil
}

// TestConfigManagerInterface validates the ConfigManager interface contract
func TestConfigManagerInterface(t *testing.T) {
	ctx := context.Background()
	manager := newMockConfigManager()

	// Ensure it implements the interface
	var _ ports.ConfigManager = manager

	// Test initial state (no current account)
	current, err := manager.GetCurrentAccount(ctx)
	if err != nil {
		t.Errorf("GetCurrentAccount() error = %v", err)
	}
	if current != nil {
		t.Error("GetCurrentAccount() should return nil initially")
	}

	// Create test account
	account, err := domain.NewAccount("config@example.com", "config", "uuid-config")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Test SetCurrentAccount
	if err := manager.SetCurrentAccount(ctx, account); err != nil {
		t.Errorf("SetCurrentAccount() error = %v", err)
	}

	// Test GetCurrentAccount after setting
	current, err = manager.GetCurrentAccount(ctx)
	if err != nil {
		t.Errorf("GetCurrentAccount() error = %v", err)
	}
	if current == nil {
		t.Error("GetCurrentAccount() should return account after setting")
	}
	if current.Email() != account.Email() {
		t.Error("GetCurrentAccount() returned wrong account")
	}
}

// TestConfigManagerUseCaseScenarios tests common use case scenarios
func TestConfigManagerUseCaseScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, manager ports.ConfigManager)
	}{
		{
			name: "AddAccount reads current Claude config",
			scenario: func(t *testing.T, manager ports.ConfigManager) {
				ctx := context.Background()

				// Simulate existing Claude account
				existing, _ := domain.NewAccount("existing@example.com", "", "uuid-existing")
				_ = manager.SetCurrentAccount(ctx, existing)

				// AddAccount would read current account
				current, err := manager.GetCurrentAccount(ctx)
				if err != nil {
					t.Errorf("failed to get current account: %v", err)
				}
				if current == nil || current.Email() != existing.Email() {
					t.Error("should return existing account from Claude config")
				}
			},
		},
		{
			name: "SwitchAccount updates Claude config",
			scenario: func(t *testing.T, manager ports.ConfigManager) {
				ctx := context.Background()

				// Start with one account
				oldAccount, _ := domain.NewAccount("old@example.com", "old", "uuid-old")
				_ = manager.SetCurrentAccount(ctx, oldAccount)

				// Switch to new account
				newAccount, _ := domain.NewAccount("new@example.com", "new", "uuid-new")
				if err := manager.SetCurrentAccount(ctx, newAccount); err != nil {
					t.Errorf("failed to switch account: %v", err)
				}

				// Verify switch
				current, _ := manager.GetCurrentAccount(ctx)
				if current == nil || current.Email() != newAccount.Email() {
					t.Error("config should be updated to new account")
				}
			},
		},
		{
			name: "RemoveAccount handling current account",
			scenario: func(t *testing.T, manager ports.ConfigManager) {
				ctx := context.Background()

				// Set current account
				account, _ := domain.NewAccount("remove@example.com", "remove", "uuid-remove")
				_ = manager.SetCurrentAccount(ctx, account)

				// After removal, config might be cleared or set to another account
				// This depends on the use case implementation
				// For now, just verify we can read the state
				current, err := manager.GetCurrentAccount(ctx)
				if err != nil {
					t.Errorf("GetCurrentAccount() error = %v", err)
				}
				// The use case will decide what to do when removing current account
				_ = current
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newMockConfigManager()
			tt.scenario(t, manager)
		})
	}
}
