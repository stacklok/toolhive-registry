package skilljson

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testNewSkill() NewSkill {
	return NewSkill{
		Namespace:   "io.github.stacklok",
		Name:        "tdd",
		Title:       "TDD",
		Description: "Test-driven development with a red-green-refactor loop — use when writing code & tests.",
		Version:     "0.1.1",
		Status:      "active",
		Repository: Repository{
			URL:  "https://github.com/mattpocock/skills",
			Type: RegistryTypeGit,
		},
		Icons: []Icon{
			{Src: "icon.svg", Type: "image/svg+xml"},
		},
		Packages: []NewPackage{
			{
				RegistryType: "oci",
				Identifier:   DockyardSkillPrefix + "tdd:0.1.1",
			},
			{
				RegistryType: RegistryTypeGit,
				URL:          "https://github.com/mattpocock/skills",
				Ref:          "f304057d61d3df3c9fd992ac2b6e3833cb9325fb",
				Subfolder:    "skills/engineering/tdd",
			},
		},
	}
}

func TestMarshalNewSkillGolden(t *testing.T) {
	t.Parallel()
	const want = `{
  "namespace": "io.github.stacklok",
  "name": "tdd",
  "title": "TDD",
  "description": "Test-driven development with a red-green-refactor loop — use when writing code & tests.",
  "version": "0.1.1",
  "status": "active",
  "repository": {
    "url": "https://github.com/mattpocock/skills",
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
      "identifier": "ghcr.io/stacklok/dockyard/skills/tdd:0.1.1"
    },
    {
      "registryType": "git",
      "url": "https://github.com/mattpocock/skills",
      "ref": "f304057d61d3df3c9fd992ac2b6e3833cb9325fb",
      "subfolder": "skills/engineering/tdd"
    }
  ]
}
`
	got, err := MarshalNewSkill(testNewSkill())
	if err != nil {
		t.Fatalf("MarshalNewSkill: %v", err)
	}
	if string(got) != want {
		t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestMarshalNewSkillNoHTMLEscaping(t *testing.T) {
	t.Parallel()
	s := testNewSkill()
	s.Description = "under 60 lines <target>, minimal & high-signal"
	got, err := MarshalNewSkill(s)
	if err != nil {
		t.Fatalf("MarshalNewSkill: %v", err)
	}
	if want := `"under 60 lines <target>, minimal & high-signal"`; !strings.Contains(string(got), want) {
		t.Errorf("expected raw %s in output, got:\n%s", want, got)
	}
}

func TestMarshalNewSkillRoundTrip(t *testing.T) {
	t.Parallel()
	data, err := MarshalNewSkill(testNewSkill())
	if err != nil {
		t.Fatalf("MarshalNewSkill: %v", err)
	}
	path := filepath.Join(t.TempDir(), "skill.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	idx, name, tag := sf.DockyardOCIPackage()
	if idx != 0 || name != "tdd" || tag != "0.1.1" {
		t.Errorf("DockyardOCIPackage = (%d, %q, %q), want (0, tdd, 0.1.1)", idx, name, tag)
	}
	if gitIdx := sf.GitPackage(); gitIdx != 1 {
		t.Errorf("GitPackage = %d, want 1", gitIdx)
	}
}
