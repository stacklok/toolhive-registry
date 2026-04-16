# ToolHive Registry Inclusion Criteria (Shared)

Source: [docs/registry-criteria.md](../../../../docs/registry-criteria.md)

For skill-specific criteria, see [skill-criteria.md](skill-criteria.md).

## Table of Contents

- [Open Source Standards](#open-source-standards)
- [Licensing](#licensing)
- [Evaluation Framework](#evaluation-framework)
- [Security Standards](#security-standards)
- [Code Quality and Testing](#code-quality-and-testing)
- [Stability](#stability)
- [Documentation](#documentation)
- [Release Process](#release-process)
- [Community and Responsiveness](#community-and-responsiveness)
- [Verification How-Tos](#verification-how-tos)

---

## Open Source Standards

- Source code **must** be publicly available -- no exceptions
- Repository must be on a public hosting platform (GitHub, GitLab, etc.)

## Licensing

**Accepted** (permissive):
- `Apache-2.0`, `MIT`, `BSD-2-Clause`, `BSD-3-Clause`

**Rejected** (copyleft):
- `AGPL-3.0`, `GPL-2.0`, `GPL-3.0`, `LGPL-*`

If license is unclear or missing, request clarification from the submitter.

## Evaluation Framework

The registry uses four severity levels to categorize criteria:

- **Required** -- Essential attributes; significant penalty if missing
- **Expected** -- Typical of well-executed projects; moderate score impact if absent
- **Recommended** -- Good practice indicators; positive contribution to evaluation
- **Bonus** -- Demonstrates excellence; no penalty for absence

## Security Standards

| Requirement | Description | Severity |
|-------------|-------------|----------|
| Provenance | Sigstore or GitHub Attestations for release artifacts | Expected |
| SLSA compliance | Supply chain security level assessment | Recommended |
| Pinned dependencies | Lockfiles present; GitHub Actions pinned to SHAs | Required |
| SBOM | Software Bill of Materials published in releases | Recommended |
| Auth mechanisms | Secure auth/authz (OAuth, TLS, API keys marked secret) | Required |
| Sensitive info handling | Secrets, tokens, and credentials properly protected | Required |
| Encryption in transit | TLS for all external communication | Required |
| Automated security scanning | Dedicated scanning in CI (not just CVE monitoring) | Expected |
| Security reporting | SECURITY.md or equivalent incident response channel | Recommended |
| No known CVEs | Dependencies free of unpatched critical/high CVEs | Required |

## Code Quality and Testing

- Automated tests with measurable coverage
- Code linting and quality checks in CI/CD
- Code review practices (branch protection, required reviews)
- CI pipeline present and passing

## Stability

- **Version consistency** -- semantic versioning (vX.Y.Z tags)
- **Breaking change frequency** -- low frequency of breaking changes
- **Backward compatibility** -- maintained across minor versions

## Documentation

- Clear README with setup/install instructions
- API/tool documentation (what each tool does, inputs, outputs)
- Deployment and operation guides
- Documentation kept up to date with releases

## Release Process

- CI-based release automation (not manual artifact uploads)
- Regular release cadence (not stale for >6 months)
- Semantic versioning compliance
- Maintained changelog (CHANGELOG.md or GitHub Releases with notes)

## Community and Responsiveness

### Responsiveness
- **Issue response time**: Issues open **3-4 weeks without any response is a red flag**
- Active bug resolution and resolution rate
- PR review turnaround
- User support quality

### Community Strength
- Project backing -- individual vs. organizational
- Number of active maintainers and contributor diversity
- Corporate or foundation support
- Governance model maturity (CODEOWNERS, CONTRIBUTING.md, clear decision-making)

**Green flags**: Regular commits, active issue triage, multiple contributors, good test coverage, clear docs, organizational backing.

**Yellow flags**: No activity >6 months, many unanswered issues, missing tests, incomplete docs, single maintainer with no org backing.

**Red flags**: Copyleft license, no source code, known unpatched CVEs, abandoned project, no clear maintainer.

## Verification Guide

Use whichever tools are available -- `gh` CLI, GitHub MCP tools, or WebFetch.

| Check | What to Look For | Where |
|-------|------------------|-------|
| License | SPDX identifier matches accepted list | Repo license endpoint, or `LICENSE` file |
| Dependency automation | Dependabot OR Renovate configured | `.github/dependabot.yml`, `renovate.json`, `.renovaterc`, `.renovaterc.json`, `.github/renovate.json` |
| Security policy | Incident reporting process exists | `SECURITY.md` in repo root |
| CI/CD | Workflows exist and run | `.github/workflows/` directory; recent workflow runs |
| Recent activity | Commits within last 6 months | Commit history (last 5-10 commits) |
| Unanswered issues | Issues open >3-4 weeks with 0 comments | Open issues sorted by creation date |
| Releases | Semver tags, changelog present | Releases list, tag names |
| Provenance | Sigstore signatures or GitHub Attestations | Release artifacts, container image signatures |
