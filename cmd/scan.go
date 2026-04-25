package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/GeiserX/mcp-upgrade/internal/config"
	"github.com/GeiserX/mcp-upgrade/internal/detect"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Auto-discover MCP servers from all supported client configs",
	RunE:  runScan,
}

func runScan(cmd *cobra.Command, args []string) error {
	scanner := config.NewScanner()
	result, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	detect.DetectAll(result.Servers)
	detect.ResolveMetadata(result.Servers)

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	blue := color.New(color.FgBlue, color.Bold)
	blue.Printf("\nFound %d client(s), %d server(s)\n\n", len(result.Clients), len(result.Servers))

	if len(result.Clients) > 0 {
		fmt.Println("Clients:")
		for _, c := range result.Clients {
			fmt.Printf("  %s (%s)\n", c.Name, c.ConfigPath)
		}
		fmt.Println()
	}

	if len(result.Servers) == 0 {
		fmt.Println("No MCP servers found.")
		return nil
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("Server", "Type", "Package", "Client")

	for _, s := range result.Servers {
		pkg := s.Package
		if pkg == "" {
			pkg = "-"
		}
		table.Append(s.Name, string(s.Type), pkg, s.Client)
	}

	table.Render()
	return nil
}
