package fix

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
)

// testPack mirrors the core pack's autofix shape without depending on pack
// loading (rules_test covers parsing; this covers planning + applying).
func testPack(t *testing.T) *rules.Pack {
	t.Helper()
	pack := &rules.Pack{Schema: rules.SchemaID, ID: "test", Version: "0", License: "MIT"}
	data := []byte(`
schema: vibeshield.rules/v1
id: test
version: 0.0.0
license: MIT
rules:
  - id: VS-SEC-014
    category: insecure-default
    severity: high
    title: "Flask debug on"
    message: "msg"
    fix: "set debug=False"
    autofix:
      match: 'debug\s*=\s*True'
      replace: 'debug=False'
    languages: [python]
    pattern:
      kind: regex
      match: 'app\.run\s*\([^)]*debug\s*=\s*True'
  - id: VS-SEC-001
    category: hardcoded-secret
    severity: critical
    title: "Secret"
    message: "msg"
    fix: "rotate"
    autofix:
      match: 'KEY\s*=\s*"[^"]*"'
      replace: 'KEY=REDACTED'
    languages: [python]
    pattern:
      kind: regex
      match: 'KEY\s*=\s*"'
`)
	p, errs := rules.ParsePackBytes(data, "test.yaml", rules.LoadOptions{})
	if len(errs) > 0 {
		t.Fatalf("pack must load: %v", errs)
	}
	_ = pack
	return p
}

func planFor(t *testing.T, root string) *Plan {
	t.Helper()
	rep := &scan.Report{Findings: []*finding.Finding{
		{RuleID: "VS-SEC-014", Severity: "high", Category: "insecure-default",
			Title: "Flask debug on", Message: "m", File: "app.py", Line: 3,
			Fix: "set debug=False", AIOrigin: "unknown"},
		{RuleID: "VS-SEC-001", Severity: "critical", Category: "hardcoded-secret",
			Title: "Secret", Message: "m", File: "app.py", Line: 5,
			Fix: "rotate", AIOrigin: "unknown"},
	}}
	return BuildPlan(root, rep, testPack(t))
}

func TestBuildPlanAppliesAutofixAndSkipsSecrets(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.py"),
		[]byte("from flask import Flask\napp = Flask(__name__)\napp.run(debug=True)\n\nAPI_KEY = \"sk-proj-FAKE1234\"\n"), 0o644)

	plan := planFor(t, dir)
	if len(plan.Files) != 1 {
		t.Fatalf("want 1 file, got %d", len(plan.Files))
	}
	fp := plan.Files[0]
	if len(fp.Patches) != 1 {
		t.Fatalf("secret category must never be patched, got %d patches", len(fp.Patches))
	}
	if fp.Patches[0].LineNo != 3 || fp.Patches[0].After != "app.run(debug=False)" {
		t.Errorf("patch = %+v", fp.Patches[0])
	}
}

func TestApplyWritesFilePreservingTailNewline(t *testing.T) {
	dir := t.TempDir()
	src := "import flask\napp = flask.Flask(__name__)\napp.run(debug=True)\n\nKEY =\n"
	os.WriteFile(filepath.Join(dir, "app.py"), []byte(src), 0o644)
	plan := planFor(t, dir)
	// drop the secret finding line so only the autofix target remains
	var audit bytes.Buffer
	g := &Gate{Mode: "yes", Stdin: strings.NewReader(""), Out: &audit}
	applied, err := Apply(dir, plan, g, &audit, "yes")
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("applied = %d, want 1", applied)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "app.py"))
	if string(got) != "import flask\napp = flask.Flask(__name__)\napp.run(debug=False)\n\nKEY =\n" {
		t.Errorf("file = %q", got)
	}
	if !strings.Contains(audit.String(), `"rule_id":"VS-SEC-014"`) {
		t.Errorf("audit trail missing entry: %q", audit.String())
	}
	// no temp files left behind
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), "vibeshield-tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestDryRunChangesNothing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("app.run(debug=True)\n"), 0o644)
	plan := planFor(t, dir)
	var out bytes.Buffer
	g := &Gate{Mode: "dry", Out: &out}
	applied, err := Apply(dir, plan, g, nil, "interactive")
	if err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Fatalf("dry run applied %d patches", applied)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "app.py"))
	if string(got) != "app.run(debug=True)\n" {
		t.Errorf("dry run modified the file: %q", got)
	}
}

func TestBuildPlanStaleFindingUnfixable(t *testing.T) {
	dir := t.TempDir()
	// line 3 no longer matches (someone fixed it manually)
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("a\nb\napp.run(debug=False)\nc\nKEY =\n"), 0o644)
	plan := planFor(t, dir)
	if plan.HasWork() {
		t.Fatal("stale finding must not produce patches")
	}
	if plan.Unfixable == 0 {
		t.Error("stale findings should be counted")
	}
}
