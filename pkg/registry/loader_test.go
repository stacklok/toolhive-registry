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
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test spec.yaml file - write raw YAML to avoid marshaling issues
	yamlData := []byte(`name: test-server
image: test/image:latest
description: Test MCP server
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

	// Test loading the entry
	loader := NewLoader(tmpDir)
	entry, err := loader.LoadEntry(specPath)

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
						Tier:        "Official",
						Status:      "Active",
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
					},
				},
			},
			wantErr: true,
			errMsg:  "image is required for image-based servers",
		},
		{
			name: "missing description",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Transport: "stdio",
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
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "transport is required",
		},
		{
			name: "invalid transport",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "invalid",
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "invalid transport",
		},
		{
			name: "invalid tier",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Tier:        "InvalidTier",
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "invalid tier",
		},
		{
			name: "invalid status",
			entry: &types.RegistryEntry{
				ImageMetadata: &toolhiveRegistry.ImageMetadata{
					BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
						Description: "Test server",
						Transport:   "stdio",
						Status:      "InvalidStatus",
					},
					Image: "test/image:latest",
				},
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := loader.validateEntry(tt.entry)
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
	entries := map[string]string{
		"server1": `name: server1
description: Test server 1
transport: stdio
image: test/server1:latest`,
		"server2": `name: server2
description: Test server 2
transport: sse
image: test/server2:latest`,
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
	assert.Len(t, loader.GetEntries(), 2)

	// Check that entries were loaded correctly
	loadedEntries := loader.GetEntries()
	assert.Contains(t, loadedEntries, "server1")
	assert.Contains(t, loadedEntries, "server2")

	// Test GetSortedEntries
	sorted := loader.GetSortedEntries()
	assert.Len(t, sorted, 2)
	assert.Equal(t, "server1", sorted[0].GetName())
	assert.Equal(t, "server2", sorted[1].GetName())
}

func TestBuilder_Build(t *testing.T) {
	t.Parallel()
	// Create a loader with test data
	loader := NewLoader("")
	loader.entries = map[string]*types.RegistryEntry{
		"test-server": {
			ImageMetadata: &toolhiveRegistry.ImageMetadata{
				BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
					Name:        "test-server",
					Description: "Test server",
					Transport:   "stdio",
					Tier:        "Community",
					Status:      "Active",
				},
				Image: "test/image:latest",
			},
		},
	}

	// Create builder and build
	builder := NewBuilder(loader)
	registry, err := builder.Build()

	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "1.0.0", registry.Version)
	assert.Len(t, registry.Servers, 1)
	assert.Contains(t, registry.Servers, "test-server")

	// Check that defaults were set
	server := registry.Servers["test-server"]
	assert.Equal(t, "Community", server.Tier)
	assert.Equal(t, "Active", server.Status)
	assert.NotNil(t, server.Tools)
	assert.NotNil(t, server.Tags)
	assert.NotNil(t, server.EnvVars)
	assert.NotNil(t, server.Args)
}

func TestBuilder_ValidateAgainstSchema(t *testing.T) {
	t.Parallel()
	// Test with valid entries
	loader := NewLoader("")
	loader.entries = map[string]*types.RegistryEntry{
		"valid-server": {
			ImageMetadata: &toolhiveRegistry.ImageMetadata{
				BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
					Description: "Valid server",
					Transport:   "stdio",
				},
				Image: "test/image:latest",
			},
		},
	}

	builder := NewBuilder(loader)
	err := builder.ValidateAgainstSchema()
	assert.NoError(t, err)

	// Test with invalid entry (missing required field)
	loader.entries = map[string]*types.RegistryEntry{
		"invalid-server": {
			ImageMetadata: &toolhiveRegistry.ImageMetadata{
				BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
					// Missing Description and Transport
				},
				Image: "test/image:latest",
			},
		},
	}

	err = builder.ValidateAgainstSchema()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description is required")
}
