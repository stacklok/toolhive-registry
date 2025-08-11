// Package types provides extended types for the toolhive registry
package types

import (
	"time"

	"github.com/stacklok/toolhive/pkg/registry"
)

// RegistryEntry extends the toolhive ImageMetadata with additional fields
// for the modular registry system
type RegistryEntry struct {
	// Embed the original ImageMetadata from toolhive
	*registry.ImageMetadata `yaml:",inline"`

	// Examples of usage
	Examples []Example `yaml:"examples,omitempty"`

	// License information
	License string `yaml:"license,omitempty"`
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
