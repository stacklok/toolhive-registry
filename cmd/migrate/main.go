// Package main provides a one-time migration tool that converts spec.yaml files
// from the legacy registry format into individual server.json files using the
// upstream MCP ServerJSON format.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	upstream "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/spf13/cobra"
	"github.com/stacklok/toolhive/pkg/registry/converters"

	"github.com/stacklok/toolhive-registry/pkg/legacy/registry"
	"github.com/stacklok/toolhive-registry/pkg/legacy/types"
)

var (
	source  string
	target  string
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate spec.yaml files to server.json format",
	Long: `migrate reads existing spec.yaml files from the legacy registry directory
and writes individual server.json files in the upstream MCP ServerJSON format.

Output structure: <target>/<name>/server.json`,
	RunE: runMigrate,
}

func init() {
	rootCmd.Flags().StringVar(&source, "source", "registry", "Path to existing spec.yaml directory")
	rootCmd.Flags().StringVar(&target, "target", "registries/toolhive/servers", "Output directory for server.json files")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be written without writing")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Log each conversion")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runMigrate(_ *cobra.Command, _ []string) error {
	loader := registry.NewLoader(source)
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("failed to load registry entries: %w", err)
	}

	entries := loader.GetEntries()
	if len(entries) == 0 {
		return fmt.Errorf("no entries found in %s", source)
	}

	// Sort names for deterministic output
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	var converted, errCount int
	var errors []string

	for _, name := range names {
		entry := entries[name]

		serverJSON, err := convertEntry(name, entry)
		if err != nil {
			errCount++
			msg := fmt.Sprintf("  SKIP %s: %v", name, err)
			errors = append(errors, msg)
			if verbose {
				log.Println(msg)
			}
			continue
		}

		if err := writeServerJSON(name, serverJSON); err != nil {
			errCount++
			msg := fmt.Sprintf("  FAIL %s: %v", name, err)
			errors = append(errors, msg)
			if verbose {
				log.Println(msg)
			}
			continue
		}

		converted++
		if verbose {
			log.Printf("  OK   %s", name)
		}
	}

	// Summary
	fmt.Printf("Migration complete: %d/%d converted", converted, len(entries))
	if dryRun {
		fmt.Print(" (dry-run)")
	}
	fmt.Println()

	if errCount > 0 {
		fmt.Printf("Errors: %d\n", errCount)
		for _, e := range errors {
			fmt.Println(e)
		}
		return fmt.Errorf("%d entries failed to convert", errCount)
	}

	return nil
}

func convertEntry(name string, entry *types.RegistryEntry) (*upstream.ServerJSON, error) {
	var sj *upstream.ServerJSON
	var err error

	switch {
	case entry.IsImage():
		sj, err = converters.ImageMetadataToServerJSON(name, entry.ImageMetadata)
		if err != nil {
			return nil, fmt.Errorf("image conversion: %w", err)
		}
		if sj == nil {
			return nil, fmt.Errorf("image conversion returned nil")
		}

	case entry.IsRemote():
		sj, err = converters.RemoteServerMetadataToServerJSON(name, entry.RemoteServerMetadata)
		if err != nil {
			return nil, fmt.Errorf("remote conversion: %w", err)
		}
		if sj == nil {
			return nil, fmt.Errorf("remote conversion returned nil")
		}

	default:
		return nil, fmt.Errorf("entry is neither image nor remote")
	}

	// Convert name to reverse-DNS format (e.g. "github" -> "io.github.stacklok/github")
	sj.Name = convertNameToReverseDNS(sj.Name)

	return sj, nil
}

// convertNameToReverseDNS converts simple server names to reverse-DNS format.
func convertNameToReverseDNS(name string) string {
	if strings.Contains(name, "/") {
		return name
	}
	return "io.github.stacklok/" + name
}

func writeServerJSON(name string, serverJSON *upstream.ServerJSON) error {
	data, err := json.MarshalIndent(serverJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	// Ensure trailing newline
	data = append(data, '\n')

	dir := filepath.Join(target, name)
	path := filepath.Join(dir, "server.json")

	if dryRun {
		if verbose {
			log.Printf("  Would write %s (%d bytes)", path, len(data))
		}
		return nil
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
