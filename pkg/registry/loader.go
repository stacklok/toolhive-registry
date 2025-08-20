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
			if entry.GetName() != "" {
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

// LoadEntry loads a single registry entry from a YAML file
func (l *Loader) LoadEntry(path string) (*types.RegistryEntry, error) {
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

	// Validate required fields
	if err := l.validateEntry(&entry); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &entry, nil
}

// validateEntry validates a registry entry
func (l *Loader) validateEntry(entry *types.RegistryEntry) error {
	// Check that we have either image or remote metadata
	if entry.ImageMetadata == nil && entry.RemoteServerMetadata == nil {
		return fmt.Errorf("entry must be either an image or remote server")
	}

	// Route to appropriate validator based on metadata presence
	if entry.ImageMetadata != nil {
		return l.validateImageEntry(entry)
	} else if entry.RemoteServerMetadata != nil {
		return l.validateRemoteEntry(entry)
	}

	return fmt.Errorf("unable to determine server type")
}

// validateImageEntry validates an image-based server entry
func (l *Loader) validateImageEntry(entry *types.RegistryEntry) error {
	if entry.ImageMetadata == nil {
		return fmt.Errorf("ImageMetadata is nil")
	}

	// Image-specific required fields
	if entry.Image == "" {
		return fmt.Errorf("image is required for image-based servers")
	}

	// Validate common fields
	return l.validateCommonFields(entry)
}

// validateRemoteEntry validates a remote server entry
func (l *Loader) validateRemoteEntry(entry *types.RegistryEntry) error {
	if entry.RemoteServerMetadata == nil {
		return fmt.Errorf("RemoteServerMetadata is nil")
	}

	// Remote-specific required fields
	if entry.URL == "" {
		return fmt.Errorf("url is required for remote servers")
	}

	// Remote servers cannot use stdio transport
	if entry.RemoteServerMetadata.Transport == "stdio" {
		return fmt.Errorf("remote servers cannot use stdio transport (use sse or streamable-http)")
	}

	// Validate common fields
	return l.validateCommonFields(entry)
}

// validateCommonFields validates fields common to both image and remote servers
func (l *Loader) validateCommonFields(entry *types.RegistryEntry) error {
	// Required fields
	if entry.GetDescription() == "" {
		return fmt.Errorf("description is required")
	}

	if entry.GetTransport() == "" {
		return fmt.Errorf("transport is required")
	}

	// Validate transport
	if err := l.validateTransport(entry.GetTransport()); err != nil {
		return err
	}

	// Validate tier if specified
	if err := l.validateTier(entry.GetTier()); err != nil {
		return err
	}

	// Validate status if specified
	return l.validateStatus(entry.GetStatus())
}

// validateTransport validates the transport type
func (*Loader) validateTransport(transport string) error {
	validTransports := map[string]bool{
		"stdio":           true,
		"sse":             true,
		"streamable-http": true,
	}

	if !validTransports[transport] {
		return fmt.Errorf("invalid transport: %s (must be stdio, sse, or streamable-http)", transport)
	}

	return nil
}

// validateTier validates the tier classification
func (*Loader) validateTier(tier string) error {
	if tier == "" {
		return nil // Tier is optional
	}

	validTiers := map[string]bool{
		"Official":  true,
		"Community": true,
	}

	if !validTiers[tier] {
		return fmt.Errorf("invalid tier: %s (must be Official or Community)", tier)
	}

	return nil
}

// validateStatus validates the status field
func (*Loader) validateStatus(status string) error {
	if status == "" {
		return nil // Status is optional
	}

	validStatuses := map[string]bool{
		"Active":     true,
		"Deprecated": true,
		"Beta":       true,
		"Alpha":      true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be Active, Deprecated, Beta, or Alpha)", status)
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

	// Validate image-based servers
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

	// Validate remote servers
	for name, server := range registry.RemoteServers {
		if server.URL == "" {
			return fmt.Errorf("remote server %s: url is required", name)
		}
		if server.Description == "" {
			return fmt.Errorf("remote server %s: description is required", name)
		}
		if server.Transport == "" {
			return fmt.Errorf("remote server %s: transport is required", name)
		}
		// Remote servers cannot use stdio
		if server.Transport == "stdio" {
			return fmt.Errorf("remote server %s: cannot use stdio transport", name)
		}
	}

	return nil
}
