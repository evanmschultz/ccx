package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/usecases"
)

// Extended mock for RemoveAccount tests
type extendedMockAccountRepository struct {
	*mockAccountRepository
	deleteErr error
}

func newExtendedMockAccountRepository() *extendedMockAccountRepository {
	return &extendedMockAccountRepository{
		mockAccountRepository: newMockAccountRepository(),
	}
}

func (m *extendedMockAccountRepository) Delete(_ context.Context, id domain.AccountID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.accounts, id)
	return nil
}

// Extended mock for credential store with delete error
type extendedMockCredentialStore struct {
	*mockCredentialStore
	deleteErr error
}

func newExtendedMockCredentialStore() *extendedMockCredentialStore {
	return &extendedMockCredentialStore{
		mockCredentialStore: newMockCredentialStore(),
	}
}

func (m *extendedMockCredentialStore) Delete(_ context.Context, accountID domain.AccountID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.credentials, accountID)
	return nil
}

// Test setup helper for RemoveAccountUseCase
type removeAccountTestSetup struct {
	accountRepo     *extendedMockAccountRepository
	credentialStore *extendedMockCredentialStore
	configManager   *mockConfigManager
	historyRepo     *mockHistoryRepository
	useCase         usecases.RemoveAccountUseCase
	testAccounts    map[string]*domain.Account
}

func setupRemoveAccountTest() *removeAccountTestSetup {
	accountRepo := newExtendedMockAccountRepository()
	credentialStore := newExtendedMockCredentialStore()
	configManager := newMockConfigManager()
	historyRepo := newMockHistoryRepository()

	// Pre-populate with test accounts
	personal, _ := domain.NewAccount(testEmailPersonal, "personal", "uuid-personal")
	work, _ := domain.NewAccount(testEmailWork, "work", "uuid-work")
	test, _ := domain.NewAccount(testEmailTest, "test", "uuid-test")

	testAccounts := map[string]*domain.Account{
		"personal": personal,
		"work":     work,
		"test":     test,
	}

	_ = accountRepo.Save(context.Background(), personal)
	_ = accountRepo.Save(context.Background(), work)
	_ = accountRepo.Save(context.Background(), test)

	// Add credentials for each account
	for _, account := range []*domain.Account{personal, work, test} {
		creds, _ := domain.NewCredentials(account.ID(), []byte(`{"sessionKey": "key-`+account.Alias()+`"}`))
		_ = credentialStore.Store(context.Background(), creds)
	}

	// Set current account to personal by default
	configManager.currentAccount = personal

	useCase := usecases.NewRemoveAccountService(
		accountRepo,
		credentialStore,
		configManager,
		historyRepo,
	)

	return &removeAccountTestSetup{
		accountRepo:     accountRepo,
		credentialStore: credentialStore,
		configManager:   configManager,
		historyRepo:     historyRepo,
		useCase:         useCase,
		testAccounts:    testAccounts,
	}
}

// TestRemoveAccountUseCase_Execute_RemoveExistingAccount tests removing an existing account
func TestRemoveAccountUseCase_Execute_RemoveExistingAccount(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	input := usecases.RemoveAccountInput{
		AccountID: string(setup.testAccounts["work"].ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.RemovedAccount.Email != testEmailWork {
		t.Errorf("Expected removed account email %s, got %s", testEmailWork, result.RemovedAccount.Email)
	}

	// Verify account was deleted from repository
	workAccount := setup.testAccounts["work"]
	_, err = setup.accountRepo.FindByID(ctx, workAccount.ID())
	if err == nil {
		t.Error("Expected account to be deleted from repository")
	}

	// Verify credentials were deleted
	_, err = setup.credentialStore.Retrieve(ctx, workAccount.ID())
	if err == nil {
		t.Error("Expected credentials to be deleted")
	}
}

// TestRemoveAccountUseCase_Execute_RemoveCurrentAccount tests removing the currently active account
func TestRemoveAccountUseCase_Execute_RemoveCurrentAccount(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Remove the current account (personal)
	currentAccount := setup.configManager.currentAccount
	input := usecases.RemoveAccountInput{
		AccountID: string(currentAccount.ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result.RemovedAccount.Email != testEmailPersonal {
		t.Errorf("Expected removed account email %s, got %s", testEmailPersonal, result.RemovedAccount.Email)
	}

	// Verify current account was cleared
	if setup.configManager.currentAccount != nil {
		t.Error("Expected current account to be cleared when removing current account")
	}

	// Verify history was updated if needed
	if setup.historyRepo.saveCalls == 0 {
		t.Error("Expected history to be saved when removing current account")
	}
}

// TestRemoveAccountUseCase_Execute_RemoveNonExistentAccount tests removing an account that doesn't exist
func TestRemoveAccountUseCase_Execute_RemoveNonExistentAccount(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	input := usecases.RemoveAccountInput{
		AccountID: "nonexistent-id",
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when removing non-existent account, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestRemoveAccountUseCase_Execute_CredentialDeletionFailure tests when credential deletion fails
func TestRemoveAccountUseCase_Execute_CredentialDeletionFailure(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Force credential deletion to fail
	setup.credentialStore.deleteErr = errors.New("keychain locked")

	input := usecases.RemoveAccountInput{
		AccountID: string(setup.testAccounts["work"].ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when credential deletion fails, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	// Verify account was NOT deleted (rollback)
	workAccount := setup.testAccounts["work"]
	_, err = setup.accountRepo.FindByID(ctx, workAccount.ID())
	if err != nil {
		t.Error("Account should not be deleted when credential deletion fails")
	}
}

// TestRemoveAccountUseCase_Execute_AccountDeletionFailure tests when account deletion fails
func TestRemoveAccountUseCase_Execute_AccountDeletionFailure(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Force account deletion to fail
	setup.accountRepo.deleteErr = errors.New("database locked")

	input := usecases.RemoveAccountInput{
		AccountID: string(setup.testAccounts["work"].ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when account deletion fails, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	// Verify credentials should be restored (rollback logic)
	workAccount := setup.testAccounts["work"]
	_, err = setup.credentialStore.Retrieve(ctx, workAccount.ID())
	if err != nil {
		t.Error("Credentials should be restored when account deletion fails")
	}
}

// TestRemoveAccountUseCase_Execute_RemoveLastAccount tests removing the last remaining account
func TestRemoveAccountUseCase_Execute_RemoveLastAccount(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Remove all accounts except one
	_ = setup.accountRepo.Delete(ctx, setup.testAccounts["work"].ID())
	_ = setup.accountRepo.Delete(ctx, setup.testAccounts["test"].ID())

	// Now try to remove the last account
	input := usecases.RemoveAccountInput{
		AccountID: string(setup.testAccounts["personal"].ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify - should succeed but warn or handle specially
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.WasLastAccount {
		t.Error("Expected WasLastAccount to be true when removing last account")
	}

	// Current account should be cleared
	if setup.configManager.currentAccount != nil {
		t.Error("Expected current account to be cleared when removing last account")
	}
}

// TestRemoveAccountUseCase_Execute_EmptyInput tests when no account ID provided
func TestRemoveAccountUseCase_Execute_EmptyInput(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	input := usecases.RemoveAccountInput{} // Empty input

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when no account ID provided, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestRemoveAccountUseCase_Execute_ContextCancellation tests context cancellation
func TestRemoveAccountUseCase_Execute_ContextCancellation(t *testing.T) {
	setup := setupRemoveAccountTest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := usecases.RemoveAccountInput{
		AccountID: string(setup.testAccounts["work"].ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if result != nil {
		t.Errorf("Execute() result = %v, want nil", result)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want %v", err, context.Canceled)
	}
}

// TestRemoveAccountUseCase_Execute_ConfigUpdateFailure tests when clearing current account fails
func TestRemoveAccountUseCase_Execute_ConfigUpdateFailure(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Force config update to fail
	setup.configManager.setErr = errors.New("config file locked")

	// Remove the current account
	currentAccount := setup.configManager.currentAccount
	input := usecases.RemoveAccountInput{
		AccountID: string(currentAccount.ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when config update fails, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	// Account and credentials should be restored (rollback)
	_, err = setup.accountRepo.FindByID(ctx, currentAccount.ID())
	if err != nil {
		t.Error("Account should be restored when config update fails")
	}
}

// TestRemoveAccountUseCase_Execute_HistoryUpdateFailure tests when history update fails
func TestRemoveAccountUseCase_Execute_HistoryUpdateFailure(t *testing.T) {
	setup := setupRemoveAccountTest()
	ctx := context.Background()

	// Force history save to fail
	setup.historyRepo.saveErr = errors.New("history file corrupted")

	// Remove current account
	currentAccount := setup.configManager.currentAccount
	input := usecases.RemoveAccountInput{
		AccountID: string(currentAccount.ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify - should succeed despite history failure
	if err != nil {
		t.Errorf("Expected success despite history failure, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Account should be removed and config cleared
	if setup.configManager.currentAccount != nil {
		t.Error("Current account should be cleared even if history fails")
	}
}

// Interface compliance test
func TestRemoveAccountService_ImplementsInterface(_ *testing.T) {
	var _ usecases.RemoveAccountUseCase = (*usecases.RemoveAccountService)(nil)
}
