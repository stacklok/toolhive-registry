# mcp-server-box

MCP server that integrates with the Box API to perform file operations, AI-based querying, metadata management, and document generation

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/uvx/mcp-server-box:0.1.2`
- **Repository:** [https://github.com/box-community/mcp-server-box](https://github.com/box-community/mcp-server-box)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 35 tools:

- `box_who_am_i` | - `box_authorize_app_tool` | - `box_search_tool`
- `box_search_folder_by_name_tool` | - `box_ai_ask_file_single_tool` | - `box_ai_ask_file_multi_tool`
- `box_ai_ask_hub_tool` | - `box_ai_extract_freeform_tool` | - `box_ai_extract_structured_using_fields_tool`
- `box_ai_extract_structured_using_template_tool` | - `box_ai_extract_structured_enhanced_using_fields_tool` | - `box_ai_extract_structured_enhanced_using_template_tool`
- `box_docgen_create_batch_tool` | - `box_docgen_get_job_by_id_tool` | - `box_docgen_list_jobs_tool`
- `box_docgen_list_jobs_by_batch_tool` | - `box_docgen_template_create_tool` | - `box_docgen_template_list_tool`
- `box_docgen_template_get_by_id_tool` | - `box_docgen_template_list_tags_tool` | - `box_docgen_template_list_jobs_tool`
- `box_docgen_template_get_by_name_tool` | - `box_docgen_create_single_file_from_user_input_tool` | - `box_read_tool`
- `box_upload_file_from_path_tool` | - `box_upload_file_from_content_tool` | - `box_download_file_tool`
- `box_list_folder_content_by_folder_id` | - `box_manage_folder_tool` | - `box_metadata_template_get_by_name_tool`
- `box_metadata_set_instance_on_file_tool` | - `box_metadata_get_instance_on_file_tool` | - `box_metadata_delete_instance_on_file_tool`
- `box_metadata_update_instance_on_file_tool` | - `box_metadata_template_create_tool`

## Environment Variables

### Required

- **BOX_CLIENT_ID**: Box API Client ID
- **BOX_CLIENT_SECRET** 🔒: Box API Client Secret

## Tags

`storage` `box` `files` `ai` `document-generation` 

## Statistics

- ⭐ Stars: 43
- 📦 Pulls: 52
- 🕐 Last Updated: 2025-08-13T08:42:34Z
