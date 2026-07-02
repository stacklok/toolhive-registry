package skilljson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NewSkill describes a skill.json to be created from scratch. The struct
// field order matches the conventional key order of existing entries under
// registries/<reg>/skills/ so generated files look hand-written.
type NewSkill struct {
	Namespace   string       `json:"namespace"`
	Name        string       `json:"name"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Status      string       `json:"status"`
	Repository  Repository   `json:"repository"`
	Icons       []Icon       `json:"icons"`
	Packages    []NewPackage `json:"packages"`
}

// Repository is the upstream source repository of a skill.
type Repository struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// Icon is a single icon reference in a skill.json.
type Icon struct {
	Src  string `json:"src"`
	Type string `json:"type"`
}

// NewPackage is a package entry of a generated skill.json. OCI packages set
// Identifier; git packages set URL, Ref, and Subfolder.
type NewPackage struct {
	RegistryType string `json:"registryType"`
	Identifier   string `json:"identifier,omitempty"`
	URL          string `json:"url,omitempty"`
	Ref          string `json:"ref,omitempty"`
	Subfolder    string `json:"subfolder,omitempty"`
}

// MarshalNewSkill renders a NewSkill in the on-disk skill.json convention:
// two-space indentation, no HTML escaping (descriptions may contain raw
// < or &), and a trailing newline.
func MarshalNewSkill(s NewSkill) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, fmt.Errorf("failed to marshal skill %q: %w", s.Name, err)
	}
	return buf.Bytes(), nil
}
