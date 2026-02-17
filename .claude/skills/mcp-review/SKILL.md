---
name: mcp-review
description: Review MCP server specifications and updates for compliance, security, and quality. Use when evaluating server.json files, PRs adding/updating servers, or assessing MCP server changes.
allowed-tools: Read, Grep, Glob, Bash, WebFetch
---

# MCP Server Specification Review

You are an expert reviewer for the ToolHive Registry, a curated catalog of MCP (Model Context Protocol) servers. Your role is to evaluate server.json files and MCP server submissions for compliance, security, quality, and completeness.

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
| **Provenance** | Software provenance verification via Sigstore or GitHub Attestations | Check for `provenance` field in `_meta` extensions |
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

## server.json Review Process

### 1. Determine Server Type

| Type | Identifier | Valid Transports |
|------|------------|------------------|
| **Container** | `packages` field | `stdio`, `streamable-http`, `sse` |
| **Remote** | `remotes` field | `streamable-http`, `sse` (NOT `stdio`) |

### 2. Validate Top-Level Fields

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.stacklok/<server-name>",
  "description": "Clear one-line description",
  "title": "<server-name>",
  "repository": { "url": "https://github.com/org/repo", "source": "github" },
  "version": "1.0.0"
}
```

### 3. Validate Package/Remote Configuration

**Container servers:**
```json
{
  "packages": [{
    "registryType": "oci",
    "identifier": "ghcr.io/org/server:v1.0.0",
    "transport": { "type": "stdio" },
    "environmentVariables": [
      { "name": "API_KEY", "description": "...", "isRequired": true, "isSecret": true }
    ]
  }]
}
```

**Remote servers:**
```json
{
  "remotes": [{
    "type": "streamable-http",
    "url": "https://api.example.com/mcp"
  }]
}
```

### 4. Validate `_meta` Extensions

The `_meta` block must follow this nesting:
```json
{
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "<extension-key>": {
          "tier": "Community",
          "status": "Active",
          "tags": ["..."],
          "tools": ["..."]
        }
      }
    }
  }
}
```

**Critical:** The `<extension-key>` must exactly match:
- For containers: `packages[0].identifier` (e.g., `ghcr.io/org/server:v1.0.0`)
- For remotes: `remotes[0].url` (e.g., `https://api.example.com/mcp`)

### 5. Security Review

**CRITICAL - Must verify:**

1. **No filesystem paths** in `permissions` - Users configure mounts at runtime
2. **Secrets marked** with `isSecret: true` in `environmentVariables`
3. **Network scoped** - No `insecure_allow_all: true` unless justified
4. **Specific image tags** - Never use `latest`, always pin versions
5. **Extension key matches** - `_meta` key must match `identifier` or `url`

**Supply Chain Security (in `_meta` extensions):**
```json
{
  "provenance": {
    "cert_issuer": "https://token.actions.githubusercontent.com",
    "repository_uri": "https://github.com/org/repo",
    "runner_environment": "github-hosted",
    "signer_identity": "/.github/workflows/release.yml",
    "sigstore_url": "tuf-repo-cdn.sigstore.dev"
  }
}
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
| Top-level fields | Pass/Fail | |
| Package/Remote config | Pass/Fail | |
| Extension key match | Pass/Fail | |
| Transport valid | Pass/Fail | |
| Tools listed | Pass/Fail | |
| Security fields | Pass/Fail | |

### Security Review

- [ ] No filesystem paths in permissions
- [ ] Secrets properly marked (isSecret: true)
- [ ] Network permissions scoped
- [ ] Image tag pinned (not `latest`)
- [ ] Extension key matches identifier/URL
- [ ] Provenance configured (if applicable)

### Findings

**Issues (must fix):**
1. ...

**Suggestions (optional):**
1. ...

---

### Validation
Run `task catalog:validate` to verify spec compliance.
```

---

## Version Update Review

When reviewing updates to existing servers:

1. **Identify changes** - What fields changed?
2. **Check both locations** - Was the image tag updated in BOTH `identifier` AND `_meta` key?
3. **Check changelog** - What's new in this version?
4. **Verify tools** - Added or removed tools?
5. **Breaking changes** - Transport, env vars, or auth changes?
6. **Security implications** - New permissions or scopes?

Focus review on changed aspects, not full re-review.

---

## Workflow Commands

```bash
# Validate all entries
task catalog:validate

# Build registry and verify
task catalog:build

# Check specific entry in built output
jq '.servers["<name>"]' build/toolhive/registry.json
jq '.remote_servers["<name>"]' build/toolhive/registry.json
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