// Package usecases contains the application business rules and orchestrates the flow
// between the domain entities and the infrastructure adapters.
package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// RemoveAccountUseCase defines the interface for removing an account from ccx
type RemoveAccountUseCase interface {
	Execute(ctx context.Context, input RemoveAccountInput) (*RemoveAccountResult, error)
}

// RemoveAccountInput contains the input data for removing an account
type RemoveAccountInput struct {
	AccountID string // Account ID to remove
}

// RemoveAccountResult contains the result of a remove operation
type RemoveAccountResult struct {
	RemovedAccount    AccountInfo // Information about the removed account
	WasCurrentAccount bool        // True if the removed account was the current account
	WasLastAccount    bool        // True if this was the last account in the system
}

// RemoveAccountService implements the RemoveAccountUseCase
type RemoveAccountService struct {
	accounts    ports.AccountRepository
	credentials ports.CredentialStore
	config      ports.ConfigManager
	history     ports.HistoryRepository
}

// Ensure RemoveAccountService implements RemoveAccountUseCase at compile time
var _ RemoveAccountUseCase = (*RemoveAccountService)(nil)

// NewRemoveAccountService creates a new RemoveAccountService
func NewRemoveAccountService(
	accounts ports.AccountRepository,
	credentials ports.CredentialStore,
	config ports.ConfigManager,
	history ports.HistoryRepository,
) RemoveAccountUseCase {
	return &RemoveAccountService{
		accounts:    accounts,
		credentials: credentials,
		config:      config,
		history:     history,
	}
}

// Execute removes an account from ccx, including its credentials and configuration
func (s *RemoveAccountService) Execute(ctx context.Context, input RemoveAccountInput) (*RemoveAccountResult, error) {
	if err := s.validateInput(ctx, input); err != nil {
		return nil, err
	}

	accountID := domain.AccountID(input.AccountID)
	account, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	metadata, err := s.getRemovalMetadata(ctx, account)
	if err != nil {
		return nil, err
	}

	if err := s.performRemoval(ctx, account, metadata); err != nil {
		return nil, err
	}

	return &RemoveAccountResult{
		RemovedAccount:    metadata.accountInfo,
		WasCurrentAccount: metadata.isCurrentAccount,
		WasLastAccount:    metadata.isLastAccount,
	}, nil
}

type removalMetadata struct {
	accountInfo       AccountInfo
	isCurrentAccount  bool
	isLastAccount     bool
	backupCredentials *domain.Credentials
}

func (s *RemoveAccountService) validateInput(ctx context.Context, input RemoveAccountInput) error {
	if input.AccountID == "" {
		return errors.New("account ID is required")
	}
	return ctx.Err()
}

func (s *RemoveAccountService) getRemovalMetadata(ctx context.Context, account *domain.Account) (*removalMetadata, error) {
	// Get current account to check if we're removing it
	currentAccount, _ := s.config.GetCurrentAccount(ctx)
	isCurrentAccount := currentAccount != nil && currentAccount.ID() == account.ID()

	// Check if this is the last account
	allAccounts, err := s.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check account count: %w", err)
	}
	isLastAccount := len(allAccounts) == 1

	// Store account info for result before deletion
	accountInfo := AccountInfo{
		ID:        string(account.ID()),
		Email:     string(account.Email()),
		Alias:     account.Alias(),
		UUID:      account.UUID(),
		CreatedAt: account.CreatedAt(),
	}

	// Backup credentials before deletion (for rollback)
	var backupCredentials *domain.Credentials
	if creds, err := s.credentials.Retrieve(ctx, account.ID()); err == nil {
		backupCredentials = creds
	}

	return &removalMetadata{
		accountInfo:       accountInfo,
		isCurrentAccount:  isCurrentAccount,
		isLastAccount:     isLastAccount,
		backupCredentials: backupCredentials,
	}, nil
}

func (s *RemoveAccountService) performRemoval(ctx context.Context, account *domain.Account, metadata *removalMetadata) error {
	// Delete credentials first (critical for security)
	if err := s.credentials.Delete(ctx, account.ID()); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	// Delete account from repository
	if err := s.accounts.Delete(ctx, account.ID()); err != nil {
		s.rollbackCredentials(ctx, metadata.backupCredentials)
		return fmt.Errorf("failed to delete account: %w", err)
	}

	// Clear current account if we're removing it
	if metadata.isCurrentAccount {
		if err := s.config.SetCurrentAccount(ctx, nil); err != nil {
			s.rollbackFull(ctx, account, metadata.backupCredentials)
			return fmt.Errorf("failed to clear current account configuration: %w", err)
		}
		s.updateHistory(ctx)
	}

	return nil
}

func (s *RemoveAccountService) rollbackCredentials(ctx context.Context, backupCredentials *domain.Credentials) {
	if backupCredentials != nil {
		_ = s.credentials.Store(ctx, backupCredentials) // Best effort restore
	}
}

func (s *RemoveAccountService) rollbackFull(ctx context.Context, account *domain.Account, backupCredentials *domain.Credentials) {
	_ = s.accounts.Save(ctx, account) // Best effort restore
	s.rollbackCredentials(ctx, backupCredentials)
}

func (s *RemoveAccountService) updateHistory(ctx context.Context) {
	if currentHistory, err := s.history.LoadHistory(ctx); err == nil {
		_ = s.history.SaveHistory(ctx, currentHistory) // Best effort, don't fail operation
	}
}
