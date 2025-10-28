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

	"github.com/stacklok/toolhive-registry/pkg/registry/converters"
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
	var serverJSON upstream.ServerJSON
	var err error

	// Use the converters package for the core conversion logic
	if entry.IsImage() {
		serverJSON, err = converters.ImageMetadataToServerJSON(name, entry.ImageMetadata)
		if err != nil {
			// This shouldn't happen with valid data, but handle it gracefully
			// Fall back to creating a minimal server entry
			serverJSON = or.createFallbackServerJSON(name, entry)
		}
	} else if entry.IsRemote() {
		serverJSON, err = converters.RemoteServerMetadataToServerJSON(name, entry.RemoteServerMetadata)
		if err != nil {
			// Fall back to creating a minimal server entry
			serverJSON = or.createFallbackServerJSON(name, entry)
		}
	} else {
		// Neither image nor remote - create a minimal entry
		serverJSON = or.createFallbackServerJSON(name, entry)
	}

	// Add additional ToolHive-specific extensions that aren't in base metadata
	// (permissions, args, examples, license, etc.)
	or.enhanceWithToolHiveExtensions(&serverJSON, entry)

	return serverJSON
}

// createRepository creates repository information from entry
func (*OfficialRegistry) createRepository(entry *types.RegistryEntry) model.Repository {
	var repositoryURL string

	if entry.IsImage() && entry.ImageMetadata.RepositoryURL != "" {
		repositoryURL = entry.ImageMetadata.RepositoryURL
	} else if entry.IsRemote() && entry.RemoteServerMetadata.RepositoryURL != "" {
		repositoryURL = entry.RemoteServerMetadata.RepositoryURL
	}

	if repositoryURL == "" {
		// Use a toolhive-registry placeholder URL to satisfy validation when no repository is available for remote servers
		repositoryURL = "https://github.com/stacklok/toolhive-registry"
		if entry.IsRemote() {
			return model.Repository{
				URL:    repositoryURL,
				Source: "github",
			}
		}
		return model.Repository{}
	}

	return model.Repository{
		URL:    repositoryURL,
		Source: "github", // Assume GitHub for now
	}
}

// createPackages creates Package entries for image-based servers
func (*OfficialRegistry) createPackages(entry *types.RegistryEntry) []model.Package {
	if !entry.IsImage() || entry.Image == "" {
		return nil
	}

	// Convert environment variables
	var envVars []model.KeyValueInput
	for _, envVar := range entry.ImageMetadata.EnvVars {
		envVars = append(envVars, model.KeyValueInput{
			Name: envVar.Name,
			InputWithVariables: model.InputWithVariables{
				Input: model.Input{
					Description: envVar.Description,
					IsRequired:  envVar.Required,
					IsSecret:    envVar.Secret,
					Default:     envVar.Default,
				},
			},
		})
	}

	// For OCI packages, use the full image reference in the identifier field
	// The version and registryBaseURL fields are not used for OCI packages
	// See: https://github.com/modelcontextprotocol/registry/blob/main/pkg/model/types.go
	identifier := entry.Image

	// Determine transport type - use entry's transport or default to stdio for containers
	transportType := entry.GetTransport()
	if transportType == "" {
		transportType = "stdio"
	}

	transport := model.Transport{
		Type: transportType,
	}

	// Add URL field for non-stdio transports (required by schema)
	if transportType == model.TransportTypeStreamableHTTP || transportType == model.TransportTypeSSE {
		// For container-based servers, construct URL template with target port
		port := 8080 // Default port if not specified
		if entry.ImageMetadata != nil && entry.TargetPort > 0 {
			port = entry.TargetPort
		}
		transport.URL = fmt.Sprintf("http://localhost:%d", port)
	}

	pkg := model.Package{
		RegistryType:         model.RegistryTypeOCI,
		Identifier:           identifier, // Full image reference including tag
		EnvironmentVariables: envVars,
		Transport:            transport,
		// Version and RegistryBaseURL are omitted for OCI packages
	}

	return []model.Package{pkg}
}

// createRemotes creates Transport entries for remote servers
func (*OfficialRegistry) createRemotes(entry *types.RegistryEntry) []model.Transport {
	if !entry.IsRemote() || entry.URL == "" {
		return nil
	}

	// Convert headers
	var headers []model.KeyValueInput
	for _, header := range entry.Headers {
		headers = append(headers, model.KeyValueInput{
			Name: header.Name,
			InputWithVariables: model.InputWithVariables{
				Input: model.Input{
					Description: header.Description,
					IsRequired:  header.Required,
					IsSecret:    header.Secret,
				},
			},
		})
	}

	remote := model.Transport{
		Type:    entry.GetTransport(),
		URL:     entry.URL,
		Headers: headers,
	}

	return []model.Transport{remote}
}

// createXPublisherExtensions creates x-publisher extensions with ToolHive-specific data
// Following the reverse DNS naming convention: io.github.stacklok
func (or *OfficialRegistry) createXPublisherExtensions(entry *types.RegistryEntry) map[string]interface{} {
	// Get the key for the ToolHive extensions (image or URL)
	var key string
	if entry.IsImage() {
		key = entry.Image
	} else if entry.IsRemote() {
		key = entry.URL
	} else {
		return map[string]interface{}{} // Empty if neither
	}

	// Create ToolHive-specific extensions
	toolhiveExtensions := or.createToolHiveExtensions(entry)

	// Use reverse DNS naming convention for vendor-specific data
	return map[string]interface{}{
		"io.github.stacklok": map[string]interface{}{
			key: toolhiveExtensions,
		},
	}
}

// createToolHiveExtensions creates the ToolHive-specific extension data
func (or *OfficialRegistry) createToolHiveExtensions(entry *types.RegistryEntry) map[string]interface{} {
	extensions := make(map[string]interface{})

	// Always include transport type
	extensions["transport"] = entry.GetTransport()

	// Add status (active/deprecated)
	extensions["status"] = string(or.convertStatus(entry.GetStatus()))

	// Add tools list
	if tools := entry.GetTools(); len(tools) > 0 {
		extensions["tools"] = tools
	}

	// Add tier
	if tier := entry.GetTier(); tier != "" {
		extensions["tier"] = tier
	}

	// Add common fields
	if entry.IsImage() {
		or.addImageSpecificExtensions(extensions, entry)
	} else if entry.IsRemote() {
		or.addRemoteSpecificExtensions(extensions, entry)
	}

	// Add common optional fields
	or.addCommonExtensions(extensions, entry)

	return extensions
}

// addImageSpecificExtensions adds image-specific ToolHive extensions
func (*OfficialRegistry) addImageSpecificExtensions(extensions map[string]interface{}, entry *types.RegistryEntry) {
	if entry.ImageMetadata == nil {
		return
	}

	// Add tags
	if len(entry.ImageMetadata.Tags) > 0 {
		extensions["tags"] = entry.ImageMetadata.Tags
	}

	// Add permissions
	if entry.Permissions != nil {
		extensions["permissions"] = entry.Permissions
	}

	// Add args (static container arguments)
	if len(entry.Args) > 0 {
		extensions["args"] = entry.Args
	}

	// Add metadata (stars, pulls, etc.)
	if entry.ImageMetadata.Metadata != nil {
		extensions["metadata"] = entry.ImageMetadata.Metadata
	}

	// Add provenance if present
	if entry.Provenance != nil {
		extensions["provenance"] = entry.Provenance
	}
}

// addRemoteSpecificExtensions adds remote-specific ToolHive extensions
func (*OfficialRegistry) addRemoteSpecificExtensions(extensions map[string]interface{}, entry *types.RegistryEntry) {
	if entry.RemoteServerMetadata == nil {
		return
	}

	// Add tags
	if len(entry.RemoteServerMetadata.Tags) > 0 {
		extensions["tags"] = entry.RemoteServerMetadata.Tags
	}

	// Add OAuth config
	if entry.OAuthConfig != nil {
		extensions["oauth_config"] = entry.OAuthConfig
	}

	// Add metadata
	if entry.RemoteServerMetadata.Metadata != nil {
		extensions["metadata"] = entry.RemoteServerMetadata.Metadata
	}
}

// addCommonExtensions adds extensions common to both image and remote servers
func (*OfficialRegistry) addCommonExtensions(extensions map[string]interface{}, entry *types.RegistryEntry) {
	// Add examples if present
	if len(entry.Examples) > 0 {
		extensions["examples"] = entry.Examples
	}

	// Add license if present
	if entry.License != "" {
		extensions["license"] = entry.License
	}
}

// convertStatus converts ToolHive status to MCP model.Status
func (*OfficialRegistry) convertStatus(status string) model.Status {
	switch status {
	case types.StatusActive, "":
		return model.StatusActive
	case types.StatusDeprecated:
		return model.StatusDeprecated
	default:
		return model.StatusActive // Default to active
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
