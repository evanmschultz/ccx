# CCX - Claude Code eXchange

## What is CCX?

CCX is a command-line tool that enables seamless switching between multiple Claude Code accounts. It securely manages credentials and configuration, allowing developers to work with different Claude accounts (personal, work, client projects) without manual re-authentication.

## Why CCX?

### The Problem
Claude Code currently supports only one account at a time. Switching between accounts requires:
- Manual logout/login process
- Re-entering credentials
- Losing current session state
- No easy way to manage multiple professional contexts

This is particularly challenging for:
- Consultants working with multiple clients
- Developers with separate personal/work accounts
- Teams sharing development environments
- Users managing different subscription tiers

### The Solution
CCX provides:
- **Quick Switching**: Change accounts with a single command
- **Secure Storage**: Platform-native credential management (macOS Keychain, Linux secret stores)
- **Account Aliases**: Name accounts for easy identification (`ccx work`, `ccx personal`)
- **Dual Interface**: Both CLI commands and interactive TUI
- **History Tracking**: See recent account switches
- **Zero Dependencies**: Single Go binary, no runtime requirements

## Key Features

### 1. Account Management
- Add current Claude account to managed list
- Remove accounts securely
- List all managed accounts with status
- Import from existing ccswitch installations

### 2. Smart Switching
```bash
# By alias
ccx work

# By email
ccx user@example.com

# By index
ccx 2

# Interactive menu
ccx
```

### 3. Security First
- Credentials stored in OS-native secure storage
- No plaintext passwords on disk
- Encrypted state management
- Secure cleanup on account removal

### 4. Developer Experience
- Shell integration for prompt status
- Quick-switch shortcuts
- Non-intrusive update notifications
- Backwards compatible with ccswitch

## Technical Design

### Architecture
- **Hexagonal Architecture**: Clean separation of business logic and infrastructure
- **TDD Development**: 100% test coverage for domain logic
- **Go Best Practices**: Idiomatic Go with strict quality standards

### Core Components
1. **Domain Layer**: Account entities, business rules
2. **Use Cases**: Add, switch, remove, list operations
3. **Adapters**: 
   - Keychain integration (99designs/keyring)
   - JSON state persistence
   - CLI interface (Cobra)
   - TUI interface (Bubbletea)

### Quality Standards
- All code passes `golangci-lint` with strict configuration
- Race-condition free (`go test -race`)
- Security audited with `gosec`
- Zero tolerance for technical debt

## Installation

```bash
# macOS
brew install evanschultz/tap/ccx

# Linux/Windows
curl -sSL https://github.com/evanschultz/ccx/releases/latest/download/ccx_$(uname -s)_$(uname -m) -o /usr/local/bin/ccx
chmod +x /usr/local/bin/ccx

# From source
go install github.com/evanschultz/ccx/cmd/ccx@latest
```

## Migration from ccswitch

CCX automatically detects and imports existing ccswitch configurations:
1. On first run, checks for `~/.claude-switch-backup/`
2. Imports accounts with preserved settings
3. Offers to clean up old data
4. Maintains full backwards compatibility

## Future Enhancements

- Team account sharing (with approval flow)
- Account-specific environment variables
- Integration with CI/CD pipelines
- REST API for automation
- Account usage analytics

## Contributing

See `DEVELOPMENT_PLAN.md` for the development approach and `TODO.md` for current tasks.