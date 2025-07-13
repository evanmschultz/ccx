package domain_test

import (
	"bytes"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestCredentials_Creation(t *testing.T) {
	tests := []struct {
		name      string
		accountID domain.AccountID
		data      []byte
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid credentials",
			accountID: "abc12345",
			data:      []byte(`{"token":"secret-token","refreshToken":"refresh-token"}`),
			wantErr:   false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			data:      []byte(`{"token":"secret-token"}`),
			wantErr:   true,
			errMsg:    "account ID cannot be empty",
		},
		{
			name:      "empty data",
			accountID: "abc12345",
			data:      []byte{},
			wantErr:   true,
			errMsg:    "credentials data cannot be empty",
		},
		{
			name:      "nil data",
			accountID: "abc12345",
			data:      nil,
			wantErr:   true,
			errMsg:    "credentials data cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds, err := domain.NewCredentials(tt.accountID, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("error message = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if creds.AccountID() != tt.accountID {
				t.Errorf("AccountID() = %v, want %v", creds.AccountID(), tt.accountID)
			}

			// Data should not be directly accessible
			// We'll test encryption/decryption separately
		})
	}
}

func TestCredentials_Encryption(t *testing.T) {
	accountID := domain.AccountID("abc12345")
	originalData := []byte(`{"token":"super-secret-token","refreshToken":"refresh-token","user":"test@example.com"}`)

	creds, err := domain.NewCredentials(accountID, originalData)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	// The encrypted data should not be the same as the original
	encryptedData := creds.EncryptedData()
	if bytes.Equal(encryptedData, originalData) {
		t.Error("encrypted data should not equal original data")
	}

	// Decrypt and verify
	decryptedData, err := creds.Decrypt()
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(decryptedData, originalData) {
		t.Errorf("decrypted data does not match original\ngot:  %s\nwant: %s", decryptedData, originalData)
	}
}

func TestCredentials_UpdateData(t *testing.T) {
	accountID := domain.AccountID("abc12345")
	originalData := []byte(`{"token":"old-token"}`)
	newData := []byte(`{"token":"new-token","refreshToken":"new-refresh"}`)

	creds, err := domain.NewCredentials(accountID, originalData)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	// Update with new data
	err = creds.UpdateData(newData)
	if err != nil {
		t.Fatalf("failed to update data: %v", err)
	}

	// Verify the update
	decryptedData, err := creds.Decrypt()
	if err != nil {
		t.Fatalf("failed to decrypt after update: %v", err)
	}

	if !bytes.Equal(decryptedData, newData) {
		t.Errorf("updated data does not match\ngot:  %s\nwant: %s", decryptedData, newData)
	}
}

func TestCredentials_UpdateData_Validation(t *testing.T) {
	accountID := domain.AccountID("abc12345")
	originalData := []byte(`{"token":"token"}`)

	creds, err := domain.NewCredentials(accountID, originalData)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	tests := []struct {
		name    string
		newData []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty data",
			newData: []byte{},
			wantErr: true,
			errMsg:  "credentials data cannot be empty",
		},
		{
			name:    "nil data",
			newData: nil,
			wantErr: true,
			errMsg:  "credentials data cannot be empty",
		},
		{
			name:    "valid update",
			newData: []byte(`{"token":"new-token"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := creds.UpdateData(tt.newData)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("error message = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCredentials_Clone(t *testing.T) {
	accountID := domain.AccountID("abc12345")
	data := []byte(`{"token":"secret"}`)

	original, err := domain.NewCredentials(accountID, data)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	// Clone the credentials
	cloned := original.Clone()

	// Verify they have the same account ID
	if cloned.AccountID() != original.AccountID() {
		t.Errorf("cloned AccountID = %v, want %v", cloned.AccountID(), original.AccountID())
	}

	// Verify they have the same encrypted data
	if !bytes.Equal(cloned.EncryptedData(), original.EncryptedData()) {
		t.Error("cloned encrypted data does not match original")
	}

	// Verify modifying the clone doesn't affect the original
	newData := []byte(`{"token":"new-secret"}`)
	err = cloned.UpdateData(newData)
	if err != nil {
		t.Fatalf("failed to update cloned data: %v", err)
	}

	// Original should still have old data
	originalDecrypted, _ := original.Decrypt()
	clonedDecrypted, _ := cloned.Decrypt()

	if bytes.Equal(originalDecrypted, clonedDecrypted) {
		t.Error("modifying clone affected original")
	}
}

func TestCredentials_Serialization(t *testing.T) {
	accountID := domain.AccountID("abc12345")
	data := []byte(`{"token":"secret","user":"test@example.com"}`)

	creds, err := domain.NewCredentials(accountID, data)
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	// Serialize
	serialized, err := creds.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize
	restored, err := domain.DeserializeCredentials(serialized)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify account ID matches
	if restored.AccountID() != creds.AccountID() {
		t.Errorf("restored AccountID = %v, want %v", restored.AccountID(), creds.AccountID())
	}

	// Verify decrypted data matches
	originalData, _ := creds.Decrypt()
	restoredData, _ := restored.Decrypt()

	if !bytes.Equal(restoredData, originalData) {
		t.Errorf("restored data does not match original\ngot:  %s\nwant: %s", restoredData, originalData)
	}
}

func TestCredentials_InvalidSerialization(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{invalid json}`),
			wantErr: true,
		},
		{
			name:    "missing account ID",
			data:    []byte(`{"encryptedData":"somedata"}`),
			wantErr: true,
		},
		{
			name:    "missing encrypted data",
			data:    []byte(`{"accountId":"abc123"}`),
			wantErr: true,
		},
		{
			name:    "invalid base64",
			data:    []byte(`{"accountId":"abc123","encryptedData":"not-base64!"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.DeserializeCredentials(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
