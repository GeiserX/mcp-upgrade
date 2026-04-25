package upgrade

import (
	"fmt"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// PipxUpgrader runs pipx upgrade for Python-based MCP servers.
type PipxUpgrader struct{}

func (u *PipxUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypePipx && server.Package != ""
}

func (u *PipxUpgrader) Upgrade(server *model.Server) error {
	out, err := execCommand("pipx", "upgrade", server.Package)
	if err != nil {
		return fmt.Errorf("pipx upgrade %s failed: %s", server.Package, out)
	}
	return nil
}
