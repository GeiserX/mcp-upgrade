package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// GitHubChecker checks versions against GitHub releases.
type GitHubChecker struct {
	baseURL string
}

func NewGitHubChecker() *GitHubChecker {
	return &GitHubChecker{baseURL: "https://api.github.com"}
}

type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func (c *GitHubChecker) GetLatestVersion(repo string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/releases/latest", c.baseURL, repo), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create github request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if n, err := strconv.Atoi(remaining); err == nil && n == 0 {
			return "", fmt.Errorf("github api rate limit exceeded")
		}
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		// Retry without auth for public repos
		if os.Getenv("GITHUB_TOKEN") != "" {
			return c.getLatestVersionNoAuth(repo)
		}
		return "", fmt.Errorf("github api returned %d for %s", resp.StatusCode, repo)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned %d for %s", resp.StatusCode, repo)
	}

	var result githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse github response: %w", err)
	}

	return result.TagName, nil
}

func (c *GitHubChecker) getLatestVersionNoAuth(repo string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/releases/latest", c.baseURL, repo), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create github request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned %d for %s (no auth)", resp.StatusCode, repo)
	}

	var result githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse github response: %w", err)
	}

	return result.TagName, nil
}

var ghVersionRegex = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)

func (c *GitHubChecker) GetCurrentVersion(server *model.Server) (string, error) {
	binary := server.BinaryPath
	if binary == "" {
		binary = server.Command
	}
	if binary == "" {
		parts := strings.Split(server.GitHubRepo, "/")
		if len(parts) == 2 {
			binary = parts[1]
		}
	}
	if binary == "" {
		return "", fmt.Errorf("no binary path for github release server %s", server.Name)
	}

	out, err := exec.Command(binary, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run %s --version: %w", binary, err)
	}

	match := ghVersionRegex.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return strings.TrimSpace(string(out)), nil
	}
	return match[1], nil
}
