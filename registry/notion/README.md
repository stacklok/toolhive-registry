# notion

Provides integration with Notion.

## Basic Information

- **Image:** `docker.io/mcp/notion:latest`
- **Repository:** [https://github.com/makenotion/notion-mcp-server](https://github.com/makenotion/notion-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 19 tools:

- `API-get-user` | - `API-get-users` | - `API-get-self`
- `API-post-database-query` | - `API-post-search` | - `API-get-block-children`
- `API-patch-block-children` | - `API-retrieve-a-block` | - `API-update-a-block`
- `API-delete-a-block` | - `API-retrieve-a-page` | - `API-patch-page`
- `API-post-page` | - `API-create-a-database` | - `API-update-a-database`
- `API-retrieve-a-database` | - `API-retrieve-a-page-property` | - `API-retrieve-a-comment`
- `API-create-a-comment`

## Environment Variables

### Required

- **OPENAPI_MCP_HEADERS** 🔒: HTTP headers for Notion API requests in JSON format. Example: {"Authorization":"Bearer ntn_****","Notion-Version":"2022-06-28"}

## Tags

`notion` `notes` 

## Statistics

- ⭐ Stars: 2923
- 📦 Pulls: 5968
- 🕐 Last Updated: 2025-08-11T00:24:58Z
