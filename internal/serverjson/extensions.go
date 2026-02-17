// Package serverjson provides utilities for reading, modifying, and writing
// individual server.json files, with support for extracting and updating
// ToolHive-specific extensions in the _meta section.
package serverjson

import (
	"encoding/json"
	"fmt"
	"os"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/stacklok/toolhive/pkg/registry/registry"
)

// ServerFile represents a loaded server.json with its file path and parsed data.
type ServerFile struct {
	// Path is the filesystem path to the server.json file.
	Path string
	// ServerJSON is the parsed upstream ServerJSON structure.
	ServerJSON upstream.ServerJSON
	// rawBytes preserves the original file bytes for round-trip fidelity.
	rawBytes []byte
}

// LoadServerFile reads and parses a server.json from the given path.
func LoadServerFile(path string) (*ServerFile, error) {
	data, err := os.ReadFile(path) // #nosec G304 - path comes from caller (registry directory walk)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var server upstream.ServerJSON
	if err := json.Unmarshal(data, &server); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &ServerFile{
		Path:       path,
		ServerJSON: server,
		rawBytes:   data,
	}, nil
}

// ExtensionKey returns the key used inside the publisher namespace.
// For OCI packages this is the image identifier (e.g. "ghcr.io/org/image:tag").
// For remote servers this is the URL (e.g. "https://api.example.com/mcp").
func (sf *ServerFile) ExtensionKey() (string, error) {
	switch {
	case len(sf.ServerJSON.Packages) > 0:
		return sf.ServerJSON.Packages[0].Identifier, nil
	case len(sf.ServerJSON.Remotes) > 0:
		return sf.ServerJSON.Remotes[0].URL, nil
	default:
		return "", fmt.Errorf("server.json has neither packages nor remotes")
	}
}

// IsPackageServer returns true if this server.json defines OCI package(s).
func (sf *ServerFile) IsPackageServer() bool {
	return len(sf.ServerJSON.Packages) > 0
}

// IsRemoteServer returns true if this server.json defines remote server(s).
func (sf *ServerFile) IsRemoteServer() bool {
	return len(sf.ServerJSON.Remotes) > 0
}

// RepositoryURL returns the repository URL if set, or empty string.
func (sf *ServerFile) RepositoryURL() string {
	if sf.ServerJSON.Repository != nil {
		return sf.ServerJSON.Repository.URL
	}
	return ""
}

// GetExtensions extracts the ServerExtensions for this server.
// It navigates: _meta → publisher-provided → io.github.stacklok → <key>
// and deserializes into registry.ServerExtensions.
func (sf *ServerFile) GetExtensions() (*registry.ServerExtensions, error) {
	if sf.ServerJSON.Meta == nil || sf.ServerJSON.Meta.PublisherProvided == nil {
		return &registry.ServerExtensions{}, nil
	}

	stacklokData, ok := sf.ServerJSON.Meta.PublisherProvided[registry.ToolHivePublisherNamespace].(map[string]any)
	if !ok {
		return &registry.ServerExtensions{}, nil
	}

	extKey, err := sf.ExtensionKey()
	if err != nil {
		return nil, err
	}

	extData, ok := stacklokData[extKey].(map[string]any)
	if !ok {
		return &registry.ServerExtensions{}, nil
	}

	jsonData, err := json.Marshal(extData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extensions: %w", err)
	}

	var ext registry.ServerExtensions
	if err := json.Unmarshal(jsonData, &ext); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extensions: %w", err)
	}

	return &ext, nil
}

// UpdateExtensions writes the modified ServerExtensions back into the server.json.
// It parses the raw file bytes into a map[string]interface{} to preserve all
// fields (including unknown ones), replaces only the extension subtree, and
// writes back with json.MarshalIndent.
func (sf *ServerFile) UpdateExtensions(ext *registry.ServerExtensions) error {
	extJSON, err := json.Marshal(ext)
	if err != nil {
		return fmt.Errorf("failed to marshal extensions: %w", err)
	}

	var extMap map[string]any
	if err := json.Unmarshal(extJSON, &extMap); err != nil {
		return fmt.Errorf("failed to unmarshal extensions map: %w", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(sf.rawBytes, &doc); err != nil {
		return fmt.Errorf("failed to parse raw JSON: %w", err)
	}

	extKey, err := sf.ExtensionKey()
	if err != nil {
		return err
	}

	meta := ensureMap(doc, "_meta")
	publisherProvided := ensureMap(meta, registry.PublisherProvidedKey)
	stacklok := ensureMap(publisherProvided, registry.ToolHivePublisherNamespace)
	stacklok[extKey] = extMap

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal server.json: %w", err)
	}
	out = append(out, '\n')

	if err := os.WriteFile(sf.Path, out, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", sf.Path, err)
	}

	sf.rawBytes = out
	return nil
}

// MetaSize returns the size in bytes of the _meta field when serialized as JSON.
// Returns 0 if there is no _meta field.
func (sf *ServerFile) MetaSize() int {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(sf.rawBytes, &doc); err != nil {
		return 0
	}
	meta, ok := doc["_meta"]
	if !ok {
		return 0
	}
	return len(meta)
}

// ensureMap gets or creates a nested map[string]any at the given key.
func ensureMap(parent map[string]any, key string) map[string]any {
	if val, ok := parent[key].(map[string]any); ok {
		return val
	}
	m := make(map[string]any)
	parent[key] = m
	return m
}
