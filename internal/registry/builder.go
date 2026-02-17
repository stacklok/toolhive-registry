package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stacklok/toolhive/pkg/registry/registry"
)

// Builder assembles an UpstreamRegistry from loaded ServerJSON entries.
type Builder struct {
	loader *Loader
}

// NewBuilder creates a new Builder backed by the given Loader.
func NewBuilder(loader *Loader) *Builder {
	return &Builder{
		loader: loader,
	}
}

// Build creates the UpstreamRegistry structure from loaded entries.
// Servers are ordered by directory name for deterministic output.
func (b *Builder) Build() *registry.UpstreamRegistry {
	names := b.loader.GetSortedNames()

	servers := make([]upstream.ServerJSON, 0, len(names))
	for _, name := range names {
		servers = append(servers, b.loader.GetEntries()[name])
	}

	return &registry.UpstreamRegistry{
		Schema:  "https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/upstream-registry.schema.json",
		Version: "1.0.0",
		Meta: registry.UpstreamMeta{
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
		Data: registry.UpstreamData{
			Servers: servers,
			Groups:  []registry.UpstreamGroup{},
		},
	}
}

// WriteJSON builds the registry, validates it, and writes JSON to the given path.
func (b *Builder) WriteJSON(path string) error {
	builtRegistry := b.Build()

	if err := validateRegistry(builtRegistry); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(builtRegistry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ValidateAgainstSchema builds the registry and validates it without writing.
func (b *Builder) ValidateAgainstSchema() error {
	builtRegistry := b.Build()
	return validateRegistry(builtRegistry)
}

// validateRegistry validates a registry object against the upstream registry schema.
func validateRegistry(upstreamRegistry *registry.UpstreamRegistry) error {
	registryJSON, err := json.Marshal(upstreamRegistry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := toolhiveRegistry.ValidateUpstreamRegistry(registryJSON); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}
