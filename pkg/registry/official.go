package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stacklok/toolhive/pkg/registry/converters"
	"github.com/stacklok/toolhive/pkg/registry/registry"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

// OfficialRegistry handles building and writing the toolhive MCP registry based on the official server format
type OfficialRegistry struct {
	loader *Loader
}

// NewOfficialRegistry creates a new instance of the official registry
func NewOfficialRegistry(loader *Loader) *OfficialRegistry {
	return &OfficialRegistry{
		loader: loader,
	}
}

// WriteJSON builds the official MCP registry and writes it to the specified path
// Individual entries and the complete registry are validated before writing - generation fails if validation fails
func (or *OfficialRegistry) WriteJSON(path string) error {
	// Validate all entries first
	if err := or.validateEntries(); err != nil {
		return fmt.Errorf("entry validation failed: %w", err)
	}

	// Build the registry structure
	builtRegistry := or.build()

	// Validate the complete registry against schema
	if err := or.validateRegistry(builtRegistry); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(builtRegistry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ValidateAgainstSchema validates the built registry against the schema
func (or *OfficialRegistry) ValidateAgainstSchema() error {
	builtRegistry := or.build()
	return or.validateRegistry(builtRegistry)
}

// validateRegistry validates a registry object against the schema
// Uses toolhive's validation which validates the entire structure including all servers
func (*OfficialRegistry) validateRegistry(upstreamRegistry *registry.UpstreamRegistry) error {
	// Serialize to JSON for validation
	registryJSON, err := json.Marshal(upstreamRegistry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Use toolhive's built-in validation for UpstreamRegistry
	// This validates the complete structure including all servers via $ref
	if err := toolhiveRegistry.ValidateUpstreamRegistry(registryJSON); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}

// validateEntries validates all individual registry entries
func (or *OfficialRegistry) validateEntries() error {
	entries := or.loader.GetEntries()
	validator := NewSchemaValidator()

	for name, entry := range entries {
		if err := validator.ValidateEntryFields(entry, name); err != nil {
			return fmt.Errorf("entry '%s' validation failed: %w", name, err)
		}
	}

	return nil
}

// build creates the UpstreamRegistry structure from loaded entries
func (or *OfficialRegistry) build() *registry.UpstreamRegistry {
	entries := or.loader.GetEntries()

	// Get sorted entry names for consistent output
	var names []string
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	// Transform entries to upstream.ServerJSON
	var servers []upstream.ServerJSON
	for _, name := range names {
		entry := entries[name]
		serverJSON := or.transformEntry(name, entry)
		servers = append(servers, serverJSON)
	}

	upstreamRegistry := &registry.UpstreamRegistry{
		Schema:  "https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/upstream-registry.schema.json",
		Version: "1.0.0",
		Meta: registry.UpstreamMeta{
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
		Data: registry.UpstreamData{
			Servers: servers,
			Groups:  []registry.UpstreamGroup{}, // Empty for now, placeholder for future use
		},
	}

	return upstreamRegistry
}

// transformEntry converts a ToolHive RegistryEntry to an official MCP ServerJSON
func (or *OfficialRegistry) transformEntry(name string, entry *types.RegistryEntry) upstream.ServerJSON {
	var serverJSONPtr *upstream.ServerJSON
	var err error

	// Use the converters package for all conversion logic
	if entry.IsImage() {
		serverJSONPtr, err = converters.ImageMetadataToServerJSON(name, entry.ImageMetadata)
		if err != nil || serverJSONPtr == nil {
			// This shouldn't happen with valid data, but handle it gracefully
			// Fall back to creating a minimal server entry
			fallback := or.createFallbackServerJSON(name, entry)
			return fallback
		}
	} else if entry.IsRemote() {
		serverJSONPtr, err = converters.RemoteServerMetadataToServerJSON(name, entry.RemoteServerMetadata)
		if err != nil || serverJSONPtr == nil {
			// Fall back to creating a minimal server entry
			fallback := or.createFallbackServerJSON(name, entry)
			return fallback
		}
	} else {
		// Neither image nor remote - create a minimal entry
		fallback := or.createFallbackServerJSON(name, entry)
		return fallback
	}

	// Convert name to reverse-DNS format as required by the upstream registry schema
	// The converters use the name from metadata, so we need to update it after conversion
	serverJSON := *serverJSONPtr
	serverJSON.Name = or.convertNameToReverseDNS(serverJSON.Name)
	return serverJSON
}

// createRepository creates repository information from entry
func (*OfficialRegistry) createRepository(entry *types.RegistryEntry) *model.Repository {
	var repositoryURL string

	if entry.IsImage() && entry.ImageMetadata.RepositoryURL != "" {
		repositoryURL = entry.ImageMetadata.RepositoryURL
	} else if entry.IsRemote() && entry.RemoteServerMetadata.RepositoryURL != "" {
		repositoryURL = entry.RemoteServerMetadata.RepositoryURL
	}

	// If no repository URL is available, return nil (will be omitted with omitempty)
	if repositoryURL == "" {
		return nil
	}

	return &model.Repository{
		URL:    repositoryURL,
		Source: "github", // Assume GitHub for now
	}
}

// convertNameToReverseDNS converts simple server names to reverse-DNS format required by v1.0.0 schema
func (*OfficialRegistry) convertNameToReverseDNS(name string) string {
	// If already in reverse-DNS format (contains '/'), return as-is
	if strings.Contains(name, "/") {
		return name
	}

	// Convert simple names to GitHub-based namespace format
	return "io.github.stacklok/" + name
}

// createFallbackServerJSON creates a minimal ServerJSON when conversion fails
func (or *OfficialRegistry) createFallbackServerJSON(name string, entry *types.RegistryEntry) upstream.ServerJSON {
	return upstream.ServerJSON{
		Schema:      model.CurrentSchemaURL,
		Name:        or.convertNameToReverseDNS(name),
		Description: entry.GetDescription(),
		Version:     "1.0.0",
		Repository:  or.createRepository(entry),
	}
}
