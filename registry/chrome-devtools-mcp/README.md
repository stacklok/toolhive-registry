# Chrome DevTools MCP Server

[![npm chrome-devtools-mcp package](https://img.shields.io/npm/v/chrome-devtools-mcp.svg)](https://npmjs.org/package/chrome-devtools-mcp)

`chrome-devtools-mcp` lets your coding agent control and inspect a live Chrome browser. It acts as a Model Context Protocol (MCP) server, giving your AI coding assistant access to the full power of Chrome DevTools for reliable automation, in-depth debugging, and performance analysis.

## Key Features

- **Performance insights**: Uses Chrome DevTools to record traces and extract actionable performance insights
- **Advanced browser debugging**: Analyze network requests, take screenshots and check the browser console
- **Reliable automation**: Uses Puppeteer to automate actions in Chrome and automatically wait for action results

## Requirements

- Node.js v20.19 or newer (latest maintenance LTS version)
- Chrome current stable version or newer
- npm

## Configuration Options

The server supports various configuration options that can be passed via the `args` property:

### Browser Connection

- `--browserUrl`, `-u`: Connect to a running Chrome instance using port forwarding
- `--wsEndpoint`, `-w`: WebSocket endpoint to connect to a running Chrome instance
- `--wsHeaders`: Custom headers for WebSocket connection in JSON format

### Browser Options

- `--headless`: Run in headless (no UI) mode (default: false)
- `--executablePath`, `-e`: Path to custom Chrome executable
- `--isolated`: Creates a temporary user-data-dir that is automatically cleaned up (default: false)
- `--channel`: Specify Chrome channel: `stable`, `canary`, `beta`, or `dev` (default: stable)
- `--viewport`: Initial viewport size (e.g., `1280x720`)
- `--proxyServer`: Proxy server configuration for Chrome
- `--acceptInsecureCerts`: Ignore errors for self-signed and expired certificates
- `--chromeArg`: Additional arguments for Chrome

### Tool Categories

- `--categoryEmulation`: Enable/disable emulation tools (default: true)
- `--categoryPerformance`: Enable/disable performance tools (default: true)
- `--categoryNetwork`: Enable/disable network tools (default: true)

### Debugging

- `--logFile`: Path to a file to write debug logs to

## Security Considerations

⚠️ **Important Security Notes:**

1. `chrome-devtools-mcp` exposes content of the browser instance to MCP clients, allowing them to inspect, debug, and modify any data in the browser or DevTools
2. When using `--browser-url` to connect to a running Chrome instance, the remote debugging port allows any application on your machine to control the browser
3. Chrome requires a non-default user data directory when enabling remote debugging to protect your regular browsing profile
4. Avoid sharing sensitive or personal information that you don't want to share with MCP clients

## User Data Directory

The server starts Chrome using the following user data directory:

- **Linux/macOS**: `$HOME/.cache/chrome-devtools-mcp/chrome-profile-$CHANNEL`
- **Windows**: `%HOMEPATH%/.cache/chrome-devtools-mcp/chrome-profile-$CHANNEL`

The directory is shared across all instances. Use `--isolated=true` for a temporary directory that's cleaned up automatically.

## Tools

The server provides 26 tools organized into categories:

### Input Automation (8 tools)

- `click`, `drag`, `fill`, `fill_form`, `handle_dialog`, `hover`, `press_key`, `upload_file`

### Navigation Automation (6 tools)

- `close_page`, `list_pages`, `navigate_page`, `new_page`, `select_page`, `wait_for`

### Emulation (2 tools)

- `emulate`, `resize_page`

### Performance (3 tools)

- `performance_analyze_insight`, `performance_start_trace`, `performance_stop_trace`

### Network (2 tools)

- `get_network_request`, `list_network_requests`

### Debugging (5 tools)

- `evaluate_script`, `get_console_message`, `list_console_messages`, `take_screenshot`, `take_snapshot`

## Known Limitations

### Operating System Sandboxes

Some MCP clients allow sandboxing the MCP server using macOS Seatbelt or Linux containers. If sandboxes are enabled, `chrome-devtools-mcp` cannot start Chrome (which requires permissions to create its own sandboxes).

**Workarounds:**

- Disable sandboxing for `chrome-devtools-mcp` in your MCP client
- Use `--browser-url` to connect to a Chrome instance started manually outside the MCP client sandbox

## Documentation

- [Tool Reference](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/tool-reference.md)
- [Troubleshooting](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/troubleshooting.md)
- [Changelog](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/CHANGELOG.md)
- [Contributing](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/CONTRIBUTING.md)

## Repository

[https://github.com/ChromeDevTools/chrome-devtools-mcp](https://github.com/ChromeDevTools/chrome-devtools-mcp)
