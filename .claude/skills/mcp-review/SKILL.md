---
name: mcp-review
description: >-
  Review MCP server specifications and updates for compliance, security, and quality.
  Use when evaluating server.json files, PRs adding/updating servers, or assessing MCP server changes.
  NOT for creating new entries (use add-mcp-server instead).
allowed-tools: Read Grep Glob Bash WebFetch
---

# MCP Server Specification Review

You are an expert reviewer for the ToolHive Registry. Evaluate server.json files and MCP server submissions for spec compliance, security, registry inclusion criteria, and completeness.

For detailed field specs, see [server-json-spec.md](references/server-json-spec.md).
For full registry inclusion criteria, see [registry-criteria.md](references/registry-criteria.md).

## Review Workflow

### Step 1: Identify Change Scope

Determine what you're reviewing:

- **New server submission** → Full review (spec + repository assessment + inclusion criteria)
- **Version update** → Focused review (changed fields, both identifier and _meta key updated, changelog)
- **Config change** → Targeted review (just the changed aspects + security implications)

### Step 2: Validate server.json

Read the server.json and check, in order:

1. **Server type** — `packages` (container) or `remotes` (remote)?
2. **Required top-level fields** — `$schema`, `name`, `description`, `title`, `version`, `repository`, `icons`
3. **Package/remote config** — identifier format, transport type valid for server type
4. **Extension key match** — `_meta` key must exactly equal `packages[0].identifier` or `remotes[0].url`
5. **Extension fields** — `tier`, `status`, `tools`, `overview` all present and valid
6. **Icons** — `icons` array present with correct `icon.svg` URL
7. **Overview format** — starts with `## Title\n\n` followed by 3-5 sentences
8. **No auto-populated fields** — reject if `metadata.*` or `tool_definitions` present in new submissions
9. **Remote-specific** — tags include `"remote"`; `oauth_config` present if OAuth required
10. **Container-specific** — `transport.url` present when type is `streamable-http`

Run `task catalog:validate` to catch schema-level issues.

### Step 3: Security Review

**Must verify:**

- [ ] Image tag pinned to specific version (never `latest`)
- [ ] Secrets marked with `isSecret: true` in `environmentVariables`
- [ ] No filesystem paths in `permissions` (users configure mounts at runtime)
- [ ] Network permissions scoped — no `insecure_allow_all: true` unless justified (e.g., fetch servers)
- [ ] Extension key matches identifier/URL exactly
- [ ] Provenance block present (Sigstore or GitHub Attestations) — applies to all tiers

### Step 4: Repository Assessment (New Submissions)

For new servers, assess the source repository against registry inclusion criteria.
See [registry-criteria.md](references/registry-criteria.md) for the full checklist and verification commands.

**Critical checks** (use `gh` CLI, GitHub MCP tools, or WebFetch — whichever is available):

1. **License** — fetch the repo's license; must be permissive (see quick reference)
2. **Dependency automation** — look for Dependabot (`.github/dependabot.yml`) OR Renovate (`renovate.json`, `.renovaterc`, `.renovaterc.json`, `.github/renovate.json`)
3. **Security policy** — check for `SECURITY.md`
4. **CI workflows** — list `.github/workflows/` contents; confirm CI runs and passes
5. **Unanswered issues** — issues open 3-4 weeks without any response is a red flag
6. **Recent activity** — check last 5 commits for recency
7. **Releases** — list recent releases; confirm semver tags and changelog

**Inclusion criteria summary:**

| Category | What to Check |
|----------|---------------|
| Open source | Public repo, permissive license (Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause) |
| Security | Provenance, pinned deps, security scanning, sensitive info handling, no known CVEs, SECURITY.md |
| Quality | CI present, tests exist, linting, code review practices |
| Stability | Semver tags, low breaking change frequency, backward compat |
| Releases | CI-based automation, regular cadence, changelog maintained |
| Documentation | README with setup, tool docs, deployment guidance |
| Community | Issues responded within 3-4 weeks, active development, contributor diversity, org backing |
| MCP compliance | Protocol support, appropriate transport type |

### Step 5: Version Update Review (Existing Servers)

For updates to existing entries:

1. **What changed?** — Diff the fields
2. **Both locations updated?** — Image tag in `packages[0].identifier` AND `_meta` extension key
3. **Changelog** — What's new in the upstream release?
4. **Tool changes** — Tools added or removed?
5. **Breaking changes** — Transport, env vars, or auth changes?
6. **Security** — New permissions, scopes, or dependencies?

Focus review on changed aspects, not full re-review.

## Output Format

```markdown
## MCP Server Review

**Server**: <name>
**Type**: Container / Remote
**Repository**: <url>
**Verdict**: APPROVE / REQUEST_CHANGES / REJECT

---

### Inclusion Criteria

| Criteria | Status | Notes |
|----------|--------|-------|
| Open Source | Pass/Fail | |
| License | Pass/Fail | <license> |
| Security Practices | Pass/Fail | |
| Code Quality | Pass/Fail | |
| Stability | Pass/Fail | |
| Documentation | Pass/Fail | |
| Community | Pass/Fail | |

### Spec Compliance

| Check | Status | Notes |
|-------|--------|-------|
| Required top-level fields | Pass/Fail | |
| Package/Remote config | Pass/Fail | |
| Extension key match | Pass/Fail | |
| Transport valid | Pass/Fail | |
| Icons present | Pass/Fail | |
| Overview format | Pass/Fail | |
| Tools listed | Pass/Fail | |
| No auto-populated fields | Pass/Fail | |
| Tags (remote tag if applicable) | Pass/Fail | |

### Security Review

- [ ] Image tag pinned (not `latest`)
- [ ] Secrets marked `isSecret: true`
- [ ] No filesystem paths in permissions
- [ ] Network permissions scoped
- [ ] Extension key matches identifier/URL
- [ ] Provenance configured

### Findings

**Issues (must fix):**
1. ...

**Suggestions (optional):**
1. ...

---

### Validation
Run `task catalog:validate` to verify spec compliance.
```

## Error Handling

| Situation | Action |
|-----------|--------|
| Repository is private or inaccessible | Note it — cannot verify inclusion criteria; ask submitter for access or evidence |
| License file missing or ambiguous | Request clarification; do not assume permissive |
| `gh` CLI errors or rate-limited | Fall back to WebFetch for README; note what couldn't be verified |
| `task catalog:validate` fails | Report the exact error; it must pass before approval |
| Unclear whether server needs OAuth | Check the upstream README/docs for auth requirements |
| Provenance info unavailable | Flag as missing — expected for all servers per registry criteria |

## Quick Reference

### Valid Values

| Field | Options |
|-------|---------|
| Tier | `Official`, `Community` |
| Status | `Active`, `Deprecated` |
| Transport | `stdio` (containers only), `streamable-http` (preferred for HTTP), `sse` (legacy) |
| Accepted licenses | `Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause` |
| Rejected licenses | `AGPL-3.0`, `GPL-2.0`, `GPL-3.0`, `LGPL-*` |

### Workflow Commands

```bash
task catalog:validate                                             # Validate all entries
task catalog:build                                                # Build registry
jq '.servers["<name>"]' build/toolhive/registry.json              # Check container entry
jq '.remote_servers["<name>"]' build/toolhive/registry.json       # Check remote entry
```
