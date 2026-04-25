package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// clientDef describes where a client stores its MCP server configuration.
type clientDef struct {
	Name       string
	Paths      func(home string) []string // returns candidate config file paths
	ServerKey  string                     // JSON key holding the server map
}

func clients() []clientDef {
	return []clientDef{
		{
			Name: "Claude Code",
			Paths: func(home string) []string {
				return []string{
					filepath.Join(home, ".claude.json"),
					filepath.Join(home, ".claude", "settings.json"),
				}
			},
			ServerKey: "mcpServers",
		},
		{
			Name:      "Claude Code (project)",
			Paths:     func(_ string) []string { return []string{".mcp.json"} },
			ServerKey: "mcpServers",
		},
		{
			Name:      "Claude Desktop",
			Paths:     claudeDesktopPaths,
			ServerKey: "mcpServers",
		},
		{
			Name:      "Cursor",
			Paths:     func(home string) []string { return []string{filepath.Join(home, ".cursor", "mcp.json")} },
			ServerKey: "mcpServers",
		},
		{
			Name:      "Windsurf",
			Paths:     func(home string) []string { return []string{filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")} },
			ServerKey: "mcpServers",
		},
		{
			Name:      "VS Code",
			Paths:     vscodePaths,
			ServerKey: "servers",
		},
		{
			Name:      "Cline",
			Paths:     func(home string) []string { return []string{filepath.Join(home, ".cline", "mcp_settings.json")} },
			ServerKey: "mcpServers",
		},
		{
			Name:      "Continue",
			Paths:     func(home string) []string { return []string{filepath.Join(home, ".continue", "config.json")} },
			ServerKey: "mcpServers",
		},
		{
			Name:      "Zed",
			Paths:     zedPaths,
			ServerKey: "context_servers",
		},
		{
			Name:      "Codex CLI",
			Paths:     func(home string) []string { return []string{filepath.Join(home, ".codex", "mcp.json")} },
			ServerKey: "mcpServers",
		},
	}
}

func claudeDesktopPaths(home string) []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")}
	case "linux":
		cfg := os.Getenv("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = filepath.Join(home, ".config")
		}
		return []string{filepath.Join(cfg, "Claude", "claude_desktop_config.json")}
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		return []string{filepath.Join(appdata, "Claude", "claude_desktop_config.json")}
	default:
		return nil
	}
}

func vscodePaths(home string) []string {
	var paths []string
	// Project-level config (always valid)
	paths = append(paths, filepath.Join(".vscode", "mcp.json"))

	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json"))
	case "linux":
		cfg := os.Getenv("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = filepath.Join(home, ".config")
		}
		paths = append(paths, filepath.Join(cfg, "Code", "User", "mcp.json"))
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		paths = append(paths, filepath.Join(appdata, "Code", "User", "mcp.json"))
	}
	return paths
}

func zedPaths(home string) []string {
	switch runtime.GOOS {
	case "darwin", "linux":
		cfg := os.Getenv("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = filepath.Join(home, ".config")
		}
		return []string{filepath.Join(cfg, "zed", "settings.json")}
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		return []string{filepath.Join(appdata, "Zed", "settings.json")}
	default:
		return nil
	}
}

// Scanner discovers MCP servers from all supported client config files.
type Scanner struct{}

// NewScanner creates a Scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// Scan resolves config paths, parses them, and returns all discovered servers.
func (s *Scanner) Scan() (*model.ScanResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	var result model.ScanResult
	seen := make(map[string]bool) // fingerprint -> already added

	for _, c := range clients() {
		for _, p := range c.Paths(home) {
			servers, err := parseConfigFile(p, c.ServerKey)
			if err != nil {
				continue // file missing or unparseable — skip silently
			}

			result.Clients = append(result.Clients, model.Client{
				Name:       c.Name,
				ConfigPath: p,
			})

			for name, srv := range servers {
				fp := fingerprint(srv.Command, srv.Args)
				if seen[fp] {
					continue
				}
				seen[fp] = true

				srv.Name = name
				srv.Client = c.Name
				srv.ConfigPath = p
				result.Servers = append(result.Servers, srv)
			}
		}
	}

	return &result, nil
}

// rawServerEntry represents the JSON shape of a single MCP server entry.
type rawServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// parseConfigFile reads a JSON config file and extracts MCP server entries
// from the given top-level key.
func parseConfigFile(path, serverKey string) (map[string]model.Server, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	raw, ok := top[serverKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found in %s", serverKey, path)
	}

	var entries map[string]rawServerEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("parse servers in %s: %w", path, err)
	}

	servers := make(map[string]model.Server, len(entries))
	for name, entry := range entries {
		servers[name] = model.Server{
			Command: entry.Command,
			Args:    entry.Args,
			Env:     entry.Env,
		}
	}
	return servers, nil
}

// fingerprint produces a deduplication key from command + args.
func fingerprint(command string, args []string) string {
	h := sha256.New()
	h.Write([]byte(command))
	for _, a := range args {
		h.Write([]byte{0})
		h.Write([]byte(a))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// FingerprintServer is exported for use by other packages that need deduplication.
func FingerprintServer(command string, args []string) string {
	return fingerprint(command, args)
}

// ConfigPathsForClient returns the resolved config paths for a given client name.
// Useful for targeted re-scanning.
func ConfigPathsForClient(clientName string) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}
	for _, c := range clients() {
		if strings.EqualFold(c.Name, clientName) {
			return c.Paths(home), nil
		}
	}
	return nil, fmt.Errorf("unknown client: %s", clientName)
}
