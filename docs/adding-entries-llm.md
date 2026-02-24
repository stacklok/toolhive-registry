# Instructions for LLMs: Adding MCP Server Entries to ToolHive Registry

## Context

You are helping to add an MCP (Model Context Protocol) server entry to the ToolHive registry. Each entry is a `server.json` file following the [upstream MCP ServerJSON schema](https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json) with ToolHive-specific extensions in the `_meta` field.

## Quick Reference

### Server Types

| Type      | Identifier         | Transport Options                      | Primary Use                   |
| --------- | ------------------ | -------------------------------------- | ----------------------------- |
| Container | `packages` field   | `stdio`, `streamable-http`, `sse`      | Self-hosted Docker containers |
| Remote    | `remotes` field    | `streamable-http`, `sse` (NOT `stdio`) | HTTP/HTTPS API endpoints      |

### Field Priority Levels

- **REQUIRED**: Must be present or validation will fail
- **HIGHLY RECOMMENDED**: Should be included when available
- **OPTIONAL**: Include when relevant or needed
- **AUTO-POPULATED**: Set by CI workflows; do NOT include for new entries

## Step-by-Step Process

### 1. Determine Server Type

**Ask yourself:** Is this a Docker container or a remote HTTP endpoint?

- If Docker/OCI image → Container-based server (uses `packages`)
- If HTTP/HTTPS API → Remote server (uses `remotes`)

### 2. Choose Server Name

- Use **only** lowercase letters, numbers, and hyphens
- Make it descriptive and unique
- Examples: `github`, `aws-pricing`, `sqlite`, `notion-remote`

### 3. Create Directory and Icon

```bash
mkdir -p registries/toolhive/servers/<server-name>
```

Each server directory must contain:
- `server.json` — the server definition
- `icon.svg` — an SVG icon representing the server or service

### 4. Gather Information (If Needed)

**When to fetch documentation:**

- User provides only a repository URL or homepage without details
- Tool names, descriptions, or authentication requirements are unclear
- Transport type needs verification for remote servers
- Environment variables or permissions are unknown

**Where to look:**

1. **Repository README** - Primary source for:

   - Tool/function lists
   - Quick start examples
   - Authentication requirements
   - Environment variables

2. **Official documentation** - Best for:

   - Complete API/tool specifications
   - Transport protocol details
   - Authentication flows (OAuth, API keys)
   - Configuration options

3. **Package/Container registry** - Useful for:
   - Verifying image references
   - Checking available versions
   - Finding pull counts

**What to extract:**

- `tools`: List of function names the server exposes
- `description`: One-line summary from README or docs
- `transport`: Protocol type (stdio/sse/streamable-http)
- `environmentVariables`: Required environment variables
- `auth`: Authentication method (headers/OAuth)
- `permissions`: Network hosts (never filesystem paths)
- `provenance`: Supply chain security details (Sigstore/cosign signatures)

**Best practices:**

- **Ask first** if uncertain whether to fetch (unless clearly needed)
- **Prioritize** official docs over README when both exist
- **Extract** only spec-relevant information, not implementation details
- **Verify** transport type for remote servers (streamable-http preferred)
- **Skip** if user has already provided comprehensive details

### 5. Create the icon.svg File

Create an SVG icon for the server at `registries/toolhive/servers/<server-name>/icon.svg`.

- Use the service's official logo/icon when available
- Keep it simple and recognizable at small sizes
- Use a standard SVG format

### 6. Create the server.json File

Create `registries/toolhive/servers/<server-name>/server.json` using the appropriate template below.

---

## server.json Structure

Every `server.json` file has two main sections:

1. **Top-level fields** — Standard MCP ServerJSON fields (`name`, `description`, `packages`/`remotes`, `icons`, etc.)
2. **`_meta` extensions** — ToolHive-specific data (tier, status, tags, tools, overview, permissions, etc.)

The `_meta` extensions are nested at:
```
_meta["io.modelcontextprotocol.registry/publisher-provided"]["io.github.stacklok"]["<extension-key>"]
```

The **extension key** must match:
- For containers: the `packages[0].identifier` value
- For remotes: the `remotes[0].url` value

---

## Container-Based Servers

### Minimal Required Fields

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/my-server",
  "description": "Clear, concise one-line description",
  "title": "My Server",
  "repository": {
    "url": "https://github.com/org/server",
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
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/my-server/icon.svg",
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
          "tools": ["tool_name"],
          "overview": "## My Server\n\nA brief markdown description of the server's purpose and capabilities."
        }
      }
    }
  }
}
```

### Complete Template

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

### Container Transport Variants

Containers support two transport patterns:

**stdio (most common — ~75% of entries):**
```json
"transport": {
  "type": "stdio"
}
```

**streamable-http (used when the server exposes an HTTP endpoint inside the container):**
```json
"transport": {
  "type": "streamable-http",
  "url": "http://localhost:8080"
}
```

Use `streamable-http` when the containerized server runs an HTTP listener rather than communicating over stdin/stdout.

---

## Remote Servers

### Minimal Required Fields

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/my-remote-server",
  "description": "Clear, concise one-line description",
  "title": "My Remote Server",
  "repository": {
    "url": "https://github.com/org/server",
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
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/my-remote-server/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://api.example.com/mcp": {
          "tier": "Official",
          "status": "Active",
          "tools": ["tool_name"],
          "overview": "## My Remote Server\n\nA brief markdown description of the server's purpose and capabilities."
        }
      }
    }
  }
}
```

### Complete Template

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

---

## Field Selection Guide

### Top-Level Fields

**Always include:**

- `$schema` — Always `"https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json"`
- `name` — Format: `"io.github.stacklok/<server-name>"`
- `description` — One-line, clear
- `title` — Human-readable display name (e.g., `"Fetch"`, `"GitHub (Remote)"`, `"AWS Knowledge Bases"`)
- `version` — `"1.0.0"` for new entries
- `repository.url` and `repository.source`
- `icons` — Array with one entry pointing to the server's `icon.svg` file (see format below)

**Icons format (always the same structure):**

```json
"icons": [
  {
    "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/<server-name>/icon.svg",
    "mimeType": "image/svg+xml",
    "sizes": ["any"]
  }
]
```

**For containers (`packages`):**

- `packages[0].registryType` — Always `"oci"`
- `packages[0].identifier` — Full image reference with tag (e.g., `ghcr.io/org/server:v1.0.0`)
- `packages[0].transport.type` — `"stdio"` or `"streamable-http"` (or `"sse"`)
- `packages[0].transport.url` — Required when transport type is `"streamable-http"` (typically `"http://localhost:8080"`)
- `packages[0].environmentVariables` — When the server needs API keys or configuration

**For remotes (`remotes`):**

- `remotes[0].type` — `"streamable-http"` or `"sse"` (NOT `"stdio"`)
- `remotes[0].url` — HTTPS endpoint

### Environment Variables

Environment variables support these fields:

| Field         | Type    | Description                                    |
| ------------- | ------- | ---------------------------------------------- |
| `name`        | string  | Variable name (e.g., `"API_KEY"`)              |
| `description` | string  | What the variable is for                       |
| `isRequired`  | boolean | Whether the variable must be set               |
| `isSecret`    | boolean | Whether the value is sensitive (keys, tokens)  |
| `default`     | string  | Default value if not provided by user          |

### Extensions (`_meta`) Fields

**Always include:**

- `tier` — `"Official"` or `"Community"`
- `status` — `"Active"` or `"Deprecated"`
- `tools` — List of tool names (or `["set_during_runtime"]` if dynamic)
- `overview` — Markdown-formatted description (see [Overview Field Format](#overview-field-format))

**Highly recommended:**

- `tags` — Categorization (include `"remote"` for remote servers)

**Include when applicable:**

- `permissions.network` — For network access (NEVER filesystem paths; see [Network Permissions](#network-permissions))
- `provenance` — Supply chain security (Sigstore/cosign)
- `custom_metadata` — Author, homepage, license
- `oauth_config` — For remote servers requiring OAuth (see [OAuth Configuration](#oauth-configuration))

**Auto-populated by CI (do NOT include for new entries):**

- `metadata.stars` — GitHub stars (updated daily)
- `metadata.pulls` — Docker pull count (updated daily)
- `metadata.last_updated` — Timestamp (updated on any CI change)
- `tool_definitions` — Full MCP Tool objects with `inputSchema` (extracted by running the server on PR)

Note: The `tools` list is also auto-updated by CI when a PR modifies a server.json. You should still include it in new entries with your best-effort list of tool names, and CI will verify/update it.

### Overview Field Format

The `overview` field is a markdown string describing the server. Follow this pattern:

```
## Server Display Name\n\nA 3-5 sentence description of the server's purpose, key capabilities, and typical use cases.
```

Example:
```json
"overview": "## Heroku MCP Server\n\nThe heroku-mcp-server is a Model Context Protocol (MCP) server that enables AI assistants and agents to interact directly with the Heroku platform. Key capabilities include application lifecycle management, deployment workflows, runtime insight for dynos and process scaling, and operational visibility through logs and platform metadata."
```

### Network Permissions

Use `permissions.network.outbound` to declare what hosts the server needs to reach.

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

**Wildcard pattern:** Prefix a hostname with `.` to allow all subdomains (e.g., `".github.com"` allows `api.github.com`, `raw.github.com`, etc.).

**Broad network access** — use `insecure_allow_all` when the server needs to reach arbitrary hosts (e.g., a general-purpose fetch/HTTP server):
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

**NEVER** include filesystem paths in permissions — only network configuration.

### OAuth Configuration

For remote servers requiring OAuth authentication, include `oauth_config`:

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

---

## Common Patterns & Examples

### Pattern 1: Container with stdio Transport

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/heroku",
  "description": "MCP server for seamless interaction between LLMs and the Heroku Platform",
  "title": "Heroku",
  "repository": {
    "url": "https://github.com/heroku/heroku-mcp-server",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/stacklok/dockyard/npx/heroku-mcp-server:1.0.7",
      "transport": { "type": "stdio" },
      "environmentVariables": [
        {
          "name": "HEROKU_API_KEY",
          "description": "Your Heroku authorization token",
          "isRequired": true,
          "isSecret": true
        },
        {
          "name": "MCP_SERVER_REQUEST_TIMEOUT",
          "description": "Timeout in milliseconds for command execution",
          "default": "15000"
        }
      ]
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/heroku/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/stacklok/dockyard/npx/heroku-mcp-server:1.0.7": {
          "tier": "Official",
          "status": "Active",
          "tags": ["heroku", "paas", "deployment", "cloud", "devops"],
          "tools": ["list_apps", "get_app_info", "create_app", "deploy_to_heroku", "ps_list", "ps_scale"],
          "overview": "## Heroku MCP Server\n\nThe heroku-mcp-server enables AI assistants and agents to interact directly with the Heroku platform. Key capabilities include application lifecycle management, deployment workflows, runtime insight for dynos and process scaling, and operational visibility through logs and platform metadata.",
          "permissions": {
            "network": {
              "outbound": {
                "allow_host": [".heroku.com", ".herokuapp.com"],
                "allow_port": [443]
              }
            }
          },
          "provenance": {
            "cert_issuer": "https://token.actions.githubusercontent.com",
            "repository_uri": "https://github.com/stacklok/dockyard",
            "runner_environment": "github-hosted",
            "signer_identity": "/.github/workflows/build-containers.yml",
            "sigstore_url": "tuf-repo-cdn.sigstore.dev"
          }
        }
      }
    }
  }
}
```

### Pattern 2: Container with streamable-http Transport

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/sqlite",
  "description": "Provides tools and resources for querying SQLite databases.",
  "title": "SQLite",
  "repository": {
    "url": "https://github.com/StacklokLabs/sqlite-mcp",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/stackloklabs/sqlite-mcp/server:0.1.0",
      "transport": {
        "type": "streamable-http",
        "url": "http://localhost:8080"
      }
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/sqlite/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/stackloklabs/sqlite-mcp/server:0.1.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["database", "sqlite", "sql", "data", "query"],
          "tools": ["execute_query", "execute_statement", "list_tables", "describe_table"],
          "overview": "## SQLite MCP Server\n\nThe SQLite MCP server provides AI assistants with direct read and write access to SQLite databases. It supports executing arbitrary SQL queries and statements, listing all tables in a database, and describing table schemas."
        }
      }
    }
  }
}
```

### Pattern 3: Remote Server

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/semgrep-remote",
  "description": "Official Semgrep MCP server for code security scanning and vulnerability detection",
  "title": "Semgrep (Remote)",
  "repository": {
    "url": "https://github.com/semgrep/mcp",
    "source": "github"
  },
  "version": "1.0.0",
  "remotes": [
    {
      "type": "streamable-http",
      "url": "https://mcp.semgrep.ai/mcp"
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/semgrep-remote/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://mcp.semgrep.ai/mcp": {
          "tier": "Official",
          "status": "Active",
          "tags": ["remote", "security", "semgrep", "static-analysis", "code-scanning"],
          "tools": ["semgrep_scan", "semgrep_findings", "security_check", "get_abstract_syntax_tree"],
          "overview": "## Semgrep Remote MCP Server\n\nThe semgrep-remote MCP server is the official Semgrep MCP server for code security scanning and vulnerability detection. It enables vulnerability scanning utilizing semantic analysis, abstract syntax tree generation, and custom security rule development.",
          "custom_metadata": {
            "author": "Semgrep",
            "homepage": "https://mcp.semgrep.ai",
            "license": "MIT"
          }
        }
      }
    }
  }
}
```

### Pattern 4: Remote Server with OAuth

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/github-remote",
  "description": "GitHub's official MCP server for repositories, issues, PRs, actions, and security with OAuth",
  "title": "GitHub (Remote)",
  "repository": {
    "url": "https://github.com/github/github-mcp-server",
    "source": "github"
  },
  "version": "1.0.0",
  "remotes": [
    {
      "type": "streamable-http",
      "url": "https://api.githubcopilot.com/mcp"
    }
  ],
  "icons": [
    {
      "src": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/registries/toolhive/servers/github-remote/icon.svg",
      "mimeType": "image/svg+xml",
      "sizes": ["any"]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://api.githubcopilot.com/mcp": {
          "tier": "Official",
          "status": "Active",
          "tags": ["remote", "github", "git", "version-control", "repositories"],
          "tools": ["create_issue", "create_pull_request", "search_code", "get_file_contents"],
          "overview": "## GitHub Remote MCP Server\n\nThe github-remote MCP server is a hosted MCP server that allows AI assistants to interact with GitHub repositories and resources without running a local MCP process. It provides secure, permission-aware access through OAuth.",
          "oauth_config": {
            "authorize_url": "https://github.com/login/oauth/authorize",
            "token_url": "https://github.com/login/oauth/access_token",
            "scopes": ["repo", "user:email"],
            "use_pkce": true
          },
          "custom_metadata": {
            "author": "GitHub",
            "homepage": "https://github.com",
            "license": "MIT"
          }
        }
      }
    }
  }
}
```

---

## Validation Rules

### Container Servers

| Rule            | Description                                                                |
| --------------- | -------------------------------------------------------------------------- |
| Image format    | Must be valid OCI reference with tag (e.g., `ghcr.io/org/name:v1.0.0`)    |
| Transport       | Must be `"stdio"`, `"sse"`, or `"streamable-http"`                         |
| Transport URL   | Required when transport type is `"streamable-http"` (e.g., `"http://localhost:8080"`) |
| Required fields | `$schema`, `name`, `description`, `title`, `version`, `packages`, `icons`, `_meta` |
| Extension key   | Must match `packages[0].identifier` exactly                                |

### Remote Servers

| Rule            | Description                                                                |
| --------------- | -------------------------------------------------------------------------- |
| URL format      | Must be valid HTTPS URL                                                    |
| Transport       | Must be `"streamable-http"` or `"sse"` (NOT `"stdio"`)                     |
| Required fields | `$schema`, `name`, `description`, `title`, `version`, `remotes`, `icons`, `_meta` |
| Extension key   | Must match `remotes[0].url` exactly                                        |
| Tags            | Should include `"remote"` tag                                              |

### Both Types

| Rule          | Description                                   |
| ------------- | --------------------------------------------- |
| Server name   | Lowercase letters, numbers, hyphens only      |
| Name format   | `io.github.stacklok/<server-name>`            |
| Description   | Single line, clear, concise                   |
| Title         | Human-readable display name                   |
| Tier values   | Exactly `"Official"` or `"Community"`         |
| Status values | Exactly `"Active"` or `"Deprecated"`          |
| Overview      | Markdown string with `## Title` heading       |
| Icons         | Must include `icons` array with `icon.svg` ref |
| JSON syntax   | Valid JSON, 2-space indentation               |

---

## Post-Creation Steps

### 1. Validate Against Schema

Validate that your server.json file conforms to the schema:

```bash
task catalog:validate
```

This checks that all required fields are present and properly formatted.

### 2. Build and Test Registry

Ensure the registry builds successfully with your new entry:

```bash
task catalog:build
```

This compiles all entries into the registry and verifies there are no conflicts or errors.

### 3. Verify Entry

```bash
# For container servers:
jq '.servers["<server-name>"]' build/toolhive/registry.json

# For remote servers:
jq '.remote_servers["<server-name>"]' build/toolhive/registry.json
```

---

## Troubleshooting

### Common Errors

**"Extension key mismatch"**

- The key in `_meta` must exactly match `packages[0].identifier` (containers) or `remotes[0].url` (remotes)
- Including the version tag: `ghcr.io/org/server:v1.0.0`, not `ghcr.io/org/server`

**"Invalid transport type"**

- Container: Use `"stdio"`, `"streamable-http"`, or `"sse"`
- Remote: Use `"streamable-http"` or `"sse"` (NOT `"stdio"`)

**"Missing required fields"**

- Check `$schema`, `name`, `description`, `title`, `version`, `icons`
- Container: Check `packages` array
- Remote: Check `remotes` array

**"Invalid tier or status"**

- Tier: Must be exactly `"Official"` or `"Community"`
- Status: Must be exactly `"Active"` or `"Deprecated"`

**"Invalid JSON"**

- Ensure proper JSON syntax (no trailing commas, quoted keys)
- Use 2-space indentation
- Validate with `jq . < server.json`

---

## Reference Examples

Study these existing entries:

- **Container, stdio, API**: `registries/toolhive/servers/github/server.json`
- **Container, streamable-http, Database**: `registries/toolhive/servers/sqlite/server.json`
- **Container, streamable-http, Network**: `registries/toolhive/servers/fetch/server.json`
- **Remote, Official**: `registries/toolhive/servers/semgrep-remote/server.json`
- **Remote, AWS**: `registries/toolhive/servers/aws-knowledge/server.json`
- **Remote, OAuth**: `registries/toolhive/servers/github-remote/server.json`

---

## Final Checklist

Before submitting:

- [ ] Server name: lowercase, numbers, hyphens only
- [ ] Directory: `registries/toolhive/servers/<server-name>/` exists
- [ ] File: named exactly `server.json`
- [ ] Icon: `icon.svg` file present in server directory
- [ ] `$schema` field present
- [ ] `name` follows format `io.github.stacklok/<server-name>`
- [ ] `description` is clear and concise
- [ ] `title` is a human-readable display name
- [ ] `version` is `"1.0.0"`
- [ ] `repository.url` and `repository.source` are set
- [ ] `icons` array present with correct `src` URL pointing to `icon.svg`
- [ ] Container: `packages` with `registryType`, `identifier`, `transport`
- [ ] Container: `transport.url` included if transport type is `streamable-http`
- [ ] Remote: `remotes` with `type` and `url`
- [ ] Transport: appropriate for server type
- [ ] `_meta` extensions block present with correct nesting
- [ ] Extension key matches `identifier` (containers) or `url` (remotes)
- [ ] `tier`: "Official" or "Community"
- [ ] `status`: "Active" or "Deprecated"
- [ ] `tools`: at least one tool listed
- [ ] `overview`: markdown description with `## Title` heading
- [ ] Tags: relevant categories included (and `"remote"` for remote servers)
- [ ] Network permissions: specified if needed (NEVER include filesystem paths)
- [ ] OAuth config: included for remote servers requiring OAuth
- [ ] `metadata` fields (stars, pulls, last_updated) are NOT included (auto-populated by CI)
- [ ] `tool_definitions` is NOT included (auto-populated by CI)
- [ ] Validation: `task catalog:validate` passes without errors
- [ ] Registry build: `task catalog:build` completes successfully

---

## Decision Tree for LLMs

```
Start
  |
Do you have complete information (tools, auth, description)?
  |-- No --> Gather information
  |   |-- Check repository README
  |   |-- Review official documentation
  |   +-- Extract: tools, description, transport, auth, env vars
  |
  +-- Yes --> Proceed to server type
      |
Is this a Docker container or HTTP endpoint?
  |-- Docker --> Use Container template
  |   |-- Set: packages with identifier, transport
  |   |-- Choose transport: stdio (default) or streamable-http (if HTTP listener)
  |   |-- Set: icons with icon.svg reference
  |   |-- Set: _meta extensions with matching key
  |   |-- Set: overview with markdown description
  |   |-- Need API keys? --> Add environmentVariables with isSecret: true
  |   +-- Need network access? --> Add permissions.network (NEVER filesystem paths)
  |       |-- Specific hosts? --> allow_host + allow_port
  |       +-- Any host? --> insecure_allow_all: true
  |
  +-- HTTP --> Use Remote template
      |-- Set: remotes with url, type (streamable-http preferred)
      |-- Set: icons with icon.svg reference
      |-- Set: _meta extensions with matching key (= the URL)
      |-- Set: overview with markdown description
      |-- Tags: include "remote"
      |-- Needs OAuth? --> Add oauth_config with authorize_url, token_url, scopes
      +-- Validate with task catalog:validate
```
