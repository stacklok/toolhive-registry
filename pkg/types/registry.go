// Package types provides extended types for the toolhive registry
package types

import (
	"fmt"
	"time"

	"github.com/stacklok/toolhive/pkg/registry"
)

const (
	// StatusActive indicates the server is actively maintained
	StatusActive = "Active"
	// StatusDeprecated indicates the server is deprecated
	StatusDeprecated = "Deprecated"

	// TierCommunity indicates the server is community-supported
	TierCommunity = "Community"

	// TierOfficial indicates the server is officially supported
	TierOfficial = "Official"
)

// RegistryEntry is a unified type that can represent either an image-based or remote MCP server
// It embeds either ImageMetadata or RemoteServerMetadata from toolhive based on what's in the spec.yaml
type RegistryEntry struct {
	// Embed the original ImageMetadata from toolhive (for image-based servers)
	*registry.ImageMetadata `yaml:",inline"`

	// Embed the RemoteServerMetadata from toolhive (for remote servers)
	*registry.RemoteServerMetadata `yaml:",inline"`

	// Extended fields for the registry (applies to both types)
	Examples []Example `yaml:"examples,omitempty"`
	License  string    `yaml:"license,omitempty"`
}

// GetServerMetadata returns the underlying ServerMetadata interface
// This allows unified access to common fields regardless of server type
func (r *RegistryEntry) GetServerMetadata() registry.ServerMetadata {
	if r.IsImage() {
		return r.ImageMetadata
	}
	if r.IsRemote() {
		return r.RemoteServerMetadata
	}
	return nil
}

// Example provides usage examples
type Example struct {
	// Name of the example
	Name string `yaml:"name"`

	// Description of what the example does
	Description string `yaml:"description"`

	// Sample usage string. This is a multi-line string that provides an example of how to use the registry entry.
	Sample string `yaml:"sample"`
}

// RegistryMetadata contains metadata about the entire registry
type RegistryMetadata struct {
	// Version of the registry format
	Version string `yaml:"version"`

	// When the registry was last updated
	LastUpdated time.Time `yaml:"last_updated"`
}

// IsRemote returns true if this is a remote server entry
func (r *RegistryEntry) IsRemote() bool {
	// A remote server has a URL field instead of an Image field
	return r.RemoteServerMetadata != nil && r.URL != ""
}

// IsImage returns true if this is an image-based server entry
func (r *RegistryEntry) IsImage() bool {
	// An image server has an Image field
	return r.ImageMetadata != nil && r.Image != ""
}

// GetName returns the name of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetName() string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetName()
	}
	return ""
}

// GetDescription returns the description of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetDescription() string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetDescription()
	}
	return ""
}

// GetTransport returns the transport of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetTransport() string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetTransport()
	}
	return ""
}

// GetTier returns the tier of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetTier() string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetTier()
	}
	return ""
}

// GetStatus returns the status of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetStatus() string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetStatus()
	}
	return ""
}

// GetTools returns the tools of the entry using the ServerMetadata interface
func (r *RegistryEntry) GetTools() []string {
	if metadata := r.GetServerMetadata(); metadata != nil {
		return metadata.GetTools()
	}
	return nil
}

// SetName sets the name on the appropriate metadata type
func (r *RegistryEntry) SetName(name string) {
	if r.ImageMetadata != nil {
		r.ImageMetadata.Name = name
	}
	if r.RemoteServerMetadata != nil {
		r.RemoteServerMetadata.Name = name
	}
}

// SetDefaults sets default values for tier and status if not specified
func (r *RegistryEntry) SetDefaults() {
	if r.ImageMetadata != nil {
		if r.ImageMetadata.Tier == "" {
			r.ImageMetadata.Tier = TierCommunity
		}
		if r.ImageMetadata.Status == "" {
			r.ImageMetadata.Status = StatusActive
		}
	}
	if r.RemoteServerMetadata != nil {
		if r.RemoteServerMetadata.Tier == "" {
			r.RemoteServerMetadata.Tier = TierCommunity
		}
		if r.RemoteServerMetadata.Status == "" {
			r.RemoteServerMetadata.Status = StatusActive
		}
	}
}

// UnmarshalYAML implements custom YAML unmarshaling to determine server type
func (r *RegistryEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First unmarshal into a map to check which fields are present
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Check for image vs url to determine type
	_, hasImage := raw["image"]
	_, hasURL := raw["url"]

	if hasImage && hasURL {
		return fmt.Errorf("entry cannot have both 'image' and 'url' fields")
	}

	if !hasImage && !hasURL {
		return fmt.Errorf("entry must have either 'image' or 'url' field")
	}

	if hasImage {
		// For image-based servers, unmarshal into ImageMetadata
		r.ImageMetadata = &registry.ImageMetadata{}
		if err := unmarshal(r.ImageMetadata); err != nil {
			return err
		}
	} else {
		// For remote servers, unmarshal into RemoteServerMetadata
		r.RemoteServerMetadata = &registry.RemoteServerMetadata{}
		if err := unmarshal(r.RemoteServerMetadata); err != nil {
			return err
		}
	}

	// Unmarshal extended fields (examples, license, oauth, headers, env_vars) separately
	type extendedFields struct {
		Examples []Example `yaml:"examples,omitempty"`
		License  string    `yaml:"license,omitempty"`
		// OAuth configuration in simplified YAML format
		OAuth *struct {
			Issuer         string            `yaml:"issuer,omitempty"`
			AuthorizeURL   string            `yaml:"authorize_url,omitempty"`
			TokenURL       string            `yaml:"token_url,omitempty"`
			ClientID       string            `yaml:"client_id,omitempty"`
			Scopes         []string          `yaml:"scopes,omitempty"`
			UsePKCE        *bool             `yaml:"use_pkce,omitempty"`
			OAuthParams    map[string]string `yaml:"oauth_params,omitempty"`
			CallbackPort   int               `yaml:"callback_port,omitempty"`
		} `yaml:"oauth,omitempty"`
		// Headers for remote server authentication
		Headers []struct {
			Name        string   `yaml:"name"`
			Description string   `yaml:"description"`
			Required    bool     `yaml:"required"`
			Default     string   `yaml:"default,omitempty"`
			Secret      bool     `yaml:"secret,omitempty"`
			Choices     []string `yaml:"choices,omitempty"`
		} `yaml:"headers,omitempty"`
		// Environment variables for server configuration
		EnvVars []struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			Required    bool   `yaml:"required"`
			Default     string `yaml:"default,omitempty"`
			Secret      bool   `yaml:"secret,omitempty"`
		} `yaml:"env_vars,omitempty"`
	}
	var extended extendedFields
	if err := unmarshal(&extended); err != nil {
		return err
	}
	r.Examples = extended.Examples
	r.License = extended.License

	// Handle OAuth configuration transformation for remote servers
	if r.RemoteServerMetadata != nil && extended.OAuth != nil {
		r.RemoteServerMetadata.OAuthConfig = &registry.OAuthConfig{
			Issuer:       extended.OAuth.Issuer,
			AuthorizeURL: extended.OAuth.AuthorizeURL,
			TokenURL:     extended.OAuth.TokenURL,
			ClientID:     extended.OAuth.ClientID,
			Scopes:       extended.OAuth.Scopes,
			OAuthParams:  extended.OAuth.OAuthParams,
			CallbackPort: extended.OAuth.CallbackPort,
		}
		// Set UsePKCE default or explicit value
		if extended.OAuth.UsePKCE != nil {
			r.RemoteServerMetadata.OAuthConfig.UsePKCE = *extended.OAuth.UsePKCE
		} else {
			r.RemoteServerMetadata.OAuthConfig.UsePKCE = true // default to true
		}
	}

	// Handle Headers transformation for remote servers
	if r.RemoteServerMetadata != nil && len(extended.Headers) > 0 {
		r.RemoteServerMetadata.Headers = make([]*registry.Header, len(extended.Headers))
		for i, h := range extended.Headers {
			r.RemoteServerMetadata.Headers[i] = &registry.Header{
				Name:        h.Name,
				Description: h.Description,
				Required:    h.Required,
				Default:     h.Default,
				Secret:      h.Secret,
				Choices:     h.Choices,
			}
		}
	}

	// Handle EnvVars transformation for remote servers
	if r.RemoteServerMetadata != nil && len(extended.EnvVars) > 0 {
		r.RemoteServerMetadata.EnvVars = make([]*registry.EnvVar, len(extended.EnvVars))
		for i, e := range extended.EnvVars {
			r.RemoteServerMetadata.EnvVars[i] = &registry.EnvVar{
				Name:        e.Name,
				Description: e.Description,
				Required:    e.Required,
				Default:     e.Default,
				Secret:      e.Secret,
			}
		}
	}

	// Handle custom metadata fields for remote servers (homepage, license, author, etc.)
	if r.RemoteServerMetadata != nil {
		// Extract custom fields from the raw YAML that aren't part of the standard schema
		customFields := make(map[string]interface{})
		
		// Check for common custom fields
		if val, exists := raw["homepage"]; exists {
			customFields["homepage"] = val
		}
		if val, exists := raw["license"]; exists {
			customFields["license"] = val
		}
		if val, exists := raw["author"]; exists {
			customFields["author"] = val
		}
		
		// Set custom metadata if we found any custom fields
		if len(customFields) > 0 {
			r.RemoteServerMetadata.CustomMetadata = customFields
		}
	}

	return nil
}
