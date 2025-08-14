# firecrawl

A powerful web scraping and content extraction MCP server that provides advanced crawling, search, and structured data extraction capabilities with LLM integration.

## Basic Information

- **Image:** `docker.io/mcp/firecrawl:latest`
- **Repository:** [https://github.com/mendableai/firecrawl-mcp-server](https://github.com/mendableai/firecrawl-mcp-server)
- **Tier:** Official
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 10 tools:

- `firecrawl_scrape`
- `firecrawl_batch_scrape`
- `firecrawl_check_batch_status`
- `firecrawl_check_crawl_status`
- `firecrawl_search`
- `firecrawl_crawl`
- `firecrawl_extract`
- `firecrawl_deep_research`
- `firecrawl_generate_llmstxt`
- `firecrawl_map`

## Environment Variables

### Required

- **FIRECRAWL_API_KEY** 🔒: API key for FireCrawl service authentication

### Optional

- **FIRECRAWL_API_URL**: FireCrawl API URL (default: https://api.firecrawl.dev/v1)
- **FIRECRAWL_RETRY_MAX_ATTEMPTS**: Maximum number of retry attempts for API calls
- **FIRECRAWL_RETRY_INITIAL_DELAY**: Initial delay in milliseconds for retry backoff
- **FIRECRAWL_RETRY_MAX_DELAY**: Maximum delay in milliseconds for retry backoff
- **FIRECRAWL_RETRY_BACKOFF_FACTOR**: Backoff factor for retry delay calculation
- **FIRECRAWL_CREDIT_WARNING_THRESHOLD**: Credit threshold for warning notifications
- **FIRECRAWL_CREDIT_CRITICAL_THRESHOLD**: Credit threshold for critical notifications

## Tags

`web-crawler` `web-scraping` `data-collection` `batch-processing` `content-extraction` `search-api` `llm-tools` `javascript-rendering` `research` `automation` 

## Statistics

- ⭐ Stars: 4135
- 📦 Pulls: 12644
- 🕐 Last Updated: 2025-08-13T08:42:42Z
