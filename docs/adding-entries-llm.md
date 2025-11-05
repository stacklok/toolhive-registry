# Instructions for LLMs: Adding MCP Server Entries to ToolHive Registry

## Context

You are helping to add an MCP (Model Context Protocol) server entry to the ToolHive registry. Each entry defines a server that provides tools and capabilities to AI assistants.

## Quick Reference

### Server Types

| Type      | Identifier     | Transport Options                      | Primary Use                   |
| --------- | -------------- | -------------------------------------- | ----------------------------- |
| Container | `image:` field | `stdio`, `streamable-http`, `sse`      | Self-hosted Docker containers |
| Remote    | `url:` field   | `streamable-http`, `sse` (NOT `stdio`) | HTTP/HTTPS API endpoints      |

### Field Priority Levels

- **REQUIRED**: Must be present or validation will fail
- **HIGHLY RECOMMENDED**: Should be included when available
- **OPTIONAL**: Include when relevant or needed

## Step-by-Step Process

### 1. Determine Server Type

**Ask yourself:** Is this a Docker container or a remote HTTP endpoint?

- If Docker/OCI image → Container-based server
- If HTTP/HTTPS API → Remote server

### 2. Choose Server Name

- Use **only** lowercase letters, numbers, and hyphens
- Make it descriptive and unique
- Examples: `github`, `aws-pricing`, `sqlite`, `notion-remote`

### 3. Create Directory

```bash
mkdir registry/<server-name>
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
- `env_vars`: Required environment variables
- `auth`: Authentication method (headers/OAuth)
- `permissions`: Network hosts (never filesystem paths)
- `provenance`: Supply chain security details (Sigstore/cosign signatures)

**Best practices:**

- **Ask first** if uncertain whether to fetch (unless clearly needed)
- **Prioritize** official docs over README when both exist
- **Extract** only spec-relevant information, not implementation details
- **Verify** transport type for remote servers (streamable-http preferred)
- **Skip** if user has already provided comprehensive details

### 5. Create Specification File

Create `registry/<server-name>/spec.yaml` using the appropriate template below.

---

## Container-Based Servers

### Minimal Required Fields

```yaml
image: ghcr.io/org/server:v1.0.0
description: Clear, concise one-line description
transport: stdio
tier: Community
status: Active
tools:
  - tool_name # Or use "set_during_runtime" if tools aren't knowable up-front
```

### Complete Template

```yaml
# ============================================
# REQUIRED FIELDS
# ============================================

# Docker/OCI image reference with tag
image: ghcr.io/organization/server-name:v1.0.0

# One-line description of server purpose
description: Enables interaction with [service/API] for [purpose]

# Communication protocol
transport: stdio # Options: "stdio", "streamable-http", "sse"

# Classification tier
tier: Community # Options: "Official", "Community"

# Development status
status: Active # Options: "Active", "Deprecated"

# List of tools this server provides (at least one required)
# Use actual tool names if known, or "set_during_runtime" if not discoverable
tools:
  - tool_name_1
  - tool_name_2
  - tool_name_3

# ============================================
# HIGHLY RECOMMENDED FIELDS
# ============================================

# Source code repository URL
repository_url: https://github.com/organization/repository

# Categorization tags
tags:
  - category1 # e.g., "database", "api", "productivity"
  - category2
  - category3

# ============================================
# OPTIONAL FIELDS
# ============================================

# Project homepage or documentation
homepage: https://docs.example.com

# License identifier (SPDX format)
license: MIT # Common: MIT, Apache-2.0, GPL-3.0

# Author or organization name
author: Organization Name

# Server name (defaults to directory name if omitted)
name: server-name

# ============================================
# CONDITIONAL FIELDS (include if applicable)
# ============================================

# Target port (REQUIRED for streamable-http or sse transports)
target_port: 8080

# Environment variables (for API keys, config, etc.)
env_vars:
  - name: API_KEY
    description: Authentication key for service
    required: true
    secret: true # Mark sensitive data as secret

  - name: BASE_URL
    description: API endpoint URL
    required: false
    default: "https://api.example.com"

# Command-line arguments
args:
  - --verbose
  - --config=/path/to/config

# Security permissions
permissions:
  # Network access control
  network:
    outbound:
      allow_host:
        - api.example.com
        - auth.example.com
      allow_port:
        - 443
        - 80
      # insecure_allow_all: false  # Only use if absolutely necessary

  # IMPORTANT: Do NOT specify filesystem paths in registry entries
  # Mounting host directories is a security risk and should be
  # configured by users at runtime, not in registry specs

# Usage metrics (auto-updated - typically omit when creating new entries)
metadata:
  stars: 0
  pulls: 0
  last_updated: 2025-01-01T00:00:00Z

# Provenance for supply chain security
provenance:
  cert_issuer: https://token.actions.githubusercontent.com
  repository_uri: https://github.com/org/repository
  runner_environment: github-hosted
  signer_identity: /.github/workflows/build-containers.yml
  sigstore_url: tuf-repo-cdn.sigstore.dev
```

---

## Remote Servers

### Minimal Required Fields

```yaml
url: https://api.example.com/mcp
description: Clear, concise one-line description
transport: streamable-http
tier: Official
status: Active
tools:
  - tool_name # Or use "set_during_runtime" if tools aren't knowable up-front
```

### Complete Template

```yaml
# ============================================
# REQUIRED FIELDS
# ============================================

# Remote server endpoint URL
url: https://api.example.com/mcp/v1

# One-line description of server purpose
description: Enables interaction with [service/API] for [purpose]

# Communication protocol (NOT "stdio" for remote servers)
transport: streamable-http # Options: "streamable-http", "sse" (deprecated)

# Classification tier (REQUIRED)
tier: Official # Options: "Official", "Community"

# Development status (REQUIRED)
status: Active # Options: "Active", "Deprecated"

# List of tools this server provides (at least one required)
# Use actual tool names if known, or "set_during_runtime" if not discoverable
tools:
  - tool_name_1
  - tool_name_2

# ============================================
# HIGHLY RECOMMENDED FIELDS
# ============================================

# Source code repository (if open source)
repository_url: https://github.com/organization/repository

# Categorization tags (include "remote")
tags:
  - remote # Always include for remote servers
  - api
  - integration

# ============================================
# OPTIONAL FIELDS
# ============================================

# Project homepage or documentation
homepage: https://docs.example.com

# Author or organization name
author: Organization Name

# ============================================
# AUTHENTICATION (choose one method)
# ============================================

# Option 1: Header-based authentication (API keys, tokens)
headers:
  - name: X-API-Key
    description: API key for authentication
    required: true
    secret: true

# Option 2: OAuth 2.0 / OIDC configuration
oauth_config:
  issuer: https://auth.example.com # For OIDC discovery
  authorize_url: https://auth.example.com/authorize # For non-OIDC
  token_url: https://auth.example.com/token # For non-OIDC
  client_id: mcp-client
  scopes:
    - read
    - write

# Usage metrics (auto-updated - typically omit when creating new entries)
metadata:
  stars: 0
  last_updated: 2025-01-01T00:00:00Z
```

---

## Field Selection Guide

### For Container Servers

**Always include:**

- `image` (with version tag)
- `description` (one-line, clear)
- `transport` (`"stdio"`, `"streamable-http"`, or `"sse"`)
- `tier` (`"Official"` or `"Community"`)
- `status` (`"Active"` or `"Deprecated"`)
- `tools` (at least one tool name, or `set_during_runtime` if unknown)

**Highly recommended:**

- `repository_url` (GitHub/GitLab link)
- `tags` (categorization)

**Include when needed:**

- `target_port` (REQUIRED when transport is `"streamable-http"` or `"sse"`)
- `name` (optional - defaults to directory name if omitted)
- `env_vars` (for API keys, configuration)
- `permissions.network` (for network access only - NEVER specify filesystem paths)
- `args` (command-line arguments)
- `provenance` (for supply chain security verification - Sigstore/cosign signatures)

### For Remote Servers

**Always include:**

- `url` (HTTPS endpoint)
- `description` (one-line, clear)
- `transport` (`"streamable-http"` or `"sse"`)
- `tier` (`"Official"` or `"Community"`)
- `status` (`"Active"` or `"Deprecated"`)
- `tools` (at least one tool name, or `set_during_runtime` if unknown)

**Highly recommended:**

- `repository_url` (if open source)
- `tags` (always include `"remote"`)

**Include when needed:**

- `headers` or `oauth_config` (authentication)
- `homepage` (documentation link)

---

## Common Patterns & Examples

### Pattern 1: Container-Based API Integration

```yaml
image: ghcr.io/org/api-server:v1.2.0
description: Integrates with ExampleAPI for data retrieval and manipulation
transport: stdio
repository_url: https://github.com/org/api-server
homepage: https://example.com/docs
author: Example Organization
tier: Community
status: Active
tools:
  - fetch_data
  - create_record
  - update_record
  - delete_record
env_vars:
  - name: API_KEY
    description: API key from example.com
    required: true
    secret: true
permissions:
  network:
    outbound:
      allow_host:
        - api.example.com
      allow_port:
        - 443
tags:
  - api
  - integration
  - data
```

### Pattern 2: Container-Based Database Tool

```yaml
image: docker.io/org/db-server:v2.0.0
description: Provides tools for querying and managing PostgreSQL databases
transport: stdio
repository_url: https://github.com/org/db-server
license: Apache-2.0
tier: Community
status: Active
tools:
  - execute_query
  - list_tables
  - describe_table
  - get_schema
env_vars:
  - name: DATABASE_URL
    description: PostgreSQL connection string
    required: true
    secret: true
tags:
  - database
  - postgresql
  - sql
  - data
```

### Pattern 3: Container-Based File Processor

```yaml
image: ghcr.io/org/file-server:v1.0.0
description: Processes and analyzes various file formats
transport: streamable-http
target_port: 8080
repository_url: https://github.com/org/file-server
tier: Community
status: Active
tools:
  - read_file
  - analyze_content
  - convert_format
  - extract_metadata
tags:
  - files
  - processing
  - conversion
# NOTE: File access should be configured by users at runtime,
# not specified in registry entries for security reasons
```

### Pattern 4: Remote API Server

```yaml
url: https://knowledge-api.example.com/mcp
description: Documentation and knowledge base API for technical content
transport: streamable-http
repository_url: https://github.com/org/knowledge-mcp
homepage: https://docs.example.com/mcp
author: Example Inc
tier: Official
status: Active
tools:
  - search_documentation
  - get_article
  - list_categories
tags:
  - remote
  - documentation
  - api
  - knowledge
headers:
  - name: X-API-Key
    description: API authentication key
    required: true
    secret: true
```

### Pattern 5: Remote OAuth Service

```yaml
url: https://api.service.com/mcp/v2
description: Integration with ServiceAPI using OAuth authentication
transport: streamable-http
repository_url: https://github.com/org/service-mcp
homepage: https://service.com/mcp-docs
author: Service Inc
tier: Official
status: Active
tools:
  - query_data
  - create_resource
  - update_resource
tags:
  - remote
  - api
  - oauth
oauth_config:
  issuer: https://auth.service.com
  client_id: mcp-client-id
  scopes:
    - read
    - write
    - admin
```

---

## Validation Rules

### Container Servers

| Rule            | Description                                                                                            |
| --------------- | ------------------------------------------------------------------------------------------------------ |
| Image format    | Must be valid Docker/OCI reference (e.g., `registry/org/name:tag`)                                     |
| Transport       | Must be `"stdio"`, `"sse"`, or `"streamable-http"`                                                     |
| Required fields | Must have `image`, `description`, `transport`, `tier`, `status`, `tools` (≥1, or `set_during_runtime`) |

### Remote Servers

| Rule            | Description                                                                                          |
| --------------- | ---------------------------------------------------------------------------------------------------- |
| URL format      | Must be valid HTTP/HTTPS URL                                                                         |
| Transport       | Must be `"streamable-http"` or `"sse"` (NOT `"stdio"`)                                               |
| Required fields | Must have `url`, `description`, `transport`, `tier`, `status`, `tools` (≥1, or `set_during_runtime`) |
| Tags            | Should include `"remote"` tag                                                                        |

### Both Types

| Rule          | Description                              |
| ------------- | ---------------------------------------- |
| Server name   | Lowercase letters, numbers, hyphens only |
| Description   | Single line, clear, concise              |
| Tier values   | Exactly `"Official"` or `"Community"`    |
| Status values | Exactly `"Active"` or `"Deprecated"`     |
| YAML syntax   | 2-space indentation, proper list format  |

---

## Post-Creation Steps

### 1. Validate the Entry

```bash
task validate
```

### 2. Build Registry

```bash
task build:registry
```

### 3. Verify Entry

```bash
# For container servers:
jq '.servers["<server-name>"]' build/registry.json

# For remote servers:
jq '.remote_servers["<server-name>"]' build/registry.json
```

---

## Troubleshooting

### Common Errors

**"Invalid transport type"**

- Container: Use `"stdio"`, `"streamable-http"`, or `"sse"`
- Remote: Use `"streamable-http"` or `"sse"` (NOT `"stdio"`)

**"Missing required fields"**

- Container: Check `image`, `description`, `transport`
- Remote: Check `url`, `description`, `transport`

**"Invalid tier or status"**

- Tier: Must be exactly `"Official"` or `"Community"`
- Status: Must be exactly `"Active"` or `"Deprecated"`

**"YAML syntax error"**

- Use 2-space indentation (not tabs)
- Quote strings with special characters
- Use `-` for list items
- Ensure proper nesting

---

## Reference Examples

Study these existing entries:

- **Container, API**: `registry/github/spec.yaml`
- **Container, Database**: `registry/sqlite/spec.yaml`
- **Container, Simple**: `registry/fetch/spec.yaml`
- **Remote, Full-featured**: `registry/notion-remote/spec.yaml`
- **Remote, AWS**: `registry/aws-knowledge/spec.yaml`

---

## Final Checklist

Before submitting:

- [ ] Server name: lowercase, numbers, hyphens only
- [ ] Directory: `registry/<server-name>/` exists
- [ ] File: named exactly `spec.yaml`
- [ ] Required fields: present and valid
  - Container: `image`, `description`, `transport`, `tier`, `status`, `tools`
  - Remote: `url`, `description`, `transport`, `tier`, `status`, `tools`
- [ ] Image/URL: complete and correct
- [ ] Description: clear and concise
- [ ] Transport: appropriate for server type
  - Container: `stdio`, `streamable-http`, or `sse`
  - Remote: `streamable-http` or `sse` (NOT `stdio`)
- [ ] Target port: specified if transport is `streamable-http` or `sse` (containers only)
- [ ] Tier: set to "Official" or "Community"
- [ ] Status: set to "Active" or "Deprecated"
- [ ] Tools: at least one tool listed (actual tool names if known, or `set_during_runtime` if tools can't be determined)
- [ ] Tags: relevant categories included (and "remote" for remote servers)
- [ ] Auth: configured if needed
- [ ] Network permissions: specified if needed (NEVER include filesystem paths)
- [ ] Validation: `task validate` passes

---

## Decision Tree for LLMs

```
Start
  ↓
Do you have complete information (tools, auth, description)?
  ├─ No → Gather information
  │   ├─ Check repository README
  │   ├─ Review official documentation
  │   └─ Extract: tools, description, transport, auth, env_vars
  │
  └─ Yes → Proceed to server type
      ↓
Is this a Docker container or HTTP endpoint?
  ├─ Docker → Use Container template
  │   ├─ Add: image, description, transport (stdio)
  │   ├─ Add: repository_url, tools, tags
  │   ├─ Need API keys? → Add env_vars with secret: true
  │   └─ Need network access? → Add permissions.network (NEVER filesystem paths)
  │
  └─ HTTP → Use Remote template
      ├─ Add: url, description, transport (streamable-http or sse)
      ├─ Add: repository_url, tools, tags (include "remote")
      ├─ Need auth?
      │   ├─ API key → Add headers
      │   └─ OAuth → Add oauth_config
      └─ Validate with task validate
```
