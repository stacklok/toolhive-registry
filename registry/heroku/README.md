# heroku

MCP server for seamless interaction between LLMs and the Heroku Platform

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/heroku-mcp-server:1.0.7`
- **Repository:** [https://github.com/heroku/heroku-mcp-server](https://github.com/heroku/heroku-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 32 tools:

- `list_apps` | - `get_app_info` | - `create_app`
- `rename_app` | - `transfer_app` | - `deploy_to_heroku`
- `deploy_one_off_dyno` | - `ps_list` | - `ps_scale`
- `ps_restart` | - `list_addons` | - `get_addon_info`
- `create_addon` | - `maintenance_on` | - `maintenance_off`
- `get_app_logs` | - `pipelines_create` | - `pipelines_promote`
- `pipelines_list` | - `pipelines_info` | - `list_teams`
- `list_private_spaces` | - `pg_psql` | - `pg_info`
- `pg_ps` | - `pg_locks` | - `pg_outliers`
- `pg_credentials` | - `pg_kill` | - `pg_maintenance`
- `pg_backups` | - `pg_upgrade`

## Environment Variables

### Required

- **HEROKU_API_KEY** 🔒: Your Heroku authorization token

### Optional

- **MCP_SERVER_REQUEST_TIMEOUT**: Timeout in milliseconds for command execution
  - Default: `15000`

## Tags

`heroku` `paas` `deployment` `cloud` `devops` 

## Statistics

- ⭐ Stars: 55
- 📦 Pulls: 104
- 🕐 Last Updated: 2025-08-13T08:42:36Z
