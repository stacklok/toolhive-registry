// Package main provides the catalog CLI tool for building registry files
// from server.json entries. It discovers registries under a root directory
// and produces output for each one.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	internalregistry "github.com/stacklok/toolhive-registry/internal/registry"
)

const (
	formatToolhive = "toolhive"
	formatUpstream = "upstream"
	formatAll      = "all"

	defaultRegistries = "registries"
	defaultOutputDir  = "build"

	// Each registry is expected to have a "servers" subdirectory containing
	// individual server directories with server.json files.
	serversSubdir = "servers"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var (
	registriesDir string
	outputDir     string
	format        string
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Build the ToolHive catalog from server.json files",
	Long: `catalog discovers registries under a root directory and builds
registry files from individual server.json entries for each one.

Given a registries directory (default: registries/), it looks for
subdirectories containing a "servers/" folder with server.json files:

  registries/
    toolhive/
      servers/
        github/server.json
        ...

For each registry found, it produces output in the build directory:

  build/
    toolhive/
      registry.json
      official-registry.json`,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build registry files for all discovered registries",
	RunE:  runBuild,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all server.json entries across all registries",
	RunE:  runValidate,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(*cobra.Command, []string) {
		fmt.Printf("catalog %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&registriesDir, "registries", "r", defaultRegistries, "Path to the registries root directory",
	)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	buildCmd.Flags().StringVarP(&outputDir, "output-dir", "o", defaultOutputDir, "Output directory")
	buildCmd.Flags().StringVarP(&format, "format", "f", formatAll,
		fmt.Sprintf("Output format (%s, %s, %s)", formatToolhive, formatUpstream, formatAll))

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateMetadataCmd)
	rootCmd.AddCommand(updateToolsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// registryInfo holds the name and loader for a discovered registry.
type registryInfo struct {
	name   string
	loader *internalregistry.Loader
}

// discoverRegistries walks the registries root directory, finds subdirectories
// that contain a "servers/" folder, and returns a loader for each.
func discoverRegistries() ([]registryInfo, error) {
	entries, err := os.ReadDir(registriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read registries directory %s: %w", registriesDir, err)
	}

	var registries []registryInfo
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		serversPath := filepath.Join(registriesDir, entry.Name(), serversSubdir)
		info, err := os.Stat(serversPath)
		if err != nil || !info.IsDir() {
			if verbose {
				fmt.Printf("Skipping %s (no %s/ directory)\n", entry.Name(), serversSubdir)
			}
			continue
		}

		loader := internalregistry.NewLoader(serversPath)
		if err := loader.LoadAll(); err != nil {
			return nil, fmt.Errorf("failed to load registry %q: %w", entry.Name(), err)
		}

		if verbose {
			fmt.Printf("Discovered registry %q with %d entries\n", entry.Name(), len(loader.GetEntries()))
		}

		registries = append(registries, registryInfo{
			name:   entry.Name(),
			loader: loader,
		})
	}

	if len(registries) == 0 {
		return nil, fmt.Errorf("no registries found under %s", registriesDir)
	}

	return registries, nil
}

func runBuild(_ *cobra.Command, _ []string) error {
	registries, err := discoverRegistries()
	if err != nil {
		return err
	}

	formats := determineFormats(format)

	for _, reg := range registries {
		regOutputDir := filepath.Join(outputDir, reg.name)
		if err := os.MkdirAll(regOutputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", regOutputDir, err)
		}

		for _, f := range formats {
			if err := buildFormat(reg.loader, f, regOutputDir); err != nil {
				return fmt.Errorf("failed to build %s format for registry %q: %w", f, reg.name, err)
			}
		}

		fmt.Printf("Built registry %q: %d entries [%s] -> %s\n",
			reg.name, len(reg.loader.GetEntries()), strings.Join(formats, ", "), regOutputDir)
	}

	return nil
}

func runValidate(_ *cobra.Command, _ []string) error {
	registries, err := discoverRegistries()
	if err != nil {
		return err
	}

	for _, reg := range registries {
		upstreamBuilder := internalregistry.NewBuilder(reg.loader)
		if err := upstreamBuilder.ValidateAgainstSchema(); err != nil {
			return fmt.Errorf("registry %q: upstream validation failed: %w", reg.name, err)
		}
		if verbose {
			fmt.Printf("  %s upstream format: valid\n", reg.name)
		}

		legacyBuilder := internalregistry.NewLegacyBuilder(reg.loader)
		if err := legacyBuilder.ValidateAgainstSchema(); err != nil {
			return fmt.Errorf("registry %q: toolhive validation failed: %w", reg.name, err)
		}
		if verbose {
			fmt.Printf("  %s toolhive format: valid\n", reg.name)
		}

		fmt.Printf("Registry %q: all %d entries valid (both formats)\n", reg.name, len(reg.loader.GetEntries()))
	}

	return nil
}

func determineFormats(f string) []string {
	switch strings.ToLower(f) {
	case formatAll:
		return []string{formatToolhive, formatUpstream}
	case formatUpstream:
		return []string{formatUpstream}
	case formatToolhive:
		return []string{formatToolhive}
	default:
		return []string{formatAll}
	}
}

func buildFormat(loader *internalregistry.Loader, f string, outDir string) error {
	switch f {
	case formatToolhive:
		return buildToolhive(loader, outDir)
	case formatUpstream:
		return buildUpstream(loader, outDir)
	default:
		return fmt.Errorf("unknown format: %s", f)
	}
}

func buildToolhive(loader *internalregistry.Loader, outDir string) error {
	builder := internalregistry.NewLegacyBuilder(loader)
	outPath := filepath.Join(outDir, "registry.json")

	if err := builder.WriteJSON(outPath); err != nil {
		return fmt.Errorf("failed to write toolhive registry: %w", err)
	}

	if verbose {
		fmt.Printf("  wrote %s\n", outPath)
	}
	return nil
}

func buildUpstream(loader *internalregistry.Loader, outDir string) error {
	builder := internalregistry.NewBuilder(loader)
	outPath := filepath.Join(outDir, "official-registry.json")

	if err := builder.WriteJSON(outPath); err != nil {
		return fmt.Errorf("failed to write upstream registry: %w", err)
	}

	if verbose {
		fmt.Printf("  wrote %s\n", outPath)
	}
	return nil
}
