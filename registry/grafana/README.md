# grafana

Provides access to your Grafana instance and the surrounding ecosystem, enabling dashboard search, datasource queries, alerting management, incident response, and Sift investigations.

## Basic Information

- **Image:** `docker.io/mcp/grafana:latest`
- **Repository:** [https://github.com/grafana/mcp-grafana](https://github.com/grafana/mcp-grafana)
- **Tier:** Official
- **Status:** Active
- **Transport:** sse

## Available Tools

This server provides 37 tools:

- `list_teams` | - `search_dashboards` | - `get_dashboard_by_uid`
- `update_dashboard` | - `get_dashboard_panel_queries` | - `list_datasources`
- `get_datasource_by_uid` | - `get_datasource_by_name` | - `query_prometheus`
- `list_prometheus_metric_metadata` | - `list_prometheus_metric_names` | - `list_prometheus_label_names`
- `list_prometheus_label_values` | - `list_incidents` | - `create_incident`
- `add_activity_to_incident` | - `resolve_incident` | - `query_loki_logs`
- `list_loki_label_names` | - `list_loki_label_values` | - `query_loki_stats`
- `list_alert_rules` | - `get_alert_rule_by_uid` | - `list_oncall_schedules`
- `get_oncall_shift` | - `get_current_oncall_users` | - `list_oncall_teams`
- `list_oncall_users` | - `get_investigation` | - `get_analysis`
- `list_investigations` | - `find_error_pattern_logs` | - `find_slow_requests`
- `list_pyroscope_label_names` | - `list_pyroscope_label_values` | - `list_pyroscope_profile_types`
- `fetch_pyroscope_profile`

## Environment Variables

### Required

- **GRAFANA_URL**: URL of the Grafana instance to connect to
- **GRAFANA_API_KEY** 🔒: Service account token with appropriate permissions

## Tags

`grafana` `dashboards` `visualization` `monitoring` `alerting` `prometheus` `loki` `tempo` `pyroscope` `incidents` `observability` `metrics` `logs` `traces` `sift` `investigations` `oncall` 

## Statistics

- ⭐ Stars: 1385
- 📦 Pulls: 8120
- 🕐 Last Updated: 2025-08-13T08:42:53Z
