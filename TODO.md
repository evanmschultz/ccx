# CCX Development TODO

## Current Status
Starting fresh Go implementation with TDD and hexagonal architecture.

## Phase 0: Project Setup ⏳

### Infrastructure
- [ ] Initialize Go module: `go mod init github.com/evanschultz/ccx`
- [ ] Create initial directory structure
- [ ] Setup GitHub repository
- [ ] Configure GitHub Actions CI pipeline
- [ ] Create initial Makefile → justfile adaptation

### Quality Tools
- [x] `.golangci.yml` - Copied and configured
- [x] `justfile` - Copied, needs ccx adaptations
- [ ] Adapt justfile commands for ccx (Emmy → ccx)
- [ ] Verify all linters work correctly
- [ ] Setup pre-commit hooks

### Documentation
- [x] `claude.md` - Updated for ccx development
- [x] `PROJECT.md` - Created with project overview
- [x] `TODO.md` - This file
- [x] `DEVELOPMENT_PLAN.md` - Comprehensive TDD plan

## Phase 1: Domain Layer (Days 2-3) ✅

### Entities
- [x] Write `domain/account_test.go` - First failing test
- [x] Implement `Account` entity with validation
- [x] Write `domain/credentials_test.go`
- [x] Implement `Credentials` with encryption
- [x] Write `domain/history_test.go`
- [x] Implement `History` tracking
- [ ] Write `domain/migration_test.go`
- [ ] Implement `Migration` from ccswitch

### Domain Rules
- [x] Account email validation
- [x] Alias uniqueness
- [x] Account ID generation
- [x] History size limits

## Phase 2: Ports Definition (Day 4) ✅

### Repository Interfaces
- [x] `AccountRepository` interface
- [x] `CredentialStore` interface
- [x] `ConfigManager` interface
- [x] `HistoryRepository` interface

### External Service Interfaces
- [ ] `UpdateChecker` interface
- [ ] `ShellIntegration` interface

## Phase 3: Use Cases (Days 5-7) 🚧

### Core Use Cases
- [x] `AddAccountUseCase` with tests ✅
- [x] `ListAccountsUseCase` with tests ✅
- [x] `SwitchAccountUseCase` with tests ✅
- [x] `RemoveAccountUseCase` with tests ✅
- [ ] `SetAliasUseCase` with tests
- [ ] `GetHistoryUseCase` with tests

### Advanced Use Cases
- [ ] `ImportFromCCSwitchUseCase`
- [ ] `ExportShellFunctionUseCase`
- [ ] `CheckUpdateUseCase`

## Phase 4: Infrastructure Adapters (Days 8-10)

### Keychain Adapter
- [ ] Implement `KeyringStore` with 99designs/keyring
- [ ] Platform-specific tests (macOS, Linux, Windows)
- [ ] Error handling and fallbacks

### JSON Repository
- [ ] Implement `JSONAccountRepository`
- [ ] Atomic file operations
- [ ] Migration support
- [ ] Backup/restore functionality

### Config Manager
- [ ] Claude config file detection
- [ ] Safe JSON manipulation
- [ ] Config merge operations

## Phase 5: CLI Implementation (Days 11-12)

### Cobra Commands
- [ ] Root command with help
- [ ] `add` command
- [ ] `list` command
- [ ] `switch` command with quick-switch
- [ ] `remove` command
- [ ] `alias` command
- [ ] `history` command
- [ ] `update` command
- [ ] `shell-integration` command

### CLI Features
- [ ] Global flags and configuration
- [ ] Error handling and user feedback
- [ ] Colored output with lipgloss
- [ ] Progress indicators

## Phase 6: TUI Implementation (Days 13-14)

### Bubbletea Components
- [ ] Account list view
- [ ] Account details view
- [ ] Confirmation dialogs
- [ ] Menu navigation

### TUI Features
- [ ] Keyboard shortcuts
- [ ] Search/filter accounts
- [ ] Visual feedback
- [ ] Responsive layout

## Phase 7: Advanced Features (Days 15-16)

### Import/Migration
- [ ] Detect ccswitch installation
- [ ] Import accounts with preservation
- [ ] Cleanup old data option
- [ ] Progress reporting

### Shell Integration
- [ ] Bash completion
- [ ] Zsh completion
- [ ] Fish completion
- [ ] Prompt function export

### Updates
- [ ] Manual update check
- [ ] Self-update functionality
- [ ] Version comparison
- [ ] Download and replace binary

## Phase 8: Release Preparation (Day 17)

### Build & Distribution
- [ ] Cross-platform build matrix
- [ ] Binary compression
- [ ] Checksum generation
- [ ] GitHub releases automation

### Package Managers
- [ ] Homebrew formula
- [ ] AUR package (Arch Linux)
- [ ] Snap package
- [ ] Windows installer

### Documentation
- [ ] README.md with badges
- [ ] Installation guide
- [ ] Usage examples
- [ ] Troubleshooting guide

## Quality Metrics

### Test Coverage Goals
- Domain layer: 100%
- Use cases: 100%
- Adapters: 80%+
- Overall: 90%+

### Performance Targets
- Startup time: < 50ms
- Switch operation: < 100ms
- Binary size: < 10MB

## Notes

- Each checkbox represents a failing test to write first (TDD)
- No task proceeds until quality gates pass
- Use `just check` before any commit
- Update this file as tasks complete

## Completed Phases

### Research & Planning ✅
- Analyzed ccswitch implementation
- Researched bash 3.2 compatibility
- Designed gum-based UI approach
- Decided on Go implementation
- Created hexagonal architecture plan

---

*Last updated: 2025-07-13*