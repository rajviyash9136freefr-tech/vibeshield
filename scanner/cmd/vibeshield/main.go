// Command vibeshield is the VibeShield CLI: a static security & dependency
// auditor for AI-generated code (contracts/cli.md). It never sends code
// anywhere; --online opts into package-intel lookups (v1.1).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/config"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/output"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scandiff"
)

// Version is stamped by -ldflags "-X main.Version=v1.2.3" at release build.
var Version = "1.0.0"

const usage = `vibeshield %s — security scanner for AI-generated code

Usage:
  vibeshield scan [path]      Scan a directory (full) or a git diff (--diff/--staged)
  vibeshield version          Print version and embedded rule-pack info

Scan flags:
  --diff <ref|->     Diff mode: scan changes vs a git ref, or "-" for stdin
  --staged           Pre-commit mode: scan git staged changes
  --format <fmt>     pretty (default) | json | github        (sarif: v1.1)
  --config <file>    Config path (default: vibeshield.yml if present)
  --mode <mode>      off | warn | block-on-critical | block-on-high+
  --rules <dir>      Load extra rule packs from a directory
  --online           (reserved) allow package-intel network lookups
  --max-cols <n>     Output width cap (default 88)
  --no-color         Disable color (also: NO_COLOR env, non-TTY auto)

Exit codes: 0 clean or warn-mode findings · 1 block threshold met · 2 config/usage error
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, usage, Version)
		return 2
	}
	switch args[0] {
	case "version", "--version", "-v":
		return cmdVersion(stdout)
	case "scan":
		return cmdScan(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		fmt.Fprintf(stdout, usage, Version)
		return 0
	default:
		fmt.Fprintf(stderr, "vibeshield: unknown command %q\n\n", args[0])
		fmt.Fprintf(stderr, usage, Version)
		return 2
	}
}

func cmdVersion(w io.Writer) int {
	pack, err := rules.LoadCore()
	if err != nil {
		fmt.Fprintf(w, "vibeshield %s — rule packs failed to load: %v\n", Version, err)
		return 2
	}
	fmt.Fprintf(w, "vibeshield %s\n", Version)
	fmt.Fprintf(w, "rule packs: %s %s (%s, %d rules)\n", pack.ID, pack.Version, pack.License, len(pack.Rules))
	fmt.Fprintf(w, "no code leaves this machine: static analysis only\n")
	return 0
}

// normalizeScanArgs moves positional args (the scan path) after the flags:
// Go's flag package stops at the first non-flag word, but users (and the
// Action) call `scan . --format json`, not `scan --format json .`.
func normalizeScanArgs(args []string) []string {
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
			if !strings.Contains(name, "=") && i+1 < len(args) && flagTakesValue(name) {
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

func flagTakesValue(name string) bool {
	switch name {
	case "diff", "format", "config", "mode", "rules", "max-cols":
		return true
	}
	return false
}

func cmdScan(args []string, stdout, stderr io.Writer) int {
	args = normalizeScanArgs(args)
	fl := flag.NewFlagSet("scan", flag.ContinueOnError)
	fl.SetOutput(stderr)
	var (
		diffRef  = fl.String("diff", "", "scan changes vs git ref, or '-' for stdin")
		staged   = fl.Bool("staged", false, "scan git staged changes (pre-commit mode)")
		format   = fl.String("format", "", "pretty | json | github (default: pretty)")
		cfgPath  = fl.String("config", "vibeshield.yml", "config file path")
		mode     = fl.String("mode", "", "off | warn | block-on-critical | block-on-high+")
		rulesDir = fl.String("rules", "", "directory of extra rule pack YAML files")
		online   = fl.Bool("online", false, "(reserved) package-intel network lookups")
		maxCols  = fl.Int("max-cols", 88, "output width cap")
		noColor  = fl.Bool("no-color", false, "disable color")
	)
	fl.Usage = func() { fmt.Fprintf(stderr, usage, Version) }
	if err := fl.Parse(args); err != nil {
		return 2
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
	cfg, err := config.Load(*cfgPath)
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

	// Full scan, then optionally filter to diff hunks.
	rep, err := scan.Dir(path, pack, Version, opts)
	if err != nil {
		fmt.Fprintf(stderr, "vibeshield: scan failed: %v\n", err)
		return 2
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
		fmt.Fprintln(stderr, "vibeshield: --format sarif lands in v1.1 (see contracts/cli.md)")
		return 2
	case "", "pretty":
		output.Pretty(stdout, rep, color, *maxCols)
	default:
		fmt.Fprintf(stderr, "vibeshield: unknown --format %q\n", *format)
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
