package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	thvregistry "github.com/stacklok/toolhive-core/registry/types"
)

// SkillLoader walks a skills directory and loads skill.json files into memory.
type SkillLoader struct {
	skillsPath string
	entries    map[string]thvregistry.Skill
}

// NewSkillLoader creates a new SkillLoader that reads from the given skills directory.
func NewSkillLoader(skillsPath string) *SkillLoader {
	return &SkillLoader{
		skillsPath: skillsPath,
		entries:    make(map[string]thvregistry.Skill),
	}
}

// LoadAll walks top-level subdirectories under skillsPath, reads skill.json
// from each, and stores the results keyed by directory name.
// Hidden directories (starting with ".") are skipped.
// Returns an error if any skill.json contains malformed JSON.
func (l *SkillLoader) LoadAll() error {
	err := filepath.Walk(l.skillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories and the root directory itself
		if !info.IsDir() || path == l.skillsPath {
			return nil
		}

		relPath, err := filepath.Rel(l.skillsPath, path)
		if err != nil {
			return err
		}

		// Skip hidden directories
		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip nested directories (only process top-level)
		if strings.Contains(relPath, string(os.PathSeparator)) {
			return filepath.SkipDir
		}

		// Read skill.json from this directory
		skillJSONPath := filepath.Join(path, "skill.json")
		data, err := os.ReadFile(skillJSONPath) // #nosec G304 - path is constructed from known directory structure
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("WARNING: no skill.json in %s, skipping", info.Name())
				return nil
			}
			return fmt.Errorf("failed to read %s: %w", skillJSONPath, err)
		}

		var skill thvregistry.Skill
		if err := json.Unmarshal(data, &skill); err != nil {
			return fmt.Errorf("failed to parse %s: %w", skillJSONPath, err)
		}

		l.entries[info.Name()] = skill
		return nil
	})

	return err
}

// GetEntries returns all loaded skill.json entries keyed by directory name.
func (l *SkillLoader) GetEntries() map[string]thvregistry.Skill {
	return l.entries
}

// GetSortedNames returns directory names in sorted order for deterministic output.
func (l *SkillLoader) GetSortedNames() []string {
	names := make([]string, 0, len(l.entries))
	for name := range l.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
