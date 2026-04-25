package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GeiserX/mcp-upgrade/internal/model"
)

func TestNPMChecker_GetLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Path is the decoded path; RawPath preserves encoding
		path := r.URL.RawPath
		if path == "" {
			path = r.URL.Path
		}
		switch path {
		case "/@modelcontextprotocol%2Fserver-filesystem/latest":
			json.NewEncoder(w).Encode(npmLatestResponse{Version: "2.1.0"})
		case "/typescript/latest":
			json.NewEncoder(w).Encode(npmLatestResponse{Version: "5.8.3"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	checker := &NPMChecker{baseURL: srv.URL}

	tests := []struct {
		name    string
		pkg     string
		want    string
		wantErr bool
	}{
		{
			name: "simple package",
			pkg:  "typescript",
			want: "5.8.3",
		},
		{
			name: "scoped package",
			pkg:  "@modelcontextprotocol/server-filesystem",
			want: "2.1.0",
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

func TestNPMChecker_GetCurrentVersion_NPX(t *testing.T) {
	checker := NewNPMChecker()
	server := &model.Server{Type: model.TypeNPX, Package: "typescript"}
	got, err := checker.GetCurrentVersion(server)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "(auto)" {
		t.Errorf("got %q, want %q", got, "(auto)")
	}
}
