package upgrade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// UVXUpgrader runs uv tool upgrade for uv-managed MCP servers.
type UVXUpgrader struct{}

func (u *UVXUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypeUVX && server.Package != ""
}

func (u *UVXUpgrader) Upgrade(server *model.Server) error {
	out, err := execCommand("uv", "tool", "upgrade", server.Package)
	if err == nil {
		return nil
	}

	// Fall back to clearing the uv tool cache if the upgrade subcommand is not available.
	home, err2 := os.UserHomeDir()
	if err2 != nil {
		return fmt.Errorf("uv tool upgrade %s failed: %s", server.Package, out)
	}

	cacheDir := filepath.Join(home, ".local", "share", "uv", "tools", server.Package)
	if err2 := os.RemoveAll(cacheDir); err2 != nil {
		return fmt.Errorf("uv tool upgrade %s failed and cache clear failed: %s", server.Package, out)
	}

	return nil
}
