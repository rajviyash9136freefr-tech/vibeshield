package cli

import "testing"

func TestNormalizeFoldsPunctuation(t *testing.T) {
	cases := map[string]string{
		"Claude-Code":       "claude code",
		"VS-SEC-017":        "vs sec 017",
		"pre_commit":        "pre commit",
		"  Pre   Commit  ":  "pre commit",
		".cursor/rules":     "cursor rules",
		"block-on-high+":    "block on high+",
		"github_action.yml": "github action yml",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTokenize(t *testing.T) {
	got := tokenize("  AWS   SECRET-Key ")
	want := []string{"aws", "secret", "key"}
	if len(got) != len(want) {
		t.Fatalf("tokenize = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokenize = %v, want %v", got, want)
		}
	}
	if n := len(tokenize("   ")); n != 0 {
		t.Errorf("tokenize(blank) len = %d, want 0", n)
	}
}

func TestScoreTextRanksExactAboveFuzzy(t *testing.T) {
	exact := scoreText("scan", "scan")
	prefix := scoreText("scan", "scan staged changes")
	boundary := scoreText("staged", "scan staged changes")
	substring := scoreText("taged", "scan staged changes")
	fuzzy := scoreText("vsec017", "vs sec 017")

	if exact <= prefix {
		t.Errorf("exact (%d) should beat prefix (%d)", exact, prefix)
	}
	if prefix <= boundary {
		t.Errorf("prefix (%d) should beat word-boundary (%d)", prefix, boundary)
	}
	if boundary <= substring {
		t.Errorf("boundary (%d) should beat substring (%d)", boundary, substring)
	}
	if fuzzy <= 0 || fuzzy >= substring {
		t.Errorf("fuzzy (%d) should be positive and below substring (%d)", fuzzy, substring)
	}
	if n := scoreText("zzzz", "scan staged changes"); n != 0 {
		t.Errorf("non-match scored %d, want 0", n)
	}
	if n := scoreText("scan", ""); n != 0 {
		t.Errorf("empty haystack scored %d, want 0", n)
	}
}

func TestScoreTextFoldsSeparators(t *testing.T) {
	// A user typing the human spelling must reach the hyphenated rule id.
	if scoreText("vs sec 017", "VS-SEC-017") == 0 {
		t.Error("expected 'vs sec 017' to match 'VS-SEC-017'")
	}
	if scoreText("claude code", "Claude Code (CLI + Desktop)") == 0 {
		t.Error("expected 'claude code' to match the Claude agent title")
	}
}

func TestScoreItemEmptyQueryMatchesEverything(t *testing.T) {
	it := Item{Title: "Scan this project"}
	if n := ScoreItem("", it); n != 1 {
		t.Errorf("ScoreItem(\"\") = %d, want 1", n)
	}
	if n := ScoreItem("   ", it); n != 1 {
		t.Errorf("ScoreItem(blank) = %d, want 1", n)
	}
}

func TestScoreItemRequiresEveryToken(t *testing.T) {
	it := Item{
		Title:    "Scan staged changes (pre-commit)",
		Group:    "Scan",
		Summary:  "The same gate the pre-commit hook runs.",
		Keywords: []string{"scan", "staged", "pre commit", "hook", "git"},
	}
	if ScoreItem("staged hook", it) == 0 {
		t.Error("both tokens are present, expected a match")
	}
	if n := ScoreItem("staged aws", it); n != 0 {
		t.Errorf("'aws' is absent, so the query must not match (scored %d) — "+
			"multi-token queries narrow, they do not widen", n)
	}
}

func TestScoreItemWeightsTitleAboveSummary(t *testing.T) {
	titled := Item{Title: "secrets", Summary: "unrelated"}
	summarised := Item{Title: "unrelated", Summary: "secrets"}
	if ScoreItem("secrets", titled) <= ScoreItem("secrets", summarised) {
		t.Error("a title hit must outrank a summary hit")
	}
}

func TestRankOrdersAndKeepsTiesStable(t *testing.T) {
	items := []Item{
		{Title: "Scan changes vs main", Keywords: []string{"diff"}},
		{Title: "Scan this project", Keywords: []string{"scan", "audit"}},
		{Title: "Apply fixes", Keywords: []string{"fix"}},
	}
	got := Rank("scan", items)
	if len(got) != 2 {
		t.Fatalf("Rank(\"scan\") returned %d items, want 2", len(got))
	}
	if got[0].Title != "Scan this project" {
		t.Errorf("best match = %q, want the exact-title item", got[0].Title)
	}

	// Ties keep catalog order so redraws never shuffle the list.
	tie := []Item{{Title: "alpha"}, {Title: "beta"}, {Title: "gamma"}}
	all := Rank("", tie)
	for i := range tie {
		if all[i].Title != tie[i].Title {
			t.Fatalf("empty query reordered the catalog: %v", all)
		}
	}
}

func TestRankReturnsNothingWhenNoMatch(t *testing.T) {
	items := []Item{{Title: "Scan this project"}}
	if got := Rank("zzzzzz", items); len(got) != 0 {
		t.Errorf("Rank on a miss returned %d items, want 0", len(got))
	}
}

func TestSubsequenceIsOrderSensitive(t *testing.T) {
	if _, ok := subsequence("abcd", "dcba"); ok {
		t.Error("subsequence must respect order")
	}
	if _, ok := subsequence("vs017", "vs sec 017"); !ok {
		t.Error("expected 'vs017' to fuzzily reach 'vs sec 017'")
	}
}

// Short tokens must never reach the fuzzy path: "aws" subsequence-matches
// "CORS allows all origins" and "sec" matches most English sentences, which
// would make a three-letter search return the entire catalog.
func TestShortTokensDoNotFuzzyMatch(t *testing.T) {
	if _, ok := subsequence("aws", "cors allows all origins"); ok {
		t.Error("'aws' must not subsequence-match 'cors allows all origins'")
	}
	if n := scoreText("aws", "CORS allows all origins"); n != 0 {
		t.Errorf("scoreText(aws, cors allows all origins) = %d, want 0", n)
	}
	if _, ok := subsequence("sec", "vs dep 017 deprecated comment marker"); ok {
		t.Error("'sec' must not subsequence-match an unrelated rule title")
	}
}

// A long query still gets fuzzy recall, which is what makes id searches work
// when the user forgets the separators.
func TestLongTokensStillFuzzyMatch(t *testing.T) {
	if n := scoreText("vssec017", "VS-SEC-017  OpenAI API key hardcoded"); n == 0 {
		t.Error("expected 'vssec017' to reach the rule title")
	}
	if n := scoreText("cursorrules", "Cursor"); n != 0 {
		t.Errorf("scoreText(cursorrules, Cursor) = %d, want 0 — the alias keyword carries that match", n)
	}
}

// End-to-end guard on the ranking that a user actually sees: searching a
// three-letter term must not return a wall of unrelated rules.
func TestSearchForAWSTermIsNotNoisy(t *testing.T) {
	items := []Item{
		{Title: "AWS access key ID embedded in source", Keywords: []string{"VS-SEC-001", "hardcoded-secret", "aws"}},
		{Title: "CORS allows all origins", Keywords: []string{"VS-SEC-052", "insecure-default", "cors"}},
		{Title: "Dependency declared without a version bound", Keywords: []string{"VS-DEP-001", "dependency-risk"}},
	}
	got := Rank("aws", items)
	if len(got) != 1 {
		t.Fatalf("Rank(aws) returned %d items, want only the AWS rule: %v", len(got), titles(got))
	}
	if got[0].Title != "AWS access key ID embedded in source" {
		t.Errorf("best match = %q, want the AWS rule", got[0].Title)
	}
}

func titles(items []Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Title
	}
	return out
}
