# elasticsearch

Connect to your Elasticsearch data.

## Basic Information

- **Image:** `docker.io/mcp/elasticsearch:latest`
- **Repository:** [https://github.com/elastic/mcp-server-elasticsearch](https://github.com/elastic/mcp-server-elasticsearch)
- **Tier:** Official
- **Status:** Active
- **Transport:** streamable-http

## Available Tools

This server provides 5 tools:

- `esql`
- `get_mappings`
- `get_shards`
- `list_indices`
- `search`

## Environment Variables

### Required

- **ES_URL**: Your Elasticsearch instance URL

### Optional

- **ES_API_KEY** 🔒: Elasticsearch API key for authentication
- **ES_USERNAME**: Elasticsearch username for basic authentication
- **ES_PASSWORD** 🔒: Elasticsearch password for basic authentication
- **ES_CA_CERT**: Path to custom CA certificate for Elasticsearch SSL/TLS
- **ES_SSL_SKIP_VERIFY**: Set to '1' or 'true' to skip SSL certificate verification
- **ES_PATH_PREFIX**: Path prefix for Elasticsearch instance exposed at a non-root path
- **ES_VERSION**: Server assumes Elasticsearch 9.x. Set to 8 target Elasticsearch 8.x

## Tags

`elasticsearch` `search` `analytics` `data` `alerting` `observability` `metrics` `logs` 

## Statistics

- ⭐ Stars: 402
- 📦 Pulls: 10632
- 🕐 Last Updated: 2025-08-11T00:24:57Z
