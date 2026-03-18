# server.json Specification Reference

Canonical schema: https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json
Full authoring guide: `docs/adding-entries-llm.md` in this repo.

## Table of Contents

- [Required Top-Level Fields](#required-top-level-fields)
- [Container Servers](#container-servers)
- [Remote Servers](#remote-servers)
- [Extensions Block](#extensions-block)
- [Auto-Populated Fields](#auto-populated-fields)
- [Validation Checklist](#validation-checklist)

---

## Required Top-Level Fields

Every server.json must have:

| Field | Format | Example |
|-------|--------|---------|
| `$schema` | Fixed URL | `"https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json"` |
| `name` | `io.github.stacklok/<server-name>` | `"io.github.stacklok/github"` |
| `description` | Single clear line | `"Enables interaction with GitHub..."` |
| `title` | Human-readable | `"GitHub"` |
| `version` | Semver | `"1.0.0"` |
| `repository.url` | GitHub URL | `"https://github.com/org/repo"` |
| `repository.source` | Platform | `"github"` |
| `icons` | Array with icon.svg ref | See below |

### Icons (always required)

```json
"icons": [
  {
    "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/<server-name>/icon.svg",
    "mimeType": "image/svg+xml",
    "sizes": ["any"]
  }
]
```

---

## Container Servers

Use `packages` field. Transports: `stdio`, `streamable-http`, `sse`.

```json
"packages": [{
  "registryType": "oci",
  "identifier": "ghcr.io/org/server:v1.0.0",
  "transport": { "type": "stdio" },
  "environmentVariables": [
    { "name": "API_KEY", "description": "Auth key", "isRequired": true, "isSecret": true }
  ]
}]
```

**When transport is `streamable-http`**, a `url` field is required:

```json
"transport": {
  "type": "streamable-http",
  "url": "http://localhost:8080"
}
```

### Environment Variables

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | Variable name |
| `description` | string | What it's for |
| `isRequired` | boolean | Must be set? |
| `isSecret` | boolean | Sensitive value? **Must be true for keys/tokens** |
| `default` | string | Default if not provided |

---

## Remote Servers

Use `remotes` field. Transports: `streamable-http` (preferred), `sse`. **Never `stdio`**.

```json
"remotes": [{
  "type": "streamable-http",
  "url": "https://api.example.com/mcp"
}]
```

### OAuth Configuration (when applicable)

```json
"oauth_config": {
  "authorize_url": "https://auth.example.com/oauth/authorize",
  "token_url": "https://auth.example.com/oauth/token",
  "scopes": ["read", "write"],
  "use_pkce": true
}
```

---

## Extensions Block

Nested at: `_meta["io.modelcontextprotocol.registry/publisher-provided"]["io.github.stacklok"]["<extension-key>"]`

**Critical**: `<extension-key>` must exactly match:
- Containers: `packages[0].identifier`
- Remotes: `remotes[0].url`

### Required extension fields

| Field | Values | Notes |
|-------|--------|-------|
| `tier` | `"Official"` or `"Community"` | |
| `status` | `"Active"` or `"Deprecated"` | |
| `tools` | `["tool_a", "tool_b"]` | At least one; or `["set_during_runtime"]` |
| `overview` | Markdown string | Must start with `## Title\n\n` followed by 3-5 sentences |

### Recommended extension fields

| Field | Notes |
|-------|-------|
| `tags` | Categories; include `"remote"` for remote servers |
| `permissions.network` | Network access declaration (never filesystem paths) |
| `provenance` | Sigstore/cosign supply chain info |
| `custom_metadata` | Author, homepage, license |

### Network Permissions

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

Use `insecure_allow_all: true` only for general-purpose fetch/HTTP servers.

**NEVER** include filesystem paths in permissions.

### Provenance

```json
"provenance": {
  "cert_issuer": "https://token.actions.githubusercontent.com",
  "repository_uri": "https://github.com/org/repo",
  "runner_environment": "github-hosted",
  "signer_identity": "/.github/workflows/release.yml",
  "sigstore_url": "tuf-repo-cdn.sigstore.dev"
}
```

---

## Auto-Populated Fields

These are set by CI — **reject if present in new submissions**:

- `metadata.stars` — GitHub stars
- `metadata.pulls` — Docker pull count
- `metadata.last_updated` — Timestamp
- `tool_definitions` — Full MCP Tool objects with `inputSchema`

The `tools` list is also auto-updated by CI on PR, but should be present in new entries as a best-effort list.

---

## Validation Checklist

### Container Servers

- [ ] `packages[0].identifier` has pinned version tag (not `latest`)
- [ ] `transport.type` is `stdio`, `streamable-http`, or `sse`
- [ ] `transport.url` present when type is `streamable-http`
- [ ] Extension key matches `packages[0].identifier` exactly
- [ ] Secrets marked with `isSecret: true`

### Remote Servers

- [ ] `remotes[0].url` is HTTPS
- [ ] `remotes[0].type` is `streamable-http` or `sse` (not `stdio`)
- [ ] Extension key matches `remotes[0].url` exactly
- [ ] Tags include `"remote"`
- [ ] `oauth_config` present if server requires OAuth

### Both Types

- [ ] All required top-level fields present
- [ ] `icons` array present with correct URL
- [ ] `overview` starts with `## Title\n\n`
- [ ] `tier` is exactly `"Official"` or `"Community"`
- [ ] `status` is exactly `"Active"` or `"Deprecated"`
- [ ] `tools` array has at least one entry
- [ ] No `metadata.*` or `tool_definitions` fields (auto-populated)
- [ ] No filesystem paths in `permissions`
- [ ] Network permissions scoped (no unjustified `insecure_allow_all`)
- [ ] Valid JSON with 2-space indentation
