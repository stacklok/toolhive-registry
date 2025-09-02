package registry

import (
	"os"
	"path/filepath"
	"testing"

	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

func TestLoader_LoadEntry(t *testing.T) {
	t.Parallel()
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test YAML file with raw YAML to avoid marshaling issues
	yamlData := []byte(`name: test-server
description: Test MCP server
image: test/image:latest
transport: stdio
tier: Community
status: Active
tools:
  - tool1
  - tool2
tags:
  - test
  - example
`)

	specPath := filepath.Join(tmpDir, "spec.yaml")
	err := os.WriteFile(specPath, yamlData, 0644)
	require.NoError(t, err)

	// Test loading the entry with a proper name
	loader := NewLoader(tmpDir)
	entry, err := loader.LoadEntryWithName(specPath, "test-server")

	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "test-server", entry.GetName())
	assert.True(t, entry.IsImage())
	assert.Equal(t, "test/image:latest", entry.Image)
	assert.Equal(t, "Test MCP server", entry.GetDescription())
	assert.Equal(t, "stdio", entry.GetTransport())
	assert.Len(t, entry.GetTools(), 2)
}

func TestLoader_ValidateEntry(t *testing.T) {
	t.Parallel()
	loader := NewLoader("")

	tests := []struct {
		name    string
		entry   *types.RegistryEntry
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid entry",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Tier:        types.TierOfficial,
						Status:      types.StatusActive,
						Tools:       []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: false,
		},
		{
			name: "missing image",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Tools:       []string{"test-tool"},
					},
				},
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "missing description",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Transport: "stdio",
						Tools:     []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "missing transport",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Tools:       []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "transport is required",
		},
		{
			name: "missing tools",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "at least one tool must be specified",
		},
		{
			name: "invalid transport",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "invalid",
						Tools:       []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "schema validation failed",
		},
		{
			name: "invalid tier",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Tier:        "InvalidTier",
						Tools:       []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "schema validation failed",
		},
		{
			name: "invalid status",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Status:      "InvalidStatus",
						Tools:       []string{"test-tool"},
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "schema validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := loader.validateEntry(tt.entry, "test-entry")
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoader_LoadAll(t *testing.T) {
	t.Parallel()
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create multiple test entries with raw YAML to avoid marshaling issues
	server1YAML := `name: server1
description: Test server 1
transport: stdio
image: test/server1:latest
tier: Community
status: Active
tools:
  - tool1`

	server2YAML := `name: server2
description: Test server 2
transport: sse
image: test/server2:latest
tier: Community
status: Active
tools:
  - tool2`

	entries := map[string]string{
		"server1": server1YAML,
		"server2": server2YAML,
	}

	// Create directories and spec files
	for name, yamlContent := range entries {
		dir := filepath.Join(tmpDir, name)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		specPath := filepath.Join(dir, "spec.yaml")
		err = os.WriteFile(specPath, []byte(yamlContent), 0644)
		require.NoError(t, err)
	}

	// Test loading all entries
	loader := NewLoader(tmpDir)
	err := loader.LoadAll()
	assert.NoError(t, err)

	loadedEntries := loader.GetEntries()
	assert.Len(t, loadedEntries, 2)
	assert.Contains(t, loadedEntries, "server1")
	assert.Contains(t, loadedEntries, "server2")

	sortedEntries := loader.GetSortedEntries()
	assert.Len(t, sortedEntries, 2)
}
