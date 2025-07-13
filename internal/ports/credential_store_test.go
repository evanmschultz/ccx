package ports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// mockCredentialStore is a test implementation of CredentialStore
type mockCredentialStore struct {
	credentials map[domain.AccountID]*domain.Credentials
	err         error
}

func newMockCredentialStore() *mockCredentialStore {
	return &mockCredentialStore{
		credentials: make(map[domain.AccountID]*domain.Credentials),
	}
}

func (m *mockCredentialStore) Store(_ context.Context, creds *domain.Credentials) error {
	if m.err != nil {
		return m.err
	}
	m.credentials[creds.AccountID()] = creds
	return nil
}

func (m *mockCredentialStore) Retrieve(_ context.Context, accountID domain.AccountID) (*domain.Credentials, error) {
	if m.err != nil {
		return nil, m.err
	}
	creds, ok := m.credentials[accountID]
	if !ok {
		return nil, errors.New("credentials not found")
	}
	return creds, nil
}

func (m *mockCredentialStore) Delete(_ context.Context, accountID domain.AccountID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.credentials, accountID)
	return nil
}

// TestCredentialStoreInterface validates the CredentialStore interface contract
func TestCredentialStoreInterface(t *testing.T) {
	ctx := context.Background()
	store := newMockCredentialStore()

	// Ensure it implements the interface
	var _ ports.CredentialStore = store

	// Create test credentials
	accountID := domain.GenerateAccountID()
	testData := []byte(`{"sessionKey": "test-key", "sessionKeyExpiresAt": 0}`)
	creds, err := domain.NewCredentials(accountID, testData)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	// Test Store
	if err := store.Store(ctx, creds); err != nil {
		t.Errorf("Store() error = %v", err)
	}

	// Test Retrieve
	retrieved, err := store.Retrieve(ctx, accountID)
	if err != nil {
		t.Errorf("Retrieve() error = %v", err)
	}
	if retrieved.AccountID() != accountID {
		t.Errorf("Retrieve() returned wrong credentials")
	}

	// Test Delete
	if err := store.Delete(ctx, accountID); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = store.Retrieve(ctx, accountID)
	if err == nil {
		t.Errorf("Retrieve() should return error after deletion")
	}
}

// TestCredentialStoreUseCaseScenarios tests common use case scenarios
func TestCredentialStoreUseCaseScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, store ports.CredentialStore)
	}{
		{
			name: "AddAccount stores credentials",
			scenario: func(t *testing.T, store ports.CredentialStore) {
				ctx := context.Background()
				accountID := domain.GenerateAccountID()

				// Create and store credentials
				creds, _ := domain.NewCredentials(accountID, []byte(`{"key": "value"}`))
				if err := store.Store(ctx, creds); err != nil {
					t.Errorf("failed to store credentials: %v", err)
				}

				// Verify storage
				retrieved, err := store.Retrieve(ctx, accountID)
				if err != nil {
					t.Errorf("failed to retrieve credentials: %v", err)
				}
				if retrieved.AccountID() != accountID {
					t.Error("retrieved wrong credentials")
				}
			},
		},
		{
			name: "SwitchAccount retrieves credentials",
			scenario: func(t *testing.T, store ports.CredentialStore) {
				ctx := context.Background()

				// Store credentials for multiple accounts
				for i := 0; i < 3; i++ {
					accountID := domain.GenerateAccountID()
					creds, _ := domain.NewCredentials(accountID, []byte(`{"id": "`+string(accountID)+`"}`))
					_ = store.Store(ctx, creds)
				}

				// Simulate switching by retrieving specific credentials
				targetID := domain.AccountID("specific-id")
				targetCreds, _ := domain.NewCredentials(targetID, []byte(`{"target": true}`))
				if err := store.Store(ctx, targetCreds); err != nil {
					t.Errorf("failed to store target credentials: %v", err)
				}

				retrieved, err := store.Retrieve(ctx, targetID)
				if err != nil {
					t.Errorf("failed to retrieve target credentials: %v", err)
				}

				// Verify we got the right credentials
				data, _ := retrieved.Decrypt()
				if string(data) != `{"target": true}` {
					t.Error("retrieved wrong credential data")
				}
			},
		},
		{
			name: "RemoveAccount deletes credentials",
			scenario: func(t *testing.T, store ports.CredentialStore) {
				ctx := context.Background()
				accountID := domain.GenerateAccountID()

				// Store credentials
				creds, _ := domain.NewCredentials(accountID, []byte(`{"temp": true}`))
				if err := store.Store(ctx, creds); err != nil {
					t.Errorf("failed to store credentials: %v", err)
				}

				// Delete credentials
				if err := store.Delete(ctx, accountID); err != nil {
					t.Errorf("failed to delete credentials: %v", err)
				}

				// Verify deletion
				_, err := store.Retrieve(ctx, accountID)
				if err == nil {
					t.Error("credentials should be deleted")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockCredentialStore()
			tt.scenario(t, store)
		})
	}
}
