package upgrade

import (
	"testing"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

func TestFindAsset(t *testing.T) {
	assets := []string{
		"tool-Darwin_arm64.tar.gz",
		"tool-Darwin_amd64.tar.gz",
		"tool-linux_amd64.tar.gz",
		"tool-linux_arm64.tar.gz",
		"tool-windows_amd64.zip",
		"tool-Darwin_arm64.tar.gz.sha256",
	}

	tests := []struct {
		name     string
		os       string
		arch     string
		expected string
	}{
		{"darwin arm64", "darwin", "arm64", "tool-Darwin_arm64.tar.gz"},
		{"darwin amd64", "darwin", "amd64", "tool-Darwin_amd64.tar.gz"},
		{"linux amd64", "linux", "amd64", "tool-linux_amd64.tar.gz"},
		{"linux arm64", "linux", "arm64", "tool-linux_arm64.tar.gz"},
		{"windows amd64", "windows", "amd64", "tool-windows_amd64.zip"},
		{"no match", "freebsd", "arm64", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAsset(assets, tt.os, tt.arch)
			if got != tt.expected {
				t.Errorf("findAsset(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.expected)
			}
		})
	}
}

func TestFindAssetAlternateNaming(t *testing.T) {
	assets := []string{
		"server-macos-aarch64.tar.gz",
		"server-linux-x86_64.tar.gz",
		"server-linux-x86_64.tar.gz.sig",
	}

	tests := []struct {
		name     string
		os       string
		arch     string
		expected string
	}{
		{"macos aarch64 alias", "darwin", "arm64", "server-macos-aarch64.tar.gz"},
		{"linux x86_64 alias", "linux", "amd64", "server-linux-x86_64.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAsset(assets, tt.os, tt.arch)
			if got != tt.expected {
				t.Errorf("findAsset(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.expected)
			}
		})
	}
}

func TestFindAssetSkipsChecksums(t *testing.T) {
	assets := []string{
		"tool-Darwin_arm64.tar.gz.sha256",
		"tool-Darwin_arm64.tar.gz.sha512",
		"tool-Darwin_arm64.tar.gz.sig",
		"tool-Darwin_arm64.tar.gz.asc",
	}

	got := findAsset(assets, "darwin", "arm64")
	if got != "" {
		t.Errorf("expected no match for checksum-only assets, got %q", got)
	}
}

func TestSelectUpgrader(t *testing.T) {
	tests := []struct {
		serverType model.ServerType
		expectNil  bool
		expectType string
	}{
		{model.TypeNPX, false, "*upgrade.NPXUpgrader"},
		{model.TypePipx, false, "*upgrade.PipxUpgrader"},
		{model.TypeUVX, false, "*upgrade.UVXUpgrader"},
		{model.TypeDocker, false, "*upgrade.DockerUpgrader"},
		{model.TypeGitHubRelease, false, "*upgrade.GitHubReleaseUpgrader"},
		{model.TypeGoBinary, false, "*upgrade.GoBinaryUpgrader"},
		{model.TypeLocal, true, ""},
		{model.TypeUnknown, true, ""},
		{model.TypeCargo, true, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.serverType), func(t *testing.T) {
			s := &model.Server{Type: tt.serverType}
			got := selectUpgrader(s)
			if tt.expectNil && got != nil {
				t.Errorf("expected nil upgrader for %s, got %T", tt.serverType, got)
			}
			if !tt.expectNil && got == nil {
				t.Errorf("expected upgrader for %s, got nil", tt.serverType)
			}
		})
	}
}

func TestManagerDryRunDoesNotExecute(t *testing.T) {
	mgr := NewManager(true)

	servers := []model.Server{
		{
			Name:           "test-npx",
			Type:           model.TypeNPX,
			Package:        "some-package",
			Status:         model.StatusUpgradable,
			CurrentVersion: "1.0.0",
			LatestVersion:  "2.0.0",
		},
		{
			Name:    "test-docker",
			Type:    model.TypeDocker,
			Status:  model.StatusUpgradable,
			DockerImage: "some/image:latest",
		},
	}

	upgraded, failed, skipped := mgr.UpgradeAll(servers)

	if failed != 0 {
		t.Errorf("dry run should not fail, got %d failures", failed)
	}
	if upgraded != 2 {
		t.Errorf("expected 2 upgraded (dry-run), got %d", upgraded)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
}

func TestManagerSkipsNonUpgradable(t *testing.T) {
	mgr := NewManager(true)

	servers := []model.Server{
		{Name: "up-to-date", Type: model.TypeNPX, Package: "p", Status: model.StatusUpToDate},
		{Name: "auto", Type: model.TypeNPX, Package: "p", Status: model.StatusAutoLatest},
		{Name: "skipped", Type: model.TypeNPX, Package: "p", Status: model.StatusSkipped},
		{Name: "upgradable", Type: model.TypeNPX, Package: "p", Status: model.StatusUpgradable},
	}

	upgraded, _, skipped := mgr.UpgradeAll(servers)

	if skipped != 3 {
		t.Errorf("expected 3 skipped, got %d", skipped)
	}
	if upgraded != 1 {
		t.Errorf("expected 1 upgraded (dry-run), got %d", upgraded)
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		status model.UpgradeStatus
		skip   bool
	}{
		{model.StatusUpToDate, true},
		{model.StatusAutoLatest, true},
		{model.StatusSkipped, true},
		{model.StatusUpgradable, false},
		{model.StatusUnknown, false},
		{model.StatusError, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := shouldSkip(tt.status); got != tt.skip {
				t.Errorf("shouldSkip(%s) = %v, want %v", tt.status, got, tt.skip)
			}
		})
	}
}
