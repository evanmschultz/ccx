// Package usecases contains the application business rules and orchestrates the flow
// between the domain entities and the infrastructure adapters.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// AddAccountUseCase defines the interface for adding a new account to ccx
type AddAccountUseCase interface {
	Execute(ctx context.Context, input AddAccountInput) error
}

// AddAccountInput contains the input data for adding an account
type AddAccountInput struct {
	// If Email is provided, use it; otherwise read from Claude config
	Email string
	// If Alias is provided, use it; otherwise generate from email
	Alias string
	// If Credentials is provided, use it; otherwise read from Claude config
	Credentials []byte
}

// AddAccountService implements the AddAccountUseCase
type AddAccountService struct {
	accounts    ports.AccountRepository
	credentials ports.CredentialStore
	config      ports.ConfigManager
}

// NewAddAccountService creates a new AddAccountService
func NewAddAccountService(
	accounts ports.AccountRepository,
	credentials ports.CredentialStore,
	config ports.ConfigManager,
) AddAccountUseCase {
	return &AddAccountService{
		accounts:    accounts,
		credentials: credentials,
		config:      config,
	}
}

// Execute adds a new account to ccx
func (s *AddAccountService) Execute(ctx context.Context, input AddAccountInput) error {
	// Step 1: Determine account details
	email, uuid, credentialData, err := s.determineAccountDetails(ctx, input)
	if err != nil {
		return err
	}

	// Step 2: Check if account already exists
	if err := s.checkAccountExists(ctx, email); err != nil {
		return err
	}

	// Step 3: Generate alias if not provided
	alias := s.generateAlias(input.Alias, email)

	// Step 4: Create and save account with credentials
	return s.createAndSaveAccount(ctx, email, alias, uuid, credentialData)
}

// determineAccountDetails resolves email, uuid, and credentials from input or Claude config
func (s *AddAccountService) determineAccountDetails(ctx context.Context, input AddAccountInput) (string, string, []byte, error) {
	if input.Email != "" {
		// Use explicit input
		if len(input.Credentials) == 0 {
			return "", "", nil, errors.New("credentials must be provided when email is specified explicitly")
		}
		uuid := "explicit-" + strings.ReplaceAll(input.Email, "@", "-")
		return input.Email, uuid, input.Credentials, nil
	}

	// Get from Claude config
	currentAccount, err := s.config.GetCurrentAccount(ctx)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get current Claude account: %w", err)
	}
	if currentAccount == nil {
		return "", "", nil, errors.New("no current Claude account found - please configure Claude or provide explicit email")
	}

	email := string(currentAccount.Email())
	uuid := currentAccount.UUID()

	// Use provided credentials or generate placeholder
	var credentialData []byte
	if len(input.Credentials) > 0 {
		credentialData = input.Credentials
	} else {
		credentialData = []byte(fmt.Sprintf(`{"account_id": "%s", "session_key": "placeholder"}`, uuid))
	}

	return email, uuid, credentialData, nil
}

// checkAccountExists verifies the account doesn't already exist
func (s *AddAccountService) checkAccountExists(ctx context.Context, email string) error {
	_, err := s.accounts.FindByEmail(ctx, domain.Email(email))
	if err == nil {
		return fmt.Errorf("account with email %s already exists", email)
	}
	return nil
}

// generateAlias creates an alias from email if not provided
func (s *AddAccountService) generateAlias(inputAlias, email string) string {
	if inputAlias != "" {
		return inputAlias
	}
	// Generate alias from email (part before @)
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// createAndSaveAccount creates the account and credentials, saving them with cleanup on failure
func (s *AddAccountService) createAndSaveAccount(ctx context.Context, email, alias, uuid string, credentialData []byte) error {
	// Create account entity
	account, err := domain.NewAccount(email, alias, uuid)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Create and store credentials
	credentials, err := domain.NewCredentials(account.ID(), credentialData)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	err = s.credentials.Store(ctx, credentials)
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	// Save account (after credentials to ensure atomic-like behavior)
	err = s.accounts.Save(ctx, account)
	if err != nil {
		// Clean up credentials on account save failure
		_ = s.credentials.Delete(ctx, account.ID())
		return fmt.Errorf("failed to save account: %w", err)
	}

	return nil
}
