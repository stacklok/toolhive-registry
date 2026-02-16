package metadata

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type mockHTTPClient struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.body)),
	}, nil
}

func TestFetchStars_Success(t *testing.T) {
	t.Parallel()
	mock := &mockHTTPClient{
		statusCode: http.StatusOK,
		body:       `{"stargazers_count": 42}`,
	}
	fetcher := NewFetcher(mock, "test-token")

	stars, err := fetcher.FetchStars(context.Background(), "https://github.com/owner/repo")
	if err != nil {
		t.Fatalf("FetchStars failed: %v", err)
	}
	if stars != 42 {
		t.Errorf("expected 42 stars, got %d", stars)
	}
}

func TestFetchStars_EmptyURL(t *testing.T) {
	t.Parallel()
	fetcher := NewFetcher(nil, "")

	stars, err := fetcher.FetchStars(context.Background(), "")
	if err != nil {
		t.Fatalf("FetchStars failed: %v", err)
	}
	if stars != 0 {
		t.Errorf("expected 0 stars for empty URL, got %d", stars)
	}
}

func TestFetchStars_NonGitHubURL(t *testing.T) {
	t.Parallel()
	fetcher := NewFetcher(nil, "")

	stars, err := fetcher.FetchStars(context.Background(), "https://gitlab.com/owner/repo")
	if err != nil {
		t.Fatalf("FetchStars failed: %v", err)
	}
	if stars != 0 {
		t.Errorf("expected 0 stars for non-GitHub URL, got %d", stars)
	}
}

func TestFetchStars_APIError(t *testing.T) {
	t.Parallel()
	mock := &mockHTTPClient{
		statusCode: http.StatusForbidden,
		body:       `{"message": "rate limit exceeded"}`,
	}
	fetcher := NewFetcher(mock, "")

	_, err := fetcher.FetchStars(context.Background(), "https://github.com/owner/repo")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestFetchStars_NetworkError(t *testing.T) {
	t.Parallel()
	mock := &mockHTTPClient{
		err: io.ErrUnexpectedEOF,
	}
	fetcher := NewFetcher(mock, "")

	_, err := fetcher.FetchStars(context.Background(), "https://github.com/owner/repo")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}

func TestFetchStars_InvalidJSON(t *testing.T) {
	t.Parallel()
	mock := &mockHTTPClient{
		statusCode: http.StatusOK,
		body:       `not json`,
	}
	fetcher := NewFetcher(mock, "")

	_, err := fetcher.FetchStars(context.Background(), "https://github.com/owner/repo")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchStars_WithToken(t *testing.T) {
	t.Parallel()
	var capturedReq *http.Request
	mock := &mockHTTPClient{
		statusCode: http.StatusOK,
		body:       `{"stargazers_count": 10}`,
	}

	// Wrap to capture request
	fetcher := NewFetcher(&requestCapture{
		client:  mock,
		lastReq: &capturedReq,
	}, "my-secret-token")

	_, err := fetcher.FetchStars(context.Background(), "https://github.com/owner/repo")
	if err != nil {
		t.Fatalf("FetchStars failed: %v", err)
	}

	if capturedReq == nil {
		t.Fatal("request was not captured")
	}
	auth := capturedReq.Header.Get("Authorization")
	if auth != "token my-secret-token" {
		t.Errorf("expected Authorization header, got %q", auth)
	}
}

func TestExtractOwnerRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url           string
		expectedOwner string
		expectedRepo  string
		expectErr     bool
	}{
		{"https://github.com/owner/repo", "owner", "repo", false},
		{"https://github.com/owner/repo.git", "owner", "repo", false},
		{"https://github.com/org/my-repo", "org", "my-repo", false},
		{"invalid", "", "", true},
	}

	for _, tt := range tests {
		owner, repo, err := extractOwnerRepo(tt.url)
		if tt.expectErr && err == nil {
			t.Errorf("expected error for %s", tt.url)
		}
		if !tt.expectErr && err != nil {
			t.Errorf("unexpected error for %s: %v", tt.url, err)
		}
		if owner != tt.expectedOwner {
			t.Errorf("for %s: expected owner %s, got %s", tt.url, tt.expectedOwner, owner)
		}
		if repo != tt.expectedRepo {
			t.Errorf("for %s: expected repo %s, got %s", tt.url, tt.expectedRepo, repo)
		}
	}
}

// requestCapture wraps an HTTPClient to capture the request.
type requestCapture struct {
	client  HTTPClient
	lastReq **http.Request
}

func (rc *requestCapture) Do(req *http.Request) (*http.Response, error) {
	*rc.lastReq = req
	return rc.client.Do(req)
}
