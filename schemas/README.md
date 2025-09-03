# ToolHive Registry Schema

This directory contains the JSON Schema for the ToolHive registry format.

## Schema Overview

The schema defines the structure for ToolHive registry JSON files, leveraging the official MCP server schema for individual server entries.

**Schema Location:** `schemas/registry.schema.json`

**Schema URL:** `https://raw.githubusercontent.com/stacklok/toolhive-registry/main/schemas/registry.schema.json`

## Schema Design

### Upstream Integration

The schema reuses the official MCP server schema:
`https://static.modelcontextprotocol.io/schemas/2025-07-09/server.schema.json`

**Benefits:**
- **Standards compliance** with MCP specification
- **Automatic compatibility** with MCP ecosystem tooling  
- **Reduced maintenance** - leverages upstream schema evolution
- **Ecosystem integration** - works with existing MCP validators

### Registry Structure

```json
{
  "$schema": "https://raw.githubusercontent.com/stacklok/toolhive-registry/main/schemas/registry.schema.json",
  "version": "1.0.0",
  "meta": {
    "last_updated": "2024-01-15T10:30:00Z"
  },
  "data": {
    "servers": [...],  // Uses official MCP server schema
    "groups": []       // Placeholder for future grouping
  }
}
```

## Usage

### In Generated Registries

The schema URL is automatically included in all generated registry files:

```go
// Generated registries automatically include the schema reference
registry := NewOfficialRegistry(loader)
registry.WriteJSON("output/registry.json")
```

### With Validation Tools

JSON Schema validators can verify registry format:

```bash
# Using ajv-cli
npx ajv validate -s schemas/registry.schema.json -d output/registry.json

# Using jsonschema (Python)
jsonschema -i output/registry.json schemas/registry.schema.json
```

### IDE Support

IDEs with JSON Schema support will provide:
- Auto-completion for registry structure
- Validation errors and warnings
- Documentation on hover

## Maintenance

### Schema Updates

For schema changes:

1. **Update schema file**: Edit `schemas/registry.schema.json`
2. **Update ID if needed**: Update `$id` field if URL changes  
3. **Test changes**: Validate existing registries against updated schema
4. **Commit schema**: Push changes to make schema available via GitHub

### Breaking Changes

For major structural changes:
- Consider backwards compatibility
- Update registry generation code accordingly
- Document breaking changes in release notes

## References

- [JSON Schema Specification](https://json-schema.org/)
- [MCP Server Schema](https://static.modelcontextprotocol.io/schemas/2025-07-09/server.schema.json)
- [Model Context Protocol](https://modelcontextprotocol.io/)