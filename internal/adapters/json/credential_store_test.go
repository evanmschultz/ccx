package json

import (
	"context"
	"os"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestFileCredentialStore_Store(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-creds-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileCredentialStore(tmpDir)
	ctx := context.Background()

	// Create test credentials
	accountID := domain.GenerateAccountID()
	testData := []byte("test-credential-data")

	creds, err := domain.NewCredentials(accountID, testData)
	if err != nil {
		t.Fatalf("Failed to create credentials: %v", err)
	}

	// Test storing
	err = store.Store(ctx, creds)
	if err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	// Verify credentials can be retrieved
	retrieved, err := store.Retrieve(ctx, accountID)
	if err != nil {
		t.Fatalf("Failed to retrieve credentials: %v", err)
	}

	if retrieved.AccountID() != accountID {
		t.Errorf("Expected account ID %v, got %v", accountID, retrieved.AccountID())
	}
}

func TestFileCredentialStore_Retrieve(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-creds-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileCredentialStore(tmpDir)
	ctx := context.Background()

	// Create and store test credentials
	accountID := domain.GenerateAccountID()
	testData := []byte("test-credential-data")

	creds, err := domain.NewCredentials(accountID, testData)
	if err != nil {
		t.Fatalf("Failed to create credentials: %v", err)
	}

	err = store.Store(ctx, creds)
	if err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	// Test retrieval
	retrieved, err := store.Retrieve(ctx, accountID)
	if err != nil {
		t.Fatalf("Failed to retrieve credentials: %v", err)
	}

	if retrieved.AccountID() != accountID {
		t.Errorf("Expected account ID %v, got %v", accountID, retrieved.AccountID())
	}

	// Test retrieving non-existent credentials
	nonExistentID := domain.GenerateAccountID()
	_, err = store.Retrieve(ctx, nonExistentID)
	if err == nil {
		t.Fatal("Expected error when retrieving non-existent credentials")
	}
}

func TestFileCredentialStore_Delete(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-creds-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileCredentialStore(tmpDir)
	ctx := context.Background()

	// Create and store test credentials
	accountID := domain.GenerateAccountID()
	testData := []byte("test-credential-data")

	creds, err := domain.NewCredentials(accountID, testData)
	if err != nil {
		t.Fatalf("Failed to create credentials: %v", err)
	}

	err = store.Store(ctx, creds)
	if err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	// Test deletion
	err = store.Delete(ctx, accountID)
	if err != nil {
		t.Fatalf("Failed to delete credentials: %v", err)
	}

	// Verify credentials are gone
	_, err = store.Retrieve(ctx, accountID)
	if err == nil {
		t.Fatal("Expected credentials to be deleted")
	}
}
