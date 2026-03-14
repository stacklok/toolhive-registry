package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	thvregistry "github.com/stacklok/toolhive-core/registry/types"
)

// Builder assembles an UpstreamRegistry from loaded ServerJSON entries and optional skills.
type Builder struct {
	loader      *Loader
	skillLoader *SkillLoader
}

// NewBuilder creates a new Builder backed by the given Loader.
func NewBuilder(loader *Loader) *Builder {
	return &Builder{
		loader: loader,
	}
}

// WithSkillLoader sets the skill loader on the builder.
func (b *Builder) WithSkillLoader(sl *SkillLoader) *Builder {
	b.skillLoader = sl
	return b
}

// Build creates the UpstreamRegistry structure from loaded entries.
// Servers and skills are ordered by directory name for deterministic output.
func (b *Builder) Build() *thvregistry.UpstreamRegistry {
	names := b.loader.GetSortedNames()

	servers := make([]upstream.ServerJSON, 0, len(names))
	for _, name := range names {
		servers = append(servers, b.loader.GetEntries()[name])
	}

	var skills []thvregistry.Skill
	if b.skillLoader != nil && len(b.skillLoader.GetEntries()) > 0 {
		skillNames := b.skillLoader.GetSortedNames()
		skills = make([]thvregistry.Skill, 0, len(skillNames))
		for _, name := range skillNames {
			skills = append(skills, b.skillLoader.GetEntries()[name])
		}
	}

	return &thvregistry.UpstreamRegistry{
		Schema:  "https://raw.githubusercontent.com/stacklok/toolhive-core/main/registry/types/data/upstream-registry.schema.json",
		Version: "1.0.0",
		Meta: thvregistry.UpstreamMeta{
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
		Data: thvregistry.UpstreamData{
			Servers: servers,
			Groups:  []thvregistry.UpstreamGroup{},
			Skills:  skills,
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
func validateRegistry(upstreamRegistry *thvregistry.UpstreamRegistry) error {
	registryJSON, err := json.Marshal(upstreamRegistry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := thvregistry.ValidateUpstreamRegistryBytes(registryJSON); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}
