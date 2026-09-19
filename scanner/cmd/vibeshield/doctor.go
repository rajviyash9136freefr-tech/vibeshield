// `vibeshield doctor` — the CLI's self-check.
//
// Every CLI that people run in more than one environment eventually grows one
// of these (brew doctor, flutter doctor, gh auth status) because the same three
// questions keep coming up: is the binary the one I think it is, is the config
// being read, and is the gate actually wired into git and CI? Answering them
// from the tool itself is cheaper than answering them in issues.
//
// doctor never modifies anything. It reports, prints the command that fixes
// each gap, and exits 1 when something needs fixing so CI can assert the setup.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/config"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/initcmd"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/output"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

// checkStatus is how a check turned out. "note" is informational and never
// fails the command; only "warn" does.
type checkStatus string

const (
	statusOK   checkStatus = "ok"
	statusWarn checkStatus = "warn"
	statusNote checkStatus = "note"
	statusInfo checkStatus = "info"
)

// check is one line of the doctor report.
type check struct {
	Name   string      `json:"name"`
	Status checkStatus `json:"status"`
	Detail string      `json:"detail"`
	Fix    string      `json:"fix,omitempty"`
}

// agentRuleFiles are the per-agent rule files VibeShield ships. doctor only
// reports which are present; the full recipes are `vibeshield agents`.
var agentRuleFiles = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".cursor/rules/vibeshield.mdc",
	".cursorrules",
	".windsurf/rules/vibeshield.md",
	".windsurfrules",
	".agents/rules/vibeshield.md",
	".github/copilot-instructions.md",
}

func cmdDoctor(args []string, stdout, stderr io.Writer) int {
	args = normalizeScanArgs("doctor", args)
	fl := newFlagSet("doctor", stderr)
	var (
		cfgPath = fl.String("config", "vibeshield.yml", "config file path to check")
		format  = fl.String("format", "pretty", "pretty | json")
		noColor = fl.Bool("no-color", false, "disable color")
		verbose = fl.Bool("verbose", false, "print every check, including the ones that passed")
	)
	fl.BoolVar(verbose, "v", false, "alias of --verbose")
	fl.Usage = func() { printUsage(stderr, "doctor") }
	if err := fl.Parse(args); err != nil {
		return usageExit(err)
	}

	path := "."
	if fl.NArg() > 0 {
		path = fl.Arg(0)
	}

	checks, next := runChecks(path, *cfgPath, *verbose)

	if *format == "json" {
		return emitDoctorJSON(stdout, stderr, path, checks)
	}
	if *format != "" && *format != "pretty" {
		fmt.Fprintf(stderr, "vibeshield: unknown --format %q (pretty | json)\n", *format)
		return 2
	}

	renderDoctor(stdout, checks, next, !*noColor && output.IsTTY(stdout))
	return doctorExit(checks)
}

// doctorExit is 1 when anything needs fixing, 0 otherwise. Documented in
// `vibeshield help doctor` and contracts/cli.md.
func doctorExit(checks []check) int {
	for _, c := range checks {
		if c.Status == statusWarn {
			return 1
		}
	}
	return 0
}

func runChecks(path, cfgFlag string, verbose bool) ([]check, string) {
	var checks []check
	next := ""

	// 1. Binary + rule pack. This is the one check that cannot be skipped:
	// if the embedded pack does not load, nothing else matters.
	pack, err := rules.LoadCore()
	if err != nil {
		checks = append(checks, check{
			Name: "binary", Status: statusWarn,
			Detail: fmt.Sprintf("vibeshield %s — embedded rule pack failed to load: %v", Version, err),
			Fix:    "reinstall vibeshield (the embedded pack should never fail)",
		})
	} else {
		detail := fmt.Sprintf("vibeshield %s · pack %s %s · %d active rules",
			Version, pack.ID, pack.Version, pack.ActiveRules())
		if r := pack.ReservedRules(); r > 0 {
			detail += fmt.Sprintf(" (+%d reserved)", r)
		}
		checks = append(checks, check{Name: "binary", Status: statusOK, Detail: detail})
	}

	// 2. Project path.
	root, absErr := filepath.Abs(path)
	fi, statErr := os.Stat(path)
	switch {
	case absErr != nil:
		checks = append(checks, check{Name: "project", Status: statusWarn,
			Detail: absErr.Error(), Fix: "pass a readable directory: vibeshield doctor <path>"})
		return checks, "vibeshield doctor <path>"
	case statErr != nil || !fi.IsDir():
		checks = append(checks, check{Name: "project", Status: statusWarn,
			Detail: fmt.Sprintf("%s is not a readable directory", path),
			Fix:    "pass a readable directory: vibeshield doctor <path>"})
		return checks, "vibeshield doctor <path>"
	}
	checks = append(checks, check{Name: "project", Status: statusOK, Detail: filepath.ToSlash(root)})

	// 3. Config. Resolution mirrors `scan`: an explicit --config wins, then the
	// scanned project's own vibeshield.yml, then the working directory's.
	explicitCfg := cfgFlag != "vibeshield.yml"
	cfgFile, cfgExists := resolveDoctorConfig(cfgFlag, root, explicitCfg)
	switch {
	case explicitCfg && !cfgExists:
		// Naming a file that is not there is a mistake worth surfacing: the
		// scanner would fall back to defaults and the user would never know.
		checks = append(checks, check{
			Name: "config", Status: statusWarn,
			Detail: fmt.Sprintf("--config %s does not exist", cfgFlag),
			Fix:    "point --config at a real file, or drop the flag for the defaults",
		})
		next = "vibeshield doctor"
	case !cfgExists:
		checks = append(checks, check{
			Name: "config", Status: statusNote,
			Detail: "no vibeshield.yml — scanner defaults are in use (mode warn)",
			Fix:    "vibeshield init",
		})
		next = "vibeshield init"
	default:
		cfg, err := config.Load(cfgFile)
		if err != nil {
			checks = append(checks, check{
				Name: "config", Status: statusWarn,
				Detail: fmt.Sprintf("%s is not valid: %v", relTo(root, cfgFile), err),
				Fix:    "fix the file, then re-run: vibeshield doctor",
			})
			next = "vibeshield doctor"
		} else {
			mode := cfg.Mode
			if mode == "" {
				mode = "warn (default)"
			}
			detail := fmt.Sprintf("%s · mode %s", relTo(root, cfgFile), mode)
			if len(cfg.Languages) > 0 {
				detail += " · languages " + strings.Join(cfg.Languages, ",")
			}
			checks = append(checks, check{Name: "config", Status: statusOK, Detail: detail})
		}
	}

	// 4. Git repository + pre-commit hook.
	gitDir, isRepo := initcmd.FindGitDir(root)
	if !isRepo {
		checks = append(checks, check{
			Name: "git", Status: statusNote,
			Detail: "not a git repository — no pre-commit hook, and --diff/--staged will not work",
			Fix:    "git init",
		})
		if next == "" {
			next = "vibeshield scan ."
		}
	} else {
		checks = append(checks, check{Name: "git", Status: statusOK, Detail: "repository detected"})
		hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
		data, err := os.ReadFile(hookPath)
		switch {
		case err != nil:
			checks = append(checks, check{
				Name: "git hook", Status: statusWarn,
				Detail: "no .git/hooks/pre-commit — commits are not scanned",
				Fix:    "vibeshield init",
			})
			next = "vibeshield init"
		case strings.Contains(string(data), "vibeshield"):
			checks = append(checks, check{Name: "git hook", Status: statusOK,
				Detail: ".git/hooks/pre-commit runs vibeshield"})
		default:
			checks = append(checks, check{
				Name: "git hook", Status: statusWarn,
				Detail: "a pre-commit hook exists but does not run vibeshield (it was left alone, by design)",
				Fix:    "add `vibeshield scan --staged` to it, or run: vibeshield init --force",
			})
			next = "vibeshield init --force"
		}
	}

	// 5. Pull-request gate.
	wfPath := filepath.Join(root, filepath.FromSlash(initcmd.WorkflowFile))
	if fi, err := os.Stat(wfPath); err == nil && !fi.IsDir() {
		checks = append(checks, check{Name: "pr gate", Status: statusOK,
			Detail: initcmd.WorkflowFile + " present"})
	} else {
		checks = append(checks, check{
			Name: "pr gate", Status: statusWarn,
			Detail: initcmd.WorkflowFile + " is missing — pull requests are not gated",
			Fix:    "vibeshield init",
		})
		if next == "" {
			next = "vibeshield init"
		}
	}

	// 6. Agent rule files — informational: a project may legitimately use none.
	var found []string
	for _, f := range agentRuleFiles {
		if fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(f))); err == nil && !fi.IsDir() {
			found = append(found, f)
		}
	}
	if len(found) > 0 {
		checks = append(checks, check{Name: "agent rules", Status: statusOK,
			Detail: fmt.Sprintf("%d of %d present: %s",
				len(found), len(agentRuleFiles), strings.Join(found, ", "))})
	} else {
		checks = append(checks, check{
			Name: "agent rules", Status: statusNote,
			Detail: "no agent rule file — your coding agent is not being told what to avoid",
			Fix:    "vibeshield agents --body > AGENTS.md",
		})
		if next == "" {
			next = "vibeshield agents --body > AGENTS.md"
		}
	}

	if verbose {
		checks = append(checks, check{Name: "terminal", Status: statusInfo,
			Detail: fmt.Sprintf("stdout is a TTY: %t", output.IsTTY(os.Stdout))})
	}
	return checks, next
}

// resolveDoctorConfig mirrors scan's resolution order: an explicit --config
// wins, then the scanned project's own vibeshield.yml, then the working
// directory's. It reports whether the file it settled on actually exists, so
// doctor can tell "no config, defaults apply" from "you named a file that is
// not there".
func resolveDoctorConfig(cfgFlag, root string, explicit bool) (string, bool) {
	if explicit {
		_, err := os.Stat(cfgFlag)
		return cfgFlag, err == nil
	}
	candidate := filepath.Join(root, cfgFlag)
	if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
		return candidate, true
	}
	if fi, err := os.Stat(cfgFlag); err == nil && !fi.IsDir() {
		return cfgFlag, true
	}
	return candidate, false
}

func relTo(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(p)
}

func renderDoctor(w io.Writer, checks []check, next string, color bool) {
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
		cCyan   = "\x1b[36m"
	)

	fmt.Fprintf(w, "\n  %s\n\n", paint(cBold, "VibeShield "+Version+" — doctor"))

	warns, notes := 0, 0
	for _, c := range checks {
		label, colour := "ok", cGreen
		switch c.Status {
		case statusWarn:
			label, colour = "warn", cYellow
			warns++
		case statusNote:
			label, colour = "note", cCyan
			notes++
		case statusInfo:
			label, colour = "info", cDim
		}
		fmt.Fprintf(w, "  %-4s %-12s %s\n", paint(colour, label), c.Name, c.Detail)
		if c.Fix != "" {
			fmt.Fprintf(w, "  %-4s %-12s %s %s\n", "", "", paint(cDim, "→"), c.Fix)
		}
	}

	fmt.Fprintln(w)
	switch {
	case warns == 0 && notes == 0:
		fmt.Fprintf(w, "  %s\n", paint(cGreen, "Everything is wired up. Nothing to fix."))
	case warns == 0:
		fmt.Fprintf(w, "  %s\n", paint(cGreen, fmt.Sprintf("No problems found (%d note(s)).", notes)))
	default:
		fmt.Fprintf(w, "  %s\n", paint(cYellow, fmt.Sprintf("%d thing(s) to fix.", warns)))
	}
	if next != "" {
		fmt.Fprintf(w, "  %s %s\n", paint(cDim, "Next:"), next)
	}
	fmt.Fprintln(w)
}

func emitDoctorJSON(w io.Writer, stderr io.Writer, path string, checks []check) int {
	type doc struct {
		Schema  string  `json:"schema"`
		Version string  `json:"version"`
		Path    string  `json:"path"`
		OK      bool    `json:"ok"`
		Checks  []check `json:"checks"`
	}
	d := doc{
		Schema:  "vibeshield.doctor/v1",
		Version: Version,
		Path:    filepath.ToSlash(path),
		OK:      doctorExit(checks) == 0,
		Checks:  checks,
	}
	if d.Checks == nil {
		d.Checks = []check{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		fmt.Fprintf(stderr, "vibeshield: %v\n", err)
		return 2
	}
	return doctorExit(checks)
}
