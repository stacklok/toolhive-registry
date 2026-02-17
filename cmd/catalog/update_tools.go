package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
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
	serverName := filepath.Base(filepath.Dir(path))

	if verbose {
		fmt.Printf("Processing server: %s\n", serverName)
		fmt.Printf("Current tools count: %d\n", len(currentTools))
	}

	newTools, err := fetchToolsFromServer(sf, serverName)
	if err != nil {
		return err
	}

	return applyToolsUpdate(sf, ext, currentTools, newTools)
}

// fetchToolsFromServer starts a temporary MCP server and queries its tools.
func fetchToolsFromServer(
	sf *serverjson.ServerFile, serverName string,
) ([]string, error) {
	client, err := thvclient.NewClient(thvPath, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create thv client: %w", err)
	}

	fmt.Printf("Starting temporary MCP server: %s\n", serverName)
	tempName, err := client.RunServer(sf, serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to run server: %w", err)
	}
	defer func() {
		if stopErr := client.StopServer(tempName); stopErr != nil {
			fmt.Printf(
				"Warning: failed to stop server %s: %v\n",
				tempName, stopErr,
			)
		}
	}()

	tools, err := client.ListTools(tempName)
	if err != nil {
		if logs, logErr := client.Logs(tempName); logErr == nil && logs != "" {
			fmt.Printf("Server logs:\n%s\n", logs)
		}
		return nil, fmt.Errorf("failed to fetch tools: %w", err)
	}

	fmt.Printf("Discovered %d tools\n", len(tools))
	return tools, nil
}

// applyToolsUpdate compares and writes the updated tools list.
func applyToolsUpdate(
	sf *serverjson.ServerFile,
	ext *toolhiveRegistry.ServerExtensions,
	currentTools, newTools []string,
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

	if slices.Equal(currentTools, newTools) {
		fmt.Println("Tools list is already up to date")
		return nil
	}

	printToolsDiff(currentTools, newTools)

	if dryRunTools {
		fmt.Println("[DRY RUN] Would update tools list")
		return nil
	}

	ext.Tools = newTools
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
