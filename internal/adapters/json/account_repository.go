// Package json provides file-based implementations of domain repositories using JSON for persistence.
package json

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// FileAccountRepository implements AccountRepository using JSON files
type FileAccountRepository struct {
	dataDir string
	mu      sync.RWMutex
}

// accountData represents the JSON structure for persistence
type accountData struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Alias     string `json:"alias"`
	UUID      string `json:"uuid"`
	CreatedAt string `json:"created_at"`
	LastUsed  string `json:"last_used"`
}

// NewFileAccountRepository creates a new file-based account repository
func NewFileAccountRepository(dataDir string) ports.AccountRepository {
	return &FileAccountRepository{
		dataDir: dataDir,
	}
}

// Save persists an account to the JSON file
func (r *FileAccountRepository) Save(_ context.Context, account *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(r.dataDir, 0o700); err != nil {
		return err
	}

	// Load existing accounts
	accounts, err := r.loadAccounts()
	if err != nil {
		return err
	}

	// Convert domain account to accountData
	data := accountData{
		ID:        string(account.ID()),
		Email:     string(account.Email()),
		Alias:     account.Alias(),
		UUID:      account.UUID(),
		CreatedAt: account.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		LastUsed:  account.LastUsed().Format("2006-01-02T15:04:05Z07:00"),
	}

	// Check if account already exists (update scenario)
	found := false
	for i, acc := range accounts {
		if acc.ID == data.ID {
			accounts[i] = data
			found = true
			break
		}
	}

	// If not found, append new account
	if !found {
		accounts = append(accounts, data)
	}

	// Save back to file
	return r.saveAccounts(accounts)
}

// FindByID retrieves an account by its ID
func (r *FileAccountRepository) FindByID(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts, err := r.loadAccounts()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.ID == string(id) {
			return r.convertToAccount(acc)
		}
	}

	return nil, errors.New("account not found")
}

// FindByEmail retrieves an account by email
func (r *FileAccountRepository) FindByEmail(_ context.Context, email domain.Email) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts, err := r.loadAccounts()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.Email == string(email) {
			return r.convertToAccount(acc)
		}
	}

	return nil, errors.New("account not found")
}

// FindByAlias retrieves an account by alias
func (r *FileAccountRepository) FindByAlias(_ context.Context, alias string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts, err := r.loadAccounts()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.Alias == alias {
			return r.convertToAccount(acc)
		}
	}

	return nil, errors.New("account not found")
}

// List returns all accounts
func (r *FileAccountRepository) List(_ context.Context) ([]*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts, err := r.loadAccounts()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Account, 0, len(accounts))
	for _, acc := range accounts {
		account, err := r.convertToAccount(acc)
		if err != nil {
			return nil, err
		}
		result = append(result, account)
	}

	return result, nil
}

// Delete removes an account
func (r *FileAccountRepository) Delete(_ context.Context, id domain.AccountID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	accounts, err := r.loadAccounts()
	if err != nil {
		return err
	}

	// Find and remove account
	found := false
	for i, acc := range accounts {
		if acc.ID == string(id) {
			accounts = append(accounts[:i], accounts[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return errors.New("account not found")
	}

	return r.saveAccounts(accounts)
}

// loadAccounts loads accounts from the JSON file
func (r *FileAccountRepository) loadAccounts() ([]accountData, error) {
	filePath := filepath.Join(r.dataDir, "accounts.json")

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []accountData{}, nil
	}

	data, err := os.ReadFile(filePath) // #nosec G304 - controlled file path within app data directory
	if err != nil {
		return nil, err
	}

	var accounts []accountData
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

// saveAccounts saves accounts to the JSON file
func (r *FileAccountRepository) saveAccounts(accounts []accountData) error {
	filePath := filepath.Join(r.dataDir, "accounts.json")

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o600)
}

// convertToAccount converts accountData to domain.Account
func (r *FileAccountRepository) convertToAccount(data accountData) (*domain.Account, error) {
	// Parse timestamps
	createdAt, err := time.Parse("2006-01-02T15:04:05Z07:00", data.CreatedAt)
	if err != nil {
		return nil, err
	}

	lastUsed, err := time.Parse("2006-01-02T15:04:05Z07:00", data.LastUsed)
	if err != nil {
		return nil, err
	}

	// Reconstruct account with original values
	return domain.ReconstructAccount(
		domain.AccountID(data.ID),
		data.Email,
		data.Alias,
		data.UUID,
		createdAt,
		lastUsed,
	)
}
