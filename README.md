# ToolHive Catalog

This repository contains the catalog of MCP (Model Context Protocol) servers and skills available for ToolHive. Each server entry provides AI assistants with specialized tools and capabilities, while skills provide reusable prompts and workflows that leverage those tools.

## What is this?

Think of this as a catalog of tools and skills that AI assistants can use. Each server entry represents a server that provides specific capabilities—like interacting with GitHub, querying databases, or fetching web content. Each skill entry is a reusable prompt or workflow that combines server tools to accomplish a specific task—like reviewing pull requests or debugging issues.

## How to Add Your MCP Server

Adding your MCP server to the registry is simple! You need to create a `server.json` file following the [upstream MCP ServerJSON schema](https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json), with ToolHive-specific extensions in the `_meta` field. We support two types of MCP servers:

1. **Container-based servers** - Run as Docker containers (use `packages` field)
2. **Remote servers** - Accessed via HTTP/HTTPS endpoints (use `remotes` field)

### Step 1: Create a Folder

Create a new folder in `registries/toolhive/servers/` with your server's name (use lowercase and hyphens):

```
registries/
  └── toolhive/
      └── servers/
      │   └── my-awesome-server/
      │       └── server.json
      └── skills/
          └── my-skill/
              ├── skill.json
              ├── icon.svg
              └── skill/          # installable content (referenced by subfolder in skill.json)
                  └── SKILL.md
```

### Step 2: Create Your server.json File

Choose the appropriate format based on your server type:

#### For Container-based Servers

Create a `server.json` file:

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/my-server",
  "description": "What your server does in one sentence",
  "title": "my-server",
  "repository": {
    "url": "https://github.com/myorg/my-server",
    "source": "github"
  },
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/myorg/my-server:v1.0.0",
      "transport": {
        "type": "stdio"
      },
      "environmentVariables": [
        {
          "name": "API_KEY",
          "description": "Your API key",
          "isRequired": true,
          "isSecret": true
        }
      ]
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/myorg/my-server:v1.0.0": {
          "tier": "Community",
          "status": "Active",
          "tags": ["api", "integration"],
          "tools": ["tool_name_1", "tool_name_2"]
        }
      }
    }
  }
}
```

#### For Remote Servers

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/my-remote-server",
  "description": "What your server does in one sentence",
  "title": "my-remote-server",
  "repository": {
    "url": "https://github.com/myorg/my-server",
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
          "tier": "Community",
          "status": "Active",
          "tags": ["remote", "api"],
          "tools": ["tool_name_1", "tool_name_2"]
        }
      }
    }
  }
}
```

### Key Concepts

**The `_meta` extensions block** holds ToolHive-specific data (tier, status, tags, tools, permissions, etc.) nested under `io.modelcontextprotocol.registry/publisher-provided` > `io.github.stacklok` > `<extension-key>`.

The **extension key** must match:
- For containers: the `packages[0].identifier` value (e.g., `ghcr.io/myorg/my-server:v1.0.0`)
- For remotes: the `remotes[0].url` value (e.g., `https://api.example.com/mcp`)

### Step 3: Add More Details (Optional but Recommended)

You can add more information in the `_meta` extensions block:

```json
{
  "ghcr.io/myorg/my-server:v1.0.0": {
    "tier": "Community",
    "status": "Active",
    "tags": ["api", "integration", "productivity"],
    "tools": ["tool_name_1", "tool_name_2"],
    "permissions": {
      "network": {
        "outbound": {
          "allow_host": ["api.example.com"],
          "allow_port": [443]
        }
      }
    },
    "provenance": {
      "cert_issuer": "https://token.actions.githubusercontent.com",
      "repository_uri": "https://github.com/myorg/my-server",
      "runner_environment": "github-hosted",
      "signer_identity": "/.github/workflows/build.yml",
      "sigstore_url": "tuf-repo-cdn.sigstore.dev"
    },
    "custom_metadata": {
      "author": "My Organization",
      "homepage": "https://docs.example.com",
      "license": "MIT"
    }
  }
}
```

## How to Add a Skill

Skills are reusable prompts and workflows that leverage MCP server tools. A skill is defined by a `SKILL.md` file (with YAML frontmatter) and a `skill.json` registry entry.

### Step 1: Create the Directory Structure

Create a new folder in `registries/toolhive/skills/` with your skill's name. Skill content (SKILL.md and any bundled resources) lives in a `skill/` subfolder, separate from registry metadata (skill.json, icon.svg):

```bash
mkdir -p registries/toolhive/skills/my-skill/skill
```

### Step 2: Create the SKILL.md File

Create this inside the `skill/` subfolder. It uses YAML frontmatter for metadata and markdown for the prompt/instructions:

```markdown
---
name: my-skill
description: What this skill does in one sentence.
version: "0.1.0"
allowed-tools:
  - server/tool_name_1
  - server/tool_name_2
license: Apache-2.0
metadata:
  author: Your Name
---

# My Skill

Instructions and prompts for the AI assistant go here...
```

### Step 3: Create the skill.json File

Create this in the skill's root directory (next to `icon.svg`, not inside `skill/`). This registers the skill in the catalog. Note that the `subfolder` points to the `skill/` subdirectory so that only skill content is installed — not registry metadata like `skill.json` and `icon.svg`:

```json
{
  "namespace": "io.github.stacklok",
  "name": "my-skill",
  "title": "My Skill",
  "description": "What this skill does in one sentence.",
  "version": "0.1.0",
  "status": "active",
  "license": "Apache-2.0",
  "allowedTools": ["server/tool_name_1", "server/tool_name_2"],
  "repository": {
    "url": "https://github.com/stacklok/toolhive-catalog",
    "type": "git"
  },
  "icons": [
    {
      "src": "icon.svg",
      "type": "image/svg+xml"
    }
  ],
  "packages": [
    {
      "registryType": "git",
      "url": "https://github.com/stacklok/toolhive-catalog",
      "ref": "main",
      "subfolder": "registries/toolhive/skills/my-skill/skill"
    }
  ]
}
```

### Step 4: Add an Icon

Add an `icon.svg` file to your skill directory.

### Step 5: Validate

```bash
task catalog:validate
task catalog:build
```

For a complete example, see `registries/toolhive/skills/code-review/`.

## Common Questions

### What is "transport"?

This tells ToolHive how to communicate with your server:

For **container-based servers** (in `packages[].transport.type`):

- `stdio` - Standard input/output (most common)
- `sse` - Server-sent events
- `streamable-http` - HTTP streaming

For **remote servers** (in `remotes[].type`):

- `streamable-http` - HTTP streaming (recommended)
- `sse` - Server-sent events (deprecated but still supported)
- **Note:** Remote servers cannot use `stdio`

If you're not sure, use `stdio` for containers and `streamable-http` for remote servers.

### What is "tier"?

- `Official` - Maintained by the MCP team or platform owners
- `Community` - Created and maintained by the community (most servers)

### What is "status"?

- `Active` - Fully functional and maintained
- `Deprecated` - No longer maintained, will be removed

### What about the "tools" field?

List all the tools your server provides in the `_meta` extensions block. If your server's tools aren't knowable until runtime, you can use `["set_during_runtime"]`.

### Do I need a Docker image?

**For container-based servers:** Yes. Your MCP server must be packaged as a Docker image and published to a registry like:

- Docker Hub (`docker.io/username/image`)
- GitHub Container Registry (`ghcr.io/username/image`)
- Other public registries

**For remote servers:** No. You just need to provide the URL endpoint where your MCP server is accessible.

### Security Note: Filesystem Permissions

**Important:** Don't specify filesystem paths or volume mounts in your registry entries. Mounting host directories is a security risk and should be configured by users at runtime, not in registry specs. Only network permissions (allowed hosts and ports) should be specified in the `permissions` section.

### How do I test my entry?

After adding your entry, you can validate it:

```bash
# Validate all entries
task catalog:validate

# Build the registry files
task catalog:build
```

Or submit a pull request and our automated checks will validate it for you.

## Submitting Your Entry

1. Fork this repository
2. Add your server entry as described above
3. Submit a pull request
4. We'll review and merge your addition

### Registry Inclusion Criteria

We evaluate submissions based on several criteria to ensure quality and usefulness for the community. Your server should:

- Provide clear value and unique capabilities
- Be well-documented with accurate tool descriptions
- Follow security best practices
- Be actively maintained with recent updates

For detailed evaluation criteria, see the [Registry Inclusion Criteria](docs/registry-criteria.md),
including [MCP server criteria](docs/server-criteria.md) and
[skill criteria](docs/skill-criteria.md).

### Before Submitting, Please Ensure

For **container-based servers**:

- [ ] Your Docker image is publicly accessible
- [ ] The `description` clearly explains what your server does
- [ ] You've listed all the tools your server provides
- [ ] Any required environment variables are documented
- [ ] Your server works with ToolHive

For **remote servers**:

- [ ] Your server endpoint is publicly accessible
- [ ] The `description` clearly explains what your server does
- [ ] You've listed all the tools your server provides
- [ ] Any required authentication is documented
- [ ] The transport is set to `streamable-http` (preferred) or `sse` (not `stdio`)
- [ ] Your server works with ToolHive's proxy command

## Need Help?

- Check existing entries in `registries/toolhive/servers/` for server examples
- Check existing entries in `registries/toolhive/skills/` for skill examples
- Open an issue if you have questions
- Join our community discussions

## For Maintainers

If you need to work with the registry programmatically:

```bash
# Validate all entries
task catalog:validate

# Build the upstream MCP registry files
task catalog:build

# Update metadata for oldest entries
task catalog:update-metadata:oldest

# Preview a sync of every dockyard-sourced skill (no writes)
task catalog:sync-skills:all:dry-run

# See all available commands
task
```

### Dockyard skill sync

Skills whose first OCI package targets `ghcr.io/stacklok/dockyard/skills/<name>:*`
are kept in lockstep with the `skills/<name>/spec.yaml` files in
[stacklok/dockyard](https://github.com/stacklok/dockyard). The
[`sync-skills` workflow](.github/workflows/sync-skills.yml) runs daily, executes
`catalog sync-skills --all`, and opens (or refreshes) a single
`catalog-sync-skills` PR that bumps both the OCI tag (`packages[].identifier`)
and the upstream git ref (`packages[].ref`) so they always match what dockyard
last published. To run it manually, use `task catalog:sync-skills:all` or
trigger the workflow via `workflow_dispatch`.

## License

Apache License 2.0
