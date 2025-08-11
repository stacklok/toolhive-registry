# redis

Enables LLMs to interact with Redis key-value databases through a set of standardized tools.

## Basic Information

- **Image:** `docker.io/mcp/redis:latest`
- **Repository:** [https://github.com/redis/mcp-redis](https://github.com/redis/mcp-redis)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 44 tools:

- `dbsize` | - `info` | - `client_list`
- `delete` | - `type` | - `expire`
- `rename` | - `scan_keys` | - `scan_all_keys`
- `get_indexes` | - `get_index_info` | - `get_indexed_keys_number`
- `create_vector_index_hash` | - `vector_search_hash` | - `hset`
- `hget` | - `hdel` | - `hgetall`
- `hexists` | - `set_vector_in_hash` | - `get_vector_from_hash`
- `lpush` | - `rpush` | - `lpop`
- `rpop` | - `lrange` | - `llen`
- `set` | - `get` | - `json_set`
- `json_get` | - `json_del` | - `zadd`
- `zrange` | - `zrem` | - `sadd`
- `srem` | - `smembers` | - `xadd`
- `xrange` | - `xdel` | - `publish`
- `subscribe` | - `unsubscribe`

## Environment Variables

### Required

- **REDIS_HOST**: Redis IP or hostname (default "127.0.0.1")

### Optional

- **REDIS_PORT**: Redis port (default 6379)
- **REDIS_DB**: Redis database number (default 0)
- **REDIS_USERNAME**: Redis username (default "default")
- **REDIS_PWD** 🔒: Redis password (default empty)
- **REDIS_SSL**: Redis TLS connection (True|False, default False)
- **REDIS_CA_PATH**: CA certificate for verifying server
- **REDIS_SSL_KEYFILE**: Client's private key file for client authentication
- **REDIS_SSL_CERTFILE**: Client's certificate file for client authentication
- **REDIS_CERT_REQS**: Whether the client should verify the server's certificate (default "required")
- **REDIS_CA_CERTS**: Path to the trusted CA certificates file
- **REDIS_CLUSTER_MODE**: Enable Redis Cluster mode (True|False, default False)
- **MCP_TRANSPORT**: Use the stdio or sse transport (default stdio)

## Tags

`redis` `database` `key-value` `storage` `cache` `data` 

## Statistics

- ⭐ Stars: 196
- 📦 Pulls: 10111
- 🕐 Last Updated: 2025-08-11T00:24:57Z
