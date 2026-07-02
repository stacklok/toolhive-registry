// Package skilljson provides utilities for reading and surgically modifying
// individual skill.json files. Updates are performed as byte-level
// substitutions so the original key order, indentation, and trailing newline
// of the file are preserved across rewrites.
package skilljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	thvregistry "github.com/stacklok/toolhive-core/registry/types"
)

// DockyardSkillPrefix is the OCI identifier prefix for skills published by
// the stacklok/dockyard pipeline.
const DockyardSkillPrefix = "ghcr.io/stacklok/dockyard/skills/"

// RegistryTypeGit is the package registryType value for git-sourced packages.
const RegistryTypeGit = "git"

// SkillFile represents a loaded skill.json with both its parsed structure
// and its original bytes for round-trip fidelity.
type SkillFile struct {
	// Path is the filesystem path to the skill.json file.
	Path string
	// Skill is the parsed skill structure.
	Skill thvregistry.Skill
	// rawBytes preserves the original file bytes for surgical edits.
	rawBytes []byte
}

// Load reads and parses a skill.json from the given path.
func Load(path string) (*SkillFile, error) {
	data, err := os.ReadFile(path) // #nosec G304 - path comes from caller (registry directory walk)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var skill thvregistry.Skill
	if err := json.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &SkillFile{
		Path:     path,
		Skill:    skill,
		rawBytes: data,
	}, nil
}

// DockyardOCIPackage returns the index of the first OCI package whose
// identifier starts with [DockyardSkillPrefix], along with the parsed
// skill name and current tag from the identifier. Returns idx == -1
// when no such package is present.
func (sf *SkillFile) DockyardOCIPackage() (idx int, skillName, currentTag string) {
	for i, p := range sf.Skill.Packages {
		if p.RegistryType != "oci" {
			continue
		}
		if !strings.HasPrefix(p.Identifier, DockyardSkillPrefix) {
			continue
		}
		rest := strings.TrimPrefix(p.Identifier, DockyardSkillPrefix)
		colon := strings.IndexByte(rest, ':')
		if colon <= 0 || colon == len(rest)-1 {
			continue
		}
		return i, rest[:colon], rest[colon+1:]
	}
	return -1, "", ""
}

// GitPackage returns the index of the first git package, or -1 if none.
func (sf *SkillFile) GitPackage() int {
	for i, p := range sf.Skill.Packages {
		if p.RegistryType == RegistryTypeGit {
			return i
		}
	}
	return -1
}

// SetIdentifier replaces the OCI identifier of the package at idx with a
// byte-level substitution. The old identifier must occur exactly once in
// the raw file bytes.
func (sf *SkillFile) SetIdentifier(idx int, newIdentifier string) error {
	if idx < 0 || idx >= len(sf.Skill.Packages) {
		return fmt.Errorf("package index %d out of range (have %d packages)", idx, len(sf.Skill.Packages))
	}
	old := sf.Skill.Packages[idx].Identifier
	if old == newIdentifier {
		return nil
	}
	oldEnc, err := json.Marshal(old)
	if err != nil {
		return fmt.Errorf("failed to encode old identifier: %w", err)
	}
	newEnc, err := json.Marshal(newIdentifier)
	if err != nil {
		return fmt.Errorf("failed to encode new identifier: %w", err)
	}
	if count := bytes.Count(sf.rawBytes, oldEnc); count != 1 {
		return fmt.Errorf("expected exactly 1 occurrence of identifier %q in %s, found %d", old, sf.Path, count)
	}
	sf.rawBytes = bytes.Replace(sf.rawBytes, oldEnc, newEnc, 1)
	sf.Skill.Packages[idx].Identifier = newIdentifier
	return nil
}

// SetRef replaces the ref of the package at idx with a byte-level
// substitution. The match is scoped to a `"ref": "<old>"` token to avoid
// touching unrelated string occurrences. Both spaced and unspaced forms
// of the JSON key are accepted.
func (sf *SkillFile) SetRef(idx int, newRef string) error {
	if idx < 0 || idx >= len(sf.Skill.Packages) {
		return fmt.Errorf("package index %d out of range (have %d packages)", idx, len(sf.Skill.Packages))
	}
	old := sf.Skill.Packages[idx].Ref
	if old == newRef {
		return nil
	}
	oldVal, err := json.Marshal(old)
	if err != nil {
		return fmt.Errorf("failed to encode old ref: %w", err)
	}
	newVal, err := json.Marshal(newRef)
	if err != nil {
		return fmt.Errorf("failed to encode new ref: %w", err)
	}

	candidates := [][2][]byte{
		{[]byte(`"ref": ` + string(oldVal)), []byte(`"ref": ` + string(newVal))},
		{[]byte(`"ref":` + string(oldVal)), []byte(`"ref":` + string(newVal))},
	}
	for _, c := range candidates {
		if bytes.Count(sf.rawBytes, c[0]) == 1 {
			sf.rawBytes = bytes.Replace(sf.rawBytes, c[0], c[1], 1)
			sf.Skill.Packages[idx].Ref = newRef
			return nil
		}
	}
	return fmt.Errorf("expected exactly 1 occurrence of `\"ref\": %s` in %s", oldVal, sf.Path)
}

// Bytes returns the current (possibly modified) raw bytes of the skill.json.
func (sf *SkillFile) Bytes() []byte {
	return sf.rawBytes
}

// Write writes the current raw bytes back to the file path.
func (sf *SkillFile) Write() error {
	return os.WriteFile(sf.Path, sf.rawBytes, 0600)
}
