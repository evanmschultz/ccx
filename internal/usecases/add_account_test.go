package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/usecases"
)

// Mock implementations for testing
type mockAccountRepository struct {
	accounts map[domain.AccountID]*domain.Account
	saveErr  error
	findErr  error
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts: make(map[domain.AccountID]*domain.Account),
	}
}

func (m *mockAccountRepository) Save(_ context.Context, account *domain.Account) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.accounts[account.ID()] = account
	return nil
}

func (m *mockAccountRepository) FindByID(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	account, ok := m.accounts[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	return account, nil
}

func (m *mockAccountRepository) FindByEmail(_ context.Context, email domain.Email) (*domain.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, account := range m.accounts {
		if account.Email() == email {
			return account, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepository) FindByAlias(_ context.Context, alias string) (*domain.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, account := range m.accounts {
		if account.Alias() == alias {
			return account, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepository) List(_ context.Context) ([]*domain.Account, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	result := make([]*domain.Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		result = append(result, account)
	}
	return result, nil
}

func (m *mockAccountRepository) Delete(_ context.Context, id domain.AccountID) error {
	delete(m.accounts, id)
	return nil
}

type mockCredentialStore struct {
	credentials map[domain.AccountID]*domain.Credentials
	storeErr    error
	retrieveErr error
}

func newMockCredentialStore() *mockCredentialStore {
	return &mockCredentialStore{
		credentials: make(map[domain.AccountID]*domain.Credentials),
	}
}

func (m *mockCredentialStore) Store(_ context.Context, creds *domain.Credentials) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.credentials[creds.AccountID()] = creds
	return nil
}

func (m *mockCredentialStore) Retrieve(_ context.Context, accountID domain.AccountID) (*domain.Credentials, error) {
	if m.retrieveErr != nil {
		return nil, m.retrieveErr
	}
	creds, ok := m.credentials[accountID]
	if !ok {
		return nil, errors.New("credentials not found")
	}
	return creds, nil
}

func (m *mockCredentialStore) Delete(_ context.Context, accountID domain.AccountID) error {
	delete(m.credentials, accountID)
	return nil
}

type mockConfigManager struct {
	currentAccount *domain.Account
	getErr         error
	setErr         error
}

func newMockConfigManager() *mockConfigManager {
	return &mockConfigManager{}
}

func (m *mockConfigManager) GetCurrentAccount(_ context.Context) (*domain.Account, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.currentAccount, nil
}

func (m *mockConfigManager) SetCurrentAccount(_ context.Context, account *domain.Account) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.currentAccount = account
	return nil
}

// Test setup helper
type testSetup struct {
	accountRepo     *mockAccountRepository
	credentialStore *mockCredentialStore
	configManager   *mockConfigManager
	useCase         usecases.AddAccountUseCase
}

func setupTest() *testSetup {
	accountRepo := newMockAccountRepository()
	credentialStore := newMockCredentialStore()
	configManager := newMockConfigManager()

	useCase := usecases.NewAddAccountService(accountRepo, credentialStore, configManager)

	return &testSetup{
		accountRepo:     accountRepo,
		credentialStore: credentialStore,
		configManager:   configManager,
		useCase:         useCase,
	}
}

// Interface compliance test
func TestAddAccountService_ImplementsInterface(_ *testing.T) {
	setup := setupTest()
	_ = setup.useCase
}

// TestAddAccountUseCase_Execute_SuccessfulAdd tests the happy path
func TestAddAccountUseCase_Execute_SuccessfulAdd(t *testing.T) {
	setup := setupTest()
	ctx := context.Background()

	// Setup: Configure a current account in Claude config
	claudeAccount, err := domain.NewAccount("test@example.com", "", "uuid-123")
	if err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}
	setup.configManager.currentAccount = claudeAccount

	// Test input with explicit alias
	input := usecases.AddAccountInput{
		Alias: "work",
	}

	// Execute
	err = setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Verify account was saved
	savedAccounts, _ := setup.accountRepo.List(ctx)
	if len(savedAccounts) != 1 {
		t.Errorf("Expected 1 account to be saved, got %d", len(savedAccounts))
	}

	if len(savedAccounts) > 0 {
		account := savedAccounts[0]
		if account.Email() != "test@example.com" {
			t.Errorf("Expected email test@example.com, got %s", account.Email())
		}
		if account.Alias() != "work" {
			t.Errorf("Expected alias work, got %s", account.Alias())
		}

		// Verify credentials were stored
		_, err = setup.credentialStore.Retrieve(ctx, account.ID())
		if err != nil {
			t.Errorf("Expected credentials to be stored for account %s", account.ID())
		}
	}
}

// TestAddAccountUseCase_Execute_NoClaudeConfig tests when Claude config is missing
func TestAddAccountUseCase_Execute_NoClaudeConfig(t *testing.T) {
	setup := setupTest()
	ctx := context.Background()

	// Setup: No current account configured
	setup.configManager.currentAccount = nil

	input := usecases.AddAccountInput{}

	// Execute
	err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when no Claude config exists, got nil")
	}
}

// TestAddAccountUseCase_Execute_DuplicateAccount tests adding an existing account
func TestAddAccountUseCase_Execute_DuplicateAccount(t *testing.T) {
	setup := setupTest()
	ctx := context.Background()

	// Setup: Add an existing account
	existingAccount, _ := domain.NewAccount("test@example.com", "existing", "uuid-existing")
	_ = setup.accountRepo.Save(ctx, existingAccount)

	// Setup: Configure same account in Claude config
	claudeAccount, _ := domain.NewAccount("test@example.com", "", "uuid-123")
	setup.configManager.currentAccount = claudeAccount

	input := usecases.AddAccountInput{}

	// Execute
	err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when adding duplicate account, got nil")
	}
}

// TestAddAccountUseCase_Execute_CredentialStoreFailure tests credential storage failure
func TestAddAccountUseCase_Execute_CredentialStoreFailure(t *testing.T) {
	setup := setupTest()
	ctx := context.Background()

	// Setup: Configure a current account
	claudeAccount, _ := domain.NewAccount("test@example.com", "", "uuid-123")
	setup.configManager.currentAccount = claudeAccount

	// Setup: Force credential store to fail
	setup.credentialStore.storeErr = errors.New("keychain unavailable")

	input := usecases.AddAccountInput{}

	// Execute
	err := setup.useCase.Execute(ctx, input)

	// Verify
	if err == nil {
		t.Error("Expected error when credential store fails, got nil")
	}

	// Verify no account was saved (atomic operation)
	savedAccounts, _ := setup.accountRepo.List(ctx)
	if len(savedAccounts) != 0 {
		t.Errorf("Expected no accounts to be saved on credential failure, got %d", len(savedAccounts))
	}
}

// TestAddAccountUseCase_Execute_ExplicitInput tests with all input provided
func TestAddAccountUseCase_Execute_ExplicitInput(t *testing.T) {
	setup := setupTest()
	ctx := context.Background()

	// Setup: No Claude config needed
	setup.configManager.currentAccount = nil

	input := usecases.AddAccountInput{
		Email:       "explicit@example.com",
		Alias:       "explicit-alias",
		Credentials: []byte(`{"sessionKey": "test-key"}`),
	}

	// Execute
	err := setup.useCase.Execute(ctx, input)
	// Verify
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Verify account was created with explicit data
	savedAccounts, _ := setup.accountRepo.List(ctx)
	if len(savedAccounts) != 1 {
		t.Errorf("Expected 1 account to be saved, got %d", len(savedAccounts))
	}

	if len(savedAccounts) > 0 {
		account := savedAccounts[0]
		if account.Email() != "explicit@example.com" {
			t.Errorf("Expected email explicit@example.com, got %s", account.Email())
		}
		if account.Alias() != "explicit-alias" {
			t.Errorf("Expected alias explicit-alias, got %s", account.Alias())
		}
	}
}
