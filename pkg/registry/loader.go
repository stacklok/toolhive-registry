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
			// Use directory name as the entry name
			entryName := info.Name()

			entry, err := l.LoadEntryWithName(specPath, entryName)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", specPath, err)
			}

			// Override with explicit name if set in the spec
			if entry.GetName() != "" && entry.GetName() != entryName {
				entryName = entry.GetName()
			} else {
				entry.SetName(entryName)
			}

			l.entries[entryName] = entry
		}

		return nil
	})

	return err
}

// LoadEntry loads a single registry entry from a YAML file without validation
// Use LoadEntryWithName for validation with proper naming
func (l *Loader) LoadEntry(path string) (*types.RegistryEntry, error) {
	return l.LoadEntryWithName(path, "")
}

// LoadEntryWithName loads a single registry entry from a YAML file with validation
func (l *Loader) LoadEntryWithName(path string, name string) (*types.RegistryEntry, error) {
	file, err := os.Open(path) // #nosec G304 - path is constructed from known directory structure
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

	// Validate with the actual name if provided
	if err := l.validateEntry(&entry, name); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &entry, nil
}

// validateEntry validates a registry entry using comprehensive schema-based validation
func (*Loader) validateEntry(entry *types.RegistryEntry, name string) error {
	// Use the new schema validator for comprehensive validation
	validator := NewSchemaValidator()

	return validator.ValidateComplete(entry, name)
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
		return entries[i].GetName() < entries[j].GetName()
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
