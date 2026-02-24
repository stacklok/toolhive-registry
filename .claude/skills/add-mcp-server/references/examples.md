# Real-World Examples

These are patterns from actual entries in the registry.

## Container with stdio Transport (heroku)

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

## Container with streamable-http Transport (sqlite)

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

## Remote Server (semgrep-remote)

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

## Remote Server with OAuth (github-remote)

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
