# ToolHive Registry

This repository contains the registry of MCP (Model Context Protocol) servers available for ToolHive. Each server entry provides AI assistants with specialized tools and capabilities.

## What is this?

Think of this as a catalog of tools that AI assistants can use. Each entry in this registry represents a server that provides specific capabilities - like interacting with GitHub, querying databases, or fetching web content.

## How to Add Your MCP Server

Adding your MCP server to the registry is simple! You just need to create a YAML file with some basic information about your server.

### Step 1: Create a Folder

Create a new folder in the `registry/` directory with your server's name (use lowercase and hyphens):

```
registry/
  └── my-awesome-server/
      └── spec.yaml
```

### Step 2: Create Your spec.yaml File

Create a `spec.yaml` file in your folder with this minimum information:

```yaml
# Required fields - you must provide these
image: docker.io/myorg/my-server:latest  # Your Docker image
description: What your server does in one sentence
transport: stdio  # How your server communicates (usually "stdio")

# Recommended fields - helps users understand your server
tools:
  - tool_name_1  # List the tools your server provides
  - tool_name_2
  
repository_url: https://github.com/myorg/my-server  # Where to find your code
```

### Step 3: Add More Details (Optional but Helpful)

You can add more information to help users:

```yaml
# ... required fields above ...

# Tell users what environment variables they need
env_vars:
  - name: API_KEY
    description: Your API key from example.com
    required: true
    secret: true  # Mark sensitive data

  - name: TIMEOUT
    description: Request timeout in seconds
    required: false
    default: "30"

# Help users find your server
tags:
  - api
  - integration
  - productivity

# Server classification
tier: Community  # or "Official" if maintained by the protocol team
status: Active   # or "Beta", "Deprecated"
```

### Real Example

Here's what a complete entry looks like:

```yaml
image: ghcr.io/github/github-mcp-server:v0.10.0
description: Provides integration with GitHub's APIs for repository management
transport: stdio
repository_url: https://github.com/github/github-mcp-server

tools:
  - create_issue
  - create_pull_request
  - get_file_contents
  - search_repositories

env_vars:
  - name: GITHUB_PERSONAL_ACCESS_TOKEN
    description: GitHub personal access token with appropriate permissions
    required: true
    secret: true

tags:
  - github
  - version-control
  - api

tier: Official
status: Active
```

## Common Questions

### What is "transport"?

This tells ToolHive how to communicate with your server. Most servers use:
- `stdio` - Standard input/output (most common)
- `sse` - Server-sent events
- `streamable-http` - HTTP streaming

If you're not sure, use `stdio`.

### What is "tier"?

- `Official` - Maintained by the MCP team or platform owners
- `Community` - Created and maintained by the community (most servers)

### What is "status"?

- `Active` - Fully functional and maintained
- `Beta` - Still in development but usable
- `Alpha` - Early development, may have issues
- `Deprecated` - No longer maintained, will be removed

### Do I need a Docker image?

Yes! Your MCP server must be packaged as a Docker image and published to a registry like:
- Docker Hub (`docker.io/username/image`)
- GitHub Container Registry (`ghcr.io/username/image`)
- Other public registries

### How do I test my entry?

After adding your entry, you can validate it:

```bash
# If you have the build tools installed
task validate
```

Or submit a pull request and our automated checks will validate it for you.

## Submitting Your Entry

1. Fork this repository
2. Add your server entry as described above
3. Submit a pull request
4. We'll review and merge your addition!

### Before Submitting, Please Ensure:

- [ ] Your Docker image is publicly accessible
- [ ] The `description` clearly explains what your server does
- [ ] You've listed all the tools your server provides
- [ ] Any required environment variables are documented
- [ ] Your server actually works with ToolHive

## Need Help?

- Check existing entries in the `registry/` folder for examples
- Open an issue if you have questions
- Join our community discussions

## For Maintainers

If you need to work with the registry programmatically:

```bash
# Import existing registry
task import

# Validate all entries
task validate

# Build the registry.json
task build:registry

# See all available commands
task
```

## License

Apache License 2.0
