# hass-mcp

Integrates Home Assistant with LLM applications, enabling direct interaction with smart home devices, sensors, and automations.

## Basic Information

- **Image:** `docker.io/voska/hass-mcp:0.1.1`
- **Repository:** [https://github.com/voska/hass-mcp](https://github.com/voska/hass-mcp)
- **Tier:** Community
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 11 tools:

- `get_version` | - `get_entity` | - `entity_action`
- `list_entities` | - `search_entities_tool` | - `domain_summary_tool`
- `list_automations` | - `call_service_tool` | - `restart_ha`
- `get_history` | - `get_error_log`

## Environment Variables

### Required

- **HA_URL**: Home Assistant instance URL (e.g. http://homeassistant.local:8123)
- **HA_TOKEN** 🔒: Home Assistant Long-Lived Access Token

## Tags

`home-assistant` `smart-home` `automation` `iot` `sensors` `devices` `control` `monitoring` `home-automation` `domotics` 

## Statistics

- ⭐ Stars: 151
- 📦 Pulls: 17082
- 🕐 Last Updated: 2025-08-13T08:42:47Z
