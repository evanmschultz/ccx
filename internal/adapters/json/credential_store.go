package json

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// FileCredentialStore implements CredentialStore using encrypted files
type FileCredentialStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileCredentialStore creates a new file-based credential store
func NewFileCredentialStore(dataDir string) ports.CredentialStore {
	return &FileCredentialStore{
		dataDir: dataDir,
	}
}

// Store securely saves credentials to an encrypted file
func (s *FileCredentialStore) Store(ctx context.Context, creds *domain.Credentials) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure credentials directory exists
	credsDir := filepath.Join(s.dataDir, "credentials")
	if err := os.MkdirAll(credsDir, 0700); err != nil { // More restrictive permissions for credentials
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Serialize credentials using domain's built-in encryption
	data, err := creds.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	// Write to file named by account ID
	filename := fmt.Sprintf("%s.json", creds.AccountID())
	filePath := filepath.Join(credsDir, filename)
	
	if err := os.WriteFile(filePath, data, 0600); err != nil { // Restrictive permissions
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// Retrieve gets credentials for an account
func (s *FileCredentialStore) Retrieve(ctx context.Context, accountID domain.AccountID) (*domain.Credentials, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build file path
	filename := fmt.Sprintf("%s.json", accountID)
	filePath := filepath.Join(s.dataDir, "credentials", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("credentials not found")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Deserialize credentials using domain's built-in decryption
	creds, err := domain.DeserializeCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize credentials: %w", err)
	}

	return creds, nil
}

// Delete removes credentials for an account
func (s *FileCredentialStore) Delete(ctx context.Context, accountID domain.AccountID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build file path
	filename := fmt.Sprintf("%s.json", accountID)
	filePath := filepath.Join(s.dataDir, "credentials", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("credentials not found")
	}

	// Remove file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}

	return nil
}