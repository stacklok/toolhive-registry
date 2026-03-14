# Simple Skill Template

A simple skill consists of just a SKILL.md file with frontmatter and instructions.

## Directory Structure

```
skill-name/
└── SKILL.md
```

## Template

```markdown
---
name: <skill-name>
description: <What the skill does. When to use it. Trigger keywords.>
---

# <Skill Name>

## Overview

<Brief description of what this skill accomplishes>

## Workflow

1. <Step 1>
2. <Step 2>
3. <Step 3>

## Guidelines

- Key guideline or constraint
- Quality standard to maintain
```

## Example: Code Review Skill

```markdown
---
name: code-review
description: Reviews code for best practices, security patterns, and conventions. Use when users ask for code review, feedback on code quality, or want to check code before committing.
---

# Code Review

Analyze code for quality, security, and maintainability.

## Review Checklist

1. Check for security vulnerabilities (injection, XSS, etc.)
2. Verify error handling is appropriate
3. Assess code readability and naming
4. Look for potential performance issues
5. Check adherence to project conventions

## Output Format

Provide feedback organized by severity:
- **Critical**: Security issues or bugs
- **Important**: Best practice violations
- **Suggestion**: Minor improvements
```
