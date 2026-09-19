// Command vibeshield is the VibeShield CLI: a static security & dependency
// auditor for AI-generated code (contracts/cli.md). It never sends code
// anywhere; --online opts into package-intel lookups (v1.1).
//
// v3 is the "usable from a terminal" release. The scanner already worked; what
// changed is that a person who has never seen it can now find their way:
// every command has its own help, `vibeshield doctor` says what is wired up
// and what is not, `vibeshield completion` adds Tab completion, and a typo
// suggests the command that was meant instead of printing the whole manual.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/cli"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/config"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/fix"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/initcmd"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/output"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scandiff"
)

// Version is stamped by -ldflags "-X main.Version=v1.2.3" at release build.
var Version = "3.0.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// A bare invocation opens the interactive console. Scripts that pipe
		// into vibeshield still get the help text and exit 2, so no existing
		// automation changes behaviour.
		if stdinIsInteractive(os.Stdin) {
			return cmdConsole(stdout, stderr)
		}
		fmt.Fprint(stderr, usageTop(Version))
		return 2
	}
	switch args[0] {
	case "version", "--version", "-v", "-V":
		return cmdVersion(stdout)
	case "scan":
		return cmdScan(args[1:], stdout, stderr)
	case "fix":
		return cmdFix(args[1:], os.Stdin, stdout, stderr)
	case "init":
		return cmdInit(args[1:], stdout, stderr)
	case "doctor":
		return cmdDoctor(args[1:], stdout, stderr)
	case "search", "find":
		return cmdSearch(args[1:], stdout, stderr)
	case "rules", "rule":
		return cmdRules(args[1:], stdout, stderr)
	case "agents", "agent":
		return cmdAgents(args[1:], stdout, stderr)
	case "completion", "completions":
		return cmdCompletion(args[1:], stdout, stderr)
	case "ui", "menu", "console":
		return cmdConsole(stdout, stderr)
	case "help", "--help", "-h":
		return cmdHelp(args[1:], stdout, stderr)
	default:
		return unknownCommand(args[0], stderr)
	}
}

// unknownCommand is the typo path. Printing the whole manual for a misspelled
// verb buries the one line the user needs, so a close match is named first.
func unknownCommand(arg string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "vibeshield: unknown command %q\n", arg)
	if s := suggest(arg); s != "" {
		fmt.Fprintf(stderr, "\n  Did you mean `vibeshield %s`?\n", s)
	}
	fmt.Fprintf(stderr, "\n  Run `vibeshield help` for the command list, or `vibeshield` for the console.\n")
	return 2
}

// cmdHelp implements `vibeshield help [command]`. With no argument it prints
// the overview; with one it prints that command's long help.
func cmdHelp(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, usageTop(Version))
		return 0
	}
	text, ok := usageFor(args[0], Version)
	if !ok {
		fmt.Fprintf(stderr, "vibeshield: no help for %q\n", args[0])
		if s := suggest(args[0]); s != "" {
			fmt.Fprintf(stderr, "\n  Did you mean `vibeshield help %s`?\n", s)
		}
		return 2
	}
	fmt.Fprint(stdout, text)
	return 0
}

// cmdConsole opens the interactive console: one search box over every action,
// rule and agent recipe. Actions are replayed through run(), so the menu can
// never drift from the documented flags.
func cmdConsole(stdout, stderr io.Writer) int {
	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: core rules failed to load: %v\n", err)
		return 2
	}
	c := &cli.Console{
		In:      os.Stdin,
		Out:     stdout,
		Items:   cli.Catalog(pack, Version),
		Version: Version,
		Color:   output.IsTTY(stdout),
		Exec:    func(a []string) int { return run(a, stdout, stderr) },
	}
	return c.Run()
}

func cmdVersion(w io.Writer) int {
	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(w, "vibeshield %s — rule packs failed to load: %v\n", Version, err)
		return 2
	}
	fmt.Fprintf(w, "vibeshield %s\n", Version)
	fmt.Fprintf(w, "rule packs: %s %s (%s, %d rules)\n", pack.ID, pack.Version, pack.License, len(pack.Rules))
	// A pack is deliberately larger than the engine: rules the matcher cannot
	// evaluate yet still load so packs stay portable. Saying only "122 rules"
	// would overstate what a scan can find, so the split is reported.
	if reserved := pack.ReservedRules(); reserved > 0 {
		fmt.Fprintf(w, "engine:     %d active · %d reserved (structural — pending the package-intel model)\n",
			pack.ActiveRules(), reserved)
	} else {
		fmt.Fprintf(w, "engine:     %d active\n", pack.ActiveRules())
	}
	fmt.Fprintf(w, "no code leaves this machine: static analysis only\n")
	return 0
}

// resolveConfigPath picks the config file for a scan. An explicit --config
// always wins. Otherwise the scanned project's own vibeshield.yml is preferred
// over the working directory's: `vibeshield scan ../other-project` must honour
// that project's configuration, not silently ignore it because the shell
// happened to be somewhere else. When the scan path is "." both candidates are
// the same file, so nothing changes for the common case.
func resolveConfigPath(fl *flag.FlagSet, configured, scanPath string) string {
	explicit := false
	fl.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			explicit = true
		}
	})
	if explicit {
		return configured
	}
	candidate := filepath.Join(scanPath, "vibeshield.yml")
	if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
		return candidate
	}
	return configured
}

// cmdInit is the setup path documented in contracts/cli.md: detect the stack,
// write vibeshield.yml + a PR-gate workflow + a pre-commit hook, then run a
// first scan. It is deliberately conservative — it never overwrites a file
// (least of all a git hook) without --force, and --dry-run prints the whole
// plan first so nothing is a surprise.
func cmdInit(args []string, stdout, stderr io.Writer) int {
	args = normalizeScanArgs("init", args)
	fl := newFlagSet("init", stderr)
	var (
		mode       = fl.String("mode", "warn", "initial gate mode written to vibeshield.yml")
		dryRun     = fl.Bool("dry-run", false, "show what would be written, change nothing")
		force      = fl.Bool("force", false, "overwrite files that already exist")
		noHook     = fl.Bool("no-hook", false, "skip the git pre-commit hook")
		noWorkflow = fl.Bool("no-workflow", false, "skip the GitHub Action workflow")
		noScan     = fl.Bool("no-scan", false, "skip the first scan")
		noColor    = fl.Bool("no-color", false, "disable color")
	)
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}
	if !config.AllowedModes[*mode] {
		fmt.Fprintf(stderr, "vibeshield: bad --mode %q (off|warn|block-on-critical|block-on-high+)\n", *mode)
		return 2
	}
	path := "."
	if fl.NArg() > 0 {
		path = fl.Arg(0)
	}
	self, err := os.Executable()
	if err != nil {
		self = ""
	}

	plan, err := initcmd.BuildPlan(initcmd.Options{
		Dir: path, Mode: *mode, Version: Version,
		Force: *force, Hook: !*noHook, Workflow: !*noWorkflow,
		BinaryPath: self,
	})
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}

	color := !*noColor && output.IsTTY(stdout)
	paint := func(code, s string) string {
		if !color {
			return s
		}
		return code + s + "\x1b[0m"
	}
	const (
		cBold   = "\x1b[1m"
		cDim    = "\x1b[90m"
		cGreen  = "\x1b[32m"
		cYellow = "\x1b[33m"
	)

	fmt.Fprintf(stdout, "\n  %s\n", paint(cBold, "VibeShield "+Version+" — project setup"))
	if plan.Stack.Empty() {
		fmt.Fprintf(stdout, "  %s\n", paint(cDim, "No manifest recognised at the root — every language will be scanned."))
	} else {
		fmt.Fprintf(stdout, "  Detected   %s\n", strings.Join(plan.Stack.Languages, ", "))
		if len(plan.Stack.Ecosystems) > 0 {
			fmt.Fprintf(stdout, "  Ecosystem  %s\n", strings.Join(plan.Stack.Ecosystems, ", "))
		}
		if len(plan.Stack.Frameworks) > 0 {
			fmt.Fprintf(stdout, "  Framework  %s\n", strings.Join(plan.Stack.Frameworks, ", "))
		}
	}
	fmt.Fprintln(stdout)

	verb := cGreen
	verbWord := "wrote"
	if *dryRun {
		verb, verbWord = cYellow, "would write"
	}
	for _, f := range plan.Files {
		fmt.Fprintf(stdout, "  %s  %s\n", paint(verb, verbWord), f.Rel)
	}
	if len(plan.Files) == 0 {
		fmt.Fprintf(stdout, "  %s\n", paint(cDim, "nothing to write"))
	}
	for _, s := range plan.Skips {
		fmt.Fprintf(stdout, "  %s  %s — %s\n", paint(cDim, "skipped"), s.Rel, s.Reason)
	}
	if plan.NoRepo {
		fmt.Fprintf(stdout, "  %s  no git repository here, so no pre-commit hook\n", paint(cDim, "skipped"))
	}

	if *dryRun {
		fmt.Fprintf(stdout, "\n  %s\n\n", paint(cDim, "--dry-run: nothing was written. Re-run without it to apply."))
		return 0
	}

	written, err := plan.Apply()
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "\n  %d file(s) written.\n", len(written))

	if *noScan {
		fmt.Fprintf(stdout, "\n  Next: vibeshield scan %s\n\n", path)
		return 0
	}
	fmt.Fprintf(stdout, "\n  Running the first scan…\n")
	return run([]string{"scan", path}, stdout, stderr)
}

// cmdSearch is the scriptable half of the console: the same ranking, on
// stdout, so an AI agent can ask "what do you know about aws keys?" without a
// TTY. It searches the rule packs plus the action and agent-setup catalog.
func cmdSearch(args []string, stdout, stderr io.Writer) int {
	return runSearch(args, stdout, stderr, "search")
}

// cmdRules is `vibeshield rules [id]`: the rule packs, without the noise of the
// actions and agent recipes that `search` also indexes. It is the same engine
// with --rules --limit 0 pre-selected, which is why the two can never disagree.
func cmdRules(args []string, stdout, stderr io.Writer) int {
	pre := []string{"--rules", "--limit", "0"}
	// An explicit --limit in the user's args comes later, so Go's flag package
	// lets it win over the pre-selected one.
	return runSearch(append(pre, args...), stdout, stderr, "rules")
}

func runSearch(args []string, stdout, stderr io.Writer, helpCmd string) int {
	args = normalizeScanArgs(helpCmd, args)
	fl := newFlagSet(helpCmd, stderr)
	var (
		list       = fl.Bool("list", false, "list every catalog entry instead of searching")
		rulesOnly  = fl.Bool("rules", false, "search rule packs only")
		agentsOnly = fl.Bool("agents", false, "search agent setup recipes only")
		limit      = fl.Int("limit", 20, "maximum number of results")
		format     = fl.String("format", "pretty", "pretty | json")
		noColor    = fl.Bool("no-color", false, "disable color")
	)
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}
	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: core rules failed to load: %v\n", err)
		return 2
	}

	items := cli.Catalog(pack, Version)
	if *rulesOnly {
		items = filterItems(items, func(it cli.Item) bool { return strings.HasPrefix(it.Group, "Rule · ") })
	}
	if *agentsOnly {
		items = filterItems(items, func(it cli.Item) bool { return it.Group == "Agent setup" })
	}

	query := strings.Join(fl.Args(), " ")
	var results []cli.Item
	if *list || query == "" {
		results = items
	} else {
		results = cli.Rank(query, items)
	}
	// --list means "everything", so the default limit must not silently truncate
	// it. An explicit --limit still wins, which is what makes
	// `search --list --limit 5` useful.
	explicitLimit := false
	fl.Visit(func(f *flag.Flag) {
		if f.Name == "limit" {
			explicitLimit = true
		}
	})
	if *list && !explicitLimit {
		*limit = 0
	}
	if *limit > 0 && len(results) > *limit {
		results = results[:*limit]
	}

	switch *format {
	case "json":
		return emitSearchJSON(stdout, stderr, query, results)
	case "", "pretty":
		color := !*noColor && output.IsTTY(stdout)
		emitSearchPretty(stdout, query, results, color)
		if len(results) == 0 {
			// An empty search is not an error, but it should not be a dead end
			// either — point at the two ways to see everything.
			fmt.Fprintf(stdout, "  %s\n\n",
				"nothing matched — try `vibeshield rules` for the full list, or a shorter word")
		}
		return 0
	default:
		fmt.Fprintf(stderr, "vibeshield: unknown --format %q (pretty | json)\n", *format)
		return 2
	}
}

func filterItems(items []cli.Item, keep func(cli.Item) bool) []cli.Item {
	out := make([]cli.Item, 0, len(items))
	for _, it := range items {
		if keep(it) {
			out = append(out, it)
		}
	}
	return out
}

func emitSearchPretty(w io.Writer, query string, results []cli.Item, color bool) {
	dim := func(s string) string {
		if color {
			return "\x1b[90m" + s + "\x1b[0m"
		}
		return s
	}
	bold := func(s string) string {
		if color {
			return "\x1b[1m" + s + "\x1b[0m"
		}
		return s
	}
	cyan := func(s string) string {
		if color {
			return "\x1b[36m" + s + "\x1b[0m"
		}
		return s
	}

	if query != "" {
		fmt.Fprintf(w, "\n  %d result(s) for %q\n", len(results), query)
	} else {
		fmt.Fprintf(w, "\n  %d catalog entries\n", len(results))
	}
	last := ""
	for _, it := range results {
		if it.Group != last {
			fmt.Fprintf(w, "\n  %s\n", cyan(strings.ToUpper(it.Group)))
			last = it.Group
		}
		fmt.Fprintf(w, "    %s\n      %s\n", bold(it.Title), dim(it.Summary))
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n\n", dim("open the interactive console with `vibeshield` to read any entry in full"))
}

func emitSearchJSON(w io.Writer, stderr io.Writer, query string, results []cli.Item) int {
	type row struct {
		Kind     string   `json:"kind"`
		Group    string   `json:"group"`
		Title    string   `json:"title"`
		Summary  string   `json:"summary"`
		Keywords []string `json:"keywords,omitempty"`
		Command  []string `json:"command,omitempty"`
	}
	doc := struct {
		Schema  string `json:"schema"`
		Version string `json:"version"`
		Query   string `json:"query"`
		Count   int    `json:"count"`
		Results []row  `json:"results"`
	}{
		Schema:  "vibeshield.search/v1",
		Version: Version,
		Query:   query,
		Count:   len(results),
		Results: make([]row, 0, len(results)),
	}
	for _, it := range results {
		doc.Results = append(doc.Results, row{
			Kind: string(it.Kind), Group: it.Group, Title: it.Title,
			Summary: it.Summary, Keywords: it.Keywords, Command: it.Args,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	return 0
}

// cmdAgents prints the per-agent setup recipes. `vibeshield agents --body`
// emits just the shared rule block, which is what you paste into an agent's
// rules file; the full listing is the human-readable install guide.
func cmdAgents(args []string, stdout, stderr io.Writer) int {
	args = normalizeScanArgs("agents", args)
	fl := newFlagSet("agents", stderr)
	var (
		body     = fl.Bool("body", false, "print only the shared rule block")
		markdown = fl.Bool("markdown", false, "print the agent matrix as markdown")
	)
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}
	if *body {
		fmt.Fprint(stdout, cli.AgentRulesBody)
		return 0
	}

	ags := cli.Agents()
	if fl.NArg() > 0 {
		want := strings.ToLower(fl.Arg(0))
		kept := make([]cli.Agent, 0, 1)
		for _, a := range ags {
			if a.Slug == want || strings.Contains(strings.ToLower(a.Name), want) {
				kept = append(kept, a)
			}
		}
		if len(kept) == 0 {
			fmt.Fprintf(stderr, "vibeshield: no agent named %q. Try: ", fl.Arg(0))
			for i, a := range ags {
				if i > 0 {
					fmt.Fprint(stderr, ", ")
				}
				fmt.Fprint(stderr, a.Slug)
			}
			fmt.Fprintln(stderr)
			return 2
		}
		ags = kept
	}

	if *markdown {
		fmt.Fprintln(stdout, "| Agent | File | Scope |")
		fmt.Fprintln(stdout, "|:---|:---|:---|")
		for _, a := range ags {
			fmt.Fprintf(stdout, "| **%s** | `%s` | %s |\n", a.Name, a.File, a.Scope)
		}
		return 0
	}

	fmt.Fprintf(stdout, "\n  VibeShield %s — agent setup recipes\n", Version)
	for _, a := range ags {
		fmt.Fprintf(stdout, "\n  %s\n  %s\n", a.Name, strings.Repeat("─", len(a.Name)+4))
		fmt.Fprintf(stdout, "  File    %s\n", a.File)
		fmt.Fprintf(stdout, "  Scope   %s\n", a.Scope)
		for i, s := range a.Steps {
			fmt.Fprintf(stdout, "    %d. %s\n", i+1, s)
		}
	}
	fmt.Fprintf(stdout, "\n  Paste the rule block with: vibeshield agents --body\n\n")
	return 0
}

// normalizeScanArgs moves positional args (the scan path, the search query)
// after the flags: Go's flag package stops at the first non-flag word, but
// users (and the Action) call `scan . --format json` and
// `search "prompt injection" --limit 5`, not the other way round.
//
// The command name matters: `--rules` takes a directory in `scan` but is a
// boolean switch in `search`, so a single global "this flag takes a value"
// table would make `search --rules --list` swallow `--list` as a directory
// name. That was a real bug before v3.
func normalizeScanArgs(cmd string, args []string) []string {
	var flags, pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "-") && len(a) > 1:
			flags = append(flags, a)
			// `-` alone is the stdin diff spec, positional here is fine
			// to pass through as a flag value below.
			name := strings.TrimLeft(a, "-")
			if !strings.Contains(name, "=") && i+1 < len(args) && flagTakesValue(cmd, name) {
				i++
				flags = append(flags, args[i])
			}
		default:
			pos = append(pos, a)
		}
		i++
	}
	return append(flags, pos...)
}

// valueFlags lists, per subcommand, the flags that consume the following
// argument. Getting this wrong swallows a positional argument.
var valueFlags = map[string]map[string]bool{
	"scan":   {"diff": true, "format": true, "config": true, "mode": true, "rules": true, "max-cols": true},
	"fix":    {"report": true, "config": true, "rules": true},
	"init":   {"mode": true},
	"doctor": {"config": true, "format": true},
	"search": {"limit": true, "format": true},
	"rules":  {"limit": true, "format": true},
}

func flagTakesValue(cmd, name string) bool {
	return valueFlags[cmd][name]
}

// cmdFix is VibePatch: it scans (or reuses a --report file), plans the
// mechanical autofixes, shows a −/+ preview of exactly which files will be
// read and changed, and applies them through a human gate. --yes skips the
// prompts for coding agents; every applied patch is appended to
// vibeshield-fixes.log (JSONL) so the change trail stays auditable.
func cmdFix(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	args = normalizeScanArgs("fix", args)
	fl := newFlagSet("fix", stderr)
	var (
		dryRun   = fl.Bool("dry-run", false, "preview only, change nothing")
		yes      = fl.Bool("yes", false, "apply without prompting (agent/CI mode; still audited)")
		report   = fl.String("report", "", "reuse a --format json scan file instead of rescanning")
		cfgPath  = fl.String("config", "vibeshield.yml", "config file path")
		rulesDir = fl.String("rules", "", "directory of extra rule pack YAML files")
		noColor  = fl.Bool("no-color", false, "disable color")
		verbose  = fl.Bool("verbose", false, "explain what was scanned, to stderr")
	)
	fl.BoolVar(verbose, "v", false, "alias of --verbose")
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}
	path := "."
	if fl.NArg() > 0 {
		path = fl.Arg(0)
	}
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "vibeshield: %s is not a readable directory\n", path)
		return 2
	}

	cfgFile := resolveConfigPath(fl, *cfgPath, path)
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: core rules failed to load: %v\n", err)
		return 2
	}
	if *rulesDir != "" {
		extra, err := loadExtraPack(*rulesDir)
		if err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
		pack.Rules = append(pack.Rules, extra.Rules...)
	}
	if *verbose {
		fmt.Fprintf(stderr, "vibeshield: config %s · pack %s %s · %d rules\n",
			cfgFile, pack.ID, pack.Version, len(pack.Rules))
	}

	var rep *scan.Report
	if *report != "" {
		data, err := os.ReadFile(*report)
		if err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
		rep = &scan.Report{}
		if err := json.Unmarshal(data, rep); err != nil {
			fmt.Fprintf(stderr, "vibeshield: --report is not a vibeshield json scan: %v\n", err)
			return 2
		}
	} else {
		opts := scan.Options{Ignores: cfg.Ignores()}
		if len(cfg.Languages) > 0 {
			opts.Languages = cfg.Languages
		}
		fmt.Fprintf(stderr, "  Reading %s…\n", path)
		rep, err = scan.Dir(path, pack, Version, opts)
		if err != nil {
			fmt.Fprintf(stderr, "vibeshield: scan failed: %v\n", err)
			return 2
		}
		fmt.Fprintf(stderr, "  Read %d files · %d findings\n", rep.Scan.FilesScanned, len(rep.Findings))
	}

	plan := fix.BuildPlan(path, rep, pack)
	color := !*noColor && output.IsTTY(stdout)
	fix.Render(stdout, plan, color)
	if !plan.HasWork() {
		return 0
	}

	gate := fix.Gate{Mode: "ask", Stdin: stdin, Out: stdout}
	mode := "interactive"
	switch {
	case *dryRun:
		gate.Mode = "dry"
		fmt.Fprintln(stdout, "  (--dry-run: nothing was changed)")
		return 0
	case *yes:
		gate.Mode = "yes"
		mode = "yes"
	case !stdinIsInteractive(stdin):
		fmt.Fprintln(stderr, "vibeshield: stdin is not a terminal — refusing to guess at the gate. Re-run with --dry-run to preview or --yes to apply.")
		return 2
	}

	logPath := filepath.Join(path, "vibeshield-fixes.log")
	audit, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: cannot open audit log %s: %v\n", logPath, err)
		return 2
	}
	defer audit.Close()

	applied, err := fix.Apply(path, plan, &gate, audit, mode)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "\n  %d fix(es) applied · trail in %s — re-scan to confirm, review the diff before you commit.\n",
		applied, filepath.ToSlash(logPath))
	if applied == 0 {
		fmt.Fprintln(stdout, "  Nothing applied.")
	}
	return 0
}

// stdinIsInteractive reports whether fix's per-file gate can actually ask.
// A piped/scripted stdin would auto-answer EOF → "N" on every file, which
// looks like a hang to a user at a terminal — so non-TTY stdin must choose
// --dry-run or --yes explicitly.
func stdinIsInteractive(stdin io.Reader) bool {
	f, ok := stdin.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func cmdScan(args []string, stdout, stderr io.Writer) int {
	args = normalizeScanArgs("scan", args)
	fl := newFlagSet("scan", stderr)
	var (
		diffRef  = fl.String("diff", "", "scan changes vs git ref, or '-' for stdin")
		staged   = fl.Bool("staged", false, "scan git staged changes (pre-commit mode)")
		format   = fl.String("format", "", "pretty | json | github | sarif (default: pretty)")
		cfgPath  = fl.String("config", "vibeshield.yml", "config file path")
		mode     = fl.String("mode", "", "off | warn | block-on-critical | block-on-high+")
		rulesDir = fl.String("rules", "", "directory of extra rule pack YAML files")
		online   = fl.Bool("online", false, "(reserved) package-intel network lookups")
		maxCols  = fl.Int("max-cols", 88, "output width cap")
		noColor  = fl.Bool("no-color", false, "disable color")
		verbose  = fl.Bool("verbose", false, "explain what was scanned, to stderr")
	)
	fl.BoolVar(verbose, "v", false, "alias of --verbose")
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}
	if *online {
		fmt.Fprintln(stderr, "vibeshield: --online package-intel is not in this build yet; continuing fully offline (harmless)")
	}
	path := "."
	if fl.NArg() > 0 {
		path = fl.Arg(0)
	}
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "vibeshield: %s is not a readable directory\n", path)
		return 2
	}

	// Config
	cfgFile := resolveConfigPath(fl, *cfgPath, path)
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	effMode := cfg.Mode
	if effMode == "" {
		effMode = "warn"
	}
	if *mode != "" {
		effMode = *mode
	}
	if !config.AllowedModes[effMode] {
		fmt.Fprintf(stderr, "vibeshield: bad mode %q\n", effMode)
		return 2
	}
	if *mode != "" { // flag path must also be valid
		if !config.AllowedModes[*mode] {
			fmt.Fprintf(stderr, "vibeshield: bad --mode %q\n", *mode)
			return 2
		}
	}

	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: core rules failed to load: %v\n", err)
		return 2
	}
	if *rulesDir != "" {
		extra, err := loadExtraPack(*rulesDir)
		if err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
		seen := map[string]bool{}
		for _, r := range pack.Rules {
			seen[r.ID] = true
		}
		for _, r := range extra.Rules {
			if seen[r.ID] {
				fmt.Fprintf(stderr, "vibeshield: duplicate rule id %s in %s\n", r.ID, *rulesDir)
				return 2
			}
			pack.Rules = append(pack.Rules, r)
		}
	}

	opts := scan.Options{Ignores: cfg.Ignores()}
	if len(cfg.Languages) > 0 {
		opts.Languages = cfg.Languages
	}
	if *verbose {
		fmt.Fprintf(stderr, "vibeshield: config %s · mode %s\n", cfgFile, effMode)
		if len(cfg.Languages) > 0 {
			fmt.Fprintf(stderr, "vibeshield: languages %s\n", strings.Join(cfg.Languages, ", "))
		} else {
			fmt.Fprintln(stderr, "vibeshield: languages all supported")
		}
		fmt.Fprintf(stderr, "vibeshield: pack %s %s · %d rules · %d ignore(s)\n",
			pack.ID, pack.Version, len(pack.Rules), len(opts.Ignores))
	}

	// Full scan, then optionally filter to diff hunks.
	rep, err := scan.Dir(path, pack, Version, opts)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: scan failed: %v\n", err)
		return 2
	}
	if *verbose {
		fmt.Fprintf(stderr, "vibeshield: walked %s · %d files in %dms\n",
			filepath.ToSlash(path), rep.Scan.FilesScanned, rep.Scan.DurationMs)
	}
	if *staged || *diffRef != "" {
		var d *scandiff.Diff
		if *diffRef == "-" {
			data, _ := io.ReadAll(stdinOrPipe())
			d = scandiff.FromStdinText(string(data))
		} else if *staged {
			d, err = scandiff.FromGit("", "--staged")
		} else {
			d, err = scandiff.FromGit(*diffRef)
		}
		if err != nil {
			fmt.Fprintf(stderr, "vibeshield: diff mode unavailable: %v\n", err)
			return 2
		}
		rep.Scan.Mode = "diff"
		rep.Scan.Ref = d.Ref
		if *staged {
			rep.Scan.Ref = "staged"
		}
		filterToDiff(rep, d)
	}

	color := !*noColor && output.IsTTY(stdout)
	switch *format {
	case "json":
		if err := output.JSON(stdout, rep); err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
	case "github":
		if err := output.GitHub(stdout, rep); err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
		// The Action also consumes the JSON summary from --format json, so
		// annotations are the only stdout here; counts go to stderr for logs.
		fmt.Fprintf(stderr, "vibeshield: %d critical, %d high, %d medium, %d low\n",
			rep.Summary.Critical, rep.Summary.High, rep.Summary.Medium, rep.Summary.Low)
	case "sarif":
		// contracts/cli.md: "targets GitHub code-scanning uploads".
		if err := output.SARIF(stdout, rep); err != nil {
			fmt.Fprintf(stderr, "vibeshield: %v\n", err)
			return 2
		}
		fmt.Fprintf(stderr, "vibeshield: %d critical, %d high, %d medium, %d low\n",
			rep.Summary.Critical, rep.Summary.High, rep.Summary.Medium, rep.Summary.Low)
	case "", "pretty":
		output.Pretty(stdout, rep, color, *maxCols)
	default:
		fmt.Fprintf(stderr, "vibeshield: unknown --format %q (pretty | json | github | sarif)\n", *format)
		return 2
	}

	if rep.Summary.Blocks(effMode) {
		fmt.Fprintf(stderr, "vibeshield: findings at or above threshold — blocked (mode %s)\n", effMode)
		return 1
	}
	return 0
}

// filterToDiff keeps only findings on added lines (diff mode); file-level
// findings survive when their file was touched by the diff.
func filterToDiff(rep *scan.Report, d *scandiff.Diff) {
	kept := rep.Findings[:0]
	summary := scan.Summary{}
	clean := 0
	for _, f := range rep.Findings {
		inDiff := d.Has(f.File, f.Line)
		if f.Line == 0 {
			_, touched := d.Added[f.File]
			inDiff = touched
		}
		if inDiff {
			kept = append(kept, f)
			summary.Add(f.Severity)
		}
	}
	// clean files: of the files in the diff, those with no kept findings
	remaining := map[string]bool{}
	for _, f := range kept {
		remaining[f.File] = true
	}
	for _, file := range d.Files {
		if !remaining[file] {
			clean++
		}
	}
	rep.Findings = kept
	rep.Summary = summary
	rep.Summary.CleanFiles = clean
}
func loadExtraPack(dir string) (*rules.Pack, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	merged := &rules.Pack{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		p, errs := rules.ParsePackBytes(data, e.Name(), rules.LoadOptions{})
		if len(errs) > 0 {
			return nil, errs[0]
		}
		merged.Rules = append(merged.Rules, p.Rules...)
	}
	return merged, nil
}

// stdinOrPipe returns os.Stdin (placeholder for the '-' diff source).
func stdinOrPipe() io.Reader { return os.Stdin }
