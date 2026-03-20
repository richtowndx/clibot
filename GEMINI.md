# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

clibot is a lightweight middleware that bridges ACP-compatible AI CLI tools (Claude Code, Gemini CLI, OpenCode) to IM platforms (Discord, Telegram, Feishu, DingTalk, QQ). It enables using desktop AI programming assistants from mobile phones with streaming responses - no public IP required.

**Important**: This project will be **open-sourced** publicly. **ALL code, documentation, and comments must be in English.** This includes: variable/function names, error messages, comments, documentation, and commit messages.

## Common Commands

### Build & Install
```bash
make build          # Build binary to bin/clibot
make install        # Install to $GOPATH/bin
make build-all      # Build for Linux and macOS
```

### Testing
```bash
make test                      # Run all tests with coverage
make test-short                # Run short tests only
make test-coverage             # Generate HTML coverage report
go test ./internal/core/...    # Run tests for specific package
go test -run TestName ./path   # Run single test
```

### Code Quality
```bash
make fmt            # Format code (required before committing)
make vet            # Run go vet
make lint           # Run golangci-lint
make check          # Run fmt, vet, and test
```

### Development
```bash
make deps           # Download dependencies
make deps-tidy      # Tidy go.mod
make dev            # Build with race detection
```

## Architecture

### Core Components

```
internal/
├── core/           # Engine, Config, Session management
│   ├── engine.go   # Central orchestration - routes messages between bots and CLIs
│   ├── config.go   # YAML configuration loading and validation
│   ├── types.go    # Config structs and Session state
│   └── hook.go     # HTTP hook server for CLI notifications
├── cli/            # CLI adapters (implements CLIAdapter interface)
│   ├── interface.go
│   ├── acp.go      # ACP protocol adapter (recommended)
│   ├── claude.go   # Claude Code via tmux (hook mode)
│   ├── gemini.go   # Gemini CLI via tmux
│   └── opencode.go # OpenCode CLI via tmux
├── bot/            # Bot adapters (implements BotAdapter interface)
│   ├── interface.go
│   ├── discord.go
│   ├── telegram.go
│   ├── feishu.go
│   ├── dingtalk.go
│   └── qq.go
├── proxy/          # Network proxy support
├── watchdog/       # Tmux monitoring for hook mode
└── logger/         # Structured logging with logrus
```

### Key Interfaces

**BotAdapter** (`internal/bot/interface.go`):
- `Start(messageHandler)` - Start bot and listen for messages
- `SendMessage(channel, message)` - Send message to IM platform
- `SupportsTypingIndicator()` / `AddTypingIndicator()` / `RemoveTypingIndicator()` - Typing feedback
- `SetProxyManager()` / `Stop()` - Proxy and lifecycle

**CLIAdapter** (`internal/cli/interface.go`):
- `SendInput(sessionName, input)` - Send user input to CLI
- `HandleHookData(data)` - Process hook notifications (hook mode only)
- `IsSessionAlive(sessionName)` - Check session status
- `CreateSession(sessionName, workDir, startCmd, transportURL)` - Create new CLI session

### Operation Modes

1. **ACP Mode (Recommended)**: Direct protocol communication via `acp-go-sdk`. No tmux required. Supports streaming responses.
2. **Hook Mode**: Uses tmux for session management. Requires CLI hook configuration. Real-time notifications via HTTP webhook.

### Message Flow

```
User Message → BotAdapter.Start callback → Engine.HandleBotMessage
    → Permission check (whitelist/admin)
    → Route to CLI: Engine.SendInput → CLIAdapter.SendInput
    → CLI processes → Response via:
        - ACP: SessionUpdate callback → Engine.SendResponseToSession
        - Hook: HTTP webhook → Engine.handleHook → Engine.SendToBot
    → BotAdapter.SendMessage → User
```

## Testing Guidelines

- Use `github.com/stretchr/testify` for all tests
- Maintain test coverage above 50%
- Write table-driven tests for multiple scenarios
- Test files follow `*_test.go` naming convention

## Code Quality Standards

- **MUST**: Code must compile successfully before considering work complete
- **MUST**: All tests must pass (`go test ./...`) before marking task as complete
- **MUST**: Run `make fmt` to format Golang code before committing
- **MUST**: No failing tests or compilation errors in final deliverables

## Git Workflow

- **DO NOT** automatically commit without explicit user instruction
- **DO NOT** run `git commit` or `git push` unless explicitly requested
- **One atomic change per commit**: Each commit should contain only one logical change

### Commit Message Convention

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation
- `refactor:` code refactoring
- `opt:` performance optimization
- `security:` security fixes
- `chore:` build/tooling

**Subject line limit**: Maximum 150 characters for readability in git logs.

## Configuration

Configuration is loaded from YAML with environment variable expansion (`${VAR_NAME}`).

Key sections:
- `sessions`: CLI tool sessions (name, cli_type, work_dir, start_cmd, transport)
- `bots`: IM platform configurations (discord, telegram, feishu, dingtalk, qq)
- `cli_adapters`: CLI adapter settings (timeout, env vars)
- `security`: Whitelist and admin user IDs per platform
- `proxy`: Network proxy configuration
