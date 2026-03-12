---
name: ci-agent-hardening
description: >-
  Audit and harden GitHub Actions workflows against prompt injection,
  pull_request_target exploits (Pwn Requests), expression injection,
  cache poisoning, credential theft, and supply chain attacks. Based on
  Clinejection and hackerbot-claw campaigns. Use when reviewing CI/CD
  security, securing AI agent workflows, hardening publishing pipelines,
  or checking for GitHub Actions misconfigurations. Also covers slash
  command authorization, CLAUDE.md protection, and network egress.
  NOT for general CI/CD optimization or non-security workflow issues.
version: "0.1.0"
license: Apache-2.0
metadata:
  author: Stacklok
  homepage: https://github.com/stacklok/toolhive-catalog
---

# CI Agent Hardening

Audit and fix security vulnerabilities in GitHub Actions workflows, with emphasis on AI-powered CI/CD. Based on two major 2026 incidents:

- **Clinejection** — Prompt injection in AI triage bot led to cache poisoning, credential theft, malicious npm publish (~4,000 developers affected)
- **hackerbot-claw** — Autonomous AI bot exploited 7 repos (Microsoft, DataDog, CNCF, Aqua Trivy), achieving full repo compromise on Trivy (25k+ stars) via PAT theft

## Prerequisites

- Repository with `.github/workflows/` directory
- `bash` available for running the audit script

## Instructions

For background and exploit details on any pattern below, see [references/ATTACK-PATTERNS.md](references/ATTACK-PATTERNS.md).

### Step 1: Run the Automated Audit

Run from the repository root:

```bash
bash <skill-path>/scripts/audit-workflows.sh .github/workflows
```

Review `[CRIT]` findings first (exploitable now), then `[WARN]` (defense gaps).

### Step 2: Fix `pull_request_target` + Fork Checkout (CRITICAL)

This is the #1 attack vector. 4 of 7 hackerbot-claw attacks exploited it. Check every `pull_request_target` workflow:

```yaml
# DANGEROUS — runs fork code with your secrets
on: pull_request_target
steps:
  - uses: actions/checkout@v4
    with:
      ref: ${{ github.event.pull_request.head.sha }}  # ATTACKER'S CODE

# SAFE — use pull_request trigger instead (no secret access)
on: pull_request

# SAFE — if pull_request_target needed, checkout base only
on: pull_request_target
steps:
  - uses: actions/checkout@v4  # defaults to base branch
```

If the workflow needs both secrets AND fork code (e.g., comment on PR), split into two workflows: one that runs untrusted code (no secrets) and one that consumes artifacts (with secrets).

### Step 3: Eliminate Expression Injection in `run:` Blocks

Any `${{ }}` expression referencing user-controlled values inside a `run:` block is a shell injection. Branch names, filenames, PR titles, commit messages are all vectors.

```yaml
# VULNERABLE — branch name is shell-evaluated
- run: echo "${{ github.event.pull_request.head.ref }}"

# SAFE — environment variable (properly quoted by runner)
- run: echo "$PR_HEAD_REF"
  env:
    PR_HEAD_REF: ${{ github.event.pull_request.head.ref }}
```

Check ALL `run:` blocks for these user-controlled fields:
- `github.event.issue.title` / `.body`
- `github.event.pull_request.title` / `.body` / `.head.ref`
- `github.event.comment.body`
- `github.event.discussion.title` / `.body`
- `github.event.review.body`
- `github.event.head_commit.message`

### Step 4: Secure Slash Commands

Any `issue_comment`-triggered workflow that runs code must check `author_association`:

```yaml
# VULNERABLE — any GitHub user can trigger
if: contains(github.event.comment.body, '/deploy')

# SAFE — restrict to maintainers
if: >-
  contains(github.event.comment.body, '/deploy') &&
  (github.event.comment.author_association == 'MEMBER' ||
   github.event.comment.author_association == 'OWNER')
```

### Step 5: Harden AI Agent Workflows

1. **Minimize tools** — Triage/review bots need `Read` at most. NEVER grant `Bash` to workflows triggered by external users.
2. **Remove wildcard users** — Delete `allowed_non_write_users: "*"`. Scope to known collaborators.
3. **Restrict permissions** — AI workflows should have `contents: read` at most.
4. **Protect CLAUDE.md** — Add to `CODEOWNERS` requiring maintainer review. Validate against expected schema in CI.
5. **Never checkout fork code** for AI review — use base branch checkout only.

### Step 6: Set Least-Privilege Permissions

Every workflow must have an explicit `permissions:` block:

```yaml
permissions:
  contents: read
  pull-requests: read  # or write only if needed
```

A quality check script does NOT need `contents: write`. A `GITHUB_TOKEN` with write access, if stolen, allows pushing commits and merging PRs.

### Step 7: Harden Credential Management

1. **Use OIDC provenance** — `npm publish --provenance`, PyPI trusted publishers. Eliminates static tokens.
2. **Prefer `GITHUB_TOKEN`** over PATs — auto-scoped, expires after workflow.
3. **If PAT required** — use fine-grained PATs with minimal scope, stored in GitHub Environments with protection rules.
4. **Separate nightly/production credentials** — different tokens, scopes, and workflows.
5. **Verify rotation** — test-publish a canary after rotating to confirm old token is revoked.

### Step 8: Eliminate Cache Attack Surface

1. **Remove `actions/cache` from release workflows** — install dependencies fresh.
2. **If cache required** — use workflow-specific key prefixes that cannot be written by other workflows.
3. **Monitor for anomalies** — cache size >10GB, sudden miss rate spikes.

### Step 9: Network Egress Monitoring

Every attack in the hackerbot-claw campaign depended on `curl` to `hackmoltrepeat.com`. Consider:
- StepSecurity Harden-Runner for network egress allowlisting
- Alert on outbound calls to unknown domains during CI
- Block `curl`/`wget` to non-allowlisted domains

### Step 10: Disclosure Readiness

1. **`SECURITY.md`** with contact info and SLA (Cline ignored reports for 40 days).
2. **Monitor security inbox** actively.
3. **Credential rotation runbook** — tested before you need it.
4. **Incident response plan** covering repo takeover scenario (Trivy: repo renamed, releases deleted, malicious extension published).

## Generating Fix PRs

Apply changes per-file, explain each change. Priority order:

1. Remove `pull_request_target` + fork checkout patterns (or split workflows)
2. Move `${{ }}` expressions to `env:` variables in all `run:` blocks
3. Add `author_association` checks to slash command workflows
4. Add explicit `permissions:` blocks with least privilege
5. Reduce AI agent `allowed_tools` to minimum
6. Remove `allowed_non_write_users: "*"`
7. Add `--provenance` to publish commands
8. Remove `actions/cache` from release workflows
9. Add `CLAUDE.md` to `CODEOWNERS`

## Error Handling

| Issue | Cause | Solution |
|-------|-------|----------|
| Script finds no workflows | Wrong path | Verify repo root, check `.github/workflows/` |
| False positive on AI detection | Keyword match in comments | Review context manually |
| `pull_request_target` needed for secrets | Workflow design requires it | Split into two workflows (see Step 2) |
| OIDC provenance fails | Registry not configured | Follow npm/PyPI OIDC setup docs |

## See Also

- [references/ATTACK-PATTERNS.md](references/ATTACK-PATTERNS.md) — Full details on Clinejection + hackerbot-claw: all 7 attacks, exploit code, IOCs, comprehensive hardening checklist
