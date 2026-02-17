package serverjson

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// makeServerJSON returns a minimal server.json with the given last_updated value.
// If lastUpdated is empty, the _meta section is omitted entirely.
func makeServerJSON(lastUpdated string) string {
	if lastUpdated == "" {
		return testNoMetaServerJSON
	}
	return `{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "test",
  "description": "test",
  "version": "1.0.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/test/img:v1",
      "transport": { "type": "stdio" }
    }
  ],
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "io.github.stacklok": {
        "ghcr.io/test/img:v1": {
          "status": "Active",
          "metadata": {
            "last_updated": "` + lastUpdated + `"
          }
        }
      }
    }
  }
}`
}

// setupRegistriesDir creates a registries directory structure for testing:
//
//	registriesDir/
//	  toolhive/
//	    servers/
//	      <name>/server.json  (for each entry in servers map)
func setupRegistriesDir(t *testing.T, servers map[string]string) string {
	t.Helper()
	root := t.TempDir()
	serversDir := filepath.Join(root, "toolhive", "servers")

	for name, content := range servers {
		dir := filepath.Join(serversDir, name)
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "server.json"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func TestFindOldestServers_BasicSort(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"newest": makeServerJSON("2026-03-01T00:00:00Z"),
		"middle": makeServerJSON("2026-02-01T00:00:00Z"),
		"oldest": makeServerJSON("2026-01-01T00:00:00Z"),
	})

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(2)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Oldest should be first
	expectTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !entries[0].LastUpdated.Equal(expectTime) {
		t.Errorf("expected oldest entry first (2026-01-01), got %s", entries[0].LastUpdated)
	}

	expectTime2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !entries[1].LastUpdated.Equal(expectTime2) {
		t.Errorf("expected second oldest entry (2026-02-01), got %s", entries[1].LastUpdated)
	}
}

func TestFindOldestServers_EmptyLastUpdated(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"with-date": makeServerJSON("2026-06-01T00:00:00Z"),
		"no-meta":   makeServerJSON(""), // no _meta → epoch
	})

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(2)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Entry with no _meta should be treated as epoch → selected first
	if !entries[0].LastUpdated.IsZero() {
		t.Errorf("expected zero time (epoch) for entry without _meta, got %s", entries[0].LastUpdated)
	}
}

func TestFindOldestServers_CountExceedsTotal(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"alpha": makeServerJSON("2026-01-01T00:00:00Z"),
		"beta":  makeServerJSON("2026-02-01T00:00:00Z"),
		"gamma": makeServerJSON("2026-03-01T00:00:00Z"),
	})

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(10)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (all available), got %d", len(entries))
	}
}

func TestFindOldestServers_SkipsHiddenDirs(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"visible": makeServerJSON("2026-01-01T00:00:00Z"),
	})

	// Add a hidden server directory
	hiddenDir := filepath.Join(root, "toolhive", "servers", ".hidden")
	if err := os.MkdirAll(hiddenDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(hiddenDir, "server.json"),
		[]byte(makeServerJSON("2025-01-01T00:00:00Z")),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	// Also add a hidden registry directory
	hiddenRegDir := filepath.Join(root, ".hidden-registry", "servers", "test")
	if err := os.MkdirAll(hiddenRegDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(hiddenRegDir, "server.json"),
		[]byte(makeServerJSON("2024-01-01T00:00:00Z")),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(10)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (hidden dirs skipped), got %d", len(entries))
	}
}

func TestFindOldestServers_SkipsMalformedJSON(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"good": makeServerJSON("2026-01-01T00:00:00Z"),
	})

	// Add a malformed server.json
	badDir := filepath.Join(root, "toolhive", "servers", "bad")
	if err := os.MkdirAll(badDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(badDir, "server.json"),
		[]byte("{invalid json"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(10)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (malformed skipped), got %d", len(entries))
	}
}

func TestFindOldestServers_InvalidCount(t *testing.T) {
	t.Parallel()

	root := setupRegistriesDir(t, map[string]string{
		"test": makeServerJSON("2026-01-01T00:00:00Z"),
	})

	scanner := NewScanner(root)

	_, err := scanner.FindOldestServers(0)
	if err == nil {
		t.Error("expected error for count=0")
	}

	_, err = scanner.FindOldestServers(-1)
	if err == nil {
		t.Error("expected error for negative count")
	}
}

func TestFindOldestServers_EmptyRegistries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	scanner := NewScanner(root)
	entries, err := scanner.FindOldestServers(5)
	if err != nil {
		t.Fatalf("FindOldestServers failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty dir, got %d", len(entries))
	}
}
