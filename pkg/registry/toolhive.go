package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/stacklok/toolhive/pkg/permissions"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
)

// Builder builds the final registry JSON from loaded entries
type Builder struct {
	loader *Loader
}

// NewBuilder creates a new registry builder
func NewBuilder(loader *Loader) *Builder {
	return &Builder{
		loader: loader,
	}
}

// Build creates the final registry structure compatible with toolhive
func (b *Builder) Build() (*toolhiveRegistry.Registry, error) {
	registry := &toolhiveRegistry.Registry{
		Version:       "1.0.0",
		LastUpdated:   time.Now().UTC().Format(time.RFC3339),
		Servers:       make(map[string]*toolhiveRegistry.ImageMetadata),
		RemoteServers: make(map[string]*toolhiveRegistry.RemoteServerMetadata),
	}

	// Get all entry names and sort them alphabetically
	var names []string
	for name := range b.loader.GetEntries() {
		names = append(names, name)
	}
	sort.Strings(names)

	// Convert our extended entries back to toolhive format in alphabetical order
	for _, name := range names {
		entry := b.loader.GetEntries()[name]

		if entry.IsImage() {
			// Process image-based server
			metadata := b.processImageMetadata(entry.ImageMetadata)
			registry.Servers[name] = metadata
		} else if entry.IsRemote() {
			// Process remote server
			metadata := b.processRemoteMetadata(entry.RemoteServerMetadata)
			registry.RemoteServers[name] = metadata
		}
	}

	return registry, nil
}

// processImageMetadata processes and normalizes ImageMetadata
func (*Builder) processImageMetadata(metadata *toolhiveRegistry.ImageMetadata) *toolhiveRegistry.ImageMetadata {
	// Create a copy of the ImageMetadata
	result := *metadata

	// Don't set the name field - the key serves as the name
	result.Name = ""

	// Set defaults if not specified
	if result.Tier == "" {
		result.Tier = "Community"
	}

	if result.Status == "" {
		result.Status = "Active"
	}

	// Initialize empty slices if nil to match JSON output
	if result.Tools == nil {
		result.Tools = []string{}
	}

	if result.Tags == nil {
		result.Tags = []string{}
	}

	if result.EnvVars == nil {
		result.EnvVars = []*toolhiveRegistry.EnvVar{}
	}

	if result.Args == nil {
		result.Args = []string{}
	}

	// Ensure permissions structure matches upstream format
	if result.Permissions != nil {
		// Initialize empty slices for read/write if nil
		if result.Permissions.Read == nil {
			result.Permissions.Read = []permissions.MountDeclaration{}
		}
		if result.Permissions.Write == nil {
			result.Permissions.Write = []permissions.MountDeclaration{}
		}

		// Ensure network permissions have explicit insecure_allow_all
		if result.Permissions.Network != nil && result.Permissions.Network.Outbound != nil {
			// Initialize empty slices if nil
			if result.Permissions.Network.Outbound.AllowHost == nil {
				result.Permissions.Network.Outbound.AllowHost = []string{}
			}
			if result.Permissions.Network.Outbound.AllowPort == nil {
				result.Permissions.Network.Outbound.AllowPort = []int{}
			}
		}
	}

	return &result
}

// processRemoteMetadata processes and normalizes RemoteServerMetadata
func (*Builder) processRemoteMetadata(metadata *toolhiveRegistry.RemoteServerMetadata) *toolhiveRegistry.RemoteServerMetadata {
	// Create a copy of the RemoteServerMetadata
	result := *metadata

	// Don't set the name field - the key serves as the name
	result.Name = ""

	// Set defaults if not specified
	if result.Tier == "" {
		result.Tier = "Community"
	}

	if result.Status == "" {
		result.Status = "Active"
	}

	// Initialize empty slices if nil to match JSON output
	if result.Tools == nil {
		result.Tools = []string{}
	}

	if result.Tags == nil {
		result.Tags = []string{}
	}

	if result.EnvVars == nil {
		result.EnvVars = []*toolhiveRegistry.EnvVar{}
	}

	if result.Headers == nil {
		result.Headers = []*toolhiveRegistry.Header{}
	}

	return &result
}

// WriteJSON writes the registry to a JSON file
func (b *Builder) WriteJSON(path string) error {
	registry, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build registry: %w", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a wrapper struct that includes the schema field
	type registryWithSchema struct {
		Schema string `json:"$schema"`
		*toolhiveRegistry.Registry
	}

	// Wrap the registry with the schema
	wrappedRegistry := registryWithSchema{
		Schema:   "https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/schema.json",
		Registry: registry,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(wrappedRegistry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ValidateAgainstSchema validates the built registry against the toolhive schema
func (b *Builder) ValidateAgainstSchema() error {
	registry, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build registry: %w", err)
	}

	// Use the comprehensive schema validator
	validator := NewSchemaValidator()

	if err := validator.ValidateRegistry(registry); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}
