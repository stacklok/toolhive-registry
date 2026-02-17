package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry/registry"

	"github.com/stacklok/toolhive-registry/internal/serverjson"
	"github.com/stacklok/toolhive-registry/internal/thvclient"
)

var (
	dryRunTools bool
	thvPath     string
)

var updateToolsCmd = &cobra.Command{
	Use:   "update-tools <server.json>",
	Short: "Update tools list by querying the running MCP server",
	Long: `Starts a temporary MCP server from the server.json package definition,
queries it for available tools, and updates the tools list in the _meta extensions.`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdateTools,
}

func init() {
	updateToolsCmd.Flags().BoolVarP(
		&dryRunTools, "dry-run", "d", false,
		"Show changes without writing",
	)
	updateToolsCmd.Flags().StringVar(
		&thvPath, "thv-path", "",
		"Path to thv binary (searches PATH if empty)",
	)
}

func runUpdateTools(_ *cobra.Command, args []string) error {
	path := args[0]

	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		return err
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		return err
	}

	currentTools := ext.Tools
	currentDefs := ext.ToolDefinitions
	serverName := filepath.Base(filepath.Dir(path))

	if verbose {
		fmt.Printf("Processing server: %s\n", serverName)
		fmt.Printf("Current tools count: %d\n", len(currentTools))
		fmt.Printf("Current tool definitions count: %d\n", len(currentDefs))
	}

	newTools, newDefs, err := fetchToolsFromServer(sf, serverName)
	if err != nil {
		return err
	}

	return applyToolsUpdate(sf, ext, currentTools, newTools, currentDefs, newDefs)
}

// fetchToolsFromServer starts a temporary MCP server and queries its tools.
func fetchToolsFromServer(
	sf *serverjson.ServerFile, serverName string,
) ([]string, []mcp.Tool, error) {
	client, err := thvclient.NewClient(thvPath, verbose)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create thv client: %w", err)
	}

	fmt.Printf("Starting temporary MCP server: %s\n", serverName)
	tempName, err := client.RunServer(sf, serverName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run server: %w", err)
	}
	defer func() {
		if stopErr := client.StopServer(tempName); stopErr != nil {
			fmt.Printf(
				"Warning: failed to stop server %s: %v\n",
				tempName, stopErr,
			)
		}
	}()

	// Fetch full tool definitions first; extract names from them.
	defs, err := client.ListToolDefinitions(tempName)
	if err != nil {
		if logs, logErr := client.Logs(tempName); logErr == nil && logs != "" {
			fmt.Printf("Server logs:\n%s\n", logs)
		}
		return nil, nil, fmt.Errorf("failed to fetch tools: %w", err)
	}

	// If we got definitions, extract tool names from them.
	if defs != nil {
		var tools []string
		for _, d := range defs {
			tools = append(tools, d.Name)
		}
		sort.Strings(tools)
		fmt.Printf("Discovered %d tools (with definitions)\n", len(tools))
		return tools, defs, nil
	}

	// Text fallback: no definitions available, get names only.
	tools, err := client.ListTools(tempName)
	if err != nil {
		if logs, logErr := client.Logs(tempName); logErr == nil && logs != "" {
			fmt.Printf("Server logs:\n%s\n", logs)
		}
		return nil, nil, fmt.Errorf("failed to fetch tools: %w", err)
	}

	fmt.Printf("Discovered %d tools (names only)\n", len(tools))
	return tools, nil, nil
}

// applyToolsUpdate compares and writes the updated tools list and tool definitions.
func applyToolsUpdate(
	sf *serverjson.ServerFile,
	ext *toolhiveRegistry.ServerExtensions,
	currentTools, newTools []string,
	currentDefs, newDefs []mcp.Tool,
) error {
	if len(newTools) == 0 && len(currentTools) > 0 {
		fmt.Printf(
			"Warning: no tools detected but had %d previously. Keeping existing.\n",
			len(currentTools),
		)
		return fmt.Errorf("empty tools list detected")
	}

	sort.Strings(currentTools)
	sort.Strings(newTools)

	toolsChanged := !slices.Equal(currentTools, newTools)
	defsChanged := !toolDefinitionsEqual(currentDefs, newDefs)

	if !toolsChanged && !defsChanged {
		fmt.Println("Tools list is already up to date")
		return nil
	}

	if toolsChanged {
		printToolsDiff(currentTools, newTools)
	}
	if defsChanged {
		printToolDefsDiff(currentDefs, newDefs)
	}

	if dryRunTools {
		fmt.Println("[DRY RUN] Would update tools list")
		return nil
	}

	ext.Tools = newTools
	ext.ToolDefinitions = newDefs
	if ext.Metadata == nil {
		ext.Metadata = &toolhiveRegistry.Metadata{}
	}
	ext.Metadata.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	if err := sf.UpdateExtensions(ext); err != nil {
		return fmt.Errorf("failed to write server.json: %w", err)
	}

	fmt.Println("Successfully updated tools list")
	return nil
}

func printToolsDiff(currentTools, newTools []string) {
	if verbose {
		diff := cmp.Diff(currentTools, newTools)
		if diff != "" {
			fmt.Printf("Diff:\n%s\n", diff)
		}
	} else {
		showToolsSummary(currentTools, newTools)
	}
}

func showToolsSummary(current, newTools []string) {
	added := diffSlices(newTools, current)
	removed := diffSlices(current, newTools)

	if len(added) > 0 {
		fmt.Printf("  Added (%d): %v\n", len(added), added)
	}
	if len(removed) > 0 {
		fmt.Printf("  Removed (%d): %v\n", len(removed), removed)
	}
}

// diffSlices returns items in a that are not in b.
func diffSlices(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	var result []string
	for _, v := range a {
		if _, ok := set[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// toolDefinitionsEqual compares two slices of tool definitions by name.
func toolDefinitionsEqual(a, b []mcp.Tool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Description != b[i].Description {
			return false
		}
	}
	return true
}

func printToolDefsDiff(currentDefs, newDefs []mcp.Tool) {
	if verbose {
		currentNames := toolDefNames(currentDefs)
		newNames := toolDefNames(newDefs)
		diff := cmp.Diff(currentNames, newNames)
		if diff != "" {
			fmt.Printf("Tool definitions diff:\n%s\n", diff)
		}
	} else {
		fmt.Printf("  Tool definitions: %d -> %d\n", len(currentDefs), len(newDefs))
	}
}

func toolDefNames(defs []mcp.Tool) []string {
	names := make([]string, 0, len(defs))
	for _, d := range defs {
		names = append(names, d.Name)
	}
	return names
}
