// Package main provides a tool to update MCP server tool lists using thv mcp list
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/stacklok/toolhive/pkg/logger"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/toolhive-registry/pkg/toolhive"
	"github.com/stacklok/toolhive-registry/pkg/types"
)

var (
	specPath    string
	dryRun      bool
	thvPath     string
	addWarnings bool
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "update-tools [spec-file]",
	Short: "Update tool lists in MCP server spec files using thv mcp list",
	Long: `update-tools fetches the current list of tools from an MCP server using
'thv mcp list --server <name>' and updates the tools section in the spec.yaml file.

If no tools are detected but the spec had tools before, it keeps the old list
and adds a warning comment.`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	logger.Initialize()

	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be changed without modifying files")
	rootCmd.Flags().StringVar(&thvPath, "thv-path", "", "Path to thv binary (defaults to searching PATH)")
	rootCmd.Flags().BoolVar(&addWarnings, "add-warnings", true, "Add warning comments when tools can't be fetched")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

func runUpdate(_ *cobra.Command, args []string) error {
	specPath = args[0]

	// Verify spec file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec file not found: %s", specPath)
	}

	// Extract server name from path
	serverDir := filepath.Dir(specPath)
	serverName := filepath.Base(serverDir)

	logger.Infof("Processing server: %s", serverName)
	if verbose {
		logger.Infof("Spec file: %s", specPath)
	}

	// Load current spec and get tools
	currentTools, err := getCurrentTools()
	if err != nil {
		return err
	}

	// Fetch new tools from thv
	newTools, err := fetchToolsFromMCP(serverName)
	if err != nil {
		return handleFetchError(err, currentTools)
	}

	logger.Infof("New tools count: %d", len(newTools))

	// Handle empty tools case
	if err := handleEmptyTools(newTools, currentTools); err != nil {
		return err
	}

	// Compare and update tools
	return compareAndUpdateTools(currentTools, newTools)
}

func getCurrentTools() ([]string, error) {
	currentSpec, err := loadSpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	currentTools := currentSpec.GetTools()
	logger.Infof("Current tools count: %d", len(currentTools))
	return currentTools, nil
}

func handleFetchError(err error, currentTools []string) error {
	logger.Warnf("Failed to fetch tools from MCP server: %v", err)

	if len(currentTools) > 0 && addWarnings {
		if !dryRun {
			if err := toolhive.AddWarningComment(specPath, "Tool list fetch failed", "Manual verification may be required"); err != nil {
				logger.Warnf("Failed to add warning comment: %v", err)
			}
		} else {
			logger.Info("[DRY RUN] Would add warning comment about fetch failure")
		}
	}
	return fmt.Errorf("failed to fetch tools: %w", err)
}

func handleEmptyTools(newTools, currentTools []string) error {
	if len(newTools) == 0 && len(currentTools) > 0 {
		logger.Warnf("No tools detected but spec file had %d tools previously", len(currentTools))
		logger.Info("Keeping existing tools list")

		if addWarnings {
			if !dryRun {
				if err := toolhive.AddWarningComment(specPath, "Tool list could not be auto-updated",
					"Please verify the tools list manually"); err != nil {
					logger.Warnf("Failed to add warning comment: %v", err)
				}
			} else {
				logger.Info("[DRY RUN] Would add warning comment about empty tool list")
			}
		}
		return fmt.Errorf("empty tools list detected")
	}
	return nil
}

func compareAndUpdateTools(currentTools, newTools []string) error {
	// Sort both lists for comparison
	sort.Strings(currentTools)
	sort.Strings(newTools)

	// Check if tools changed using slices.Equal
	if slices.Equal(currentTools, newTools) {
		logger.Info("Tools list is already up to date")
		return nil
	}

	// Show changes
	logger.Info("Tools list changes detected:")
	if verbose {
		showDetailedDiff(currentTools, newTools)
	} else {
		showSummaryDiff(currentTools, newTools)
	}

	// Update the spec file
	if !dryRun {
		if err := toolhive.UpdateSpecTools(specPath, newTools); err != nil {
			return fmt.Errorf("failed to update spec file: %w", err)
		}
		logger.Info("Successfully updated tools list")
	} else {
		logger.Info("[DRY RUN] Would update tools list in spec file")
	}

	return nil
}

func loadSpec(path string) (*types.RegistryEntry, error) {
	data, err := os.ReadFile(path) // #nosec G304 - path is controlled by application
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var entry types.RegistryEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &entry, nil
}

func fetchToolsFromMCP(serverName string) ([]string, error) {
	// Load the spec to get the configuration
	spec, err := loadSpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	// Create ToolHive client
	client, err := toolhive.NewClient(thvPath, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create ToolHive client: %w", err)
	}

	// Run the MCP server
	logger.Infof("Starting temporary MCP server: %s", serverName)
	tempName, err := client.RunServer(spec, serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to run server: %w", err)
	}
	defer func() {
		// Clean up the temporary server
		if err := client.StopServer(tempName); err != nil {
			logger.Warnf("Failed to stop temporary server %s: %v", tempName, err)
		}
		//if err := client.RemoveServer(tempName); err != nil {
		//	logger.Warnf("Failed to remove temporary server %s: %v", tempName, err)
		//}
	}()

	// Query the server for tools
	tools, err := client.ListTools(tempName)
	if err != nil {
		// Get Logs for debugging
		logs, logErr := client.Logs(tempName)
		if logErr != nil {
			logger.Warnf("Failed to fetch logs from temporary server %s: %v", tempName, logErr)
		}
		if logs != "" {
			logger.Infof("Logs from temporary server %s:\n%s", tempName, logs)
		}
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return tools, nil
}

func showDetailedDiff(current, newTools []string) {
	diff := cmp.Diff(current, newTools)
	if diff != "" {
		logger.Info("Detailed diff:")
		fmt.Println(diff)
	}
}

func showSummaryDiff(current, newTools []string) {
	currentSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, t := range current {
		currentSet[t] = true
	}
	for _, t := range newTools {
		newSet[t] = true
	}

	// Find added tools
	var added []string
	for t := range newSet {
		if !currentSet[t] {
			added = append(added, t)
		}
	}

	// Find removed tools
	var removed []string
	for t := range currentSet {
		if !newSet[t] {
			removed = append(removed, t)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)

	if len(added) > 0 {
		logger.Infof("  Added tools (%d):", len(added))
		for _, t := range added {
			logger.Infof("    + %s", t)
		}
	}

	if len(removed) > 0 {
		logger.Infof("  Removed tools (%d):", len(removed))
		for _, t := range removed {
			logger.Infof("    - %s", t)
		}
	}
}
