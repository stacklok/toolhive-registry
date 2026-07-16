package thvclient

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	mcp "github.com/stacklok/toolhive-core/mcpcompat/mcp"

	"github.com/stacklok/toolhive-catalog/internal/serverjson"
)

// Client interacts with the ToolHive CLI (thv) to run MCP servers
// and discover their available tools.
type Client struct {
	thvPath string
	verbose bool
}

// NewClient creates a new ToolHive client. If thvPath is empty, it searches PATH.
func NewClient(thvPath string, verbose bool) (*Client, error) {
	if thvPath == "" {
		var err error
		thvPath, err = exec.LookPath("thv")
		if err != nil {
			return nil, fmt.Errorf("thv binary not found in PATH: %w", err)
		}
	}

	return &Client{
		thvPath: thvPath,
		verbose: verbose,
	}, nil
}

// RunServer starts a temporary MCP server from a server.json file.
// Returns the temporary server name for subsequent operations.
func (c *Client) RunServer(sf *serverjson.ServerFile, serverName string) (string, error) {
	if sf.IsRemoteServer() {
		return "", fmt.Errorf("remote servers cannot be run locally")
	}
	if !sf.IsPackageServer() {
		return "", fmt.Errorf("no packages found in server.json")
	}

	image := sf.ServerJSON.Packages[0].Identifier
	if image == "" {
		return "", fmt.Errorf("empty image identifier in server.json")
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		return "", fmt.Errorf("failed to get extensions: %w", err)
	}

	tempName := fmt.Sprintf("temp-%s-%d", serverName, time.Now().Unix())
	runArgs := BuildRunCommand(sf, ext, tempName, image)

	if c.verbose {
		fmt.Printf("Running command: thv %s\n", strings.Join(runArgs, " "))
	}

	// #nosec G204 - thvPath is validated in NewClient via exec.LookPath
	runCmd := exec.Command(c.thvPath, runArgs...)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"failed to start MCP server: %w\nOutput: %s",
			err, string(runOutput),
		)
	}

	// Give the server time to start
	time.Sleep(5 * time.Second)

	return tempName, nil
}

// ListTools queries a running MCP server for its tools.
func (c *Client) ListTools(serverName string) ([]string, error) {
	listArgs := NewCommandBuilder("mcp").
		AddPositional("list").
		AddPositional("tools").
		AddFlag("--server", serverName).
		AddFlag("--format", "json").
		Build()

	// #nosec G204 - thvPath is validated in NewClient via exec.LookPath
	listCmd := exec.Command(c.thvPath, listArgs...)
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"thv mcp list failed: %w\nOutput: %s",
			err, string(output),
		)
	}

	return ParseToolsJSON(string(output))
}

// ListToolDefinitions queries a running MCP server for full tool definitions.
// Returns nil (not an error) if the output is text-only.
func (c *Client) ListToolDefinitions(serverName string) ([]mcp.Tool, error) {
	listArgs := NewCommandBuilder("mcp").
		AddPositional("list").
		AddPositional("tools").
		AddFlag("--server", serverName).
		AddFlag("--format", "json").
		Build()

	// #nosec G204 - thvPath is validated in NewClient via exec.LookPath
	listCmd := exec.Command(c.thvPath, listArgs...)
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"thv mcp list failed: %w\nOutput: %s",
			err, string(output),
		)
	}

	return ParseToolDefinitions(string(output))
}

// Logs retrieves logs from a running MCP server.
func (c *Client) Logs(serverName string) (string, error) {
	logsArgs := NewCommandBuilder("logs").
		AddFlag("--follow", "false").
		AddPositional(serverName).
		Build()

	// #nosec G204 - thvPath is validated in NewClient via exec.LookPath
	logsCmd := exec.Command(c.thvPath, logsArgs...)
	output, err := logsCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"thv logs failed: %w\nOutput: %s",
			err, string(output),
		)
	}

	return string(output), nil
}

// StopServer stops a running MCP server.
func (c *Client) StopServer(serverName string) error {
	// #nosec G204 - thvPath is validated in NewClient via exec.LookPath
	stopCmd := exec.Command(c.thvPath, "stop", serverName)
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop server %s: %w", serverName, err)
	}
	return nil
}
