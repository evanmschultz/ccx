package ports

import (
	"context"

	"github.com/evanschultz/ccx/internal/domain"
)

// CredentialStore defines the interface for secure credential storage.
// This will be implemented using system keychains (Keychain on macOS, etc).
type CredentialStore interface {
	// Store securely saves credentials. Used by AddAccount use case.
	Store(ctx context.Context, creds *domain.Credentials) error

	// Retrieve gets credentials for an account. Used by SwitchAccount use case.
	Retrieve(ctx context.Context, accountID domain.AccountID) (*domain.Credentials, error)

	// Delete removes credentials. Used by RemoveAccount use case.
	Delete(ctx context.Context, accountID domain.AccountID) error
}
