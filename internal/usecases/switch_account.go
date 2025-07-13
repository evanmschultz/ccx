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

// SwitchAccountUseCase defines the interface for switching between accounts
type SwitchAccountUseCase interface {
	Execute(ctx context.Context, input SwitchAccountInput) (*SwitchAccountResult, error)
}

// SwitchAccountInput contains the input data for switching accounts
type SwitchAccountInput struct {
	// Exactly one of these should be provided
	AccountID string // Direct ID lookup
	Email     string // Email lookup
	Alias     string // Alias lookup
	Index     int    // Quick-switch by index (1-based for CLI)
	Previous  bool   // Switch to previous account (toggle)
}

// SwitchAccountResult contains the result of a switch operation
type SwitchAccountResult struct {
	From *AccountInfo // Previous account (nil for first switch)
	To   AccountInfo  // New current account
}

// SwitchAccountService implements the SwitchAccountUseCase
type SwitchAccountService struct {
	accounts    ports.AccountRepository
	credentials ports.CredentialStore
	config      ports.ConfigManager
	history     ports.HistoryRepository
}

// Ensure SwitchAccountService implements SwitchAccountUseCase at compile time
var _ SwitchAccountUseCase = (*SwitchAccountService)(nil)

// NewSwitchAccountService creates a new SwitchAccountService
func NewSwitchAccountService(
	accounts ports.AccountRepository,
	credentials ports.CredentialStore,
	config ports.ConfigManager,
	history ports.HistoryRepository,
) SwitchAccountUseCase {
	return &SwitchAccountService{
		accounts:    accounts,
		credentials: credentials,
		config:      config,
		history:     history,
	}
}

// Execute switches the current account based on the provided input
func (s *SwitchAccountService) Execute(ctx context.Context, input SwitchAccountInput) (*SwitchAccountResult, error) {
	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Get current account (may be nil for first switch)
	currentAccount, err := s.config.GetCurrentAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current account: %w", err)
	}

	// Determine target account based on input
	targetAccount, err := s.determineTargetAccount(ctx, input)
	if err != nil {
		return nil, err
	}

	// Check if switching to same account
	if currentAccount != nil && currentAccount.ID() == targetAccount.ID() {
		// This is a no-op, return success
		currentInfo := s.accountToInfo(currentAccount)
		return &SwitchAccountResult{
			From: &currentInfo,
			To:   currentInfo,
		}, nil
	}

	// Verify credentials exist for target account
	_, err = s.credentials.Retrieve(ctx, targetAccount.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials for account %s: %w", targetAccount.Alias(), err)
	}

	// Update config with new account (this is the critical operation)
	err = s.config.SetCurrentAccount(ctx, targetAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to set current account: %w", err)
	}

	// Save switch to history (non-critical - warn on failure)
	if currentAccount != nil {
		err = s.saveToHistory(ctx, currentAccount.Email(), targetAccount.Email())
		if err != nil {
			// Log warning but don't fail the operation
			// In a real implementation, this would use a proper logger
			// For now, we'll just continue
			_ = err // Acknowledge the error was handled
		}
	}

	// Build result
	result := &SwitchAccountResult{
		To: s.accountToInfo(targetAccount),
	}
	if currentAccount != nil {
		fromInfo := s.accountToInfo(currentAccount)
		result.From = &fromInfo
	}

	return result, nil
}

// determineTargetAccount resolves the target account based on input
func (s *SwitchAccountService) determineTargetAccount(ctx context.Context, input SwitchAccountInput) (*domain.Account, error) {
	// Handle Previous flag
	if input.Previous {
		// Ensure no other inputs are provided
		if input.AccountID != "" || input.Email != "" || input.Alias != "" || input.Index > 0 {
			return nil, errors.New("previous flag cannot be combined with other input methods")
		}
		return s.getPreviousAccount(ctx)
	}

	// Count provided inputs
	inputCount := 0
	if input.AccountID != "" {
		inputCount++
	}
	if input.Email != "" {
		inputCount++
	}
	if input.Alias != "" {
		inputCount++
	}
	if input.Index > 0 {
		inputCount++
	}

	// Validate exactly one input
	if inputCount == 0 {
		return nil, errors.New("no account identifier provided")
	}
	if inputCount > 1 {
		return nil, errors.New("multiple account identifiers provided; use only one")
	}

	// Find account by the provided method
	switch {
	case input.AccountID != "":
		return s.accounts.FindByID(ctx, domain.AccountID(input.AccountID))
	case input.Email != "":
		return s.accounts.FindByEmail(ctx, domain.Email(input.Email))
	case input.Alias != "":
		return s.accounts.FindByAlias(ctx, input.Alias)
	case input.Index > 0:
		return s.findByIndex(ctx, input.Index)
	default:
		return nil, errors.New("internal error: invalid input state")
	}
}

// getPreviousAccount retrieves the previous account from history
func (s *SwitchAccountService) getPreviousAccount(ctx context.Context) (*domain.Account, error) {
	history, err := s.history.LoadHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	lastSwitch := history.GetLastSwitch()
	if lastSwitch == nil {
		return nil, errors.New("no previous account in history")
	}

	// The "from" of the last switch is our target
	return s.accounts.FindByEmail(ctx, lastSwitch.From())
}

// findByIndex finds an account by its position in the list (1-based)
func (s *SwitchAccountService) findByIndex(ctx context.Context, index int) (*domain.Account, error) {
	if index <= 0 {
		return nil, fmt.Errorf("invalid index %d: must be positive", index)
	}

	accounts, err := s.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	if index > len(accounts) {
		return nil, fmt.Errorf("index %d out of range (have %d accounts)", index, len(accounts))
	}

	// Convert 1-based to 0-based index
	return accounts[index-1], nil
}

// saveToHistory saves a switch entry to history
func (s *SwitchAccountService) saveToHistory(ctx context.Context, from, to domain.Email) error {
	// Load current history
	history, err := s.history.LoadHistory(ctx)
	if err != nil {
		// If we can't load history, create a new one
		history = domain.NewHistory(50) // Default to 50 entries
	}

	// Create switch entry
	entry, err := domain.NewSwitchEntry(from, to)
	if err != nil {
		return fmt.Errorf("failed to create switch entry: %w", err)
	}

	// Add to history
	history.AddEntry(entry)

	// Save updated history
	return s.history.SaveHistory(ctx, history)
}

// accountToInfo converts a domain Account to AccountInfo DTO
func (s *SwitchAccountService) accountToInfo(account *domain.Account) AccountInfo {
	return AccountInfo{
		ID:        string(account.ID()),
		Email:     string(account.Email()),
		Alias:     account.Alias(),
		UUID:      account.UUID(),
		CreatedAt: account.CreatedAt(),
	}
}
