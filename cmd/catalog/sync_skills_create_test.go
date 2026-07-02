package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureIconSVG = `<svg xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="10"/></svg>
`

const newSkillName = "new-skill"

// seedRegistry creates a registry root containing one existing
// dockyard-sourced skill (security-review) with an icon, and returns the
// root and its skills directory.
func seedRegistry(t *testing.T) (root, skillsDir string) {
	t.Helper()
	root = t.TempDir()
	skillsDir = filepath.Join(root, "toolhive", "skills")
	writeEntry(t, skillsDir, securityReview, fixtureSkillJSON, fixtureIconSVG)
	return root, skillsDir
}

func writeEntry(t *testing.T, skillsDir, name, skillJSON, icon string) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(skillJSON), 0600); err != nil {
		t.Fatalf("write skill.json: %v", err)
	}
	if icon != "" {
		if err := os.WriteFile(filepath.Join(dir, "icon.svg"), []byte(icon), 0600); err != nil {
			t.Fatalf("write icon.svg: %v", err)
		}
	}
}

func TestCreateMissingSkills_CreatesEntry(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		securityReview: dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			originalRef, "skills/security-review", "0.1.0",
		),
		newSkillName: dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/new-skill", "0.2.0",
		),
	})

	root, skillsDir := seedRegistry(t)
	results, err := createMissingSkills(context.Background(), optsFor(srv, false), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d (%+v)", len(results), results)
	}
	res := results[0]
	if res.Status != statusCreated || res.Op != opCreate {
		t.Fatalf("expected created/create, got %s/%s (%s)", res.Status, res.Op, res.Detail)
	}
	if !strings.Contains(res.Detail, "placeholder icon") {
		t.Errorf("expected placeholder icon detail, got %q", res.Detail)
	}

	const wantSkillJSON = `{
  "namespace": "io.github.stacklok",
  "name": "new-skill",
  "title": "New Skill",
  "description": "test",
  "version": "0.2.0",
  "status": "active",
  "repository": {
    "url": "https://github.com/foo/bar",
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
      "identifier": "ghcr.io/stacklok/dockyard/skills/new-skill:0.2.0"
    },
    {
      "registryType": "git",
      "url": "https://github.com/foo/bar",
      "ref": "f2cff985bcec174fcd096db65a23f06ec9bdde29",
      "subfolder": "skills/new-skill"
    }
  ]
}
`
	got, err := os.ReadFile(filepath.Join(skillsDir, newSkillName, "skill.json"))
	if err != nil {
		t.Fatalf("read created skill.json: %v", err)
	}
	if string(got) != wantSkillJSON {
		t.Errorf("skill.json mismatch:\ngot:\n%s\nwant:\n%s", got, wantSkillJSON)
	}
	icon, err := os.ReadFile(filepath.Join(skillsDir, newSkillName, "icon.svg"))
	if err != nil {
		t.Fatalf("read created icon.svg: %v", err)
	}
	if string(icon) != placeholderIconSVG {
		t.Errorf("expected placeholder icon, got:\n%s", icon)
	}
}

func TestCreateMissingSkills_CopiesSiblingIcon(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		securityReview: dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			originalRef, "skills/security-review", "0.1.0",
		),
		"agents-md": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "skills/agents-md", "0.1.1",
		),
	})

	root, skillsDir := seedRegistry(t)
	results, err := createMissingSkills(context.Background(), optsFor(srv, false), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 1 || results[0].Status != statusCreated {
		t.Fatalf("expected 1 created result, got %+v", results)
	}
	if want := "icon copied from " + securityReview + "; license needs review"; results[0].Detail != want {
		t.Errorf("expected detail %q, got %q", want, results[0].Detail)
	}
	icon, err := os.ReadFile(filepath.Join(skillsDir, "agents-md", "icon.svg"))
	if err != nil {
		t.Fatalf("read created icon.svg: %v", err)
	}
	if string(icon) != fixtureIconSVG {
		t.Errorf("expected sibling icon copied, got:\n%s", icon)
	}
}

func TestFindSiblingIcon_FirstSortedWins(t *testing.T) {
	t.Parallel()
	skillsDir := filepath.Join(t.TempDir(), "skills")
	secondIcon := strings.Replace(fixtureIconSVG, "circle", "rect", 1)
	otherSkill := strings.Replace(fixtureSkillJSON, `"name": "security-review"`, `"name": "a-first"`, 1)
	writeEntry(t, skillsDir, "security-review", fixtureSkillJSON, fixtureIconSVG)
	writeEntry(t, skillsDir, "a-first", otherSkill, secondIcon)

	icon, detail, err := findSiblingIcon(skillsDir, "https://github.com/getsentry/skills")
	if err != nil {
		t.Fatalf("findSiblingIcon: %v", err)
	}
	if string(icon) != secondIcon {
		t.Errorf("expected icon from alphabetically first sibling, got:\n%s", icon)
	}
	if detail != "icon copied from a-first" {
		t.Errorf("unexpected detail %q", detail)
	}
}

func TestCreateMissingSkills_SkipsExistingDir(t *testing.T) {
	t.Parallel()
	// Dockyard's code-review is a different skill from the catalog-hosted
	// code-review entry; an existing directory must never be overwritten.
	srv := fakeDockyard(t, map[string]string{
		"code-review": dockyardSpecYAML(
			"https://github.com/getsentry/skills",
			bumpedRef, "skills/code-review", "0.1.0",
		),
	})

	// The registry participates in dockyard syncing via security-review.
	root, skillsDir := seedRegistry(t)
	writeEntry(t, skillsDir, "code-review", fixtureNonDockyardSkillJSON, fixtureIconSVG)
	before, _ := os.ReadFile(filepath.Join(skillsDir, "code-review", "skill.json"))

	results, err := createMissingSkills(context.Background(), optsFor(srv, false), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %+v", results)
	}
	after, _ := os.ReadFile(filepath.Join(skillsDir, "code-review", "skill.json"))
	if string(before) != string(after) {
		t.Errorf("existing entry was modified")
	}
}

func TestCreateMissingSkills_SkipsNonParticipatingRegistry(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		newSkillName: dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/new-skill", "0.2.0",
		),
	})

	// One registry with no dockyard-sourced skills and one with an empty
	// skills directory: neither has opted into dockyard syncing, so no
	// entries are onboarded.
	root := t.TempDir()
	writeEntry(t, filepath.Join(root, "toolhive", "skills"), "code-review", fixtureNonDockyardSkillJSON, fixtureIconSVG)
	if err := os.MkdirAll(filepath.Join(root, "official", "skills"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	results, err := createMissingSkills(context.Background(), optsFor(srv, false), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %+v", results)
	}
	for _, reg := range []string{"toolhive", "official"} {
		if _, err := os.Stat(filepath.Join(root, reg, "skills", newSkillName)); !os.IsNotExist(err) {
			t.Errorf("skill was created in non-participating registry %s", reg)
		}
	}
}

func TestCreateMissingSkills_DryRunWritesNothing(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		newSkillName: dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/new-skill", "0.2.0",
		),
	})

	root, skillsDir := seedRegistry(t)
	results, err := createMissingSkills(context.Background(), optsFor(srv, true), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %+v", results)
	}
	res := results[0]
	if res.Status != statusNoop || !res.Planned || res.Op != opCreate {
		t.Fatalf("expected planned noop create, got %+v", res)
	}
	if !strings.Contains(res.Detail, "would create") {
		t.Errorf("expected would-create detail, got %q", res.Detail)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, newSkillName)); !os.IsNotExist(err) {
		t.Errorf("dry-run created a directory on disk")
	}
}

func TestCreateMissingSkills_IgnoreFile(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"unwanted-skill": dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/unwanted-skill", "0.2.0",
		),
	})

	root, skillsDir := seedRegistry(t)
	ignore := "# consolidated upstream; catalog keeps split entries\nunwanted-skill\n"
	if err := os.WriteFile(filepath.Join(skillsDir, syncIgnoreFile), []byte(ignore), 0600); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	results, err := createMissingSkills(context.Background(), optsFor(srv, false), root)
	if err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected ignored skill to be skipped, got %+v", results)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "unwanted-skill")); !os.IsNotExist(err) {
		t.Errorf("ignored skill was created")
	}
}

func TestCreateSkillEntry_MissingVersionErrors(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		newSkillName: dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/new-skill", "",
		),
	})

	_, skillsDir := seedRegistry(t)
	res := createSkillEntry(context.Background(), optsFor(srv, false), skillsDir, newSkillName)
	if res.Status != statusError {
		t.Fatalf("expected error, got %s (%s)", res.Status, res.Detail)
	}
	if !strings.Contains(res.Detail, "spec.version") {
		t.Errorf("expected spec.version in detail, got %q", res.Detail)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, newSkillName)); !os.IsNotExist(err) {
		t.Errorf("failed create left a directory behind")
	}
}

func TestCreateSkillEntry_MissingSpecIsNonFatal(t *testing.T) {
	t.Parallel()
	// fakeDockyard 404s any name absent from its specs map, mimicking a
	// dockyard skills/ dir with no fetchable spec.yaml. This must be skipped
	// (missing), not turned into a fatal error that aborts the whole sync.
	srv := fakeDockyard(t, map[string]string{})

	_, skillsDir := seedRegistry(t)
	res := createSkillEntry(context.Background(), optsFor(srv, false), skillsDir, "no-such-skill")
	if res.Status != statusMissing {
		t.Fatalf("expected missing, got %s (%s)", res.Status, res.Detail)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "no-such-skill")); !os.IsNotExist(err) {
		t.Errorf("missing skill left a directory behind")
	}
}

func TestCreateThenSyncIsNoop(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		newSkillName: dockyardSpecYAML(
			fixtureRepoURL, bumpedRef, "skills/new-skill", "0.2.0",
		),
	})

	root, skillsDir := seedRegistry(t)
	opts := optsFor(srv, false)
	if _, err := createMissingSkills(context.Background(), opts, root); err != nil {
		t.Fatalf("createMissingSkills: %v", err)
	}

	path := filepath.Join(skillsDir, newSkillName, "skill.json")
	res := syncOneSkill(context.Background(), opts, path)
	if res.Status != statusNoop {
		t.Fatalf("expected noop after create, got %s (%s)", res.Status, res.Detail)
	}
}

func TestListDockyardSkillNames_AuthHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"foo","type":"dir"}]`))
	}))
	t.Cleanup(srv.Close)

	opts := syncOptions{
		APIBaseURL:  srv.URL,
		DockyardRef: "main",
		HTTP:        srv.Client(),
	}
	names, err := listDockyardSkillNames(context.Background(), opts)
	if err != nil {
		t.Fatalf("listDockyardSkillNames: %v", err)
	}
	if len(names) != 1 || names[0] != "foo" {
		t.Errorf("unexpected names %v", names)
	}
	if gotAuth != "" {
		t.Errorf("expected no Authorization header without token, got %q", gotAuth)
	}

	opts.GitHubToken = "test-token"
	if _, err := listDockyardSkillNames(context.Background(), opts); err != nil {
		t.Fatalf("listDockyardSkillNames with token: %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("expected bearer token header, got %q", gotAuth)
	}
}

func TestListDockyardSkillNames_FiltersFiles(t *testing.T) {
	t.Parallel()
	srv := fakeDockyard(t, map[string]string{
		"b-skill": "",
		"a-skill": "",
	})
	names, err := listDockyardSkillNames(context.Background(), optsFor(srv, false))
	if err != nil {
		t.Fatalf("listDockyardSkillNames: %v", err)
	}
	// README.md ("type":"file") must be filtered; names come back sorted.
	if len(names) != 2 || names[0] != "a-skill" || names[1] != "b-skill" {
		t.Errorf("unexpected names %v", names)
	}
}

func TestDeriveTitle(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"tdd", "TDD"},
		{newSkillName, "New Skill"},
		{"vercel-cli-with-tokens", "Vercel CLI With Tokens"},
		{"dd-llmo-eval-trace-rca", "Dd Llmo Eval Trace RCA"},
		{"obsidian-markdown", "Obsidian Markdown"},
		{"json-canvas", "JSON Canvas"},
	}
	for _, tc := range cases {
		if got := deriveTitle(tc.in); got != tc.want {
			t.Errorf("deriveTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAddMissingRequiresAll(t *testing.T) {
	t.Parallel()
	syncSkillsAddMissing = true
	syncSkillsAll = false
	t.Cleanup(func() {
		syncSkillsAddMissing = false
		syncSkillsAll = false
	})
	err := validateSyncSkillsArgs(nil, []string{"some/skill.json"})
	if err == nil || !strings.Contains(err.Error(), "--add-missing requires --all") {
		t.Fatalf("expected --add-missing requires --all error, got %v", err)
	}
}
