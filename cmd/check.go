package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/GeiserX/mcp-upgrade/internal/config"
	"github.com/GeiserX/mcp-upgrade/internal/detect"
	"github.com/GeiserX/mcp-upgrade/internal/model"
	"github.com/GeiserX/mcp-upgrade/internal/registry"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check all MCP servers for available updates",
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	scanner := config.NewScanner()
	result, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	detect.DetectAll(result.Servers)
	detect.ResolveMetadata(result.Servers)
	_ = registry.CheckAll(result.Servers)

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result.Servers)
	}

	printCheckTable(result.Servers)
	return nil
}

func printCheckTable(servers []model.Server) {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	blue := color.New(color.FgBlue, color.Bold)
	cyan := color.New(color.FgCyan)

	_, _ = blue.Printf("\n  MCP Server Status\n\n")

	table := tablewriter.NewTable(os.Stdout)
	table.Header("Server", "Type", "Current", "Latest", "Status")

	var upgradable, auto, skipped, upToDate, errored int

	for _, s := range servers {
		cur := s.CurrentVersion
		if cur == "" {
			cur = "-"
		}
		lat := s.LatestVersion
		if lat == "" {
			lat = "-"
		}

		var statusStr string
		switch s.Status {
		case model.StatusUpgradable:
			statusStr = yellow.Sprint("UPDATE")
			upgradable++
		case model.StatusUpToDate:
			statusStr = green.Sprint("UP-TO-DATE")
			upToDate++
		case model.StatusAutoLatest:
			statusStr = cyan.Sprint("AUTO")
			auto++
		case model.StatusSkipped:
			statusStr = "SKIP"
			skipped++
		case model.StatusError:
			statusStr = red.Sprint("ERROR")
			errored++
		default:
			statusStr = "UNKNOWN"
		}

		_ = table.Append(s.Name, string(s.Type), cur, lat, statusStr)
	}

	_ = table.Render()

	fmt.Printf("\n  %s upgradable, %s up-to-date, %s auto-latest, %s skipped",
		yellow.Sprintf("%d", upgradable),
		green.Sprintf("%d", upToDate),
		cyan.Sprintf("%d", auto),
		fmt.Sprintf("%d", skipped),
	)
	if errored > 0 {
		fmt.Printf(", %s errored", red.Sprintf("%d", errored))
	}
	fmt.Println()
	fmt.Println()
}
