# Skill Inclusion Criteria

Source: [docs/skill-criteria.md](../../../../docs/skill-criteria.md)

For shared criteria (open source, licensing, community health), see
[registry-criteria.md](registry-criteria.md).

## MCP Server Dependencies

Every MCP server referenced by the skill (via `allowedTools` or other
configuration) **must** already exist in the ToolHive catalog. This is a hard
requirement and a blocker for inclusion.

- Checked at submission time.
- If the required server is not yet in catalog, the skill is rejected until the
  server is added.

## Specification Compliance

- Skills **must** comply with the
  [agent skill specification](https://github.com/anthropics/agent-skill-spec).
- `thv skills validate` verifies compliance in CI.
- Skills that fail validation are rejected.

## Distribution and Packaging

- Skills **must** be published as OCI artifacts (canonical format).
- Git references are a secondary mechanism; OCI is authoritative.
- OCI provides reproducibility, versioning, and supply chain traceability.

## Versioning

- Skills **must** use a versioning scheme (semantic versioning preferred).
- OCI image tags provide versioning; git-only references without versioning are
  insufficient.

## Security and Supply Chain

These items are **Recommended** for skills (relaxed from MCP servers):

- Software provenance (Sigstore, GitHub Attestations)
- SLSA compliance
- Pinned dependencies / Actions pinned to SHAs
- Published SBOM
- Automated security scanning
- CVE monitoring

**Rationale**: The skill ecosystem is still maturing. Enforcing full
server-level requirements would exclude nearly all current skill submissions.

## Skill Stability

Primary concern: **skill shadowing** -- a new version retains the same
name/description but behaves differently or maliciously.

Watch for:

- **Name/description consistency** -- same identity, substantially changed
  behavior.
- **Scope changes** -- new versions requesting additional MCP server
  dependencies or tool permissions.
- **Behavioral drift** -- prompt changes that alter purpose without updating
  metadata.

## Code Quality

- Validated via `thv skills validate`.
- Reviewers evaluate clarity/correctness of prompts, proper tool references,
  and overall coherence.

## Documentation

- Clear README or description explaining purpose and usage.
- Documentation of required MCP server dependencies.
- Explanation of any scripts or executable components.
- Kept up to date with releases.

## Community Health

Same general framework as all registry entries, with emphasis on:

- **Repository activity** -- regular commits, recent activity.
- **Author reputation** -- established orgs carry more weight; same-day account
  creation is a red flag.
- **Stars and adoption** -- not a hard requirement but helpful signal.

## Security Requirements

### Authentication and Authorization
- Skills interacting with authenticated services must document required
  credentials clearly.
- API keys and tokens must be marked as secret.

### Data Protection
- No embedded secrets, tokens, or credentials in prompts or scripts.

### Security Practices
- `SECURITY.md` or equivalent recommended.
- No known unpatched critical/high CVEs in scripts or dependencies.

## Scoring Table

| Requirement | Severity | Difference from MCP servers |
|-------------|----------|-----------------------------|
| Open source with public source code | Required | -- |
| Permissive license | Required | -- |
| Agent skill specification compliance | Required | Skills only |
| MCP server dependencies in catalog | Required | Skills only |
| OCI distribution | Required | Skills only |
| Versioning (tags or semantic versioning) | Required | Skills only |
| No known unpatched critical/high CVEs | Required | -- |
| Secure auth mechanisms (API keys marked secret) | Required | -- |
| Sensitive information handling | Required | -- |
| Repository activity and author reputation | Expected | Higher importance for skills |
| Documentation of dependencies and usage | Expected | -- |
| Pinned dependencies / Actions pinned to SHAs | Recommended | Required for MCP servers |
| Software provenance (Sigstore / GitHub Attestations) | Recommended | Expected for MCP servers |
| Automated security scanning in CI | Recommended | Expected for MCP servers |
| CVE monitoring | Recommended | Implicit in MCP server CI requirements |
| SLSA compliance | Recommended | -- |
| Published SBOM | Recommended | -- |
| Security reporting (`SECURITY.md`) | Recommended | -- |
| Skill quality metrics and validation | Bonus | Skills only |
