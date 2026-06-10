package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stacklok/toolhive-core/registry/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoaderWithEntries(t *testing.T) *Loader {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alpha"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alpha", "server.json"), []byte(minimalServerJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "beta"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "beta", "server.json"), []byte(minimalServerJSON2), 0600))

	loader := NewLoader(dir)
	require.NoError(t, loader.LoadAll())
	return loader
}

func TestBuilder_Build(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	builder := NewBuilder(loader)

	result := builder.Build()

	assert.Equal(t, "https://raw.githubusercontent.com/stacklok/toolhive-core/main/registry/types/data/upstream-registry.schema.json", result.Schema)
	assert.Equal(t, "1.0.0", result.Version)
	assert.NotEmpty(t, result.Meta.LastUpdated)
	assert.Len(t, result.Data.Servers, 2)

	// Verify sorted order: alpha before beta
	assert.Equal(t, "io.github.stacklok/test-server", result.Data.Servers[0].Name)
	assert.Equal(t, "io.github.stacklok/another-server", result.Data.Servers[1].Name)
}

func TestBuilder_WriteJSON(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	builder := NewBuilder(loader)

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "output", "registry-upstream.json")

	err := builder.WriteJSON(outPath)
	require.NoError(t, err)

	// Read the file back and verify it's valid JSON
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var reg registry.UpstreamRegistry
	require.NoError(t, json.Unmarshal(data, &reg))

	assert.Equal(t, "1.0.0", reg.Version)
	assert.Len(t, reg.Data.Servers, 2)
}

func TestBuilder_ValidateAgainstSchema(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	builder := NewBuilder(loader)

	err := builder.ValidateAgainstSchema()
	assert.NoError(t, err)
}

func setupSkillLoaderWithEntries(t *testing.T) *SkillLoader {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "code-review"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "code-review", "skill.json"), []byte(minimalSkillJSON), 0600))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "debugging"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "debugging", "skill.json"), []byte(minimalSkillJSON2), 0600))

	loader := NewSkillLoader(dir)
	require.NoError(t, loader.LoadAll())
	return loader
}

func TestBuilder_BuildWithSkills(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	skillLoader := setupSkillLoaderWithEntries(t)
	builder := NewBuilder(loader).WithSkillLoader(skillLoader)

	result := builder.Build()

	assert.Len(t, result.Data.Servers, 2)
	assert.Len(t, result.Data.Skills, 2)

	// Verify skills are sorted: code-review before debugging
	assert.Equal(t, "test-skill", result.Data.Skills[0].Name)
	assert.Equal(t, "another-skill", result.Data.Skills[1].Name)
}

func TestBuilder_BuildWithoutSkills(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	builder := NewBuilder(loader)

	result := builder.Build()

	assert.Len(t, result.Data.Servers, 2)
	assert.Nil(t, result.Data.Skills)
}

func TestBuilder_ValidateWithSkills(t *testing.T) {
	t.Parallel()

	loader := setupLoaderWithEntries(t)
	skillLoader := setupSkillLoaderWithEntries(t)
	builder := NewBuilder(loader).WithSkillLoader(skillLoader)

	err := builder.ValidateAgainstSchema()
	assert.NoError(t, err)
}
