---
name: skill-review
description: >-
  Review skill submissions and updates for compliance, security, and quality.
  Use when evaluating skill.json files, SKILL.md content, PRs adding/updating skills,
  or assessing skill changes in the ToolHive registry.
  NOT for reviewing MCP server entries (use mcp-review) or creating new skills (use add-mcp-server).
allowed-tools: Read Grep Glob Bash WebFetch
---

# Skill Submission Review

You are an expert reviewer for the ToolHive Registry. Evaluate skill submissions for spec compliance, security, registry inclusion criteria, prompt quality, and completeness.

For skill.json field specs, see [skill-json-spec.md](references/skill-json-spec.md).
For registry inclusion criteria, see [registry-criteria.md](references/registry-criteria.md) and [skill-criteria.md](references/skill-criteria.md).

## Review Workflow

### Step 1: Identify Change Scope

Determine what you're reviewing:

- **New skill submission** -- Full review (spec + content + repository assessment + inclusion criteria)
- **Version update** -- Focused review (changed fields, prompt diffs, scope changes, skill shadowing)
- **Config change** -- Targeted review (just the changed aspects + security implications)

### Step 2: Validate Directory Structure

Check the skill directory at `registries/toolhive/skills/<name>/`:

1. **Required files** -- `skill.json`, `icon.svg`, and `skill/SKILL.md` must all exist
2. **Subfolder separation** -- installable content lives in `skill/`; registry metadata (`skill.json`, `icon.svg`) at root
3. **No stray files** -- only recognized directories inside `skill/` (scripts/, references/, assets/)

### Step 3: Validate skill.json

Read the skill.json and check:

1. **Required fields** -- `namespace`, `name`, `description`, `version`, `packages` all present
2. **Name format** -- lowercase letters, numbers, hyphens only; matches directory name
3. **Name consistency** -- `name` in skill.json matches `name` in SKILL.md frontmatter
4. **Namespace** -- valid reverse-DNS (e.g., `io.github.stacklok`)
5. **Version** -- semantic versioning format (e.g., `0.1.0`)
6. **Packages** -- at least one entry with valid `registryType` (`oci` or `git`)
7. **Subfolder path** -- `packages[].subfolder` ends with `/skill` (points to skill content, not root)
8. **Icons** -- `icons` array present with `icon.svg` reference
9. **allowedTools** -- if present, lists tools as `server/tool_name` format
10. **No auto-populated fields** -- reject if `metadata` contains CI-populated data in new submissions

Run `task catalog:validate` to catch schema-level issues.

### Step 4: Validate SKILL.md Content

Read `skill/SKILL.md` and evaluate:

1. **Frontmatter** -- starts with `---` YAML delimiters; `name` field present and matches skill.json
2. **Description** -- present, states WHAT the skill does and WHEN to use it
3. **Role definition** -- body starts with clear expertise statement
4. **Workflow structure** -- numbered steps, actionable instructions
5. **Tool references** -- if the skill uses MCP tools, they are referenced by name
6. **No embedded secrets** -- no API keys, tokens, or credentials in prompts or scripts
7. **Length** -- SKILL.md under 500 lines; detailed content split to references/
8. **Quality** -- prompts are clear, focused, and coherent; one skill does one thing well

### Step 5: MCP Server Dependency Check

If the skill declares `allowedTools` or references MCP servers:

1. **Catalog presence** -- every referenced MCP server must already exist in `registries/toolhive/servers/`
2. **Tool existence** -- verify referenced tools appear in the server's `_meta` extensions `tools` list
3. **Scope appropriateness** -- tools requested match the skill's stated purpose

```bash
# Check if a referenced server exists
ls registries/toolhive/servers/<server-name>/server.json
# Check available tools for a server
jq '.. | ._meta? // empty | .tools // empty' registries/toolhive/servers/<server-name>/server.json
```

### Step 6: Security Review

**Must verify:**

- [ ] No secrets, tokens, or credentials embedded in SKILL.md or scripts
- [ ] API keys referenced by the skill are documented as requiring secret handling
- [ ] Scripts (if any) don't contain hardcoded credentials or unsafe operations
- [ ] No known unpatched critical/high CVEs in shipped dependencies
- [ ] Scope is appropriate -- skill doesn't request more tools/permissions than needed

**Recommended checks** (positive signals, not blockers):

- [ ] Provenance configured (Sigstore or GitHub Attestations)
- [ ] Pinned dependencies / Actions pinned to SHAs
- [ ] Automated security scanning in CI
- [ ] Security reporting mechanism (`SECURITY.md`)

### Step 7: Repository Assessment (New Submissions)

For new skills, assess the source repository against registry inclusion criteria.
See [skill-criteria.md](references/skill-criteria.md) for the full checklist.

**Critical checks** (use `gh` CLI, GitHub MCP tools, or WebFetch):

1. **License** -- must be permissive (Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause)
2. **Dependency automation** -- Dependabot or Renovate configured
3. **Security policy** -- check for `SECURITY.md`
4. **CI workflows** -- list `.github/workflows/` contents; confirm CI runs
5. **Recent activity** -- check last 5 commits for recency
6. **Author reputation** -- GitHub account age; established org vs. new account (same-day creation is a red flag)
7. **Releases** -- version tags present; changelog maintained

**Inclusion criteria summary:**

| Category | What to Check |
|----------|---------------|
| Open source | Public repo, permissive license |
| Spec compliance | Agent skill specification compliance, validated by `thv skills validate` |
| MCP dependencies | All referenced servers exist in catalog |
| Distribution | OCI artifact published (Required); git ref acceptable as secondary |
| Versioning | Semver tags, version in skill.json |
| Security | No embedded secrets, auth mechanisms documented, no known CVEs |
| Documentation | README, dependency docs, script explanations |
| Community | Active repo, responsive maintainers, contributor diversity |

### Step 8: Version Update / Skill Shadowing Review

For updates to existing skills:

1. **What changed?** -- diff skill.json and SKILL.md
2. **Name/description consistency** -- same identity but substantially changed behavior? Flag for closer review
3. **Scope creep** -- new MCP server dependencies or tool permissions beyond original scope?
4. **Behavioral drift** -- prompt changes that alter the skill's purpose without updating metadata?
5. **Version bumped** -- version in skill.json updated appropriately?
6. **Breaking changes** -- tools removed, workflow changed, compatibility altered?

Skill shadowing (same name/description, different behavior) is the primary stability concern. Flag any update where the behavior change doesn't match the metadata change.

## Output Format

```markdown
## Skill Review

**Skill**: <name>
**Repository**: <url>
**Verdict**: APPROVE / REQUEST_CHANGES / REJECT

---

### Inclusion Criteria

| Criteria | Status | Notes |
|----------|--------|-------|
| Open Source | Pass/Fail | |
| License | Pass/Fail | <license> |
| Spec Compliance | Pass/Fail | |
| MCP Dependencies | Pass/Fail/N/A | |
| Distribution | Pass/Fail | |
| Versioning | Pass/Fail | |
| Security | Pass/Fail | |
| Documentation | Pass/Fail | |
| Community | Pass/Fail | |

### Spec Compliance

| Check | Status | Notes |
|-------|--------|-------|
| Required fields (skill.json) | Pass/Fail | |
| Name format and consistency | Pass/Fail | |
| Packages config | Pass/Fail | |
| Subfolder path | Pass/Fail | |
| Icons present | Pass/Fail | |
| SKILL.md frontmatter | Pass/Fail | |
| SKILL.md quality | Pass/Fail | |
| No auto-populated fields | Pass/Fail | |

### Security Review

- [ ] No embedded secrets or credentials
- [ ] Auth requirements documented
- [ ] Scripts safe (if applicable)
- [ ] No known CVEs in dependencies
- [ ] Scope appropriate

### Findings

**Issues (must fix):**
1. ...

**Suggestions (optional):**
1. ...

---

### Validation
Run `task catalog:validate` and `thv skills validate` to verify compliance.
```

## Submitting the Verdict

When posting to GitHub, the verdict must carry its blocking state — a plain comment does **not** gate the merge. Map the verdict to the right `gh` mechanism:

| Verdict | Command | Effect |
|---------|---------|--------|
| APPROVE | `gh pr review <pr> --approve --body-file <file>` | Approves; unblocks merge |
| REQUEST_CHANGES / REJECT | `gh pr review <pr> --request-changes --body-file <file>` | **Blocks** merge until resolved |
| Non-binding notes only | `gh pr review <pr> --comment --body-file <file>` | Review comment, no gate |

`gh pr comment` posts an ordinary comment that does **not** block — only use it for FYI notes, never to record a REJECT/REQUEST_CHANGES decision. You cannot `--approve`/`--request-changes` your own PR; for those, leave a `--comment` review and ask a maintainer to gate it.

## Error Handling

| Situation | Action |
|-----------|--------|
| Repository is private or inaccessible | Note it -- cannot verify inclusion criteria; ask submitter for evidence |
| License file missing or ambiguous | Request clarification; do not assume permissive |
| `gh` CLI errors or rate-limited | Fall back to WebFetch; note what couldn't be verified |
| `task catalog:validate` fails | Report the exact error; it must pass before approval |
| `thv skills validate` fails | Report the exact error; spec compliance is a hard requirement |
| Referenced MCP server not in catalog | Hard blocker -- skill cannot be accepted until the server is added |
| Unclear tool dependencies | Ask submitter to clarify which MCP servers/tools are needed |
| SKILL.md exceeds 500 lines | Flag as needing content split to references/ |

## Quick Reference

### Valid Values

| Field | Options |
|-------|---------|
| Status | `active`, `deprecated`, `archived` |
| Registry type | `oci`, `git` |
| Accepted licenses | `Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause` |
| Rejected licenses | `AGPL-3.0`, `GPL-2.0`, `GPL-3.0`, `LGPL-*` |

### Severity Levels (Skills vs Servers)

| Requirement | Skill Severity | Server Severity |
|-------------|---------------|-----------------|
| Open source + permissive license | Required | Required |
| Spec compliance | Required | Required |
| No known CVEs | Required | Required |
| Secure auth / sensitive info | Required | Required |
| MCP deps in catalog | Required | N/A |
| OCI distribution | Required | N/A |
| Versioning | Required | Required |
| Pinned deps / Actions | Recommended | Required |
| Provenance | Recommended | Expected |
| Security scanning | Recommended | Expected |

### Workflow Commands

```bash
task catalog:validate                                                         # Validate all entries
task catalog:build                                                            # Build registry
jq '.data.skills[] | select(.name == "<name>")' build/toolhive/registry-upstream.json  # Check skill entry
```
