// Package official provides the pre-built official vendors registry as embedded bytes.
//
// This registry contains MCP servers and skills from official vendors,
// companies, and widely popular community projects.
//
// The data files are updated daily by the build workflow and committed to the
// repository, so the embedded content reflects the last daily build.
//
// Two formats are available via [Legacy] and [Upstream]:
//   - Legacy: ToolHive's own registry format (registry-legacy.json)
//   - Upstream: The upstream MCP official registry format (registry-upstream.json)
package official

import _ "embed"

//go:embed data/registry-legacy.json
var legacyRegistry []byte

//go:embed data/registry-upstream.json
var upstreamRegistry []byte

// Legacy returns the registry in ToolHive's legacy format.
func Legacy() []byte {
	return legacyRegistry
}

// Upstream returns the registry in the upstream MCP official format.
func Upstream() []byte {
	return upstreamRegistry
}
