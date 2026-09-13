// Package fix implements VibePatch: mechanical, same-line rewrites of
// findings whose rule carries an autofix (contracts/rulepack.md). It plans
// patches from a scan report and a loaded pack, renders a diff preview, and
// applies them only through an explicit human gate (per-file y/N) or an
// explicit --yes flag for agent/CI use. It never invents code: an autofix
// replaces text inside the matched line and must not add or remove lines.
//
// Privacy: like the rest of the scanner, this package never touches the
// network. Everything it writes comes from the rule pack and your own files.
package fix

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
)

// LinePatch is one line rewrite inside one file.
type LinePatch struct {
	LineNo int    // 1-based
	RuleID string // finding that motivated it
	Fix    string // rule's one-line suggestion (context for the human)
	Before string
	After  string
}

// FilePatch is the set of rewrites planned for one file.
type FilePatch struct {
	Path     string   // repo-relative POSIX
	Lines    []string // original content split on \n
	EndNL    bool     // original ended with a newline
	Patches  []LinePatch
	newLines []string // computed on first Diff
}

// Plan is the full fix plan across files (deterministic order).
type Plan struct {
	Files     []*FilePatch
	Findings  int // autofix-eligible findings seen
	Unfixable int // eligible-by-rule but line didn't match (stale scan)
}

// unsafeCategories never receive mechanical patches regardless of rule data:
// a secret needs rotation, not a rewrite, and a prompt-injection finding must
// not be "fixed" by an agent editing its own instruction file.
var unsafeCategories = map[string]bool{
	finding.CategoryHardcodedSecret: true,
	finding.CategoryPromptInjection: true,
}

// BuildPlan maps findings to their rule's autofix and computes line
// rewrites. Findings whose line no longer matches (or whose rule has no
// autofix) are counted in Unfixable, not applied.
func BuildPlan(root string, rep *scan.Report, pack *rules.Pack) *Plan {
	byID := map[string]*rules.Rule{}
	for i := range pack.Rules {
		r := &pack.Rules[i]
		byID[r.ID] = r
	}
	plan := &Plan{}
	perFile := map[string]*FilePatch{}
	for _, f := range rep.Findings {
		r := byID[f.RuleID]
		if r == nil || r.Autofix == nil || unsafeCategories[f.Category] {
			continue
		}
		if f.Line < 1 {
			continue
		}
		plan.Findings++
		fp := perFile[f.File]
		if fp == nil {
			fp = readFile(root, f.File)
			if fp == nil {
				plan.Unfixable++
				continue
			}
			perFile[f.File] = fp
		}
		if f.Line > len(fp.Lines) {
			plan.Unfixable++
			continue
		}
		line := fp.Lines[f.Line-1]
		re := r.Autofix.Re()
		if re == nil || !re.MatchString(line) {
			plan.Unfixable++
			continue
		}
		out := re.ReplaceAllString(line, r.Autofix.Replace)
		if out == line || strings.ContainsAny(out, "\n\r") {
			plan.Unfixable++
			continue
		}
		fp.Patches = append(fp.Patches, LinePatch{
			LineNo: f.Line, RuleID: r.ID, Fix: r.Fix, Before: line, After: out,
		})
	}
	files := make([]*FilePatch, 0, len(perFile))
	for _, fp := range perFile {
		if len(fp.Patches) > 0 {
			files = append(files, fp)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	plan.Files = files
	return plan
}

// readFile loads a file for patching; nil when unreadable.
func readFile(root, rel string) *FilePatch {
	p := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	s := string(data)
	return &FilePatch{
		Path:  rel,
		Lines: strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n"),
		EndNL: strings.HasSuffix(s, "\n"),
	}
}

// newContent returns the patched line set (computed once).
func (fp *FilePatch) newContent() []string {
	if fp.newLines != nil {
		return fp.newLines
	}
	out := append([]string{}, fp.Lines...)
	for _, p := range fp.Patches {
		out[p.LineNo-1] = p.After
	}
	fp.newLines = out
	return out
}

// Bytes renders the patched file content preserving CRLF when the original
// used it.
func (fp *FilePatch) Bytes() []byte {
	lines := fp.newContent()
	crlf := false
	if len(fp.Lines) > 0 {
		// Cheap heuristic: if any line ends with \r the original was CRLF.
		for _, l := range fp.Lines {
			if strings.HasSuffix(l, "\r") {
				crlf = true
				break
			}
		}
	}
	if crlf {
		for i := range lines {
			lines[i] = strings.TrimSuffix(lines[i], "\r") + "\r"
		}
	}
	s := strings.Join(lines, "\n")
	if !fp.EndNL && strings.HasSuffix(s, "\n") {
		s = strings.TrimSuffix(s, "\n")
	}
	return []byte(s)
}

// HasWork reports whether the plan would change anything.
func (p *Plan) HasWork() bool { return len(p.Files) > 0 }

// Apply writes one file atomically (temp + rename).
func (fp *FilePatch) Apply(root string) error {
	p := filepath.Join(root, filepath.FromSlash(fp.Path))
	tmp := p + ".vibeshield-tmp"
	if err := os.WriteFile(tmp, fp.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// AuditEntry is one line of vibeshield-fixes.log (JSONL) — the trail that
// makes agent-driven --yes runs reviewable after the fact.
type AuditEntry struct {
	Time   string `json:"time"`
	RuleID string `json:"rule_id"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Before string `json:"before"`
	After  string `json:"after"`
	Mode   string `json:"mode"` // interactive | yes
}

// Render writes the human-readable preview: every file that was read, and
// for each pending patch a −/+ line pair with the rule and its one-line fix.
func Render(w io.Writer, p *Plan, color bool) {
	if !p.HasWork() {
		fmt.Fprintln(w, "  Nothing to patch — no finding carries a mechanical autofix.")
		if p.Unfixable > 0 {
			fmt.Fprintf(w, "  (%d findings were stale or their line did not match)\n", p.Unfixable)
		}
		return
	}
	fmt.Fprintf(w, "\n  %d file(s) to change · %d patch(es)\n\n", len(p.Files), len(p.allPatches()))
	for _, fp := range p.Files {
		head := fmt.Sprintf("  %s  (%d fixes)", fp.Path, len(fp.Patches))
		if color {
			head = "\x1b[36m" + head + "\x1b[0m"
		}
		fmt.Fprintln(w, head)
		for _, pt := range fp.Patches {
			minus, plus := "−", "+"
			if color {
				minus = "\x1b[31m−\x1b[0m"
				plus = "\x1b[32m+\x1b[0m"
			}
			fmt.Fprintf(w, "     line %d · %s\n", pt.LineNo, pt.RuleID)
			fmt.Fprintf(w, "     %s %s\n", minus, pt.Before)
			fmt.Fprintf(w, "     %s %s\n", plus, pt.After)
			if pt.Fix != "" {
				fmt.Fprintf(w, "       → %s\n", pt.Fix)
			}
		}
		fmt.Fprintln(w)
	}
}

func (p *Plan) allPatches() []LinePatch {
	var out []LinePatch
	for _, fp := range p.Files {
		out = append(out, fp.Patches...)
	}
	return out
}

// Gate decides per-file application. Interactive prompts read from stdin;
// mode "yes" applies everything, mode "dry" applies nothing, mode "ask"
// prompts [y/N] per file.
type Gate struct {
	Mode  string // dry | ask | yes
	Stdin io.Reader
	Out   io.Writer
	// SkipAll becomes true after the user chooses s(kip all).
	SkipAll bool
}

// confirm asks once per file. Returns (apply, err). EOF counts as "N".
func (g *Gate) confirm(fp *FilePatch) bool {
	if g.SkipAll {
		return false
	}
	fmt.Fprintf(g.Out, "  Apply %d fix(es) to %s? [y/N/a=ll/s=kip all] ", len(fp.Patches), fp.Path)
	var ans string
	if _, err := fmt.Fscanln(g.Stdin, &ans); err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "a", "all":
		g.Mode = "yes"
		return true
	case "s", "skip", "skipall":
		g.SkipAll = true
		return false
	case "y", "yes":
		return true
	default:
		return false
	}
}

// Apply executes the plan through the gate, writing audit entries (JSONL)
// to audit when non-nil. Returns the number of patches applied.
func Apply(root string, p *Plan, g *Gate, audit io.Writer, mode string) (int, error) {
	applied := 0
	for _, fp := range p.Files {
		proceed := g.Mode == "yes" || (g.Mode == "ask" && g.confirm(fp))
		if !proceed {
			continue
		}
		if err := fp.Apply(root); err != nil {
			return applied, fmt.Errorf("writing %s: %w", fp.Path, err)
		}
		for _, pt := range fp.Patches {
			applied++
			if audit != nil {
				_ = json.NewEncoder(audit).Encode(AuditEntry{
					Time: time.Now().UTC().Format(time.RFC3339), RuleID: pt.RuleID,
					File: fp.Path, Line: pt.LineNo, Before: pt.Before, After: pt.After, Mode: mode,
				})
			}
		}
		fmt.Fprintf(g.Out, "  ✓ %s — %d fix(es) applied\n", fp.Path, len(fp.Patches))
	}
	return applied, nil
}
