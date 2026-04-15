# MCP Server Inclusion Criteria

This document defines the criteria specific to MCP servers in the ToolHive
registry. For shared criteria that apply to all registry entries (open source
requirements, acceptable licenses, community health), see the
[registry inclusion criteria](registry-criteria.md).

An MCP server is a containerized or remote service that exposes tools and
resources to AI assistants via the Model Context Protocol (MCP).

## Table of contents

- [Security](#security)
- [Continuous integration](#continuous-integration)
- [API compliance](#api-compliance)
- [Tool stability](#tool-stability)
- [Code quality](#code-quality)
- [Documentation](#documentation)
- [Release process](#release-process)
- [Security requirements](#security-requirements)
- [Scoring](#scoring)

## Security

- **Required** -- Pinned dependencies and GitHub Actions pinned to SHAs.
- **Expected** -- Software provenance verification (Sigstore, GitHub
  Attestations).
- **Recommended** -- SLSA compliance level assessment.
- **Recommended** -- Published Software Bill of Materials (SBOM).

## Continuous integration

- Automated dependency updates (Dependabot, Renovate, etc.).
- Automated security scanning.
- CVE monitoring.
- Code linting and quality checks.

## API compliance

- Full MCP API specification support.
- Implementation of all required endpoints (tools, resources, etc.).
- Protocol version compatibility with the current MCP spec.
- Transport type appropriate for the deployment model.

## Tool stability

- **Version consistency** -- semantic versioning (`vX.Y.Z` tags).
- **Breaking change frequency** -- low frequency of breaking changes.
- **Backward compatibility** -- maintained across minor versions.
- **Tool signatures** -- tools don't disappear or change signatures
  unexpectedly.

## Code quality

- Presence of automated tests with measurable coverage.
- Quality CI/CD implementation.
- Code linting and quality checks in CI.
- Code review practices (branch protection, required reviews).

## Documentation

- Clear README with setup and install instructions.
- API and tool documentation (what each tool does, inputs, outputs).
- Deployment and operation guides.
- Documentation kept up to date with releases.

## Release process

- CI-based release automation (not manual artifact uploads).
- Regular release cadence (no releases or meaningful commits in the last
  6 months is a yellow flag).
- Semantic versioning compliance.
- Maintained changelog (`CHANGELOG.md` or GitHub Releases with notes).

## Security requirements

### Authentication and authorization

- Secure authentication mechanisms.
- Proper authorization controls.
- Standard security protocol support (OAuth, TLS).
- API keys and tokens marked as secret.

### Data protection

- Encryption for data in transit (TLS for all external communication).
- Proper sensitive information handling (secrets, tokens, and credentials
  protected).

### Security practices

- Security issue reporting mechanisms (`SECURITY.md` or equivalent).
- Clear incident response channels.
- No known unpatched critical or high CVEs.

## Scoring

The following table maps specific requirements to their severity levels for
MCP servers. See the [evaluation framework](registry-criteria.md#evaluation-framework)
for an explanation of the severity levels.

| Requirement | Severity |
|-------------|----------|
| Open source with public source code | Required |
| Permissive license | Required |
| Pinned dependencies / Actions pinned to SHAs | Required |
| Secure auth mechanisms (OAuth, TLS, API keys marked secret) | Required |
| Sensitive information handling | Required |
| Encryption in transit (TLS) | Required |
| No known unpatched critical/high CVEs | Required |
| Software provenance (Sigstore / GitHub Attestations) | Expected |
| Automated security scanning in CI | Expected |
| SLSA compliance | Recommended |
| Published SBOM | Recommended |
| Security reporting (`SECURITY.md`) | Recommended |
