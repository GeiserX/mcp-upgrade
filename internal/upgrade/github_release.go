package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

// GitHubReleaseUpgrader downloads the latest release binary from GitHub.
type GitHubReleaseUpgrader struct{}

func (u *GitHubReleaseUpgrader) CanUpgrade(server *model.Server) bool {
	return server.Type == model.TypeGitHubRelease &&
		server.GitHubRepo != "" &&
		server.BinaryPath != ""
}

func (u *GitHubReleaseUpgrader) Upgrade(server *model.Server) error {
	assets, err := fetchReleaseAssets(server.GitHubRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases for %s: %w", server.GitHubRepo, err)
	}

	assetName := findAsset(assets, runtime.GOOS, runtime.GOARCH)
	if assetName == "" {
		// No matching binary — try building from source if it's a Go project.
		out, buildErr := execCommand("go", "install", "github.com/"+server.GitHubRepo+"@latest")
		if buildErr != nil {
			return fmt.Errorf("no matching release asset for %s/%s and go install failed: %s",
				runtime.GOOS, runtime.GOARCH, out)
		}
		return nil
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s",
		server.GitHubRepo, assetName)

	tmpDir, err := os.MkdirTemp("", "mcp-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dlPath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, dlPath); err != nil {
		return fmt.Errorf("failed to download %s: %w", assetName, err)
	}

	// Determine the final binary path.
	binaryPath := server.BinaryPath
	lower := strings.ToLower(assetName)

	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") || strings.HasSuffix(lower, ".zip") {
		extractDir := filepath.Join(tmpDir, "extracted")
		if err := extractArchive(dlPath, extractDir); err != nil {
			return fmt.Errorf("failed to extract %s: %w", assetName, err)
		}

		bin, err := findExtractedBinary(extractDir, filepath.Base(binaryPath))
		if err != nil {
			return fmt.Errorf("could not find binary in archive: %w", err)
		}
		dlPath = bin
	}

	// Replace the old binary.
	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old binary %s: %w", binaryPath, err)
	}

	if err := copyFile(dlPath, binaryPath); err != nil {
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", binaryPath, err)
	}

	return nil
}

type ghAsset struct {
	Name string `json:"name"`
}

type ghRelease struct {
	Assets []ghAsset `json:"assets"`
}

func fetchReleaseAssets(repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	names := make([]string, len(release.Assets))
	for i, a := range release.Assets {
		names[i] = a.Name
	}
	return names, nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// findExtractedBinary walks the extracted directory looking for a binary
// matching the expected name.
func findExtractedBinary(dir, expectedName string) (string, error) {
	var found string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == expectedName || strings.TrimSuffix(base, ".exe") == strings.TrimSuffix(expectedName, ".exe") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("binary %q not found in extracted archive", expectedName)
	}
	return found, nil
}
