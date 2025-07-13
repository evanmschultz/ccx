// Package usecases contains the application business rules and orchestrates the flow
// between the domain entities and the infrastructure adapters.
package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/evanschultz/ccx/internal/ports"
)

// ListAccountsUseCase defines the interface for listing all accounts in ccx
type ListAccountsUseCase interface {
	Execute(ctx context.Context) ([]AccountInfo, error)
}

// AccountInfo represents account information returned to the presentation layer
type AccountInfo struct {
	ID        string    // Account ID as string for presentation
	Email     string    // Account email
	Alias     string    // Account alias
	UUID      string    // Claude UUID
	CreatedAt time.Time // When the account was added to ccx
}

// ListAccountsService implements the ListAccountsUseCase
type ListAccountsService struct {
	accounts ports.AccountRepository
}

// Ensure ListAccountsService implements ListAccountsUseCase at compile time
var _ ListAccountsUseCase = (*ListAccountsService)(nil)

// NewListAccountsService creates a new ListAccountsService
func NewListAccountsService(accounts ports.AccountRepository) ListAccountsUseCase {
	return &ListAccountsService{
		accounts: accounts,
	}
}

// Execute lists all accounts in ccx
func (s *ListAccountsService) Execute(ctx context.Context) ([]AccountInfo, error) {
	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Retrieve all accounts from repository
	accounts, err := s.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	// Convert domain entities to DTOs
	result := make([]AccountInfo, len(accounts))
	for i, account := range accounts {
		result[i] = AccountInfo{
			ID:        string(account.ID()),
			Email:     string(account.Email()),
			Alias:     account.Alias(),
			UUID:      account.UUID(),
			CreatedAt: account.CreatedAt(),
		}
	}

	return result, nil
}
