package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

func corePack(t *testing.T) *rules.Pack {
	t.Helper()
	p, err := rules.LoadCore()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// writeTree creates files under a temp dir and returns the dir.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestDirDetectsSeededFindings(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"src/app.ts": "const AWS_ACCESS_KEY_ID = \"AKIAFAKEFAKEFAKEFAKE\";\n",
		"src/ok.js":  "export const x = 1;\n",
	})
	rep, err := Dir(dir, corePack(t), "test", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Scan.FilesScanned != 2 {
		t.Errorf("files scanned = %d, want 2", rep.Scan.FilesScanned)
	}
	if rep.Summary.Critical == 0 {
		t.Fatalf("expected a critical AKIA finding, got %+v", rep.Summary)
	}
	var f *finding.Finding
	for _, x := range rep.Findings {
		if x.RuleID == "VS-SEC-001" {
			f = x
		}
	}
	if f == nil {
		t.Fatal("VS-SEC-001 not emitted")
	}
	if f.File != "src/app.ts" {
		t.Errorf("file = %q, want src/app.ts", f.File)
	}
	if f.Line != 1 || f.Column == 0 {
		t.Errorf("line/col = %d/%d", f.Line, f.Column)
	}
	if f.DismissHash == "" {
		t.Error("dismiss hash must be set")
	}
}

func TestSecretRedaction(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"a.py": "key = \"AKIAFAKEFAKEFAKEFAKE\"\n",
	})
	rep, err := Dir(dir, corePack(t), "test", Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rep.Findings {
		if f.Category == finding.CategoryHardcodedSecret {
			if strings.Contains(f.Snippet, "AKIAFAKEFAKEFAKEFAKE") {
				t.Errorf("snippet leaks full secret: %q", f.Snippet)
			}
			if !strings.Contains(f.Snippet, string(finding.RedactionChar)) {
				t.Errorf("expected redaction bullets in %q", f.Snippet)
			}
		}
	}
}

func TestSkipDirs(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"node_modules/evil/index.js": "const k = \"AKIAFAKEFAKEFAKEFAKE\";\n",
		".git/hooks/x":               "const k = \"AKIAFAKEFAKEFAKEFAKE\";\n",
	})
	rep, err := Dir(dir, corePack(t), "test", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Scan.FilesScanned != 0 {
		t.Errorf("vendored/git dirs must be skipped, scanned %d", rep.Scan.FilesScanned)
	}
}

func TestIgnores(t *testing.T) {
	dir := writeTree(t, map[string]string{"tests/fixture.js": "const k = \"AKIAFAKEFAKEFAKEFAKE\";\n"})
	pack := corePack(t)
	rep, err := Dir(dir, pack, "test", Options{Ignores: []Ignore{{Rule: "VS-SEC-001", Globs: []string{"tests/**"}}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rep.Findings {
		if f.File == "tests/fixture.js" {
			t.Errorf("ignored finding leaked through: %+v", f)
		}
	}
}

func TestLanguageFilter(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"a.ts": "const k = \"AKIAFAKEFAKEFAKEFAKE\";\n",
		"b.md": "key: AKIAFAKEFAKEFAKEFAKE\n",
	})
	rep, err := Dir(dir, corePack(t), "test", Options{Languages: []string{"typescript"}})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Scan.FilesScanned != 1 {
		t.Errorf("files = %d, want only a.ts", rep.Scan.FilesScanned)
	}
}

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		glob, path string
		want       bool
	}{
		{"tests/**", "tests/unit/a.js", true},
		{"tests/**", "src/tests/a.js", false},
		{"**/*.py", "a/b/c.py", true},
		{"*.js", "src/a.js", false},
		{"src/*.js", "src/a.js", true},
	}
	for _, c := range cases {
		got, err := MatchGlob(c.glob, c.path)
		if err != nil {
			t.Errorf("%s: %v", c.glob, err)
		}
		if got != c.want {
			t.Errorf("MatchGlob(%q,%q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

func TestSummaryBlocks(t *testing.T) {
	s := Summary{Critical: 1}
	if !s.Blocks("block-on-critical") || !s.Blocks("block-on-high+") || s.Blocks("warn") {
		t.Error("block matrix wrong")
	}
	s2 := Summary{High: 2}
	if s2.Blocks("block-on-critical") || !s2.Blocks("block-on-high+") {
		t.Error("high+ matrix wrong")
	}
}

func TestBinaryFilesSkipped(t *testing.T) {
	dir := writeTree(t, map[string]string{"x.js": "const k = \"AKIAFAKEFAKEFAKEFAKE\";\n"})
	if err := os.WriteFile(filepath.Join(dir, "blob.bin"), []byte{0, 1, 'A', 'K', 'I', 'A'}, 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := Dir(dir, corePack(t), "test", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Scan.FilesScanned != 1 {
		t.Errorf("binary must be skipped; scanned %d", rep.Scan.FilesScanned)
	}
}

func TestLangForFile(t *testing.T) {
	cases := map[string]string{"a.ts": "typescript", "b.py": "python", "c.mjs": "javascript", "d.unknown": "generic", ".env": "generic"}
	for p, want := range cases {
		if got := LangForFile(p); got != want {
			t.Errorf("LangForFile(%s) = %s, want %s", p, got, want)
		}
	}
}
