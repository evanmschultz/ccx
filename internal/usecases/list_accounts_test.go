package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/usecases"
)

// Test setup helper for ListAccountsUseCase
type listAccountsTestSetup struct {
	accountRepo *mockAccountRepository
	useCase     usecases.ListAccountsUseCase
}

func setupListAccountsTest() *listAccountsTestSetup {
	accountRepo := newMockAccountRepository()
	useCase := usecases.NewListAccountsService(accountRepo)

	return &listAccountsTestSetup{
		accountRepo: accountRepo,
		useCase:     useCase,
	}
}

// TestListAccountsUseCase_Execute_EmptyList tests when no accounts exist
func TestListAccountsUseCase_Execute_EmptyList(t *testing.T) {
	setup := setupListAccountsTest()
	ctx := context.Background()

	// Execute
	accounts, err := setup.useCase.Execute(ctx)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if len(accounts) != 0 {
		t.Errorf("Expected empty list, got %d accounts", len(accounts))
	}
}

// TestListAccountsUseCase_Execute_MultipleAccounts tests listing multiple accounts
func TestListAccountsUseCase_Execute_MultipleAccounts(t *testing.T) {
	setup := setupListAccountsTest()
	ctx := context.Background()

	// Setup: Add multiple accounts
	account1, _ := domain.NewAccount("user1@example.com", "personal", "uuid-1")
	account2, _ := domain.NewAccount("user2@example.com", "work", "uuid-2")
	account3, _ := domain.NewAccount("user3@example.com", "test", "uuid-3")

	_ = setup.accountRepo.Save(ctx, account1)
	_ = setup.accountRepo.Save(ctx, account2)
	_ = setup.accountRepo.Save(ctx, account3)

	// Execute
	accounts, err := setup.useCase.Execute(ctx)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if len(accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(accounts))
	}

	// Verify account details are preserved
	emailSet := make(map[string]bool)
	aliasSet := make(map[string]bool)
	idSet := make(map[string]bool)

	for _, account := range accounts {
		emailSet[account.Email] = true
		aliasSet[account.Alias] = true
		idSet[account.ID] = true
	}

	expectedEmails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range expectedEmails {
		if !emailSet[email] {
			t.Errorf("Expected to find email %s in results", email)
		}
	}

	expectedAliases := []string{"personal", "work", "test"}
	for _, alias := range expectedAliases {
		if !aliasSet[alias] {
			t.Errorf("Expected to find alias %s in results", alias)
		}
	}
}

// TestListAccountsUseCase_Execute_RepositoryError tests repository failure handling
func TestListAccountsUseCase_Execute_RepositoryError(t *testing.T) {
	setup := setupListAccountsTest()
	ctx := context.Background()

	// Setup: Force repository to return error
	repoErr := errors.New("database connection failed")
	setup.accountRepo.findErr = repoErr

	// Execute
	accounts, err := setup.useCase.Execute(ctx)

	// Verify
	if err == nil {
		t.Error("Expected error when repository fails, got nil")
	}

	if accounts != nil {
		t.Errorf("Expected nil accounts on error, got %v", accounts)
	}

	// Check that the original repository error is wrapped
	if !errors.Is(err, repoErr) {
		t.Errorf("Expected error to wrap '%v', but it did not. Got: %v", repoErr, err)
	}
}

// TestListAccountsUseCase_Execute_AccountInfo tests that AccountInfo contains all necessary fields
func TestListAccountsUseCase_Execute_AccountInfo(t *testing.T) {
	setup := setupListAccountsTest()
	ctx := context.Background()

	// Setup: Add a test account
	account, _ := domain.NewAccount("test@example.com", "myalias", "uuid-test")
	_ = setup.accountRepo.Save(ctx, account)

	// Execute
	accounts, err := setup.useCase.Execute(ctx)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}

	info := accounts[0]
	if info.ID == "" {
		t.Error("AccountInfo.ID should not be empty")
	}
	if info.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", info.Email)
	}
	if info.Alias != "myalias" {
		t.Errorf("Expected alias myalias, got %s", info.Alias)
	}
	if info.UUID != "uuid-test" {
		t.Errorf("Expected UUID uuid-test, got %s", info.UUID)
	}
	if info.CreatedAt.IsZero() {
		t.Error("AccountInfo.CreatedAt should not be zero")
	}
}

// TestListAccountsUseCase_Execute_ContextCancellation tests context cancellation handling
func TestListAccountsUseCase_Execute_ContextCancellation(t *testing.T) {
	setup := setupListAccountsTest()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Execute with cancelled context
	accounts, err := setup.useCase.Execute(ctx)

	// Verify
	if accounts != nil {
		t.Errorf("Execute() accounts = %v, want nil", accounts)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want %v", err, context.Canceled)
	}
}
