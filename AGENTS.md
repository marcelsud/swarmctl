# AGENTS.md - Agent Configuration for swarmctl

## Commands

### Build
```bash
go build -o swarmctl ./cmd/swarmctl
```

### Test
```bash
# All tests
go test ./...

# Specific package tests
go test ./internal/accessories
go test ./internal/accessories -v
go test ./internal/accessories -run TestName
go test ./internal/accessories -run TestManager_Start_InvalidName

# Test all packages with verbose output
go test ./... -v

# Test with coverage
go test ./... -cover

# Run static analysis
go vet ./...
```

### Run
```bash
go run ./cmd/swarmctl <command>
```

## Code Style Guidelines

### Project Structure
- **Package naming**: Use short, clear names (`ssh`, `config`, `executor`)
- **Location**: All code in `internal/` or `pkg/cli/`, no external packages
- **Main entry**: `cmd/swarmctl/main.go`

### Imports
- **Group imports**: 
  1. Standard library
  2. External dependencies
  3. Internal packages (always prefixed with `github.com/marcelsud/swarmctl/`)
- **Example**:
```go
import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/kballard/go-shellquote"
    "github.com/marcelsud/swarmctl/internal/config"
    "github.com/marcelsud/swarmctl/internal/executor"
)
```

### Naming Conventions
- **Types**: PascalCase (e.g., `AccessoryStatus`, `Manager`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `stackName`, `runCommands`)
- **Constants**: PascalCase (e.g., `ModeSwarm`, `ModeCompose`)
- **Test functions**: `TestType_Method_Scenario` (e.g., `TestManager_Start_InvalidName`)

### Error Handling
- **Always wrap errors**: Use `fmt.Errorf("context: %w", err)`
- **Validation errors**: Clear, actionable messages
- **Example**:
```go
return fmt.Errorf("failed to start accessory: %w", err)
return fmt.Errorf("invalid name '%s': must contain only alphanumeric characters", name)
```

### Security
- **Input validation**: Strict allowlists using regex (e.g., `^[a-zA-Z0-9][a-zA-Z0-9_]{0,62}$`)
- **Shell commands**: Always use `shellquote.Join()` for user input
- **SSH**: Implement proper host key verification (no `InsecureIgnoreHostKey`)

### Testing
- **Mock executors**: Create mock implementations for testing (e.g., `AccessoriesMockExecutor`)
- **Test structure**: Use table-driven tests for multiple scenarios
- **Test naming**: Descriptive names indicating what's being tested
- **Example pattern**:
```go
func TestManager_Start_InvalidName(t *testing.T) {
    mockExec := NewAccessoriesMockExecutor()
    manager := NewManager(mockExec, "test-stack", config.ModeSwarm)
    
    err := manager.Start("redis; rm -rf /")
    if err == nil {
        t.Error("Start() should reject invalid name")
    }
}
```

### Documentation
- **Comments**: Use for complex algorithms, security considerations, and public APIs
- **TODOs**: Use `TODO:` comments for temporary workarounds
- **Function comments**: Add for public functions when behavior isn't obvious

### Environment Variables
- **Insecure mode**: `SWARMCTL_INSECURE_SSH` bypasses security (with warning)
- **Registry password**: `SWARMCTL_REGISTRY_PASSWORD` for Docker registry auth

### Git Commits
- Follow conventional commits format: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`
- Security fixes: Start with `fix(security):` and document impact
- Breaking changes: Include `BREAKING CHANGE:` in commit message body