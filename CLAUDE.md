# CCX Development Assistant

You are working on ccx (Claude Code eXchange) - a multi-account switcher for Claude Code that enables seamless switching between different Claude accounts.

## Core Principles

### 🚨 BLOCKING Quality Standards
**ALL of these are MANDATORY - code CANNOT proceed until fixed:**
- ❌ **Formatting errors** (`gofumpt -l .` must show no files)
- ❌ **Linting errors** (`golangci-lint run` must pass)
- ❌ **Security violations** (`gosec` must pass - file permissions, etc.)
- ❌ **Race conditions** (`go test -race` must pass)
- ❌ **Dead code** (`unused` linter must be clean)
- ❌ **Test failures** (ALL tests must pass)
- ❌ **Build errors** (`go build` must succeed)
- ❌ **Vet issues** (`go vet` must pass)

**We DO NOT move forward with ANY of these present. Period.**

### 🔒 Security Requirements (CI Enforced)
- **File permissions**: Use 0o600 for files, 0o700 for directories
- **Directory access**: Owner-only (no group/world read)
- **Credential storage**: Must use encryption + restrictive permissions
- **No hardcoded secrets**: All sensitive data through secure channels

### Testing Philosophy
- **No mocks in domain layer** - Use real implementations or interfaces
- **Table-driven tests** - Use subtests with test cases
- **Integration tests** - Test real behavior, not units
- **100% domain coverage** - Business logic must be fully tested
- **TDD approach** - Tests BEFORE implementation

### Go Idioms & Best Practices
- **Accept interfaces, return structs** - NEVER return interfaces from constructors
- **Errors are values** - Check and handle explicitly
- **No naked returns** - Always explicit
- **No panic in libraries** - Return errors instead
- **Context for cancellation** - ALWAYS pass as first parameter
- **Small interfaces** - 1-4 methods max
- **Consumer-defined interfaces** - Each use case defines what it needs
- **Hexagonal architecture** - Core domain has zero external dependencies

### Development Workflow

1. **TDD Cycle**
   - Write failing test first (RED)
   - Minimal code to pass (GREEN)
   - Refactor with all checks passing (REFACTOR)

2. **Quality Gates** (MANDATORY ORDER - prevents CI failures)
   - `just fmt` - **ALWAYS RUN FIRST** - Format code (gofumpt)
   - `just check` - Full quality suite:
     - `gofumpt -l -w .` - Verify formatting (after fmt)
     - `golangci-lint run` - All linters + security (gosec)
     - `go test -race -cover` - Tests with race detection
     - `go build` - Successful compilation

3. **Pre-Commit Workflow** (CRITICAL for CI)
   ```bash
   just fmt      # Format first (fixes most issues)
   just check    # Verify all quality gates
   git add .     # Only after checks pass
   git commit    # Commit with confidence
   ```

4. **Incremental Development**
   - Each phase produces working, tested code
   - No moving forward until quality gates pass
   - Small, focused commits
   - **NEVER commit without running quality gates**

## Architecture

ccx follows **Hexagonal Architecture** (Ports & Adapters):

```
├── cmd/ccx/              # Application entry point
├── internal/
│   ├── domain/          # Core business logic (zero dependencies)
│   ├── ports/           # Interfaces defined by use cases
│   ├── usecases/        # Application business rules
│   └── adapters/        # Infrastructure implementations
│       ├── keychain/    # Credential storage (99designs/keyring)
│       ├── json/        # State persistence
│       ├── cli/         # CLI interface (cobra)
│       └── tui/         # Interactive UI (bubbletea)
```

## Project References
- `PROJECT.md` - What ccx is and why it exists
- `TODO.md` - Current development tasks and progress
- `DEVELOPMENT_PLAN.md` - Comprehensive TDD development plan

## Development Commands

All commands are in the `justfile`:
- `just check` - Run full quality suite (format, lint, test, build)
- `just quick` - Quick check without race detection
- `just test` - Run tests with coverage
- `just coverage` - Generate coverage report
- `just audit` - Full security and quality audit

## Success Criteria

Every feature must meet:
1. ✅ All tests pass with `-race` flag
2. ✅ 100% test coverage for business logic
3. ✅ Zero linter warnings
4. ✅ Zero security issues (gosec)
5. ✅ Documentation for public APIs
6. ✅ Integration tests for user flows

## Current Focus

See `TODO.md` for current development phase and tasks.