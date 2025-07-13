# CCX Development Plan - TDD with Hexagonal Architecture

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Presentation Layer                        │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │   CLI/Cobra  │  │  Bubbletea   │  │   REST API       │    │
│  │   Commands   │  │     TUI      │  │   (future)       │    │
│  └──────┬──────┘  └──────┬───────┘  └────────┬──────────┘    │
└─────────┼────────────────┼───────────────────┼─────────────────┘
          │                │                   │
          ▼                ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Application Core                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Use Cases (Ports)                     │   │
│  │  • SwitchAccount  • AddAccount   • RemoveAccount       │   │
│  │  • ListAccounts   • SetAlias     • GetHistory          │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   Domain Entities                        │   │
│  │  • Account        • Credentials   • History             │   │
│  │  • Alias          • Migration     • Update              │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
          ▲                ▲                   ▲
          │                │                   │
┌─────────┼────────────────┼───────────────────┼─────────────────┐
│         │          Infrastructure Layer       │                 │
│  ┌──────┴──────┐  ┌──────┴───────┐  ┌───────┴──────────┐     │
│  │  Keychain   │  │     JSON     │  │    HTTP Client   │     │
│  │   Adapter   │  │  Repository  │  │   (for updates)  │     │
│  └─────────────┘  └──────────────┘  └──────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

## Core Principles

1. **Dependency Inversion**: Core domain defines interfaces, adapters implement them
2. **Consumer-Defined Interfaces**: Each use case defines exactly what it needs
3. **TDD**: Write failing tests first, make them pass, refactor
4. **Zero Tolerance**: All linters and tests must pass before moving forward
5. **Incremental Delivery**: Each phase produces a working, tested component

## Development Phases

### Phase 0: Project Setup (Day 1)
```bash
# Initial structure
ccx/
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   └── ccx/
│       └── main.go
├── internal/
│   ├── domain/
│   ├── ports/
│   ├── adapters/
│   └── usecases/
└── test/
    └── integration/
```

**Tasks:**
1. [ ] Initialize Go module: `go mod init github.com/evanschultz/ccx`
2. [ ] Create Makefile with all quality gates
3. [ ] Setup golangci-lint configuration
4. [ ] Create GitHub Actions CI pipeline
5. [ ] Write first failing test: `domain/account_test.go`

**Makefile:**
```makefile
.PHONY: all test lint fmt vet sec

all: fmt vet lint test

fmt:
	@echo "Running gofumpt..."
	@gofumpt -l -w .
	@echo "✓ gofumpt passed"

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run
	@echo "✓ golangci-lint passed"

test:
	@echo "Running tests with race detector..."
	@go test -race -coverprofile=coverage.out ./...
	@echo "✓ tests passed"

sec:
	@echo "Running security checks..."
	@gosec -quiet ./...
	@echo "✓ gosec passed"

check: fmt vet lint test sec
	@echo "✅ All checks passed!"
```

### Phase 1: Domain Layer (Day 2-3)

**Domain Entities (Pure Go, No Dependencies):**

```go
// internal/domain/account.go
type AccountID string
type Email string

type Account struct {
    ID        AccountID
    Email     Email
    Alias     string
    UUID      string
    CreatedAt time.Time
    LastUsed  time.Time
}

// internal/domain/credentials.go
type Credentials struct {
    AccountID AccountID
    Data      []byte // encrypted
}
```

**TDD Steps:**
1. [ ] Write test for Account creation validation
2. [ ] Implement Account with validation rules
3. [ ] Write test for Credentials encryption
4. [ ] Implement Credentials with encryption
5. [ ] Write test for History tracking
6. [ ] Implement History entity

**Quality Gates:** Each entity must have 100% test coverage

### Phase 2: Ports Definition (Day 4)

**Consumer-Defined Interfaces:**

```go
// internal/ports/repositories.go
package ports

// AccountRepository - defined by use cases that need accounts
type AccountRepository interface {
    Save(ctx context.Context, account *domain.Account) error
    FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
    FindByEmail(ctx context.Context, email domain.Email) (*domain.Account, error)
    FindByAlias(ctx context.Context, alias string) (*domain.Account, error)
    List(ctx context.Context) ([]*domain.Account, error)
    Delete(ctx context.Context, id domain.AccountID) error
}

// CredentialStore - defined by use cases that need credentials
type CredentialStore interface {
    Store(ctx context.Context, creds *domain.Credentials) error
    Retrieve(ctx context.Context, accountID domain.AccountID) (*domain.Credentials, error)
    Delete(ctx context.Context, accountID domain.AccountID) error
}

// ConfigManager - defined by use cases that need Claude config
type ConfigManager interface {
    GetCurrentAccount(ctx context.Context) (*domain.Account, error)
    SetCurrentAccount(ctx context.Context, account *domain.Account) error
}
```

**TDD Steps:**
1. [ ] Write interface usage tests (mocked implementations)
2. [ ] Ensure interfaces are minimal and focused
3. [ ] Document each interface's contract

### Phase 3: Use Cases (Day 5-7)

**Implement One Use Case at a Time:**

```go
// internal/usecases/add_account.go
type AddAccountUseCase struct {
    accounts    ports.AccountRepository
    credentials ports.CredentialStore
    config      ports.ConfigManager
}

func (uc *AddAccountUseCase) Execute(ctx context.Context) error {
    // 1. Get current account from Claude config
    // 2. Create Account entity
    // 3. Save credentials
    // 4. Save account
    // 5. Update active account
}
```

**TDD Order:**
1. [ ] AddAccount use case
2. [ ] ListAccounts use case
3. [ ] SwitchAccount use case
4. [ ] RemoveAccount use case
5. [ ] SetAlias use case
6. [ ] GetHistory use case

**Each Use Case:**
- Write integration test first
- Mock all dependencies
- Implement use case
- Ensure 100% coverage

### Phase 4: Infrastructure Adapters (Day 8-10)

**Keychain Adapter:**
```go
// internal/adapters/keychain/keyring.go
type KeyringStore struct {
    ring keyring.Keyring
}

func (k *KeyringStore) Store(ctx context.Context, creds *domain.Credentials) error {
    // Implement using 99designs/keyring
}
```

**TDD Approach:**
1. [ ] Write adapter tests using test keyring
2. [ ] Implement KeyringStore
3. [ ] Test on each platform (CI matrix)

**JSON Repository:**
1. [ ] Write file-based repository tests
2. [ ] Implement with atomic writes
3. [ ] Add migration support

### Phase 5: CLI Implementation (Day 11-12)

**Cobra Commands:**
```go
// cmd/ccx/add.go
var addCmd = &cobra.Command{
    Use:   "add",
    Short: "Add current Claude account",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Setup dependencies
        // 2. Create use case
        // 3. Execute
        // 4. Handle output
    },
}
```

**TDD with testscript:**
```
# test/integration/add_account.txtar
exec ccx add
stdout 'Added account'
exists $HOME/.ccx/state.json
```

### Phase 6: TUI Implementation (Day 13-14)

**Bubbletea Interface:**
1. [ ] Create account list component
2. [ ] Add keyboard navigation
3. [ ] Implement selection handling
4. [ ] Add confirmation dialogs

**Test with teatest package**

### Phase 7: Advanced Features (Day 15-16)

1. [ ] Import from ccswitch
2. [ ] Shell integration
3. [ ] Update mechanism
4. [ ] History viewing

### Phase 8: Release Preparation (Day 17)

1. [ ] Cross-platform builds
2. [ ] Installation scripts
3. [ ] Documentation
4. [ ] Homebrew formula

## Testing Strategy

### Unit Tests
- Domain entities: 100% coverage
- Use cases: Mock all dependencies
- Adapters: Test against interfaces

### Integration Tests
- testscript for CLI commands
- Real keychain on CI (platform matrix)
- File system operations

### Quality Tools Configuration

**.golangci.yml:**
```yaml
linters:
  enable:
    - gofumpt
    - gosec
    - errcheck
    - ineffassign
    - deadcode
    - structcheck
    - varcheck
    - govet
    - unconvert
    - prealloc
    - misspell
    - nakedret
    - gocritic
    - gocyclo
    - dupl
    - gocognit

linters-settings:
  gocyclo:
    min-complexity: 10
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
```

## Success Criteria

Each phase must meet:
1. ✅ All tests pass with -race flag
2. ✅ 100% test coverage for business logic
3. ✅ Zero linter warnings
4. ✅ Zero security issues (gosec)
5. ✅ Documentation for public APIs
6. ✅ Integration tests for user flows

## Development Workflow

1. **Start each feature:**
   ```bash
   git checkout -b feature/add-account
   ```

2. **TDD Cycle:**
   ```bash
   # Write failing test
   make test  # ❌ Red
   
   # Implement feature
   make test  # ✅ Green
   
   # Refactor
   make check # ✅ All quality gates
   ```

3. **Before commit:**
   ```bash
   make check  # Must pass 100%
   ```

4. **CI enforces all checks**

## Incremental Milestones

1. **Week 1**: Domain + Ports + Core Use Cases
   - Deliverable: Core business logic fully tested
   
2. **Week 2**: Adapters + CLI
   - Deliverable: Working CLI with real persistence
   
3. **Week 3**: TUI + Polish
   - Deliverable: Full-featured ccx ready for release

This plan ensures we build a maintainable, well-tested, production-quality tool that follows Go best practices and hexagonal architecture principles.