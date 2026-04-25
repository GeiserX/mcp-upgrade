# mcp-upgrade - AI Agent Instructions

## Project Overview

**Description**: Universal upgrade tool for MCP (Model Context Protocol) servers across all AI coding clients and package ecosystems

**Visibility**: Public repository
**Development OS**: macOS

### Repository
- **Platform**: GitHub
- **URL**: https://github.com/GeiserX/mcp-upgrade

## Technology Stack

### Languages
- Go

### Frameworks & Tools
- **CLI**: cobra
- **Release**: goreleaser (cross-platform: linux/darwin/windows, amd64/arm64)
- **Distribution**: Homebrew (`GeiserX/mcp-upgrade/mcp-upgrade`), `go install`, GitHub Releases
- **Linting**: golangci-lint
- **Coverage**: Codecov
- **CI/CD**: GitHub Actions (CI on push/PR to main, Release on `v*` tags)

### AI Technology Selection
For technologies beyond those listed, analyze the codebase and suggest appropriate solutions.

## Architecture

### Commands (`cmd/`)
- `root.go` — cobra root command and global flags
- `scan.go` — discover MCP servers from all supported client configs
- `check.go` — check servers for available updates via package registries
- `upgrade.go` — execute upgrades using appropriate package managers

### Internal packages (`internal/`)
- `config` — client config file discovery and parsing (Claude Code, Cursor, Windsurf, VS Code, Cline, Continue, Zed, Codex CLI)
- `detect` — server type inference (npx, pipx, uvx, go-binary, github-release, docker, local)
- `model` — shared data types
- `registry` — version checking against npm, PyPI, GitHub Releases, Docker Hub
- `upgrade` — upgrade execution per server type

### Version injection
Version is injected at build time via ldflags: `-X github.com/GeiserX/mcp-upgrade/cmd.version={{.Version}}`

## Development Guidelines

### Communication Style
- Be concise and direct
- Developer context: devops
- Skill level: Senior

### Workflow Rules
- Match the codebase's existing style and patterns
- Confirm before making significant architectural changes
- All commands must support `--json` for machine-readable output
- Concurrent operations should respect the max-parallelism pattern (currently 5)

### Testing Changes
```bash
go build ./...
go test -v -race ./...
```

### Important Files to Read First
Before making changes, read these files to understand the project:
- README.md
- ROADMAP.md
- SECURITY.md

### CI/CD & Infrastructure
- **CI/CD Platform**: GitHub Actions
- **CI**: Runs `go build`, `go test -v -race -coverprofile`, and `golangci-lint` on ubuntu + macOS
- **Release**: Triggered by pushing a `v*` tag. GoReleaser builds cross-platform binaries and creates GitHub Releases

### Releasing New Versions

1. Update ROADMAP.md if the release includes planned features
2. Commit changes: `git commit -m "feat: description"` or `fix:` etc.
3. Tag and push: `git tag v<version> && git push && git push origin v<version>`

CI (`.github/workflows/release.yml`) does the rest automatically:
- Builds binaries for linux/darwin/windows (amd64/arm64)
- Creates GitHub Release with archives and checksums
- Generates changelog from conventional commits

**Do NOT manually** create GitHub releases or upload binaries — CI will handle them.

### Adding Support for New Clients
1. Add config path detection in `internal/config/`
2. Parse the client's MCP server config format
3. Map to the shared `model` types
4. Add to the client discovery loop

### Adding Support for New Server Types
1. Add detection logic in `internal/detect/`
2. Add version checking in `internal/registry/`
3. Add upgrade execution in `internal/upgrade/`
4. Update the type detection table in README.md

## Best Practices

- **Write clean code**: Prioritize readability and maintainability
- **Handle errors properly**: Don't ignore errors, handle them appropriately
- **Consider security**: Review code for potential security vulnerabilities
- **Conventional commits**: Use conventional commit messages (feat:, fix:, docs:, chore:, refactor:, test:, style:)
- **Semantic versioning**: Follow semver (MAJOR.MINOR.PATCH) for version numbers
- **Concurrent safety**: Use goroutines with proper synchronization for parallel registry queries
- **Cross-platform**: Test on both Linux and macOS; handle path differences

## Learned Patterns

- **CI handles releases end-to-end** — pushing a `v*` tag triggers GoReleaser. Do NOT manually create releases.
- **Version is build-time injected** — never hardcode version strings. GoReleaser sets it via ldflags.
- **npx/uvx/docker servers are "auto-latest"** — they resolve to the latest version on each run, so they show AUTO status, not UPDATE.
- **Server deduplication** — the same binary referenced by multiple clients should appear once in output.
- **goreleaser `skip_upload: true` for brew** — Homebrew formula is generated but not auto-pushed; update the tap repo manually after release.

## Self-Improving Configuration

This file should evolve as we work together:
1. Track coding patterns and preferences
2. Note corrections made to suggestions
3. Update periodically with learned preferences

## Security Notice

> **Do not commit secrets to the repository or to the live app.**
> Always use secure standards to transmit sensitive information.
> Use environment variables, secret managers, or secure vaults for credentials.

**Security Audit Recommendation:** When making changes that involve authentication, data handling, API endpoints, or dependencies, proactively offer to perform a security review of the affected code.

---

*Generated by [LynxPrompt](https://lynxprompt.com) CLI*
