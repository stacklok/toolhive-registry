package serverjson

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stacklok/toolhive/pkg/registry/registry"
)

const statusActive = "Active"

const testPackageServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-server",
  "description": "A test server",
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/test/server:v1.0.0",
      "transport": { "type": "stdio" }
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/test/server:v1.0.0": {
          "status": "Active",
          "tier": "Official",
          "tools": ["tool_a", "tool_b"],
          "tags": ["test"],
          "metadata": {
            "stars": 100,
            "last_updated": "2026-01-01T00:00:00Z"
          }
        }
      }
    }
  }
}
`

const testRemoteServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-remote",
  "description": "A test remote server",
  "version": "1.0.0",
  "remotes": [
    {
      "type": "sse",
      "url": "https://api.example.com/mcp"
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://api.example.com/mcp": {
          "status": "Active",
          "tier": "Community",
          "tools": ["remote_tool"],
          "metadata": {
            "last_updated": "2026-01-01T00:00:00Z"
          }
        }
      }
    }
  }
}
`

const testNoMetaServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/no-meta",
  "description": "Server with no _meta",
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/test/no-meta:v1.0.0",
      "transport": { "type": "stdio" }
    }
  ]
}
`

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadServerFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testPackageServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatalf("LoadServerFile failed: %v", err)
	}

	if sf.ServerJSON.Name != "io.github.stacklok/test-server" {
		t.Errorf("expected name io.github.stacklok/test-server, got %s", sf.ServerJSON.Name)
	}
	if sf.Path != path {
		t.Errorf("expected path %s, got %s", path, sf.Path)
	}
}

func TestLoadServerFile_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", "{invalid json")

	_, err := LoadServerFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadServerFile_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := LoadServerFile("/nonexistent/server.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtensionKey_Package(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testPackageServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	key, err := sf.ExtensionKey()
	if err != nil {
		t.Fatalf("ExtensionKey failed: %v", err)
	}
	if key != "ghcr.io/test/server:v1.0.0" {
		t.Errorf("expected ghcr.io/test/server:v1.0.0, got %s", key)
	}
}

func TestExtensionKey_Remote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testRemoteServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	key, err := sf.ExtensionKey()
	if err != nil {
		t.Fatalf("ExtensionKey failed: %v", err)
	}
	if key != "https://api.example.com/mcp" {
		t.Errorf("expected https://api.example.com/mcp, got %s", key)
	}
}

func TestIsPackageServer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	pkgPath := writeTestFile(t, dir, "pkg.json", testPackageServerJSON)
	remotePath := writeTestFile(t, dir, "remote.json", testRemoteServerJSON)

	pkgSF, _ := LoadServerFile(pkgPath)
	remoteSF, _ := LoadServerFile(remotePath)

	if !pkgSF.IsPackageServer() {
		t.Error("expected package server to be true")
	}
	if pkgSF.IsRemoteServer() {
		t.Error("expected remote server to be false for package server")
	}
	if !remoteSF.IsRemoteServer() {
		t.Error("expected remote server to be true")
	}
	if remoteSF.IsPackageServer() {
		t.Error("expected package server to be false for remote server")
	}
}

func TestGetExtensions_Package(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testPackageServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatalf("GetExtensions failed: %v", err)
	}

	if ext.Status != statusActive {
		t.Errorf("expected status Active, got %s", ext.Status)
	}
	if ext.Tier != "Official" {
		t.Errorf("expected tier Official, got %s", ext.Tier)
	}
	if len(ext.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(ext.Tools))
	}
	if ext.Metadata == nil || ext.Metadata.Stars != 100 {
		t.Errorf("expected 100 stars, got %v", ext.Metadata)
	}
}

func TestGetExtensions_Remote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testRemoteServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatalf("GetExtensions failed: %v", err)
	}

	if ext.Status != statusActive {
		t.Errorf("expected status Active, got %s", ext.Status)
	}
	if ext.Tier != "Community" {
		t.Errorf("expected tier Community, got %s", ext.Tier)
	}
	if len(ext.Tools) != 1 || ext.Tools[0] != "remote_tool" {
		t.Errorf("expected [remote_tool], got %v", ext.Tools)
	}
}

func TestGetExtensions_NoMeta(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testNoMetaServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatalf("GetExtensions failed: %v", err)
	}

	// Should return empty extensions, not nil
	if ext.Status != "" {
		t.Errorf("expected empty status, got %s", ext.Status)
	}
}

func TestUpdateExtensions_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testPackageServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatal(err)
	}

	// Modify extensions
	ext.Metadata.Stars = 200
	ext.Metadata.LastUpdated = "2026-02-01T00:00:00Z"
	ext.Tools = []string{"tool_a", "tool_b", "tool_c"}

	if err := sf.UpdateExtensions(ext); err != nil {
		t.Fatalf("UpdateExtensions failed: %v", err)
	}

	// Reload and verify
	sf2, err := LoadServerFile(path)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	ext2, err := sf2.GetExtensions()
	if err != nil {
		t.Fatalf("GetExtensions after reload failed: %v", err)
	}

	if ext2.Metadata.Stars != 200 {
		t.Errorf("expected 200 stars, got %d", ext2.Metadata.Stars)
	}
	if ext2.Metadata.LastUpdated != "2026-02-01T00:00:00Z" {
		t.Errorf("expected 2026-02-01T00:00:00Z, got %s", ext2.Metadata.LastUpdated)
	}
	if len(ext2.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(ext2.Tools))
	}
}

func TestUpdateExtensions_PreservesFields(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testPackageServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatal(err)
	}

	// Update only stars
	ext.Metadata.Stars = 999

	if err := sf.UpdateExtensions(ext); err != nil {
		t.Fatal(err)
	}

	// Verify top-level fields are preserved
	var doc map[string]any
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}

	if doc["$schema"] != "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json" {
		t.Error("$schema field was lost")
	}
	if doc["name"] != "io.github.stacklok/test-server" {
		t.Error("name field was lost")
	}
	if doc["description"] != "A test server" {
		t.Error("description field was lost")
	}

	packages, ok := doc["packages"].([]interface{})
	if !ok || len(packages) != 1 {
		t.Error("packages field was lost or modified")
	}
}

func TestRepositoryURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Server with repo
	serverWithRepo := `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "repository": {"url": "https://github.com/test/repo"},
  "packages": [{"registryType": "oci", "identifier": "test:v1", "transport": {"type": "stdio"}}]
}`
	path := writeTestFile(t, dir, "with-repo.json", serverWithRepo)
	sf, _ := LoadServerFile(path)
	if sf.RepositoryURL() != "https://github.com/test/repo" {
		t.Errorf("expected repo URL, got %s", sf.RepositoryURL())
	}

	// Server without repo
	path2 := writeTestFile(t, dir, "no-repo.json", testNoMetaServerJSON)
	sf2, _ := LoadServerFile(path2)
	if sf2.RepositoryURL() != "" {
		t.Errorf("expected empty repo URL, got %s", sf2.RepositoryURL())
	}
}

func TestUpdateExtensions_RemovesStaleKeys(t *testing.T) {
	t.Parallel()

	// Simulate a server that had its package identifier bumped from v1.0.0 to v2.0.0
	// but still has a stale extension entry under the old key.
	serverWithStaleKey := `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-server",
  "description": "A test server",
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/test/server:v2.0.0",
      "transport": { "type": "stdio" }
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/test/server:v1.0.0": {
          "status": "Active",
          "tools": ["old_tool"]
        },
        "ghcr.io/test/server:v2.0.0": {
          "status": "Active",
          "tools": ["new_tool"]
        }
      }
    }
  }
}`

	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", serverWithStaleKey)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		t.Fatal(err)
	}

	// Update extensions — should remove the stale v1.0.0 key
	ext.Tools = []string{"new_tool", "another_tool"}
	if err := sf.UpdateExtensions(ext); err != nil {
		t.Fatalf("UpdateExtensions failed: %v", err)
	}

	// Reload and verify only the current key remains
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}

	meta := doc["_meta"].(map[string]any)
	pp := meta["io.modelcontextprotocol.registry/publisher-provided"].(map[string]any)
	stacklok := pp["io.github.stacklok"].(map[string]any)

	if len(stacklok) != 1 {
		t.Errorf("expected 1 extension key, got %d: %v", len(stacklok), keys(stacklok))
	}
	if _, ok := stacklok["ghcr.io/test/server:v2.0.0"]; !ok {
		t.Error("expected ghcr.io/test/server:v2.0.0 key to exist")
	}
	if _, ok := stacklok["ghcr.io/test/server:v1.0.0"]; ok {
		t.Error("stale ghcr.io/test/server:v1.0.0 key should have been removed")
	}
}

func keys(m map[string]any) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func TestUpdateExtensions_NoMeta(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeTestFile(t, dir, "server.json", testNoMetaServerJSON)

	sf, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := &registry.ServerExtensions{
		Status: statusActive,
		Tier:   "Community",
		Metadata: &registry.Metadata{
			Stars:       50,
			LastUpdated: "2026-01-01T00:00:00Z",
		},
	}

	if err := sf.UpdateExtensions(ext); err != nil {
		t.Fatalf("UpdateExtensions on no-meta server failed: %v", err)
	}

	// Reload and verify extensions were created
	sf2, err := LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext2, err := sf2.GetExtensions()
	if err != nil {
		t.Fatal(err)
	}

	if ext2.Status != statusActive {
		t.Errorf("expected Active, got %s", ext2.Status)
	}
	if ext2.Metadata == nil || ext2.Metadata.Stars != 50 {
		t.Errorf("expected 50 stars, got %v", ext2.Metadata)
	}
}
