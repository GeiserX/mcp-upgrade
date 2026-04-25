package upgrade

import (
	"fmt"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// DockerUpgrader pulls the latest Docker image for Docker-based MCP servers.
type DockerUpgrader struct{}

func (u *DockerUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypeDocker && server.DockerImage != ""
}

func (u *DockerUpgrader) Upgrade(server *model.Server) error {
	// Verify Docker daemon is available.
	if _, err := execCommand("docker", "info"); err != nil {
		return fmt.Errorf("docker daemon is not available: run 'docker info' to check")
	}

	out, err := execCommand("docker", "pull", server.DockerImage)
	if err != nil {
		return fmt.Errorf("docker pull %s failed: %s", server.DockerImage, out)
	}
	return nil
}
