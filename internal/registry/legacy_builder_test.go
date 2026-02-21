package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	toolhiveRegistry "github.com/stacklok/toolhive-core/registry/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Realistic container-based server.json (minimal but complete for converter)
const imageServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-image",
  "description": "A container-based test server",
  "title": "test-image",
  "repository": {"url": "https://github.com/example/test", "source": "github"},
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "ghcr.io/example/test:1.0.0",
    "transport": {"type": "stdio"},
    "environmentVariables": [
      {"name": "API_KEY", "description": "The API key", "isRequired": true, "isSecret": true}
    ]
  }],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/example/test:1.0.0": {
          "status": "Active",
          "tier": "Official",
          "tools": ["do_thing"],
          "tags": ["test"],
          "metadata": {"stars": 10, "pulls": 5, "last_updated": "2026-01-01T00:00:00Z"},
          "permissions": {
            "network": {
              "outbound": {"insecure_allow_all": true}
            }
          }
        }
      }
    }
  }
}`

// Realistic remote server.json (minimal but complete for converter)
const remoteServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-remote",
  "description": "A remote test server",
  "title": "test-remote",
  "repository": {"url": "https://github.com/example/remote", "source": "github"},
  "version": "1.0.0",
  "remotes": [{
    "type": "streamable-http",
    "url": "https://mcp.example.com/mcp"
  }],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://mcp.example.com/mcp": {
          "status": "Active",
          "tier": "Official",
          "tools": ["remote_tool"],
          "tags": ["remote"],
          "custom_metadata": {"author": "Example", "homepage": "https://example.com"}
        }
      }
    }
  }
}`

func setupLegacyLoader(t *testing.T) *Loader {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "test-image"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test-image", "server.json"), []byte(imageServerJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "test-remote"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test-remote", "server.json"), []byte(remoteServerJSON), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())
	return loader
}

func TestLegacyBuilder_Build(t *testing.T) {
	t.Parallel()

	loader := setupLegacyLoader(t)
	builder := NewLegacyBuilder(loader)

	reg, err := builder.Build()
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", reg.Version)
	assert.NotEmpty(t, reg.LastUpdated)

	// Image server should be in Servers
	assert.Len(t, reg.Servers, 1)
	assert.Contains(t, reg.Servers, "test-image")

	imgMeta := reg.Servers["test-image"]
	assert.Empty(t, imgMeta.Name, "name should be cleared — the map key serves as the name")
	assert.Equal(t, "A container-based test server", imgMeta.Description)
	assert.Equal(t, "stdio", imgMeta.Transport)
	assert.Equal(t, "ghcr.io/example/test:1.0.0", imgMeta.Image)
	assert.Equal(t, "Official", imgMeta.Tier)
	assert.Equal(t, "Active", imgMeta.Status)
	assert.Equal(t, []string{"do_thing"}, imgMeta.Tools)
	assert.Equal(t, []string{"test"}, imgMeta.Tags)
	require.Len(t, imgMeta.EnvVars, 1)
	assert.Equal(t, "API_KEY", imgMeta.EnvVars[0].Name)

	// Remote server should be in RemoteServers
	assert.Len(t, reg.RemoteServers, 1)
	assert.Contains(t, reg.RemoteServers, "test-remote")

	remoteMeta := reg.RemoteServers["test-remote"]
	assert.Empty(t, remoteMeta.Name)
	assert.Equal(t, "A remote test server", remoteMeta.Description)
	assert.Equal(t, "streamable-http", remoteMeta.Transport)
	assert.Equal(t, "https://mcp.example.com/mcp", remoteMeta.URL)
	assert.Equal(t, "Official", remoteMeta.Tier)
	assert.Equal(t, []string{"remote_tool"}, remoteMeta.Tools)
	assert.Equal(t, []string{"remote"}, remoteMeta.Tags)
}

func TestLegacyBuilder_Build_Defaults(t *testing.T) {
	t.Parallel()

	// Minimal server.json with no extensions — should get defaults
	minimalImage := `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/bare",
  "description": "Bare server",
  "version": "1.0.0",
  "packages": [{
    "registryType": "oci",
    "identifier": "ghcr.io/example/bare:1.0.0",
    "transport": {"type": "stdio"}
  }]
}`

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bare"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bare", "server.json"), []byte(minimalImage), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())

	builder := NewLegacyBuilder(loader)
	reg, err := builder.Build()
	require.NoError(t, err)

	imgMeta := reg.Servers["bare"]
	require.NotNil(t, imgMeta)

	// Defaults applied
	assert.Equal(t, "Community", imgMeta.Tier)
	assert.Equal(t, "Active", imgMeta.Status)

	// Nil slices initialized to empty
	assert.NotNil(t, imgMeta.Tools)
	assert.NotNil(t, imgMeta.Tags)
	assert.NotNil(t, imgMeta.EnvVars)
	assert.NotNil(t, imgMeta.Args)
}

func TestLegacyBuilder_WriteJSON(t *testing.T) {
	t.Parallel()

	loader := setupLegacyLoader(t)
	builder := NewLegacyBuilder(loader)

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "output", "registry.json")

	err := builder.WriteJSON(outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	// Verify it's valid JSON with the schema wrapper
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "$schema")
	assert.Contains(t, raw, "version")
	assert.Contains(t, raw, "servers")
	assert.Contains(t, raw, "remote_servers")

	// Verify schema value
	var schema string
	require.NoError(t, json.Unmarshal(raw["$schema"], &schema))
	assert.Equal(t, "https://raw.githubusercontent.com/stacklok/toolhive-core/main/registry/types/data/toolhive-legacy-registry.schema.json", schema)

	// Verify servers content
	var servers map[string]*toolhiveRegistry.ImageMetadata
	require.NoError(t, json.Unmarshal(raw["servers"], &servers))
	assert.Contains(t, servers, "test-image")

	var remotes map[string]*toolhiveRegistry.RemoteServerMetadata
	require.NoError(t, json.Unmarshal(raw["remote_servers"], &remotes))
	assert.Contains(t, remotes, "test-remote")
}

func TestLegacyBuilder_ValidateAgainstSchema(t *testing.T) {
	t.Parallel()

	loader := setupLegacyLoader(t)
	builder := NewLegacyBuilder(loader)

	err := builder.ValidateAgainstSchema()
	assert.NoError(t, err)
}
