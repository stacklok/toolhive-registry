---
name: skill-creator
description: Guide for creating effective skills. Use when users want to create a new skill, update an existing skill, build a slash command, or extend agent capabilities with specialized knowledge, workflows, tool integrations, or custom commands.
version: "0.1.0"
license: Apache-2.0
metadata:
  author: Stacklok
  homepage: https://github.com/stacklok/toolhive-catalog
---

# Skill Creator

This skill provides guidance for creating effective skills.

## About Skills

Skills are modular, self-contained packages that extend an agent's capabilities by providing
specialized knowledge, workflows, and tools. Think of them as "onboarding guides" for specific
domains or tasks—they transform a general-purpose agent into a specialized agent
equipped with procedural knowledge that no model can fully possess.

### What Skills Provide

1. Specialized workflows - Multi-step procedures for specific domains
2. Tool integrations - Instructions for working with specific file formats or APIs
3. Domain expertise - Company-specific knowledge, schemas, business logic
4. Bundled resources - Scripts, references, and assets for complex and repetitive tasks

## Core Principles

### Concise is Key

The context window is a public good. Skills share the context window with everything else the agent needs: system prompt, conversation history, other Skills' metadata, and the actual user request.

**Default assumption: The agent is already very smart.** Only add context the agent doesn't already have. Challenge each piece of information: "Does the agent really need this explanation?" and "Does this paragraph justify its token cost?"

Prefer concise examples over verbose explanations.

### Set Appropriate Degrees of Freedom

Match the level of specificity to the task's fragility and variability:

**High freedom (text-based instructions)**: Use when multiple approaches are valid, decisions depend on context, or heuristics guide the approach.

**Medium freedom (pseudocode or scripts with parameters)**: Use when a preferred pattern exists, some variation is acceptable, or configuration affects behavior.

**Low freedom (specific scripts, few parameters)**: Use when operations are fragile and error-prone, consistency is critical, or a specific sequence must be followed.

Think of the agent as exploring a path: a narrow bridge with cliffs needs specific guardrails (low freedom), while an open field allows many routes (high freedom).

### When to Use Subagents

Skills can instruct the agent to spawn subagents (via the Task tool) for parallel or specialized work.

**Use subagents when:**
- Parallel independent work (batch processing multiple files)
- Codebase exploration requiring multiple search iterations
- Isolated complex subtasks that don't need user interaction
- Long-running background operations

**Avoid subagents when:**
- Frequent user interaction is needed
- Sequential dependencies between steps
- Task is simple enough that overhead isn't justified

See [assets/subagent-template.md](assets/subagent-template.md) for types, prompt structure, and examples.

### Anatomy of a Skill

Every skill consists of a required SKILL.md file and optional bundled resources:

```
skill-name/
├── SKILL.md (required)
│   ├── YAML frontmatter metadata (required)
│   │   ├── name: (required)
│   │   └── description: (required)
│   └── Markdown instructions (required)
└── Bundled Resources (optional)
    ├── scripts/          - Executable code (Python/Bash/etc.)
    ├── references/       - Documentation intended to be loaded into context as needed
    └── assets/           - Files used in output (templates, icons, fonts, etc.)
```

#### SKILL.md (required)

Every SKILL.md consists of:

- **Frontmatter** (YAML): Contains the required `name` and `description` fields. The description is what the agent reads to determine when to use the skill.
- **Body** (Markdown): Instructions and guidance for using the skill. Only loaded AFTER the skill triggers.

#### Bundled Resources (optional)

##### Scripts (`scripts/`)

Executable code (Python/Bash/etc.) for tasks that require deterministic reliability or are repeatedly rewritten.

- **When to include**: When the same code is being rewritten repeatedly or deterministic reliability is needed
- **Example**: `scripts/rotate_pdf.py` for PDF rotation tasks
- **Benefits**: Token efficient, deterministic, may be executed without loading into context
- **Note**: Scripts may still need to be read by the agent for patching or environment-specific adjustments
- **Coding Principles**: Avoid modifying the user's environment to manage packages or other installs. E.g., use uv with inline dependencies (PEP 723) to run python scripts.

```python
# /// script
# requires-python = ">=3.11"
# dependencies = ["pandas", "openpyxl"]
# ///

import pandas as pd
# ... rest of script
```

Run with: `uv run script.py`

##### References (`references/`)

Documentation and reference material intended to be loaded as needed into context to inform the agent's process and thinking.

- **When to include**: For documentation that the agent should reference while working
- **Examples**: `references/finance.md` for financial schemas, `references/api_docs.md` for API specifications
- **Use cases**: Database schemas, API documentation, domain knowledge, company policies, detailed workflow guides
- **Benefits**: Keeps SKILL.md lean, loaded only when the agent determines it's needed
- **Best practice**: If files are large (>10k words), include grep search patterns in SKILL.md
- **Avoid duplication**: Information should live in either SKILL.md or references files, not both. Prefer references files for detailed information unless it's truly core to the skill.

##### Assets (`assets/`)

Files not intended to be loaded into context, but rather used within the output the agent produces.

- **When to include**: When the skill needs files that will be used in the final output
- **Examples**: `assets/logo.png` for brand assets, `assets/frontend-template/` for HTML/React boilerplate
- **Use cases**: Templates, images, icons, boilerplate code, fonts, sample documents that get copied or modified
- **Benefits**: Separates output resources from documentation, enables the agent to use files without loading them into context

#### What to Not Include in a Skill

A skill should only contain essential files that directly support its functionality. Do NOT create extraneous documentation or auxiliary files, including:

- README.md
- INSTALLATION_GUIDE.md
- QUICK_REFERENCE.md
- CHANGELOG.md

### Progressive Disclosure Design Principle

Skills use a three-level loading system to manage context efficiently:

1. **Metadata (name + description)** - Always in context (~100 words)
2. **SKILL.md body** - When skill triggers (<5k words)
3. **Bundled resources** - As needed by the agent (Unlimited because scripts can be executed without reading into context window)

#### Progressive Disclosure Patterns

Keep SKILL.md body to the essentials and under 500 lines to minimize context bloat. Split content into separate files when approaching this limit. When splitting out content into other files, it is very important to reference them from SKILL.md and describe clearly when to read them, to ensure the reader of the skill knows they exist and when to use them.

**Key principle:** When a skill supports multiple variations, frameworks, or options, keep only the core workflow and selection guidance in SKILL.md. Move variant-specific details (patterns, examples, configuration) into separate reference files.

**Pattern 1: High-level guide with references**

```markdown
# PDF Processing

## Quick start

Extract text with pdfplumber:
[code example]

## Advanced features

- **Form filling**: See [FORMS.md](FORMS.md) for complete guide
- **API reference**: See [REFERENCE.md](REFERENCE.md) for all methods
```

**Pattern 2: Domain-specific organization**

```
bigquery-skill/
├── SKILL.md (overview and navigation)
└── reference/
    ├── finance.md (revenue, billing metrics)
    ├── sales.md (opportunities, pipeline)
    └── product.md (API usage, features)
```

**Important guidelines:**

- **Avoid deeply nested references** - Keep references one level deep from SKILL.md.
- **Structure longer reference files** - For files longer than 100 lines, include a table of contents at the top.

## Skill Creation Process

1. Understand the skill with concrete examples
2. Plan reusable skill contents (scripts, references, assets)
3. Initialize the skill (create directory structure)
4. Edit the skill (implement resources and write SKILL.md)
5. Iterate based on real usage

### Step 1: Understanding the Skill with Concrete Examples

Skip this step only when the skill's usage patterns are already clearly understood.

To create an effective skill, clearly understand concrete examples of how the skill will be used. Ask the user questions like:

- "What functionality should the skill support?"
- "Can you give some examples of how this skill would be used?"
- "What would a user say that should trigger this skill?"

Conclude this step when there is a clear sense of the functionality the skill should support.

### Step 2: Planning the Reusable Skill Contents

Analyze each example by:

1. Considering how to execute on the example from scratch
2. Identifying what scripts, references, and assets would be helpful when executing these workflows repeatedly

### Step 3: Initializing the Skill

Use the appropriate template from `assets/`:

- **Simple skills**: See [assets/simple-skill-template.md](assets/simple-skill-template.md)
- **Complex skills**: See [assets/complex-skill-template.md](assets/complex-skill-template.md)
- **Subagent**: See [assets/subagent-template.md](assets/subagent-template.md)

### Step 4: Edit the Skill

Start with the reusable resources identified above: `scripts/`, `references/`, `assets/` files. Then update SKILL.md.

**Writing Guidelines:** Always use imperative/infinitive form.

##### Frontmatter

- `name`: The skill name
- `description`: Include both what the Skill does and specific triggers/contexts for when to use it. Include all "when to use" information here — not in the body.

##### Body

Write instructions for using the skill and its bundled resources.

### Step 5: Iterate

1. Use the skill on real tasks
2. Notice struggles or inefficiencies
3. Identify how SKILL.md or bundled resources should be updated
4. Implement changes and test again
