package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/stacklok/toolhive/pkg/container/verifier"
	"github.com/stacklok/toolhive/pkg/registry/converters"
	toolhiveRegistry "github.com/stacklok/toolhive/pkg/registry/registry"

	"github.com/stacklok/toolhive-registry/internal/metadata"
	"github.com/stacklok/toolhive-registry/internal/serverjson"
)

var (
	dryRunMetadata   bool
	githubToken      string
	verifyProvenance bool
	oldestCount      int
)

var updateMetadataCmd = &cobra.Command{
	Use:   "update-metadata [server.json]",
	Short: "Update GitHub stars and last_updated for a server.json file",
	Long: `Fetches the latest GitHub stars for the specified server.json file
and updates its _meta extensions. Optionally verifies provenance.

Modes:
  catalog update-metadata <server.json>   Update a single file
  catalog update-metadata --oldest 5      Update the 5 oldest entries`,
	Args: validateUpdateMetadataArgs,
	RunE: runUpdateMetadata,
}

func init() {
	updateMetadataCmd.Flags().BoolVarP(
		&dryRunMetadata, "dry-run", "d", false,
		"Show changes without writing",
	)
	updateMetadataCmd.Flags().StringVarP(
		&githubToken, "github-token", "t", "",
		"GitHub API token (or set GITHUB_TOKEN env var)",
	)
	updateMetadataCmd.Flags().BoolVar(
		&verifyProvenance, "verify-provenance", false,
		"Verify provenance before updating",
	)
	updateMetadataCmd.Flags().IntVar(
		&oldestCount, "oldest", 0,
		"Update the N oldest entries by last_updated timestamp",
	)
}

func validateUpdateMetadataArgs(_ *cobra.Command, args []string) error {
	if oldestCount > 0 && len(args) > 0 {
		return fmt.Errorf("cannot use --oldest with a positional argument")
	}
	if oldestCount == 0 && len(args) == 0 {
		return fmt.Errorf("requires a server.json path or --oldest N")
	}
	if oldestCount < 0 {
		return fmt.Errorf("--oldest must be positive, got %d", oldestCount)
	}
	if len(args) > 1 {
		return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
	}
	return nil
}

func runUpdateMetadata(_ *cobra.Command, args []string) error {
	if oldestCount > 0 {
		return runBatchUpdateMetadata()
	}
	return runSingleUpdateMetadata(args[0])
}

func runSingleUpdateMetadata(path string) error {
	ctx := context.Background()

	token := githubToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	sf, err := serverjson.LoadServerFile(path)
	if err != nil {
		return err
	}

	ext, err := sf.GetExtensions()
	if err != nil {
		return err
	}

	if ext.Metadata == nil {
		ext.Metadata = &toolhiveRegistry.Metadata{}
	}

	if verifyProvenance {
		if err := verifyServerProvenance(sf, ext); err != nil {
			return fmt.Errorf("provenance verification failed: %w", err)
		}
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	fetcher := metadata.NewFetcher(httpClient, token)

	repoURL := sf.RepositoryURL()
	newStars, err := fetcher.FetchStars(ctx, repoURL)
	if err != nil {
		fmt.Printf("Warning: failed to fetch stars: %v\n", err)
		newStars = ext.Metadata.Stars
	}

	if dryRunMetadata {
		fmt.Printf("[DRY RUN] %s: stars %d -> %d\n",
			sf.ServerJSON.Name, ext.Metadata.Stars, newStars)
		return nil
	}

	fmt.Printf("Updating %s: stars %d -> %d\n",
		sf.ServerJSON.Name, ext.Metadata.Stars, newStars)

	ext.Metadata.Stars = newStars
	ext.Metadata.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	return sf.UpdateExtensions(ext)
}

func runBatchUpdateMetadata() error {
	scanner := serverjson.NewScanner(registriesDir)
	entries, err := scanner.FindOldestServers(oldestCount)
	if err != nil {
		return fmt.Errorf("failed to scan registries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No server.json entries found")
		return nil
	}

	fmt.Printf("Updating %d oldest entries:\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("  %d. %s (last updated: %s)\n", i+1, entry.Path, formatLastUpdated(entry.LastUpdated))
	}
	fmt.Println()

	var errs []error
	for _, entry := range entries {
		if err := runSingleUpdateMetadata(entry.Path); err != nil {
			fmt.Printf("Error updating %s: %v\n", entry.Path, err)
			errs = append(errs, fmt.Errorf("%s: %w", entry.Path, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update %d/%d entries", len(errs), len(entries))
	}

	return nil
}

func formatLastUpdated(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}

func verifyServerProvenance(
	sf *serverjson.ServerFile,
	ext *toolhiveRegistry.ServerExtensions,
) error {
	if ext.Provenance == nil {
		if verbose {
			fmt.Printf("No provenance information, skipping verification\n")
		}
		return nil
	}

	if !sf.IsPackageServer() {
		return fmt.Errorf("provenance verification is only supported for package servers")
	}

	imgMeta, err := converters.ServerJSONToImageMetadata(&sf.ServerJSON)
	if err != nil {
		return fmt.Errorf("failed to convert for verification: %w", err)
	}

	v, err := verifier.New(imgMeta)
	if err != nil {
		return fmt.Errorf("failed to create verifier: %w", err)
	}

	isVerified, err := v.VerifyServer(imgMeta.Image, imgMeta)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if !isVerified {
		return fmt.Errorf("no verified signatures found")
	}

	if verbose {
		fmt.Println("Provenance verified successfully")
	}

	return nil
}
