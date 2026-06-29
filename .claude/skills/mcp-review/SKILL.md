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
For full registry inclusion criteria, see [registry-criteria.md](references/registry-criteria.md) and [server-criteria.md](references/server-criteria.md).

## Registry Layout (read first)

The repo has two server registries, and submissions land in the wrong one often enough to check first:

- `registries/toolhive/servers/` — the **contribution path**. All new submissions go here.
- `registries/official/` — a **curated subset** maintained by the ToolHive team. Contributors should not hand-add entries here.

Both normalize every entry under the **`io.github.stacklok`** publisher namespace (`name: io.github.stacklok/<server-name>` and the `_meta` publisher key), with a repo-hosted `icon.svg`. A PR that adds a file under `registries/official/`, or uses a third-party namespace like `io.github.<contributor>`, is a finding — not a style nit. `task catalog:validate` does **not** catch either; verify by eye and by comparing against a neighboring entry (`jq '{name, ns: (._meta[...]|keys)}' <neighbor>/server.json`).

## Review Workflow

### Step 1: Identify Change Scope

Determine what you're reviewing:

- **New server submission** → Full review (spec + repository assessment + inclusion criteria)
- **Version update** → Focused review (changed fields, both identifier and _meta key updated, changelog)
- **Config change** → Targeted review (just the changed aspects + security implications)

### Step 2: Validate server.json

Read the server.json and check, in order:

1. **Location & namespace** — file is under `registries/toolhive/servers/<name>/`; `name` and the `_meta` publisher key are both `io.github.stacklok` (see [Registry Layout](#registry-layout-read-first))
2. **Server type** — `packages` (container) or `remotes` (remote)?
3. **Required top-level fields** — `$schema`, `name`, `description`, `title`, `version`, `repository`, `icons`
4. **Package/remote config** — identifier format, transport type valid for server type
5. **Extension key match** — `_meta` key must exactly equal `packages[0].identifier` or `remotes[0].url`
6. **Extension fields** — `tier`, `status`, `tools`, `overview` all present and valid
7. **Tier matches maintainer** — `Official` is reserved for the ToolHive team, the MCP spec authors, or the platform owner of the integrated service. A third-party contributor's server is `Community`, regardless of what the PR claims (don't trust the field — check who maintains the repo)
8. **Icons** — repo-hosted `icon.svg` (`mimeType: image/svg+xml`) AND an `icon.svg` file present in the server dir. Flag external icon URLs (e.g. a vendor PNG) — they're a stability/supply-chain risk and break the convention
9. **Overview format** — starts with `## Title\n\n` followed by 3-5 sentences
10. **No auto-populated fields** — reject if `metadata.*` or `tool_definitions` present in new submissions
11. **Remote-specific** — tags include `"remote"`; `oauth_config` present if OAuth required
12. **Container-specific** — `transport.url` present when type is `streamable-http`

Run `task catalog:validate` to catch schema-level issues. Note its limits: it validates the JSON schema, **not** the registry-location, namespace, tier-semantics, or icon-host conventions above — a passing validate does not mean the entry is correctly placed or classified.

**Remote endpoint check (remotes only):** confirm the URL actually serves MCP and the advertised `tools` match. Don't trust the PR description:

```bash
curl -sS -X POST "<remote-url>" \
  -H "Content-Type: application/json" -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"review","version":"0"}}}'
# then tools/list (id:2, method:"tools/list") and compare names against the entry's `tools`
```

A plain `GET` often 404s on streamable-HTTP endpoints — that's not a failure; use the `initialize` POST. If the endpoint needs auth, note it and confirm `oauth_config`/auth handling is declared.

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
See [server-criteria.md](references/server-criteria.md) for the full checklist and verification commands.

**Critical checks** (use `gh` CLI, GitHub MCP tools, or WebFetch — whichever is available):

1. **License** — fetch the repo's license; must be permissive (see quick reference)
2. **Dependency automation** — look for Dependabot (`.github/dependabot.yml`) OR Renovate (`renovate.json`, `.renovaterc`, `.renovaterc.json`, `.github/renovate.json`)
3. **Security policy** — check for `SECURITY.md`
4. **CI workflows** — list `.github/workflows/` contents; confirm CI runs and passes
5. **Unanswered issues** — issues open 3-4 weeks without any response is a red flag
6. **Recent activity** — check last 5 commits for recency
7. **Releases** — list recent releases; confirm semver tags and changelog
8. **Repo & author maturity** — repo age, star count, contributor diversity, and GitHub account age. A days-old repo, zero stars, no releases, and a single brand-new account are a cluster of red flags (e.g. `gh repo view <owner>/<repo> --json createdAt,stargazerCount,licenseInfo,pushedAt`)
9. **Documentation relevance** — the README must actually describe *this* server and its tools. A generic or boilerplate README (e.g. an unmodified upstream template that documents unrelated demo tools) fails the documentation criterion even when a README file exists

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
| Location & namespace (`toolhive/`, `io.github.stacklok`) | Pass/Fail | |
| Required top-level fields | Pass/Fail | |
| Package/Remote config | Pass/Fail | |
| Extension key match | Pass/Fail | |
| Tier matches maintainer | Pass/Fail | |
| Transport valid | Pass/Fail | |
| Remote endpoint serves MCP (remotes) | Pass/Fail/N/A | |
| Icons (repo-hosted SVG + file present) | Pass/Fail | |
| Overview format | Pass/Fail | |
| Tools listed (match live server) | Pass/Fail | |
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
| Repository is private or inaccessible | Note it — cannot verify inclusion criteria; ask submitter for access or evidence |
| License file missing or ambiguous | Request clarification; do not assume permissive |
| `gh` CLI errors or rate-limited | Fall back to WebFetch for README; note what couldn't be verified |
| `task catalog:validate` fails | Report the exact error; it must pass before approval |
| Unclear whether server needs OAuth | Check the upstream README/docs for auth requirements |
| Provenance info unavailable | Flag as missing — expected for all servers per registry criteria |
| Entry added under `registries/official/` or with a non-`io.github.stacklok` namespace | Flag as a structural finding; direct the contributor to `registries/toolhive/servers/` under the stacklok namespace |
| Remote endpoint unreachable or doesn't speak MCP | Re-check with the `initialize` POST (not a bare `GET`); if it still fails, this blocks approval — the URL is the entire contract for a remote |

## Quick Reference

### Severity (what blocks vs. what's a note)

| Requirement | Severity |
|-------------|----------|
| Open source + permissive license | Required (hard gate — no license = reject) |
| Spec compliance (`task catalog:validate` passes) | Required |
| Correct location & `io.github.stacklok` namespace | Required |
| Tier matches maintainer identity | Required |
| Remote endpoint serves MCP + tools match | Required (remotes) |
| No known critical/high CVEs | Required |
| Secrets `isSecret: true`, no filesystem paths | Required |
| Semver versioning | Required |
| Documentation describes this server | Required |
| Pinned deps / Actions pinned to SHAs | Required |
| Provenance (Sigstore / attestations) | Expected |
| Security scanning in CI, `SECURITY.md` | Expected |
| Repo/author maturity, stars, community traction | Weighed (a cluster of misses can sink an otherwise-valid entry) |

A `Required` miss is REQUEST_CHANGES or REJECT. `Expected` misses are called out and weigh on the verdict. `Weighed` signals inform borderline calls.

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
jq '.data.servers[] | select(.name == "io.github.stacklok/<name>")' build/toolhive/registry-upstream.json
```
