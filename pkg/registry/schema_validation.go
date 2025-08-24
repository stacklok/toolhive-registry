// Package registry provides schema-based validation using the toolhive library
package registry

import (
	"encoding/json"
	"fmt"

	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"github.com/stacklok/toolhive-registry/pkg/types"
)

// SchemaValidator provides comprehensive schema-based validation using the toolhive library
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// ValidateEntry validates a single registry entry using the toolhive schema
func (v *SchemaValidator) ValidateEntry(entry *types.RegistryEntry, name string) error {
	// Convert our entry to the toolhive registry format for validation
	registry, err := v.convertToToolhiveRegistry(entry, name)
	if err != nil {
		return fmt.Errorf("failed to convert entry for validation: %w", err)
	}

	// Serialize to JSON for schema validation
	registryJSON, err := json.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry for validation: %w", err)
	}

	// Use toolhive's schema validation
	if err := toolhiveRegistry.ValidateRegistrySchema(registryJSON); err != nil {
		return fmt.Errorf("schema validation failed for entry '%s': %w", name, err)
	}

	return nil
}

// ValidateRegistry validates a complete registry using the toolhive schema
func (v *SchemaValidator) ValidateRegistry(registry *toolhiveRegistry.Registry) error {
	// Serialize to JSON for schema validation
	registryJSON, err := json.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry for validation: %w", err)
	}

	// Use toolhive's schema validation
	if err := toolhiveRegistry.ValidateRegistrySchema(registryJSON); err != nil {
		return fmt.Errorf("registry schema validation failed: %w", err)
	}

	return nil
}

// convertToToolhiveRegistry converts our RegistryEntry to a minimal toolhive Registry for validation
func (v *SchemaValidator) convertToToolhiveRegistry(entry *types.RegistryEntry, name string) (*toolhiveRegistry.Registry, error) {
	registry := &toolhiveRegistry.Registry{
		Version:     "1.0.0",
		LastUpdated: "2024-01-01T00:00:00Z", // Placeholder for validation
		Servers:     make(map[string]*toolhiveRegistry.ImageMetadata),
		RemoteServers: make(map[string]*toolhiveRegistry.RemoteServerMetadata),
	}

	if entry.IsImage() {
		// Set the name if not already set
		if entry.ImageMetadata.Name == "" {
			entry.ImageMetadata.Name = name
		}
		registry.Servers[name] = entry.ImageMetadata
	} else if entry.IsRemote() {
		// Set the name if not already set
		if entry.RemoteServerMetadata.Name == "" {
			entry.RemoteServerMetadata.Name = name
		}
		registry.RemoteServers[name] = entry.RemoteServerMetadata
	} else {
		return nil, fmt.Errorf("entry must be either image-based or remote server")
	}

	return registry, nil
}

// ValidateEntryFields performs additional field-level validation beyond schema validation
func (v *SchemaValidator) ValidateEntryFields(entry *types.RegistryEntry, name string) error {
	// Basic type validation
	if entry.ImageMetadata == nil && entry.RemoteServerMetadata == nil {
		return fmt.Errorf("entry '%s' must be either an image or remote server", name)
	}

	if entry.ImageMetadata != nil && entry.RemoteServerMetadata != nil {
		return fmt.Errorf("entry '%s' cannot be both image and remote server", name)
	}

	// Image-specific validation
	if entry.IsImage() {
		if entry.Image == "" {
			return fmt.Errorf("entry '%s': image field is required for image-based servers", name)
		}
	}

	// Remote-specific validation
	if entry.IsRemote() {
		if entry.URL == "" {
			return fmt.Errorf("entry '%s': url field is required for remote servers", name)
		}
		
		// Remote servers cannot use stdio transport
		if entry.GetTransport() == "stdio" {
			return fmt.Errorf("entry '%s': remote servers cannot use stdio transport (use sse or streamable-http)", name)
		}
	}

	// Common field validation
	if entry.GetDescription() == "" {
		return fmt.Errorf("entry '%s': description is required", name)
	}

	if entry.GetTransport() == "" {
		return fmt.Errorf("entry '%s': transport is required", name)
	}

	if len(entry.GetTools()) == 0 {
		return fmt.Errorf("entry '%s': at least one tool must be specified", name)
	}

	return nil
}

// ValidateComplete performs both schema validation and field validation
func (v *SchemaValidator) ValidateComplete(entry *types.RegistryEntry, name string) error {
	// First perform field validation
	if err := v.ValidateEntryFields(entry, name); err != nil {
		return err
	}

	// Then perform schema validation
	if err := v.ValidateEntry(entry, name); err != nil {
		return err
	}

	return nil
}