package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPyPIChecker_GetLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pypi/mcp/json":
			json.NewEncoder(w).Encode(pypiResponse{
				Info: struct {
					Version string `json:"version"`
				}{Version: "1.9.2"},
			})
		case "/pypi/uvx/json":
			json.NewEncoder(w).Encode(pypiResponse{
				Info: struct {
					Version string `json:"version"`
				}{Version: "0.5.1"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	checker := &PyPIChecker{baseURL: srv.URL}

	tests := []struct {
		name    string
		pkg     string
		want    string
		wantErr bool
	}{
		{
			name: "mcp package",
			pkg:  "mcp",
			want: "1.9.2",
		},
		{
			name: "uvx package",
			pkg:  "uvx",
			want: "0.5.1",
		},
		{
			name:    "not found",
			pkg:     "nonexistent-pkg-xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.GetLatestVersion(tt.pkg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePipxJSON(t *testing.T) {
	data := []byte(`{
		"venvs": {
			"mcp-server-fetch": {
				"metadata": {
					"main_package": {
						"package_version": "0.6.2"
					}
				}
			}
		}
	}`)

	got, err := parsePipxJSON(data, "mcp-server-fetch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "0.6.2" {
		t.Errorf("got %q, want %q", got, "0.6.2")
	}

	_, err = parsePipxJSON(data, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}
