# Registry Inclusion Criteria

The ToolHive registry is a curated list of MCP servers that meet specific
criteria. We aim to establish a community-auditable list of high-quality MCP
servers through clear, observable, and objective criteria.

These criteria ensure that the servers in the registry meet the standards of
security, quality, and usability that ToolHive aims to uphold.

## Table of contents

- [Contribute to the registry](#contribute-to-the-registry)
- [Criteria for MCP servers](#criteria-for-mcp-servers)
  - [Open source requirements](#open-source-requirements)
  - [Acceptable licenses](#acceptable-licenses)
  - [Security](#security)
  - [Continuous integration](#continuous-integration)
  - [API compliance](#api-compliance)
  - [Tool stability](#tool-stability)
  - [Code quality](#code-quality)
  - [Documentation](#documentation)
  - [Release process](#release-process)
  - [Community health](#community-health)
  - [Security requirements](#security-requirements)
- [Evaluation framework](#evaluation-framework)
  - [Scoring system](#scoring-system)
  - [Tiered classifications](#tiered-classifications)

## Contribute to the registry

If you have an MCP server that you'd like to add to the ToolHive registry, you
can [open an issue](https://github.com/stacklok/toolhive-catalog/issues/new?template=add-an-mcp-server.md)
or submit a pull request. The ToolHive team will review your submission and
consider adding it to the registry.

For instructions on how to structure your submission, see the
[README](../README.md#how-to-add-your-mcp-server).

## Criteria for MCP servers

### Open source requirements

- Must be fully open source with no exceptions.
- Source code must be publicly accessible on a public hosting platform (GitHub,
  GitLab, etc.).
- Must use an [acceptable open source license](#acceptable-licenses).

### Acceptable licenses

**Accepted** (permissive):

Licenses such as `Apache-2.0`, `MIT`, `BSD-2-Clause`, and `BSD-3-Clause` allow
maximum flexibility for integration, modification, and redistribution with
minimal restrictions, making MCP servers accessible across all project types and
commercial applications.

**Excluded** (copyleft / restrictive):

We exclude copyleft and restrictive licenses such as `AGPL-3.0`, `GPL-2.0`,
`GPL-3.0`, and `LGPL-*` to ensure MCP servers can be freely integrated into
various commercial and open source projects without legal complications or viral
licensing requirements.

If the license is unclear or missing, we will request clarification from the
submitter.

### Security

- Software provenance verification (Sigstore, GitHub Attestations).
- SLSA compliance level assessment.
- Pinned dependencies and GitHub Actions pinned to SHAs.
- Published Software Bill of Materials (SBOMs).

### Continuous integration

- Automated dependency updates (Dependabot, Renovate, etc.).
- Automated security scanning.
- CVE monitoring.
- Code linting and quality checks.

### API compliance

- Full MCP API specification support.
- Implementation of all required endpoints (tools, resources, etc.).
- Protocol version compatibility with the current MCP spec.
- Transport type appropriate for the deployment model.

### Tool stability

- **Version consistency** -- semantic versioning (`vX.Y.Z` tags).
- **Breaking change frequency** -- low frequency of breaking changes.
- **Backward compatibility** -- maintained across minor versions.
- **Tool signatures** -- tools don't disappear or change signatures
  unexpectedly.

### Code quality

- Presence of automated tests with measurable coverage.
- Quality CI/CD implementation.
- Code linting and quality checks in CI.
- Code review practices (branch protection, required reviews).

### Documentation

- Clear README with setup and install instructions.
- API and tool documentation (what each tool does, inputs, outputs).
- Deployment and operation guides.
- Documentation kept up to date with releases.

### Release process

- CI-based release automation (not manual artifact uploads).
- Regular release cadence (not stale for more than 6 months).
- Semantic versioning compliance.
- Maintained changelog (`CHANGELOG.md` or GitHub Releases with notes).

### Community health

#### Responsiveness

- **Issue response time**: issues open 3-4 weeks without any response is a red
  flag.
- Active bug resolution and resolution rate.
- PR review turnaround.
- User support quality.

#### Community strength

- Project backing (individual vs. organizational).
- Number of active maintainers and contributor diversity.
- Corporate or foundation support.
- Governance model maturity (`CODEOWNERS`, `CONTRIBUTING.md`, clear
  decision-making).

**Green flags**: regular commits, active issue triage, multiple contributors,
good test coverage, clear documentation, organizational backing.

**Yellow flags**: no activity for more than 6 months, many unanswered issues,
missing tests, incomplete documentation, single maintainer with no org backing.

**Red flags**: copyleft license, no source code, known unpatched CVEs, abandoned
project, no clear maintainer.

### Security requirements

#### Authentication and authorization

- Secure authentication mechanisms.
- Proper authorization controls.
- Standard security protocol support (OAuth, TLS).
- API keys and tokens marked as secret.

#### Data protection

- Encryption for data in transit (TLS for all external communication).
- Proper sensitive information handling (secrets, tokens, and credentials
  protected).

#### Security practices

- Security issue reporting mechanisms (`SECURITY.md` or equivalent).
- Clear incident response channels.
- No known unpatched critical or high CVEs.

## Evaluation framework

### Scoring system

The registry uses four severity levels to categorize criteria:

| Level | Description | Impact |
|-------|-------------|--------|
| **Required** | Essential attributes | Significant penalty if missing |
| **Expected** | Typical of well-executed projects | Moderate score impact if absent |
| **Recommended** | Good practice indicators | Positive contribution to evaluation |
| **Bonus** | Demonstrates excellence | No penalty for absence |

The following table maps specific requirements to their severity levels:

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

### Tiered classifications

- **Official** -- maintained by the MCP team or platform owners.
- **Community** -- created and maintained by the community.

Minimum threshold requirements (stars, maintainers, community indicators) and
regular re-evaluation ensure entries continue to meet the criteria over time.
