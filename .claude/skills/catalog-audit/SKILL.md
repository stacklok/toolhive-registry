---
name: catalog-audit
description: >-
  Audit existing ToolHive catalog server entries for drift: whether the pinned
  image still exists, the upstream repo is still alive and active, and the declared
  env vars and network-permission profile still match what the server actually
  needs. Use whenever asked to audit, health-check, sweep, or spot-check catalog
  entries / server.json files for staleness, broken images, dead repos, abandoned
  projects, or wrong env-vars/permissions - for the community tier, the official
  tier, the whole registry, or a named subset. Report-only; it never edits entries.
  NOT for reviewing a new submission (use mcp-review) or creating one (use add-mcp-server).
allowed-tools: Read Grep Glob Bash WebFetch Task
---

# Catalog audit

Audit catalog entries that are **already in the registry** for three kinds of drift:

1. **Still valid** - the pinned OCI image still resolves; `repository.url` still loads.
2. **Still active** - the upstream repo isn't archived/disabled and isn't stale.
3. **Still canonically correct** - declared env vars and the network-permission
   profile match what the upstream project actually documents and requires.

This is **report-only**. Produce findings for a human to triage; never edit a
`server.json`, open a PR, or change permissions. The output is a dated markdown
report under `audits/`.

Checks 1 and 2 are mechanical and run via the bundled script. Check 3 needs
judgment over upstream docs and is done by fanning out sub-agents. Grading rules
and severity live in [references/audit-criteria.md](references/audit-criteria.md) -
read it before writing findings.

## Step 1 - resolve scope

Infer scope from the request, don't ask for rigid flags. Map intent to a server set:

- "the community servers" / "community tier" -> `--tier community`
- "official servers" -> `--tier official`
- "the whole catalog" / "everything" -> `--tier all`
- "re-check X, Y, Z" / a list of names -> `--names X,Y,Z`
- a specific registry tree -> `--registry toolhive|official|all` (default `toolhive`)

The repo has two server trees: `registries/toolhive/servers/` (contribution path,
default) and `registries/official/servers/` (curated subset). Default to `toolhive`
unless the user clearly means otherwise.

Skills (`registries/toolhive/skills/`) are **out of scope** unless explicitly asked.
If asked, audit them separately (their `packages[]` carry `registryType: git` with a
`ref`/`subfolder` and `registryType: oci` images - check the git ref resolves and the
repo is active; there are no env vars or network permissions to canonicalize).

State the resolved scope and the entry count back to the user before crawling, so a
mistaken scope is caught early.

## Step 2 - mechanical checks (script)

Run the bundled deterministic script. It needs `gh` (authenticated) and `docker` or
`skopeo` on PATH.

```bash
python3 .claude/skills/catalog-audit/scripts/check_entries.py \
  --tier community \
  --out <scratch>/mechanical.json
```

Use `--names a,b,c` for a subset, or `--paths .../server.json ...` for exact files.
The script prints a JSON array (also written to `--out`): per entry it reports
`imageExists` / `imageMethod`, `repoResolves`, `archived`, `disabled`, `lastPushISO`,
`monthsStale` / `isStale`, `openIssues`, plus `dockyard` and `tier`. `null` on an
image/repo field means inconclusive (tooling couldn't run / transient), **not** a pass.

It also reports `officialDrift`: for an Official-tier entry, whether its
`registries/official` mirror has drifted from the `registries/toolhive` source -
`false` (in sync), a list of differing top-level keys (e.g. `["repository"]`), or `null`
(no mirror, e.g. Community-tier). A non-empty list is a finding worth reporting and means
a fix needs applying to both copies.

## Step 3 - deep canonical-correctness checks (sub-agents)

For each in-scope entry, spawn a sub-agent to do check 3. Keep it cheap: these are
independent, doc-reading tasks. For a handful of entries, launch `Task` agents in
parallel (batches of ~8). For a large sweep (dozens), prefer the `Workflow` tool to
pipeline the fan-out if the user has opted into it; otherwise batch `Task` agents.

Give each sub-agent:

- the `server.json` path and its parsed env vars + `permissions` block,
- the `repository.url` (and note if `dockyard:true`, so it reads the upstream repo,
  not the dockyard wrapper),
- the relevant criteria from [references/audit-criteria.md](references/audit-criteria.md).

Instruct each agent to read the upstream README and relevant source (via `gh api
repos/{owner}/{repo}/readme`, `gh api`/`get_file_contents` for files, `search_code`
for env-var references, or `WebFetch`), then return **structured findings only**:

```json
{
  "name": "<server>",
  "envVars": [{"name": "...", "issue": "...", "severity": "...", "evidence": "<url-or-file:line>"}],
  "permissions": [{"issue": "...", "severity": "...", "evidence": "..."}],
  "couldNotVerify": ["..."]
}
```

Each finding must cite upstream evidence (a URL or `file:line`). Findings without
evidence are not actionable - drop them or mark `couldNotVerify`.

## Step 4 - verify findings before reporting (adversarial)

Single-pass deep checks over-report - this is not optional cleanup, it's a core step.
A cheap finder reliably flags plausible-but-wrong issues: a host that a leading-dot
`allow_host` already covers, an "unused" var its own evidence shows is used, an
`insecure_allow_all` that is actually justified, a 401 from a healthy OAuth-gated remote.

For every high/critical finding (and ideally all of them), spawn an **independent agent -
a stronger model than the finder** - to re-check it against upstream, prompted to refute by
default: "confirm only if you can re-derive this from upstream; mark refuted if the entry is
actually correct; uncertain if you can't tell." Drop refuted findings; keep confirmed ones
with the verifier's note. In the first official-tier run this stage refuted ~6 high and ~21
medium false positives that had survived the finder.

When fixes are later applied, re-verify each one against upstream while editing (grep the
actual var name / host) - a cheap third check that catches the cases a verifier still misses
(e.g. a contested host the second pass got wrong).

## Step 5 - merge and write the report

Merge mechanical results (step 2) with deep findings (step 3) into severity-ranked
findings using the scale in the criteria file. Write to
`audits/<scope>-<YYYY-MM-DD>.md` (create `audits/` if absent; get the date with
`date +%F`). Use this structure:

```markdown
# Catalog audit - <scope> - <date>

Scope: <description> (<N> entries). Methods: docker/skopeo image probe, gh repo API,
sub-agent upstream-doc review. Report only - no entries were modified.

## Summary
| entry | image | repo | active | env issues | perm issues | worst |
|-------|-------|------|--------|-----------|-------------|-------|
...

## Critical
...
## High
...
## Medium
...
## Low
...

## Could not verify
<entries/fields where a check was inconclusive>

## Clean
<entries with no findings>
```

Each finding lists the entry, category, detail, evidence (URL/command), and a
suggested action for the human (since this is report-only). Lead with the worst.

## Step 6 - hand off

Show the user the summary table and the top findings inline. Do **not** commit the
report or edit any entry unless the user explicitly asks. If they want fixes, that's
a separate pass (and `mcp-review` / `add-mcp-server` cover the editing rules). When fixing
an Official-tier entry, remember it has a mirror copy (see Notes) that must be updated too.

## Notes

- **Inconclusive is not clean.** Surface `null` image/repo results and unreachable
  upstream docs as "could not verify," never as passing.
- **Check wildcard coverage before flagging a missing host.** `allow_host` is Squid-style:
  a leading-dot entry (`.example.com`) already matches every subdomain. List the actual
  `allow_host` and confirm the needed host isn't already covered before reporting it missing.
  (Both the finder and a verifier missed this on `aws-documentation` in the first run.)
- **A remote endpoint returning 401/403 is alive,** not a finding - it's auth/OAuth-gated.
  Only DNS failure, connection refused, 404, 5xx, or TLS errors mean it may be down.
- **Official-tier entries are mirrored in `registries/official/servers/`** (a copy, not a
  symlink). Flag when an entry's `registries/official` copy has drifted from its
  `registries/toolhive` source - that drift is itself a finding. And any fix to an
  Official-tier entry must be applied to both copies (CI reconciles only the auto-populated
  `metadata` / `tool_definitions`, not the hand-edited fields).
- **Be fair on staleness.** A small, finished server can be quiet for good reasons;
  weigh `monthsStale` against the server's scope before flagging.
- **Be fair on `insecure_allow_all`.** Servers that reach arbitrary user-supplied
  hosts (fetch, scrapers, browsers, db/cluster/registry clients) may legitimately need it;
  flag it only where a concrete scoped host list is feasible, and propose that list.
