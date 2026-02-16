// Package metadata provides utilities for fetching server metadata
// (GitHub stars) from external APIs.
package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is an interface for making HTTP requests, enabling testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Fetcher retrieves metadata from external APIs.
type Fetcher struct {
	client      HTTPClient
	githubToken string
}

// NewFetcher creates a new Fetcher with the given HTTP client and optional GitHub token.
func NewFetcher(client HTTPClient, githubToken string) *Fetcher {
	return &Fetcher{
		client:      client,
		githubToken: githubToken,
	}
}

// FetchStars returns the stargazers_count for a GitHub repository.
// The repoURL should be a GitHub repository URL like "https://github.com/owner/repo".
// Returns (0, nil) if the URL is empty or not a GitHub URL.
func (f *Fetcher) FetchStars(ctx context.Context, repoURL string) (int, error) {
	if repoURL == "" {
		return 0, nil
	}

	if !strings.Contains(repoURL, "github.com") {
		return 0, nil
	}

	owner, repo, err := extractOwnerRepo(repoURL)
	if err != nil {
		return 0, fmt.Errorf("failed to parse repo URL: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if f.githubToken != "" {
		req.Header.Set("Authorization", "token "+f.githubToken)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf(
			"GitHub API returned %s: %s", resp.Status, string(body),
		)
	}

	var repoInfo struct {
		StargazersCount int `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return repoInfo.StargazersCount, nil
}

// extractOwnerRepo extracts the owner and repo from a GitHub repository URL.
func extractOwnerRepo(repoURL string) (string, string, error) {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", repoURL)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	return owner, repo, nil
}
