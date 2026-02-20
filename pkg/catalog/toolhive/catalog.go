// Package toolhive provides the pre-built ToolHive registry as embedded bytes.
//
// The data files are updated daily by the build workflow and committed to the
// repository, so the embedded content reflects the last daily build.
//
// Two formats are available via [Legacy] and [Upstream]:
//   - Legacy: ToolHive's own registry format (registry.json)
//   - Upstream: The upstream MCP official registry format (official-registry.json)
package toolhive

import _ "embed"

//go:embed data/registry.json
var legacyRegistry []byte

//go:embed data/official-registry.json
var upstreamRegistry []byte

// Legacy returns the registry in ToolHive's legacy format.
func Legacy() []byte {
	return legacyRegistry
}

// Upstream returns the registry in the upstream MCP official format.
func Upstream() []byte {
	return upstreamRegistry
}
