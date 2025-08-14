// Package main provides the registry builder CLI tool
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stacklok/toolhive-registry/pkg/registry"
)

var (
	// Version information (set during build)
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "registry-builder",
	Short: "Build and manage the ToolHive registry",
	Long: `registry-builder is a tool for building and managing the ToolHive registry.
It converts modular YAML registry entries into various output formats
including ToolHive JSON and upstream MCP Registry formats.`,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the registry from YAML files",
	Long: `Build the registry by loading all YAML files from the registry directory
and generating output in the specified format.

Supported formats:
  - toolhive: ToolHive JSON format (default)
  - mcp-registry: Upstream MCP Registry format (future)
  - all: Build all supported formats`,
	RunE: runBuild,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate registry entries",
	Long:  `Validate all registry entries without building the output files.`,
	RunE:  runValidate,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registry entries",
	Long:  `List all registry entries found in the registry directory.`,
	RunE:  runList,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("registry-builder %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

var (
	registryPath string
	outputDir    string
	outputFormat string
	verbose      bool
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&registryPath, "registry", "r", "registry", "Path to the registry directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Build command flags
	buildCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "build", "Output directory for built registry files")
	buildCmd.Flags().StringVarP(&outputFormat, "format", "f", "toolhive", "Output format (toolhive, mcp-registry, all)")

	// Add commands
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runBuild(cmd *cobra.Command, args []string) error {
	if verbose {
		log.Printf("Building registry from %s", registryPath)
	}

	// Create loader
	loader := registry.NewLoader(registryPath)

	// Load all entries
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("failed to load registry entries: %w", err)
	}

	entries := loader.GetEntries()
	if verbose {
		log.Printf("Loaded %d registry entries", len(entries))
	}

	// Determine which formats to build
	formats := determineFormats(outputFormat)

	// Build each format
	var builtFormats []string
	for _, format := range formats {
		if err := buildFormat(loader, format, outputDir); err != nil {
			return fmt.Errorf("failed to build %s format: %w", format, err)
		}
		builtFormats = append(builtFormats, format)
	}

	fmt.Printf("✓ Successfully built registry with %d entries\n", len(entries))
	fmt.Printf("  Formats: %s\n", strings.Join(builtFormats, ", "))
	fmt.Printf("  Output directory: %s\n", outputDir)

	return nil
}

func determineFormats(format string) []string {
	switch strings.ToLower(format) {
	case "all":
		// Return all supported formats
		// For now, just toolhive, but will expand to include mcp-registry
		return []string{"toolhive"}
	case "mcp-registry", "mcp":
		// Future: Upstream MCP Registry format
		fmt.Println("Note: MCP Registry format support is planned for a future release")
		fmt.Println("This will generate output compatible with https://github.com/modelcontextprotocol/registry")
		return []string{}
	case "toolhive":
		fallthrough
	default:
		return []string{"toolhive"}
	}
}

func buildFormat(loader *registry.Loader, format string, outputDir string) error {
	switch format {
	case "toolhive":
		return buildToolhiveFormat(loader, outputDir)
	case "mcp-registry":
		// Future implementation
		return fmt.Errorf("MCP Registry format not yet implemented")
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func buildToolhiveFormat(loader *registry.Loader, outputDir string) error {
	// Create builder
	builder := registry.NewBuilder(loader)

	// Validate against schema
	if err := builder.ValidateAgainstSchema(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write JSON output
	outputPath := filepath.Join(outputDir, "registry.json")
	if err := builder.WriteJSON(outputPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if verbose {
		log.Printf("Written ToolHive format to %s", outputPath)
	}

	return nil
}

// Future: buildMCPRegistryFormat function will be added here
// func buildMCPRegistryFormat(loader *registry.Loader, outputDir string) error {
//     // Implementation for upstream MCP Registry format
//     // This will create output compatible with the MCP Registry service:
//     // https://github.com/modelcontextprotocol/registry
//     // The format will evolve as the upstream standard evolves
// }

func runValidate(cmd *cobra.Command, args []string) error {
	if verbose {
		log.Printf("Validating registry entries in %s", registryPath)
	}

	// Create loader
	loader := registry.NewLoader(registryPath)

	// Load all entries
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("failed to load registry entries: %w", err)
	}

	entries := loader.GetEntries()

	// Create builder for validation
	builder := registry.NewBuilder(loader)

	// Validate against schema
	if err := builder.ValidateAgainstSchema(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✓ All %d registry entries are valid\n", len(entries))

	if verbose {
		fmt.Println("\nValidated entries:")
		for _, entry := range loader.GetSortedEntries() {
			fmt.Printf("  - %s: %s\n", entry.Name, entry.Description)
		}
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	// Create loader
	loader := registry.NewLoader(registryPath)

	// Load all entries
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("failed to load registry entries: %w", err)
	}

	entries := loader.GetSortedEntries()

	fmt.Printf("Found %d registry entries:\n\n", len(entries))

	for _, entry := range entries {
		status := entry.Status
		if status == "" {
			status = "Active"
		}

		tier := entry.Tier
		if tier == "" {
			tier = "Community"
		}

		fmt.Printf("%-30s [%s/%s] %s\n", entry.Name, tier, status, entry.Image)
		if verbose {
			fmt.Printf("  Description: %s\n", entry.Description)
			fmt.Printf("  Transport:   %s\n", entry.Transport)
			if len(entry.Tools) > 0 {
				fmt.Printf("  Tools:       %d available\n", len(entry.Tools))
			}
			if entry.RepositoryURL != "" {
				fmt.Printf("  Repository:  %s\n", entry.RepositoryURL)
			}
			if entry.License != "" {
				fmt.Printf("  License:     %s\n", entry.License)
			}
			if len(entry.Examples) > 0 {
				fmt.Printf("  Examples:    %d available\n", len(entry.Examples))
			}
			fmt.Println()
		}
	}

	return nil
}
