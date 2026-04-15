# Skill Inclusion Criteria

This document defines the criteria specific to skills in the ToolHive
registry. For shared criteria that apply to all registry entries (open source
requirements, acceptable licenses, community health), see the
[registry inclusion criteria](registry-criteria.md).

A skill is a reusable prompt or workflow that leverages MCP server tools to
perform a specific task -- such as reviewing pull requests, triaging issues,
or hardening CI pipelines. Skills bundle prompts, tool references, and
optional scripts into a package that AI assistants can invoke.

Because the skill ecosystem is still maturing, some criteria are intentionally
relaxed compared to MCP servers. This document explains where the criteria
align, where they differ, and the rationale behind each decision.

## Table of contents

- [MCP server dependencies](#mcp-server-dependencies)
- [Specification compliance](#specification-compliance)
- [Distribution and packaging](#distribution-and-packaging)
- [Versioning](#versioning)
- [Security and supply chain](#security-and-supply-chain)
- [Skill stability](#skill-stability)
- [Code quality](#code-quality)
- [Documentation](#documentation)
- [Release process](#release-process)
- [Community health and repository signals](#community-health-and-repository-signals)
- [Security requirements](#security-requirements)
- [Scoring](#scoring)

## MCP server dependencies

If a skill declares a dependency on one or more MCP servers (via
`allowedTools` or other configuration), every referenced MCP server **must**
already be included in the ToolHive catalog. This is a hard requirement and a
blocker for inclusion.

- The dependency check is performed at submission time.
- If the required MCP server is not yet in the catalog, the skill submission
  will be rejected until the server is added.
- This ensures that every skill in the catalog can be fully resolved and
  installed from the catalog itself, without requiring external or unvetted
  server dependencies.

## Specification compliance

- Skills **must** be fully compliant with the
  [agent skill specification](https://github.com/anthropics/agent-skill-spec).
- The `thv skills validate` command is used in CI to verify compliance
  automatically.
- Skills that fail validation are rejected.

## Distribution and packaging

The official ToolHive catalog requires OCI (Open Container Initiative) images
as the canonical distribution format for all skills.

- Skills **must** be published as OCI artifacts to be included in the catalog.
- Git references are supported as a secondary distribution mechanism, but OCI
  is the authoritative source for the catalog.
- OCI distribution provides reproducibility, versioning, and supply chain
  traceability that git references alone cannot guarantee.

## Versioning

- Skills **must** use a versioning scheme (semantic versioning preferred).
- OCI images provide built-in versioning through container tags, which the
  catalog leverages.
- Git-only references without a versioning mechanism are problematic for
  enterprise consumers and are not sufficient on their own.
- Versions are used as container tags when the catalog CI builds OCI images.

## Security and supply chain

Security and CI requirements are **relaxed** for skills compared to MCP
servers. The following items are recommended rather than required:

- Software provenance verification (Sigstore, GitHub Attestations).
- SLSA compliance level assessment.
- Pinned dependencies and GitHub Actions pinned to SHAs.
- Published Software Bill of Materials (SBOM).
- Automated security scanning.
- CVE monitoring.

**Rationale**: The skill ecosystem is still emerging. Very few skill authors
currently publish SBOMs, perform CVE monitoring, or follow SLSA compliance
practices for skills. Enforcing these requirements today would exclude nearly
all skill submissions and slow adoption. As the ecosystem matures and tooling
improves, these requirements will be revisited and tightened.

Skills can ship scripts (executable software), which is one of the motivations
for requiring OCI distribution -- it provides a clearer supply chain boundary
than raw git references.

## Skill stability

Skill stability is evaluated differently from MCP server tool stability.
Skills are more free-form and do not expose the same kind of typed tool
signatures that MCP servers do, making stability harder to measure
programmatically.

The primary concern is **skill shadowing** -- a new version of a skill that
retains the same name and description but behaves in a fundamentally different
or malicious way. Reviewers should watch for:

- **Name/description consistency** -- a skill update that keeps the same
  identity but changes its behavior substantially should be flagged for
  closer review.
- **Scope changes** -- new versions that request additional MCP server
  dependencies or new tool permissions beyond the original scope.
- **Behavioral drift** -- prompt changes that alter the skill's purpose in
  ways not reflected by its metadata.

## Code quality

- Validation against the agent skill specification via `thv skills validate`.
- Skill quality metrics are still being developed; both ToolHive and Anthropic
  have validation tooling that is expected to converge.
- Reviewers should evaluate the clarity and correctness of skill prompts,
  proper use of tool references, and overall coherence of the skill's purpose.

## Documentation

- Clear README or description explaining what the skill does and how to use
  it.
- Documentation of required MCP server dependencies.
- Explanation of any scripts or executable components included in the skill.
- Documentation kept up to date with releases.

## Release process

- Git tags or another versioning mechanism are required.
- Versions are used as container tags when the catalog CI builds OCI images.
- Regular release cadence (no releases or meaningful commits in the last
  6 months is a yellow flag).
- Semantic versioning compliance is preferred.

## Community health and repository signals

Community health criteria for skills follow the same general framework as
all registry entries (see
[community health](registry-criteria.md#community-health)), with additional
emphasis on repository-level signals that are particularly important for
evaluating skills:

- **Repository activity** -- regular commits and recent activity are strong
  positive signals.
- **Developer activity** -- active issue triage, PR reviews, and responsive
  maintainers.
- **Author reputation** -- skills from established organizations or known
  maintainers carry more weight than submissions from newly created accounts.
  A GitHub account created the same day as the submission is a red flag.
- **Stars and community adoption** -- while not a hard requirement, these
  indicators help gauge community trust.

## Security requirements

Skills share the same security principles as MCP servers for authentication,
data protection, and security practices, but the enforcement level differs.

### Authentication and authorization

- Skills that interact with authenticated services should document the
  required credentials clearly.
- API keys and tokens referenced by the skill must be marked as secret.

### Data protection

- Skills should not embed or expose secrets, tokens, or credentials in their
  prompts or scripts.

### Security practices

- Security issue reporting mechanisms (`SECURITY.md` or equivalent) are
  recommended.
- No known unpatched critical or high CVEs in any scripts or dependencies
  shipped by the skill.

## Scoring

The following table maps specific requirements to their severity levels for
skills. Key differences from MCP servers are noted in the rightmost column.
See the [evaluation framework](registry-criteria.md#evaluation-framework)
for an explanation of the severity levels.

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
