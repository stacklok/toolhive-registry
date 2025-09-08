package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
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

	// Validate the complete registry against schema
	if err := or.validateRegistry(registry); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
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
func (*OfficialRegistry) validateRegistry(registry *ToolHiveRegistryType) error {
	// Marshal registry to JSON
	registryJSON, err := json.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Load schema from local file (fallback to remote if needed)
	schemaPath := "schemas/registry.schema.json"
	var schemaLoader gojsonschema.JSONLoader

	// Try local schema first
	if _, err := os.Stat(schemaPath); err == nil {
		schemaLoader = gojsonschema.NewReferenceLoader("file://" + schemaPath)
	} else {
		// Fall back to remote schema
		schemaLoader = gojsonschema.NewReferenceLoader(
			"https://raw.githubusercontent.com/stacklok/toolhive-registry/main/schemas/registry.schema.json")
	}

	// Create document loader from registry data
	documentLoader := gojsonschema.NewBytesLoader(registryJSON)

	// Perform validation
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errorMessages []string
		for _, desc := range result.Errors() {
			errorMessages = append(errorMessages, desc.String())
		}
		return fmt.Errorf("validation errors: %v", errorMessages)
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
		Name:          name,
		Description:   entry.GetDescription(),
		Status:        or.convertStatus(entry.GetStatus()),
		Repository:    or.createRepository(entry),
		VersionDetail: or.createVersionDetail(),
		Meta: &upstream.ServerMeta{
			PublisherProvided: or.createXPublisherExtensions(entry),
			// The registry extensions are not supposed to be set by us.
			// They are generated by the registry system.
			// We include them here so we can start using them in toolhive,
			// and they are available when we support an official MCP registry.
			Official: or.createRegistryExtensions(),
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

// createVersionDetail creates version information (fixed at 1.0.0 for now)
func (*OfficialRegistry) createVersionDetail() model.VersionDetail {
	return model.VersionDetail{
		Version: "1.0.0",
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

	// Extract registry and version information from the image reference
	registryBaseURL, identifier, version, err := parseImageReference(entry.Image)
	if err != nil {
		// Continue with fallback values
		registryBaseURL = ""
		identifier = entry.Image
		version = ""
	}

	pkg := model.Package{
		RegistryType:         model.RegistryTypeOCI,
		RegistryBaseURL:      registryBaseURL,
		Identifier:           identifier,
		Version:              version,
		EnvironmentVariables: envVars,
	}

	return []model.Package{pkg}
}

// createRemotes creates Remote entries for remote servers
func (*OfficialRegistry) createRemotes(entry *types.RegistryEntry) []model.Remote {
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

	remote := model.Remote{
		TransportType: entry.GetTransport(),
		URL:           entry.URL,
		Headers:       headers,
	}

	return []model.Remote{remote}
}

// createRegistryExtensions creates registry-generated metadata
func (*OfficialRegistry) createRegistryExtensions() *upstream.RegistryExtensions {
	now := time.Now().UTC()
	return &upstream.RegistryExtensions{
		ID:          uuid.NewString(),
		PublishedAt: now,
		UpdatedAt:   now,
		IsLatest:    true,
		ReleaseDate: now.Format("2006-01-02"),
	}
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
