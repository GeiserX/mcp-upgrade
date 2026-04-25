package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

func TestDetectType(t *testing.T) {
	// Create a temporary executable file for binary detection tests.
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "my-mcp-server")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	goBinPath := filepath.Join(tmpDir, "go", "bin", "mcp-tool")
	if err := os.MkdirAll(filepath.Dir(goBinPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goBinPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		server      model.Server
		wantType    model.ServerType
		wantPkg     string
		wantBinary  string
		wantDocker  string
	}{
		{
			name: "npx with scoped package",
			server: model.Server{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			},
			wantType: model.TypeNPX,
			wantPkg:  "@modelcontextprotocol/server-github",
		},
		{
			name: "npx with unscoped package",
			server: model.Server{
				Command: "npx",
				Args:    []string{"-y", "firecrawl-mcp"},
			},
			wantType: model.TypeNPX,
			wantPkg:  "firecrawl-mcp",
		},
		{
			name: "npx with @latest suffix",
			server: model.Server{
				Command: "npx",
				Args:    []string{"-y", "@upstash/context7-mcp@latest"},
			},
			wantType: model.TypeNPX,
			wantPkg:  "@upstash/context7-mcp",
		},
		{
			name: "npx with --yes flag",
			server: model.Server{
				Command: "npx",
				Args:    []string{"--yes", "some-package@1.2.3"},
			},
			wantType: model.TypeNPX,
			wantPkg:  "some-package",
		},
		{
			name: "npx with scoped package and version",
			server: model.Server{
				Command: "npx",
				Args:    []string{"-y", "@org/pkg@2.0.0"},
			},
			wantType: model.TypeNPX,
			wantPkg:  "@org/pkg",
		},
		{
			name: "node command",
			server: model.Server{
				Command: "node",
				Args:    []string{"/path/to/server.js"},
			},
			wantType: model.TypeLocalNode,
		},
		{
			name: "uvx package",
			server: model.Server{
				Command: "uvx",
				Args:    []string{"mcp-server-fetch"},
			},
			wantType: model.TypeUVX,
			wantPkg:  "mcp-server-fetch",
		},
		{
			name: "uvx with flags before package",
			server: model.Server{
				Command: "uvx",
				Args:    []string{"--quiet", "mcp-server-fetch"},
			},
			wantType: model.TypeUVX,
			wantPkg:  "mcp-server-fetch",
		},
		{
			name: "python -m module",
			server: model.Server{
				Command: "python3",
				Args:    []string{"-m", "mcp_server_time"},
			},
			wantType: model.TypePipx,
			wantPkg:  "mcp_server_time",
		},
		{
			name: "python without -m",
			server: model.Server{
				Command: "python",
				Args:    []string{"server.py"},
			},
			wantType: model.TypeUnknown,
		},
		{
			name: "docker run with flags",
			server: model.Server{
				Command: "docker",
				Args:    []string{"run", "-i", "--rm", "-e", "API_KEY=xxx", "mcp/fetch:latest"},
			},
			wantType:   model.TypeDocker,
			wantDocker: "mcp/fetch:latest",
		},
		{
			name: "docker run image as last arg",
			server: model.Server{
				Command: "docker",
				Args:    []string{"run", "--rm", "ghcr.io/org/image:v1"},
			},
			wantType:   model.TypeDocker,
			wantDocker: "ghcr.io/org/image:v1",
		},
		{
			name: "docker without run",
			server: model.Server{
				Command: "docker",
				Args:    []string{"build", "."},
			},
			wantType: model.TypeUnknown,
		},
		{
			name: "go binary path",
			server: model.Server{
				Command: goBinPath,
			},
			wantType:   model.TypeGoBinary,
			wantBinary: goBinPath,
		},
		{
			name: "plain executable binary",
			server: model.Server{
				Command: binPath,
			},
			wantType:   model.TypeGitHubRelease,
			wantBinary: binPath,
		},
		{
			name: "cargo command",
			server: model.Server{
				Command: "cargo",
				Args:    []string{"run", "--release"},
			},
			wantType: model.TypeCargo,
		},
		{
			name: "unknown command",
			server: model.Server{
				Command: "some-random-thing",
				Args:    []string{"--foo"},
			},
			wantType: model.TypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.server
			DetectType(&s)

			if s.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", s.Type, tt.wantType)
			}
			if tt.wantPkg != "" && s.Package != tt.wantPkg {
				t.Errorf("Package = %q, want %q", s.Package, tt.wantPkg)
			}
			if tt.wantBinary != "" && s.BinaryPath != tt.wantBinary {
				t.Errorf("BinaryPath = %q, want %q", s.BinaryPath, tt.wantBinary)
			}
			if tt.wantDocker != "" && s.DockerImage != tt.wantDocker {
				t.Errorf("DockerImage = %q, want %q", s.DockerImage, tt.wantDocker)
			}
		})
	}
}

func TestDetectAll(t *testing.T) {
	servers := []model.Server{
		{Command: "npx", Args: []string{"-y", "some-pkg"}},
		{Command: "uvx", Args: []string{"another-pkg"}},
		{Command: "unknown-cmd"},
	}

	DetectAll(servers)

	if servers[0].Type != model.TypeNPX {
		t.Errorf("servers[0].Type = %q, want %q", servers[0].Type, model.TypeNPX)
	}
	if servers[1].Type != model.TypeUVX {
		t.Errorf("servers[1].Type = %q, want %q", servers[1].Type, model.TypeUVX)
	}
	if servers[2].Type != model.TypeUnknown {
		t.Errorf("servers[2].Type = %q, want %q", servers[2].Type, model.TypeUnknown)
	}
}
