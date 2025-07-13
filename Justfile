# CCX Development Commands
# https://github.com/casey/just

# Default recipe to display help
default:
    @just --list

# Format code with gofumpt
fmt:
    gofumpt -l -w .

# Run comprehensive linting
lint:
    golangci-lint run

# Run tests with coverage
test:
    go test -cover ./...

# Run tests with race detection
test-race:
    go test -race ./...

# Run tests with coverage and race detection
test-all:
    go test -cover -race ./...

# Build all packages
build:
    go build ./...

# Run the complete check suite (format, lint, test with race, build)
check: fmt lint test-all build

# Quick check during development (format, lint, test without race)
quick: fmt lint test

# Generate test coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Generate test coverage by function
coverage-func:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run benchmarks
bench:
    go test -bench=. ./...

# Run benchmarks with memory profiling
bench-mem:
    go test -bench=. -benchmem ./...

# Clean up generated files
clean:
    rm -f coverage.out coverage.html
    rm -f cpu.prof mem.prof
    go clean -cache

# Verify and tidy module dependencies
mod-tidy:
    go mod tidy
    go mod verify

# Check for vulnerabilities (requires govulncheck)
vuln:
    govulncheck ./...

# Run the application
run:
    go run ./cmd/ccx

# Install the application
install:
    go install ./cmd/ccx

# CI/CD check (no file modifications)
ci-check:
    gofumpt -l .
    golangci-lint run
    go test -race -coverprofile=coverage.out ./...

# Merge a branch with full testing
merge branch:
    git merge {{branch}}
    @just check

# Run integration tests
test-integration:
    go test -tags=integration -timeout=5m ./...

# Watch for changes and run tests (requires entr)
watch:
    find . -name '*.go' | entr -c just quick

# Profile CPU usage during tests
profile-cpu:
    go test -cpuprofile=cpu.prof -bench=.
    @echo "Run 'go tool pprof -http=:8080 cpu.prof' to view profile"

# Profile memory usage during tests
profile-mem:
    go test -memprofile=mem.prof -bench=.
    @echo "Run 'go tool pprof -http=:8080 mem.prof' to view profile"

# Update all tools
tools-update:
    go install mvdan.cc/gofumpt@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest

# Show current Go version and environment
env:
    go version
    go env GOPATH
    go env GOROOT
    go env GOOS
    go env GOARCH

# Find potentially dead code and unused exports
dead-code:
    @echo "=== Checking for dead code ==="
    go mod tidy
    golangci-lint run --enable-only=unused,unparam
    @echo "=== Checking for unused dependencies ==="
    go mod why -m all | grep -B1 "# " || echo "No unused dependencies found"

# Security audit - check for known vulnerabilities and security issues
security:
    @echo "=== Running security checks ==="
    golangci-lint run --enable-only=gosec
    @echo "=== Checking for vulnerabilities ==="
    govulncheck ./...
    @echo "=== Checking go.sum for issues ==="
    go mod verify

# Check code complexity and maintainability
complexity:
    @echo "=== Checking cyclomatic complexity ==="
    golangci-lint run --enable-only=gocyclo
    @echo "=== Checking code duplication ==="
    golangci-lint run --enable-only=dupl
    @echo "=== Checking function length ==="
    golangci-lint run --enable-only=funlen || true

# Run static analysis for potential bugs
static-analysis:
    @echo "=== Running staticcheck ==="
    staticcheck ./...
    @echo "=== Running go vet with all checks ==="
    go vet -all ./...

# Generate and check test coverage with thresholds
coverage-check threshold="90":
    @go test -coverprofile=coverage.out ./...
    @total=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'); \
    echo "Total coverage: $total%"; \
    if [ $(echo "$total < {{threshold}}" | bc -l) -eq 1 ]; then \
        echo "Coverage $total% is below threshold {{threshold}}%"; \
        exit 1; \
    fi

# Find TODO/FIXME comments in code
todos:
    @echo "=== TODO comments ==="
    @grep -rn "TODO\|FIXME\|XXX\|HACK" --include="*.go" . | grep -v vendor/ || echo "No TODOs found"

# Check for missing error handling
errors:
    @echo "=== Checking error handling ==="
    golangci-lint run --enable-only=errcheck
    @echo "=== Checking error wrapping ==="
    grep -rn 'fmt\.Errorf' --include="*.go" . | grep -v '%w' | head -20 || echo "All errors properly wrapped"

# Full audit - runs all quality checks
audit: security dead-code complexity static-analysis coverage-check todos errors
    @echo "=== Audit complete ==="

# Complete check - runs standard checks plus full audit
full: check audit
    @echo "=== Full check and audit complete ==="

# Aliases for common operations
alias t := test
alias ta := test-all
alias c := check
alias q := quick
alias f := fmt
alias l := lint
alias b := build