# slack-mcp-server

MCP server for Slack with channels, DMs, message history, search, and smart pagination

## Basic Information

- **Image:** `ghcr.io/korotovsky/slack-mcp-server:v1.1.28`
- **Repository:** [https://github.com/korotovsky/slack-mcp-server](https://github.com/korotovsky/slack-mcp-server)
- **Tier:** Community
- **Status:** Active
- **Transport:** sse
- **Target Port:** 13080

## Available Tools

This server provides 5 tools:

- `conversations_history` - Fetch message history from channels and DMs
- `conversations_replies` - Get thread replies for messages
- `conversations_add_message` - Post messages to channels (when enabled)
- `conversations_search_messages` - Search messages across the workspace
- `channels_list` - List available channels in the workspace

## Environment Variables

### Authentication Options (use one)

- **SLACK_MCP_XOXP_TOKEN** (recommended): User OAuth token (xoxp-...) for Slack API access. Provides full access to channels and search.
- **SLACK_MCP_XOXB_TOKEN**: Bot token (xoxb-...) for Slack API access. Bot has limited access (invited channels only, no search).
- **SLACK_MCP_XOXC_TOKEN** + **SLACK_MCP_XOXD_TOKEN**: Browser session tokens for session-based authentication.

#### How to get a User OAuth Token (xoxp)

1. Go to [api.slack.com/apps](https://api.slack.com/apps) and click **Create New App** → **From scratch**
2. Name it (e.g., "Slack MCP") and select your workspace
3. In the sidebar, click **OAuth & Permissions**
4. Scroll to **User Token Scopes** and add:
   - `channels:history`, `channels:read`
   - `groups:history`, `groups:read`
   - `im:history`, `im:read`, `im:write`
   - `mpim:history`, `mpim:read`, `mpim:write`
   - `users:read`, `chat:write`, `search:read`
5. Click **Install to Workspace** and authorize
6. Copy the **User OAuth Token** (starts with `xoxp-`)

### Server Configuration

- **SLACK_MCP_HOST**: Host address for the MCP server to bind to. Default: `0.0.0.0` (required for container accessibility).

### Optional Configuration

- **SLACK_MCP_ADD_MESSAGE_TOOL**: Enable message posting. Set to 'true' for all channels, or comma-separated channel IDs to whitelist.
- **SLACK_MCP_LOG_LEVEL**: Log level (debug, info, warn, error, panic, fatal). Default is 'info'.

## Tags

`slack` `messaging` `channels` `search` `history` `communication` `workspace`

## Network Permissions

This server requires outbound network access to:
- `*.slack.com` (port 443)
- `*.slack-edge.com` (port 443)
