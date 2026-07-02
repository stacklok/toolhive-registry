package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/stacklok/toolhive-catalog/internal/skilljson"
)

const (
	defaultDockyardBaseURL = "https://raw.githubusercontent.com/stacklok/dockyard"

	statusUpdated = "updated"
	statusCreated = "created"
	statusNoop    = "noop"
	statusSkipped = "skipped"
	statusMissing = "missing"
	statusError   = "error"

	// opCreate marks results produced by the --add-missing create path so
	// the summary can report creations separately from updates.
	opCreate = "create"
)

var (
	syncSkillsAll        bool
	syncSkillsAddMissing bool
	syncSkillsDryRun     bool
	syncSkillsDockyard   string
	syncSkillsBaseURL    string
	syncSkillsAPIURL     string
)

var syncSkillsCmd = &cobra.Command{
	Use:   "sync-skills [skill.json]",
	Short: "Sync dockyard-published skills with their upstream spec.yaml",
	Long: `Reads each skill.json whose first OCI package targets
ghcr.io/stacklok/dockyard/skills/<name>:* and aligns the OCI tag and
git ref with the corresponding skills/<name>/spec.yaml in
github.com/stacklok/dockyard.

Modes:
  catalog sync-skills <skill.json>            Sync a single file
  catalog sync-skills --all                   Sync every dockyard-sourced skill
  catalog sync-skills --all --add-missing     Also create entries for dockyard
                                              skills absent from the catalog`,
	Args: validateSyncSkillsArgs,
	RunE: runSyncSkills,
}

func init() {
	syncSkillsCmd.Flags().BoolVar(
		&syncSkillsAll, "all", false,
		"Sync all dockyard-sourced skills under the registries directory",
	)
	syncSkillsCmd.Flags().BoolVar(
		&syncSkillsAddMissing, "add-missing", false,
		"Create catalog entries for dockyard skills that have none (requires --all)",
	)
	syncSkillsCmd.Flags().BoolVarP(
		&syncSkillsDryRun, "dry-run", "d", false,
		"Show changes without writing",
	)
	syncSkillsCmd.Flags().StringVar(
		&syncSkillsDockyard, "dockyard-ref", "main",
		"Branch or commit ref of stacklok/dockyard to read spec.yaml from",
	)
	syncSkillsCmd.Flags().StringVar(
		&syncSkillsBaseURL, "dockyard-base-url", defaultDockyardBaseURL,
		"Base URL for dockyard raw content (mainly for tests)",
	)
	syncSkillsCmd.Flags().StringVar(
		&syncSkillsAPIURL, "dockyard-api-url", defaultDockyardAPIBaseURL,
		"Base URL for the dockyard GitHub API (mainly for tests)",
	)
}

func validateSyncSkillsArgs(_ *cobra.Command, args []string) error {
	if syncSkillsAll && len(args) > 0 {
		return fmt.Errorf("cannot use --all with a positional argument")
	}
	if !syncSkillsAll && len(args) == 0 {
		return fmt.Errorf("requires a skill.json path or --all")
	}
	if syncSkillsAddMissing && !syncSkillsAll {
		return fmt.Errorf("--add-missing requires --all")
	}
	if len(args) > 1 {
		return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
	}
	return nil
}

// syncOptions bundles all state needed to perform a sync. Decoupling these
// from package-level cobra flags makes the underlying logic unit-testable
// in parallel.
type syncOptions struct {
	BaseURL     string
	APIBaseURL  string
	DockyardRef string
	DryRun      bool
	AddMissing  bool
	GitHubToken string
	HTTP        *http.Client
}

// dockyardSpec mirrors the relevant subset of dockyard's skills/<name>/spec.yaml.
type dockyardSpec struct {
	Metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"metadata"`
	Spec struct {
		Repository string `yaml:"repository"`
		Ref        string `yaml:"ref"`
		Path       string `yaml:"path"`
		Version    string `yaml:"version"`
	} `yaml:"spec"`
}

// syncResult captures what happened to a single skill.
type syncResult struct {
	Path      string
	SkillName string
	Status    string
	// Op distinguishes create-mode results (opCreate) from the default
	// update path, so summaries can count the two separately.
	Op        string
	Detail    string
	IDBefore  string
	IDAfter   string
	RefBefore string
	RefAfter  string
	// Planned is set in dry-run mode when the sync would have written
	// changes, so the summary can distinguish would-update from true noop.
	Planned bool
}

func defaultSyncOptions() syncOptions {
	return syncOptions{
		BaseURL:     syncSkillsBaseURL,
		APIBaseURL:  syncSkillsAPIURL,
		DockyardRef: syncSkillsDockyard,
		DryRun:      syncSkillsDryRun,
		AddMissing:  syncSkillsAddMissing,
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		HTTP:        &http.Client{Timeout: 15 * time.Second},
	}
}

func runSyncSkills(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	opts := defaultSyncOptions()
	if syncSkillsAll {
		return runSyncSkillsAll(ctx, opts, registriesDir)
	}
	res := syncOneSkill(ctx, opts, args[0])
	printSyncResult(res, true)
	if res.Status == statusError {
		return fmt.Errorf("sync failed for %s: %s", res.Path, res.Detail)
	}
	return nil
}

func runSyncSkillsAll(ctx context.Context, opts syncOptions, root string) error {
	paths, err := findDockyardSkillFiles(root)
	if err != nil {
		return err
	}
	if len(paths) == 0 && !opts.AddMissing {
		fmt.Println("No dockyard-sourced skills found")
		return nil
	}

	fmt.Printf("Syncing %d dockyard-sourced skill(s)\n\n", len(paths))

	results := make([]syncResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, syncOneSkill(ctx, opts, p))
	}

	if opts.AddMissing {
		created, err := createMissingSkills(ctx, opts, root)
		if err != nil {
			return err
		}
		results = append(results, created...)
	}

	printSyncSummary(results)

	if errs := countByStatus(results, statusError); errs > 0 {
		return fmt.Errorf("%d/%d sync(s) failed", errs, len(results))
	}
	return nil
}

// findDockyardSkillFiles walks the registries directory and returns paths to
// every skill.json whose first OCI package targets the dockyard prefix.
func findDockyardSkillFiles(root string) ([]string, error) {
	regs, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read registries directory %s: %w", root, err)
	}
	var out []string
	for _, reg := range regs {
		if !reg.IsDir() || strings.HasPrefix(reg.Name(), ".") {
			continue
		}
		paths, err := dockyardSkillsInRegistry(filepath.Join(root, reg.Name(), skillsSubdir))
		if err != nil {
			return nil, err
		}
		out = append(out, paths...)
	}
	sort.Strings(out)
	return out, nil
}

func dockyardSkillsInRegistry(skillsDir string) ([]string, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", skillsDir, err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(skillsDir, e.Name(), "skill.json")
		sf, err := skilljson.Load(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if idx, _, _ := sf.DockyardOCIPackage(); idx >= 0 {
			out = append(out, path)
		}
	}
	return out, nil
}

// syncOneSkill performs the full sync flow for a single skill.json path.
// It never panics; failures are reported via the returned syncResult.
func syncOneSkill(ctx context.Context, opts syncOptions, path string) syncResult {
	res := syncResult{Path: path}

	sf, err := skilljson.Load(path)
	if err != nil {
		return errorResult(res, err.Error())
	}

	ociIdx, name, currentTag := sf.DockyardOCIPackage()
	if ociIdx < 0 {
		res.Status = statusSkipped
		res.Detail = "no dockyard OCI package"
		return res
	}
	res.SkillName = name
	res.IDBefore = sf.Skill.Packages[ociIdx].Identifier

	spec, errRes, ok := loadDockyardSpec(ctx, opts, res, name)
	if !ok {
		return errRes
	}

	gitIdx := sf.GitPackage()
	if skipRes, ok := guardGitPackage(sf, gitIdx, spec, &res); !ok {
		return skipRes
	}

	return applySync(sf, ociIdx, gitIdx, currentTag, name, spec, opts, res)
}

// loadDockyardSpec fetches and validates a dockyard spec.yaml for the named
// skill. The (result, ok) sentinel mirrors a "happy path returns true" pattern
// so the caller can short-circuit on missing/invalid specs without nesting.
func loadDockyardSpec(
	ctx context.Context, opts syncOptions, res syncResult, name string,
) (*dockyardSpec, syncResult, bool) {
	spec, err := fetchDockyardSpec(ctx, opts, name)
	if err != nil {
		if isNotFoundErr(err) {
			res.Status = statusMissing
			res.Detail = "not present in dockyard"
			return nil, res, false
		}
		return nil, errorResult(res, err.Error()), false
	}
	if spec.Spec.Version == "" {
		return nil, errorResult(res, "dockyard spec.yaml is missing spec.version"), false
	}
	return spec, res, true
}

// guardGitPackage validates the git package against the dockyard spec.
// On mismatch it returns (skipResult, false). On match (or absence of a git
// package) it records the current ref into res via pointer and returns ok=true.
func guardGitPackage(
	sf *skilljson.SkillFile, gitIdx int, spec *dockyardSpec, res *syncResult,
) (syncResult, bool) {
	if gitIdx < 0 {
		return syncResult{}, true
	}
	gp := sf.Skill.Packages[gitIdx]
	if skip, detail := checkGitMatches(gp.URL, gp.Subfolder, spec); skip {
		out := *res
		out.Status = statusSkipped
		out.Detail = detail
		return out, false
	}
	res.RefBefore = gp.Ref
	return syncResult{}, true
}

// applySync performs the actual rewrite and write phase based on whether the
// tag/ref need to bump. dryRun short-circuits before touching disk.
func applySync(
	sf *skilljson.SkillFile,
	ociIdx, gitIdx int,
	currentTag, name string,
	spec *dockyardSpec,
	opts syncOptions,
	res syncResult,
) syncResult {
	newID := skilljson.DockyardSkillPrefix + name + ":" + spec.Spec.Version
	tagBumped := currentTag != spec.Spec.Version
	refBumped := gitIdx >= 0 && spec.Spec.Ref != "" && sf.Skill.Packages[gitIdx].Ref != spec.Spec.Ref

	if !tagBumped && !refBumped {
		res.Status = statusNoop
		res.IDAfter = res.IDBefore
		res.RefAfter = res.RefBefore
		return res
	}

	if tagBumped {
		if err := sf.SetIdentifier(ociIdx, newID); err != nil {
			return errorResult(res, err.Error())
		}
	}
	if refBumped {
		if err := sf.SetRef(gitIdx, spec.Spec.Ref); err != nil {
			return errorResult(res, err.Error())
		}
	}

	res.IDAfter = sf.Skill.Packages[ociIdx].Identifier
	if gitIdx >= 0 {
		res.RefAfter = sf.Skill.Packages[gitIdx].Ref
	}

	if opts.DryRun {
		res.Status = statusNoop
		res.Detail = "[DRY RUN] would update"
		res.Planned = true
		return res
	}

	if err := sf.Write(); err != nil {
		return errorResult(res, err.Error())
	}
	res.Status = statusUpdated
	return res
}

func errorResult(res syncResult, detail string) syncResult {
	res.Status = statusError
	res.Detail = detail
	return res
}

// checkGitMatches enforces that the catalog git package and the dockyard
// spec point at the same upstream repository and subfolder. A mismatch
// causes the skill to be skipped (not silently rewritten).
func checkGitMatches(catalogURL, catalogSubfolder string, spec *dockyardSpec) (skip bool, detail string) {
	if !repoURLsMatch(catalogURL, spec.Spec.Repository) {
		return true, fmt.Sprintf(
			"repository mismatch: catalog has %q, dockyard advertises %q",
			catalogURL, spec.Spec.Repository)
	}
	if spec.Spec.Path != "" && catalogSubfolder != "" && catalogSubfolder != spec.Spec.Path {
		return true, fmt.Sprintf(
			"subfolder mismatch: catalog has %q, dockyard advertises %q",
			catalogSubfolder, spec.Spec.Path)
	}
	return false, ""
}

// fetchDockyardSpec retrieves and parses the spec.yaml for the given skill name
// from stacklok/dockyard at the configured ref.
func fetchDockyardSpec(ctx context.Context, opts syncOptions, name string) (*dockyardSpec, error) {
	specURL := fmt.Sprintf(
		"%s/%s/skills/%s/spec.yaml",
		strings.TrimRight(opts.BaseURL, "/"),
		url.PathEscape(opts.DockyardRef),
		url.PathEscape(name),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	resp, err := opts.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", specURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound{url: specURL}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, specURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var spec dockyardSpec
	if err := yaml.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse spec.yaml for %s: %w", name, err)
	}
	return &spec, nil
}

type errNotFound struct {
	url string
}

func (e errNotFound) Error() string { return "not found: " + e.url }

func isNotFoundErr(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

// repoURLsMatch compares two git repository URLs while tolerating common
// suffix differences (".git", trailing slash). Empty inputs always return
// false so the sanity guard fails closed.
func repoURLsMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return canonRepoURL(a) == canonRepoURL(b)
}

func canonRepoURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSuffix(s, ".git")
	return s
}

func countByStatus(rs []syncResult, status string) int {
	n := 0
	for _, r := range rs {
		if r.Status == status {
			n++
		}
	}
	return n
}

func printSyncResult(r syncResult, single bool) {
	switch r.Status {
	case statusUpdated:
		fmt.Printf("UPDATED  %s (%s)\n", r.Path, r.SkillName)
		printDelta(r)
	case statusCreated:
		fmt.Printf("CREATED  %s (%s): %s\n", r.Path, r.SkillName, r.Detail)
	case statusNoop:
		if r.Planned {
			fmt.Printf("DRY-RUN  %s (%s)\n", r.Path, r.SkillName)
			if r.Detail != "" {
				fmt.Printf("  %s\n", r.Detail)
			}
			printDelta(r)
			return
		}
		if single || verbose {
			fmt.Printf("NOOP     %s (%s)\n", r.Path, r.SkillName)
		}
	case statusSkipped:
		fmt.Printf("SKIPPED  %s (%s): %s\n", r.Path, r.SkillName, r.Detail)
	case statusMissing:
		fmt.Printf("MISSING  %s (%s): %s\n", r.Path, r.SkillName, r.Detail)
	case statusError:
		fmt.Fprintf(os.Stderr, "ERROR    %s (%s): %s\n", r.Path, r.SkillName, r.Detail)
	}
}

func printDelta(r syncResult) {
	if r.IDBefore != r.IDAfter {
		fmt.Printf("  identifier: %s -> %s\n", r.IDBefore, r.IDAfter)
	}
	if r.RefBefore != r.RefAfter && r.RefAfter != "" {
		fmt.Printf("  ref:        %s -> %s\n", r.RefBefore, r.RefAfter)
	}
}

func printSyncSummary(results []syncResult) {
	for _, r := range results {
		printSyncResult(r, false)
	}
	plannedUpdate := 0
	plannedCreate := 0
	noop := 0
	for _, r := range results {
		if r.Status != statusNoop {
			continue
		}
		switch {
		case r.Planned && r.Op == opCreate:
			plannedCreate++
		case r.Planned:
			plannedUpdate++
		default:
			noop++
		}
	}
	fmt.Println()
	if plannedUpdate > 0 || plannedCreate > 0 {
		fmt.Printf("Summary: %d would-update, %d would-create, %d noop, %d skipped, %d missing, %d error (of %d) [dry run]\n",
			plannedUpdate,
			plannedCreate,
			noop,
			countByStatus(results, statusSkipped),
			countByStatus(results, statusMissing),
			countByStatus(results, statusError),
			len(results),
		)
		return
	}
	fmt.Printf("Summary: %d updated, %d created, %d noop, %d skipped, %d missing, %d error (of %d)\n",
		countByStatus(results, statusUpdated),
		countByStatus(results, statusCreated),
		noop,
		countByStatus(results, statusSkipped),
		countByStatus(results, statusMissing),
		countByStatus(results, statusError),
		len(results),
	)
}
