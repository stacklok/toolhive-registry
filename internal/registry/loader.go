// Package registry provides functionality for loading server.json files
// and building the aggregate upstream registry.
package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

// Loader walks a servers directory and loads server.json files into memory.
type Loader struct {
	serversPath string
	entries     map[string]upstream.ServerJSON
}

// NewLoader creates a new Loader that reads from the given servers directory.
func NewLoader(serversPath string) *Loader {
	return &Loader{
		serversPath: serversPath,
		entries:     make(map[string]upstream.ServerJSON),
	}
}

// LoadAll walks top-level subdirectories under serversPath, reads server.json
// from each, and stores the results keyed by directory name.
// Hidden directories (starting with ".") are skipped.
// Returns an error if any server.json contains malformed JSON.
func (l *Loader) LoadAll() error {
	err := filepath.Walk(l.serversPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories and the root directory itself
		if !info.IsDir() || path == l.serversPath {
			return nil
		}

		relPath, err := filepath.Rel(l.serversPath, path)
		if err != nil {
			return err
		}

		// Skip hidden directories
		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip nested directories (only process top-level)
		if strings.Contains(relPath, string(os.PathSeparator)) {
			return filepath.SkipDir
		}

		// Read server.json from this directory
		serverJSONPath := filepath.Join(path, "server.json")
		data, err := os.ReadFile(serverJSONPath) // #nosec G304 - path is constructed from known directory structure
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("WARNING: no server.json in %s, skipping", info.Name())
				return nil
			}
			return fmt.Errorf("failed to read %s: %w", serverJSONPath, err)
		}

		var server upstream.ServerJSON
		if err := json.Unmarshal(data, &server); err != nil {
			return fmt.Errorf("failed to parse %s: %w", serverJSONPath, err)
		}

		l.entries[info.Name()] = server
		return nil
	})

	return err
}

// GetEntries returns all loaded server.json entries keyed by directory name.
func (l *Loader) GetEntries() map[string]upstream.ServerJSON {
	return l.entries
}

// GetSortedNames returns directory names in sorted order for deterministic output.
func (l *Loader) GetSortedNames() []string {
	names := make([]string, 0, len(l.entries))
	for name := range l.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
