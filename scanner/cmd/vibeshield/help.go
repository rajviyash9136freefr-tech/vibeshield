// Help text for the v3 CLI.
//
// Design rules, borrowed from clig.dev because they are the ones that actually
// help a first-time CLI user:
//
//   - lead with examples, not prose — people copy a command before they read a
//     sentence about it;
//   - every subcommand has its own help, reachable as `vibeshield <cmd> --help`
//     and as `vibeshield help <cmd>`, so `-h` works anywhere;
//   - say what the exit codes mean, because that is the whole contract for a
//     gate that runs in CI;
//   - end with the next command to run, so the tool teaches itself.
//
// Keep this file free of runtime state: it is pure text, so the binary stays a
// single static build with no dependencies.
package main

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
)

// commandDoc is one row of the command table, used for both the top-level
// listing and `vibeshield help <command>`.
type commandDoc struct {
	Name    string // canonical verb
	Aliases []string
	Group   string // grouping in the top-level listing
	Summary string // one line, imperative
	Usage   string // "vibeshield scan [path] [flags]"
	Body    string // the long help: examples first, then flags
}

// docs is the single source of truth for help text. A command that is
// dispatched in run() but missing here is a documentation bug, and
// help_test.go fails the build when that happens.
func docs() []commandDoc {
	return []commandDoc{
		{
			Name:    "scan",
			Group:   "Everyday",
			Summary: "Audit a project, a git diff, or the staged changes",
			Usage:   "vibeshield scan [path] [flags]",
			Body: `Audit a directory for the failure modes AI coding agents actually
produce: invented package names, secrets pasted from chat, insecure
scaffolding, stripped licences. Fully offline — nothing leaves the machine.

Examples:

  vibeshield scan .                     audit the current project
  vibeshield scan ../api                audit another project (its own config wins)
  vibeshield scan --diff main           only findings on lines you added vs main
  vibeshield scan --staged              what the pre-commit hook runs
  vibeshield scan . --format json       machine-readable, for agents and CI
  vibeshield scan . --format sarif      upload to GitHub code scanning
  vibeshield scan . --mode block-on-critical   fail the build on a critical

Flags:

  --diff <ref|->     Diff mode: scan changes vs a git ref, or "-" for stdin
  --staged           Pre-commit mode: scan git staged changes
  --format <fmt>     pretty (default) | json | github | sarif
                     github = ::error/::warning annotations in the log
                     sarif  = GitHub code-scanning upload (upload-sarif)
  --config <file>    Config path (default: <path>/vibeshield.yml, then ./vibeshield.yml)
  --mode <mode>      off | warn | block-on-critical | block-on-high+
  --rules <dir>      Load extra rule packs from a directory
  --online           (reserved) allow package-intel network lookups
  --max-cols <n>     Output width cap (default 88)
  --no-color         Disable colour (also: NO_COLOR env, non-TTY auto)
  -v, --verbose      Explain what was scanned, on stderr

Exit codes: 0 clean or warn-mode findings · 1 block threshold met · 2 usage or config error

Next: vibeshield fix . --dry-run      preview the fixes it can make for you`,
		},
		{
			Name:    "fix",
			Group:   "Everyday",
			Summary: "Preview and apply the mechanical fixes (VibePatch)",
			Usage:   "vibeshield fix [path] [flags]",
			Body: `VibePatch applies the one-line fixes the rules already know how to make.
It is opt-in and human-gated: nothing is written until you say so, and every
patch is appended to vibeshield-fixes.log (JSONL) so the change trail stays
auditable.

Secrets and prompt-injection findings are never patched mechanically — a key
needs rotation, and an agent editing its own instruction file is the injection
we came to catch.

Examples:

  vibeshield fix . --dry-run            show the −/+ diff, change nothing
  vibeshield fix .                      prompt per file: [y/N/a/s]
  vibeshield fix . --yes                apply everything, no prompts (agents / CI)
  vibeshield fix . --report scan.json   reuse a previous --format json scan

Flags:

  --dry-run          Preview the diff, change nothing
  --yes              Apply without prompting; every patch is still logged
  --report <file>    Reuse a --format json scan instead of rescanning
  --config <file>    Config path (default: vibeshield.yml if present)
  --rules <dir>      Load extra rule packs from a directory
  --no-color         Disable colour
  -v, --verbose      Explain what was scanned, on stderr

A non-interactive stdin must pass --dry-run or --yes: the gate never guesses.

Exit codes: 0 applied or nothing to do · 2 usage or config error`,
		},
		{
			Name:    "init",
			Group:   "Everyday",
			Summary: "Set a project up: config, PR gate, pre-commit hook, first scan",
			Usage:   "vibeshield init [path] [flags]",
			Body: `Detects the stack from the project's manifests, then writes at most three
files — vibeshield.yml, .github/workflows/vibeshield.yml and
.git/hooks/pre-commit — and runs the first scan.

It is deliberately conservative: --dry-run prints the whole plan first, an
existing file is never overwritten without --force, and an existing git hook is
never touched at all.

Examples:

  vibeshield init --dry-run             show the plan, write nothing
  vibeshield init                       set the project up and scan it
  vibeshield init . --mode block-on-critical
  vibeshield init . --no-hook           CI repo: workflow but no local hook

Flags:

  --mode <mode>      Gate mode written to vibeshield.yml (default: warn)
  --dry-run          Show what would be written, change nothing
  --force            Overwrite files that already exist
  --no-hook          Skip the git pre-commit hook
  --no-workflow      Skip the GitHub Action workflow
  --no-scan          Skip the first scan
  --no-color         Disable colour

Exit codes: 0 written (or planned) · 2 usage error

Next: vibeshield doctor              check the setup you just wrote`,
		},
		{
			Name:    "doctor",
			Group:   "Everyday",
			Summary: "Check the install, the config and the git hooks",
			Usage:   "vibeshield doctor [path] [flags]",
			Body: `Reads the project and reports what is set up and what is not, with the
exact command that fixes each gap. Nothing is modified.

Use it when a scan behaves unexpectedly, after cloning a repo, or in CI to
assert that the gate is actually wired up.

Examples:

  vibeshield doctor                     check the current project
  vibeshield doctor ../api              check another project
  vibeshield doctor --format json       machine-readable checklist

Flags:

  --config <file>    Config path to check (default: vibeshield.yml if present)
  --format <fmt>     pretty (default) | json
  --no-color         Disable colour
  -v, --verbose      Print every check, including the ones that passed

Exit codes: 0 healthy · 1 something needs fixing · 2 usage error`,
		},
		{
			Name:    "search",
			Aliases: []string{"find"},
			Group:   "Explore",
			Summary: "Search every rule, action and agent recipe",
			Usage:   "vibeshield search [flags] [query]",
			Body: `The ranking behind the interactive console, on stdout — so scripts and AI
agents can ask "what do you know about aws keys?" without a TTY.

Multi-word queries narrow rather than widen: every token must match.

Examples:

  vibeshield search aws                 AWS credential rules
  vibeshield search vs-sec-017          one rule, by id
  vibeshield search "prompt injection"  every token must match
  vibeshield search --agents cursor     agent setup recipes only
  vibeshield search --rules --list      every shipped rule
  vibeshield search aws --format json   vibeshield.search/v1 for agents and CI

Flags:

  --list             List every catalog entry instead of searching
  --rules            Search rule packs only
  --agents           Search agent setup recipes only
  --limit <n>        Max results (default 20)
  --format <fmt>     pretty (default) | json
  --no-color         Disable colour

Exit codes: 0 always (an empty result is not an error) · 2 usage error`,
		},
		{
			Name:    "rules",
			Aliases: []string{"rule"},
			Group:   "Explore",
			Summary: "List the rule packs, or read one rule in full",
			Usage:   "vibeshield rules [id] [flags]",
			Body: `A shortcut for the most common search: no query lists every rule grouped by
category, an id or a word reads one rule in full.

Examples:

  vibeshield rules                      every rule, grouped by category
  vibeshield rules VS-SEC-017           one rule, in full
  vibeshield rules cors                 rules matching a word
  vibeshield rules --format json        machine-readable list

Flags:

  --limit <n>        Max results (default 0 = no limit)
  --format <fmt>     pretty (default) | json
  --no-color         Disable colour

Exit codes: 0 always · 2 usage error`,
		},
		{
			Name:    "agents",
			Aliases: []string{"agent"},
			Group:   "Explore",
			Summary: "Per-agent setup recipes (Codex, Claude Code, Cursor, …)",
			Usage:   "vibeshield agents [--body|--markdown] [name]",
			Body: `Prints the setup recipe for each coding agent VibeShield ships rules for.
Every file is generated from one source of truth and checked in CI, so the
docs cannot claim support the binary does not ship.

Examples:

  vibeshield agents                     every recipe
  vibeshield agents cursor              one agent
  vibeshield agents --body > AGENTS.md  just the pasteable rule block
  vibeshield agents --markdown          the support matrix as a table

Flags:

  --body             Print only the shared rule block
  --markdown         Print the agent matrix as markdown

Exit codes: 0 · 2 unknown agent name`,
		},
		{
			Name:    "completion",
			Aliases: []string{"completions"},
			Group:   "Explore",
			Summary: "Print a shell completion script",
			Usage:   "vibeshield completion <bash|zsh|fish|powershell>",
			Body: `Prints a completion script for the given shell on stdout. Source it once and
every command and flag completes with Tab.

Examples:

  vibeshield completion bash   >> ~/.bashrc
  vibeshield completion zsh    > "${fpath[1]}/_vibeshield"
  vibeshield completion fish   > ~/.config/fish/completions/vibeshield.fish
  vibeshield completion powershell | Out-String | Invoke-Expression

Exit codes: 0 · 2 unknown or missing shell name`,
		},
		{
			Name:    "version",
			Group:   "Meta",
			Summary: "Version, rule-pack info and engine coverage",
			Usage:   "vibeshield version",
			Body: `Prints the binary version, the embedded rule packs and how many rules the
engine can actually evaluate. Reserved rules load and validate so packs stay
portable across versions, but the matcher skips them — this command says so
rather than overstating the rule count.

Also available as: vibeshield --version`,
		},
		{
			Name:    "ui",
			Aliases: []string{"console", "menu"},
			Group:   "Meta",
			Summary: "Open the interactive console (same as a bare vibeshield)",
			Usage:   "vibeshield ui",
			Body: `One search box over every action, rule and agent recipe. Selecting an action
hands the terminal back and runs the real command, so the menu can never drift
from the documented flags.

Keys: ↑/↓ move · Tab next group · Enter open · Esc clear then quit · Ctrl+U reset · Ctrl+C quit

A bare ` + "`vibeshield`" + ` opens the same console on a terminal, and falls back to
this help text when stdin is a pipe — so scripts are unaffected.

Exit codes: 0 quit normally · 2 rules failed to load`,
		},
		{
			Name:    "help",
			Group:   "Meta",
			Summary: "Help for a command",
			Usage:   "vibeshield help [command]",
			Body: `Prints the full help. Every command also accepts -h and --help.

Examples:

  vibeshield help                this overview
  vibeshield help scan           flags and examples for one command
  vibeshield scan --help         the same thing`,
		},
	}
}

// docsIndex maps every accepted verb (canonical name and aliases) to its doc.
func docsIndex() map[string]commandDoc {
	idx := map[string]commandDoc{}
	for _, d := range docs() {
		idx[d.Name] = d
		for _, a := range d.Aliases {
			idx[a] = d
		}
	}
	return idx
}

// usageTop is the help a bare `vibeshield` prints when stdin is not a terminal.
// Examples first, then the command table, then the exit-code contract.
func usageTop(version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "vibeshield %s — security scanner for AI-generated code\n\n", version)
	b.WriteString(`  Finds what AI coding agents leave behind — hallucinated packages,
  secrets pasted from chat, insecure defaults, stripped licences — and
  gives you a one-line fix for each. Static analysis only: no code,
  key or prompt ever leaves this machine.

Start here:

  vibeshield init              set this project up, then scan it
  vibeshield scan .            audit everything, right now
  vibeshield scan --staged     audit only what you are about to commit
  vibeshield doctor            check what is wired up, and what is not

`)

	groups := []string{"Everyday", "Explore", "Meta"}
	for _, g := range groups {
		fmt.Fprintf(&b, "%s:\n\n", g)
		for _, d := range docs() {
			if d.Group != g {
				continue
			}
			fmt.Fprintf(&b, "  %-13s %s\n", d.Name, d.Summary)
		}
		b.WriteString("\n")
	}

	b.WriteString(`Every command takes -h/--help, and "vibeshield help <command>" works too.

Exit codes: 0 clean or warn-mode findings · 1 findings at the block threshold · 2 usage or config error
Docs & examples: https://github.com/rajviyash9136freefr-tech/vibeshield#readme
`)
	return b.String()
}

// usageFor renders the long help for one command. Unknown names fall back to
// the top-level help so `vibeshield help nonsense` still teaches something.
func usageFor(cmd, version string) (string, bool) {
	d, ok := docsIndex()[strings.ToLower(cmd)]
	if !ok {
		return usageTop(version), false
	}
	var b strings.Builder
	fmt.Fprintf(&b, "vibeshield %s — %s\n\n", d.Name, strings.ToLower(d.Summary))
	fmt.Fprintf(&b, "Usage:\n  %s\n\n", d.Usage)
	b.WriteString(d.Body)
	if !strings.HasSuffix(d.Body, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\nGlobal: -h/--help anywhere · vibeshield help <command> · vibeshield version\n")
	return b.String(), true
}

// suggest returns the closest known verb to an unknown one, or "" when nothing
// is close enough. clig.dev asks for this: a typo should teach the correct
// spelling instead of dumping the whole manual.
//
// Similarity is 1 − distance/length, with transpositions counted as one edit —
// "scna" and "fxi" are the typos people actually make, and plain Levenshtein
// scores each of them as two edits. The 0.6 floor is deliberately conservative:
// a confident wrong suggestion is worse than none, so "kubernetes" gets silence
// rather than "completion".
func suggest(cmd string) string {
	best, bestRatio := "", 0.0
	for _, d := range docs() {
		for _, name := range append([]string{d.Name}, d.Aliases...) {
			n := max(len(cmd), len(name))
			if n == 0 {
				continue
			}
			ratio := 1 - float64(damerau(cmd, name))/float64(n)
			if ratio > bestRatio {
				best, bestRatio = d.Name, ratio
			}
		}
	}
	if bestRatio < 0.6 {
		return ""
	}
	return best
}

// damerau is the Optimal String Alignment distance: Levenshtein plus adjacent
// transpositions. Command names are short, so the O(n·m) table is fine.
func damerau(a, b string) int {
	ar, br := []rune(a), []rune(b)
	d := make([][]int, len(ar)+1)
	for i := range d {
		d[i] = make([]int, len(br)+1)
		d[i][0] = i
	}
	for j := 0; j <= len(br); j++ {
		d[0][j] = j
	}
	for i := 1; i <= len(ar); i++ {
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+cost))
			if i > 1 && j > 1 && ar[i-1] == br[j-2] && ar[i-2] == br[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(ar)][len(br)]
}

// knownCommands lists every dispatched verb, aliases included. Used by the
// console catalog and by tests that guard against help text drift.
func knownCommands() []string {
	var out []string
	for _, d := range docs() {
		out = append(out, d.Name)
		out = append(out, d.Aliases...)
	}
	sort.Strings(out)
	return out
}

// --- shared flag plumbing ---------------------------------------------------
//
// Every subcommand builds its FlagSet through newFlagSet so that one rule holds
// everywhere: `-h`/`--help` prints that command's own help and exits 0, and a
// bad flag prints that command's help and exits 2. Before v3 each subcommand
// printed the *global* manual, which meant `vibeshield scan --help` answered a
// question nobody asked.

// newFlagSet returns a FlagSet wired to the per-command help for name.
func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fl := flag.NewFlagSet(name, flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { printUsage(stderr, name) }
	return fl
}

// printUsage writes the long help for one command to w.
func printUsage(w io.Writer, cmd string) {
	text, ok := usageFor(cmd, Version)
	if !ok {
		// usageFor already fell back to the top-level help; nothing else to do.
		fmt.Fprint(w, text)
		return
	}
	fmt.Fprint(w, text)
}

// usageExit maps a flag-parse error to the documented exit codes: asking for
// help is a success, a malformed invocation is not.
func usageExit(err error) int {
	if err == flag.ErrHelp {
		return 0
	}
	return 2
}
