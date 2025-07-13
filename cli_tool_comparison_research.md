# CLI Tool Comparison: Bash + Gum vs Go + Cobra/Bubbletea for CCX

## Executive Summary

After researching the trade-offs between building CCX (Claude Code account switcher) in Bash with Gum versus Go with Cobra and Bubbletea, I recommend **Go with Cobra and Bubbletea** for this project. While Bash with Gum offers rapid prototyping and zero-compile deployment, Go provides superior type safety, cross-platform compatibility, testing capabilities, and maintainability as the project grows.

## Key Findings

### 1. Complexity Management

**Bash + Gum**
- ✅ Simple for basic operations
- ✅ Quick prototyping
- ❌ Becomes unwieldy as features grow
- ❌ Limited structure for complex logic
- ❌ Error handling is primitive

**Go + Cobra/Bubbletea**
- ✅ Excellent structure with packages and modules
- ✅ Type safety catches errors at compile time
- ✅ Clear separation of concerns
- ✅ Scales well with project complexity
- ✅ Better error handling with explicit error returns

### 2. Distribution and Installation

**Bash + Gum**
- ❌ Requires users to install Gum separately
- ❌ Dependency on shell environment
- ✅ No compilation needed
- ❌ Platform-specific shell differences

**Go + Cobra/Bubbletea**
- ✅ Single binary distribution
- ✅ No runtime dependencies
- ✅ Cross-compile for all platforms
- ✅ Users don't need Go installed
- ✅ Recent tools like `soar` enable fast binary distribution

### 3. Cross-Platform Compatibility

**Bash + Gum**
- ❌ Shell differences between macOS/Linux/Windows
- ❌ Path handling varies by OS
- ❌ Windows support requires WSL or Git Bash

**Go + Cobra/Bubbletea**
- ✅ Native support for all major platforms
- ✅ Consistent behavior across OSes
- ✅ Built-in path handling abstractions

### 4. Keychain Integration

For credential storage, Go has mature libraries:
- **99designs/keyring**: Most comprehensive, supports macOS Keychain, Windows Credential Manager, Linux Secret Service
- **zalando/go-keyring**: Simpler implementation without C bindings
- **keybase/go-keychain**: Good for macOS/iOS focus

Bash would require platform-specific `security` commands on macOS, different tools on Linux/Windows.

### 5. Testing Capabilities

**Bash + Gum**
- Limited testing frameworks (BATS, bash_unit)
- No built-in coverage tools
- Runtime-only error detection

**Go + Cobra/Bubbletea**
- Built-in testing with `go test`
- Coverage analysis with `-cover`
- Unit and integration testing support
- Type checking at compile time
- testscript package for CLI testing

### 6. JSON State Management

**Go** provides:
- Native JSON marshaling/unmarshaling
- Struct tags for field mapping
- Type-safe data structures
- Error handling for malformed JSON

**Bash** requires:
- External tools like `jq`
- String manipulation prone to errors
- No type safety

### 7. Development Speed

**Initial Development**
- Bash + Gum: Faster for simple prototypes
- Go: Slightly slower initial setup

**Long-term Development**
- Bash: Slows down as complexity increases
- Go: Maintains consistent development speed

### 8. Real-World Examples

**Account Switchers in Go:**
- GitHub CLI (`gh`) - native account switching with `gh auth switch`
- Various Git account switchers on GitHub

**Credential Managers:**
- Git Credential Manager (GCM) - .NET based
- GitHub CLI - Go based

## Specific Recommendations for CCX

Given CCX's requirements:

### Use Go with Cobra + Bubbletea because:

1. **Credential Management**: The 99designs/keyring library provides unified cross-platform keychain access
2. **JSON State Files**: Go's native JSON support with structs ensures type safety
3. **Interactive Menus**: Bubbletea provides rich TUI components through the Bubbles library
4. **CLI Flags + Interactive Mode**: Cobra handles CLI structure, Bubbletea handles interactive parts
5. **Distribution**: Single binary that works everywhere without dependencies

### Architecture Recommendation:

```go
// Suggested structure
ccx/
├── cmd/           // Cobra commands
│   ├── root.go
│   ├── add.go
│   ├── switch.go
│   └── list.go
├── internal/
│   ├── config/    // JSON state management
│   ├── keychain/  // Credential storage
│   └── ui/        // Bubbletea interactive components
├── go.mod
└── main.go
```

### Implementation Pattern:
1. Use Cobra for command structure
2. Use Bubbletea for interactive account selection
3. Use 99designs/keyring for credential storage
4. Use standard library for JSON handling

## Migration Path

If you've already started with Bash + Gum:
1. Keep the Bash version as a prototype/reference
2. Implement core logic in Go incrementally
3. Start with non-interactive commands (add, list)
4. Add Bubbletea UI for interactive selection
5. Ensure feature parity before deprecating Bash version

## Conclusion

While Bash + Gum is excellent for quick prototypes and simple scripts, Go with Cobra and Bubbletea is the better choice for a production-ready CLI tool like CCX that needs to:
- Handle credentials securely across platforms
- Manage complex state with JSON files
- Provide both CLI and interactive interfaces
- Be distributed as a single binary
- Scale with additional features

The initial investment in Go setup will pay dividends in maintainability, reliability, and user experience.