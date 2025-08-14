# azure

The Azure MCP Server, bringing the power of Azure to your agents.

## Basic Information

- **Image:** `mcr.microsoft.com/azure-sdk/azure-mcp:0.5.1`
- **Repository:** [https://github.com/Azure/azure-mcp](https://github.com/Azure/azure-mcp)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 28 tools:

- `documentation` | - `bestpractices` | - `group`
- `subscription` | - `aks` | - `appconfig`
- `role` | - `datadog` | - `cosmos`
- `foundry` | - `grafana` | - `keyvault`
- `kusto` | - `marketplace` | - `monitor`
- `postgres` | - `redis` | - `search`
- `servicebus` | - `sql` | - `storage`
- `workbooks` | - `bicepschema` | - `azureterraformbestpractices`
- `loadtesting` | - `extension_az` | - `extension_azd`
- `extension_azqr`

## Environment Variables

### Required

- **AZURE_TENANT_ID** 🔒: Your Azure tenant ID
- **AZURE_CLIENT_ID** 🔒: Your Azure client ID for authentication
- **AZURE_CLIENT_SECRET** 🔒: Your Azure client secret for authentication

### Optional

- **HTTP_PROXY**: HTTP proxy URL for outbound requests (optional)
- **HTTPS_PROXY**: HTTPS proxy URL for outbound requests (optional)
- **NO_PROXY**: Comma-separated list of hosts to exclude from proxying (optional)

## Tags

`azure` `microsoft` `cloud` `iaas` `paas` `infrastructure` `database` `storage` 

## Statistics

- ⭐ Stars: 1080
- 📦 Pulls: 1809
- 🕐 Last Updated: 2025-08-13T08:42:38Z
