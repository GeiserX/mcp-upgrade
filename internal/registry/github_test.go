package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubChecker_GetLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/mark3labs/mcp-go/releases/latest":
			_ = json.NewEncoder(w).Encode(githubReleaseResponse{TagName: "v0.26.0"})
		case "/repos/ratelimited/repo/releases/latest":
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusForbidden)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	checker := &GitHubChecker{baseURL: srv.URL}

	tests := []struct {
		name    string
		repo    string
		want    string
		wantErr bool
	}{
		{
			name: "valid repo",
			repo: "mark3labs/mcp-go",
			want: "v0.26.0",
		},
		{
			name:    "rate limited",
			repo:    "ratelimited/repo",
			wantErr: true,
		},
		{
			name:    "not found",
			repo:    "nonexistent/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.GetLatestVersion(tt.repo)
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

func TestGitHubVersionRegex(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"mcp-go version v0.26.0", "0.26.0"},
		{"tool 2.0.1-beta", "2.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			match := ghVersionRegex.FindStringSubmatch(tt.input)
			if len(match) < 2 {
				t.Fatalf("no match for %q", tt.input)
			}
			if match[1] != tt.want {
				t.Errorf("got %q, want %q", match[1], tt.want)
			}
		})
	}
}
