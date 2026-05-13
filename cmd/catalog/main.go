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

	internalregistry "github.com/stacklok/toolhive-catalog/internal/registry"
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

	// Registries may optionally have a "skills" subdirectory containing
	// individual skill directories with skill.json files.
	skillsSubdir = "skills"
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
      registry-upstream.json`,
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
	buildCmd.Flags().StringVarP(
		&format, "format", "f", formatUpstream,
		"Deprecated: output format; only upstream is supported",
	)
	if err := buildCmd.Flags().MarkDeprecated(
		"format", "legacy format output was removed; only registry-upstream.json is built",
	); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateMetadataCmd)
	rootCmd.AddCommand(updateToolsCmd)
	rootCmd.AddCommand(syncSkillsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// registryInfo holds the name and loaders for a discovered registry.
type registryInfo struct {
	name        string
	loader      *internalregistry.Loader
	skillLoader *internalregistry.SkillLoader
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

		// Optionally load skills if the skills subdirectory exists
		var skillLoader *internalregistry.SkillLoader
		skillsPath := filepath.Join(registriesDir, entry.Name(), skillsSubdir)
		if info, serr := os.Stat(skillsPath); serr == nil && info.IsDir() {
			skillLoader = internalregistry.NewSkillLoader(skillsPath)
			if err := skillLoader.LoadAll(); err != nil {
				return nil, fmt.Errorf("failed to load skills for registry %q: %w", entry.Name(), err)
			}
			if verbose {
				fmt.Printf("  loaded %d skills\n", len(skillLoader.GetEntries()))
			}
		}

		if verbose {
			fmt.Printf("Discovered registry %q with %d entries\n", entry.Name(), len(loader.GetEntries()))
		}

		registries = append(registries, registryInfo{
			name:        entry.Name(),
			loader:      loader,
			skillLoader: skillLoader,
		})
	}

	if len(registries) == 0 {
		return nil, fmt.Errorf("no registries found under %s", registriesDir)
	}

	return registries, nil
}

func runBuild(_ *cobra.Command, _ []string) error {
	if err := validateBuildFormat(format); err != nil {
		return err
	}

	registries, err := discoverRegistries()
	if err != nil {
		return err
	}

	for _, reg := range registries {
		regOutputDir := filepath.Join(outputDir, reg.name)
		if err := os.MkdirAll(regOutputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", regOutputDir, err)
		}

		if err := buildUpstream(reg, regOutputDir); err != nil {
			return fmt.Errorf("failed to build registry %q: %w", reg.name, err)
		}

		fmt.Printf("Built registry %q: %d entries -> %s\n",
			reg.name, len(reg.loader.GetEntries()), regOutputDir)
	}

	return nil
}

func validateBuildFormat(f string) error {
	switch strings.ToLower(f) {
	case "", formatUpstream, formatAll:
		return nil
	case formatToolhive, "legacy":
		return fmt.Errorf("legacy registry format was removed; only %q output is supported", formatUpstream)
	default:
		return fmt.Errorf("unknown format %q: only %q output is supported", f, formatUpstream)
	}
}

func runValidate(_ *cobra.Command, _ []string) error {
	registries, err := discoverRegistries()
	if err != nil {
		return err
	}

	for _, reg := range registries {
		upstreamBuilder := internalregistry.NewBuilder(reg.loader)
		if reg.skillLoader != nil {
			upstreamBuilder.WithSkillLoader(reg.skillLoader)
		}
		if err := upstreamBuilder.ValidateAgainstSchema(); err != nil {
			return fmt.Errorf("registry %q: upstream validation failed: %w", reg.name, err)
		}
		if verbose {
			fmt.Printf("  %s upstream format: valid\n", reg.name)
		}

		skillCount := 0
		if reg.skillLoader != nil {
			skillCount = len(reg.skillLoader.GetEntries())
		}
		fmt.Printf("Registry %q: all %d servers and %d skills valid\n",
			reg.name, len(reg.loader.GetEntries()), skillCount)
	}

	return nil
}

func buildUpstream(reg registryInfo, outDir string) error {
	builder := internalregistry.NewBuilder(reg.loader)
	if reg.skillLoader != nil {
		builder.WithSkillLoader(reg.skillLoader)
	}
	outPath := filepath.Join(outDir, "registry-upstream.json")

	if err := builder.WriteJSON(outPath); err != nil {
		return fmt.Errorf("failed to write upstream registry: %w", err)
	}

	if verbose {
		fmt.Printf("  wrote %s\n", outPath)
	}
	return nil
}
