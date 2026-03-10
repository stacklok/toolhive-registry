package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalSkillJSON = `{
  "namespace": "io.github.stacklok",
  "name": "test-skill",
  "description": "A test skill",
  "version": "0.1.0"
}`

const minimalSkillJSON2 = `{
  "namespace": "io.github.stacklok",
  "name": "another-skill",
  "description": "Another test skill",
  "version": "0.2.0"
}`

func TestSkillLoader_LoadAll(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alpha"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alpha", "skill.json"), []byte(minimalSkillJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "beta"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "beta", "skill.json"), []byte(minimalSkillJSON2), 0600))

	loader := NewSkillLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 2)
	assert.Contains(t, entries, "alpha")
	assert.Contains(t, entries, "beta")

	// Verify sorted names
	names := loader.GetSortedNames()
	assert.Equal(t, []string{"alpha", "beta"}, names)

	// Verify content was parsed correctly
	assert.Equal(t, "test-skill", entries["alpha"].Name)
	assert.Equal(t, "A test skill", entries["alpha"].Description)
	assert.Equal(t, "0.1.0", entries["alpha"].Version)
}

func TestSkillLoader_LoadAll_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "broken"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken", "skill.json"), []byte(`{invalid json`), 0600))

	loader := NewSkillLoader(dir)
	err := loader.LoadAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestSkillLoader_LoadAll_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	loader := NewSkillLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Empty(t, entries)

	names := loader.GetSortedNames()
	assert.Empty(t, names)
}

func TestSkillLoader_LoadAll_SkipsHiddenDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".hidden"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".hidden", "skill.json"), []byte(minimalSkillJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "visible"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "visible", "skill.json"), []byte(minimalSkillJSON), 0600))

	loader := NewSkillLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries, "visible")
	assert.NotContains(t, entries, ".hidden")
}

func TestSkillLoader_LoadAll_SkipsMissingSkillJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a directory without skill.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "empty-dir"), 0750))

	// Create a directory with skill.json
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "valid"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "valid", "skill.json"), []byte(minimalSkillJSON), 0600))

	loader := NewSkillLoader(dir)
	require.NoError(t, loader.LoadAll())

	entries := loader.GetEntries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries, "valid")
}
