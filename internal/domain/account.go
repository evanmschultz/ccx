// Package domain contains the core business logic and entities for ccx.
// It has no external dependencies and represents the heart of the hexagonal architecture.
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

// AccountID represents a unique identifier for an account
type AccountID string

// Email represents a validated email address
type Email string

// Account represents a Claude Code account
type Account struct {
	id        AccountID
	email     Email
	alias     string
	uuid      string
	createdAt time.Time
	lastUsed  time.Time
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Alias validation regex (letters, numbers, hyphens, underscores)
var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// NewAccount creates a new Account with validation
func NewAccount(email, alias, uuid string) (*Account, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if uuid == "" {
		return nil, errors.New("uuid cannot be empty")
	}

	if alias != "" {
		if err := validateAlias(alias); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	return &Account{
		id:        GenerateAccountID(),
		email:     Email(email),
		alias:     alias,
		uuid:      uuid,
		createdAt: now,
		lastUsed:  now,
	}, nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// validateAlias validates an alias
func validateAlias(alias string) error {
	if strings.Contains(alias, " ") {
		return errors.New("alias cannot contain spaces")
	}

	if !aliasRegex.MatchString(alias) {
		return errors.New("alias can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}

// GenerateAccountID generates a unique 8-character account ID
func GenerateAccountID() AccountID {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return AccountID(time.Now().Format("20060102"))
	}
	return AccountID(hex.EncodeToString(bytes))
}

// ID returns the account ID
func (a *Account) ID() AccountID {
	return a.id
}

// Email returns the account email
func (a *Account) Email() Email {
	return a.email
}

// Alias returns the account alias
func (a *Account) Alias() string {
	return a.alias
}

// UUID returns the account UUID
func (a *Account) UUID() string {
	return a.uuid
}

// CreatedAt returns when the account was created
func (a *Account) CreatedAt() time.Time {
	return a.createdAt
}

// LastUsed returns when the account was last used
func (a *Account) LastUsed() time.Time {
	return a.lastUsed
}

// UpdateAlias updates the account alias with validation
func (a *Account) UpdateAlias(newAlias string) error {
	if newAlias != "" {
		if err := validateAlias(newAlias); err != nil {
			return err
		}
	}
	a.alias = newAlias
	return nil
}

// MarkUsed updates the last used timestamp
func (a *Account) MarkUsed() {
	a.lastUsed = time.Now()
}
