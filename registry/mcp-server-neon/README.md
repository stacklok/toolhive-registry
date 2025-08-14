# mcp-server-neon

MCP server for interacting with Neon Management API and databases

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/mcp-server-neon:0.6.3`
- **Repository:** [https://github.com/neondatabase-labs/mcp-server-neon](https://github.com/neondatabase-labs/mcp-server-neon)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 21 tools:

- `list_projects` | - `describe_project` | - `create_project`
- `delete_project` | - `create_branch` | - `delete_branch`
- `describe_branch` | - `list_branch_computes` | - `list_organizations`
- `get_connection_string` | - `run_sql` | - `run_sql_transaction`
- `get_database_tables` | - `describe_table_schema` | - `list_slow_queries`
- `prepare_database_migration` | - `complete_database_migration` | - `explain_sql_statement`
- `prepare_query_tuning` | - `complete_query_tuning` | - `provision_neon_auth`

## Environment Variables

### Required

- **NEON_API_KEY** 🔒: API key for Neon database service

## Tags

`database` `postgresql` `api` `management` `sql` `migration` `branching` 

## Statistics

- ⭐ Stars: 405
- 📦 Pulls: 55
- 🕐 Last Updated: 2025-08-13T08:42:33Z
