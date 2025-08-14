#!/usr/bin/env python3
import json
import sys

# List of uncommitted servers
uncommitted_servers = [
    "arxiv-mcp-server",
    "brightdata-mcp",
    "ida-pro-mcp",
    "magic-mcp",
    "mcp-jetbrains",
    "mcp-neo4j-aura-manager",
    "mcp-neo4j-cypher",
    "mcp-neo4j-memory",
    "mcp-server-neon",
    "tavily-mcp"
]

# Read the full registry
with open('build/registry.json', 'r') as f:
    registry = json.load(f)

# Create a new registry with only uncommitted servers
uncommitted_registry = {
    "$schema": registry.get("$schema", ""),
    "version": registry.get("version", ""),
    "last_updated": registry.get("last_updated", ""),
    "servers": {}
}

# Extract only the uncommitted servers
for server_name in uncommitted_servers:
    if server_name in registry["servers"]:
        uncommitted_registry["servers"][server_name] = registry["servers"][server_name]
    else:
        print(f"Warning: Server '{server_name}' not found in registry.json", file=sys.stderr)

# Write the result to a new file
output_file = 'build/uncommitted_registry.json'
with open(output_file, 'w') as f:
    json.dump(uncommitted_registry, f, indent=2)

print(f"Extracted {len(uncommitted_registry['servers'])} uncommitted servers to {output_file}")
print(f"Servers included: {', '.join(uncommitted_registry['servers'].keys())}")
