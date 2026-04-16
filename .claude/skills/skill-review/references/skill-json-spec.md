# Skill Field Reference

Source: [docs/adding-skills-llm.md](../../../../docs/adding-skills-llm.md)

## Directory Layout

```
registries/toolhive/skills/<skill-name>/
├── skill.json          # Registry metadata (namespace, name, packages)
├── icon.svg            # Visual icon
└── skill/              # Installable content (only this gets installed)
    ├── SKILL.md        # Skill prompt with YAML frontmatter
    ├── scripts/        # Optional: executable code
    ├── references/     # Optional: documentation loaded on demand
    └── assets/         # Optional: templates, images, data files
```

`packages[].subfolder` in skill.json points to the `skill/` directory so only
skill content is installed, not registry metadata.

## skill.json Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | string | Yes | Reverse-DNS namespace (e.g., `io.github.stacklok`). 3-128 chars. |
| `name` | string | Yes | Skill identifier. 1-64 chars, pattern `^[a-z0-9][a-z0-9-]*[a-z0-9]$`. |
| `description` | string | Yes | One-line description. 1-1024 chars. |
| `version` | string | Yes | Semantic version (e.g., `0.1.0`). |
| `status` | string | No | `active`, `deprecated`, or `archived`. Default: `active`. |
| `title` | string | No | Human-readable display name. Max 100 chars. |
| `license` | string | No | SPDX license identifier. Max 256 chars. |
| `compatibility` | string | No | Environment requirements. Max 500 chars. |
| `allowedTools` | string[] | No | Tools the skill uses (format: `server/tool_name`). Experimental. |
| `repository` | object | No | Source repository (`url`, `type`). |
| `icons` | array | No | Icon references (`src`, `type`). |
| `packages` | array | Yes | Distribution packages. |
| `metadata` | map | No | Official metadata from SKILL.md (auto-populated by CI). |
| `_meta` | map | No | Extended metadata. |

## SKILL.md Frontmatter Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Skill identifier (must match skill.json `name`). |
| `description` | string | Yes | One-line description. |
| `version` | string | No | Semantic version. |
| `allowed-tools` | string[] | No | Tools the skill uses. |
| `license` | string | No | SPDX license identifier. |
| `compatibility` | string | No | Environment requirements. |
| `metadata` | map | No | Additional key-value metadata. |

## Package Types

### Git (catalog-hosted skills)

```json
{
  "registryType": "git",
  "url": "https://github.com/stacklok/toolhive-catalog",
  "ref": "main",
  "subfolder": "registries/toolhive/skills/<skill-name>/skill"
}
```

### OCI (independently published skills)

```json
{
  "registryType": "oci",
  "identifier": "ghcr.io/org/skill-name:v0.1.0"
}
```

## Validation Rules

| Rule | Description |
|------|-------------|
| Skill name | Lowercase letters, numbers, hyphens only (`^[a-z0-9][a-z0-9-]*[a-z0-9]$`) |
| `name` match | skill.json `name` must match SKILL.md frontmatter `name` |
| Required files | `skill.json`, `icon.svg`, and `skill/SKILL.md` must all exist |
| SKILL.md format | Must start with YAML frontmatter (`---` delimiters) |
| Frontmatter | `name` field is required in frontmatter |
| JSON syntax | Valid JSON, 2-space indentation |
| Version | Semantic version format (e.g., `0.1.0`) |
| Subfolder | `packages[].subfolder` must end with `/skill` |

## Validation Commands

```bash
task catalog:validate                                                         # Schema validation
task catalog:build                                                            # Build registry
jq '.data.skills[] | select(.name == "<name>")' build/toolhive/registry-upstream.json  # Check entry
```
