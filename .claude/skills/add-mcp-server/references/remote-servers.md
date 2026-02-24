# Remote Server Details

## Complete Remote Template

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/<server-name>",
  "description": "Enables interaction with [service/API] for [purpose]",
  "title": "<Human-Readable Title> (Remote)",
  "repository": {
    "url": "https://github.com/organization/repository",
    "source": "github"
  },
  "version": "1.0.0",
  "remotes": [
    {
      "type": "streamable-http",
      "url": "https://api.example.com/mcp/v1"
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
        "https://api.example.com/mcp/v1": {
          "tier": "Official",
          "status": "Active",
          "tags": ["remote", "api", "integration"],
          "tools": [
            "tool_name_1",
            "tool_name_2"
          ],
          "overview": "## Server Name (Remote)\n\nA markdown-formatted description of the server's purpose, key capabilities, and typical use cases. Keep to 3-5 sentences.",
          "oauth_config": {
            "authorize_url": "https://auth.example.com/oauth/authorize",
            "token_url": "https://auth.example.com/oauth/token",
            "scopes": ["read", "write"],
            "use_pkce": true
          },
          "custom_metadata": {
            "author": "Organization Name",
            "homepage": "https://docs.example.com",
            "license": "MIT"
          }
        }
      }
    }
  }
}
```

## Remote-Specific Rules

- Transport: `"streamable-http"` (preferred) or `"sse"`. NEVER `"stdio"`.
- Tags: always include `"remote"`.
- Extension key: must exactly match `remotes[0].url`.
- URL: must be HTTPS.

## OAuth Configuration

For remote servers requiring OAuth authentication:

```json
"oauth_config": {
  "authorize_url": "https://auth.example.com/oauth/authorize",
  "token_url": "https://auth.example.com/oauth/token",
  "scopes": ["read", "write"],
  "use_pkce": true
}
```

| Field           | Type     | Description                                |
| --------------- | -------- | ------------------------------------------ |
| `authorize_url` | string   | OAuth authorization endpoint               |
| `token_url`     | string   | OAuth token exchange endpoint              |
| `scopes`        | string[] | Required OAuth scopes                      |
| `use_pkce`      | boolean  | Whether to use PKCE flow (recommended)     |

Only include `oauth_config` when the remote server actually requires OAuth. Many remote servers work without authentication or use simpler auth.

## Custom Metadata

Optional but recommended for Official-tier servers:

```json
"custom_metadata": {
  "author": "Organization Name",
  "homepage": "https://docs.example.com",
  "license": "MIT"
}
```
