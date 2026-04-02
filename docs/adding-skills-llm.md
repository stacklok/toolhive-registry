# Instructions for LLMs: Adding Skills to ToolHive Registry

## Context

You are helping to add a skill entry to the ToolHive registry. A skill is a reusable prompt or workflow that leverages MCP server tools. Each skill consists of a `SKILL.md` file (the actual prompt with YAML frontmatter), a `skill.json` registry entry, and an `icon.svg`.

## Quick Reference

### Skill Components

| File         | Purpose                                          | Required |
| ------------ | ------------------------------------------------ | -------- |
| `skill.json` | Registry metadata (namespace, name, packages)    | Yes      |
| `SKILL.md`   | Skill prompt with YAML frontmatter               | Yes      |
| `icon.svg`   | Visual icon for the skill                        | Yes      |

### SKILL.md Frontmatter Fields

| Field           | Type     | Required | Description                                    |
| --------------- | -------- | -------- | ---------------------------------------------- |
| `name`          | string   | Yes      | Skill identifier (lowercase, hyphens)          |
| `description`   | string   | Yes      | One-line description                           |
| `version`       | string   | No       | Semantic version (e.g., `"0.1.0"`)            |
| `allowed-tools` | string[] | No       | Tools the skill uses (e.g., `github/get_pull_request`) |
| `license`       | string   | No       | SPDX license identifier                        |
| `compatibility` | string   | No       | Environment requirements                       |
| `metadata`      | map      | No       | Additional key-value metadata                  |

### skill.json Fields

| Field          | Type     | Required | Description                                    |
| -------------- | -------- | -------- | ---------------------------------------------- |
| `namespace`    | string   | Yes      | Reverse-DNS namespace (e.g., `io.github.stacklok`) |
| `name`         | string   | Yes      | Skill identifier (must match SKILL.md name)    |
| `description`  | string   | Yes      | One-line description                           |
| `version`      | string   | Yes      | Semantic version                               |
| `status`       | string   | No       | `active`, `deprecated`, or `archived`          |
| `title`        | string   | No       | Human-readable display name                    |
| `license`      | string   | No       | SPDX license identifier                        |
| `allowedTools` | string[] | No       | Tools the skill uses                           |
| `repository`   | object   | No       | Source repository (`url`, `type`)              |
| `icons`        | array    | No       | Icon references                                |
| `packages`     | array    | Yes      | Distribution packages (OCI or git)             |

## Step-by-Step Process

### 1. Choose a Skill Name

- Use **only** lowercase letters, numbers, and hyphens
- Make it descriptive and action-oriented
- Examples: `code-review`, `bug-triage`, `api-design`, `test-generator`

### 2. Create the Directory

```bash
mkdir -p registries/toolhive/skills/<skill-name>
```

### 3. Identify Required Tools

Determine which MCP server tools the skill needs. Use the format `server/tool_name`:

- Browse existing servers in `registries/toolhive/servers/` to find available tools
- Check the `tools` field in each server's `_meta` extensions
- List only the tools the skill actually uses

### 4. Create the SKILL.md File

The SKILL.md has two parts: YAML frontmatter and markdown content.

```markdown
---
name: my-skill
description: What this skill does in one sentence.
version: "0.1.0"
allowed-tools:
  - server/tool_name_1
  - server/tool_name_2
license: Apache-2.0
metadata:
  author: Author Name
  homepage: https://example.com
---

# Skill Title

You are an expert at [domain]. When asked to [task], follow this process:

## 1. First Step

- Use `tool_name_1` to gather information.
- Analyze the results.

## 2. Second Step

- Use `tool_name_2` to take action.
- Provide structured output.

## Guidelines

- Be specific and actionable.
- Explain your reasoning.
```

**SKILL.md Best Practices:**

- Start with a clear role definition ("You are an expert...")
- Break the workflow into numbered steps
- Reference specific tools by name so the LLM knows what to call
- Include guidelines for tone, output format, and edge cases
- Keep prompts focused — one skill should do one thing well

### 5. Create the skill.json File

```json
{
  "namespace": "io.github.stacklok",
  "name": "<skill-name>",
  "title": "Human-Readable Title",
  "description": "What this skill does in one sentence.",
  "version": "0.1.0",
  "status": "active",
  "license": "Apache-2.0",
  "allowedTools": [
    "server/tool_name_1",
    "server/tool_name_2"
  ],
  "repository": {
    "url": "https://github.com/stacklok/toolhive-catalog",
    "type": "git"
  },
  "icons": [
    {
      "src": "icon.svg",
      "type": "image/svg+xml"
    }
  ],
  "packages": [
    {
      "registryType": "git",
      "url": "https://github.com/stacklok/toolhive-catalog",
      "ref": "main",
      "subfolder": "registries/toolhive/skills/<skill-name>"
    }
  ]
}
```

### 6. Add an Icon

Create an `icon.svg` file in the skill directory. Keep it simple, recognizable at small sizes, and relevant to the skill's purpose.

### 7. Validate

```bash
task catalog:validate
task catalog:build
```

Verify the skill appears in the output:

```bash
jq '.data.skills[] | select(.name == "<skill-name>")' build/toolhive/registry-upstream.json
```

---

## Package Types

Skills support two distribution methods in the `packages` array:

### Git (recommended for catalog-hosted skills)

```json
{
  "registryType": "git",
  "url": "https://github.com/stacklok/toolhive-catalog",
  "ref": "main",
  "subfolder": "registries/toolhive/skills/my-skill"
}
```

### OCI (for independently published skills)

```json
{
  "registryType": "oci",
  "identifier": "ghcr.io/org/skill-name:v0.1.0"
}
```

You can include both if the skill is available through multiple channels.

---

## Complete Example

See `registries/toolhive/skills/code-review/` for a complete working example that:

- References GitHub server tools via `allowedTools`
- Provides a structured code review workflow in `SKILL.md`
- Uses a git package reference for distribution

---

## Validation Rules

| Rule            | Description                                        |
| --------------- | -------------------------------------------------- |
| Skill name      | Lowercase letters, numbers, hyphens only           |
| `name` match    | `skill.json` name must match `SKILL.md` frontmatter name |
| Required files  | `skill.json`, `SKILL.md`, `icon.svg` must all exist |
| SKILL.md format | Must start with YAML frontmatter (`---` delimiters) |
| Frontmatter     | `name` field is required in frontmatter            |
| JSON syntax     | Valid JSON, 2-space indentation                    |
| Version         | Should be semantic version (e.g., `0.1.0`)         |

---

## Final Checklist

Before submitting:

- [ ] Skill name: lowercase, numbers, hyphens only
- [ ] Directory: `registries/toolhive/skills/<skill-name>/` exists
- [ ] `skill.json`: valid JSON with `namespace`, `name`, `description`, `version`, `packages`
- [ ] `SKILL.md`: starts with `---` YAML frontmatter, `name` field present
- [ ] `icon.svg`: file present in skill directory
- [ ] `allowedTools`: lists the actual tools the skill uses
- [ ] `name` in `skill.json` matches `name` in `SKILL.md` frontmatter
- [ ] Validation: `task catalog:validate` passes
- [ ] Build: `task catalog:build` completes successfully
- [ ] Skill appears in `build/toolhive/registry-upstream.json` under `data.skills`

---

## Decision Tree for LLMs

```
Start
  |
What is the skill's purpose?
  |-- Review/analyze something --> Include read-only tools
  |-- Create/modify something --> Include write tools
  |-- Both --> Include both, with clear workflow steps
  |
Which servers provide the needed tools?
  |-- Check registries/toolhive/servers/*/server.json
  |-- Look at the "tools" field in _meta extensions
  |-- List tools as server/tool_name in allowedTools
  |
Create SKILL.md
  |-- Frontmatter: name, description, version, allowed-tools
  |-- Body: role definition, numbered steps, guidelines
  |-- Reference specific tool names in the instructions
  |
Create skill.json
  |-- namespace: io.github.stacklok
  |-- name: must match SKILL.md frontmatter
  |-- packages: git reference to this repo
  |
Validate with task catalog:validate
```