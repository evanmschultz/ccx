package json

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestFileAccountRepository_Save(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo := NewFileAccountRepository(tmpDir)
	ctx := context.Background()

	// Create test account
	account, err := domain.NewAccount(
		"test@example.com",
		"test-alias",
		"550e8400-e29b-41d4-a716-446655440000",
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Test saving
	err = repo.Save(ctx, account)
	if err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "accounts.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("accounts.json file was not created")
	}

	// Verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read accounts file: %v", err)
	}

	var accounts []accountData
	if err := json.Unmarshal(data, &accounts); err != nil {
		t.Fatalf("Failed to unmarshal accounts: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}

	if accounts[0].Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", accounts[0].Email)
	}
}

func TestFileAccountRepository_FindByID(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo := NewFileAccountRepository(tmpDir)
	ctx := context.Background()

	// Create and save test account
	account, err := domain.NewAccount(
		"test@example.com",
		"test-alias",
		"550e8400-e29b-41d4-a716-446655440000",
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	err = repo.Save(ctx, account)
	if err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Test finding by ID
	found, err := repo.FindByID(ctx, account.ID())
	if err != nil {
		t.Fatalf("Failed to find account: %v", err)
	}

	if found.ID() != account.ID() {
		t.Errorf("Expected ID %v, got %v", account.ID(), found.ID())
	}

	if string(found.Email()) != string(account.Email()) {
		t.Errorf("Expected email %s, got %s", string(account.Email()), string(found.Email()))
	}
}

func TestFileAccountRepository_List(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo := NewFileAccountRepository(tmpDir)
	ctx := context.Background()

	// Create multiple test accounts
	account1, _ := domain.NewAccount("test1@example.com", "alias1", "uuid1")
	account2, _ := domain.NewAccount("test2@example.com", "alias2", "uuid2")

	// Save accounts
	repo.Save(ctx, account1)
	repo.Save(ctx, account2)

	// Test listing
	accounts, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts) != 2 {
		t.Fatalf("Expected 2 accounts, got %d", len(accounts))
	}
}

func TestFileAccountRepository_Delete(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := os.MkdirTemp("", "ccx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo := NewFileAccountRepository(tmpDir)
	ctx := context.Background()

	// Create and save test account
	account, _ := domain.NewAccount("test@example.com", "alias", "uuid")
	repo.Save(ctx, account)

	// Test deletion
	err = repo.Delete(ctx, account.ID())
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	// Verify account is gone
	_, err = repo.FindByID(ctx, account.ID())
	if err == nil {
		t.Fatal("Expected account to be deleted")
	}
}