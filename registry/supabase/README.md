# supabase

Connect your Supabase projects to AI assistants for managing tables, fetching config, and querying data

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/supabase-mcp-server:latest`
- **Repository:** [https://github.com/supabase-community/supabase-mcp](https://github.com/supabase-community/supabase-mcp)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 31 tools:

- `list_projects` | - `get_project` | - `create_project`
- `pause_project` | - `restore_project` | - `list_organizations`
- `get_organization` | - `get_cost` | - `confirm_cost`
- `search_docs` | - `list_tables` | - `list_extensions`
- `list_migrations` | - `apply_migration` | - `execute_sql`
- `get_logs` | - `get_advisors` | - `get_project_url`
- `get_anon_key` | - `generate_typescript_types` | - `list_edge_functions`
- `deploy_edge_function` | - `create_branch` | - `list_branches`
- `delete_branch` | - `merge_branch` | - `reset_branch`
- `rebase_branch` | - `list_storage_buckets` | - `get_storage_config`
- `update_storage_config`

## Environment Variables

### Required

- **SUPABASE_ACCESS_TOKEN** 🔒: Personal access token from Supabase dashboard

## Tags

`supabase` `database` `backend` `baas` `postgresql` 

## Statistics

- ⭐ Stars: 1942
- 📦 Pulls: 102
- 🕐 Last Updated: 2025-08-13T08:42:36Z
