package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/usecases"
)

// Test email constants
const (
	testEmailPersonal = "personal@example.com"
	testEmailWork     = "work@example.com"
	testEmailTest     = "test@example.com"
)

// Test setup helper for SwitchAccountUseCase
type switchAccountTestSetup struct {
	accountRepo     *mockAccountRepository
	credentialStore *mockCredentialStore
	configManager   *mockConfigManager
	historyRepo     *mockHistoryRepository
	useCase         usecases.SwitchAccountUseCase
	testAccounts    map[string]*domain.Account
	testCredentials map[domain.AccountID]*domain.Credentials
}

// Mock history repository
type mockHistoryRepository struct {
	history   *domain.History
	saveErr   error
	loadErr   error
	saveCalls int
}

func newMockHistoryRepository() *mockHistoryRepository {
	return &mockHistoryRepository{
		history: domain.NewHistory(10),
	}
}

func (m *mockHistoryRepository) SaveHistory(_ context.Context, history *domain.History) error {
	m.saveCalls++
	if m.saveErr != nil {
		return m.saveErr
	}
	m.history = history
	return nil
}

func (m *mockHistoryRepository) LoadHistory(_ context.Context) (*domain.History, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.history, nil
}

func setupSwitchAccountTest() *switchAccountTestSetup {
	accountRepo := newMockAccountRepository()
	credentialStore := newMockCredentialStore()
	configManager := newMockConfigManager()
	historyRepo := newMockHistoryRepository()

	// Pre-populate test accounts
	accounts := make(map[string]*domain.Account)
	credentials := make(map[domain.AccountID]*domain.Credentials)

	// Create test accounts
	personal, _ := domain.NewAccount(testEmailPersonal, "personal", "uuid-personal")
	work, _ := domain.NewAccount(testEmailWork, "work", "uuid-work")
	test, _ := domain.NewAccount(testEmailTest, "test", "uuid-test")

	accounts["personal"] = personal
	accounts["work"] = work
	accounts["test"] = test

	// Add to repository
	_ = accountRepo.Save(context.Background(), personal)
	_ = accountRepo.Save(context.Background(), work)
	_ = accountRepo.Save(context.Background(), test)

	// Create credentials for each account
	for _, account := range accounts {
		creds, _ := domain.NewCredentials(account.ID(), []byte(`{"sessionKey": "key-`+account.Alias()+`"}`))
		credentials[account.ID()] = creds
		_ = credentialStore.Store(context.Background(), creds)
	}

	// Set current account to personal by default
	configManager.currentAccount = personal

	useCase := usecases.NewSwitchAccountService(
		accountRepo,
		credentialStore,
		configManager,
		historyRepo,
	)

	return &switchAccountTestSetup{
		accountRepo:     accountRepo,
		credentialStore: credentialStore,
		configManager:   configManager,
		historyRepo:     historyRepo,
		useCase:         useCase,
		testAccounts:    accounts,
		testCredentials: credentials,
	}
}

// TestSwitchAccountUseCase_Execute_ByAlias tests switching by alias
func TestSwitchAccountUseCase_Execute_ByAlias(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Alias: "work",
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

	if result.From.Email != testEmailPersonal {
		t.Errorf("Expected from email %s, got %s", testEmailPersonal, result.From.Email)
	}

	if result.To.Email != testEmailWork {
		t.Errorf("Expected to email %s, got %s", testEmailWork, result.To.Email)
	}

	// Verify config was updated
	if setup.configManager.currentAccount.Email() != testEmailWork {
		t.Error("Config was not updated to work account")
	}

	// Verify history was saved
	if setup.historyRepo.saveCalls != 1 {
		t.Errorf("Expected history to be saved once, got %d calls", setup.historyRepo.saveCalls)
	}
}

// TestSwitchAccountUseCase_Execute_ByEmail tests switching by email
func TestSwitchAccountUseCase_Execute_ByEmail(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Email: testEmailTest,
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result.To.Email != testEmailTest {
		t.Errorf("Expected to email %s, got %s", testEmailTest, result.To.Email)
	}
}

// TestSwitchAccountUseCase_Execute_ByID tests switching by account ID
func TestSwitchAccountUseCase_Execute_ByID(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	workAccount := setup.testAccounts["work"]
	input := usecases.SwitchAccountInput{
		AccountID: string(workAccount.ID()),
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result.To.ID != string(workAccount.ID()) {
		t.Errorf("Expected to ID %s, got %s", workAccount.ID(), result.To.ID)
	}
}

// TestSwitchAccountUseCase_Execute_ByIndex tests switching by index
func TestSwitchAccountUseCase_Execute_ByIndex(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Index: 2, // Second account in list (1-based)
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Result should be one of the accounts (order not guaranteed in map iteration)
	validEmails := map[string]bool{
		testEmailPersonal: true,
		testEmailWork:     true,
		testEmailTest:     true,
	}

	if !validEmails[result.To.Email] {
		t.Errorf("Unexpected account email: %s", result.To.Email)
	}
}

// TestSwitchAccountUseCase_Execute_Previous tests switching to previous account
func TestSwitchAccountUseCase_Execute_Previous(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// First switch from personal to work
	_, _ = setup.useCase.Execute(ctx, usecases.SwitchAccountInput{Alias: "work"})

	// Now switch to previous (should go back to personal)
	input := usecases.SwitchAccountInput{
		Previous: true,
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result.To.Email != testEmailPersonal {
		t.Errorf("Expected to switch back to %s, got %s", testEmailPersonal, result.To.Email)
	}
}

// TestSwitchAccountUseCase_Execute_FirstSwitch tests when no current account
func TestSwitchAccountUseCase_Execute_FirstSwitch(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// Clear current account
	setup.configManager.currentAccount = nil

	input := usecases.SwitchAccountInput{
		Alias: "work",
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// From should be nil for first switch
	if result.From != nil {
		t.Errorf("Expected From to be nil for first switch, got %+v", result.From)
	}

	if result.To.Email != testEmailWork {
		t.Errorf("Expected to email %s, got %s", testEmailWork, result.To.Email)
	}
}

// TestSwitchAccountUseCase_Execute_SameAccount tests switching to same account
func TestSwitchAccountUseCase_Execute_SameAccount(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Alias: "personal", // Already the current account
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify - should succeed but not actually switch
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if result.From.Email != result.To.Email {
		t.Error("Expected From and To to be the same for same-account switch")
	}

	// History should not be updated for same-account switch
	if setup.historyRepo.saveCalls != 0 {
		t.Error("History should not be saved for same-account switch")
	}
}

// TestSwitchAccountUseCase_Execute_NoInputMethod tests when no input method provided
func TestSwitchAccountUseCase_Execute_NoInputMethod(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{} // Empty input

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when no input method provided, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestSwitchAccountUseCase_Execute_MultipleInputMethods tests when multiple methods provided
func TestSwitchAccountUseCase_Execute_MultipleInputMethods(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Alias: "work",
		Email: testEmailTest, // Both alias and email
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when multiple input methods provided, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestSwitchAccountUseCase_Execute_AccountNotFound tests when account doesn't exist
func TestSwitchAccountUseCase_Execute_AccountNotFound(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Alias: "nonexistent",
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when account not found, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestSwitchAccountUseCase_Execute_CredentialsNotFound tests when credentials missing
func TestSwitchAccountUseCase_Execute_CredentialsNotFound(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// Remove credentials for work account
	workAccount := setup.testAccounts["work"]
	_ = setup.credentialStore.Delete(ctx, workAccount.ID())

	input := usecases.SwitchAccountInput{
		Alias: "work",
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when credentials not found, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	// Config should not be updated
	if setup.configManager.currentAccount.Email() != testEmailPersonal {
		t.Error("Config should not be updated when credentials are missing")
	}
}

// TestSwitchAccountUseCase_Execute_ConfigUpdateFailure tests config update failure
func TestSwitchAccountUseCase_Execute_ConfigUpdateFailure(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// Force config update to fail
	setup.configManager.setErr = errors.New("config file locked")

	input := usecases.SwitchAccountInput{
		Alias: "work",
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

	// History should not be saved if config update failed
	if setup.historyRepo.saveCalls != 0 {
		t.Error("History should not be saved when config update fails")
	}
}

// TestSwitchAccountUseCase_Execute_HistoryFailure tests history save failure
func TestSwitchAccountUseCase_Execute_HistoryFailure(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// Force history save to fail
	setup.historyRepo.saveErr = errors.New("history file corrupted")

	input := usecases.SwitchAccountInput{
		Alias: "work",
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)
	// Verify - should succeed despite history failure
	if err != nil {
		t.Errorf("Expected success despite history failure, got error: %v", err)
	}

	if result.To.Email != testEmailWork {
		t.Error("Switch should succeed even if history save fails")
	}

	// Config should be updated
	if setup.configManager.currentAccount.Email() != testEmailWork {
		t.Error("Config should be updated even if history save fails")
	}
}

// TestSwitchAccountUseCase_Execute_ContextCancellation tests context cancellation
func TestSwitchAccountUseCase_Execute_ContextCancellation(t *testing.T) {
	setup := setupSwitchAccountTest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := usecases.SwitchAccountInput{
		Alias: "work",
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

// TestSwitchAccountUseCase_Execute_PreviousWithNoHistory tests previous when no history
func TestSwitchAccountUseCase_Execute_PreviousWithNoHistory(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	// Clear history
	setup.historyRepo.history.Clear()

	input := usecases.SwitchAccountInput{
		Previous: true,
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when no previous account in history, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestSwitchAccountUseCase_Execute_PreviousWithOtherInput tests previous flag with other input
func TestSwitchAccountUseCase_Execute_PreviousWithOtherInput(t *testing.T) {
	setup := setupSwitchAccountTest()
	ctx := context.Background()

	input := usecases.SwitchAccountInput{
		Previous: true,
		Alias:    "work", // Both previous and alias
	}

	// Execute
	result, err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when previous flag used with other input, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}
}

// TestSwitchAccountUseCase_Execute_InvalidIndex tests invalid index values
func TestSwitchAccountUseCase_Execute_InvalidIndex(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{"zero index", 0},
		{"negative index", -1},
		{"out of bounds", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := setupSwitchAccountTest()
			ctx := context.Background()

			input := usecases.SwitchAccountInput{
				Index: tt.index,
			}

			// Execute
			result, err := setup.useCase.Execute(ctx, input)

			// Verify
			if err == nil {
				t.Error("Expected error for invalid index, got nil")
			}

			if result != nil {
				t.Errorf("Expected nil result on error, got %+v", result)
			}
		})
	}
}
