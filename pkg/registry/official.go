package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

type OfficialRegistry struct {
	loader *Loader
}

// NewOfficialRegistry creates a new instance of the official registry
func NewOfficialRegistry(loader *Loader) *OfficialRegistry {
	return &OfficialRegistry{
		loader: loader,
	}
}

func (or *OfficialRegistry) WriteJSON(path string) error {
	// Build the registry structure
	registry, err := or.build()
	if err != nil {
		return fmt.Errorf("failed to build official registry: %w", err)
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

// build creates the ToolHiveRegistryType structure from loaded entries
func (or *OfficialRegistry) build() (*ToolHiveRegistryType, error) {
	entries := or.loader.GetEntries()

	// TODO: Transform entries to upstream.ServerRecord (Phase 2)
	var servers []upstream.ServerRecord
	for range entries {
		// Placeholder - will be implemented in Phase 2
		servers = append(servers, upstream.ServerRecord{})
	}

	registry := &ToolHiveRegistryType{
		Schema:  "", // TODO: Add schema URL once applicable
		Version: "1.0.0",
		Meta: Meta{
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
		Data: Data{
			Servers: servers,
			Groups:  []Group{}, // Empty for now, placeholder for future use
		},
	}

	return registry, nil
}
