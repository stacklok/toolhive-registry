// Package main provides a tool to import ToolHive registry.json into modular YAML format
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry"
	"gopkg.in/yaml.v3"
)

var (
	sourceURL  string
	sourceFile string
	outputDir  string
	verbose    bool
	dryRun     bool
)

var rootCmd = &cobra.Command{
	Use:   "import-from-toolhive",
	Short: "Import ToolHive registry.json into modular YAML format",
	Long: `Import the existing ToolHive registry.json and convert it to the modular YAML format.
Each registry entry will be converted to its own directory with a spec.yaml file.

This tool is specifically for importing from ToolHive's format. For migrating to
upstream MCP Registry format, use the 'migrate' command (future).`,
	RunE: runImport,
}

func init() {
	rootCmd.Flags().StringVarP(&sourceURL, "url", "u",
		"https://raw.githubusercontent.com/stacklok/toolhive/main/pkg/registry/data/registry.json",
		"URL to fetch registry.json from")
	rootCmd.Flags().StringVarP(&sourceFile, "file", "f", "", "Local registry.json file (overrides URL)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "registry", "Output directory for YAML files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without actually creating files")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runImport(cmd *cobra.Command, args []string) error {
	// Load the registry JSON
	var registryData []byte
	var err error

	if sourceFile != "" {
		// Load from file
		if verbose {
			log.Printf("Loading registry from file: %s", sourceFile)
		}
		registryData, err = os.ReadFile(sourceFile) // #nosec G304 - file path comes from command line flag
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		// Fetch from URL
		if verbose {
			log.Printf("Fetching registry from URL: %s", sourceURL)
		}
		resp, err := http.Get(sourceURL) // #nosec G107 - URL comes from command line flag
		if err != nil {
			return fmt.Errorf("failed to fetch registry: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch registry: HTTP %d", resp.StatusCode)
		}

		registryData, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// Parse the JSON
	var registry toolhiveRegistry.Registry
	if err := json.Unmarshal(registryData, &registry); err != nil {
		return fmt.Errorf("failed to parse registry JSON: %w", err)
	}

	fmt.Printf("Found %d registry entries to import\n", len(registry.Servers))

	if dryRun {
		fmt.Println("\nDry run mode - no files will be created")
		fmt.Println("\nWould create the following structure:")
	}

	// Process each server entry in alphabetical order
	var names []string
	for name := range registry.Servers {
		names = append(names, name)
	}
	// Sort names alphabetically
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}

	successCount := 0
	for _, name := range names {
		server := registry.Servers[name]
		if err := importEntry(name, server, outputDir, dryRun); err != nil {
			log.Printf("Warning: Failed to import %s: %v", name, err)
			continue
		}
		successCount++
	}

	if !dryRun {
		fmt.Printf("\n✓ Successfully imported %d/%d entries to %s\n", successCount, len(registry.Servers), outputDir)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review the imported entries in the registry/ directory")
		fmt.Println("  2. Run 'registry-builder validate' to validate all entries")
		fmt.Println("  3. Run 'registry-builder build' to generate the registry.json")
	} else {
		fmt.Printf("\n✓ Would import %d/%d entries\n", successCount, len(registry.Servers))
	}

	return nil
}

func importEntry(name string, server *toolhiveRegistry.ImageMetadata, outputDir string, dryRun bool) error {
	// Sanitize the name for use as a directory
	dirName := sanitizeName(name)
	entryDir := filepath.Join(outputDir, dirName)
	specPath := filepath.Join(entryDir, "spec.yaml")

	if verbose || dryRun {
		fmt.Printf("  %s -> %s\n", name, specPath)
	}

	if dryRun {
		return nil
	}

	// Create the directory
	if err := os.MkdirAll(entryDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Ensure the name is set in the metadata
	if server.Name == "" {
		server.Name = name
	}

	// Create YAML content with proper formatting
	yamlData, err := yaml.Marshal(server)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Add a header comment with metadata
	header := fmt.Sprintf(`# %s MCP Server Registry Entry
# Auto-imported from ToolHive registry.json
# 
# Original source: https://github.com/stacklok/toolhive
# Import timestamp: %s
# ---
`, name, time.Now().UTC().Format(time.RFC3339))

	finalContent := header + string(yamlData)

	// Write the spec.yaml file
	if err := os.WriteFile(specPath, []byte(finalContent), 0600); err != nil {
		return fmt.Errorf("failed to write spec.yaml: %w", err)
	}

	// Optionally create a README for complex entries
	if shouldCreateReadme(server) {
		readmePath := filepath.Join(entryDir, "README.md")
		readmeContent := generateReadme(name, server)
		if err := os.WriteFile(readmePath, []byte(readmeContent), 0600); err != nil {
			// Non-fatal error
			if verbose {
				log.Printf("Warning: Failed to write README for %s: %v", name, err)
			}
		}
	}

	return nil
}

func sanitizeName(name string) string {
	// Replace problematic characters with hyphens
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		".", "-",
		"/", "-",
		"\\", "-",
	)
	sanitized := replacer.Replace(name)

	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// Remove any remaining non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remove leading/trailing hyphens
	finalName := strings.Trim(result.String(), "-")

	// Collapse multiple hyphens into one
	for strings.Contains(finalName, "--") {
		finalName = strings.ReplaceAll(finalName, "--", "-")
	}

	return finalName
}

func shouldCreateReadme(server *toolhiveRegistry.ImageMetadata) bool {
	// Create README for entries with substantial documentation needs
	return len(server.Tools) > 10 || len(server.EnvVars) > 5 || len(server.Tags) > 10
}

func generateReadme(name string, server *toolhiveRegistry.ImageMetadata) string {
	var readme strings.Builder

	readme.WriteString(fmt.Sprintf("# %s\n\n", name))

	if server.Description != "" {
		readme.WriteString(fmt.Sprintf("%s\n\n", server.Description))
	}

	// Basic information section
	readme.WriteString("## Basic Information\n\n")

	if server.Image != "" {
		readme.WriteString(fmt.Sprintf("- **Image:** `%s`\n", server.Image))
	}

	if server.RepositoryURL != "" {
		readme.WriteString(fmt.Sprintf("- **Repository:** [%s](%s)\n", server.RepositoryURL, server.RepositoryURL))
	}

	if server.Tier != "" {
		readme.WriteString(fmt.Sprintf("- **Tier:** %s\n", server.Tier))
	}

	if server.Status != "" {
		readme.WriteString(fmt.Sprintf("- **Status:** %s\n", server.Status))
	}

	if server.Transport != "" {
		readme.WriteString(fmt.Sprintf("- **Transport:** %s\n", server.Transport))
	}

	// Tools section
	if len(server.Tools) > 0 {
		readme.WriteString("\n## Available Tools\n\n")
		readme.WriteString(fmt.Sprintf("This server provides %d tools:\n\n", len(server.Tools)))

		// Group tools in columns for better readability if there are many
		if len(server.Tools) > 10 {
			for i := 0; i < len(server.Tools); i += 3 {
				for j := 0; j < 3 && i+j < len(server.Tools); j++ {
					readme.WriteString(fmt.Sprintf("- `%s`", server.Tools[i+j]))
					if j < 2 && i+j+1 < len(server.Tools) {
						readme.WriteString(" | ")
					}
				}
				readme.WriteString("\n")
			}
		} else {
			for _, tool := range server.Tools {
				readme.WriteString(fmt.Sprintf("- `%s`\n", tool))
			}
		}
	}

	// Environment Variables section
	if len(server.EnvVars) > 0 {
		readme.WriteString("\n## Environment Variables\n\n")

		// Separate required and optional
		var required, optional []*toolhiveRegistry.EnvVar
		for _, env := range server.EnvVars {
			if env.Required {
				required = append(required, env)
			} else {
				optional = append(optional, env)
			}
		}

		if len(required) > 0 {
			readme.WriteString("### Required\n\n")
			for _, env := range required {
				secret := ""
				if env.Secret {
					secret = " 🔒"
				}
				readme.WriteString(fmt.Sprintf("- **%s**%s: %s\n", env.Name, secret, env.Description))
			}
		}

		if len(optional) > 0 {
			readme.WriteString("\n### Optional\n\n")
			for _, env := range optional {
				secret := ""
				if env.Secret {
					secret = " 🔒"
				}
				readme.WriteString(fmt.Sprintf("- **%s**%s: %s\n", env.Name, secret, env.Description))
				if env.Default != "" {
					readme.WriteString(fmt.Sprintf("  - Default: `%s`\n", env.Default))
				}
			}
		}
	}

	// Tags section
	if len(server.Tags) > 0 {
		readme.WriteString("\n## Tags\n\n")
		for _, tag := range server.Tags {
			readme.WriteString(fmt.Sprintf("`%s` ", tag))
		}
		readme.WriteString("\n")
	}

	// Metadata section
	if server.Metadata != nil {
		readme.WriteString("\n## Statistics\n\n")
		if server.Metadata.Stars > 0 {
			readme.WriteString(fmt.Sprintf("- ⭐ Stars: %d\n", server.Metadata.Stars))
		}
		if server.Metadata.Pulls > 0 {
			readme.WriteString(fmt.Sprintf("- 📦 Pulls: %d\n", server.Metadata.Pulls))
		}
		if server.Metadata.LastUpdated != "" {
			readme.WriteString(fmt.Sprintf("- 🕐 Last Updated: %s\n", server.Metadata.LastUpdated))
		}
	}

	return readme.String()
}
