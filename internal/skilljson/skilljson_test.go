package skilljson

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSkill = `{
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

func writeFixture(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "skill.json")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if sf.Skill.Name != "security-review" {
		t.Errorf("expected name security-review, got %q", sf.Skill.Name)
	}
	if len(sf.Skill.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(sf.Skill.Packages))
	}
}

func TestLoadMalformed(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, "{not json}")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error loading malformed skill.json")
	}
}

func TestDockyardOCIPackage(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	idx, name, tag := sf.DockyardOCIPackage()
	if idx != 0 {
		t.Errorf("expected idx 0, got %d", idx)
	}
	if name != "security-review" {
		t.Errorf("expected skill name security-review, got %q", name)
	}
	if tag != "0.1.0" {
		t.Errorf("expected tag 0.1.0, got %q", tag)
	}
}

func TestDockyardOCIPackageNotFound(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, strings.Replace(sampleSkill,
		"ghcr.io/stacklok/dockyard/skills/security-review:0.1.0",
		"ghcr.io/example/other:1.0.0", 1))
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if idx, _, _ := sf.DockyardOCIPackage(); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

func TestGitPackage(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got := sf.GitPackage(); got != 1 {
		t.Errorf("expected git package idx 1, got %d", got)
	}
}

func TestSetIdentifierBumpsTag(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	newID := "ghcr.io/stacklok/dockyard/skills/security-review:0.2.0"
	if err := sf.SetIdentifier(0, newID); err != nil {
		t.Fatalf("SetIdentifier failed: %v", err)
	}
	if sf.Skill.Packages[0].Identifier != newID {
		t.Errorf("identifier not updated in struct, got %q", sf.Skill.Packages[0].Identifier)
	}
	if !strings.Contains(string(sf.Bytes()), newID) {
		t.Errorf("new identifier not present in raw bytes")
	}
	if strings.Contains(string(sf.Bytes()), ":0.1.0") {
		t.Errorf("old tag still present in raw bytes")
	}
}

func TestSetIdentifierNoOp(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	before := string(sf.Bytes())
	if err := sf.SetIdentifier(0, sf.Skill.Packages[0].Identifier); err != nil {
		t.Fatalf("SetIdentifier failed on no-op: %v", err)
	}
	if string(sf.Bytes()) != before {
		t.Errorf("no-op SetIdentifier mutated bytes")
	}
}

func TestSetIdentifierOutOfRange(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if err := sf.SetIdentifier(99, "ghcr.io/foo:1"); err == nil {
		t.Error("expected out-of-range error")
	}
}

func TestSetRefBumpsRef(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	newRef := "f2cff985bcec174fcd096db65a23f06ec9bdde29"
	if err := sf.SetRef(1, newRef); err != nil {
		t.Fatalf("SetRef failed: %v", err)
	}
	if sf.Skill.Packages[1].Ref != newRef {
		t.Errorf("ref not updated in struct, got %q", sf.Skill.Packages[1].Ref)
	}
	if !strings.Contains(string(sf.Bytes()), newRef) {
		t.Errorf("new ref not present in raw bytes")
	}
	if strings.Contains(string(sf.Bytes()), "94ea2a26c70f3f646f07a613ffe5cd3d4eca1955") {
		t.Errorf("old ref still present in raw bytes")
	}
}

func TestSetRefUnspaced(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, strings.ReplaceAll(sampleSkill,
		`"ref": "94ea2a26c70f3f646f07a613ffe5cd3d4eca1955"`,
		`"ref":"94ea2a26c70f3f646f07a613ffe5cd3d4eca1955"`))
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if err := sf.SetRef(1, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"); err != nil {
		t.Fatalf("SetRef failed on unspaced form: %v", err)
	}
	if !strings.Contains(string(sf.Bytes()),
		`"ref":"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"`) {
		t.Errorf("unspaced ref not rewritten correctly")
	}
}

func TestSetRefNoOp(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	before := string(sf.Bytes())
	if err := sf.SetRef(1, sf.Skill.Packages[1].Ref); err != nil {
		t.Fatalf("SetRef failed on no-op: %v", err)
	}
	if string(sf.Bytes()) != before {
		t.Errorf("no-op SetRef mutated bytes")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	t.Parallel()
	path := writeFixture(t, sampleSkill)
	sf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	newID := "ghcr.io/stacklok/dockyard/skills/security-review:0.2.0"
	if err := sf.SetIdentifier(0, newID); err != nil {
		t.Fatalf("SetIdentifier failed: %v", err)
	}
	if err := sf.Write(); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(got), newID) {
		t.Errorf("written file does not contain new identifier")
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Errorf("trailing newline not preserved")
	}
}
