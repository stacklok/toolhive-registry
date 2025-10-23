// Package registry provides conversion functions between upstream MCP ServerJSON format
// and toolhive ImageMetadata/RemoteServerMetadata formats.
package registry

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
	"github.com/stacklok/toolhive/pkg/registry"
)

// ServerJSONToImageMetadata converts an upstream ServerJSON (with OCI packages) to toolhive ImageMetadata
// This function only handles OCI packages and will error if there are multiple OCI packages
func ServerJSONToImageMetadata(serverJSON *upstream.ServerJSON) (*registry.ImageMetadata, error) {
	if serverJSON == nil {
		return nil, fmt.Errorf("serverJSON cannot be nil")
	}

	if len(serverJSON.Packages) == 0 {
		return nil, fmt.Errorf("serverJSON has no packages (not a container-based server)")
	}

	// Filter for OCI packages only
	var ociPackages []model.Package
	for _, pkg := range serverJSON.Packages {
		if pkg.RegistryType == model.RegistryTypeOCI {
			ociPackages = append(ociPackages, pkg)
		}
	}

	if len(ociPackages) == 0 {
		return nil, fmt.Errorf("serverJSON has no OCI packages")
	}

	if len(ociPackages) > 1 {
		return nil, fmt.Errorf("serverJSON has %d OCI packages, expected exactly 1", len(ociPackages))
	}

	pkg := ociPackages[0]

	imageMetadata := &registry.ImageMetadata{
		BaseServerMetadata: registry.BaseServerMetadata{
			Description: serverJSON.Description,
			Transport:   pkg.Transport.Type,
		},
		Image: pkg.Identifier, // OCI packages store full image ref in Identifier
	}

	// Set repository URL
	if serverJSON.Repository.URL != "" {
		imageMetadata.RepositoryURL = serverJSON.Repository.URL
	}

	// Convert environment variables
	if len(pkg.EnvironmentVariables) > 0 {
		imageMetadata.EnvVars = make([]*registry.EnvVar, 0, len(pkg.EnvironmentVariables))
		for _, envVar := range pkg.EnvironmentVariables {
			imageMetadata.EnvVars = append(imageMetadata.EnvVars, &registry.EnvVar{
				Name:        envVar.Name,
				Description: envVar.Description,
				Required:    envVar.IsRequired,
				Secret:      envVar.IsSecret,
				Default:     envVar.Default,
			})
		}
	}

	// Extract target port from transport URL if present
	if pkg.Transport.URL != "" {
		// Parse URL like "http://localhost:8080"
		parsedURL, err := url.Parse(pkg.Transport.URL)
		if err == nil && parsedURL.Port() != "" {
			if port, err := strconv.Atoi(parsedURL.Port()); err == nil {
				imageMetadata.TargetPort = port
			}
		}
	}

	// Extract publisher-provided extensions
	extractImageExtensions(serverJSON, imageMetadata)

	return imageMetadata, nil
}

// ImageMetadataToServerJSON converts toolhive ImageMetadata to an upstream ServerJSON
// The name parameter should be the simple server name (e.g., "fetch")
func ImageMetadataToServerJSON(name string, imageMetadata *registry.ImageMetadata) (*upstream.ServerJSON, error) {
	if imageMetadata == nil {
		return nil, fmt.Errorf("imageMetadata cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Create ServerJSON with basic fields
	serverJSON := &upstream.ServerJSON{
		Schema:      model.CurrentSchemaURL,
		Name:        BuildReverseDNSName(name),
		Description: imageMetadata.Description,
		Version:     "1.0.0", // TODO: Extract from image tag or metadata
	}

	// Set repository
	if imageMetadata.RepositoryURL != "" {
		serverJSON.Repository = model.Repository{
			URL:    imageMetadata.RepositoryURL,
			Source: "github", // Assume GitHub
		}
	}

	// Create package
	serverJSON.Packages = createPackagesFromImageMetadata(imageMetadata)

	// Create publisher extensions
	serverJSON.Meta = &upstream.ServerMeta{
		PublisherProvided: createImageExtensions(imageMetadata),
	}

	return serverJSON, nil
}

// ServerJSONToRemoteServerMetadata converts an upstream ServerJSON (with remotes) to toolhive RemoteServerMetadata
// This function extracts remote server data and reconstructs RemoteServerMetadata format
func ServerJSONToRemoteServerMetadata(serverJSON *upstream.ServerJSON) (*registry.RemoteServerMetadata, error) {
	if serverJSON == nil {
		return nil, fmt.Errorf("serverJSON cannot be nil")
	}

	if len(serverJSON.Remotes) == 0 {
		return nil, fmt.Errorf("serverJSON has no remotes (not a remote server)")
	}

	remote := serverJSON.Remotes[0] // Use first remote

	remoteMetadata := &registry.RemoteServerMetadata{
		BaseServerMetadata: registry.BaseServerMetadata{
			Description: serverJSON.Description,
			Transport:   remote.Type,
		},
		URL: remote.URL,
	}

	// Set repository URL
	if serverJSON.Repository.URL != "" {
		remoteMetadata.RepositoryURL = serverJSON.Repository.URL
	}

	// Convert headers
	if len(remote.Headers) > 0 {
		remoteMetadata.Headers = make([]*registry.Header, 0, len(remote.Headers))
		for _, header := range remote.Headers {
			remoteMetadata.Headers = append(remoteMetadata.Headers, &registry.Header{
				Name:        header.Name,
				Description: header.Description,
				Required:    header.IsRequired,
				Secret:      header.IsSecret,
			})
		}
	}

	// Extract publisher-provided extensions
	extractRemoteExtensions(serverJSON, remoteMetadata)

	return remoteMetadata, nil
}

// RemoteServerMetadataToServerJSON converts toolhive RemoteServerMetadata to an upstream ServerJSON
// The name parameter should be the simple server name (e.g., "github-remote")
func RemoteServerMetadataToServerJSON(name string, remoteMetadata *registry.RemoteServerMetadata) (*upstream.ServerJSON, error) {
	if remoteMetadata == nil {
		return nil, fmt.Errorf("remoteMetadata cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Create ServerJSON with basic fields
	serverJSON := &upstream.ServerJSON{
		Schema:      model.CurrentSchemaURL,
		Name:        BuildReverseDNSName(name),
		Description: remoteMetadata.Description,
		Version:     "1.0.0", // TODO: Version management
	}

	// Set repository
	if remoteMetadata.RepositoryURL != "" {
		serverJSON.Repository = model.Repository{
			URL:    remoteMetadata.RepositoryURL,
			Source: "github", // Assume GitHub
		}
	}

	// Create remote
	serverJSON.Remotes = createRemotesFromRemoteMetadata(remoteMetadata)

	// Create publisher extensions
	serverJSON.Meta = &upstream.ServerMeta{
		PublisherProvided: createRemoteExtensions(remoteMetadata),
	}

	return serverJSON, nil
}

// Helper functions

// createPackagesFromImageMetadata creates OCI Package entries from ImageMetadata
func createPackagesFromImageMetadata(imageMetadata *registry.ImageMetadata) []model.Package {
	// Convert environment variables
	var envVars []model.KeyValueInput
	for _, envVar := range imageMetadata.EnvVars {
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

	// Determine transport
	transportType := imageMetadata.Transport
	if transportType == "" {
		transportType = model.TransportTypeStdio
	}

	transport := model.Transport{
		Type: transportType,
	}

	// Add URL for non-stdio transports
	if transportType == model.TransportTypeStreamableHTTP || transportType == model.TransportTypeSSE {
		port := 8080
		if imageMetadata.TargetPort > 0 {
			port = imageMetadata.TargetPort
		}
		transport.URL = fmt.Sprintf("http://localhost:%d", port)
	}

	return []model.Package{{
		RegistryType:         model.RegistryTypeOCI,
		Identifier:           imageMetadata.Image,
		EnvironmentVariables: envVars,
		Transport:            transport,
	}}
}

// createRemotesFromRemoteMetadata creates Transport entries from RemoteServerMetadata
func createRemotesFromRemoteMetadata(remoteMetadata *registry.RemoteServerMetadata) []model.Transport {
	// Convert headers
	var headers []model.KeyValueInput
	for _, header := range remoteMetadata.Headers {
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

	return []model.Transport{{
		Type:    remoteMetadata.Transport,
		URL:     remoteMetadata.URL,
		Headers: headers,
	}}
}

// createImageExtensions creates publisher extensions map from ImageMetadata
func createImageExtensions(imageMetadata *registry.ImageMetadata) map[string]interface{} {
	extensions := make(map[string]interface{})

	// Always include transport and status
	extensions["transport"] = imageMetadata.Transport
	extensions["status"] = imageMetadata.Status
	if extensions["status"] == "" {
		extensions["status"] = "active"
	}

	// Add tools
	if len(imageMetadata.Tools) > 0 {
		extensions["tools"] = imageMetadata.Tools
	}

	// Add tier
	if imageMetadata.Tier != "" {
		extensions["tier"] = imageMetadata.Tier
	}

	// Add tags
	if len(imageMetadata.Tags) > 0 {
		extensions["tags"] = imageMetadata.Tags
	}

	// Add metadata
	if imageMetadata.Metadata != nil {
		extensions["metadata"] = map[string]interface{}{
			"stars":        imageMetadata.Metadata.Stars,
			"pulls":        imageMetadata.Metadata.Pulls,
			"last_updated": imageMetadata.Metadata.LastUpdated,
		}
	}

	return map[string]interface{}{
		"io.github.stacklok": map[string]interface{}{
			imageMetadata.Image: extensions,
		},
	}
}

// createRemoteExtensions creates publisher extensions map from RemoteServerMetadata
func createRemoteExtensions(remoteMetadata *registry.RemoteServerMetadata) map[string]interface{} {
	extensions := make(map[string]interface{})

	// Always include transport and status
	extensions["transport"] = remoteMetadata.Transport
	extensions["status"] = remoteMetadata.Status
	if extensions["status"] == "" {
		extensions["status"] = "active"
	}

	// Add tools
	if len(remoteMetadata.Tools) > 0 {
		extensions["tools"] = remoteMetadata.Tools
	}

	// Add tier
	if remoteMetadata.Tier != "" {
		extensions["tier"] = remoteMetadata.Tier
	}

	// Add tags
	if len(remoteMetadata.Tags) > 0 {
		extensions["tags"] = remoteMetadata.Tags
	}

	// Add metadata
	if remoteMetadata.Metadata != nil {
		extensions["metadata"] = map[string]interface{}{
			"stars":        remoteMetadata.Metadata.Stars,
			"pulls":        remoteMetadata.Metadata.Pulls,
			"last_updated": remoteMetadata.Metadata.LastUpdated,
		}
	}

	return map[string]interface{}{
		"io.github.stacklok": map[string]interface{}{
			remoteMetadata.URL: extensions,
		},
	}
}

// extractImageExtensions extracts publisher-provided extensions into ImageMetadata
func extractImageExtensions(serverJSON *upstream.ServerJSON, imageMetadata *registry.ImageMetadata) {
	if serverJSON.Meta == nil || serverJSON.Meta.PublisherProvided == nil {
		return
	}

	stacklokData, ok := serverJSON.Meta.PublisherProvided["io.github.stacklok"].(map[string]interface{})
	if !ok {
		return
	}

	// Find the extension data (keyed by image reference)
	for _, extensionsData := range stacklokData {
		extensions, ok := extensionsData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract fields
		if status, ok := extensions["status"].(string); ok {
			imageMetadata.Status = status
		}
		if tier, ok := extensions["tier"].(string); ok {
			imageMetadata.Tier = tier
		}
		if toolsData, ok := extensions["tools"].([]interface{}); ok {
			imageMetadata.Tools = interfaceSliceToStringSlice(toolsData)
		}
		if tagsData, ok := extensions["tags"].([]interface{}); ok {
			imageMetadata.Tags = interfaceSliceToStringSlice(tagsData)
		}
		if metadataData, ok := extensions["metadata"].(map[string]interface{}); ok {
			imageMetadata.Metadata = &registry.Metadata{}
			if stars, ok := metadataData["stars"].(float64); ok {
				imageMetadata.Metadata.Stars = int(stars)
			}
			if pulls, ok := metadataData["pulls"].(float64); ok {
				imageMetadata.Metadata.Pulls = int(pulls)
			}
			if lastUpdated, ok := metadataData["last_updated"].(string); ok {
				imageMetadata.Metadata.LastUpdated = lastUpdated
			}
		}

		break // Only process first entry
	}
}

// extractRemoteExtensions extracts publisher-provided extensions into RemoteServerMetadata
func extractRemoteExtensions(serverJSON *upstream.ServerJSON, remoteMetadata *registry.RemoteServerMetadata) {
	if serverJSON.Meta == nil || serverJSON.Meta.PublisherProvided == nil {
		return
	}

	stacklokData, ok := serverJSON.Meta.PublisherProvided["io.github.stacklok"].(map[string]interface{})
	if !ok {
		return
	}

	// Find the extension data (keyed by URL)
	for _, extensionsData := range stacklokData {
		extensions, ok := extensionsData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract fields
		if status, ok := extensions["status"].(string); ok {
			remoteMetadata.Status = status
		}
		if tier, ok := extensions["tier"].(string); ok {
			remoteMetadata.Tier = tier
		}
		if toolsData, ok := extensions["tools"].([]interface{}); ok {
			remoteMetadata.Tools = interfaceSliceToStringSlice(toolsData)
		}
		if tagsData, ok := extensions["tags"].([]interface{}); ok {
			remoteMetadata.Tags = interfaceSliceToStringSlice(tagsData)
		}
		if metadataData, ok := extensions["metadata"].(map[string]interface{}); ok {
			remoteMetadata.Metadata = &registry.Metadata{}
			if stars, ok := metadataData["stars"].(float64); ok {
				remoteMetadata.Metadata.Stars = int(stars)
			}
			if pulls, ok := metadataData["pulls"].(float64); ok {
				remoteMetadata.Metadata.Pulls = int(pulls)
			}
			if lastUpdated, ok := metadataData["last_updated"].(string); ok {
				remoteMetadata.Metadata.LastUpdated = lastUpdated
			}
		}

		break // Only process first entry
	}
}

// Utility functions

// interfaceSliceToStringSlice converts []interface{} to []string
func interfaceSliceToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, item := range input {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// ExtractServerName extracts the simple server name from a reverse-DNS format name
// Example: "io.github.stacklok/fetch" -> "fetch"
func ExtractServerName(reverseDNSName string) string {
	parts := strings.Split(reverseDNSName, "/")
	if len(parts) == 2 {
		return parts[1]
	}
	return reverseDNSName
}

// BuildReverseDNSName builds a reverse-DNS format name from a simple name
// Example: "fetch" -> "io.github.stacklok/fetch"
func BuildReverseDNSName(simpleName string) string {
	if strings.Contains(simpleName, "/") {
		return simpleName // Already in reverse-DNS format
	}
	return "io.github.stacklok/" + simpleName
}
