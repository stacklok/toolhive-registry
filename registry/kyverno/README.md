# kyverno

Provides Kyverno policy management capabilities enabling AI assistants to interact with Kyverno policies in Kubernetes clusters for comprehensive policy management, security assessment, and compliance monitoring.

## Basic Information

- **Image:** `ghcr.io/nirmata/kyverno-mcp:v0.2.2`
- **Repository:** [https://github.com/nirmata/kyverno-mcp](https://github.com/nirmata/kyverno-mcp)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 5 tools:

- `list_contexts`
- `switch_context`
- `apply_policies`
- `show_violations`
- `help`

## Environment Variables


### Optional

- **KUBECONFIG**: Path to the kubeconfig file for Kubernetes API authentication (mounted into the container with --volume)

## Tags

`kyverno` `kubernetes` `policy-management` `security` `compliance` `governance` `policy-as-code` `nirmata` `admission-control` `validation` `mutation` `generation` `policy-violations` `security-assessment` `rbac` `pod-security` `best-practices` 

## Statistics

- ⭐ Stars: 5
- 📦 Pulls: 4550
- 🕐 Last Updated: 2025-08-11T00:24:54Z
