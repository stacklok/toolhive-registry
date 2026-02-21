package thvclient

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stacklok/toolhive-core/permissions"
	"github.com/stacklok/toolhive-core/registry/types"

	"github.com/stacklok/toolhive-catalog/internal/serverjson"
)

func writeTestServerJSON(t *testing.T, dir, content string) string {
	t.Helper()
	serverDir := filepath.Join(dir, "test")
	if err := os.MkdirAll(serverDir, 0750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(serverDir, "server.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuildRunCommand_Basic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	serverJSON := `{
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "ghcr.io/test/server:v1",
    "transport": {"type": "stdio"}
  }]
}`
	path := writeTestServerJSON(t, dir, serverJSON)
	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := &registry.ServerExtensions{}
	args := BuildRunCommand(sf, ext, "temp-test-123", "ghcr.io/test/server:v1")

	expected := []string{
		"run", "--name", "temp-test-123",
		"--transport", "stdio",
		"ghcr.io/test/server:v1",
	}

	if !slices.Equal(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

func TestBuildRunCommand_WithEnvVars(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	serverJSON := `{
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "ghcr.io/test/server:v1",
    "transport": {"type": "stdio"},
    "environmentVariables": [
      {"name": "TOKEN", "isRequired": true, "isSecret": true},
      {"name": "HOST", "default": "localhost"},
      {"name": "OPTIONAL_SECRET", "isSecret": true}
    ]
  }]
}`
	path := writeTestServerJSON(t, dir, serverJSON)
	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := &registry.ServerExtensions{}
	args := BuildRunCommand(sf, ext, "temp-test-123", "ghcr.io/test/server:v1")

	if !containsEnvVar(args, "TOKEN", "placeholder") {
		t.Error("expected TOKEN=placeholder")
	}
	if !containsEnvVar(args, "HOST", "localhost") {
		t.Error("expected HOST=localhost")
	}
	if !containsEnvVar(args, "OPTIONAL_SECRET", "placeholder") {
		t.Error("expected OPTIONAL_SECRET=placeholder")
	}
}

func TestBuildRunCommand_WithPermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	serverJSON := `{
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "test:v1",
    "transport": {"type": "stdio"}
  }]
}`
	path := writeTestServerJSON(t, dir, serverJSON)
	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := &registry.ServerExtensions{
		Permissions: &permissions.Profile{
			Network: &permissions.NetworkPermissions{
				Outbound: &permissions.OutboundNetworkPermissions{
					AllowHost: []string{"example.com"},
				},
			},
		},
	}
	args := BuildRunCommand(sf, ext, "temp-test", "test:v1")

	if !slices.Contains(args, "--permission-profile") {
		t.Error("expected --permission-profile flag")
	}
	idx := slices.Index(args, "--permission-profile")
	if idx+1 >= len(args) || args[idx+1] != "network" {
		t.Error("expected --permission-profile network")
	}
}

func TestBuildRunCommand_WithArgs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	serverJSON := `{
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "test:v1",
    "transport": {"type": "stdio"}
  }]
}`
	path := writeTestServerJSON(t, dir, serverJSON)
	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := &registry.ServerExtensions{
		Args: []string{"--mode", "read-only"},
	}
	args := BuildRunCommand(sf, ext, "temp-test", "test:v1")

	// Should have: ... test:v1 -- --mode read-only
	dashIdx := slices.Index(args, "--")
	if dashIdx == -1 {
		t.Fatal("expected -- separator")
	}
	if !slices.Equal(args[dashIdx+1:], []string{"--mode", "read-only"}) {
		t.Errorf("expected [--mode read-only] after --, got %v", args[dashIdx+1:])
	}
}

func TestCommandBuilder(t *testing.T) {
	t.Parallel()
	args := NewCommandBuilder("run").
		AddFlag("--name", "test").
		AddBoolFlag("--verbose", true).
		AddBoolFlag("--quiet", false).
		AddEnvVar("KEY", "value").
		AddPositional("image:latest").
		Build()

	expected := []string{
		"run", "--name", "test", "--verbose",
		"-e", "KEY=value", "image:latest",
	}
	if !slices.Equal(args, expected) {
		t.Errorf("expected %v, got %v", expected, args)
	}
}

// containsEnvVar checks if args contain -e NAME=VALUE pair.
func containsEnvVar(args []string, name, value string) bool {
	expected := fmt.Sprintf("%s=%s", name, value)
	for i, arg := range args {
		if arg == "-e" && i+1 < len(args) && args[i+1] == expected {
			return true
		}
	}
	return false
}
