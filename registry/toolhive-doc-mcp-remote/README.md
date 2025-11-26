# ToolHive Documentation MCP Server (Remote)

Remote MCP server that provides semantic search capabilities over Stacklok documentation using vector embeddings. This is the hosted version of the toolhive-doc-mcp server, accessible via HTTPS without requiring local container deployment.

## Features

- **No Local Setup Required**: Access the server directly via HTTPS
- **Semantic Search**: Vector-based similarity search using local embeddings
- **Pre-built Documentation Index**: Ready-to-use documentation database
- **Always Up-to-date**: Maintained and updated by Stacklok
- **High Availability**: Deployed on Kubernetes with monitoring and observability

## Tools

### `query_docs`

Search documentation using semantic similarity.

**Parameters:**
- `query` (string, required): The search query
- `limit` (integer, optional): Maximum number of results to return (default: 5)
- `query_type` (string, optional): Type of search - "semantic", "keyword", or "hybrid" (default: "semantic")
- `min_score` (number, optional): Minimum relevance score (0.0-1.0)

**Returns:** List of relevant documentation chunks with content, metadata, and relevance scores.

**Example:**
```json
{
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
}
```

### `get_chunk`

Retrieve full details of a specific documentation chunk by its ID.

**Parameters:**
- `chunk_id` (string, required): UUID of the documentation chunk

**Returns:** Full chunk details including content, source, and metadata.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_chunk",
    "arguments": {
      "chunk_id": "12345678-1234-1234-1234-123456789abc"
    }
  }
}
```

## Usage

### Direct Access

The server is available at: `https://toolhive-doc-mcp.stacklok.dev`

You can connect to it directly using any MCP client that supports the `streamable-http` transport.

### Using ToolHive CLI

```bash
# The remote server will be automatically available through ToolHive
thv ls

# Use it via the proxy command
thv proxy toolhive-doc-mcp-remote --port 9090
```

### Testing with curl

```bash
curl -X POST https://toolhive-doc-mcp.stacklok.dev/sse \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "query_docs",
      "arguments": {
        "query": "How do I create an MCP server?",
        "limit": 3
      }
    }
  }'
```

## Documentation Sources

The server includes pre-built, regularly updated documentation from:
- Stacklok ToolHive documentation
- ToolHive GitHub repository
- Related Stacklok projects and resources

## Advantages of Remote Version

1. **No Container Required**: Access immediately without Docker
2. **Always Updated**: Documentation index is refreshed automatically
3. **High Performance**: Hosted on optimized infrastructure
4. **No Resource Usage**: Doesn't consume local CPU/memory
5. **Managed Service**: Maintained and monitored by Stacklok

## Container Version Alternative

If you prefer to run the server locally or need offline access, use the container-based version: [`toolhive-doc-mcp`](../toolhive-doc-mcp)

## Architecture

- **Deployment**: Kubernetes on AWS EKS
- **Vector Store**: SQLite with sqlite-vec
- **Embeddings**: BAAI/bge-small-en-v1.5
- **Protocol**: MCP over HTTP with Server-Sent Events (SSE)
- **Monitoring**: OpenTelemetry with Prometheus and Grafana

## Links

- **Endpoint**: https://toolhive-doc-mcp.stacklok.dev
- **Repository**: https://github.com/stacklok/toolhive-doc-mcp
- **Container Version**: [`toolhive-doc-mcp`](../toolhive-doc-mcp)
- **Documentation**: https://github.com/stacklok/toolhive-doc-mcp#readme

## Support

For issues or questions:
- Open an issue at: https://github.com/stacklok/toolhive-doc-mcp/issues
- View documentation at: https://github.com/stacklok/toolhive-doc-mcp#readme

