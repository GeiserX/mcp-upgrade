package main

import (
	"os"

	"github.com/GeiserX/mcp-upgrade/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
