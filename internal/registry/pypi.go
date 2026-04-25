package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// PyPIChecker checks versions against the PyPI registry.
type PyPIChecker struct {
	baseURL string
}

func NewPyPIChecker() *PyPIChecker {
	return &PyPIChecker{baseURL: "https://pypi.org"}
}

type pypiResponse struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func (c *PyPIChecker) GetLatestVersion(pkg string) (string, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/pypi/%s/json", c.baseURL, pkg))
	if err != nil {
		return "", fmt.Errorf("pypi request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pypi returned %d for %s", resp.StatusCode, pkg)
	}

	var result pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse pypi response: %w", err)
	}

	return result.Info.Version, nil
}

var versionRegex = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)

func (c *PyPIChecker) GetCurrentVersion(server *model.Server) (string, error) {
	if server.Type == model.TypeUVX {
		return "(auto)", nil
	}

	// Try pipx list --json first
	out, err := exec.Command("pipx", "list", "--json").Output()
	if err == nil {
		ver, parseErr := parsePipxJSON(out, server.Package)
		if parseErr == nil {
			return ver, nil
		}
	}

	// Fallback: run the binary with --version
	binary := server.Command
	if binary == "" {
		binary = server.Package
	}
	out, err = exec.Command(binary, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("could not determine version for %s: %w", server.Package, err)
	}

	match := versionRegex.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return strings.TrimSpace(string(out)), nil
	}
	return match[1], nil
}

func parsePipxJSON(data []byte, pkg string) (string, error) {
	var result struct {
		Venvs map[string]struct {
			Metadata struct {
				MainPackage struct {
					PackageVersion string `json:"package_version"`
				} `json:"main_package"`
			} `json:"metadata"`
		} `json:"venvs"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	venv, ok := result.Venvs[pkg]
	if !ok {
		return "", fmt.Errorf("package %s not found in pipx list", pkg)
	}

	return venv.Metadata.MainPackage.PackageVersion, nil
}
