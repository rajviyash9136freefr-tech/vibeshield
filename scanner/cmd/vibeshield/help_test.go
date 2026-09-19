package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// dispatched is every verb run() answers to, aliases included. It is written
// out by hand on purpose: the test exists to fail when a command is added to
// the dispatcher without a matching entry in docs(), which is exactly the kind
// of drift that leaves `vibeshield <newthing> --help` printing the wrong manual.
var dispatched = []string{
	"scan", "fix", "init", "doctor", "search", "find", "rules", "rule",
	"agents", "agent", "completion", "completions", "ui", "menu", "console",
	"version", "help",
}

func TestEveryDispatchedCommandHasHelp(t *testing.T) {
	for _, cmd := range dispatched {
		if _, ok := docsIndex()[cmd]; !ok {
			t.Errorf("`vibeshield %s` is dispatched but has no entry in docs()", cmd)
		}
	}
}

func TestEveryDocumentedCommandIsDispatched(t *testing.T) {
	known := map[string]bool{}
	for _, c := range dispatched {
		known[c] = true
	}
	for _, d := range docs() {
		if !known[d.Name] {
			t.Errorf("docs() documents %q but run() does not dispatch it", d.Name)
		}
	}
}

// `-h` has to work on every command, and it has to answer with that command's
// help rather than the global manual. Before v3 every subcommand printed the
// top-level usage, so `vibeshield scan --help` never mentioned --diff.
func TestHelpFlagPrintsTheCommandOwnHelp(t *testing.T) {
	cases := []struct{ cmd, want string }{
		{"scan", "--staged"},
		{"fix", "--dry-run"},
		{"init", "--no-workflow"},
		{"doctor", "--format"},
		{"search", "--agents"},
		{"rules", "--limit"},
		{"agents", "--markdown"},
		{"completion", "powershell"},
	}
	for _, tc := range cases {
		var out, errOut bytes.Buffer
		code := run([]string{tc.cmd, "--help"}, &out, &errOut)
		if code != 0 {
			t.Errorf("%s --help exited %d, want 0", tc.cmd, code)
			continue
		}
		text := out.String() + errOut.String()
		if !strings.Contains(text, tc.want) {
			t.Errorf("%s --help did not mention %q\n---\n%s", tc.cmd, tc.want, text)
		}
		if !strings.Contains(text, "vibeshield "+tc.cmd+" —") {
			t.Errorf("%s --help did not print its own header\n---\n%s", tc.cmd, text)
		}
	}
}

func TestHelpCommandAcceptsACommandName(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"help", "scan"}, &out, &errOut); code != 0 {
		t.Fatalf("`vibeshield help scan` exited %d, want 0", code)
	}
	if !strings.Contains(out.String(), "vibeshield scan —") {
		t.Errorf("`vibeshield help scan` did not print the scan help:\n%s", out.String())
	}

	out.Reset()
	errOut.Reset()
	if code := run([]string{"help", "nonsense"}, &out, &errOut); code != 2 {
		t.Errorf("`vibeshield help nonsense` exited %d, want 2", code)
	}
}

func TestBareInvocationOnAPipePrintsHelpAndExitsTwo(t *testing.T) {
	// Scripts that pipe into vibeshield must keep getting the usage text and
	// exit 2 — the console only opens on a real terminal. run() reads os.Stdin
	// directly, so this test can only assert the branch when the test process
	// itself does not have a terminal on stdin.
	if stdinIsInteractive(os.Stdin) {
		t.Skip("stdin is a terminal here, so run() would open the console")
	}
	var out, errOut bytes.Buffer
	code := run(nil, &out, &errOut)
	if code != 2 {
		t.Errorf("bare vibeshield on a pipe exited %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "Start here") {
		t.Errorf("bare vibeshield did not print the help:\n%s", errOut.String())
	}
}

func TestUnknownCommandSuggestsTheNearestVerb(t *testing.T) {
	cases := map[string]string{
		"scna":     "scan",
		"scn":      "scan",
		"fxi":      "fix",
		"doctr":    "doctor",
		"serach":   "search",
		"complete": "completion",
	}
	for typo, want := range cases {
		if got := suggest(typo); got != want {
			t.Errorf("suggest(%q) = %q, want %q", typo, got, want)
		}
	}
	// A word with nothing in common must not produce a confident wrong answer.
	if got := suggest("kubernetes"); got != "" {
		t.Errorf("suggest(%q) = %q, want no suggestion", "kubernetes", got)
	}
}

func TestUnknownCommandExitsTwoAndNamesTheSuggestion(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"scna"}, &out, &errOut); code != 2 {
		t.Fatalf("exited %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "Did you mean `vibeshield scan`?") {
		t.Errorf("no suggestion in:\n%s", errOut.String())
	}
}

// Completion has to know about every flag the help text advertises, or Tab
// silently stops working for a flag that exists.
func TestCompletionCoversEveryDocumentedSubcommandFlag(t *testing.T) {
	for _, d := range docs() {
		flags, hasCompletion := completionFlags[d.Name]
		if !hasCompletion {
			continue // commands without flags, or meta commands
		}
		// Every flag the completion offers must be a real flag, i.e. the
		// command's own help body has to mention it.
		for _, f := range flags {
			if !strings.Contains(d.Body, f) {
				t.Errorf("%s: completion offers %s but the help text never mentions it", d.Name, f)
			}
		}
	}
}

// The reverse direction: a flag documented in a help body should be offered by
// completion, so the two tables cannot drift apart in the quiet direction.
func TestDocumentedFlagsAreCompletable(t *testing.T) {
	for _, d := range docs() {
		flags, hasCompletion := completionFlags[d.Name]
		if !hasCompletion {
			continue
		}
		offered := map[string]bool{}
		for _, f := range flags {
			offered[f] = true
		}
		for _, line := range strings.Split(d.Body, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "--") && !strings.HasPrefix(trimmed, "-v") {
				continue
			}
			name := strings.Fields(trimmed)[0]
			name = strings.TrimSuffix(name, ",")
			if name == "--help" || name == "-h" {
				continue
			}
			if !offered[name] {
				t.Errorf("%s: help documents %s but completion does not offer it", d.Name, name)
			}
		}
	}
}

func TestCompletionEmitsAScriptForEveryShell(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		var out, errOut bytes.Buffer
		if code := run([]string{"completion", shell}, &out, &errOut); code != 0 {
			t.Errorf("completion %s exited %d, want 0", shell, code)
			continue
		}
		if out.Len() == 0 {
			t.Errorf("completion %s printed nothing", shell)
		}
		if !strings.Contains(out.String(), "vibeshield") {
			t.Errorf("completion %s does not mention vibeshield", shell)
		}
		// Every verb should be completable.
		for _, cmd := range completionCommands {
			if !strings.Contains(out.String(), cmd) {
				t.Errorf("completion %s is missing the %q verb", shell, cmd)
			}
		}
	}
}

func TestCompletionRejectsAnUnknownShell(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"completion", "tcsh"}, &out, &errOut); code != 2 {
		t.Errorf("completion tcsh exited %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "unknown shell") {
		t.Errorf("no explanation in:\n%s", errOut.String())
	}
}

// doctor must never exit 0 on a broken setup, and must never modify anything.
func TestDoctorReportsAndExitsNonZeroWhenSomethingIsMissing(t *testing.T) {
	dir := t.TempDir() // empty: no config, no git, no workflow
	var out, errOut bytes.Buffer
	code := run([]string{"doctor", dir, "--no-color"}, &out, &errOut)
	if code != 1 {
		t.Errorf("doctor on an empty dir exited %d, want 1\n%s", code, out.String())
	}
	text := out.String()
	for _, want := range []string{"VibeShield " + Version + " — doctor", "pr gate", "→"} {
		if !strings.Contains(text, want) {
			t.Errorf("doctor output is missing %q:\n%s", want, text)
		}
	}
}

func TestDoctorJSONIsMachineReadable(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	code := run([]string{"doctor", dir, "--format", "json"}, &out, &errOut)
	if code != 1 {
		t.Errorf("doctor --format json exited %d, want 1", code)
	}
	for _, want := range []string{`"schema": "vibeshield.doctor/v1"`, `"checks":`, `"status":`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("doctor json is missing %q:\n%s", want, out.String())
		}
	}
}

func TestDoctorRejectsAnUnknownFormat(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"doctor", t.TempDir(), "--format", "xml"}, &out, &errOut); code != 2 {
		t.Errorf("doctor --format xml exited %d, want 2", code)
	}
}

// `rules` is `search --rules`, and the two must agree about what a rule is.
func TestRulesListsOnlyRules(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"rules", "--no-color"}, &out, &errOut); code != 0 {
		t.Fatalf("rules exited %d, want 0\n%s", code, errOut.String())
	}
	text := out.String()
	if !strings.Contains(text, "VS-SEC-") {
		t.Errorf("rules printed no rule ids:\n%s", text)
	}
	// Agent recipes are not rules and must not appear.
	if strings.Contains(text, "AGENT SETUP") {
		t.Errorf("rules leaked the agent-setup catalog:\n%s", text)
	}
}

func TestRulesAcceptsAPositionalQueryBeforeAFlag(t *testing.T) {
	// Regression: `--rules` is a switch here but a value flag in `scan`, and a
	// shared table made this invocation fail with a flag parse error.
	var out, errOut bytes.Buffer
	if code := run([]string{"rules", "VS-SEC-017", "--no-color"}, &out, &errOut); code != 0 {
		t.Fatalf("rules VS-SEC-017 --no-color exited %d, want 0\n%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "VS-SEC-017") {
		t.Errorf("the rule did not come back:\n%s", out.String())
	}
}

func TestRulesEmptyQueryIsNotAnError(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"rules", "zzzznotathing", "--no-color"}, &out, &errOut); code != 0 {
		t.Errorf("a rule search with no hits exited %d, want 0", code)
	}
	if !strings.Contains(out.String(), "nothing matched") {
		t.Errorf("an empty result should point somewhere:\n%s", out.String())
	}
}

func TestScanHelpListsEveryFormatTheBinaryAccepts(t *testing.T) {
	text, ok := usageFor("scan", Version)
	if !ok {
		t.Fatal("no help for scan")
	}
	for _, f := range []string{"pretty", "json", "github", "sarif"} {
		if !strings.Contains(text, f) {
			t.Errorf("scan help does not mention --format %s", f)
		}
	}
}

// `--list` promises everything, so the default --limit 20 must not quietly cut
// it down. The README documents `vibeshield search --rules --list` as "every
// shipped rule", and it printed 20 entries.
func TestListIsNotTruncatedByTheDefaultLimit(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"search", "--rules", "--list", "--no-color"}, &out, &errOut); code != 0 {
		t.Fatalf("search --rules --list exited %d, want 0\n%s", code, errOut.String())
	}
	if strings.Contains(out.String(), "20 catalog entries") {
		t.Errorf("--list was truncated by the default limit:\n%s", out.String())
	}

	// An explicit --limit still applies.
	out.Reset()
	errOut.Reset()
	if code := run([]string{"search", "--rules", "--list", "--limit", "3", "--no-color"}, &out, &errOut); code != 0 {
		t.Fatalf("explicit limit exited %d, want 0", code)
	}
	if !strings.Contains(out.String(), "3 catalog entries") {
		t.Errorf("an explicit --limit was ignored:\n%s", out.String())
	}
}
