// Package main is the entry point for the regup command
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stacklok/toolhive/pkg/container/verifier"
	"github.com/stacklok/toolhive/pkg/logger"
	"github.com/stacklok/toolhive/pkg/registry"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/toolhive-registry/pkg/types"
)

var (
	specPath         string
	dryRun           bool
	githubToken      string
	verifyProvenance bool
)

type serverWithName struct {
	name  string
	path  string
	entry *types.RegistryEntry
}

// ProvenanceVerificationError represents an error during provenance verification
type ProvenanceVerificationError struct {
	ServerName string
	Reason     string
}

func (e *ProvenanceVerificationError) Error() string {
	return fmt.Sprintf("provenance verification failed for server %s: %s", e.ServerName, e.Reason)
}

var rootCmd = &cobra.Command{
	Use:   "regup [spec-file]",
	Short: "Update a single MCP server registry entry with latest information",
	Long: `regup is a utility for updating a single MCP server registry entry with the latest information.
It updates the GitHub stars and pulls data for the specified spec.yaml file.
This tool is designed to be run by Renovate when updating image versions.`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	// Initialize the logger system
	logger.Initialize()

	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry run without making changes")
	rootCmd.Flags().StringVarP(&githubToken, "github-token", "t", "",
		"GitHub token for API authentication (can also be set via GITHUB_TOKEN env var)")
	rootCmd.Flags().BoolVar(&verifyProvenance, "verify-provenance", false,
		"Verify provenance information and fail if verification fails")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

func runUpdate(_ *cobra.Command, args []string) error {
	specPath = args[0]

	// If token not provided via flag, check environment variable
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	// Load the single spec file
	server, err := loadSpec(specPath)
	if err != nil {
		return fmt.Errorf("failed to load spec file: %w", err)
	}

	// Update the server
	if err := updateServerInfo(server); err != nil {
		var provenanceErr *ProvenanceVerificationError
		if errors.As(err, &provenanceErr) {
			return fmt.Errorf("provenance verification failed: %w", err)
		}
		return fmt.Errorf("failed to update server: %w", err)
	}

	if dryRun {
		logger.Info("Dry run completed, no changes made")
	} else {
		logger.Infof("Successfully updated %s", server.name)
	}

	return nil
}

func loadSpec(path string) (serverWithName, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return serverWithName{}, fmt.Errorf("spec file not found: %s", path)
	}

	// Read the spec file
	data, err := os.ReadFile(path) // #nosec G304 - file path is constructed from known directory
	if err != nil {
		return serverWithName{}, fmt.Errorf("failed to read spec file: %w", err)
	}

	// Parse YAML into our RegistryEntry type
	var entry types.RegistryEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return serverWithName{}, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract server name from path (parent directory name)
	dir := filepath.Dir(path)
	name := filepath.Base(dir)

	// Set the name if not already set
	if entry.GetName() == "" {
		entry.SetName(name)
	}

	return serverWithName{
		name:  name,
		path:  path,
		entry: &entry,
	}, nil
}

func updateServerInfo(server serverWithName) error {
	// Verify provenance if requested
	if verifyProvenance {
		if err := verifyServerProvenance(server); err != nil {
			return &ProvenanceVerificationError{
				ServerName: server.name,
				Reason:     err.Error(),
			}
		}
	}

	repoURL, metadata, err := getServerMetadata(server)
	if err != nil {
		return err
	}

	currentStars := metadata.Stars
	currentPulls := metadata.Pulls

	newStars := getUpdatedStars(repoURL, currentStars, server.name)
	newPulls := getUpdatedPulls(server, currentPulls)

	return updateServerMetadata(server, currentStars, newStars, currentPulls, newPulls)
}

func getServerMetadata(server serverWithName) (string, *registry.Metadata, error) {
	var repoURL string
	var metadata *registry.Metadata

	if server.entry.IsImage() && server.entry.ImageMetadata != nil {
		repoURL = server.entry.ImageMetadata.RepositoryURL
		if server.entry.ImageMetadata.Metadata == nil {
			server.entry.ImageMetadata.Metadata = &registry.Metadata{}
		}
		metadata = server.entry.ImageMetadata.Metadata
	} else if server.entry.IsRemote() && server.entry.RemoteServerMetadata != nil {
		repoURL = server.entry.RemoteServerMetadata.RepositoryURL
		if server.entry.RemoteServerMetadata.Metadata == nil {
			server.entry.RemoteServerMetadata.Metadata = &registry.Metadata{}
		}
		metadata = server.entry.RemoteServerMetadata.Metadata
	} else {
		return "", nil, fmt.Errorf("unable to determine server type for %s", server.name)
	}

	if repoURL == "" {
		logger.Warnf("Server %s has no repository URL, skipping GitHub stars update", server.name)
	}

	return repoURL, metadata, nil
}

func getUpdatedStars(repoURL string, currentStars int, serverName string) int {
	if repoURL == "" {
		return currentStars
	}

	owner, repo, err := extractOwnerRepo(repoURL)
	if err != nil {
		logger.Warnf("Failed to extract owner/repo from URL %s: %v", repoURL, err)
		return currentStars
	}

	// Get repository info from GitHub API
	stars, _, err := getGitHubRepoInfo(owner, repo, serverName, currentStars)
	if err != nil {
		logger.Warnf("Failed to get GitHub repo info for %s: %v", serverName, err)
		return currentStars
	}

	return stars
}

func getUpdatedPulls(server serverWithName, currentPulls int) int {
	if !server.entry.IsImage() || server.entry.ImageMetadata == nil || server.entry.Image == "" {
		return currentPulls
	}

	pullCount, err := getContainerPullCount(server.entry.Image)
	if err != nil {
		logger.Warnf("Failed to get pull count for image %s: %v", server.entry.Image, err)
		return currentPulls
	}

	if pullCount > 0 {
		return pullCount
	}

	// No pull count available (GHCR or private registry)
	return currentPulls
}

func updateServerMetadata(server serverWithName, currentStars, newStars, currentPulls, newPulls int) error {
	if dryRun {
		logger.Infof("[DRY RUN] Would update %s: stars %d -> %d, pulls %d -> %d",
			server.name, currentStars, newStars, currentPulls, newPulls)
		return nil
	}

	// Log the changes
	logger.Infof("Updating %s: stars %d -> %d, pulls %d -> %d",
		server.name, currentStars, newStars, currentPulls, newPulls)

	// Use yaml.v3 Node API to preserve comments and structure
	return updateYAMLPreservingStructure(server.path, newStars, newPulls)
}

// updateYAMLPreservingStructure updates the YAML file while preserving comments and structure
func updateYAMLPreservingStructure(path string, stars, pulls int) error {
	// Read the original file
	data, err := os.ReadFile(path) // #nosec G304 - file path is constructed from known directory
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse with yaml.v3 to preserve structure
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Update the metadata fields
	if err := updateMetadataInNode(&doc, stars, pulls); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// Marshal back preserving structure
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(4)
	if err := encoder.Encode(&doc); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	// Write back to file
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// updateMetadataInNode updates metadata fields in the YAML node tree
func updateMetadataInNode(node *yaml.Node, stars, pulls int) error {
	// Navigate to the document content
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return updateMetadataInNode(node.Content[0], stars, pulls)
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	// Find or create metadata section
	metadataIndex := -1
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "metadata" {
			metadataIndex = i
			break
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)

	if metadataIndex >= 0 {
		// Update existing metadata
		metadataNode := node.Content[metadataIndex+1]
		if metadataNode.Kind != yaml.MappingNode {
			return fmt.Errorf("metadata is not a mapping")
		}

		// Update or add fields
		updated := map[string]bool{
			"stars":       false,
			"pulls":       false,
			"lastupdated": false,
		}

		for i := 0; i < len(metadataNode.Content); i += 2 {
			key := metadataNode.Content[i].Value
			switch key {
			case "stars":
				metadataNode.Content[i+1].Value = fmt.Sprintf("%d", stars)
				updated["stars"] = true
			case "pulls":
				metadataNode.Content[i+1].Value = fmt.Sprintf("%d", pulls)
				updated["pulls"] = true
			case "lastupdated":
				metadataNode.Content[i+1].Value = now
				updated["lastupdated"] = true
			}
		}

		// Add missing fields
		if !updated["stars"] {
			metadataNode.Content = append(metadataNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "stars"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", stars)})
		}
		if !updated["pulls"] {
			metadataNode.Content = append(metadataNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "pulls"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", pulls)})
		}
		if !updated["lastupdated"] {
			metadataNode.Content = append(metadataNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "lastupdated"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: now})
		}
	} else {
		// Add new metadata section
		metadataKey := &yaml.Node{Kind: yaml.ScalarNode, Value: "metadata"}
		metadataValue := &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "stars"},
				{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", stars)},
				{Kind: yaml.ScalarNode, Value: "pulls"},
				{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", pulls)},
				{Kind: yaml.ScalarNode, Value: "lastupdated"},
				{Kind: yaml.ScalarNode, Value: now},
			},
		}
		node.Content = append(node.Content, metadataKey, metadataValue)
	}

	return nil
}

// verifyServerProvenance verifies the provenance information for a server
func verifyServerProvenance(server serverWithName) error {
	// Check if provenance information exists
	if server.entry.Provenance == nil {
		logger.Warnf("Server %s has no provenance information, skipping verification", server.name)
		return nil
	}

	// Get image reference
	if server.entry.Image == "" {
		return fmt.Errorf("no image reference provided")
	}

	logger.Infof("Verifying provenance for server %s with image %s", server.name, server.entry.Image)

	// The entry already has ImageMetadata embedded, so we can use it directly
	v, err := verifier.New(server.entry.ImageMetadata)
	if err != nil {
		return fmt.Errorf("failed to create verifier: %w", err)
	}

	// Get verification results
	isVerified, err := v.VerifyServer(server.entry.Image, server.entry.ImageMetadata)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Check if we have valid verification results
	if isVerified {
		logger.Infof("Server %s verified successfully", server.name)
		return nil
	}

	return fmt.Errorf("no verified signatures found")
}

// extractOwnerRepo extracts the owner and repo from a GitHub repository URL
func extractOwnerRepo(url string) (string, string, error) {
	// Remove trailing .git if present
	url = strings.TrimSuffix(url, ".git")

	// Handle different GitHub URL formats
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
	}

	// The owner and repo should be the last two parts
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	return owner, repo, nil
}

// getGitHubRepoInfo gets the stars count for a GitHub repository
func getGitHubRepoInfo(owner, repo, _ string, currentPulls int) (stars int, pulls int, err error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	if githubToken != "" {
		req.Header.Add("Authorization", "token "+githubToken)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("GitHub API returned %s: %s", resp.Status, string(body))
	}

	// Parse response
	var repoInfo struct {
		StargazersCount int `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return 0, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Return current pulls - we'll fetch container pulls separately
	return repoInfo.StargazersCount, currentPulls, nil
}

// getContainerPullCount fetches the pull count for a container image
func getContainerPullCount(image string) (int, error) {
	// Parse the image reference
	parts := strings.Split(image, ":")
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid image format: %s", image)
	}

	imageName := parts[0]

	// Determine registry and fetch accordingly
	if strings.HasPrefix(imageName, "ghcr.io/") {
		return getGHCRPullCount(imageName)
	} else if strings.Contains(imageName, "/") && !strings.Contains(imageName, ".") {
		// Likely Docker Hub (no dots in the hostname part)
		return getDockerHubPullCount(imageName)
	}

	// Unknown registry, return 0
	logger.Warnf("Unknown registry for image %s, cannot fetch pull count", image)
	return 0, nil
}

// getGHCRPullCount fetches pull count for GitHub Container Registry images
func getGHCRPullCount(imageName string) (int, error) {
	// GHCR requires authentication to get package statistics
	if githubToken == "" {
		logger.Debugf("No GitHub token available, cannot fetch GHCR pull count for %s", imageName)
		return 0, nil
	}

	owner, packageName, err := parseGHCRImageName(imageName)
	if err != nil {
		return 0, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url, err := fetchGHCRPackageInfo(client, owner, packageName)
	if err != nil {
		return 0, err
	}

	return fetchGHCRVersions(client, url, imageName)
}

func parseGHCRImageName(imageName string) (string, string, error) {
	// Parse the image name: ghcr.io/owner/repo/package or ghcr.io/owner/package
	imageName = strings.TrimPrefix(imageName, "ghcr.io/")
	parts := strings.Split(imageName, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GHCR image format: %s", imageName)
	}

	owner := parts[0]
	// The package name is everything after the owner
	packageName := strings.Join(parts[1:], "/")
	return owner, packageName, nil
}

func fetchGHCRPackageInfo(client *http.Client, owner, packageName string) (string, error) {
	// GitHub Packages API endpoint for container packages
	url := fmt.Sprintf("https://api.github.com/users/%s/packages/container/%s", owner, packageName)

	resp, err := makeGHCRRequest(client, url)
	if err != nil {
		// Try org endpoint if user endpoint fails
		url = fmt.Sprintf("https://api.github.com/orgs/%s/packages/container/%s", owner, packageName)
		resp, err = makeGHCRRequest(client, url)
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound && strings.Contains(url, "/users/") {
		// Try org endpoint if user endpoint returned 404
		url = strings.Replace(url, "/users/", "/orgs/", 1)
		resp, err = makeGHCRRequest(client, url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		logger.Debugf("Could not fetch GHCR package stats (status %d)", resp.StatusCode)
		return "", fmt.Errorf("package not found or no access")
	}

	var packageInfo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return url, nil
}

func makeGHCRRequest(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", "token "+githubToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func fetchGHCRVersions(client *http.Client, baseURL, imageName string) (int, error) {
	versionsURL := fmt.Sprintf("%s/versions?per_page=100", baseURL)
	resp, err := makeGHCRRequest(client, versionsURL)
	if err != nil {
		return 0, fmt.Errorf("failed to create versions request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Debugf("Could not fetch GHCR package versions (status %d) for %s", resp.StatusCode, imageName)
		return 0, nil
	}

	var versions []struct {
		Metadata struct {
			Container struct {
				Tags []string `json:"tags"`
			} `json:"container"`
		} `json:"metadata"`
		// Unfortunately, GitHub API doesn't expose download_count for container packages
		// in the same way it does for other package types
	}

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return 0, fmt.Errorf("failed to parse versions response: %w", err)
	}

	// GitHub doesn't expose container download counts through the API
	// even with authentication. This is a known limitation.
	// Return 0 to indicate we couldn't get the data
	logger.Debugf("GHCR package found but download count not available through API for %s", imageName)
	return 0, nil
}

// getDockerHubPullCount fetches pull count for Docker Hub images
func getDockerHubPullCount(imageName string) (int, error) {
	// Remove docker.io prefix if present
	imageName = strings.TrimPrefix(imageName, "docker.io/")

	// Docker Hub API endpoint
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/", imageName)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Not found or error - return 0
		return 0, nil
	}

	var dockerHubResp struct {
		PullCount int `json:"pull_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dockerHubResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return dockerHubResp.PullCount, nil
}
