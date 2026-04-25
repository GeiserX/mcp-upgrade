package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// NPMChecker checks versions against the npm registry.
type NPMChecker struct {
	baseURL string
}

func NewNPMChecker() *NPMChecker {
	return &NPMChecker{baseURL: "https://registry.npmjs.org"}
}

type npmLatestResponse struct {
	Version string `json:"version"`
}

func (c *NPMChecker) GetLatestVersion(pkg string) (string, error) {
	encoded := pkg
	if strings.HasPrefix(pkg, "@") {
		parts := strings.SplitN(pkg, "/", 2)
		if len(parts) == 2 {
			encoded = parts[0] + "%2F" + parts[1]
		}
	}

	resp, err := httpClient.Get(fmt.Sprintf("%s/%s/latest", c.baseURL, encoded))
	if err != nil {
		return "", fmt.Errorf("npm registry request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned %d for %s", resp.StatusCode, pkg)
	}

	var result npmLatestResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse npm response: %w", err)
	}

	return result.Version, nil
}

func (c *NPMChecker) GetCurrentVersion(server *model.Server) (string, error) {
	if server.Type == model.TypeNPX {
		return "(auto)", nil
	}

	out, err := exec.Command("npm", "list", "-g", server.Package, "--json").Output()
	if err != nil {
		return "", fmt.Errorf("npm list failed: %w", err)
	}

	var result struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("failed to parse npm list output: %w", err)
	}

	dep, ok := result.Dependencies[server.Package]
	if !ok {
		return "", fmt.Errorf("package %s not found in global npm packages", server.Package)
	}

	return dep.Version, nil
}
