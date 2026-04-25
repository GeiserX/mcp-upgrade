package upgrade

import (
	"fmt"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// GoBinaryUpgrader handles Go-installed MCP server binaries.
// If a GitHub repo is known, it delegates to GitHubReleaseUpgrader.
// Otherwise it attempts go install @latest.
type GoBinaryUpgrader struct{}

func (u *GoBinaryUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypeGoBinary
}

func (u *GoBinaryUpgrader) Upgrade(server *model.Server) error {
	// If we know the GitHub repo, try a release download first.
	if server.GitHubRepo != "" && server.BinaryPath != "" {
		gh := &GitHubReleaseUpgrader{}
		if err := gh.Upgrade(server); err == nil {
			return nil
		}
		// Fall through to go install on release failure.
	}

	// Derive the module path from GitHubRepo or Package.
	module := server.Package
	if module == "" && server.GitHubRepo != "" {
		module = "github.com/" + server.GitHubRepo
	}
	if module == "" {
		return fmt.Errorf("no package or GitHub repo known for go binary %s", server.Name)
	}

	out, err := execCommand("go", "install", module+"@latest")
	if err != nil {
		return fmt.Errorf("go install %s@latest failed: %s", module, out)
	}
	return nil
}
