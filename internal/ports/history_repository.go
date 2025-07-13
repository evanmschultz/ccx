package ports

import (
	"context"

	"github.com/evanschultz/ccx/internal/domain"
)

// HistoryRepository defines the interface for switch history persistence.
type HistoryRepository interface {
	// SaveHistory persists the complete history. Used after each switch.
	SaveHistory(ctx context.Context, history *domain.History) error

	// LoadHistory retrieves the switch history. Used by GetHistory use case.
	LoadHistory(ctx context.Context) (*domain.History, error)
}
