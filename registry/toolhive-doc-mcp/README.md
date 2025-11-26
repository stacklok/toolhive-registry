# ToolHive Documentation MCP Server

An MCP server that provides semantic search capabilities over Stacklok documentation using vector embeddings. This server enables AI assistants to find relevant documentation content through natural language queries.

## Features

- **Semantic Search**: Vector-based similarity search using local embeddings
- **Multiple Documentation Sources**: Supports both website crawling and GitHub repository documentation
- **Incremental Sync**: Efficient caching to avoid re-fetching unchanged pages
- **Local Embeddings**: Uses fastembed (BAAI/bge-small-en-v1.5) - no API keys required for embeddings
- **GitHub Authentication**: Optional token support for higher API rate limits

## Tools

### `query_docs`

Search documentation using semantic similarity.

**Parameters:**
- `query` (string, required): The search query
- `limit` (integer, optional): Maximum number of results to return (default: 5)
- `query_type` (string, optional): Type of search - "semantic", "keyword", or "hybrid" (default: "semantic")
- `min_score` (number, optional): Minimum relevance score (0.0-1.0)

**Returns:** List of relevant documentation chunks with content, metadata, and relevance scores.

### `get_chunk`

Retrieve full details of a specific documentation chunk by its ID.

**Parameters:**
- `chunk_id` (string, required): UUID of the documentation chunk

**Returns:** Full chunk details including content, source, and metadata.

## Configuration

### Environment Variables

#### Optional Configuration

- `GITHUB_TOKEN`: GitHub personal access token for higher API rate limits
  - Without token: 60 requests/hour
  - With token: 5,000 requests/hour
  - No special scopes needed for public repositories
  - Create at: https://github.com/settings/tokens

#### Telemetry (OpenTelemetry)

- `OTEL_ENABLED`: Enable/disable OpenTelemetry logging (default: `true`)
- `OTEL_ENDPOINT`: OpenTelemetry collector endpoint (default: `http://otel-collector.otel.svc.cluster.local:4318`)
- `OTEL_SERVICE_NAME`: Service name for telemetry (default: `toolhive-doc-mcp`)
- `OTEL_SERVICE_VERSION`: Service version for telemetry (default: `1.0.0`)

## Usage Example

Using ToolHive CLI to query the documentation:

```bash
# Start the MCP server via ToolHive
thv proxy toolhive-doc-mcp --port 9090

# In another terminal, query the docs
curl -X POST http://localhost:9090/sse \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "query_docs",
      "arguments": {
        "query": "What is toolhive?",
        "limit": 5
      }
    }
  }'
```

## Architecture

- **Vector Store**: SQLite with sqlite-vec for vector similarity search
- **Embeddings**: Local fastembed model (BAAI/bge-small-en-v1.5)
- **Content Extraction**: BeautifulSoup4 + lxml for HTML parsing
- **GitHub Integration**: GitHub API client with authentication support
- **MCP Server**: FastMCP with HTTP/SSE protocol
- **Telemetry**: OpenTelemetry logging for query and response data

## Documentation Sources

The server includes pre-built documentation from:
- Stacklok ToolHive documentation
- GitHub repositories with markdown files
- Additional sources configurable at build time

## Network Permissions

This server requires network access to:
- `*.github.com` and `*.githubusercontent.com` on port 443 for GitHub API access during runtime operations

## Links

- **Repository**: https://github.com/stacklok/toolhive-doc-mcp
- **Documentation**: See the repository README for detailed configuration options
- **Image**: ghcr.io/stackloklabs/toolhive-doc-mcp

## Support

For issues or questions:
- Open an issue at: https://github.com/stacklok/toolhive-doc-mcp/issues
- View documentation at: https://github.com/stacklok/toolhive-doc-mcp#readme

