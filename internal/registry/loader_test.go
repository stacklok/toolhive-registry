package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalServerJSON = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/test-server",
  "description": "A test server",
  "version": "1.0.0"
}`

const minimalServerJSON2 = `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/another-server",
  "description": "Another test server",
  "version": "1.0.0"
}`

func TestLoader_LoadAll(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create two server directories with server.json files
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alpha"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alpha", "server.json"), []byte(minimalServerJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "beta"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "beta", "server.json"), []byte(minimalServerJSON2), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 2)
	assert.Contains(t, entries, "alpha")
	assert.Contains(t, entries, "beta")

	// Verify sorted names
	names := loader.GetSortedNames()
	assert.Equal(t, []string{"alpha", "beta"}, names)

	// Verify content was parsed correctly
	assert.Equal(t, "io.github.stacklok/test-server", entries["alpha"].Name)
	assert.Equal(t, "A test server", entries["alpha"].Description)
}

func TestLoader_LoadAll_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "broken"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken", "server.json"), []byte(`{invalid json`), 0600))

	loader := NewLoader(dir)
	err := loader.LoadAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestLoader_LoadAll_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Empty(t, entries)

	names := loader.GetSortedNames()
	assert.Empty(t, names)
}

func TestLoader_LoadAll_SkipsHiddenDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a hidden directory with server.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".hidden"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".hidden", "server.json"), []byte(minimalServerJSON), 0600))

	// Create a visible directory with server.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "visible"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "visible", "server.json"), []byte(minimalServerJSON), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries, "visible")
	assert.NotContains(t, entries, ".hidden")
}

func TestLoader_LoadAll_SkipsMissingServerJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a directory without server.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "empty-dir"), 0750))

	// Create a directory with server.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "valid"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "valid", "server.json"), []byte(minimalServerJSON), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries, "valid")
}
