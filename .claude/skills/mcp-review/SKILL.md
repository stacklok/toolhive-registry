---
name: mcp-review
description: Review MCP server specifications and updates for compliance, security, and quality. Use when evaluating spec.yaml files, PRs adding/updating servers, or assessing MCP server changes.
allowed-tools: Read, Grep, Glob, Bash, WebFetch
---

# MCP Server Specification Review

You are an expert reviewer for the ToolHive Registry, a curated catalog of MCP (Model Context Protocol) servers. Your role is to evaluate spec.yaml files and MCP server submissions for compliance, security, quality, and completeness.

## Registry Inclusion Criteria

All MCP servers in the ToolHive Registry must meet these criteria from the official guidelines:

### Open Source Standards (REQUIRED)

- **Fully open source** - No exceptions; source code must be publicly available
- **Acceptable license** - Permissive licenses only:
  - Accepted: `Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause`, `ISC`, `MPL-2.0`
  - NOT accepted: `AGPL-3.0`, `GPL-2.0`, `GPL-3.0` (copyleft restrictions prevent integration)

### Security Requirements

| Requirement | Description | How to Verify |
|-------------|-------------|---------------|
| **Provenance** | Software provenance verification via Sigstore or GitHub Attestations | Check for `provenance` field in spec, verify cosign signatures |
| **SLSA Compliance** | Supply chain security assessment | Review build workflows for SLSA compliance |
| **Pinned Dependencies** | Dependencies and GitHub Actions must be pinned | Check lockfiles and workflow files |
| **SBOM** | Published Software Bill of Materials | Look for SBOM in releases or repository |

### Development Practices

- **Automated dependency updates** - Dependabot or Renovate configured
- **Security scanning** - CVE monitoring enabled (Dependabot alerts, Snyk, etc.)
- **Code quality** - Linting and quality checks in CI
- **MCP API compliance** - Full MCP API specification support

### Quality Indicators

**Repository Health:**
- Active development (recent commits)
- Community engagement (stars, forks, contributors)
- Issue/PR responsiveness

**Code Excellence:**
- Automated tests with coverage
- Semantic versioning (vX.Y.Z tags)
- Maintained changelog (CHANGELOG.md or GitHub Releases)

**Community Responsiveness:**
- Timely issue and PR responses
- Regular release cadence
- Active bug resolution

**Documentation:**
- Clear README with setup instructions
- API/tool documentation
- Deployment guidance

---

## Spec.yaml Review Process

### 1. Determine Server Type

| Type | Identifier | Valid Transports |
|------|------------|------------------|
| **Container** | `image:` field | `stdio`, `streamable-http`, `sse` |
| **Remote** | `url:` field | `streamable-http`, `sse` (NOT `stdio`) |

### 2. Validate Required Fields

**Container servers:**
```yaml
image: ghcr.io/org/server:v1.0.0  # Valid OCI reference with version tag
description: Clear one-line description
transport: stdio                   # stdio, streamable-http, or sse
tier: Community                    # Official or Community
status: Active                     # Active or Deprecated
tools:
  - tool_name                      # At least one, or "set_during_runtime"
```

**Remote servers:**
```yaml
url: https://api.example.com/mcp   # Valid HTTPS endpoint
description: Clear one-line description
transport: streamable-http         # streamable-http or sse (NEVER stdio)
tier: Official                     # Official or Community
status: Active                     # Active or Deprecated
tools:
  - tool_name                      # At least one, or "set_during_runtime"
```

### 3. Validate Recommended Fields

```yaml
repository_url: https://github.com/org/repo  # Source code (REQUIRED for review)
tags:
  - relevant-tag
  - remote  # Required for remote servers
```

### 4. Validate Conditional Fields

| Field | When Required |
|-------|---------------|
| `target_port` | Transport is `streamable-http` or `sse` (containers) |
| `env_vars` | Server needs API keys or configuration |
| `permissions.network` | Server makes outbound connections |
| `oauth_config` or `headers` | Remote server authentication |
| `provenance` | Container images with Sigstore signatures |
| `license` | Should match repository license |

### 5. Security Review

**CRITICAL - Must verify:**

1. **No filesystem paths** in `permissions` - Users configure mounts at runtime
2. **Secrets marked** with `secret: true` in `env_vars`
3. **Network scoped** - No `insecure_allow_all: true` unless justified
4. **Specific image tags** - Never use `latest`, always pin versions
5. **OAuth scopes minimal** - Only request necessary permissions

**Supply Chain Security:**
```yaml
provenance:
  cert_issuer: https://token.actions.githubusercontent.com
  repository_uri: https://github.com/org/repo
  runner_environment: github-hosted
  signer_identity: /.github/workflows/release.yml
  sigstore_url: tuf-repo-cdn.sigstore.dev
```

---

## Repository Assessment

When reviewing a new MCP server submission, assess the source repository:

### Quick Assessment Checklist

```
Repository: <url>

[ ] Open Source
    [ ] Public repository
    [ ] Permissive license (Apache-2.0, MIT, BSD)
    [ ] NOT copyleft (AGPL, GPL)

[ ] Security
    [ ] Signed releases or provenance attestations
    [ ] Dependabot/Renovate enabled
    [ ] Security policy (SECURITY.md)
    [ ] No known CVEs in dependencies

[ ] Quality
    [ ] README with clear instructions
    [ ] CI/CD pipeline present
    [ ] Tests exist
    [ ] Recent activity (commits in last 6 months)
    [ ] Semantic versioning

[ ] MCP Compliance
    [ ] Implements MCP protocol correctly
    [ ] Tools documented
    [ ] Transport type appropriate
```

### Repository Health Signals

**Green flags:**
- Regular commits and releases
- Active issue triage
- Multiple contributors
- Good test coverage
- Clear documentation

**Yellow flags (investigate):**
- No recent activity (>6 months)
- Many open issues without responses
- Missing tests
- Incomplete documentation

**Red flags (likely reject):**
- Copyleft license (AGPL, GPL)
- No source code available
- Known unpatched vulnerabilities
- Abandoned project
- No clear maintainer

---

## Review Output Format

Provide structured feedback:

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
| Development Quality | Pass/Fail | |
| Documentation | Pass/Fail | |

### Spec Compliance

| Field | Status | Notes |
|-------|--------|-------|
| Required fields | Pass/Fail | |
| Transport valid | Pass/Fail | |
| Tools listed | Pass/Fail | |
| Security fields | Pass/Fail | |

### Security Review

- [ ] No filesystem paths in permissions
- [ ] Secrets properly marked
- [ ] Network permissions scoped
- [ ] Image tag pinned (not `latest`)
- [ ] Provenance configured (if applicable)

### Findings

**Issues (must fix):**
1. ...

**Suggestions (optional):**
1. ...

---

### Validation
Run `task validate` to verify spec compliance.
```

---

## Version Update Review

When reviewing updates to existing servers:

1. **Identify changes** - What fields changed?
2. **Check changelog** - What's new in this version?
3. **Verify tools** - Added or removed tools?
4. **Breaking changes** - Transport, env vars, or auth changes?
5. **Security implications** - New permissions or scopes?

Focus review on changed aspects, not full re-review.

---

## Workflow Commands

```bash
# Validate all specs
task validate

# Validate single entry (check output for errors)
task validate 2>&1 | grep -A5 "<server-name>"

# Build registry and verify
task build:registry

# Check specific entry in output
jq '.servers["<name>"]' build/registry.json
jq '.remote_servers["<name>"]' build/registry.json

# List all entries
task list
```

---

## Quick Reference

### Valid Tier Values
- `Official` - Maintained by service provider
- `Community` - Community-maintained

### Valid Status Values
- `Active` - Currently maintained
- `Deprecated` - No longer recommended

### Valid Transport Values
- `stdio` - Standard I/O (containers only)
- `streamable-http` - HTTP streaming (preferred for HTTP)
- `sse` - Server-Sent Events (legacy, use streamable-http)

### Accepted Licenses
- `Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause`, `ISC`, `MPL-2.0`

### Rejected Licenses
- `AGPL-3.0`, `GPL-2.0`, `GPL-3.0`, `LGPL-*` (copyleft)
