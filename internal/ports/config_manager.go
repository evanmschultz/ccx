package ports

import (
	"context"

	"github.com/evanschultz/ccx/internal/domain"
)

// ConfigManager defines the interface for managing Claude's configuration.
// This handles reading and updating the Claude config file.
type ConfigManager interface {
	// GetCurrentAccount reads the current account from Claude config.
	// Returns nil if no account is configured.
	GetCurrentAccount(ctx context.Context) (*domain.Account, error)

	// SetCurrentAccount updates Claude config with the new account.
	// Used by SwitchAccount and AddAccount use cases.
	SetCurrentAccount(ctx context.Context, account *domain.Account) error
}
