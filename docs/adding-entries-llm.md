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

## Step-by-Step Process

### 1. Determine Server Type

**Ask yourself:** Is this a Docker container or a remote HTTP endpoint?

- If Docker/OCI image → Container-based server (uses `packages`)
- If HTTP/HTTPS API → Remote server (uses `remotes`)

### 2. Choose Server Name

- Use **only** lowercase letters, numbers, and hyphens
- Make it descriptive and unique
- Examples: `github`, `aws-pricing`, `sqlite`, `notion-remote`

### 3. Create Directory

```bash
mkdir -p registries/toolhive/servers/<server-name>
```

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

### 5. Create the server.json File

Create `registries/toolhive/servers/<server-name>/server.json` using the appropriate template below.

---

## server.json Structure

Every `server.json` file has two main sections:

1. **Top-level fields** - Standard MCP ServerJSON fields (`name`, `description`, `packages`/`remotes`, etc.)
2. **`_meta` extensions** - ToolHive-specific data (tier, status, tags, tools, permissions, etc.)

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
  "title": "my-server",
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
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/org/server:v1.0.0": {
          "tier": "Community",
          "status": "Active",
          "tools": ["tool_name"]
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
  "title": "<server-name>",
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
          "description": "API endpoint URL (optional)"
        }
      ],
      "arguments": ["--verbose"]
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

---

## Remote Servers

### Minimal Required Fields

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/my-remote-server",
  "description": "Clear, concise one-line description",
  "title": "my-remote-server",
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
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://api.example.com/mcp": {
          "tier": "Official",
          "status": "Active",
          "tools": ["tool_name"]
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
  "title": "<server-name>",
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

- `$schema` - Always `"https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json"`
- `name` - Format: `"io.github.stacklok/<server-name>"`
- `description` - One-line, clear
- `title` - Same as the directory name
- `version` - `"1.0.0"` for new entries
- `repository.url` and `repository.source`

**For containers (`packages`):**

- `packages[0].registryType` - Always `"oci"`
- `packages[0].identifier` - Full image reference with tag (e.g., `ghcr.io/org/server:v1.0.0`)
- `packages[0].transport.type` - `"stdio"`, `"streamable-http"`, or `"sse"`
- `packages[0].environmentVariables` - When the server needs API keys or configuration

**For remotes (`remotes`):**

- `remotes[0].type` - `"streamable-http"` or `"sse"` (NOT `"stdio"`)
- `remotes[0].url` - HTTPS endpoint

### Extensions (`_meta`) Fields

**Always include:**

- `tier` - `"Official"` or `"Community"`
- `status` - `"Active"` or `"Deprecated"`
- `tools` - List of tool names (or `["set_during_runtime"]` if dynamic)

**Highly recommended:**

- `tags` - Categorization (include `"remote"` for remote servers)

**Include when applicable:**

- `permissions.network` - For network access only (NEVER filesystem paths)
- `provenance` - Supply chain security (Sigstore/cosign)
- `custom_metadata` - Author, homepage, license

**Auto-populated (do NOT include for new entries):**

- `metadata.stars` - GitHub stars (updated by CI)
- `metadata.pulls` - Docker pull count (updated by CI)
- `metadata.last_updated` - Timestamp (updated by CI)

---

## Common Patterns & Examples

### Pattern 1: Container-Based API Integration

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/api-server",
  "description": "Integrates with ExampleAPI for data retrieval and manipulation",
  "title": "api-server",
  "repository": {
    "url": "https://github.com/org/api-server",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/org/api-server:v1.2.0",
      "transport": { "type": "stdio" },
      "environmentVariables": [
        {
          "name": "API_KEY",
          "description": "API key from example.com",
          "isRequired": true,
          "isSecret": true
        }
      ]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/org/api-server:v1.2.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["api", "integration", "data"],
          "tools": ["fetch_data", "create_record", "update_record", "delete_record"],
          "permissions": {
            "network": {
              "outbound": {
                "allow_host": ["api.example.com"],
                "allow_port": [443]
              }
            }
          }
        }
      }
    }
  }
}
```

### Pattern 2: Container-Based Database Tool

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/db-server",
  "description": "Provides tools for querying and managing PostgreSQL databases",
  "title": "db-server",
  "repository": {
    "url": "https://github.com/org/db-server",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "docker.io/org/db-server:v2.0.0",
      "transport": { "type": "stdio" },
      "environmentVariables": [
        {
          "name": "DATABASE_URL",
          "description": "PostgreSQL connection string",
          "isRequired": true,
          "isSecret": true
        }
      ]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "docker.io/org/db-server:v2.0.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["database", "postgresql", "sql", "data"],
          "tools": ["execute_query", "list_tables", "describe_table", "get_schema"],
          "custom_metadata": {
            "license": "Apache-2.0"
          }
        }
      }
    }
  }
}
```

### Pattern 3: Remote API Server

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/knowledge-remote",
  "description": "Documentation and knowledge base API for technical content",
  "title": "knowledge-remote",
  "repository": {
    "url": "https://github.com/org/knowledge-mcp",
    "source": "github"
  },
  "version": "1.0.0",
  "remotes": [
    {
      "type": "streamable-http",
      "url": "https://knowledge-api.example.com/mcp"
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "https://knowledge-api.example.com/mcp": {
          "tier": "Official",
          "status": "Active",
          "tags": ["remote", "documentation", "api", "knowledge"],
          "tools": ["search_documentation", "get_article", "list_categories"],
          "custom_metadata": {
            "author": "Example Inc",
            "homepage": "https://docs.example.com/mcp"
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
| Required fields | `$schema`, `name`, `description`, `title`, `version`, `packages`, `_meta`  |
| Extension key   | Must match `packages[0].identifier` exactly                                |

### Remote Servers

| Rule            | Description                                                                |
| --------------- | -------------------------------------------------------------------------- |
| URL format      | Must be valid HTTPS URL                                                    |
| Transport       | Must be `"streamable-http"` or `"sse"` (NOT `"stdio"`)                     |
| Required fields | `$schema`, `name`, `description`, `title`, `version`, `remotes`, `_meta`   |
| Extension key   | Must match `remotes[0].url` exactly                                        |
| Tags            | Should include `"remote"` tag                                              |

### Both Types

| Rule          | Description                              |
| ------------- | ---------------------------------------- |
| Server name   | Lowercase letters, numbers, hyphens only |
| Name format   | `io.github.stacklok/<server-name>`       |
| Description   | Single line, clear, concise              |
| Tier values   | Exactly `"Official"` or `"Community"`    |
| Status values | Exactly `"Active"` or `"Deprecated"`     |
| JSON syntax   | Valid JSON, 2-space indentation          |

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

- Check `$schema`, `name`, `description`, `title`, `version`
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

- **Container, API**: `registries/toolhive/servers/github/server.json`
- **Container, Database**: `registries/toolhive/servers/sqlite/server.json`
- **Container, Simple**: `registries/toolhive/servers/fetch/server.json`
- **Remote, Full-featured**: `registries/toolhive/servers/semgrep-remote/server.json`
- **Remote, AWS**: `registries/toolhive/servers/aws-knowledge/server.json`

---

## Final Checklist

Before submitting:

- [ ] Server name: lowercase, numbers, hyphens only
- [ ] Directory: `registries/toolhive/servers/<server-name>/` exists
- [ ] File: named exactly `server.json`
- [ ] `$schema` field present
- [ ] `name` follows format `io.github.stacklok/<server-name>`
- [ ] `description` is clear and concise
- [ ] `title` matches directory name
- [ ] `version` is `"1.0.0"`
- [ ] `repository.url` and `repository.source` are set
- [ ] Container: `packages` with `registryType`, `identifier`, `transport`
- [ ] Remote: `remotes` with `type` and `url`
- [ ] Transport: appropriate for server type
- [ ] `_meta` extensions block present with correct nesting
- [ ] Extension key matches `identifier` (containers) or `url` (remotes)
- [ ] `tier`: "Official" or "Community"
- [ ] `status`: "Active" or "Deprecated"
- [ ] `tools`: at least one tool listed
- [ ] Tags: relevant categories included (and "remote" for remote servers)
- [ ] Network permissions: specified if needed (NEVER include filesystem paths)
- [ ] `metadata` fields (stars, pulls, last_updated) are NOT included (auto-populated by CI)
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
  |   |-- Set: packages with identifier, transport (stdio)
  |   |-- Set: _meta extensions with matching key
  |   |-- Need API keys? --> Add environmentVariables with isSecret: true
  |   +-- Need network access? --> Add permissions.network (NEVER filesystem paths)
  |
  +-- HTTP --> Use Remote template
      |-- Set: remotes with url, type (streamable-http)
      |-- Set: _meta extensions with matching key (= the URL)
      |-- Tags: include "remote"
      +-- Validate with task catalog:validate
```
