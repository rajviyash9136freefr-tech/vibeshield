package rules

import (
	"strings"
	"testing"
)

// minimal valid pack body around a list of rule YAML fragments.
func packYAML(rules ...string) string {
	return "schema: vibeshield.rules/v1\nid: core\nversion: 1.0.0\nlicense: MIT\nrules:\n" +
		strings.Join(rules, "\n")
}

const ruleGood = `  - id: VS-SEC-001
    category: hardcoded-secret
    severity: critical
    title: "AWS key"
    message: "A key that was real."
    fix: "Rotate it."
    languages: [generic]
    pattern:
      kind: regex
      match: 'AKIA[0-9A-Z]{16}'`

func TestParsePackBytes(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string // substring; "" = expect success
		rules   int
	}{
		{"valid single rule", packYAML(ruleGood), "", 1},
		{"two rules", packYAML(ruleGood, strings.Replace(ruleGood, "VS-SEC-001", "VS-PKG-002", 1)), "", 2},
		{"empty pack is legal", "schema: vibeshield.rules/v1\nid: core\nversion: 1.0.0\nlicense: MIT\nrules: []\n", "", 0},
		{"bad schema", strings.Replace(packYAML(ruleGood), SchemaID, "other/1", 1), "unsupported schema", 0},
		{"missing id", strings.Replace(packYAML(ruleGood), "id: core\n", "", 1), "pack id required", 0},
		{"unknown pack field", strings.Replace(packYAML(ruleGood), "license: MIT", "license: MIT\nextra: 1", 1), "unknown field", 0},
		{"unknown rule id format", strings.Replace(packYAML(ruleGood), "VS-SEC-001", "VS-BAD-001", 1), "id must match", 0},
		{"unknown severity", strings.Replace(packYAML(ruleGood), "severity: critical", "severity: fatal", 1), "unknown severity", 0},
		{"unknown category", strings.Replace(packYAML(ruleGood), "category: hardcoded-secret", "category: spooky", 1), "unknown category", 0},
		{"unknown language", strings.Replace(packYAML(ruleGood), "languages: [generic]", "languages: [brainfuck]", 1), "unknown language", 0},
		{"bad regex", strings.Replace(packYAML(ruleGood), "'AKIA[0-9A-Z]{16}'", "'AKIA('", 1), "does not compile", 0},
		{"lookahead rejected", strings.Replace(packYAML(ruleGood), "'AKIA[0-9A-Z]{16}'", "'AKIA(?=1)[0-9A-Z]{16}'", 1), "RE2 does not support", 0},
		{"backreference rejected", strings.Replace(packYAML(ruleGood), "'AKIA[0-9A-Z]{16}'", "'(A)\\\\1'", 1), "RE2 does not support", 0},
		{"possessive rejected", strings.Replace(packYAML(ruleGood), "'AKIA[0-9A-Z]{16}'", "'A++K'", 1), "RE2 does not support", 0},
		{"bad flag", strings.Replace(packYAML(ruleGood), "kind: regex", "kind: regex\n      flags: [ungreedy]", 1), "unsupported pattern flag", 0},
		{"unknown rule field", strings.Replace(packYAML(ruleGood), "fix: \"Rotate it.\"", "fix: \"Rotate it.\"\n    when: friday", 1), "unknown field", 0},
		{"missing title", strings.Replace(packYAML(ruleGood), "    title: \"AWS key\"\n", "", 1), "title required", 0},
		{"bad glob", strings.Replace(packYAML(ruleGood), "languages: [generic]", "languages: [generic]\n    paths:\n      include: [\"**/[bad\"]", 1), "bad glob", 0},
		{"literal kind valid", packYAML(strings.Replace(ruleGood, "kind: regex", "kind: literal", 1)), "", 1},
		{"structural reserved", packYAML(strings.Replace(ruleGood, "kind: regex\n      match: 'AKIA[0-9A-Z]{16}'", "kind: structural\n      match: namescore >= 0.7", 1)), "", 1},
		{"invalid YAML", "schema: [", "invalid YAML", 0},
		{"empty document", "", "empty pack document", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, errs := ParsePackBytes([]byte(tc.yaml), "test.yaml", LoadOptions{})
			if tc.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				if len(p.Rules) != tc.rules {
					t.Fatalf("rules = %d, want %d", len(p.Rules), tc.rules)
				}
				return
			}
			if len(errs) == 0 {
				t.Fatalf("expected error containing %q, got none", tc.wantErr)
			}
			if !strings.Contains(errs[0].Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", errs[0], tc.wantErr)
			}
		})
	}
}

func TestRuleMatching(t *testing.T) {
	p, errs := ParsePackBytes([]byte(packYAML(ruleGood)), "t.yaml", LoadOptions{})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	r := &p.Rules[0]
	if !r.MatchesPath("src/app.ts") {
		t.Error("generic include should match")
	}
	if !r.AppliesToLanguage("python") {
		t.Error("generic rules apply to every language")
	}
	loc := r.Pattern.Re().FindStringIndex("AWS_ACCESS_KEY_ID = AKIAFAKEFAKEFAKEFAKE")
	if loc == nil {
		t.Fatal("expected AKIA match")
	}
}

func TestLoadCore(t *testing.T) {
	pack, err := LoadCore()
	if err != nil {
		t.Fatalf("embedded core must load: %v", err)
	}
	if pack.ID != "core" || pack.License != "MIT" {
		t.Errorf("pack = %s/%s, want core/MIT", pack.ID, pack.License)
	}
	if len(pack.Rules) < 100 {
		t.Errorf("core pack should carry ~122 rules, got %d", len(pack.Rules))
	}
	ids := map[string]bool{}
	for _, r := range pack.Rules {
		if ids[r.ID] {
			t.Errorf("duplicate rule id %s", r.ID)
		}
		ids[r.ID] = true
	}
}
