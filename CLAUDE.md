# toolhive-catalog

Curated registry of MCP (Model Context Protocol) servers and skills for ToolHive.

## Skills

- `/add-mcp-server` — use when adding a new server entry to the catalog
- `/mcp-review` — use when reviewing a PR that adds or updates a server entry

## Key commands

Uses [Taskfile](https://taskfile.dev), not Make:

```sh
task catalog:validate   # validate all server.json files
task catalog:build      # build the catalog output
task test               # run Go tests
task lint               # run linters
```

## Registry structure

Each server lives at `registries/toolhive/servers/<name>/` containing:
- `server.json` — server definition (see `docs/adding-entries-llm.md` for full schema)
- `icon.svg` — server icon

Each skill lives at `registries/toolhive/skills/<name>/` containing:
- `skill.json` — skill definition (uses `Skill` type from toolhive-core)
- `SKILL.md` — skill prompt with YAML frontmatter (name, description, version, allowed-tools)
- `icon.svg` — skill icon

## Non-obvious gotchas

- The `_meta` extension key must exactly match `packages[0].identifier` or `remotes[0].url`
- `metadata.*` and `tool_definitions` are auto-populated by CI — never add them manually
- Commits require a `Signed-off-by` trailer (DCO)

## Go code

CLI entry point: `cmd/catalog/`. Packages: `internal/`. Requires Go 1.25+.
