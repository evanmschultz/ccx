package domain_test

import (
	"testing"
	"time"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestAccount_Creation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		alias   string
		uuid    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid account",
			email:   "user@example.com",
			alias:   "work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid account without alias",
			email:   "user@example.com",
			alias:   "",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "invalid email - empty",
			email:   "",
			alias:   "work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "invalid email - no @",
			email:   "userexample.com",
			alias:   "work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email - no domain",
			email:   "user@",
			alias:   "work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email - no local part",
			email:   "@example.com",
			alias:   "work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid uuid - empty",
			email:   "user@example.com",
			alias:   "work",
			uuid:    "",
			wantErr: true,
			errMsg:  "uuid cannot be empty",
		},
		{
			name:    "invalid alias - contains spaces",
			email:   "user@example.com",
			alias:   "my work",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "alias cannot contain spaces",
		},
		{
			name:    "invalid alias - special characters",
			email:   "user@example.com",
			alias:   "work@home",
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
			errMsg:  "alias can only contain letters, numbers, hyphens, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := domain.NewAccount(tt.email, tt.alias, tt.uuid)

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

			if account.Email() != domain.Email(tt.email) {
				t.Errorf("Email() = %v, want %v", account.Email(), tt.email)
			}

			if account.Alias() != tt.alias {
				t.Errorf("Alias() = %v, want %v", account.Alias(), tt.alias)
			}

			if account.UUID() != tt.uuid {
				t.Errorf("UUID() = %v, want %v", account.UUID(), tt.uuid)
			}

			if account.ID() == "" {
				t.Error("ID() should not be empty")
			}

			if account.CreatedAt().IsZero() {
				t.Error("CreatedAt() should not be zero")
			}

			if account.LastUsed().IsZero() {
				t.Error("LastUsed() should not be zero")
			}
		})
	}
}

func TestAccount_UpdateAlias(t *testing.T) {
	account, err := domain.NewAccount("user@example.com", "work", "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	tests := []struct {
		name     string
		newAlias string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid alias update",
			newAlias: "personal",
			wantErr:  false,
		},
		{
			name:     "remove alias",
			newAlias: "",
			wantErr:  false,
		},
		{
			name:     "invalid alias - spaces",
			newAlias: "my personal",
			wantErr:  true,
			errMsg:   "alias cannot contain spaces",
		},
		{
			name:     "valid alias - with hyphen",
			newAlias: "work-laptop",
			wantErr:  false,
		},
		{
			name:     "valid alias - with underscore",
			newAlias: "work_laptop",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := account.UpdateAlias(tt.newAlias)

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

			if account.Alias() != tt.newAlias {
				t.Errorf("Alias() = %v, want %v", account.Alias(), tt.newAlias)
			}
		})
	}
}

func TestAccount_MarkUsed(t *testing.T) {
	account, err := domain.NewAccount("user@example.com", "work", "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	originalLastUsed := account.LastUsed()

	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	account.MarkUsed()

	if !account.LastUsed().After(originalLastUsed) {
		t.Error("LastUsed() should be updated to a later time")
	}
}

func TestEmail_Validation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with dots", "first.last@example.com", false},
		{"empty email", "", true},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no local part", "@example.com", true},
		{"multiple at signs", "user@@example.com", true},
		{"spaces in email", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountID_Generation(t *testing.T) {
	// Test that account IDs are unique
	seen := make(map[domain.AccountID]bool)

	for i := 0; i < 100; i++ {
		id := domain.GenerateAccountID()
		if id == "" {
			t.Fatal("GenerateAccountID() returned empty string")
		}

		if seen[id] {
			t.Fatalf("GenerateAccountID() returned duplicate ID: %s", id)
		}
		seen[id] = true

		// Verify ID format (8 characters, alphanumeric)
		if len(string(id)) != 8 {
			t.Errorf("GenerateAccountID() = %s, want 8 characters", id)
		}
	}
}
