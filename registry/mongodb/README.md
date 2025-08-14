# mongodb

Provides support for interacting with MongoDB Databases and MongoDB Atlas.

## Basic Information

- **Image:** `docker.io/mongodb/mongodb-mcp-server:0.2.0`
- **Repository:** [https://github.com/mongodb-js/mongodb-mcp-server](https://github.com/mongodb-js/mongodb-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 32 tools:

- `atlas-list-orgs` | - `atlas-list-projects` | - `atlas-create-project`
- `atlas-list-clusters` | - `atlas-inspect-cluster` | - `atlas-create-free-cluster`
- `atlas-connect-cluster` | - `atlas-inspect-access-list` | - `atlas-create-access-list`
- `atlas-list-db-users` | - `atlas-create-db-user` | - `atlas-list-alerts`
- `connect` | - `find` | - `aggregate`
- `count` | - `insert-one` | - `insert-many`
- `create-index` | - `update-one` | - `update-many`
- `rename-collection` | - `delete-one` | - `delete-many`
- `drop-collection` | - `drop-database` | - `list-databases`
- `list-collections` | - `collection-indexes` | - `collection-schema`
- `collection-storage-size` | - `db-stats`

## Environment Variables


### Optional

- **MDB_MCP_CONNECTION_STRING** 🔒: MongoDB connection string for direct database connections (optional, if not set, you'll need to call the connect tool before interacting with MongoDB data)
- **MDB_MCP_API_CLIENT_ID** 🔒: Atlas API client ID for authentication (required for running Atlas tools)
- **MDB_MCP_API_CLIENT_SECRET** 🔒: Atlas API client secret for authentication (required for running Atlas tools)
- **MDB_MCP_API_BASE_URL**: Atlas API base URL (default is https://cloud.mongodb.com/)
- **MDB_MCP_SERVER_ADDRESS**: MongoDB server address for direct connections (optional, used for connect tool)
- **MDB_MCP_SERVER_PORT**: MongoDB server port for direct connections (optional, used for connect tool)
- **MDB_MCP_LOG_PATH**: Folder to store logs (inside the container)
- **MDB_MCP_DISABLED_TOOLS**: Comma-separated list of tool names, operation types, and/or categories of tools to disable
- **MDB_MCP_READ_ONLY**: When set to true, only allows read and metadata operation types
- **MDB_MCP_TELEMETRY**: When set to disabled, disables telemetry collection

## Tags

`mongodb` `mongo` `atlas` `database` `data` `query` 

## Statistics

- ⭐ Stars: 556
- 📦 Pulls: 5060
- 🕐 Last Updated: 2025-08-13T08:42:37Z
