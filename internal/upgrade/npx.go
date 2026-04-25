package upgrade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// NPXUpgrader clears the npx cache so the next invocation resolves the latest version.
type NPXUpgrader struct{}

func (u *NPXUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypeNPX && server.Package != ""
}

func (u *NPXUpgrader) Upgrade(server *model.Server) error {
	// Try the built-in cache clear first.
	if _, err := execCommand("npx", "clear-npx-cache"); err == nil {
		return nil
	}

	// Fall back to removing the package entry from the npx cache directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".npm", "_npx")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		// Cache directory doesn't exist — nothing to clear.
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgPath := filepath.Join(cacheDir, entry.Name(), "node_modules", server.Package)
		if _, err := os.Stat(pkgPath); err == nil {
			if err := os.RemoveAll(filepath.Join(cacheDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to clear npx cache for %s: %w", server.Package, err)
			}
		}
	}

	// npx always resolves latest on next run, so this always succeeds.
	return nil
}
