package upgrade

import "github.com/GeiserX/mcp-upgrade/internal/model"

// Target wraps a Server pointer for upgrade operations.
type Target struct {
	Server *model.Server
}

// Upgrader defines the interface for package-manager-specific upgrade strategies.
type Upgrader interface {
	Upgrade(server *model.Server) error
	CanUpgrade(server *model.Server) bool
}
