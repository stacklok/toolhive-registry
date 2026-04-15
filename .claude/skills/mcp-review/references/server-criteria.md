# MCP Server Inclusion Criteria

Source: [docs/server-criteria.md](../../../../docs/server-criteria.md)

For shared criteria (open source, licensing, community health), see
[registry-criteria.md](registry-criteria.md).

## Security

| Requirement | Severity |
|-------------|----------|
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

## Continuous Integration

- Automated dependency updates (Dependabot, Renovate, etc.)
- Automated security scanning
- CVE monitoring
- Code linting and quality checks

## API Compliance

- Full MCP API specification support
- All required endpoints implemented (tools, resources, prompts as applicable)
- Protocol version compatibility with current MCP spec
- Transport type appropriate for the deployment model

## Tool Stability

- **Version consistency** -- semantic versioning (`vX.Y.Z` tags)
- **Breaking change frequency** -- low frequency of breaking changes
- **Backward compatibility** -- maintained across minor versions
- **Tool signatures** -- tools don't disappear or change signatures unexpectedly

## Code Quality

- Automated tests with measurable coverage
- Code linting and quality checks in CI/CD
- Code review practices (branch protection, required reviews)
- CI pipeline present and passing

## Documentation

- Clear README with setup/install instructions
- API/tool documentation (what each tool does, inputs, outputs)
- Deployment and operation guides
- Documentation kept up to date with releases

## Release Process

- CI-based release automation (not manual artifact uploads)
- Regular release cadence (no releases or meaningful commits in the last
  6 months is a yellow flag)
- Semantic versioning compliance
- Maintained changelog (`CHANGELOG.md` or GitHub Releases with notes)

## Verification Guide

Use whichever tools are available -- `gh` CLI, GitHub MCP tools, or WebFetch.

| Check | What to Look For | Where |
|-------|------------------|-------|
| License | SPDX identifier matches accepted list | Repo license endpoint, or `LICENSE` file |
| Dependency automation | Dependabot OR Renovate configured | `.github/dependabot.yml`, `renovate.json`, `.renovaterc`, `.renovaterc.json`, `.github/renovate.json` |
| Security policy | Incident reporting process exists | `SECURITY.md` in repo root |
| CI/CD | Workflows exist and run | `.github/workflows/` directory; recent workflow runs |
| Recent activity | Commits within last 6 months | Commit history (last 5-10 commits) |
| Unanswered issues | Issues open >3 weeks with 0 comments | Open issues sorted by creation date |
| Releases | Semver tags, changelog present | Releases list, tag names |
| Provenance | Sigstore signatures or GitHub Attestations | Release artifacts, container image signatures |
