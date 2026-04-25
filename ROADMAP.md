# mcp-upgrade Roadmap

Universal upgrade tool for MCP (Model Context Protocol) servers across all clients and package ecosystems.

## Problem

No AI coding client (Claude Code, Cursor, Windsurf, Zed, Cline, etc.) has MCP server upgrade commands. `mcpm` only upgrades servers it installed. Servers come from 6+ ecosystems (npm/npx, pip/pipx/uvx, Go, Docker, Rust/cargo, raw binaries) and are configured as static entries. There is no tool that:

1. Auto-discovers which MCP servers you have installed across all clients
2. Detects what type each server is (npm, pip, Go binary, Docker, etc.)
3. Checks all of them for updates in one pass
4. Upgrades them

## Architecture

Single-binary CLI written in **Go** (fast startup, easy cross-compilation, no runtime deps).

```
mcp-upgrade [command]
  scan        Auto-discover MCP servers from all supported client configs
  check       Compare installed versions against latest available
  upgrade     Upgrade all (or specific) servers
  list        Show all discovered servers with type, version, source
  version     Show mcp-upgrade version
```

## Core Design

### 1. Client Config Discovery

Parse config files from all known MCP clients:

| Client | Config locations |
|--------|-----------------|
| Claude Code | `~/.claude/settings.json` → `.mcpServers`, `~/.claude/projects/*/settings.json`, `.mcp.json` (project) |
| Claude Desktop | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS), `%APPDATA%\Claude\` (Win) |
| Cursor | `~/.cursor/mcp.json`, `.cursor/mcp.json` (project) |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` |
| VS Code | `~/.vscode/mcp.json`, `.vscode/mcp.json` (project) |
| Cline | `~/.cline/mcp_settings.json` |
| Continue | `~/.continue/config.json` |
| Zed | `~/.config/zed/settings.json` → `context_servers` |
| Codex CLI | `~/.codex/mcp.json` |
| Goose CLI | `~/.config/goose/config.yaml` |

**Deduplication**: Same server binary referenced by multiple clients should show once.

### 2. Server Type Detection

Infer server type from the command field:

| Pattern | Detected type | Example |
|---------|--------------|---------|
| `npx -y @org/pkg` | npm | `npx -y @upstash/context7-mcp` |
| `node /path/to/file` | local-node | `node ./server.js` |
| `uvx pkg` | pypi-uvx | `uvx mcp-jenkins` |
| `python -m pkg` | pypi | `python -m mcp_server` |
| `docker run ... image` | docker | `docker run -i mcp/kagisearch` |
| Binary in `GOPATH/bin` | go-binary | `/home/user/go/bin/terraform-mcp-server` |
| Binary with GitHub release | github-release | Detect via `--version` + GitHub API |
| Anything else | unknown | Manual upgrade |

### 3. Version Detection

For each server type:

- **npm**: `npm view <pkg> version` for latest; parse `@version` from command or check `node_modules`
- **pypi**: `pip index versions <pkg>` or PyPI JSON API
- **go-binary**: Run `<binary> --version` (common convention), compare against `gh release view`
- **docker**: `docker inspect` for current digest, `docker pull --dry-run` or registry API for latest
- **github-release**: Parse `--version` output, compare against GitHub Releases API

### 4. Upgrade Execution

Per type:

- **npm/npx**: Clear npx cache for that package (`rm -rf ~/.npm/_npx/<hash>`) + verify
- **pypi-pipx**: `pipx upgrade <pkg>`
- **pypi-uvx**: Clear uv cache for that tool (`uv cache clean <pkg>`)
- **go-binary**: Download from GitHub Releases (prefer binary) or `go install` from source
- **docker**: `docker pull <image>:<tag>`
- **github-release**: Download + replace binary (rm old, cp new, chmod +x)

### 5. Output Format

```
$ mcp-upgrade check

  Server                     Type          Current   Latest    Status
  ─────────────────────────  ────────────  ────────  ────────  ──────────
  github-mcp-server          github-rel    0.31.0    1.0.3     UPDATE
  terraform-mcp-server       go-binary     0.4.0     0.5.1     UPDATE
  mcp-atlassian              pipx          0.21.0    0.21.1    UPDATE
  @upstash/context7-mcp      npx           (cached)  2.6.1     OK (auto)
  mcp-jenkins                uvx           (cached)  3.2.0     OK (auto)
  mcp/kagisearch             docker        abc123    def456    UPDATE
  telegram-archive-mcp       local         -         -         SKIP

  3 upgradable, 2 auto-latest, 1 skipped, 1 local
```

JSON output with `--json` flag for scripting.

---

## Milestones

### v0.1.0 — Core MVP

- [ ] Go project scaffolding (cobra CLI, goreleaser)
- [ ] Config parser for Claude Code (`settings.json`, `.mcp.json`)
- [ ] Config parser for Claude Desktop
- [ ] Server type auto-detection from command string
- [ ] `scan` command — list discovered servers
- [ ] `list` command — show servers with detected types
- [ ] Version checking for npm packages (npm registry API)
- [ ] Version checking for PyPI packages (PyPI JSON API)
- [ ] Version checking for GitHub releases (GitHub API)
- [ ] `check` command — show upgrade availability
- [ ] `upgrade` command for npm (npx cache clear)
- [ ] `upgrade` command for pipx
- [ ] `upgrade` command for GitHub release binaries
- [ ] `upgrade` command for Docker images
- [ ] `--check` (dry-run) flag
- [ ] `--json` output flag
- [ ] README with usage examples
- [ ] Unit tests for config parsing and type detection
- [ ] CI with GitHub Actions (build + test + release via goreleaser)
- [ ] GPL-3.0 license
- [ ] Homebrew formula

### v0.2.0 — Multi-Client Support

- [ ] Config parser for Cursor
- [ ] Config parser for Windsurf
- [ ] Config parser for VS Code / Copilot
- [ ] Config parser for Cline
- [ ] Config parser for Continue
- [ ] Config parser for Zed
- [ ] Config parser for Codex CLI
- [ ] Config parser for Goose CLI
- [ ] Server deduplication across clients
- [ ] `--client` flag to filter by client
- [ ] Cross-platform config path resolution (macOS, Linux, Windows)

### v0.3.0 — Smart Detection

- [ ] Go binary detection via GOPATH heuristics
- [ ] Cargo/Rust binary detection
- [ ] Auto-detect GitHub repo from binary metadata (embedded version strings, help text)
- [ ] Handle `uvx` servers (uv tool cache management)
- [ ] Detect version from `--version`, `version`, `-v`, `-V` subcommands
- [ ] Handle servers with non-standard version output

### v0.4.0 — Safety & UX

- [ ] Interactive confirmation before upgrades (default on)
- [ ] `--yes` flag to skip confirmation
- [ ] Backup binary before overwriting
- [ ] Rollback on failed upgrade
- [ ] Colored terminal output with progress bars
- [ ] `--verbose` and `--quiet` flags
- [ ] Config file (`~/.config/mcp-upgrade/config.yaml`) for overrides (pin versions, skip servers, custom GitHub repos)

### v1.0.0 — Production Ready

- [ ] Comprehensive test suite (>90% coverage)
- [ ] Integration tests with mock registries
- [ ] Man page
- [ ] Shell completions (bash, zsh, fish)
- [ ] `self-update` command
- [ ] Cron/launchd scheduling helper
- [ ] Windows support
- [ ] Stable API for `--json` output

### Future Ideas

- Plugin system for custom server types
- Web UI dashboard
- Notification integration (Slack/Discord/Telegram when updates available)
- Proxy/mirror support for air-gapped environments
- Lock file for reproducible MCP server sets
- `init` command to bootstrap MCP servers from a manifest
