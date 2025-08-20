# Instructions for LLMs: Adding MCP Server Entries to ToolHive Registry

## Context
You are helping to add an MCP (Model Context Protocol) server entry to the ToolHive registry. Each entry defines a server that provides tools and capabilities to AI assistants.

## Task Overview
Create a YAML specification file for an MCP server in the correct directory structure. MCP servers can be either container-based (Docker images) or remote (HTTP/HTTPS endpoints).

## Step-by-Step Process

### 1. Determine the Server Name
- Use lowercase letters, numbers, and hyphens only
- Choose a descriptive, unique name
- Examples: `github`, `aws-pricing`, `sqlite`, `notion`

### 2. Create Directory Structure
```bash
mkdir registry/<server-name>
```

### 3. Create spec.yaml File
Create `registry/<server-name>/spec.yaml` with the appropriate structure based on server type:

#### For Container-based Servers

##### Minimal Required Fields
```yaml
image: <docker-image-reference>  # e.g., ghcr.io/org/server:v1.0.0
description: <one-line-description>  # Clear, concise explanation
transport: <transport-type>  # Usually "stdio", can be "sse" or "streamable-http"
```

#### For Remote Servers

##### Minimal Required Fields
```yaml
url: <server-endpoint>  # e.g., https://api.example.com/mcp
description: <one-line-description>  # Clear, concise explanation
transport: <transport-type>  # "sse" or "streamable-http" (NOT "stdio")
```

#### Complete Template with All Fields

##### Container-based Servers
```yaml
# Docker/OCI image reference (REQUIRED for container servers)
image: ghcr.io/organization/server-name:v1.0.0

# One-line description (REQUIRED)
description: Enables interaction with [service/API] for [purpose]

# Communication protocol (REQUIRED)
transport: stdio  # Most common, alternatives: "sse", "streamable-http"

# Source code repository (HIGHLY RECOMMENDED)
repository_url: https://github.com/organization/repository

# Project homepage/documentation (OPTIONAL)
homepage: https://docs.example.com

# License identifier (OPTIONAL)
license: MIT  # Common: MIT, Apache-2.0, GPL-3.0

# Author/organization (OPTIONAL)
author: Organization Name

# Classification tier (OPTIONAL, defaults to "Community")
tier: Community  # Options: "Official", "Community"

# Development status (OPTIONAL, defaults to "Active")
status: Active  # Options: "Active", "Beta", "Alpha", "Deprecated"

# Categorization tags (RECOMMENDED)
tags:
  - category1  # e.g., "database", "api", "productivity"
  - category2
  - category3

# List of tools provided (HIGHLY RECOMMENDED)
tools:
  - tool_name_1  # Actual function names the server exposes
  - tool_name_2
  - tool_name_3

# Environment variables (IF APPLICABLE)
env_vars:
  - name: API_KEY
    description: Authentication key for service
    required: true
    secret: true  # Mark as secret for sensitive data
  
  - name: BASE_URL
    description: API endpoint URL
    required: false
    default: "https://api.example.com"

# Command-line arguments (IF APPLICABLE)
args:
  - --verbose
  - --config=/path/to/config

# Security permissions (IF APPLICABLE)
permissions:
  # Network access
  network:
    outbound:
      allow_host:
        - api.example.com
        - auth.example.com
      allow_port:
        - 443
        - 80
      # insecure_allow_all: false  # Only set to true if absolutely necessary
  
  # File system access
  read:
    - /config
    - /data
  
  write:
    - /cache
    - /logs

# Usage metrics (OPTIONAL, auto-updated)
metrics:
  stars: 0  # GitHub stars
  pulls: 0  # Docker pulls
```

##### Remote Servers
```yaml
url: https://api.example.com/mcp/v1  # REQUIRED endpoint URL
description: Enables interaction with [service/API] for [purpose]  # REQUIRED
transport: sse  # REQUIRED: "sse" or "streamable-http" (NOT "stdio")
repository_url: https://github.com/organization/repository
tools:
  - tool_name_1
  - tool_name_2
tags:
  - remote
  - api

# Authentication options:
headers:
  - name: X-API-Key
    description: API key for authentication
    required: true
    secret: true

# Option 2: OAuth configuration
oauth_config:
  issuer: https://auth.example.com  # OIDC discovery
  authorize_url: https://auth.example.com/authorize  # Non-OIDC
  token_url: https://auth.example.com/token  # Non-OIDC
  client_id: mcp-client
  scopes:
    - read
    - write

```

## Field Selection Guidelines

### Always Include
**Container servers**: `image`, `description`, `transport` (usually "stdio")
**Remote servers**: `url`, `description`, `transport` ("sse" or "streamable-http")

### Include When Available
- `repository_url`, `tools`, `tags`

### Include When Needed
**Both types**: `env_vars` with `secret: true` for sensitive data
**Container servers**: `permissions`, `args`
**Remote servers**: `headers` or `oauth_config` for auth

### Tier/Status
- Tier: `"Official"`, `"Community"` (default)
- Status: `"Active"` (default), `"Beta"`, `"Alpha"`, `"Deprecated"`

## Validation Rules

**Container servers**:
- `image` must be valid Docker/OCI reference
- `transport` must be `"stdio"`, `"sse"`, or `"streamable-http"`

**Remote servers**:
- `url` must be valid HTTP/HTTPS URL
- `transport` must be `"sse"` or `"streamable-http"` (NOT `"stdio"`)

## Common Patterns

### Container-based API Integration Server
```yaml
image: ghcr.io/org/api-server:latest
description: Integrates with ExampleAPI for data retrieval and manipulation
transport: stdio
repository_url: https://github.com/org/api-server
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

### Container-based Database Server
```yaml
image: docker.io/org/db-server:latest
description: Provides tools for querying and managing PostgreSQL databases
transport: stdio
repository_url: https://github.com/org/db-server
tools:
  - execute_query
  - list_tables
  - describe_table
env_vars:
  - name: DATABASE_URL
    description: PostgreSQL connection string
    required: true
    secret: true
tags:
  - database
  - postgresql
  - sql
```

### Container-based File Processing Server
```yaml
image: ghcr.io/org/file-server:latest
description: Processes and analyzes various file formats
transport: stdio
repository_url: https://github.com/org/file-server
tools:
  - read_file
  - analyze_content
  - convert_format
permissions:
  read:
    - /input
  write:
    - /output
tags:
  - files
  - processing
  - conversion
```

## Post-Creation Steps

After creating the spec.yaml file:

1. **Validate the entry:**
   ```bash
   task validate
   ```

2. **Build the registry to verify inclusion:**
   ```bash
   task build:registry
   ```

3. **Check the generated entry:**
   ```bash
   # For container servers:
   jq '.servers["<server-name>"]' build/registry.json
   # For remote servers:
   jq '.remote_servers["<server-name>"]' build/registry.json
   ```

## Error Resolution

### Common Issues and Solutions

1. **Invalid transport type**
   - Ensure transport is exactly one of: `"stdio"`, `"sse"`, `"streamable-http"`

2. **Missing required fields**
   - Container: Verify `image`, `description`, and `transport` are present
   - Remote: Verify `url`, `description`, and `transport` are present

3. **Invalid tier or status**
   - Check spelling and capitalization match exactly

4. **YAML syntax errors**
   - Ensure proper indentation (2 spaces)
   - Quote strings containing special characters
   - Use proper list syntax with `-` for arrays

## Examples to Reference

Look at these existing entries for patterns:
- `registry/github/spec.yaml` - Complex API integration
- `registry/sqlite/spec.yaml` - Database server
- `registry/fetch/spec.yaml` - Simple tool server
- `registry/notion/spec.yaml` - Remote server example

## Final Checklist

Before completing:
- [ ] Server name uses only lowercase, numbers, hyphens
- [ ] Directory created at `registry/<server-name>/`
- [ ] File named exactly `spec.yaml`
- [ ] All required fields present (image/url, description, transport)
- [ ] Image reference complete with tag (container) OR URL valid (remote)
- [ ] Description is clear and concise
- [ ] Tools list matches actual server capabilities
- [ ] Environment variables documented if needed
- [ ] Permissions specified if network/filesystem access required
- [ ] Validation passes with `task validate`
