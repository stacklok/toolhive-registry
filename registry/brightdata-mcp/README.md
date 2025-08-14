# brightdata-mcp

An MCP interface into the Bright Data toolset for web scraping and data extraction

## Basic Information

- **Image:** `ghcr.io/stacklok/dockyard/npx/brightdata-mcp:2.4.2`
- **Repository:** [https://github.com/brightdata/brightdata-mcp](https://github.com/brightdata/brightdata-mcp)
- **Tier:** Community
- **Status:** Active
- **Transport:** stdio

## Available Tools

This server provides 58 tools:

- `search_engine` | - `scrape_as_markdown` | - `scrape_as_html`
- `extract` | - `session_stats` | - `web_data_amazon_product`
- `web_data_amazon_product_reviews` | - `web_data_amazon_product_search` | - `web_data_walmart_product`
- `web_data_walmart_seller` | - `web_data_ebay_product` | - `web_data_homedepot_products`
- `web_data_zara_products` | - `web_data_etsy_products` | - `web_data_bestbuy_products`
- `web_data_linkedin_person_profile` | - `web_data_linkedin_company_profile` | - `web_data_linkedin_job_listings`
- `web_data_linkedin_posts` | - `web_data_linkedin_people_search` | - `web_data_crunchbase_company`
- `web_data_zoominfo_company_profile` | - `web_data_instagram_profiles` | - `web_data_instagram_posts`
- `web_data_instagram_reels` | - `web_data_instagram_comments` | - `web_data_facebook_posts`
- `web_data_facebook_marketplace_listings` | - `web_data_facebook_company_reviews` | - `web_data_facebook_events`
- `web_data_tiktok_profiles` | - `web_data_tiktok_posts` | - `web_data_tiktok_shop`
- `web_data_tiktok_comments` | - `web_data_google_maps_reviews` | - `web_data_google_shopping`
- `web_data_google_play_store` | - `web_data_apple_app_store` | - `web_data_reuter_news`
- `web_data_github_repository_file` | - `web_data_yahoo_finance_business` | - `web_data_x_posts`
- `web_data_zillow_properties_listing` | - `web_data_booking_hotel_listings` | - `web_data_youtube_profiles`
- `web_data_youtube_comments` | - `web_data_youtube_videos` | - `web_data_reddit_posts`
- `browser_create_session` | - `browser_navigate` | - `browser_click`
- `browser_type` | - `browser_scroll` | - `browser_screenshot`
- `browser_get_page_content` | - `browser_wait_for_element` | - `browser_execute_script`
- `browser_close_session`

## Environment Variables

### Required

- **API_TOKEN** 🔒: Bright Data API token for authentication

### Optional

- **RATE_LIMIT**: Rate limiting configuration (format: limit/time+unit, e.g., 100/1h, 50/30m, 10/5s)
- **WEB_UNLOCKER_ZONE**: Custom Web Unlocker zone name
  - Default: `mcp_unlocker`
- **BROWSER_ZONE**: Custom Browser API zone name
  - Default: `mcp_browser`
- **PRO_MODE**: Enable pro mode to access all tools including browser automation and web data extraction
  - Default: `false`

## Tags

`web-scraping` `data-extraction` `api` `automation` `browser` 

## Statistics

- ⭐ Stars: 1065
- 📦 Pulls: 65
- 🕐 Last Updated: 2025-08-13T08:42:29Z
