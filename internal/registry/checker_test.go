package registry

import (
	"testing"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

func TestCheckVersion_SkipsLocal(t *testing.T) {
	server := model.Server{Name: "local-tool", Type: model.TypeLocal}
	err := CheckVersion(&server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Status != model.StatusSkipped {
		t.Errorf("got status %q, want %q", server.Status, model.StatusSkipped)
	}
}

func TestCheckVersion_SkipsUnknown(t *testing.T) {
	server := model.Server{Name: "unknown-tool", Type: model.TypeUnknown}
	err := CheckVersion(&server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Status != model.StatusSkipped {
		t.Errorf("got status %q, want %q", server.Status, model.StatusSkipped)
	}
}

func TestCheckVersion_SkipsCargo(t *testing.T) {
	server := model.Server{Name: "cargo-tool", Type: model.TypeCargo}
	err := CheckVersion(&server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Status != model.StatusSkipped {
		t.Errorf("got status %q, want %q", server.Status, model.StatusSkipped)
	}
}

func TestCheckerForType(t *testing.T) {
	tests := []struct {
		serverType model.ServerType
		isNil      bool
	}{
		{model.TypeNPX, false},
		{model.TypeLocalNode, false},
		{model.TypePipx, false},
		{model.TypeUVX, false},
		{model.TypeGitHubRelease, false},
		{model.TypeGoBinary, false},
		{model.TypeDocker, false},
		{model.TypeLocal, true},
		{model.TypeUnknown, true},
		{model.TypeCargo, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.serverType), func(t *testing.T) {
			checker, _ := checkerForType(tt.serverType)
			if tt.isNil && checker != nil {
				t.Errorf("expected nil checker for %s", tt.serverType)
			}
			if !tt.isNil && checker == nil {
				t.Errorf("expected non-nil checker for %s", tt.serverType)
			}
		})
	}
}

func TestCheckAll_Empty(t *testing.T) {
	err := CheckAll(nil)
	if err != nil {
		t.Fatalf("unexpected error for empty slice: %v", err)
	}
}

func TestCheckAll_SkipsAll(t *testing.T) {
	servers := []model.Server{
		{Name: "a", Type: model.TypeLocal},
		{Name: "b", Type: model.TypeUnknown},
		{Name: "c", Type: model.TypeCargo},
	}
	err := CheckAll(servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range servers {
		if s.Status != model.StatusSkipped {
			t.Errorf("server %s: got status %q, want %q", s.Name, s.Status, model.StatusSkipped)
		}
	}
}

func TestPackageName(t *testing.T) {
	tests := []struct {
		name   string
		server model.Server
		want   string
	}{
		{
			name:   "npm package",
			server: model.Server{Type: model.TypeNPX, Package: "@org/pkg"},
			want:   "@org/pkg",
		},
		{
			name:   "github release",
			server: model.Server{Type: model.TypeGitHubRelease, GitHubRepo: "owner/repo"},
			want:   "owner/repo",
		},
		{
			name:   "docker image",
			server: model.Server{Type: model.TypeDocker, DockerImage: "mcp/server"},
			want:   "mcp/server",
		},
		{
			name:   "pypi package",
			server: model.Server{Type: model.TypePipx, Package: "mcp-server"},
			want:   "mcp-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := packageName(&tt.server)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
