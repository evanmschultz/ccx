package ports_test

import (
	"context"
	"testing"

	"github.com/evanschultz/ccx/internal/domain"
	"github.com/evanschultz/ccx/internal/ports"
)

// mockHistoryRepository is a test implementation of HistoryRepository
type mockHistoryRepository struct {
	history *domain.History
	err     error
}

func newMockHistoryRepository() *mockHistoryRepository {
	return &mockHistoryRepository{}
}

func (m *mockHistoryRepository) SaveHistory(_ context.Context, history *domain.History) error {
	if m.err != nil {
		return m.err
	}
	m.history = history
	return nil
}

func (m *mockHistoryRepository) LoadHistory(_ context.Context) (*domain.History, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.history == nil {
		// Return empty history if none exists
		return domain.NewHistory(10), nil
	}
	return m.history, nil
}

// TestHistoryRepositoryInterface validates the HistoryRepository interface contract
func TestHistoryRepositoryInterface(t *testing.T) {
	ctx := context.Background()
	repo := newMockHistoryRepository()

	// Ensure it implements the interface
	var _ ports.HistoryRepository = repo

	// Test loading empty history
	history, err := repo.LoadHistory(ctx)
	if err != nil {
		t.Errorf("LoadHistory() error = %v", err)
	}
	if history == nil {
		t.Error("LoadHistory() should return empty history, not nil")
	}
	if len(history.Entries()) != 0 {
		t.Error("LoadHistory() should return empty history initially")
	}

	// Create history with entries
	history = domain.NewHistory(10)
	entry, _ := domain.NewSwitchEntry("old@example.com", "new@example.com")
	history.AddEntry(entry)

	// Test SaveHistory
	if err := repo.SaveHistory(ctx, history); err != nil {
		t.Errorf("SaveHistory() error = %v", err)
	}

	// Test LoadHistory after saving
	loaded, err := repo.LoadHistory(ctx)
	if err != nil {
		t.Errorf("LoadHistory() error = %v", err)
	}
	if len(loaded.Entries()) != 1 {
		t.Errorf("LoadHistory() returned %d entries, want 1", len(loaded.Entries()))
	}
}

// TestHistoryRepositoryUseCaseScenarios tests common use case scenarios
func TestHistoryRepositoryUseCaseScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario func(t *testing.T, repo ports.HistoryRepository)
	}{
		{
			name: "SwitchAccount adds history entry",
			scenario: func(t *testing.T, repo ports.HistoryRepository) {
				ctx := context.Background()

				// Load existing history
				history, _ := repo.LoadHistory(ctx)

				// Add switch entry
				entry, _ := domain.NewSwitchEntry("user1@example.com", "user2@example.com")
				history.AddEntry(entry)

				// Save updated history
				if err := repo.SaveHistory(ctx, history); err != nil {
					t.Errorf("failed to save history: %v", err)
				}

				// Verify persistence
				loaded, _ := repo.LoadHistory(ctx)
				entries := loaded.Entries()
				if len(entries) != 1 {
					t.Error("history should contain one entry")
				}
				if entries[0].From() != "user1@example.com" || entries[0].To() != "user2@example.com" {
					t.Error("history entry has wrong data")
				}
			},
		},
		{
			name: "GetHistory retrieves switch history",
			scenario: func(t *testing.T, repo ports.HistoryRepository) {
				ctx := context.Background()
				history := domain.NewHistory(5)

				// Add multiple entries
				emails := []struct{ from, to string }{
					{"a@example.com", "b@example.com"},
					{"b@example.com", "c@example.com"},
					{"c@example.com", "a@example.com"},
				}

				for _, e := range emails {
					entry, _ := domain.NewSwitchEntry(domain.Email(e.from), domain.Email(e.to))
					history.AddEntry(entry)
				}

				_ = repo.SaveHistory(ctx, history)

				// Load and verify
				loaded, err := repo.LoadHistory(ctx)
				if err != nil {
					t.Errorf("failed to load history: %v", err)
				}

				entries := loaded.Entries()
				if len(entries) != 3 {
					t.Errorf("loaded %d entries, want 3", len(entries))
				}

				// Verify most recent first
				if entries[0].From() != "c@example.com" {
					t.Error("entries should be in reverse chronological order")
				}
			},
		},
		{
			name: "History respects max entries limit",
			scenario: func(t *testing.T, repo ports.HistoryRepository) {
				ctx := context.Background()
				history := domain.NewHistory(3) // Max 3 entries

				// Add more than max entries
				for i := 0; i < 5; i++ {
					from := domain.Email(string(rune('a'+i)) + "@example.com")
					to := domain.Email(string(rune('a'+i+1)) + "@example.com")
					entry, _ := domain.NewSwitchEntry(from, to)
					history.AddEntry(entry)
				}

				_ = repo.SaveHistory(ctx, history)

				// Load and verify limit
				loaded, _ := repo.LoadHistory(ctx)
				entries := loaded.Entries()
				if len(entries) != 3 {
					t.Errorf("history has %d entries, want max 3", len(entries))
				}

				// Verify we kept the most recent
				if entries[0].From() != "e@example.com" {
					t.Error("should keep most recent entries")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockHistoryRepository()
			tt.scenario(t, repo)
		})
	}
}
