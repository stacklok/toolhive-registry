# browserbase

MCP server for cloud browser automation with Browserbase and Stagehand

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/browserbase-mcp-server:2.0.0`
- **Repository:** [https://github.com/browserbase/mcp-server-browserbase](https://github.com/browserbase/mcp-server-browserbase)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 16 tools:

- `createSession` | - `listSessions` | - `closeSession`
- `navigateWithSession` | - `actWithSession` | - `extractWithSession`
- `observeWithSession` | - `getUrlWithSession` | - `getAllUrlsWithSession`
- `closeAllSessions` | - `navigate` | - `act`
- `extract` | - `observe` | - `screenshot`
- `getUrl`

## Environment Variables

### Required

- **BROWSERBASE_API_KEY** 🔒: Browserbase API key
- **BROWSERBASE_PROJECT_ID**: Browserbase project ID
- **GEMINI_API_KEY** 🔒: Google Gemini API key for Stagehand

## Tags

`browser` `automation` `web-scraping` `testing` `stagehand` 

## Statistics

- ⭐ Stars: 2411
- 📦 Pulls: 133
- 🕐 Last Updated: 2025-08-13T08:42:35Z
