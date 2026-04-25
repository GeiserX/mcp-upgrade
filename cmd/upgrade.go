package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/GeiserX/mcp-upgrade/internal/config"
	"github.com/GeiserX/mcp-upgrade/internal/detect"
	"github.com/GeiserX/mcp-upgrade/internal/registry"
	"github.com/GeiserX/mcp-upgrade/internal/upgrade"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	yesFlag bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [server-name...]",
	Short: "Upgrade all or specific MCP servers",
	Long:  "Upgrade MCP servers to their latest versions. Specify server names to upgrade specific ones, or omit to upgrade all.",
	RunE:  runUpgrade,
}

func init() {
	upgradeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be upgraded without making changes")
	upgradeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	scanner := config.NewScanner()
	result, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	detect.DetectAll(result.Servers)
	detect.ResolveMetadata(result.Servers)
	_ = registry.CheckAll(result.Servers)

	targets := result.Servers
	if len(args) > 0 {
		nameSet := make(map[string]bool)
		for _, a := range args {
			nameSet[a] = true
		}
		var filtered []upgrade.Target
		for i := range targets {
			if nameSet[targets[i].Name] {
				filtered = append(filtered, upgrade.Target{Server: &targets[i]})
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no matching servers found for: %v", args)
		}
	}

	if jsonOut && dryRun {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Servers)
	}

	printCheckTable(result.Servers)

	mgr := upgrade.NewManager(dryRun)
	upgraded, failed, skipped := mgr.UpgradeAll(result.Servers)

	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed)

	if dryRun {
		fmt.Printf("\n  Dry run complete. Use 'mcp-upgrade upgrade' to apply changes.\n\n")
	} else {
		if upgraded > 0 {
			_, _ = green.Printf("\n  %d server(s) upgraded.", upgraded)
		}
		if failed > 0 {
			_, _ = red.Printf(" %d failed.", failed)
		}
		if skipped > 0 {
			fmt.Printf(" %d skipped.", skipped)
		}
		if upgraded > 0 {
			fmt.Printf("\n  Restart your AI coding client to pick up updated servers.")
		}
		fmt.Println()
		fmt.Println()
	}

	if failed > 0 {
		return fmt.Errorf("%d server(s) failed to upgrade", failed)
	}
	return nil
}
