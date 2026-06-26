#!/usr/bin/env python3
"""Deterministic mechanical checks for a catalog audit (checks 1 & 2).

For each server.json this resolves:
  - validity : does the pinned OCI image still exist? does repository.url resolve?
  - activity : is the upstream repo archived/disabled? how stale is it?

No LLM involved. The deep canonical-correctness check (env vars + network
permissions) is handled by the agent that drives the catalog-audit skill; this
script only produces the cheap, objective signals.

Output is a JSON array on stdout (one object per entry), ready for the skill to
merge with the deep findings.

Scope selection (the skill usually passes explicit --paths; the other flags are
conveniences):
  --paths a/server.json b/server.json   audit exactly these files
  --registry toolhive|official|all      which registry tree to scan (default toolhive)
  --tier community|official|all         filter by _meta tier (default all)
  --names arxiv-mcp-server,sqlite       restrict to these server dir names

Examples:
  ./check_entries.py --tier community
  ./check_entries.py --names hass-mcp,arxiv-mcp-server,slack-mcp-server
  ./check_entries.py --paths registries/toolhive/servers/sqlite/server.json
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

PUBLISHER_KEY = "io.modelcontextprotocol.registry/publisher-provided"
NAMESPACE_KEY = "io.github.stacklok"
DOCKYARD_PREFIX = "ghcr.io/stacklok/dockyard/"

# How long (months) before we treat a repo as stale. Surfaced as a finding by the
# skill, not a hard failure -- a stable, finished server can be legitimately quiet.
STALE_MONTHS = 12


def repo_root() -> Path:
    """Locate the repo root by walking up to the registries/ dir."""
    here = Path(__file__).resolve()
    for parent in here.parents:
        if (parent / "registries").is_dir():
            return parent
    return Path.cwd()


def discover_paths(args, root: Path) -> list[Path]:
    if args.paths:
        return [Path(p) if Path(p).is_absolute() else root / p for p in args.paths]

    registries = ["toolhive", "official"] if args.registry == "all" else [args.registry]
    names = set(args.names.split(",")) if args.names else None
    found: list[Path] = []
    for reg in registries:
        servers_dir = root / "registries" / reg / "servers"
        if not servers_dir.is_dir():
            continue
        for sj in sorted(servers_dir.glob("*/server.json")):
            if names and sj.parent.name not in names:
                continue
            found.append(sj)
    return found


def get_extension(data: dict) -> tuple[str | None, dict]:
    """Return (extension_key, extension_block). Key is identifier or remote url."""
    pkgs = data.get("packages") or []
    remotes = data.get("remotes") or []
    if pkgs:
        key = pkgs[0].get("identifier")
    elif remotes:
        key = remotes[0].get("url")
    else:
        key = None
    block = (
        data.get("_meta", {})
        .get(PUBLISHER_KEY, {})
        .get(NAMESPACE_KEY, {})
        .get(key, {})
    )
    return key, block


def parse_github_owner_repo(url: str) -> tuple[str, str] | None:
    if not url:
        return None
    m = re.search(r"github\.com[:/]+([^/]+)/([^/#?]+?)(?:\.git)?/?$", url.strip())
    if not m:
        return None
    return m.group(1), m.group(2)


def run(cmd: list[str], timeout: int) -> tuple[int, str, str]:
    try:
        proc = subprocess.run(
            cmd, capture_output=True, text=True, timeout=timeout
        )
        return proc.returncode, proc.stdout, proc.stderr
    except subprocess.TimeoutExpired:
        return 124, "", f"timeout after {timeout}s"
    except FileNotFoundError:
        return 127, "", f"command not found: {cmd[0]}"


def check_image(identifier: str, timeout: int) -> dict:
    """Confirm the *pinned* image ref still resolves. docker -> skopeo fallback."""
    if not identifier:
        return {"imageExists": None, "imageMethod": None, "imageError": "no identifier"}

    rc, _, err = run(["docker", "manifest", "inspect", identifier], timeout)
    if rc == 0:
        return {"imageExists": True, "imageMethod": "docker", "imageError": None}
    docker_err = err.strip()

    rc, _, err = run(["skopeo", "inspect", f"docker://{identifier}"], timeout)
    if rc == 0:
        return {"imageExists": True, "imageMethod": "skopeo", "imageError": None}

    # Both failed. If both tools are simply absent we cannot conclude "missing".
    if rc == 127 and "command not found: docker" in docker_err:
        return {
            "imageExists": None,
            "imageMethod": None,
            "imageError": "no image tooling available (docker/skopeo)",
        }
    return {
        "imageExists": False,
        "imageMethod": None,
        "imageError": (err.strip() or docker_err)[:300],
    }


def check_repo(repo_url: str, timeout: int) -> dict:
    """Resolve repository.url and pull activity signals via gh api."""
    out = {
        "repoResolves": None,
        "archived": None,
        "disabled": None,
        "lastPushISO": None,
        "monthsStale": None,
        "openIssues": None,
        "repoError": None,
    }
    parsed = parse_github_owner_repo(repo_url)
    if not parsed:
        out["repoError"] = "repository.url is not a parseable GitHub URL"
        return out
    owner, repo = parsed

    rc, stdout, err = run(["gh", "api", f"repos/{owner}/{repo}"], timeout)
    if rc != 0:
        # 404 / gone vs transient. gh prints the status in stderr.
        if "404" in err or "Not Found" in err:
            out["repoResolves"] = False
            out["repoError"] = "repo not found (404)"
        else:
            out["repoError"] = (err.strip() or "gh api failed")[:300]
        return out

    try:
        info = json.loads(stdout)
    except json.JSONDecodeError:
        out["repoError"] = "could not parse gh api response"
        return out

    out["repoResolves"] = True
    out["archived"] = bool(info.get("archived"))
    out["disabled"] = bool(info.get("disabled"))
    out["openIssues"] = info.get("open_issues_count")
    pushed = info.get("pushed_at")
    out["lastPushISO"] = pushed
    if pushed:
        try:
            dt = datetime.fromisoformat(pushed.replace("Z", "+00:00"))
            delta = datetime.now(timezone.utc) - dt
            out["monthsStale"] = round(delta.days / 30.44, 1)
        except ValueError:
            pass
    return out


def _strip_volatile(data: dict) -> dict:
    """Copy of the entry with CI-managed fields (metadata, tool_definitions) removed,
    so the two registry copies can be compared on their hand-edited content only."""
    d = json.loads(json.dumps(data))
    pub = d.get("_meta", {}).get(PUBLISHER_KEY, {}).get(NAMESPACE_KEY, {})
    for ext in pub.values():
        if isinstance(ext, dict):
            ext.pop("metadata", None)
            ext.pop("tool_definitions", None)
    return d


def check_official_drift(path: Path) -> dict:
    """Official-tier entries live in BOTH registries/toolhive and registries/official
    (a copy, not a symlink). From the toolhive side, compare hand-edited content against
    the official mirror and report which top-level fields have drifted.

    Returns officialDrift = None (no mirror / not a toolhive path), False (in sync), or a
    list of differing top-level keys (e.g. ["_meta", "repository"]). metadata and
    tool_definitions are ignored since CI updates each copy independently.
    """
    parts = path.parts
    if "registries" not in parts:
        return {"officialDrift": None}
    i = parts.index("registries")
    if i + 1 >= len(parts) or parts[i + 1] != "toolhive":
        return {"officialDrift": None}  # only check from the toolhive side
    official = Path(*parts[: i + 1], "official", *parts[i + 2:])
    if not official.is_file():
        return {"officialDrift": None}  # no mirror (e.g. Community-tier entries)
    try:
        a = _strip_volatile(json.loads(path.read_text()))
        b = _strip_volatile(json.loads(official.read_text()))
    except (OSError, json.JSONDecodeError) as exc:
        return {"officialDrift": f"compare error: {exc}"}
    diffs = sorted(k for k in set(a) | set(b) if a.get(k) != b.get(k))
    return {"officialDrift": diffs or False}


def audit_one(path: Path, timeout: int) -> dict:
    rel = str(path)
    result: dict = {"name": path.parent.name, "path": rel}
    try:
        data = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        result["error"] = f"could not read server.json: {exc}"
        return result

    key, ext = get_extension(data)
    identifier = None
    pkgs = data.get("packages") or []
    if pkgs:
        identifier = pkgs[0].get("identifier")
    repo_url = (data.get("repository") or {}).get("url")

    result.update(
        {
            "identifier": identifier,
            "remoteUrl": (data.get("remotes") or [{}])[0].get("url") if data.get("remotes") else None,
            "repoUrl": repo_url,
            "tier": ext.get("tier"),
            "status": ext.get("status"),
            "dockyard": bool(identifier and identifier.startswith(DOCKYARD_PREFIX)),
        }
    )

    # Check 1a: image existence (containers only).
    if identifier:
        result.update(check_image(identifier, timeout))
    else:
        result.update({"imageExists": None, "imageMethod": None, "imageError": "remote server (no image)"})

    # Check 1b + 2: repo resolves & activity. Always keyed off repository.url so
    # dockyard-repackaged entries report on the real upstream, not the wrapper.
    result.update(check_repo(repo_url, timeout))

    # Cross-registry drift: Official-tier entries are mirrored in registries/official.
    result.update(check_official_drift(path))

    # Convenience rollup the skill can sort/group on.
    stale = result.get("monthsStale")
    result["isStale"] = bool(stale is not None and stale >= STALE_MONTHS)
    return result


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--paths", nargs="*", help="explicit server.json paths")
    ap.add_argument("--registry", default="toolhive", choices=["toolhive", "official", "all"])
    ap.add_argument("--tier", default="all", choices=["community", "official", "all"])
    ap.add_argument("--names", help="comma-separated server dir names to restrict to")
    ap.add_argument("--timeout", type=int, default=60, help="per-command timeout (s)")
    ap.add_argument("--concurrency", type=int, default=8, help="parallel entries")
    ap.add_argument("--out", help="also write JSON to this file")
    args = ap.parse_args()

    root = repo_root()
    paths = discover_paths(args, root)

    # Tier filter needs the parsed file, so apply after discovery unless --paths given.
    if args.tier != "all" and not args.paths:
        want = args.tier.capitalize()  # community -> Community, official -> Official
        filtered = []
        for p in paths:
            try:
                _, ext = get_extension(json.loads(p.read_text()))
                if ext.get("tier") == want:
                    filtered.append(p)
            except (OSError, json.JSONDecodeError):
                continue
        paths = filtered

    if not paths:
        print("[]")
        print("no matching entries", file=sys.stderr)
        return 0

    print(f"auditing {len(paths)} entr{'y' if len(paths) == 1 else 'ies'}...", file=sys.stderr)
    results: list[dict] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as pool:
        futs = {pool.submit(audit_one, p, args.timeout): p for p in paths}
        for fut in concurrent.futures.as_completed(futs):
            results.append(fut.result())

    results.sort(key=lambda r: r.get("name", ""))
    payload = json.dumps(results, indent=2)
    print(payload)
    if args.out:
        Path(args.out).write_text(payload)
        print(f"wrote {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
