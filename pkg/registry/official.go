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
	"github.com/xeipuuv/gojsonschema"

	"github.com/stacklok/toolhive/pkg/registry/converters"
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
	registry := or.build()

	// Validate the complete registry against schema (warnings only for now)
	if err := or.validateRegistry(registry); err != nil {
		fmt.Printf("⚠️  Schema validation warnings: %v\n", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(registry, "", "  ")
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
	registry := or.build()
	return or.validateRegistry(registry)
}

// validateRegistry validates a registry object against the schema
// This validates both the wrapper structure and each server entry
func (or *OfficialRegistry) validateRegistry(registry *ToolHiveRegistryType) error {
	var allErrors []string

	// Step 1: Validate the wrapper structure against the ToolHive registry schema
	registryJSON, err := json.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Load the wrapper schema from local file
	wrapperSchemaPath := "schemas/registry.schema.json"
	wrapperSchemaLoader := gojsonschema.NewReferenceLoader("file://" + wrapperSchemaPath)

	wrapperLoader := gojsonschema.NewBytesLoader(registryJSON)
	wrapperResult, err := gojsonschema.Validate(wrapperSchemaLoader, wrapperLoader)
	if err != nil {
		return fmt.Errorf("wrapper schema validation failed: %w", err)
	}

	if !wrapperResult.Valid() {
		for _, desc := range wrapperResult.Errors() {
			allErrors = append(allErrors, fmt.Sprintf("wrapper: %s", desc.String()))
		}
	}

	// Step 2: Validate each server individually against the upstream MCP server schema
	if err := or.validateServers(registry.Data.Servers, &allErrors); err != nil {
		return err
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("validation errors: %v", allErrors)
	}

	return nil
}

// validateServers validates each server entry against the upstream MCP server schema
// This function can be used standalone to validate individual servers
func (*OfficialRegistry) validateServers(servers []upstream.ServerJSON, allErrors *[]string) error {
	// Use the upstream schema URL directly from the registry package
	// This ensures we're always validating against the same schema version
	// that the code is built with, eliminating the need for manual schema syncing
	serverSchemaLoader := gojsonschema.NewReferenceLoader(model.CurrentSchemaURL)

	for i, server := range servers {
		// Marshal server to JSON
		serverJSON, err := json.Marshal(server)
		if err != nil {
			return fmt.Errorf("failed to marshal server %d: %w", i, err)
		}

		// Create document loader from server data
		documentLoader := gojsonschema.NewBytesLoader(serverJSON)

		// Perform validation
		result, err := gojsonschema.Validate(serverSchemaLoader, documentLoader)
		if err != nil {
			return fmt.Errorf("schema validation failed for server %d (%s): %w", i, server.Name, err)
		}

		if !result.Valid() {
			for _, desc := range result.Errors() {
				*allErrors = append(*allErrors, fmt.Sprintf("data.servers.%d: %s", i, desc.String()))
			}
		}
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

// build creates the ToolHiveRegistryType structure from loaded entries
func (or *OfficialRegistry) build() *ToolHiveRegistryType {
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

	registry := &ToolHiveRegistryType{
		Schema:  "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/schemas/registry.schema.json",
		Version: "1.0.0",
		Meta: Meta{
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
		Data: Data{
			Servers: servers,
			Groups:  []Group{}, // Empty for now, placeholder for future use
		},
	}

	return registry
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

	return *serverJSONPtr
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
