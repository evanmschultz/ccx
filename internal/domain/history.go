package domain

import (
	"errors"
	"time"
)

// SwitchEntry represents a single account switch event
type SwitchEntry struct {
	from      Email
	to        Email
	timestamp time.Time
}

// History manages account switch history
type History struct {
	entries    []*SwitchEntry
	maxEntries int
}

// NewSwitchEntry creates a new switch entry with validation
func NewSwitchEntry(from, to Email) (*SwitchEntry, error) {
	if from == "" {
		return nil, errors.New("from email cannot be empty")
	}

	if to == "" {
		return nil, errors.New("to email cannot be empty")
	}

	if from == to {
		return nil, errors.New("cannot switch to the same account")
	}

	return &SwitchEntry{
		from:      from,
		to:        to,
		timestamp: time.Now(),
	}, nil
}

// From returns the source account email
func (s *SwitchEntry) From() Email {
	return s.from
}

// To returns the destination account email
func (s *SwitchEntry) To() Email {
	return s.to
}

// Timestamp returns when the switch occurred
func (s *SwitchEntry) Timestamp() time.Time {
	return s.timestamp
}

// NewHistory creates a new history tracker with a maximum number of entries
func NewHistory(maxEntries int) *History {
	if maxEntries <= 0 {
		maxEntries = 10 // Default to 10 entries
	}

	return &History{
		entries:    make([]*SwitchEntry, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// MaxEntries returns the maximum number of entries this history will keep
func (h *History) MaxEntries() int {
	return h.maxEntries
}

// AddEntry adds a new switch entry to the history
// Most recent entries are kept at the beginning of the slice
func (h *History) AddEntry(entry *SwitchEntry) {
	if entry == nil {
		return
	}

	// Add to the beginning
	h.entries = append([]*SwitchEntry{entry}, h.entries...)

	// Trim if we exceed max entries
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[:h.maxEntries]
	}
}

// Entries returns a copy of all history entries (most recent first)
func (h *History) Entries() []*SwitchEntry {
	// Return a copy to prevent external modification
	result := make([]*SwitchEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// Clear removes all entries from the history
func (h *History) Clear() {
	h.entries = h.entries[:0]
}

// GetLastSwitch returns the most recent switch entry, or nil if history is empty
func (h *History) GetLastSwitch() *SwitchEntry {
	if len(h.entries) == 0 {
		return nil
	}
	return h.entries[0]
}

// FindSwitchesFrom returns all switches from a specific email address
func (h *History) FindSwitchesFrom(email Email) []*SwitchEntry {
	var result []*SwitchEntry
	for _, entry := range h.entries {
		if entry.from == email {
			result = append(result, entry)
		}
	}
	return result
}

// FindSwitchesTo returns all switches to a specific email address
func (h *History) FindSwitchesTo(email Email) []*SwitchEntry {
	var result []*SwitchEntry
	for _, entry := range h.entries {
		if entry.to == email {
			result = append(result, entry)
		}
	}
	return result
}
