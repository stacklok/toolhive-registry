package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/stacklok/toolhive/pkg/permissions"
	toolhiveRegistryPkg "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stacklok/toolhive/pkg/registry/converters"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry/registry"
)

const legacyRegistrySchema = "https://raw.githubusercontent.com/stacklok/toolhive/main/" +
	"pkg/registry/data/toolhive-legacy-registry.schema.json"

// LegacyBuilder assembles a legacy toolhive Registry from loaded ServerJSON entries.
// It converts each ServerJSON back to ImageMetadata or RemoteServerMetadata
// using the converters package.
type LegacyBuilder struct {
	loader *Loader
}

// NewLegacyBuilder creates a new LegacyBuilder backed by the given Loader.
func NewLegacyBuilder(loader *Loader) *LegacyBuilder {
	return &LegacyBuilder{
		loader: loader,
	}
}

// Build creates the legacy toolhive Registry structure from loaded ServerJSON entries.
// Entries are processed in sorted order for deterministic output.
func (b *LegacyBuilder) Build() (*toolhiveRegistry.Registry, error) {
	reg := &toolhiveRegistry.Registry{
		Version:       "1.0.0",
		LastUpdated:   time.Now().UTC().Format(time.RFC3339),
		Servers:       make(map[string]*toolhiveRegistry.ImageMetadata),
		RemoteServers: make(map[string]*toolhiveRegistry.RemoteServerMetadata),
	}

	names := b.loader.GetSortedNames()
	entries := b.loader.GetEntries()

	for _, name := range names {
		server := entries[name]

		switch {
		case len(server.Packages) > 0:
			imgMeta, err := converters.ServerJSONToImageMetadata(&server)
			if err != nil {
				return nil, fmt.Errorf("failed to convert '%s' to image metadata: %w", name, err)
			}
			normalizeImageMetadata(imgMeta)
			reg.Servers[name] = imgMeta

		case len(server.Remotes) > 0:
			remoteMeta, err := converters.ServerJSONToRemoteServerMetadata(&server)
			if err != nil {
				return nil, fmt.Errorf("failed to convert '%s' to remote metadata: %w", name, err)
			}
			normalizeRemoteMetadata(remoteMeta)
			reg.RemoteServers[name] = remoteMeta
		}
	}

	return reg, nil
}

// registryWithSchema wraps the legacy Registry with a $schema field for JSON output.
type registryWithSchema struct {
	Schema string `json:"$schema"`
	*toolhiveRegistry.Registry
}

// WriteJSON builds the legacy registry, validates it, and writes JSON to the given path.
func (b *LegacyBuilder) WriteJSON(path string) error {
	reg, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build registry: %w", err)
	}

	if err := validateLegacyRegistry(reg); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	wrapped := registryWithSchema{
		Schema:   legacyRegistrySchema,
		Registry: reg,
	}

	data, err := json.MarshalIndent(wrapped, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ValidateAgainstSchema builds the legacy registry and validates it without writing.
func (b *LegacyBuilder) ValidateAgainstSchema() error {
	reg, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build registry: %w", err)
	}
	return validateLegacyRegistry(reg)
}

// validateLegacyRegistry validates a legacy registry against the toolhive schema.
func validateLegacyRegistry(reg *toolhiveRegistry.Registry) error {
	registryJSON, err := json.Marshal(reg)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := toolhiveRegistryPkg.ValidateRegistrySchema(registryJSON); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}

// normalizeImageMetadata applies defaults and initializes nil slices to empty
// so JSON output uses [] instead of null, matching the legacy builder behavior.
func normalizeImageMetadata(m *toolhiveRegistry.ImageMetadata) {
	m.Name = ""

	if m.Tier == "" {
		m.Tier = "Community"
	}
	if m.Status == "" {
		m.Status = "Active"
	}
	if m.Tools == nil {
		m.Tools = []string{}
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	if m.EnvVars == nil {
		m.EnvVars = []*toolhiveRegistry.EnvVar{}
	}
	if m.Args == nil {
		m.Args = []string{}
	}

	if m.Permissions != nil {
		if m.Permissions.Read == nil {
			m.Permissions.Read = []permissions.MountDeclaration{}
		}
		if m.Permissions.Write == nil {
			m.Permissions.Write = []permissions.MountDeclaration{}
		}
		if m.Permissions.Network != nil && m.Permissions.Network.Outbound != nil {
			if m.Permissions.Network.Outbound.AllowHost == nil {
				m.Permissions.Network.Outbound.AllowHost = []string{}
			}
			if m.Permissions.Network.Outbound.AllowPort == nil {
				m.Permissions.Network.Outbound.AllowPort = []int{}
			}
		}
	}
}

// normalizeRemoteMetadata applies defaults and initializes nil slices to empty.
func normalizeRemoteMetadata(m *toolhiveRegistry.RemoteServerMetadata) {
	m.Name = ""

	if m.Tier == "" {
		m.Tier = "Community"
	}
	if m.Status == "" {
		m.Status = "Active"
	}
	if m.Tools == nil {
		m.Tools = []string{}
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	if m.EnvVars == nil {
		m.EnvVars = []*toolhiveRegistry.EnvVar{}
	}
	if m.Headers == nil {
		m.Headers = []*toolhiveRegistry.Header{}
	}
}
