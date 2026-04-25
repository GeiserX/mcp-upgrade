package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "mcp-upgrade",
	Short: "Universal upgrade tool for MCP servers",
	Long:  "Auto-discover, check, and upgrade MCP servers across all AI coding clients and package ecosystems.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show mcp-upgrade version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-upgrade %s\n", version)
	},
}
