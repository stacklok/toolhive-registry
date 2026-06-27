# Catalog audit criteria

What each check looks for, how to rate severity, and the conventions the deep
check must apply. The bar here is the same one the `mcp-review` skill enforces on
new submissions; this audit applies it to entries already in the catalog. When a
detail isn't covered below, defer to:

- [`../../mcp-review/references/server-criteria.md`](../../mcp-review/references/server-criteria.md) - MCP-specific security/stability/docs bar
- [`../../mcp-review/references/registry-criteria.md`](../../mcp-review/references/registry-criteria.md) - shared open-source/license/quality bar
- [`../../mcp-review/references/server-json-spec.md`](../../mcp-review/references/server-json-spec.md) - field-level schema, env vars, permissions

## Check 1 - still valid (mechanical, from `check_entries.py`)

- **Image exists.** The pinned `packages[0].identifier` must still resolve. The
  script tries `docker manifest inspect` then `skopeo inspect`. `imageExists:false`
  is a hard problem; `imageExists:null` means the tooling couldn't run (inconclusive,
  not a pass).
- **Repo resolves.** `repository.url` must return 200. `repoResolves:false` with a
  404 means the upstream repo was deleted, renamed, or made private.

## Check 2 - still active (mechanical, from `check_entries.py`)

- **archived / disabled.** A GitHub-archived or disabled repo means the project is
  no longer maintained.
- **staleness.** `monthsStale` is months since the last push. `isStale` flips at 12
  months. Staleness is a signal, not a verdict: a small, finished server (e.g. a
  stdlib `time` server) can be legitimately quiet, while a fast-moving one going
  quiet for a year is a real concern. Weigh it against the server's scope.
- For **dockyard-repackaged** images (`ghcr.io/stacklok/dockyard/...`), activity is
  measured on the upstream `repository.url`, never the dockyard wrapper.

## Check 3 - still canonically correct (deep, agent reasoning)

Read the upstream README + relevant source, then compare reality against the entry.

### Environment variables (`packages[0].environmentVariables[]`)

Flag, with the upstream evidence (file + line/section or URL):

- **Stale / declared-but-unused** - the var no longer appears in upstream config or code.
- **Missing required** - upstream documents a required var the entry omits (the server
  would fail to start or silently misbehave without it).
- **`isSecret` mismatch** - a credential (token, key, password, secret) is not marked
  `isSecret: true`. Conversely, a non-credential (a file path, a URL, a flag) that is
  marked secret. `isSecret` is for *actual credentials only*, not file references.
- **`isRequired` mismatch** - the entry's required/optional flag disagrees with upstream.
- **Weak description** - empty, placeholder, or copied-boilerplate description.

Conventions to apply (house style for this registry):

- **`default` field is allowed, not a defect.** It's a valid schema field, and a
  machine-readable default can be valuable (a UI can prefill it; tooling can show the
  effective value). Do **not** flag a `default` just for existing. Only flag a default that
  is *wrong or misleading*: it contradicts upstream's actual default, or bakes in a value
  that breaks under ToolHive's container model (e.g. binding `127.0.0.1` where the container
  needs `0.0.0.0`). Whether to prefer an explicit `default` vs. noting it in the description
  is a judgment call, not settled precedent - don't litigate it.
- Descriptions should say what the var is for and, where relevant, the expected
  format or example.

### Network permissions (`_meta[...]io.github.stacklok.<id>.permissions.network.outbound`)

Shapes in use: `{}` (no outbound), `{allow_host:[...], allow_port:[...]}` (scoped),
`{insecure_allow_all:true}` (open). Flag:

- **`insecure_allow_all: true` where scoping is feasible** - if the upstream docs make
  clear which hosts the server contacts (e.g. a single SaaS API), an open profile is
  too broad. Propose the concrete `allow_host` list it should use. (A server that
  legitimately reaches arbitrary user-supplied hosts - a fetch/scraper/browser - may
  justify it; say so rather than flagging blindly.)
- **Missing host** - the server needs a host that isn't in `allow_host`; the entry as
  written would break that functionality. (Highest-value finding: it's a real bug.)
  **Before flagging, check wildcard coverage:** `allow_host` is Squid-style, so a
  leading-dot entry (`.example.com`) already matches `example.com` and *every* subdomain.
  List the actual `allow_host` and confirm the needed host isn't already covered by a
  leading-dot entry - e.g. `.docs.aws.amazon.com` covers `x.docs.aws.amazon.com`, but does
  NOT cover a different domain like `proxy.search.docs.aws.com`. (This exact distinction was
  missed by both the finder and a verifier in the first run; check it explicitly.)
- **Overly broad** - `allow_host`/`allow_port` wider than what the server uses, including a
  bare apex (`sentry.io`) that a sibling leading-dot entry (`.sentry.io`) already covers.
- **Host format** - subdomain wildcards are a leading-dot prefix (`.github.com`), not
  globs. Flag malformed hosts.

#### ToolHive runs every server in its own container - mind localhost

This changes how to read any profile that targets a local service. Inside the container,
`localhost` / `127.0.0.1` is the **container itself**, not the user's host machine. So:

- A server documented to reach a **host-side** service (a local app, an on-host daemon, a
  `127.0.0.1:<port>` API the user runs) cannot get there via `localhost`. The correct profile
  uses `host.docker.internal` in `allow_host` plus the runtime flag `--allow-docker-gateway`
  (on Linux without Docker Desktop, the bridge gateway IP such as `172.17.0.1`). Critically,
  the Docker gateway is denied by default **even when `insecure_allow_all: true`**, so an open
  profile alone does *not* grant host access.
- Therefore: if `allow_host` lists `localhost`/`127.0.0.1`, or upstream docs say the target is
  a localhost service, that's a finding - the entry likely needs `host.docker.internal` and may
  not function as written. Do **not** recommend "scope to localhost"; recommend
  `host.docker.internal` (and note the `--allow-docker-gateway` requirement).
- To reach **another workload on the same container network**, the profile lists that
  workload's hostname and port (not localhost).
- Ref: https://docs.stacklok.com/toolhive/guides-cli/network-isolation#accessing-other-workloads-on-the-same-container-network

### Remote servers (`remotes[]`, no container)

Remote entries have no env vars or permission profile, so check 3 is mostly endpoint health:

- **Endpoint liveness** - POST an MCP `initialize` to `remotes[0].url`. A `200`, an SSE
  stream, or a `401`/`403` all mean the endpoint is **alive** - auth/OAuth-gated is healthy,
  not a finding. Only DNS failure, connection refused, `404`, `5xx`, or a TLS/cert error mean
  it may be down (a high finding). Do not report an OAuth-gated `401` as broken.
- **oauth_config sanity** - if present, `authorize_url`/`token_url` hosts should be consistent
  with the endpoint.

### Dual-registry drift (Official tier)

Official-tier entries exist in **both** `registries/toolhive/servers/<name>/` and
`registries/official/servers/<name>/` (a copy, not a symlink). If the two copies disagree on
hand-edited fields (env vars, permissions, `repository.url`, tier), that drift is a finding -
report it. Auto-populated `metadata` / `tool_definitions` differences are expected (CI updates
each copy independently) and are not findings.

### Not findings (do not flag these)

The first community-tier run over-reported these; they are expected and correct:

- **ToolHive-managed runtime vars.** `MCP_PORT`, `MCP_TRANSPORT`, `*_MCP_BIND_HOST` / `_PORT`,
  `*_TRANSPORT`, `SLACK_MCP_PORT`, and similar transport/port knobs are injected by ToolHive at
  runtime. They are intentionally absent from catalog entries - never flag them as "missing."
- **Undeclared *optional* env vars.** The registry does not require exhaustive enumeration of
  every tuning knob (timeouts, retry counts, log levels, output dimensions, rate limits). An
  optional var that upstream defaults sensibly is not a defect when omitted. Only flag a missing
  env var when it is genuinely **required** for the server (or a major tool) to function.
- **Alternative auth methods.** If an entry supports one documented auth path (e.g. an API
  token) and upstream also offers another (e.g. full OAuth), the unused alternative's vars are
  not "missing required" - they're optional.

Be self-consistent: if your evidence shows a var *is* used upstream, it is not "declared-but-
unused." Re-read your own evidence before writing the finding.

## Severity scale

- **critical** - entry is broken: `imageExists:false`, or `repoResolves:false` (404).
- **high** - project archived/disabled; a required env var is missing; the permission
  profile omits a host the server needs (breaks functionality).
- **medium** - stale (~12 months+) for an actively-scoped server; `insecure_allow_all`
  where a concrete scoped list is feasible; `isSecret`/`isRequired` mismatch; stale
  declared env var.
- **low** - weak/placeholder descriptions, overly-broad-but-harmless permissions,
  cosmetic issues.

Inconclusive results (`null` image/repo checks, upstream docs that couldn't be reached)
are **not** passes - report them as "could not verify" so a human knows to look.
