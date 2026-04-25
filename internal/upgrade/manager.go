package upgrade

import (
	"fmt"

	"github.com/GeiserX/mcp-upgrade/internal/model"
	"github.com/fatih/color"
)

// Manager orchestrates upgrades across all detected MCP servers.
type Manager struct {
	DryRun  bool
	Verbose bool
}

// NewManager creates a Manager with the given dry-run setting.
func NewManager(dryRun bool) *Manager {
	return &Manager{DryRun: dryRun}
}

// UpgradeServer upgrades a single server using the appropriate upgrader.
func (m *Manager) UpgradeServer(server *model.Server) error {
	upgrader := selectUpgrader(server)
	if upgrader == nil {
		return fmt.Errorf("no upgrader available for type %s", server.Type)
	}

	if !upgrader.CanUpgrade(server) {
		return fmt.Errorf("upgrader cannot handle server %s", server.Name)
	}

	return upgrader.Upgrade(server)
}

// UpgradeAll processes all servers: skipping those that are up-to-date or auto-latest,
// and upgrading the rest. Returns counts of upgraded, failed, and skipped servers.
func (m *Manager) UpgradeAll(servers []model.Server) (upgraded, failed, skipped int) {
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	for i := range servers {
		s := &servers[i]

		if shouldSkip(s.Status) {
			skipped++
			continue
		}

		upgrader := selectUpgrader(s)
		if upgrader == nil || !upgrader.CanUpgrade(s) {
			skipped++
			continue
		}

		if m.DryRun {
			_, _ = yellow.Printf("  [DRY-RUN] Would upgrade %s (%s)", s.Name, s.Type)
			if s.CurrentVersion != "" && s.LatestVersion != "" {
				_, _ = cyan.Printf(" %s -> %s", s.CurrentVersion, s.LatestVersion)
			}
			fmt.Println()
			upgraded++
			continue
		}

		fmt.Printf("  Upgrading %s ...", s.Name)
		if err := upgrader.Upgrade(s); err != nil {
			_, _ = red.Printf(" FAILED: %s\n", err)
			failed++
		} else {
			_, _ = green.Println(" OK")
			upgraded++
		}
	}

	return upgraded, failed, skipped
}

func shouldSkip(status model.UpgradeStatus) bool {
	switch status {
	case model.StatusUpToDate, model.StatusAutoLatest, model.StatusSkipped:
		return true
	default:
		return false
	}
}

func selectUpgrader(server *model.Server) Upgrader {
	switch server.Type {
	case model.TypeNPX:
		return &NPXUpgrader{}
	case model.TypePipx:
		return &PipxUpgrader{}
	case model.TypeUVX:
		return &UVXUpgrader{}
	case model.TypeDocker:
		return &DockerUpgrader{}
	case model.TypeGitHubRelease:
		return &GitHubReleaseUpgrader{}
	case model.TypeGoBinary:
		return &GoBinaryUpgrader{}
	default:
		return nil
	}
}
