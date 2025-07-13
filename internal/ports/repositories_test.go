package ports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// mockAccountRepository is a test implementation of AccountRepository
type mockAccountRepository struct {
	accounts map[domain.AccountID]*domain.Account
	err      error
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts: make(map[domain.AccountID]*domain.Account),
	}
}

func (m *mockAccountRepository) Save(_ context.Context, account *domain.Account) error {
	if m.err != nil {
		return m.err
	}
	m.accounts[account.ID()] = account
	return nil
}

func (m *mockAccountRepository) FindByID(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	account, ok := m.accounts[id]
	if !ok {
		return nil, errors.New("account not found")
	}
	return account, nil
}

func (m *mockAccountRepository) FindByEmail(_ context.Context, email domain.Email) (*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, account := range m.accounts {
		if account.Email() == email {
			return account, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepository) FindByAlias(_ context.Context, alias string) (*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, account := range m.accounts {
		if account.Alias() == alias {
			return account, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *mockAccountRepository) List(_ context.Context) ([]*domain.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*domain.Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		result = append(result, account)
	}
	return result, nil
}

func (m *mockAccountRepository) Delete(_ context.Context, id domain.AccountID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.accounts, id)
	return nil
}

// TestAccountRepositoryInterface validates the AccountRepository interface contract
func TestAccountRepositoryInterface(t *testing.T) {
	ctx := context.Background()
	repo := newMockAccountRepository()

	// Ensure it implements the interface
	var _ ports.AccountRepository = repo

	// Test save and retrieve
	account, err := domain.NewAccount("test@example.com", "test-alias", "uuid-123")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Test Save
	if err := repo.Save(ctx, account); err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Test FindByID
	found, err := repo.FindByID(ctx, account.ID())
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.ID() != account.ID() {
		t.Errorf("FindByID() returned wrong account")
	}

	// Test FindByEmail
	found, err = repo.FindByEmail(ctx, account.Email())
	if err != nil {
		t.Errorf("FindByEmail() error = %v", err)
	}
	if found.Email() != account.Email() {
		t.Errorf("FindByEmail() returned wrong account")
	}

	// Test FindByAlias
	found, err = repo.FindByAlias(ctx, account.Alias())
	if err != nil {
		t.Errorf("FindByAlias() error = %v", err)
	}
	if found.Alias() != account.Alias() {
		t.Errorf("FindByAlias() returned wrong account")
	}

	// Test List
	accounts, err := repo.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("List() returned %d accounts, want 1", len(accounts))
	}

	// Test Delete
	if err := repo.Delete(ctx, account.ID()); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(ctx, account.ID())
	if err == nil {
		t.Errorf("FindByID() should return error after deletion")
	}
}

// TestAccountRepositoryUseCaseScenarios tests common use case scenarios
func TestAccountRepositoryUseCaseScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, repo ports.AccountRepository)
	}{
		{
			name: "AddAccount use case",
			scenario: func(t *testing.T, repo ports.AccountRepository) {
				ctx := context.Background()
				account, _ := domain.NewAccount("new@example.com", "new", "uuid-new")

				// Check if account already exists
				existing, _ := repo.FindByEmail(ctx, account.Email())
				if existing != nil {
					t.Error("account should not exist before adding")
				}

				// Save new account
				if err := repo.Save(ctx, account); err != nil {
					t.Errorf("failed to save account: %v", err)
				}
			},
		},
		{
			name: "ListAccounts use case",
			scenario: func(t *testing.T, repo ports.AccountRepository) {
				ctx := context.Background()

				// Add multiple accounts
				for i := 1; i <= 3; i++ {
					account, _ := domain.NewAccount(
						string(rune('a'+i))+"@example.com",
						string(rune('a'+i)),
						"uuid-"+string(rune('0'+i)),
					)
					if err := repo.Save(ctx, account); err != nil {
						t.Errorf("failed to save account: %v", err)
					}
				}

				// List all accounts
				accounts, err := repo.List(ctx)
				if err != nil {
					t.Errorf("List() error = %v", err)
				}
				if len(accounts) < 3 {
					t.Errorf("List() returned %d accounts, want at least 3", len(accounts))
				}
			},
		},
		{
			name: "SwitchAccount by alias",
			scenario: func(t *testing.T, repo ports.AccountRepository) {
				ctx := context.Background()
				account, _ := domain.NewAccount("switch@example.com", "dev", "uuid-switch")
				if err := repo.Save(ctx, account); err != nil {
					t.Errorf("failed to save account: %v", err)
				}

				// Find by alias for quick switch
				found, err := repo.FindByAlias(ctx, "dev")
				if err != nil {
					t.Errorf("FindByAlias() error = %v", err)
				}
				if found.Email() != account.Email() {
					t.Errorf("FindByAlias() returned wrong account")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockAccountRepository()
			tt.scenario(t, repo)
		})
	}
}
