# Field Reference

## Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `$schema` | Yes | Always `"https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json"` |
| `name` | Yes | `"io.github.stacklok/<server-name>"` |
| `description` | Yes | One-line, clear, concise |
| `title` | Yes | Human-readable display name (e.g., `"Fetch"`, `"GitHub (Remote)"`) |
| `version` | Yes | `"1.0.0"` for new entries |
| `repository` | Yes | `{ "url": "https://github.com/...", "source": "github" }` |
| `packages` | Container | Array with one package entry |
| `remotes` | Remote | Array with one remote entry |
| `icons` | Yes | Array with `icon.svg` reference (see below) |

### Icons Format

Always the same structure — only the server name changes:

```json
"icons": [
  {
    "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/<server-name>/icon.svg",
    "mimeType": "image/svg+xml",
    "sizes": ["any"]
  }
]
```

## Extension (`_meta`) Fields

Nested at: `_meta["io.modelcontextprotocol.registry/publisher-provided"]["io.github.stacklok"]["<extension-key>"]`

| Field | Required | Description |
|-------|----------|-------------|
| `tier` | Yes | `"Official"` or `"Community"` |
| `status` | Yes | `"Active"` or `"Deprecated"` |
| `tools` | Yes | List of tool names |
| `overview` | Yes | Markdown description (see below) |
| `tags` | Recommended | Categorization array; include `"remote"` for remote servers |
| `permissions` | When needed | Network access config |
| `provenance` | When available | Sigstore/cosign supply chain data |
| `custom_metadata` | Optional | `{ "author": "...", "homepage": "...", "license": "..." }` |
| `oauth_config` | Remote+OAuth | OAuth configuration for remote servers |

### Auto-Populated by CI (Do NOT Include)

- `metadata.stars` — GitHub stars
- `metadata.pulls` — Docker pulls
- `metadata.last_updated` — Timestamp
- `tool_definitions` — Full MCP Tool objects with inputSchema

## Overview Field Format

Markdown string with heading + 3-5 sentence description:

```json
"overview": "## Heroku MCP Server\n\nThe heroku-mcp-server is a Model Context Protocol (MCP) server that enables AI assistants and agents to interact directly with the Heroku platform. Key capabilities include application lifecycle management, deployment workflows, runtime insight for dynos and process scaling, and operational visibility through logs and platform metadata."
```

Pattern: `## <Display Name>\n\n<3-5 sentences about purpose, capabilities, and use cases.>`

## Network Permissions

**Specific hosts (preferred):**
```json
"permissions": {
  "network": {
    "outbound": {
      "allow_host": ["api.example.com", ".example.com"],
      "allow_port": [443]
    }
  }
}
```

- Prefix with `.` for wildcard subdomains (`.github.com` → `api.github.com`, `raw.github.com`, etc.)

**Broad network access** (fetch/HTTP servers):
```json
"permissions": {
  "network": {
    "outbound": {
      "allow_port": [443],
      "insecure_allow_all": true
    }
  }
}
```

NEVER include filesystem paths in permissions.

## Troubleshooting

**"Extension key mismatch"** — `_meta` key must exactly match `packages[0].identifier` or `remotes[0].url`, including version tag.

**"Invalid transport type"** — Container: `stdio`, `streamable-http`, `sse`. Remote: `streamable-http`, `sse` (not `stdio`).

**"Missing required fields"** — Check `$schema`, `name`, `description`, `title`, `version`, `icons`, and `packages`/`remotes`.

**"Invalid tier or status"** — Must be exactly `"Official"`/`"Community"` and `"Active"`/`"Deprecated"`.

**"Invalid JSON"** — No trailing commas, quoted keys, 2-space indentation. Validate with `jq . < server.json`.

## Checklist

- [ ] Server name: lowercase, numbers, hyphens only
- [ ] Directory: `registries/toolhive/servers/<name>/` exists
- [ ] `server.json` and `icon.svg` present
- [ ] `$schema`, `name`, `description`, `title`, `version`, `repository`, `icons` set
- [ ] Container: `packages` with `registryType`, `identifier`, `transport`
- [ ] Container: `transport.url` if transport is `streamable-http`
- [ ] Remote: `remotes` with `type` and `url`
- [ ] `_meta` extension key matches `identifier` or `url` exactly
- [ ] `tier`, `status`, `tools`, `overview` set
- [ ] Remote tags include `"remote"`
- [ ] `metadata`, `tool_definitions` NOT included (auto-populated)
- [ ] `task catalog:validate` passes
- [ ] `task catalog:build` succeeds
