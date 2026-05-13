// Package official provides the pre-built official vendors registry as embedded bytes.
//
// This registry contains MCP servers and skills from official vendors,
// companies, and widely popular community projects.
//
// The data file is updated daily by the build workflow and committed to the
// repository, so the embedded content reflects the last daily build. The
// registry is exposed via [Upstream] in the upstream MCP official registry
// format (registry-upstream.json).
package official

import _ "embed"

//go:embed data/registry-upstream.json
var upstreamRegistry []byte

// Upstream returns the registry in the upstream MCP official format.
func Upstream() []byte {
	return upstreamRegistry
}
