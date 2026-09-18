package cli

import (
	"strings"
	"testing"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

// testPack mirrors the shape the scanner loads: a couple of rules with the
// fields the catalog reads.
func testPack() *rules.Pack {
	return &rules.Pack{
		Schema:  rules.SchemaID,
		ID:      "core",
		Version: "2.0.0",
		License: "MIT",
		Rules: []rules.Rule{
			{
				ID: "VS-SEC-017", Category: "hardcoded-secret", Severity: "high",
				Title: "API key hardcoded in source", Message: "Keys in source leak on push.",
				Fix: "Read the key from the environment.", Languages: []string{"typescript"},
			},
			{
				ID: "VS-PKG-001", Category: "hallucinated-package", Severity: "critical",
				Title:      "Hallucinated package name added as a dependency",
				Message:    "Attackers register invented names.",
				Fix:        "Use the standard library.",
				Autofix:    &rules.Autofix{Match: "old", Replace: "new"},
				Languages:  []string{"javascript", "python"},
				References: []string{"https://vibeshield.dev/docs/hallucinated-packages"},
			},
		},
	}
}

func TestCatalogIncludesActionsAgentsAndRules(t *testing.T) {
	items := Catalog(testPack(), "2.0.0")
	if len(items) == 0 {
		t.Fatal("Catalog returned no items")
	}
	var actions, agents, ruleEntries int
	for _, it := range items {
		if it.Title == "" || it.Group == "" {
			t.Errorf("item %q is missing a title or group", it.Title)
		}
		switch {
		case it.Kind == KindAction:
			actions++
		case it.Group == "Agent setup":
			agents++
		case strings.HasPrefix(it.Group, "Rule · "):
			ruleEntries++
		}
	}
	if actions == 0 {
		t.Error("no actions in the catalog")
	}
	if agents == 0 {
		t.Error("no agent setup entries in the catalog")
	}
	if ruleEntries != 2 {
		t.Errorf("rule entries = %d, want 2", ruleEntries)
	}
}

func TestCatalogSurvivesNilPack(t *testing.T) {
	items := Catalog(nil, "2.0.0")
	if len(items) == 0 {
		t.Fatal("Catalog(nil) must still return the action and agent entries")
	}
	for _, it := range items {
		if strings.HasPrefix(it.Group, "Rule · ") {
			t.Fatalf("nil pack produced rule entry %q", it.Title)
		}
	}
}

func TestEveryActionHasRunnableArgs(t *testing.T) {
	for _, it := range actionItems("2.0.0") {
		if it.Kind != KindAction {
			t.Errorf("%q: actions must have KindAction", it.Title)
		}
		if len(it.Args) == 0 {
			t.Errorf("%q: no args, so the console could not run it", it.Title)
		}
		if len(it.Keywords) == 0 {
			t.Errorf("%q: no search keywords", it.Title)
		}
	}
}

func TestRuleItemsAreSearchableByIDCategoryAndFix(t *testing.T) {
	items := ruleItems(testPack())
	byID := map[string]Item{}
	for _, it := range items {
		byID[strings.Fields(it.Title)[0]] = it
	}
	for _, id := range []string{"VS-SEC-017", "VS-PKG-001"} {
		it, ok := byID[id]
		if !ok {
			t.Fatalf("rule %s missing from the catalog", id)
		}
		if ScoreItem(id, it) == 0 {
			t.Errorf("rule %s does not match its own id", id)
		}
		if !strings.Contains(it.Body, "Why it matters") || !strings.Contains(it.Body, "Fix") {
			t.Errorf("rule %s body is missing the Why/Fix sections", id)
		}
	}
	secret := byID["VS-SEC-017"]
	if ScoreItem("hardcoded secret", secret) == 0 {
		t.Error("rule is not reachable by its category")
	}
	if ScoreItem("environment", secret) == 0 {
		t.Error("rule is not reachable by words from its fix")
	}
}

func TestRuleBodyRendersAutofixAndReferences(t *testing.T) {
	body := ruleBody(testPack().Rules[1])
	for _, want := range []string{"VS-PKG-001", "Autofix (VibePatch)", "match:", "References", "vibeshield.dev"} {
		if !strings.Contains(body, want) {
			t.Errorf("rule body missing %q\n%s", want, body)
		}
	}
}

func TestSeverityLabelsCoverContractTokens(t *testing.T) {
	want := map[string]string{
		"critical": "CRITICAL", "high": "HIGH", "medium": "MEDIUM",
		"low": "LOW", "info": "INFO",
	}
	for sev, token := range want {
		if got := severityLabel(sev); !strings.Contains(got, token) {
			t.Errorf("severityLabel(%q) = %q, want it to contain %q", sev, got, token)
		}
	}
}

func TestAgentsAreWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range Agents() {
		if a.Slug == "" || a.Name == "" || a.File == "" {
			t.Errorf("agent %q is missing slug/name/file", a.Name)
		}
		if seen[a.Slug] {
			t.Errorf("duplicate agent slug %q", a.Slug)
		}
		seen[a.Slug] = true
		if len(a.Steps) == 0 {
			t.Errorf("agent %q has no install steps", a.Slug)
		}
		if !strings.Contains(a.Snippet, "HALLUCINATED PACKAGES") {
			t.Errorf("agent %q does not carry the shared rule body", a.Slug)
		}
		if len(a.Aliases) == 0 {
			t.Errorf("agent %q has no search aliases", a.Slug)
		}
	}
}

func TestAgentsCoverTheNamedTargets(t *testing.T) {
	// v2 ships first-class setup for every agent the project claims support
	// for; a missing slug means the README would be lying.
	want := []string{"codex", "claude-code", "antigravity", "cursor", "windsurf", "copilot"}
	have := map[string]bool{}
	for _, a := range Agents() {
		have[a.Slug] = true
	}
	for _, slug := range want {
		if !have[slug] {
			t.Errorf("missing agent recipe %q", slug)
		}
	}
}

func TestAgentRulesBodyIsAPasteableRawString(t *testing.T) {
	// Regression guard: a backtick inside the raw string literal silently
	// truncates the constant and breaks the build.
	if strings.Contains(AgentRulesBody, "`") {
		t.Fatal("AgentRulesBody must not contain backticks (it is a raw string literal)")
	}
	if len(AgentRulesBody) > 2048 {
		t.Errorf("AgentRulesBody is %d bytes; agents truncate long context, keep it under 2 KB", len(AgentRulesBody))
	}
	for i := 1; i <= 5; i++ {
		if !strings.Contains(AgentRulesBody, "\n"+string(rune('0'+i))+".") {
			t.Errorf("AgentRulesBody is missing rule %d", i)
		}
	}
}

func TestAgentEntriesAreSearchableByAlias(t *testing.T) {
	items := agentItems()
	find := func(query string) bool {
		for _, it := range items {
			if ScoreItem(query, it) > 0 {
				return true
			}
		}
		return false
	}
	for _, q := range []string{"codex", "claude desktop", "antigravity", "cursorrules", "windsurf", "copilot"} {
		if !find(q) {
			t.Errorf("no agent entry matches %q", q)
		}
	}
}
