# buildkite

Connect your Buildkite data (pipelines, builds, jobs, tests) to AI tooling and editors.

## Basic Information

- **Image:** `ghcr.io/buildkite/buildkite-mcp-server:0.5.8`
- **Repository:** [https://github.com/buildkite/buildkite-mcp-server](https://github.com/buildkite/buildkite-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 24 tools:

- `get_cluster` | - `list_clusters` | - `get_cluster_queue`
- `list_cluster_queues` | - `get_pipeline` | - `list_pipelines`
- `create_pipeline` | - `update_pipeline` | - `list_builds`
- `get_build` | - `create_build` | - `get_build_test_engine_runs`
- `get_jobs` | - `get_job_logs` | - `list_artifacts`
- `get_artifact` | - `list_annotations` | - `list_test_runs`
- `get_test_run` | - `get_failed_executions` | - `get_test`
- `access_token` | - `current_user` | - `user_token_organization`

## Environment Variables

### Required

- **BUILDKITE_API_TOKEN** 🔒: Your Buildkite API access token

### Optional

- **JOB_LOG_TOKEN_THRESHOLD**: Token threshold for job logs. If exceeded, logs will be written to disk and returned by path (for local use only).

## Tags

`buildkite` `continuous-integration` `continuous-delivery` `pipelines` `builds` `jobs` `devops` `testing` 

## Statistics

- ⭐ Stars: 24
- 📦 Pulls: 3103
- 🕐 Last Updated: 2025-08-11T00:24:58Z
