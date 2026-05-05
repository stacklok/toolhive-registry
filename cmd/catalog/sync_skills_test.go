package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSkillJSON = `{
  "namespace": "io.github.stacklok",
  "name": "security-review",
  "title": "Sentry Security Review Skill",
  "description": "Security code review for vulnerabilities.",
  "version": "0.1.0",
  "status": "active",
  "license": "Apache-2.0",
  "repository": {
    "url": "https://github.com/getsentry/skills",
    "type": "git"
  },
  "icons": [
    {
      "src": "icon.svg",
      "type": "image/svg+xml"
    }
  ],
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/stacklok/dockyard/skills/security-review:0.1.0"
    },
    {
      "registryType": "git",
      "url": "https://github.com/getsentry/skills",
      "ref": "94ea2a26c70f3f646f07a613ffe5cd3d4eca1955",
      "subfolder": "skills/security-review"
    }
  ]
}
`

const fixtureNonDockyardSkillJSON = `{
  "namespace": "io.github.stacklok",
  "name": "skill-creator",
  "title": "Skill Creator",
  "description": "Catalog-hosted skill.",
  "version": "0.1.0",
  "packages": [
    {
      "registryType": "git",
      "url": "https://github.com/stacklok/toolhive-catalog",
      "ref": "main",
      "subfolder": "registries/toolhive/skills/skill-creator/skill"
    }
  ]
}
`

const (
	originalRef = "94ea2a26c70f3f646f07a613ffe5cd3d4eca1955"
	bumpedRef   = "f2cff985bcec174fcd096db65a23f06ec9bdde29"
)

// dockyardSpecYAML returns a spec.yaml body with the given fields. The skill
// name is fixed to "security-review" because all current tests share that
// fixture.
func dockyardSpecYAML(repo, ref, path, version string) string {
	return fmt.Sprintf(`metadata:
  name: security-review
  description: "test"

spec:
  repository: %q
  ref: %q
  path: %q
  version: %q
`, repo, ref, path, version)
}

// fakeDockyard returns an httptest.Server that serves spec.yaml files for the
// given map of skill name -> body. Skills missing from the map yield 404.
func fakeDockyard(t *testing.T, specs map[string]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/main/skills/", func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/main/skills/"
		rest := strings.TrimPrefix(r.URL.Path, prefix)
		name, _, ok := strings.Cut(rest, "/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, exists := specs[name]
		if !exists {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(body))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func optsFor(srv *httptest.Server, dryRun bool) syncOptions {
	return syncOptions{
		BaseURL:     srv.URL,
		DockyardRef: "main",
		DryRun:      dryRun,
		HTTP:        srv.Client(),
	}
}

func writeSkill(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "skill.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

func TestSyncOneSkill_TagOnly(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			originalRef, "skills/security-review", "0.2.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusUpdated {
		t.Fatalf("expected updated, got %s (%s)", res.Status, res.Detail)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), ":0.2.0") {
		t.Errorf("new tag not present: %s", got)
	}
	if !strings.Contains(string(got), originalRef) {
		t.Errorf("ref should not have changed")
	}
}

func TestSyncOneSkill_RefOnly(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "skills/security-review", "0.1.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusUpdated {
		t.Fatalf("expected updated, got %s (%s)", res.Status, res.Detail)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), bumpedRef) {
		t.Errorf("new ref not present in file")
	}
	if !strings.Contains(string(got), ":0.1.0") {
		t.Errorf("tag should not have changed")
	}
}

func TestSyncOneSkill_Both(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "skills/security-review", "0.2.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusUpdated {
		t.Fatalf("expected updated, got %s (%s)", res.Status, res.Detail)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), ":0.2.0") || !strings.Contains(string(got), bumpedRef) {
		t.Errorf("expected both updates in file:\n%s", got)
	}
}

func TestSyncOneSkill_NoOp(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			originalRef, "skills/security-review", "0.1.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	before, _ := os.ReadFile(path)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusNoop {
		t.Fatalf("expected noop, got %s (%s)", res.Status, res.Detail)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("file mutated despite noop result")
	}
}

func TestSyncOneSkill_DryRun(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "skills/security-review", "0.2.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	before, _ := os.ReadFile(path)
	res := syncOneSkill(context.Background(), optsFor(srv, true), path)
	if res.Status != statusNoop {
		t.Fatalf("expected noop in dry-run, got %s (%s)", res.Status, res.Detail)
	}
	if res.IDAfter == res.IDBefore && res.RefAfter == res.RefBefore {
		t.Errorf("expected dry-run to record a planned change")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("dry-run mutated the file on disk")
	}
}

func TestSyncOneSkill_RepoMismatchSkips(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/example/other-skills",
			bumpedRef, "skills/security-review", "0.2.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	before, _ := os.ReadFile(path)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusSkipped {
		t.Fatalf("expected skipped, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "repository mismatch") {
		t.Errorf("expected repository mismatch detail, got %q", res.Detail)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("file mutated despite skipped result")
	}
}

func TestSyncOneSkill_SubfolderMismatchSkips(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "different/path", "0.2.0",
		),
	})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusSkipped {
		t.Fatalf("expected skipped, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "subfolder mismatch") {
		t.Errorf("expected subfolder mismatch detail, got %q", res.Detail)
	}
}

func TestSyncOneSkill_Missing404(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusMissing {
		t.Fatalf("expected missing, got %s (%s)", res.Status, res.Detail)
	}
}

func TestSyncOneSkill_NoDockyardPackageSkips(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{})

	path := writeSkill(t, fixtureNonDockyardSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusSkipped {
		t.Fatalf("expected skipped, got %s (%s)", res.Status, res.Detail)
	}
}

func TestSyncOneSkill_MalformedYAMLErrors(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": "spec:\n  version: [unterminated",
	})

	path := writeSkill(t, fixtureSkillJSON)
	res := syncOneSkill(context.Background(), optsFor(srv, false), path)
	if res.Status != statusError {
		t.Fatalf("expected error, got %s (%s)", res.Status, res.Detail)
	}
}

func TestFindDockyardSkillFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	mustWrite := func(rel, body string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(body), 0600); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	mustWrite("toolhive/skills/security-review/skill.json", fixtureSkillJSON)
	mustWrite("toolhive/skills/skill-creator/skill.json", fixtureNonDockyardSkillJSON)
	mustWrite(".hidden/skills/foo/skill.json", fixtureSkillJSON)
	mustWrite("official/servers/foo/server.json", "{}")

	paths, err := findDockyardSkillFiles(root)
	if err != nil {
		t.Fatalf("findDockyardSkillFiles: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 dockyard skill, got %d (%v)", len(paths), paths)
	}
	if !strings.HasSuffix(paths[0], "security-review/skill.json") {
		t.Errorf("unexpected path: %s", paths[0])
	}
}

func TestRunSyncSkillsAllReportsErrorExit(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"security-review": "spec:\n  version: [unterminated",
	})

	root := t.TempDir()
	dir := filepath.Join(root, "toolhive", "skills", "security-review")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(fixtureSkillJSON), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := runSyncSkillsAll(context.Background(), optsFor(srv, false), root); err == nil {
		t.Fatal("expected error from --all when a skill fails")
	}
}

func TestRepoURLsMatch(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b string
		want bool
	}{
		{"https://github.com/foo/bar", "https://github.com/foo/bar", true},
		{"https://github.com/foo/bar.git", "https://github.com/foo/bar", true},
		{"https://github.com/foo/bar/", "https://github.com/foo/bar.git", true},
		{"https://github.com/foo/bar", "https://github.com/foo/baz", false},
		{"", "https://github.com/foo/bar", false},
		{"https://github.com/foo/bar", "", false},
	}
	for _, tc := range cases {
		if got := repoURLsMatch(tc.a, tc.b); got != tc.want {
			t.Errorf("repoURLsMatch(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
