package serverjson

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ServerEntry represents a server.json file with its path and last-updated timestamp.
type ServerEntry struct {
	Path        string
	LastUpdated time.Time
}

// Scanner walks a registries directory to find and sort server.json files.
type Scanner struct {
	registriesDir string
}

// NewScanner creates a Scanner rooted at the given registries directory.
func NewScanner(registriesDir string) *Scanner {
	return &Scanner{registriesDir: registriesDir}
}

// FindOldestServers returns the N server.json entries with the oldest LastUpdated timestamps.
// Entries with missing or empty LastUpdated are treated as epoch (selected first).
// Malformed files are logged and skipped.
func (s *Scanner) FindOldestServers(count int) ([]ServerEntry, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %d", count)
	}

	registries, err := os.ReadDir(s.registriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read registries directory %s: %w", s.registriesDir, err)
	}

	var entries []ServerEntry

	for _, regEntry := range registries {
		if !regEntry.IsDir() || strings.HasPrefix(regEntry.Name(), ".") {
			continue
		}

		serversPath := filepath.Join(s.registriesDir, regEntry.Name(), "servers")
		info, err := os.Stat(serversPath)
		if err != nil || !info.IsDir() {
			continue
		}

		serverDirs, err := os.ReadDir(serversPath)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", serversPath, err)
			continue
		}

		for _, serverDir := range serverDirs {
			if !serverDir.IsDir() || strings.HasPrefix(serverDir.Name(), ".") {
				continue
			}

			serverJSONPath := filepath.Join(serversPath, serverDir.Name(), "server.json")
			entry, err := loadServerEntry(serverJSONPath)
			if err != nil {
				fmt.Printf("Warning: skipping %s: %v\n", serverJSONPath, err)
				continue
			}

			entries = append(entries, *entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUpdated.Before(entries[j].LastUpdated)
	})

	if count > len(entries) {
		count = len(entries)
	}

	return entries[:count], nil
}

// loadServerEntry loads a single server.json and extracts its LastUpdated timestamp.
func loadServerEntry(path string) (*ServerEntry, error) {
	sf, err := LoadServerFile(path)
	if err != nil {
		return nil, err
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		return nil, fmt.Errorf("failed to get extensions: %w", err)
	}

	lastUpdated := time.Time{} // epoch
	if ext.Metadata != nil && ext.Metadata.LastUpdated != "" {
		parsed, err := time.Parse(time.RFC3339, ext.Metadata.LastUpdated)
		if err != nil {
			fmt.Printf("Warning: invalid last_updated in %s: %v\n", path, err)
		} else {
			lastUpdated = parsed
		}
	}

	return &ServerEntry{
		Path:        path,
		LastUpdated: lastUpdated,
	}, nil
}
