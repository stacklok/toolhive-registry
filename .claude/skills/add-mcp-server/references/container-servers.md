# Container Server Details

## Complete Container Template

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/<server-name>",
  "description": "Enables interaction with [service/API] for [purpose]",
  "title": "<Human-Readable Title>",
  "repository": {
    "url": "https://github.com/organization/repository",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/organization/server-name:v1.0.0",
      "transport": {
        "type": "stdio"
      },
      "environmentVariables": [
        {
          "name": "API_KEY",
          "description": "Authentication key for service",
          "isRequired": true,
          "isSecret": true
        },
        {
          "name": "BASE_URL",
          "description": "API endpoint URL (optional)",
          "default": "https://api.example.com"
        }
      ],
      "arguments": ["--verbose"]
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
        "ghcr.io/organization/server-name:v1.0.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["category1", "category2"],
          "tools": [
            "tool_name_1",
            "tool_name_2",
            "tool_name_3"
          ],
          "overview": "## Server Name\n\nA markdown-formatted description of the server's purpose, key capabilities, and typical use cases. Keep to 3-5 sentences.",
          "permissions": {
            "network": {
              "outbound": {
                "allow_host": ["api.example.com", "auth.example.com"],
                "allow_port": [443, 80]
              }
            }
          },
          "provenance": {
            "cert_issuer": "https://token.actions.githubusercontent.com",
            "repository_uri": "https://github.com/organization/repository",
            "runner_environment": "github-hosted",
            "signer_identity": "/.github/workflows/build-containers.yml",
            "sigstore_url": "tuf-repo-cdn.sigstore.dev"
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

## Transport Variants

**stdio** (~75% of entries — most common):
```json
"transport": {
  "type": "stdio"
}
```

**streamable-http** (when the container runs an HTTP listener):
```json
"transport": {
  "type": "streamable-http",
  "url": "http://localhost:8080"
}
```

Use `streamable-http` when the containerized server exposes an HTTP endpoint rather than communicating over stdin/stdout.

## Environment Variables

| Field         | Type    | Description                                    |
| ------------- | ------- | ---------------------------------------------- |
| `name`        | string  | Variable name (e.g., `"API_KEY"`)              |
| `description` | string  | What the variable is for                       |
| `isRequired`  | boolean | Whether the variable must be set               |
| `isSecret`    | boolean | Whether the value is sensitive (keys, tokens)  |
| `default`     | string  | Default value if not provided by user          |

## Provenance (Supply Chain Security)

Include when the container image is signed via Sigstore/cosign:

```json
"provenance": {
  "cert_issuer": "https://token.actions.githubusercontent.com",
  "repository_uri": "https://github.com/organization/repository",
  "runner_environment": "github-hosted",
  "signer_identity": "/.github/workflows/build-containers.yml",
  "sigstore_url": "tuf-repo-cdn.sigstore.dev"
}
```

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

Prefix hostname with `.` for wildcard subdomain matching (`.github.com` allows `api.github.com`, etc.).

**Broad network access** (for fetch/HTTP servers):
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
