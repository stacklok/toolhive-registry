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
// This validates each server entry against the upstream MCP server schema,
// ensuring compatibility with the official MCP registry format
func (*OfficialRegistry) validateRegistry(registry *ToolHiveRegistryType) error {
	// Use the upstream schema URL directly from the registry package
	// This ensures we're always validating against the same schema version
	// that the code is built with, eliminating the need for manual schema syncing
	schemaLoader := gojsonschema.NewReferenceLoader(model.CurrentSchemaURL)

	// Validate each server individually against the upstream schema
	var allErrors []string
	for i, server := range registry.Data.Servers {
		// Marshal server to JSON
		serverJSON, err := json.Marshal(server)
		if err != nil {
			return fmt.Errorf("failed to marshal server %d: %w", i, err)
		}

		// Create document loader from server data
		documentLoader := gojsonschema.NewBytesLoader(serverJSON)

		// Perform validation
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return fmt.Errorf("schema validation failed for server %d (%s): %w", i, server.Name, err)
		}

		if !result.Valid() {
			for _, desc := range result.Errors() {
				allErrors = append(allErrors, fmt.Sprintf("data.servers.%d: %s", i, desc.String()))
			}
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("validation errors: %v", allErrors)
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
	// Create the flattened server JSON with _meta extensions
	serverJSON := upstream.ServerJSON{
		Schema:      model.CurrentSchemaURL,
		Name:        or.convertNameToReverseDNS(name),
		Description: entry.GetDescription(),
		Repository:  or.createRepository(entry),
		Version:     "1.0.0", // TODO: Default server version for now, fix this to use package/remote version
		Meta: &upstream.ServerMeta{
			PublisherProvided: or.createXPublisherExtensions(entry),
		},
	}

	// Add packages for image-based servers
	if entry.IsImage() {
		serverJSON.Packages = or.createPackages(entry)
	}

	// Add remotes for remote servers
	if entry.IsRemote() {
		serverJSON.Remotes = or.createRemotes(entry)
	}

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
		if entry.ImageMetadata != nil && entry.ImageMetadata.TargetPort > 0 {
			port = entry.ImageMetadata.TargetPort
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

	return map[string]interface{}{
		"toolhive": map[string]interface{}{
			key: toolhiveExtensions,
		},
	}
}

// createToolHiveExtensions creates the ToolHive-specific extension data
func (or *OfficialRegistry) createToolHiveExtensions(entry *types.RegistryEntry) map[string]interface{} {
	extensions := make(map[string]interface{})

	// Always include transport type
	extensions["transport"] = entry.GetTransport()

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

// parseImageReference parses a container image reference into basic components
// Returns error if registry has a port (not supported)
func parseImageReference(image string) (registryBaseURL, identifier, version string, err error) {
	// Check for port in registry (not supported)
	if strings.Contains(image, ":") && strings.Count(image, ":") > 1 {
		// Multiple colons might indicate registry:port/image:tag
		parts := strings.Split(image, "/")
		if len(parts) > 0 && strings.Contains(parts[0], ":") {
			// First part has colon, likely registry:port
			return "", "", "", fmt.Errorf("registry with port not supported: %s", parts[0])
		}
	}

	// Handle digest (@sha256:...)
	if strings.Contains(image, "@") {
		parts := strings.SplitN(image, "@", 2)
		imageRef := parts[0]
		digest := parts[1]

		reg, name := splitRegistryAndName(imageRef)
		return reg, name, digest, nil
	}

	// Handle tag (:tag)
	if strings.Contains(image, ":") {
		parts := strings.SplitN(image, ":", 2)
		imageRef := parts[0]
		tag := parts[1]

		reg, name := splitRegistryAndName(imageRef)
		return reg, name, tag, nil
	}

	// No tag or digest - default to latest
	reg, name := splitRegistryAndName(image)
	return reg, name, "latest", nil
}

// splitRegistryAndName splits image into registry and name parts
func splitRegistryAndName(image string) (registryBaseURL, identifier string) {
	// No slash = Docker Hub image
	if !strings.Contains(image, "/") {
		return "https://docker.io", image
	}

	// Has slash - check if first part looks like registry
	parts := strings.SplitN(image, "/", 2)
	firstPart := parts[0]

	// If first part has dot, assume it's a registry hostname
	if strings.Contains(firstPart, ".") {
		return "https://" + firstPart, parts[1]
	}

	// Otherwise assume Docker Hub with namespace
	return "https://docker.io", image
}

// convertNameToReverseDNS converts simple server names to reverse-DNS format required by v1.0.0 schema
func (*OfficialRegistry) convertNameToReverseDNS(name string) string {
	// If already in reverse-DNS format (contains '/'), return as-is
	if strings.Contains(name, "/") {
		return name
	}

	// Convert simple names to toolhive namespace format
	return "io.stacklok.toolhive/" + name
}
