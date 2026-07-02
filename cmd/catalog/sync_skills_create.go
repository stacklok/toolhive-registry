package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/stacklok/toolhive-catalog/internal/skilljson"
)

const (
	// defaultDockyardAPIBaseURL is the GitHub REST API base for the dockyard
	// repository, used to enumerate published skills.
	defaultDockyardAPIBaseURL = "https://api.github.com/repos/stacklok/dockyard"

	// defaultSkillNamespace is the namespace used by every catalog skill entry.
	defaultSkillNamespace = "io.github.stacklok"

	// syncIgnoreFile lists dockyard skill names (one per line, # comments)
	// that must never be auto-onboarded into a registry's skills directory.
	syncIgnoreFile = ".sync-skills-ignore"
)

// placeholderIconSVG is written for new entries when no sibling skill shares
// the upstream repository. It matches the 24x24 stroke style of existing
// catalog icons and is intended to be replaced during PR review.
const placeholderIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` +
	`stroke="#6b7280" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` +
	`<rect x="3" y="3" width="18" height="18" rx="4"/><path d="M12 8v8M8 12h8"/></svg>` + "\n"

// contentsEntry is the subset of the GitHub contents API response needed to
// enumerate dockyard's skills directory.
type contentsEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// listDockyardSkillNames enumerates the skill directories in dockyard's
// skills/ folder at the configured ref via the GitHub contents API.
func listDockyardSkillNames(ctx context.Context, opts syncOptions) ([]string, error) {
	listURL := fmt.Sprintf(
		"%s/contents/skills?ref=%s",
		strings.TrimRight(opts.APIBaseURL, "/"),
		url.QueryEscape(opts.DockyardRef),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if opts.GitHubToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.GitHubToken)
	}
	resp, err := opts.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", listURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, listURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var entries []contentsEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse contents listing: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.Type == "dir" && e.Name != "" {
			names = append(names, e.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// createMissingSkills creates catalog entries for every dockyard skill that
// has no directory under a registry's skills/ subdir and is not listed in
// that registry's ignore file. Only registries that already carry at least
// one dockyard-sourced skill participate — a registry with an empty (or
// dockyard-free) skills directory has not opted into dockyard syncing.
// A listing failure is fatal: the caller should surface it rather than
// silently skip onboarding.
func createMissingSkills(ctx context.Context, opts syncOptions, root string) ([]syncResult, error) {
	dockyardNames, err := listDockyardSkillNames(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list dockyard skills: %w", err)
	}

	regs, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read registries directory %s: %w", root, err)
	}

	var results []syncResult
	for _, reg := range regs {
		if !reg.IsDir() || strings.HasPrefix(reg.Name(), ".") {
			continue
		}
		skillsDir := filepath.Join(root, reg.Name(), skillsSubdir)
		dockyardSourced, err := dockyardSkillsInRegistry(skillsDir)
		if err != nil {
			return nil, err
		}
		if len(dockyardSourced) == 0 {
			// Registry has no dockyard-sourced skills (or no skills/ subdir
			// at all); nothing to onboard into.
			continue
		}
		existing, err := existingSkillDirs(skillsDir)
		if err != nil {
			return nil, err
		}
		ignored, err := loadSyncIgnore(skillsDir)
		if err != nil {
			return nil, err
		}
		for _, name := range dockyardNames {
			if existing[name] || ignored[name] {
				continue
			}
			results = append(results, createSkillEntry(ctx, opts, skillsDir, name))
		}
	}
	return results, nil
}

// createSkillEntry fetches the dockyard spec for name and writes a new
// skills/<name>/ directory with skill.json and icon.svg. Failures are
// reported via the returned syncResult, never a panic.
func createSkillEntry(ctx context.Context, opts syncOptions, skillsDir, name string) syncResult {
	dir := filepath.Join(skillsDir, name)
	res := syncResult{
		Path:      filepath.Join(dir, "skill.json"),
		SkillName: name,
		Op:        opCreate,
	}

	spec, err := fetchDockyardSpec(ctx, opts, name)
	if err != nil {
		// A directory listed under dockyard's skills/ that has no fetchable
		// spec.yaml (mid-publish, a shared/template dir, a transient 404) is
		// skipped rather than failing the whole run — matching the
		// single-skill path in loadDockyardSpec. A create error would
		// otherwise abort the entire sync, including legitimate tag/ref
		// updates for other skills.
		if isNotFoundErr(err) {
			res.Status = statusMissing
			res.Detail = "not present in dockyard"
			return res
		}
		return errorResult(res, err.Error())
	}

	skill, err := buildNewSkill(name, spec)
	if err != nil {
		return errorResult(res, err.Error())
	}
	data, err := skilljson.MarshalNewSkill(skill)
	if err != nil {
		return errorResult(res, err.Error())
	}

	icon, iconDetail, err := findSiblingIcon(skillsDir, spec.Spec.Repository)
	if err != nil {
		return errorResult(res, err.Error())
	}
	// Dockyard's spec.yaml carries no license, so generated entries never
	// have one — flag it alongside the icon so reviewers fill it in before
	// merging (every hand-written catalog skill has a license).
	detail := iconDetail + "; license needs review"
	res.Detail = detail
	res.IDAfter = skill.Packages[0].Identifier
	res.RefAfter = spec.Spec.Ref

	if opts.DryRun {
		res.Status = statusNoop
		res.Detail = "[DRY RUN] would create (" + detail + ")"
		res.Planned = true
		return res
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return errorResult(res, fmt.Sprintf("failed to create %s: %s", dir, err))
	}
	if err := writeEntryFiles(dir, data, icon); err != nil {
		// Remove the partially written entry so a failed run does not leave
		// a directory that later runs would treat as an existing skill.
		_ = os.RemoveAll(dir)
		return errorResult(res, err.Error())
	}

	res.Status = statusCreated
	return res
}

func writeEntryFiles(dir string, skillJSON, icon []byte) error {
	skillPath := filepath.Join(dir, "skill.json")
	if err := os.WriteFile(skillPath, skillJSON, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", skillPath, err)
	}
	iconPath := filepath.Join(dir, "icon.svg")
	if err := os.WriteFile(iconPath, icon, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", iconPath, err)
	}
	return nil
}

// buildNewSkill maps a dockyard spec.yaml onto a catalog skill.json entry.
func buildNewSkill(name string, spec *dockyardSpec) (skilljson.NewSkill, error) {
	switch {
	case spec.Spec.Version == "":
		return skilljson.NewSkill{}, fmt.Errorf("dockyard spec.yaml for %s is missing spec.version", name)
	case spec.Spec.Repository == "":
		return skilljson.NewSkill{}, fmt.Errorf("dockyard spec.yaml for %s is missing spec.repository", name)
	case spec.Spec.Ref == "":
		return skilljson.NewSkill{}, fmt.Errorf("dockyard spec.yaml for %s is missing spec.ref", name)
	case spec.Metadata.Description == "":
		return skilljson.NewSkill{}, fmt.Errorf("dockyard spec.yaml for %s is missing metadata.description", name)
	}

	return skilljson.NewSkill{
		Namespace:   defaultSkillNamespace,
		Name:        name,
		Title:       deriveTitle(name),
		Description: spec.Metadata.Description,
		Version:     spec.Spec.Version,
		Status:      "active",
		Repository: skilljson.Repository{
			URL:  spec.Spec.Repository,
			Type: skilljson.RegistryTypeGit,
		},
		Icons: []skilljson.Icon{
			{Src: "icon.svg", Type: "image/svg+xml"},
		},
		Packages: []skilljson.NewPackage{
			{
				RegistryType: "oci",
				Identifier:   skilljson.DockyardSkillPrefix + name + ":" + spec.Spec.Version,
			},
			{
				RegistryType: skilljson.RegistryTypeGit,
				URL:          spec.Spec.Repository,
				Ref:          spec.Spec.Ref,
				Subfolder:    spec.Spec.Path,
			},
		},
	}, nil
}

// titleAcronyms are lowercase words that render fully uppercased in derived
// titles instead of title case.
var titleAcronyms = map[string]bool{
	"ai": true, "api": true, "aws": true, "cd": true, "ci": true,
	"cli": true, "css": true, "gcp": true, "html": true, "http": true,
	"json": true, "md": true, "rca": true, "sdk": true, "sql": true,
	"tdd": true, "ui": true, "url": true, "yaml": true,
}

// deriveTitle turns a kebab-case skill name into a human-readable title,
// e.g. "vercel-cli-with-tokens" -> "Vercel CLI With Tokens". Results are
// polished by reviewers in the onboarding PR.
func deriveTitle(name string) string {
	words := strings.Split(name, "-")
	for i, w := range words {
		switch {
		case w == "":
			continue
		case titleAcronyms[w]:
			words[i] = strings.ToUpper(w)
		default:
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// findSiblingIcon returns the icon.svg bytes of the first existing skill
// (in sorted directory order) whose upstream repository matches repoURL.
// When no sibling matches, the generic placeholder icon is returned along
// with a detail string asking for review.
func findSiblingIcon(skillsDir, repoURL string) ([]byte, string, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte(placeholderIconSVG), "placeholder icon — needs review", nil
		}
		return nil, "", fmt.Errorf("failed to read %s: %w", skillsDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		sf, err := skilljson.Load(filepath.Join(skillsDir, e.Name(), "skill.json"))
		if err != nil {
			continue
		}
		if sf.Skill.Repository == nil || !repoURLsMatch(sf.Skill.Repository.URL, repoURL) {
			continue
		}
		icon, err := os.ReadFile(filepath.Join(skillsDir, e.Name(), "icon.svg")) // #nosec G304 - registry directory walk
		if err != nil {
			continue
		}
		return icon, "icon copied from " + e.Name(), nil
	}
	return []byte(placeholderIconSVG), "placeholder icon — needs review", nil
}

// loadSyncIgnore reads the per-registry ignore file. A missing file is not
// an error and yields an empty set.
func loadSyncIgnore(skillsDir string) (map[string]bool, error) {
	path := filepath.Join(skillsDir, syncIgnoreFile)
	f, err := os.Open(path) // #nosec G304 - registry directory walk
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	defer f.Close()

	ignored := map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignored[line] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return ignored, nil
}

// existingSkillDirs returns the set of skill directory names under
// skillsDir, or nil (no error) when the registry has no skills/ subdir.
func existingSkillDirs(skillsDir string) (map[string]bool, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", skillsDir, err)
	}
	existing := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			existing[e.Name()] = true
		}
	}
	return existing, nil
}
