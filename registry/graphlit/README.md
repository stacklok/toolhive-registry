# graphlit

MCP server for Graphlit platform - ingest, search, and retrieve knowledge from multiple sources

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/graphlit-mcp-server:1.0.20250808001`
- **Repository:** [https://github.com/graphlit/graphlit-mcp-server](https://github.com/graphlit/graphlit-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 47 tools:

- `query_contents` | - `query_collections` | - `query_feeds`
- `query_conversations` | - `retrieve_relevant_sources` | - `retrieve_similar_images`
- `visually_describe_image` | - `prompt_llm_conversation` | - `extract_structured_json_from_text`
- `publish_as_audio` | - `publish_as_image` | - `ingest_files`
- `ingest_web_pages` | - `ingest_messages` | - `ingest_posts`
- `ingest_emails` | - `ingest_issues` | - `ingest_text`
- `ingest_memory` | - `web_crawling` | - `web_search`
- `web_mapping` | - `screenshot_page` | - `configure_project`
- `create_collection` | - `add_contents_to_collection` | - `remove_contents_from_collection`
- `delete_collections` | - `delete_feeds` | - `delete_contents`
- `delete_conversations` | - `is_feed_done` | - `is_content_done`
- `list_slack_channels` | - `list_microsoft_teams_teams` | - `list_microsoft_teams_channels`
- `list_sharepoint_libraries` | - `list_sharepoint_folders` | - `list_linear_projects`
- `list_notion_databases` | - `list_notion_pages` | - `list_dropbox_folders`
- `list_box_folders` | - `list_discord_guilds` | - `list_discord_channels`
- `list_google_calendars` | - `list_microsoft_calendars`

## Environment Variables

### Required

- **GRAPHLIT_ENVIRONMENT_ID**: Your Graphlit environment ID
- **GRAPHLIT_ORGANIZATION_ID**: Your Graphlit organization ID
- **GRAPHLIT_JWT_SECRET** 🔒: Your JWT secret for signing the JWT token

### Optional

- **SLACK_BOT_TOKEN** 🔒: Slack bot token for Slack integration
- **DISCORD_BOT_TOKEN** 🔒: Discord bot token for Discord integration
- **TWITTER_TOKEN** 🔒: Twitter/X API token
- **GOOGLE_EMAIL_REFRESH_TOKEN** 🔒: Google refresh token for Gmail integration
- **GOOGLE_EMAIL_CLIENT_ID**: Google client ID for Gmail integration
- **GOOGLE_EMAIL_CLIENT_SECRET** 🔒: Google client secret for Gmail integration
- **LINEAR_API_KEY** 🔒: Linear API key for Linear integration
- **GITHUB_PERSONAL_ACCESS_TOKEN** 🔒: GitHub personal access token
- **JIRA_EMAIL**: Jira email for authentication
- **JIRA_TOKEN** 🔒: Jira API token
- **NOTION_API_KEY** 🔒: Notion API key for Notion integration

## Tags

`knowledge-base` `rag` `search` `ingestion` `data-connectors` 

## Statistics

- ⭐ Stars: 347
- 📦 Pulls: 109
- 🕐 Last Updated: 2025-08-13T08:42:34Z
