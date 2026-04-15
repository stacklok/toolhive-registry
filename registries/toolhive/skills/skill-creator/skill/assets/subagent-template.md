# Subagent Template

Guide for using subagents in skills via the Task tool.

## Subagent Types

### Built-in Types

| Type | Purpose |
|------|---------|
| `Explore` | Codebase exploration, file search, code questions |
| `general-purpose` | Complex multi-step autonomous tasks |
| `Bash` | Shell command execution |
| `Plan` | Implementation planning |

## Context Isolation

Subagents do **not** inherit conversation history. Prompts must include:
- Absolute file paths
- All required context
- Expected output format
- Where to save results

## Skill Instruction Template

Use in SKILL.md to instruct when/how to spawn subagents:

```markdown
## [Task Name]

Use a subagent for [scenario].

**Subagent type:** `[type]`

**Prompt template:**
> [Action] [task] from `{input}`. [Output format]. Save to `{output_path}`.

**Workflow:**
1. [Gather inputs]
2. Spawn subagent(s) with prompt above
3. [Handle results]
```

## Examples

### Batch Processing

```markdown
## Batch PDF Processing

For 3+ PDFs, use parallel subagents.

**Subagent type:** `general-purpose`

**Prompt template:**
> Extract text and tables from `{filepath}`. Save to `{output_dir}/{basename}.txt`.

**Workflow:**
1. List PDF files
2. Spawn one subagent per file (parallel Task calls in single message)
3. Report completion status
```

### Background Task

```markdown
## Large File Processing

**Subagent type:** `general-purpose`

**Prompt template:**
> Process `{filepath}`. Save results to `{output_path}`.

**Workflow:**
1. Spawn with `run_in_background: true`
2. Continue other work
3. Check status with `TaskOutput` tool
```

## Best Practices

- **Self-contained prompts**: Include all context—subagents have no conversation history
- **Absolute paths**: No relative references
- **Define output format**: JSON, markdown, or file structure
- **Unique temp files**: Prefix with task ID to avoid conflicts
- **Error handling**: Specify behavior for missing/malformed inputs
