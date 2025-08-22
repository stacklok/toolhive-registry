// Package main provides a tool to import ToolHive registry.json into modular YAML format
package main

import (
	"bytes"
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

func runImport(_ *cobra.Command, _ []string) error {
	registryData, err := loadRegistryData()
	if err != nil {
		return err
	}

	registry, err := parseRegistry(registryData)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d registry entries to import\n", len(registry.Servers))

	if dryRun {
		fmt.Println("\nDry run mode - no files will be created")
		fmt.Println("\nWould create the following structure:")
	}

	successCount := processRegistryEntries(registry)
	printImportSummary(successCount, len(registry.Servers))

	return nil
}

func loadRegistryData() ([]byte, error) {
	if sourceFile != "" {
		return loadFromFile()
	}
	return loadFromURL()
}

func loadFromFile() ([]byte, error) {
	if verbose {
		log.Printf("Loading registry from file: %s", sourceFile)
	}
	registryData, err := os.ReadFile(sourceFile) // #nosec G304 - file path comes from command line flag
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return registryData, nil
}

func loadFromURL() ([]byte, error) {
	if verbose {
		log.Printf("Fetching registry from URL: %s", sourceURL)
	}
	resp, err := http.Get(sourceURL) // #nosec G107 - URL comes from command line flag
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch registry: HTTP %d", resp.StatusCode)
	}

	registryData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return registryData, nil
}

func parseRegistry(registryData []byte) (*toolhiveRegistry.Registry, error) {
	var registry toolhiveRegistry.Registry
	if err := json.Unmarshal(registryData, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry JSON: %w", err)
	}
	return &registry, nil
}

func processRegistryEntries(registry *toolhiveRegistry.Registry) int {
	names := getSortedServerNames(registry)

	successCount := 0
	for _, name := range names {
		server := registry.Servers[name]
		if err := importEntry(name, server, outputDir, dryRun); err != nil {
			log.Printf("Warning: Failed to import %s: %v", name, err)
			continue
		}
		successCount++
	}
	return successCount
}

func getSortedServerNames(registry *toolhiveRegistry.Registry) []string {
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
	return names
}

func printImportSummary(successCount, totalCount int) {
	if !dryRun {
		fmt.Printf("\n✓ Successfully imported %d/%d entries to %s\n", successCount, totalCount, outputDir)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review the imported entries in the registry/ directory")
		fmt.Println("  2. Run 'registry-builder validate' to validate all entries")
		fmt.Println("  3. Run 'registry-builder build' to generate the registry.json")
	} else {
		fmt.Printf("\n✓ Would import %d/%d entries\n", successCount, totalCount)
	}
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

	// Create YAML content with proper formatting (2-space indentation)
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(server); err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	err := encoder.Close()
	if err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}
	yamlData := buf.Bytes()

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

	addReadmeHeader(&readme, name, server.Description)
	addBasicInformation(&readme, server)
	addToolsSection(&readme, server.Tools)
	addEnvironmentVariablesSection(&readme, server.EnvVars)
	addTagsSection(&readme, server.Tags)
	addMetadataSection(&readme, server.Metadata)

	return readme.String()
}

func addReadmeHeader(readme *strings.Builder, name, description string) {
	fmt.Fprintf(readme, "# %s\n\n", name)
	if description != "" {
		fmt.Fprintf(readme, "%s\n\n", description)
	}
}

func addBasicInformation(readme *strings.Builder, server *toolhiveRegistry.ImageMetadata) {
	readme.WriteString("## Basic Information\n\n")

	if server.Image != "" {
		fmt.Fprintf(readme, "- **Image:** `%s`\n", server.Image)
	}
	if server.RepositoryURL != "" {
		fmt.Fprintf(readme, "- **Repository:** [%s](%s)\n", server.RepositoryURL, server.RepositoryURL)
	}
	if server.Tier != "" {
		fmt.Fprintf(readme, "- **Tier:** %s\n", server.Tier)
	}
	if server.Status != "" {
		fmt.Fprintf(readme, "- **Status:** %s\n", server.Status)
	}
	if server.Transport != "" {
		fmt.Fprintf(readme, "- **Transport:** %s\n", server.Transport)
	}
}

func addToolsSection(readme *strings.Builder, tools []string) {
	if len(tools) == 0 {
		return
	}

	readme.WriteString("\n## Available Tools\n\n")
	fmt.Fprintf(readme, "This server provides %d tools:\n\n", len(tools))

	if len(tools) > 10 {
		addToolsInColumns(readme, tools)
	} else {
		addToolsList(readme, tools)
	}
}

func addToolsInColumns(readme *strings.Builder, tools []string) {
	for i := 0; i < len(tools); i += 3 {
		for j := 0; j < 3 && i+j < len(tools); j++ {
			fmt.Fprintf(readme, "- `%s`", tools[i+j])
			if j < 2 && i+j+1 < len(tools) {
				readme.WriteString(" | ")
			}
		}
		readme.WriteString("\n")
	}
}

func addToolsList(readme *strings.Builder, tools []string) {
	for _, tool := range tools {
		fmt.Fprintf(readme, "- `%s`\n", tool)
	}
}

func addEnvironmentVariablesSection(readme *strings.Builder, envVars []*toolhiveRegistry.EnvVar) {
	if len(envVars) == 0 {
		return
	}

	readme.WriteString("\n## Environment Variables\n\n")

	required, optional := separateEnvVars(envVars)
	addRequiredEnvVars(readme, required)
	addOptionalEnvVars(readme, optional)
}

func separateEnvVars(envVars []*toolhiveRegistry.EnvVar) ([]*toolhiveRegistry.EnvVar, []*toolhiveRegistry.EnvVar) {
	var required, optional []*toolhiveRegistry.EnvVar
	for _, env := range envVars {
		if env.Required {
			required = append(required, env)
		} else {
			optional = append(optional, env)
		}
	}
	return required, optional
}

func addRequiredEnvVars(readme *strings.Builder, required []*toolhiveRegistry.EnvVar) {
	if len(required) == 0 {
		return
	}

	readme.WriteString("### Required\n\n")
	for _, env := range required {
		secret := getSecretIndicator(env.Secret)
		fmt.Fprintf(readme, "- **%s**%s: %s\n", env.Name, secret, env.Description)
	}
}

func addOptionalEnvVars(readme *strings.Builder, optional []*toolhiveRegistry.EnvVar) {
	if len(optional) == 0 {
		return
	}

	readme.WriteString("\n### Optional\n\n")
	for _, env := range optional {
		secret := getSecretIndicator(env.Secret)
		fmt.Fprintf(readme, "- **%s**%s: %s\n", env.Name, secret, env.Description)
		if env.Default != "" {
			fmt.Fprintf(readme, "  - Default: `%s`\n", env.Default)
		}
	}
}

func getSecretIndicator(isSecret bool) string {
	if isSecret {
		return " 🔒"
	}
	return ""
}

func addTagsSection(readme *strings.Builder, tags []string) {
	if len(tags) == 0 {
		return
	}

	readme.WriteString("\n## Tags\n\n")
	for _, tag := range tags {
		fmt.Fprintf(readme, "`%s` ", tag)
	}
	readme.WriteString("\n")
}

func addMetadataSection(readme *strings.Builder, metadata *toolhiveRegistry.Metadata) {
	if metadata == nil {
		return
	}

	readme.WriteString("\n## Statistics\n\n")
	if metadata.Stars > 0 {
		fmt.Fprintf(readme, "- ⭐ Stars: %d\n", metadata.Stars)
	}
	if metadata.Pulls > 0 {
		fmt.Fprintf(readme, "- 📦 Pulls: %d\n", metadata.Pulls)
	}
	if metadata.LastUpdated != "" {
		fmt.Fprintf(readme, "- 🕐 Last Updated: %s\n", metadata.LastUpdated)
	}
}
