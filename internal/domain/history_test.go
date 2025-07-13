package domain_test

import (
	"testing"
	"time"

	"github.com/evanschultz/ccx/internal/domain"
)

func TestSwitchEntry_Creation(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.Email
		to      domain.Email
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid switch entry",
			from:    "user1@example.com",
			to:      "user2@example.com",
			wantErr: false,
		},
		{
			name:    "empty from email",
			from:    "",
			to:      "user2@example.com",
			wantErr: true,
			errMsg:  "from email cannot be empty",
		},
		{
			name:    "empty to email",
			from:    "user1@example.com",
			to:      "",
			wantErr: true,
			errMsg:  "to email cannot be empty",
		},
		{
			name:    "same from and to",
			from:    "user@example.com",
			to:      "user@example.com",
			wantErr: true,
			errMsg:  "cannot switch to the same account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := domain.NewSwitchEntry(tt.from, tt.to)

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

			if entry.From() != tt.from {
				t.Errorf("From() = %v, want %v", entry.From(), tt.from)
			}

			if entry.To() != tt.to {
				t.Errorf("To() = %v, want %v", entry.To(), tt.to)
			}

			if entry.Timestamp().IsZero() {
				t.Error("Timestamp() should not be zero")
			}

			// Timestamp should be recent (within last second)
			if time.Since(entry.Timestamp()) > time.Second {
				t.Error("Timestamp() should be recent")
			}
		})
	}
}

func TestHistory_Creation(t *testing.T) {
	history := domain.NewHistory(10)

	if history.MaxEntries() != 10 {
		t.Errorf("MaxEntries() = %v, want %v", history.MaxEntries(), 10)
	}

	if len(history.Entries()) != 0 {
		t.Errorf("len(Entries()) = %v, want 0", len(history.Entries()))
	}
}

func TestHistory_AddEntry(t *testing.T) {
	history := domain.NewHistory(5)

	// Add first entry
	entry1, _ := domain.NewSwitchEntry("user1@example.com", "user2@example.com")
	history.AddEntry(entry1)

	if len(history.Entries()) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(history.Entries()))
	}

	// Add more entries
	for i := 2; i <= 5; i++ {
		from := domain.Email("user@example.com")
		to := domain.Email(string(rune('a'+i)) + "@example.com")
		entry, _ := domain.NewSwitchEntry(from, to)
		history.AddEntry(entry)
	}

	if len(history.Entries()) != 5 {
		t.Errorf("expected 5 entries, got %d", len(history.Entries()))
	}

	// Add one more entry (should remove oldest)
	entry6, _ := domain.NewSwitchEntry("userX@example.com", "userY@example.com")
	history.AddEntry(entry6)

	entries := history.Entries()
	if len(entries) != 5 {
		t.Errorf("expected 5 entries after exceeding max, got %d", len(entries))
	}

	// Most recent entry should be first
	if entries[0].From() != "userX@example.com" || entries[0].To() != "userY@example.com" {
		t.Error("most recent entry should be first")
	}

	// Oldest entry (entry1) should be removed
	for _, e := range entries {
		if e.From() == "user1@example.com" && e.To() == "user2@example.com" {
			t.Error("oldest entry should have been removed")
		}
	}
}

func TestHistory_EntriesAreCopied(t *testing.T) {
	history := domain.NewHistory(5)

	entry, _ := domain.NewSwitchEntry("user1@example.com", "user2@example.com")
	history.AddEntry(entry)

	// Get entries and try to modify
	entries1 := history.Entries()
	entries2 := history.Entries()

	// Modifying one slice shouldn't affect the other or the internal state
	if len(entries1) > 0 {
		entries1[0] = nil
	}

	if entries2[0] == nil {
		t.Error("modifying returned slice affected another copy")
	}

	// Original should still be intact
	entries3 := history.Entries()
	if len(entries3) != 1 || entries3[0] == nil {
		t.Error("modifying returned slice affected internal state")
	}
}

func TestHistory_Clear(t *testing.T) {
	history := domain.NewHistory(5)

	// Add some entries
	for i := 0; i < 3; i++ {
		entry, _ := domain.NewSwitchEntry(
			domain.Email(string(rune('a'+i))+"@example.com"),
			domain.Email(string(rune('x'+i))+"@example.com"),
		)
		history.AddEntry(entry)
	}

	if len(history.Entries()) != 3 {
		t.Fatalf("expected 3 entries before clear, got %d", len(history.Entries()))
	}

	history.Clear()

	if len(history.Entries()) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(history.Entries()))
	}
}

func TestHistory_GetLastSwitch(t *testing.T) {
	history := domain.NewHistory(5)

	// Empty history
	lastSwitch := history.GetLastSwitch()
	if lastSwitch != nil {
		t.Error("GetLastSwitch() should return nil for empty history")
	}

	// Add entries
	entry1, _ := domain.NewSwitchEntry("user1@example.com", "user2@example.com")
	history.AddEntry(entry1)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	entry2, _ := domain.NewSwitchEntry("user2@example.com", "user3@example.com")
	history.AddEntry(entry2)

	lastSwitch = history.GetLastSwitch()
	if lastSwitch == nil {
		t.Fatal("GetLastSwitch() should not return nil")
	}

	if lastSwitch.From() != "user2@example.com" || lastSwitch.To() != "user3@example.com" {
		t.Error("GetLastSwitch() returned wrong entry")
	}
}

func TestHistory_FindSwitchesFrom(t *testing.T) {
	history := domain.NewHistory(10)

	// Add various switches
	switches := []struct {
		from string
		to   string
	}{
		{"user1@example.com", "user2@example.com"},
		{"user2@example.com", "user3@example.com"},
		{"user1@example.com", "user3@example.com"},
		{"user3@example.com", "user1@example.com"},
		{"user2@example.com", "user1@example.com"},
	}

	for _, sw := range switches {
		entry, _ := domain.NewSwitchEntry(domain.Email(sw.from), domain.Email(sw.to))
		history.AddEntry(entry)
		time.Sleep(5 * time.Millisecond) // Ensure different timestamps
	}

	// Find all switches from user1
	fromUser1 := history.FindSwitchesFrom("user1@example.com")
	if len(fromUser1) != 2 {
		t.Errorf("expected 2 switches from user1, got %d", len(fromUser1))
	}

	// Find all switches from user2
	fromUser2 := history.FindSwitchesFrom("user2@example.com")
	if len(fromUser2) != 2 {
		t.Errorf("expected 2 switches from user2, got %d", len(fromUser2))
	}

	// Find switches from non-existent user
	fromNonExistent := history.FindSwitchesFrom("nonexistent@example.com")
	if len(fromNonExistent) != 0 {
		t.Errorf("expected 0 switches from non-existent user, got %d", len(fromNonExistent))
	}
}

func TestHistory_FindSwitchesTo(t *testing.T) {
	history := domain.NewHistory(10)

	// Add various switches
	switches := []struct {
		from string
		to   string
	}{
		{"user1@example.com", "user2@example.com"},
		{"user3@example.com", "user2@example.com"},
		{"user1@example.com", "user3@example.com"},
		{"user2@example.com", "user1@example.com"},
	}

	for _, sw := range switches {
		entry, _ := domain.NewSwitchEntry(domain.Email(sw.from), domain.Email(sw.to))
		history.AddEntry(entry)
	}

	// Find all switches to user2
	toUser2 := history.FindSwitchesTo("user2@example.com")
	if len(toUser2) != 2 {
		t.Errorf("expected 2 switches to user2, got %d", len(toUser2))
	}

	// Verify the switches are correct
	foundFromUser1 := false
	foundFromUser3 := false
	for _, entry := range toUser2 {
		if entry.From() == "user1@example.com" {
			foundFromUser1 = true
		}
		if entry.From() == "user3@example.com" {
			foundFromUser3 = true
		}
	}

	if !foundFromUser1 || !foundFromUser3 {
		t.Error("FindSwitchesTo did not return expected entries")
	}
}
