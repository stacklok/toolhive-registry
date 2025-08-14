# mcp-clickhouse

MCP server for ClickHouse that provides tools to execute SQL queries, list databases and tables, and optionally use chDB's embedded OLAP engine

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/uvx/mcp-clickhouse:0.1.11`
- **Repository:** [https://github.com/ClickHouse/mcp-clickhouse](https://github.com/ClickHouse/mcp-clickhouse)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 4 tools:

- `run_select_query`
- `list_databases`
- `list_tables`
- `run_chdb_select_query`

## Environment Variables

### Required

- **CLICKHOUSE_HOST**: The hostname of your ClickHouse server
- **CLICKHOUSE_USER**: The username for authentication
- **CLICKHOUSE_PASSWORD** 🔒: The password for authentication

### Optional

- **CLICKHOUSE_PORT**: The port number of your ClickHouse server
  - Default: `8443`
- **CLICKHOUSE_SECURE**: Enable/disable HTTPS connection
  - Default: `true`
- **CLICKHOUSE_VERIFY**: Enable/disable SSL certificate verification
  - Default: `true`
- **CLICKHOUSE_DATABASE**: Default database to use
- **CHDB_ENABLED**: Enable/disable chDB functionality
  - Default: `false`
- **CHDB_DATA_PATH**: The path to the chDB data directory
  - Default: `:memory:`

## Tags

`database` `clickhouse` `sql` `analytics` `olap` 

## Statistics

- ⭐ Stars: 487
- 📦 Pulls: 81
- 🕐 Last Updated: 2025-08-13T08:42:33Z
