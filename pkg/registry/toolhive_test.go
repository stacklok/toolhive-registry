package registry

import (
	"testing"

	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stretchr/testify/assert"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

func TestBuilder_Build(t *testing.T) {
	t.Parallel()
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
					Tools:       []string{"test-tool"},
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
	assert.Len(t, registry.Servers, 1)
	assert.Contains(t, registry.Servers, "test-server")
}

func TestBuilder_ValidateAgainstSchema(t *testing.T) {
	t.Parallel()
	loader := NewLoader("")
	loader.entries = map[string]*types.RegistryEntry{
		"valid-server": {
			ImageMetadata: &toolhiveRegistry.ImageMetadata{
				BaseServerMetadata: toolhiveRegistry.BaseServerMetadata{
					Name:        "valid-server",
					Description: "Valid test server",
					Transport:   "stdio",
					Tier:        "Community",
					Status:      "Active",
					Tools:       []string{"test-tool"},
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
					Name:      "invalid-server",
					Transport: "stdio",
					Tools:     []string{"test-tool"},
				},
				Image: "test/image:latest",
			},
		},
	}

	err = builder.ValidateAgainstSchema()
	assert.Error(t, err)
}
