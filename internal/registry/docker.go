package registry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// DockerChecker checks versions for Docker-based MCP servers.
type DockerChecker struct {
	baseURL string
}

func NewDockerChecker() *DockerChecker {
	return &DockerChecker{baseURL: "https://registry.hub.docker.com"}
}

type dockerTagResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func (c *DockerChecker) GetLatestVersion(image string) (string, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/v2/repositories/%s/tags/?ordering=last_updated&page_size=1", c.baseURL, image))
	if err != nil {
		return "", fmt.Errorf("docker hub request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "latest", nil
	}

	var result dockerTagResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "latest", nil
	}

	if len(result.Results) == 0 {
		return "latest", nil
	}

	return result.Results[0].Name, nil
}

func (c *DockerChecker) GetCurrentVersion(server *model.Server) (string, error) {
	return "(local)", nil
}
