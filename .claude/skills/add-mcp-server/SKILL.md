---
name: add-mcp-server
description: >-
  Add new MCP server entries to the ToolHive registry. Creates server.json and
  icon.svg files with correct schema, _meta extensions, and validation.
  Use when adding a server, creating a registry entry, onboarding an MCP server,
  or writing server.json. NOT for reviewing existing entries (use mcp-review).
allowed-tools: Read Grep Glob Bash Write Edit WebFetch WebSearch
---

# Add MCP Server to ToolHive Registry

Each entry is a `server.json` file in `registries/toolhive/servers/<name>/` following the
[MCP ServerJSON schema](https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json)
with ToolHive extensions in `_meta`.

## Workflow

1. **Determine type**: Docker/OCI image → Container (`packages`). HTTP endpoint → Remote (`remotes`).
2. **Choose name**: lowercase, numbers, hyphens only. Append `-remote` for remote variants.
3. **Create directory**: `mkdir -p registries/toolhive/servers/<name>`
4. **Gather info** if needed: fetch README/docs for tools, transport, env vars, auth.
5. **Create `icon.svg`**: service logo, simple, standard SVG format.
6. **Create `server.json`**: use templates below.
7. **Validate**: `task catalog:validate && task catalog:build`

## Minimal Container Template

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/<server-name>",
  "description": "Clear, concise one-line description",
  "title": "<Human-Readable Title>",
  "repository": {
    "url": "https://github.com/org/repo",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/org/server:v1.0.0",
      "transport": {
        "type": "stdio"
      }
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/<server-name>/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/org/server:v1.0.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["category1", "category2"],
          "tools": ["tool_name"],
          "overview": "## Server Title\n\nA 3-5 sentence markdown description of purpose and capabilities."
        }
      }
    }
  }
}
```

For complete templates with env vars, permissions, provenance: see [references/container-servers.md](references/container-servers.md).

## Minimal Remote Template

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/<server-name>",
  "description": "Clear, concise one-line description",
  "title": "<Human-Readable Title> (Remote)",
  "repository": {
    "url": "https://github.com/org/repo",
    "source": "github"
  },
  "version": "1.0.0",
  "remotes": [
    {
      "type": "streamable-http",
      "url": "https://api.example.com/mcp"
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/<server-name>/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://api.example.com/mcp": {
          "tier": "Community",
          "status": "Active",
          "tags": ["remote", "category1"],
          "tools": ["tool_name"],
          "overview": "## Server Title (Remote)\n\nA 3-5 sentence markdown description of purpose and capabilities."
        }
      }
    }
  }
}
```

For complete templates with OAuth, custom_metadata: see [references/remote-servers.md](references/remote-servers.md).

## Critical Rules

| Rule | Detail |
|------|--------|
| **Extension key** | `_meta` key MUST exactly match `packages[0].identifier` (containers) or `remotes[0].url` (remotes) |
| **Icons** | Every entry needs `icons` array + `icon.svg` file in server directory |
| **Overview** | Markdown string starting with `## Title\n\n` followed by 3-5 sentences |
| **Name format** | `io.github.stacklok/<server-name>` |
| **Title** | Human-readable display name (e.g., `"Fetch"`, `"GitHub (Remote)"`, `"AWS Knowledge Bases"`) |
| **Tier** | Exactly `"Official"` or `"Community"`. Default to `"Community"` — `"Official"` is reserved for servers maintained by the ToolHive team, the MCP spec authors, or the platform owner of the integrated service |
| **Status** | Exactly `"Active"` or `"Deprecated"` |
| **Remote tags** | Must include `"remote"` in tags |
| **Remote transport** | `"streamable-http"` or `"sse"` only (NEVER `"stdio"`) |
| **Container transport** | `"stdio"` (default) or `"streamable-http"` with `"url": "http://localhost:8080"` |
| **No filesystem paths** | NEVER include filesystem paths in permissions — network only |
| **Image tags** | Always pin version tags (never `latest`) |

## Auto-Populated Fields (Do NOT Include)

CI workflows automatically populate these — omit from new entries:

- `metadata.stars`, `metadata.pulls`, `metadata.last_updated`
- `tool_definitions` (full MCP Tool objects with inputSchema)

The `tools` list is also auto-updated by CI on PR, but include your best-effort list in new entries.

## Information Gathering

When the user provides only a URL or incomplete details:

1. **Repository README**: tools, env vars, auth, quick start
2. **Official docs**: transport, OAuth flows, API specs
3. **Container registry**: image references, versions

Extract: tool names, description, transport type, env vars, auth method, network hosts.

## Validation

```bash
# Validate schema compliance
task catalog:validate

# Build full registry
task catalog:build

# Verify entry in output
jq '.data.servers[] | select(.name == "io.github.stacklok/<name>")' build/toolhive/registry-upstream.json
```

## Reference Examples

Study existing entries for patterns:

- **Container, stdio**: `registries/toolhive/servers/github/server.json`
- **Container, streamable-http**: `registries/toolhive/servers/sqlite/server.json`
- **Remote, Official**: `registries/toolhive/servers/semgrep-remote/server.json`
- **Remote, OAuth**: `registries/toolhive/servers/github-remote/server.json`

## See Also

- [references/container-servers.md](references/container-servers.md) — Full container template, transport variants, env vars, provenance, permissions
- [references/remote-servers.md](references/remote-servers.md) — Full remote template, OAuth configuration
- [references/field-reference.md](references/field-reference.md) — All field details, overview format, network permissions, troubleshooting
- [references/examples.md](references/examples.md) — Real-world patterns from the registry (heroku, sqlite, semgrep-remote, github-remote)
