# sentry

Sentry's MCP service for human-in-the-loop coding agents, optimized for developer workflows and debugging use cases

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/sentry-mcp-server:latest`
- **Repository:** [https://github.com/getsentry/sentry-mcp](https://github.com/getsentry/sentry-mcp)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 19 tools:

- `whoami` | - `find_organizations` | - `find_teams`
- `find_projects` | - `find_releases` | - `get_issue_details`
- `get_trace_details` | - `get_event_attachment` | - `update_issue`
- `search_events` | - `create_team` | - `create_project`
- `update_project` | - `create_dsn` | - `find_dsns`
- `analyze_issue_with_seer` | - `search_docs` | - `get_doc`
- `search_issues`

## Environment Variables

### Required

- **SENTRY_ACCESS_TOKEN** 🔒: Sentry user auth token with necessary scopes

### Optional

- **SENTRY_HOST**: Sentry host URL (e.g., sentry.example.com)
- **OPENAI_API_KEY** 🔒: OpenAI API key for AI-powered search tools (search_events, search_issues)

## Tags

`sentry` `debugging` `monitoring` `error-tracking` `observability` 

## Statistics

- ⭐ Stars: 292
- 📦 Pulls: 127
- 🕐 Last Updated: 2025-08-13T08:42:35Z
