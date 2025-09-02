package registry

import upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"

// ToolHiveRegistryType represents the structure of a ToolHive registry file using the official format
type ToolHiveRegistryType struct {
	// Schema is the JSON schema URL
	Schema string `json:"$schema"`
	// Version is the schema version of the registry
	Version string `json:"version" yaml:"version"`
	// Meta contains metadata about the registry
	Meta Meta `json:"meta" yaml:"meta"`
	// Data contains the actual registry data
	Data Data `json:"data" yaml:"data"`
}

// Group represents a group of servers (not implemented yet, placeholder for future use)
type Group struct {
}

// Data holds the servers and groups in the registry
type Data struct {
	Servers []upstream.ServerJSON `json:"servers" yaml:"servers"`
	Groups  []Group               `json:"groups" yaml:"groups"`
}

// Meta contains metadata about the registry
type Meta struct {
	// LastUpdated is the timestamp when the registry was last updated, in RFC3339 format
	LastUpdated string `json:"last_updated" yaml:"last_updated"`
}
