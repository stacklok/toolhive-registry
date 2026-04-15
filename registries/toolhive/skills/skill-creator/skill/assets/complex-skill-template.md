# Complex Skill Template

A complex skill includes SKILL.md plus bundled resources (scripts, references, and/or assets).

## Phases for Complex Skills

1. **Gather** - Clarify requirements with user
2. **Research** - Explore codebase, gather context
3. **Plan** - Decompose into tasks, propose approach
4. **Execute** - Generate output, allow user feedback

## Directory Structure

```
skill-name/
├── skill.json         # Registry metadata (not installed)
├── icon.svg           # Registry icon (not installed)
└── skill/             # Installable content
    ├── SKILL.md
    ├── scripts/       # Executable code for deterministic tasks
    ├── references/    # Documentation loaded into context as needed
    └── assets/        # Templates and files used in output
```

## Choosing Resource Types

| Resource | Use When | Examples |
|----------|----------|----------|
| **Instructions** (SKILL.md) | Agent knows how, needs guidance | Workflows, checklists, guidelines |
| **References** | Domain knowledge needed on-demand | Schemas, API docs, policies |
| **Assets/Templates** | Boilerplate to copy and customize | Code templates, config files, scaffolds |
| **Scripts** | Binary manipulation, deterministic ops | PDF/image processing |

**Prefer templates over scripts** when the agent can adapt boilerplate code.

## SKILL.md Template

```markdown
---
name: skill-name
description: Comprehensive description. Include all triggers and contexts.
---

# Skill Name

Brief overview.

## Quick Start

Most common workflow in minimal steps.

## Workflows

### Workflow Name

1. Step one
2. Step two
3. Step three

## Resources

- **Templates**: Copy from `assets/` and customize
- **References**: See `references/` for domain knowledge

## Guidelines

- Key constraints
- Quality standards
```

## Complete Example: Data Pipeline Skill

```
data-pipeline/
├── skill.json
├── icon.svg
└── skill/
    ├── SKILL.md
    ├── assets/
    │   └── etl-template.py
    └── references/
        └── sources.md
```

### SKILL.md

```markdown
---
name: data-pipeline
description: Build data pipelines for ETL tasks. Use when users need to extract, transform, and load data between sources.
---

# Data Pipeline

Build ETL pipelines using templates.

## Quick Start

1. Copy `assets/etl-template.py`
2. Configure source and destination
3. Add transformation logic
4. Run with `uv run pipeline.py`

## Resources

- **Pipeline template**: `assets/etl-template.py`
- **Data sources**: See [references/sources.md](references/sources.md)
```
