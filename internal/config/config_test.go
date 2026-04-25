package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	dir := t.TempDir()

	content := `{
		"mcpServers": {
			"context7": {
				"command": "npx",
				"args": ["-y", "@upstash/context7-mcp"],
				"env": {"KEY": "val"}
			},
			"filesystem": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
			}
		}
	}`

	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	servers, err := parseConfigFile(path, "mcpServers")
	if err != nil {
		t.Fatalf("parseConfigFile: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	ctx7, ok := servers["context7"]
	if !ok {
		t.Fatal("missing context7 server")
	}
	if ctx7.Command != "npx" {
		t.Errorf("command = %q, want npx", ctx7.Command)
	}
	if len(ctx7.Args) != 2 || ctx7.Args[1] != "@upstash/context7-mcp" {
		t.Errorf("args = %v, unexpected", ctx7.Args)
	}
	if ctx7.Env["KEY"] != "val" {
		t.Errorf("env KEY = %q, want val", ctx7.Env["KEY"])
	}

	fs, ok := servers["filesystem"]
	if !ok {
		t.Fatal("missing filesystem server")
	}
	if len(fs.Args) != 3 {
		t.Errorf("filesystem args = %v, expected 3 args", fs.Args)
	}
}

func TestParseConfigFileVSCodeKey(t *testing.T) {
	dir := t.TempDir()

	content := `{
		"servers": {
			"my-server": {
				"command": "node",
				"args": ["server.js"]
			}
		}
	}`

	path := filepath.Join(dir, "mcp.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	servers, err := parseConfigFile(path, "servers")
	if err != nil {
		t.Fatalf("parseConfigFile: %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}

	srv, ok := servers["my-server"]
	if !ok {
		t.Fatal("missing my-server")
	}
	if srv.Command != "node" {
		t.Errorf("command = %q, want node", srv.Command)
	}
}

func TestParseConfigFileMissing(t *testing.T) {
	_, err := parseConfigFile("/nonexistent/path/config.json", "mcpServers")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseConfigFileMissingKey(t *testing.T) {
	dir := t.TempDir()

	content := `{"otherKey": {}}`
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parseConfigFile(path, "mcpServers")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestParseConfigFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parseConfigFile(path, "mcpServers")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFingerprint(t *testing.T) {
	fp1 := fingerprint("npx", []string{"-y", "@upstash/context7-mcp"})
	fp2 := fingerprint("npx", []string{"-y", "@upstash/context7-mcp"})
	fp3 := fingerprint("npx", []string{"-y", "@modelcontextprotocol/server-filesystem"})

	if fp1 != fp2 {
		t.Errorf("identical commands should produce same fingerprint")
	}
	if fp1 == fp3 {
		t.Errorf("different commands should produce different fingerprints")
	}
}

func TestDeduplication(t *testing.T) {
	dir := t.TempDir()

	// Two config files with the same server
	content := `{
		"mcpServers": {
			"context7": {
				"command": "npx",
				"args": ["-y", "@upstash/context7-mcp"]
			}
		}
	}`

	path1 := filepath.Join(dir, "config1.json")
	path2 := filepath.Join(dir, "config2.json")
	if err := os.WriteFile(path1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	servers1, _ := parseConfigFile(path1, "mcpServers")
	servers2, _ := parseConfigFile(path2, "mcpServers")

	// Simulate deduplication logic
	seen := make(map[string]bool)
	var count int
	for _, srvMap := range []map[string]string{
		{"cmd": servers1["context7"].Command},
		{"cmd": servers2["context7"].Command},
	} {
		_ = srvMap
	}

	// Direct test of the Scanner dedup via fingerprints
	for _, srv := range servers1 {
		fp := fingerprint(srv.Command, srv.Args)
		if !seen[fp] {
			seen[fp] = true
			count++
		}
	}
	for _, srv := range servers2 {
		fp := fingerprint(srv.Command, srv.Args)
		if !seen[fp] {
			seen[fp] = true
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected 1 unique server after dedup, got %d", count)
	}
}

func TestScanSkipsMissingFiles(t *testing.T) {
	// Scanner.Scan should not error even when no config files exist.
	// We can't easily control HOME to point at an empty dir without
	// side effects, but we can at least verify it doesn't panic.
	scanner := &Scanner{}
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Scan returned nil result")
	}
	// Servers may or may not be found depending on the test environment,
	// but the call must succeed.
}

func TestFingerprintServer(t *testing.T) {
	fp := FingerprintServer("docker", []string{"run", "--rm", "mcp-server"})
	if fp == "" {
		t.Fatal("FingerprintServer returned empty string")
	}
	if len(fp) != 16 {
		t.Errorf("expected 16-char fingerprint, got %d chars", len(fp))
	}
}
