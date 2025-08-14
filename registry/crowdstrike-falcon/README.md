# crowdstrike-falcon

Connects AI agents with the CrowdStrike Falcon platform for intelligent security analysis, providing programmatic access to detections, incidents, behaviors, threat intelligence, hosts, vulnerabilities, and identity protection capabilities.

## Basic Information

- **Image:** `quay.io/crowdstrike/falcon-mcp:latest`
- **Repository:** [https://github.com/crowdstrike/falcon-mcp](https://github.com/crowdstrike/falcon-mcp)
- **Tier:** Official
- **Status:** Active
- **Transport:** streamable-http

## Available Tools

This server provides 19 tools:

- `falcon_check_connectivity` | - `falcon_get_available_modules` | - `falcon_search_detections`
- `falcon_get_detection_details` | - `falcon_show_crowd_score` | - `falcon_search_incidents`
- `falcon_get_incident_details` | - `falcon_search_behaviors` | - `falcon_get_behavior_details`
- `falcon_search_actors` | - `falcon_search_indicators` | - `falcon_search_reports`
- `falcon_search_hosts` | - `falcon_get_host_details` | - `falcon_search_vulnerabilities`
- `falcon_search_kubernetes_containers` | - `falcon_count_kubernetes_containers` | - `falcon_search_images_vulnerabilities`
- `idp_investigate_entity`

## Environment Variables

### Required

- **FALCON_CLIENT_ID** 🔒: CrowdStrike API client ID
- **FALCON_CLIENT_SECRET** 🔒: CrowdStrike API client secret
- **FALCON_BASE_URL**: CrowdStrike API base URL (e.g., https://api.crowdstrike.com, https://api.us-2.crowdstrike.com, https://api.eu-1.crowdstrike.com)

### Optional

- **FALCON_MCP_MODULES**: Comma-separated list of modules to enable (detections,incidents,intel,hosts,spotlight,cloud,idp). If not set, all modules are enabled.
- **FALCON_MCP_DEBUG**: Enable debug logging - true or false (default: false)

## Tags

`crowdstrike` `falcon` `security` `cybersecurity` `threat-intelligence` `detections` `incidents` `vulnerabilities` `endpoint-security` `threat-hunting` `incident-response` `malware-analysis` `identity-protection` `cloud-security` 

## Statistics

- ⭐ Stars: 36
- 📦 Pulls: 3771
- 🕐 Last Updated: 2025-08-13T08:42:49Z
