// Package ports defines the interfaces (ports) that the use cases depend on.
// These interfaces follow the dependency inversion principle - the core domain
// defines what it needs, and the infrastructure layer provides implementations.
package ports

import (
	"context"

	"github.com/evanschultz/ccx/internal/domain"
)

// AccountRepository defines the interface for account persistence.
// Each method is designed to support specific use case needs.
type AccountRepository interface {
	// Save persists an account. Used by AddAccount use case.
	Save(ctx context.Context, account *domain.Account) error

	// FindByID retrieves an account by its ID. Used by multiple use cases.
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)

	// FindByEmail retrieves an account by email. Used by SwitchAccount use case.
	FindByEmail(ctx context.Context, email domain.Email) (*domain.Account, error)

	// FindByAlias retrieves an account by alias. Used by quick-switch feature.
	FindByAlias(ctx context.Context, alias string) (*domain.Account, error)

	// List returns all accounts. Used by ListAccounts use case.
	List(ctx context.Context) ([]*domain.Account, error)

	// Delete removes an account. Used by RemoveAccount use case.
	Delete(ctx context.Context, id domain.AccountID) error
}
