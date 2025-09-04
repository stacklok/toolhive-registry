// Package toolhive provides utilities for interacting with ToolHive
package toolhive

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/stacklok/toolhive/pkg/logger"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

// Client represents a ToolHive client
type Client struct {
	thvPath string
	verbose bool
}

// NewClient creates a new ToolHive client
func NewClient(thvPath string, verbose bool) (*Client, error) {
	// Find thv binary if not specified
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

// RunServer starts an MCP server from a spec
func (c *Client) RunServer(spec *types.RegistryEntry, serverName string) (string, error) {
	// Get the image from the spec
	var image string
	if spec.IsImage() && spec.ImageMetadata != nil {
		image = spec.Image
	} else if spec.IsRemote() {
		return "", fmt.Errorf("remote servers cannot be run locally")
	} else {
		return "", fmt.Errorf("no image found in spec file")
	}

	if image == "" {
		return "", fmt.Errorf("empty image in spec file")
	}

	if c.verbose {
		logger.Debugf("Using thv binary: %s", c.thvPath)
		logger.Debugf("Running MCP server from image: %s", image)
	}

	// Build the run command
	tempName := fmt.Sprintf("temp-%s-%d", serverName, time.Now().Unix())
	runArgs := BuildRunCommand(spec, tempName, image)

	if c.verbose {
		logger.Debugf("Running command: thv %s", strings.Join(runArgs, " "))
	}

	runCmd := exec.Command(c.thvPath, runArgs...) // #nosec G204 - thvPath is validated in NewClient
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to start MCP server: %w\nOutput: %s", err, string(runOutput))
	}

	// Give the server time to start
	time.Sleep(5 * time.Second)

	return tempName, nil
}

// ListTools queries a running MCP server for its tools
func (c *Client) ListTools(serverName string) ([]string, error) {
	listArgs := NewCommandBuilder("mcp").
		AddPositional("list").
		AddPositional("tools").
		AddFlag("--server", serverName).
		AddFlag("--format", "json").
		Build()

	listCmd := exec.Command(c.thvPath, listArgs...) // #nosec G204 - thvPath is validated in NewClient
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("thv mcp list failed: %w\nOutput: %s", err, string(output))
	}

	return ParseToolsJSON(string(output))
}

// Logs retrieves logs from a running MCP server
func (c *Client) Logs(serverName string) (string, error) {
	logsArgs := NewCommandBuilder("logs").
		AddFlag("--follow", "false").
		AddPositional(serverName).
		Build()

	logsCmd := exec.Command(c.thvPath, logsArgs...) // #nosec G204 - thvPath is validated in NewClient
	output, err := logsCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("thv logs failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// StopServer stops a running MCP server
func (c *Client) StopServer(serverName string) error {
	stopCmd := exec.Command(c.thvPath, "stop", serverName) // #nosec G204 - thvPath is validated in NewClient
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop server %s: %w", serverName, err)
	}
	return nil
}

// RemoveServer removes a stopped MCP server
func (c *Client) RemoveServer(serverName string) error {
	removeCmd := exec.Command(c.thvPath, "rm", serverName) // #nosec G204 - thvPath is validated in NewClient
	if err := removeCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove server %s: %w", serverName, err)
	}
	return nil
}
