# atlassian

Connect to Atlassian products like Confluence, Jira Cloud and Server/Data deployments.

## Basic Information

- **Image:** `ghcr.io/sooperset/mcp-atlassian:0.11.9`
- **Repository:** [https://github.com/sooperset/mcp-atlassian](https://github.com/sooperset/mcp-atlassian)
- **Tier:** Community
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 41 tools:

- `confluence_search` | - `confluence_get_page` | - `confluence_get_page_children`
- `confluence_get_comments` | - `confluence_get_labels` | - `confluence_add_label`
- `confluence_create_page` | - `confluence_update_page` | - `confluence_delete_page`
- `confluence_add_comment` | - `confluence_search_user` | - `jira_get_user_profile`
- `jira_get_issue` | - `jira_search` | - `jira_search_fields`
- `jira_get_project_issues` | - `jira_get_transitions` | - `jira_get_worklog`
- `jira_download_attachments` | - `jira_get_agile_boards` | - `jira_get_board_issues`
- `jira_get_sprints_from_board` | - `jira_get_sprint_issues` | - `jira_get_link_types`
- `jira_create_issue` | - `jira_batch_create_issues` | - `jira_batch_get_changelogs`
- `jira_update_issue` | - `jira_delete_issue` | - `jira_add_comment`
- `jira_add_worklog` | - `jira_link_to_epic` | - `jira_create_issue_link`
- `jira_remove_issue_link` | - `jira_transition_issue` | - `jira_create_sprint`
- `jira_update_sprint` | - `jira_get_project_versions` | - `jira_get_all_projects`
- `jira_create_version` | - `jira_batch_create_versions`

## Environment Variables


### Optional

- **CONFLUENCE_URL**: Confluence URL (e.g., https://your-domain.atlassian.net/wiki)
- **CONFLUENCE_USERNAME**: Confluence username/email for Cloud deployments
- **CONFLUENCE_API_TOKEN** 🔒: Confluence API token for Cloud deployments
- **CONFLUENCE_PERSONAL_TOKEN** 🔒: Confluence Personal Access Token for Server/Data Center deployments
- **CONFLUENCE_SSL_VERIFY**: Verify SSL certificates for Confluence Server/Data Center (true/false)
- **CONFLUENCE_SPACES_FILTER**: Comma-separated list of Confluence space keys to filter search results
- **JIRA_URL**: Jira URL (e.g., https://your-domain.atlassian.net)
- **JIRA_USERNAME**: Jira username/email for Cloud deployments
- **JIRA_API_TOKEN** 🔒: Jira API token for Cloud deployments
- **JIRA_PERSONAL_TOKEN** 🔒: Jira Personal Access Token for Server/Data Center deployments
- **JIRA_SSL_VERIFY**: Verify SSL certificates for Jira Server/Data Center (true/false)
- **JIRA_PROJECTS_FILTER**: Comma-separated list of Jira project keys to filter search results
- **READ_ONLY_MODE**: Run in read-only mode (disables all write operations)
- **MCP_VERBOSE**: Increase logging verbosity
- **ENABLED_TOOLS**: Comma-separated list of tool names to enable (if not set, all tools are enabled)

## Tags

`atlassian` `confluence` `jira` `wiki` `issue-tracking` `project-management` `documentation` `cloud` `server` `data-center` 

## Statistics

- ⭐ Stars: 2741
- 📦 Pulls: 12519
- 🕐 Last Updated: 2025-08-11T00:24:54Z
