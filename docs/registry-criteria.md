# Registry Inclusion Criteria

The ToolHive registry is a curated list of MCP servers and skills that meet
specific criteria. We aim to establish a community-auditable list of
high-quality entries through clear, observable, and objective criteria.

These criteria ensure that entries in the registry meet the standards of
security, quality, and usability that ToolHive aims to uphold.

The registry accepts two types of entries:

- **MCP servers** -- containerized or remote services that expose tools and
  resources via the Model Context Protocol (MCP). See the
  [MCP server criteria](server-criteria.md) for server-specific requirements.
- **Skills** -- reusable prompts and workflows that leverage MCP server tools
  to perform specific tasks. See the
  [skill criteria](skill-criteria.md) for skill-specific requirements.

This document covers the shared requirements and evaluation framework that
apply to both entry types. The linked documents cover criteria specific to
each type.

## Table of contents

- [Contribute to the registry](#contribute-to-the-registry)
- [Key terms](#key-terms)
- [Shared criteria](#shared-criteria)
  - [Open source requirements](#open-source-requirements)
  - [Acceptable licenses](#acceptable-licenses)
  - [Proprietary service references](#proprietary-service-references)
  - [Community health](#community-health)
- [Evaluation framework](#evaluation-framework)
  - [Scoring system](#scoring-system)
  - [Tiered classifications](#tiered-classifications)

## Contribute to the registry

If you have an MCP server or skill that you'd like to add to the ToolHive
registry, you can
[open an issue](https://github.com/stacklok/toolhive-catalog/issues/new)
or submit a pull request. The ToolHive team will review your submission and
consider adding it to the registry.

For instructions on how to structure your submission, see the
[README](../README.md). Automated validation is available:

- **MCP servers**: `task catalog:validate` checks server definitions.
- **Skills**: `task catalog:validate` checks skill definitions, and
  `thv skills validate` verifies compliance with the agent skill
  specification.

## Key terms

| Term | Definition |
|------|------------|
| **MCP** | Model Context Protocol -- an open standard for connecting AI assistants to external tools and data sources. |
| **OCI** | Open Container Initiative -- a set of standards for container image formats and distribution. Used as the canonical packaging format for skills in the catalog. |
| **SLSA** | Supply chain Levels for Software Artifacts -- a framework for ensuring the integrity of software artifacts throughout the supply chain. |
| **SBOM** | Software Bill of Materials -- a formal record of the components and dependencies in a piece of software. |
| **Sigstore** | An open source project for signing, verifying, and protecting software supply chains. |
| **Skill shadowing** | When a new version of a skill retains the same name and description but changes its behavior substantially, potentially in a malicious way. |

## Shared criteria

The following criteria apply to both MCP servers and skills.

### Open source requirements

- Must be fully open source with no exceptions.
- Source code must be publicly accessible on a public hosting platform (GitHub,
  GitLab, etc.).
- Must use an [acceptable open source license](#acceptable-licenses).

### Acceptable licenses

**Accepted** (permissive):

Licenses such as `Apache-2.0`, `MIT`, `BSD-2-Clause`, and `BSD-3-Clause` allow
maximum flexibility for integration, modification, and redistribution with
minimal restrictions, making entries accessible across all project types and
commercial applications.

**Excluded** (copyleft / restrictive):

We exclude copyleft and restrictive licenses such as `AGPL-3.0`, `GPL-2.0`,
`GPL-3.0`, and `LGPL-*` to ensure entries can be freely integrated into
various commercial and open source projects without legal complications or
viral licensing requirements.

If the license is unclear or missing, we will request clarification from the
submitter.

### Proprietary service references

Entries (both MCP servers and skills) may reference or integrate with
proprietary services (e.g., GitHub, Stripe, Jira) provided the entry itself
is published by the official vendor of that service. For example, a GitHub MCP
server published by GitHub is acceptable, even though GitHub itself is not
open source.

### Community health

#### Responsiveness

- **Issue response time**: issues open for more than 3 weeks without any
  response is a red flag.
- Active bug triage and resolution rate.
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

**Yellow flags**: no releases or meaningful commits in the last 6 months, many
unanswered issues, missing tests, incomplete documentation, single maintainer
with no org backing.

**Red flags**: copyleft or restricted license (see
[Acceptable licenses](#acceptable-licenses)), no source code, known unpatched
CVEs, abandoned project, no clear maintainer.

## Evaluation framework

### Scoring system

The registry uses four severity levels to categorize criteria:

| Level | Description | Impact |
|-------|-------------|--------|
| **Required** | Essential attributes | Significant penalty if missing |
| **Expected** | Typical of well-executed projects | Moderate score impact if absent |
| **Recommended** | Good practice indicators | Positive contribution to evaluation |
| **Bonus** | Demonstrates excellence | No penalty for absence |

For the specific scoring tables, see:

- [MCP server scoring](server-criteria.md#scoring)
- [Skill scoring](skill-criteria.md#scoring)

### Tiered classifications

- **Official** -- maintained by the ToolHive team, the MCP specification
  authors, or the platform owners of the integrated service.
- **Community** -- created and maintained by third-party contributors.

Both tiers are subject to the same criteria. Community health signals (stars,
active maintainers, community adoption) are evaluated during review and during
periodic re-evaluation to ensure entries continue to meet the criteria over
time.
