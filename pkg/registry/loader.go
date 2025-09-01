// Package registry provides functionality for loading and managing registry entries
package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

// Loader handles loading registry entries from YAML files
type Loader struct {
	registryPath string
	entries      map[string]*types.RegistryEntry
}

// NewLoader creates a new registry loader
func NewLoader(registryPath string) *Loader {
	return &Loader{
		registryPath: registryPath,
		entries:      make(map[string]*types.RegistryEntry),
	}
}

// LoadAll loads all registry entries from the registry directory
func (l *Loader) LoadAll() error {
	// Walk through the registry directory
	err := filepath.Walk(l.registryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory or if it's the root directory
		if !info.IsDir() || path == l.registryPath {
			return nil
		}

		// Get the relative path from registry root
		relPath, err := filepath.Rel(l.registryPath, path)
		if err != nil {
			return err
		}

		// Skip hidden directories and nested directories
		if strings.HasPrefix(info.Name(), ".") || strings.Contains(relPath, string(os.PathSeparator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Try to load spec.yaml from this directory
		specPath := filepath.Join(path, "spec.yaml")
		if _, err := os.Stat(specPath); err == nil {
			// Use directory name as the entry name
			entryName := info.Name()

			entry, err := l.LoadEntryWithName(specPath, entryName)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", specPath, err)
			}

			// Override with explicit name if set in the spec
			if entry.GetName() != "" && entry.GetName() != entryName {
				entryName = entry.GetName()
			} else {
				entry.SetName(entryName)
			}

			l.entries[entryName] = entry
		}

		return nil
	})

	return err
}

// LoadEntry loads a single registry entry from a YAML file without validation
// Use LoadEntryWithName for validation with proper naming
func (l *Loader) LoadEntry(path string) (*types.RegistryEntry, error) {
	return l.LoadEntryWithName(path, "")
}

// LoadEntryWithName loads a single registry entry from a YAML file with validation
func (l *Loader) LoadEntryWithName(path string, name string) (*types.RegistryEntry, error) {
	file, err := os.Open(path) // #nosec G304 - path is constructed from known directory structure
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var entry types.RegistryEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate with the actual name if provided
	if err := l.validateEntry(&entry, name); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &entry, nil
}

// validateEntry validates a registry entry using comprehensive schema-based validation
func (*Loader) validateEntry(entry *types.RegistryEntry, name string) error {
	// Use the new schema validator for comprehensive validation
	validator := NewSchemaValidator()

	return validator.ValidateComplete(entry, name)
}

// GetEntries returns all loaded entries
func (l *Loader) GetEntries() map[string]*types.RegistryEntry {
	return l.entries
}

// GetSortedEntries returns entries sorted by name
func (l *Loader) GetSortedEntries() []*types.RegistryEntry {
	var entries []*types.RegistryEntry
	for _, entry := range l.entries {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GetName() < entries[j].GetName()
	})

	return entries
}
