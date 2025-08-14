// Package registry provides functionality for loading and managing registry entries
package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/stacklok/toolhive/pkg/permissions"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

// Loader handles loading registry entries from YAML files
type Loader struct {
	registryPath string
	entries      map[string]*types.RegistryEntry
}

// NewLoader creates a new registry loader
func NewLoader(registryPath string) *Loader {
	return &Loader{
		registryPath: registryPath,
		entries:      make(map[string]*types.RegistryEntry),
	}
}

// LoadAll loads all registry entries from the registry directory
func (l *Loader) LoadAll() error {
	// Walk through the registry directory
	err := filepath.Walk(l.registryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory or if it's the root directory
		if !info.IsDir() || path == l.registryPath {
			return nil
		}

		// Get the relative path from registry root
		relPath, err := filepath.Rel(l.registryPath, path)
		if err != nil {
			return err
		}

		// Skip hidden directories and nested directories
		if strings.HasPrefix(info.Name(), ".") || strings.Contains(relPath, string(os.PathSeparator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Try to load spec.yaml from this directory
		specPath := filepath.Join(path, "spec.yaml")
		if _, err := os.Stat(specPath); err == nil {
			entry, err := l.LoadEntry(specPath)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", specPath, err)
			}

			// Use directory name as the key if Name is not set
			entryName := info.Name()
			if entry.Name != "" {
				entryName = entry.Name
			} else {
				entry.Name = entryName
			}

			l.entries[entryName] = entry
		}

		return nil
	})

	return err
}

// LoadEntry loads a single registry entry from a YAML file
func (l *Loader) LoadEntry(path string) (*types.RegistryEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var entry types.RegistryEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Initialize embedded ImageMetadata if nil
	if entry.ImageMetadata == nil {
		entry.ImageMetadata = &toolhiveRegistry.ImageMetadata{}
	}

	// Validate required fields
	if err := l.validateEntry(&entry); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &entry, nil
}

// validateEntry validates a registry entry
func (l *Loader) validateEntry(entry *types.RegistryEntry) error {
	if entry.Image == "" {
		return fmt.Errorf("image is required")
	}

	if entry.Description == "" {
		return fmt.Errorf("description is required")
	}

	if entry.Transport == "" {
		return fmt.Errorf("transport is required")
	}

	// Validate transport type
	validTransports := map[string]bool{
		"stdio":           true,
		"sse":             true,
		"streamable-http": true,
	}

	if !validTransports[entry.Transport] {
		return fmt.Errorf("invalid transport: %s (must be stdio, sse, or streamable-http)", entry.Transport)
	}

	// Validate tier if specified
	if entry.Tier != "" {
		validTiers := map[string]bool{
			"Official":  true,
			"Community": true,
			"Partner":   true,
		}

		if !validTiers[entry.Tier] {
			return fmt.Errorf("invalid tier: %s (must be Official, Community, or Partner)", entry.Tier)
		}
	}

	// Validate status if specified
	if entry.Status != "" {
		validStatuses := map[string]bool{
			"Active":     true,
			"Deprecated": true,
			"Beta":       true,
			"Alpha":      true,
		}

		if !validStatuses[entry.Status] {
			return fmt.Errorf("invalid status: %s", entry.Status)
		}
	}

	return nil
}

// GetEntries returns all loaded entries
func (l *Loader) GetEntries() map[string]*types.RegistryEntry {
	return l.entries
}

// GetSortedEntries returns entries sorted by name
func (l *Loader) GetSortedEntries() []*types.RegistryEntry {
	var entries []*types.RegistryEntry
	for _, entry := range l.entries {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

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
		Version:     "1.0.0",
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Servers:     make(map[string]*toolhiveRegistry.ImageMetadata),
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

		// Create a copy of the ImageMetadata
		metadata := *entry.ImageMetadata

		// Don't set the name field - the key serves as the name
		metadata.Name = ""

		// Set defaults if not specified
		if metadata.Tier == "" {
			metadata.Tier = "Community"
		}

		if metadata.Status == "" {
			metadata.Status = "Active"
		}

		// Initialize empty slices if nil to match JSON output
		if metadata.Tools == nil {
			metadata.Tools = []string{}
		}

		if metadata.Tags == nil {
			metadata.Tags = []string{}
		}

		if metadata.EnvVars == nil {
			metadata.EnvVars = []*toolhiveRegistry.EnvVar{}
		}

		if metadata.Args == nil {
			metadata.Args = []string{}
		}

		// Ensure permissions structure matches upstream format
		if metadata.Permissions != nil {
			// Initialize empty slices for read/write if nil
			if metadata.Permissions.Read == nil {
				metadata.Permissions.Read = []permissions.MountDeclaration{}
			}
			if metadata.Permissions.Write == nil {
				metadata.Permissions.Write = []permissions.MountDeclaration{}
			}

			// Ensure network permissions have explicit insecure_allow_all
			if metadata.Permissions.Network != nil && metadata.Permissions.Network.Outbound != nil {
				// InsecureAllowAll is already a bool, so it will be false by default
				// But we want to ensure it's explicitly in the output

				// Initialize empty slices if nil
				if metadata.Permissions.Network.Outbound.AllowHost == nil {
					metadata.Permissions.Network.Outbound.AllowHost = []string{}
				}
				if metadata.Permissions.Network.Outbound.AllowPort == nil {
					metadata.Permissions.Network.Outbound.AllowPort = []int{}
				}
			}
		}

		registry.Servers[name] = &metadata
	}

	return registry, nil
}

// WriteJSON writes the registry to a JSON file
func (b *Builder) WriteJSON(path string) error {
	registry, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build registry: %w", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a wrapper struct that includes the schema field
	type registryWithSchema struct {
		Schema string `json:"$schema"`
		*toolhiveRegistry.Registry
	}

	// Wrap the registry with the schema
	wrappedRegistry := registryWithSchema{
		Schema:   "https://raw.githubusercontent.com/stacklok/toolhive/main/docs/registry/schema.json",
		Registry: registry,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(wrappedRegistry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
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

	// Basic validation - ensure required fields are present
	for name, server := range registry.Servers {
		if server.Image == "" {
			return fmt.Errorf("server %s: image is required", name)
		}
		if server.Description == "" {
			return fmt.Errorf("server %s: description is required", name)
		}
		if server.Transport == "" {
			return fmt.Errorf("server %s: transport is required", name)
		}
	}

	return nil
}
