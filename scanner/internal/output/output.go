// Package output renders a scan report in the contracts/cli.md formats:
// pretty (the VibeCheck terminal format), json, github (:error / :warning
// annotations for runner logs) and sarif (GitHub code-scanning uploads, see
// sarif.go).
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
)

// Severity display order and glyphs (contracts/cli.md pretty format).
var sevOrder = map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3, "info": 4}

type sevMeta struct {
	dot   string // colored dot for TTY
	emoji string // emoji kept in both modes
	name  string // token
	color string // ANSI, contracts/cli.md §Output format colors
}

var sevMetas = map[string]sevMeta{
	"critical": {"●", "🔴", "CRITICAL", "\x1b[31m"},
	"high":     {"●", "🟠", "HIGH", "\x1b[33m"},
	"medium":   {"●", "🟡", "MEDIUM", "\x1b[33m"},
	"low":      {"●", "🔵", "LOW", "\x1b[36m"},
	"info":     {"○", "⚪", "INFO", "\x1b[90m"},
}

// IsTTY reports whether output should be colored.
func IsTTY(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// JSON writes the contracts/cli.md document.
func JSON(w io.Writer, rep *scan.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

// GitHub writes ::error::/::warning:: workflow annotations (one per finding).
func GitHub(w io.Writer, rep *scan.Report) error {
	for _, f := range rep.Findings {
		lvl := "::warning"
		if f.Severity == finding.SeverityCritical || f.Severity == finding.SeverityHigh {
			lvl = "::error"
		}
		fmt.Fprintf(w, "%s file=%s,line=%d,col=%d,title=%s::%s (%s)\n",
			lvl, f.File, f.Line, f.Column, f.RuleID, oneLine(f.Message, 180), f.Title)
	}
	return nil
}

// Pretty writes the VibeCheck format (contracts/cli.md sample is law).
func Pretty(w io.Writer, rep *scan.Report, color bool, maxCols int) {
	if maxCols <= 0 {
		maxCols = 88
	}
	modeWord := "full"
	if rep.Scan.Mode == "diff" {
		modeWord = "diff mode"
	} else {
		modeWord = "full mode"
	}
	fmt.Fprintf(w, "\n  Scanning %d files (%s)… done in %.1fs\n\n",
		rep.Scan.FilesScanned, modeWord, float64(rep.Scan.DurationMs)/1000)

	fs := append([]*finding.Finding{}, rep.Findings...)
	sortFindings(fs)
	for _, f := range fs {
		m := sevMetas[f.Severity]
		head := fmt.Sprintf("  %s %-8s %-10s %s", m.emoji, strings.ToUpper(m.name), f.RuleID, f.Category)
		if color {
			head = m.color + head + "\x1b[0m"
		}
		fmt.Fprintln(w, head)
		for _, ln := range wrap(f.Message+": "+f.Snippet, maxCols-5) {
			fmt.Fprintf(w, "     %s\n", ln)
		}
		fix := "→ Fix: " + f.Fix
		for _, ln := range wrap(fix, maxCols-5) {
			if color {
				ln = "\x1b[32m" + ln + "\x1b[0m" // ok green
			}
			fmt.Fprintf(w, "     %s\n", ln)
		}
		fmt.Fprintln(w)
	}

	s := rep.Summary
	total := s.Critical + s.High + s.Medium + s.Low + s.Info
	check := "✓"
	if color {
		check = "\x1b[32m✓\x1b[0m"
	}
	fmt.Fprintf(w, "  %s %d files clean · %d findings", check, s.CleanFiles, total)
	deps := depWord(rep)
	if deps != "" {
		fmt.Fprintf(w, " · %s", deps)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)
}

func depWord(rep *scan.Report) string { return "" } // deps model lands with v1.1 (hallucinated-package scoring is rule-based today)

func sortFindings(fs []*finding.Finding) {
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && less(fs[j], fs[j-1]); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}

func less(a, b *finding.Finding) bool {
	if sa, sb := sevOrder[a.Severity], sevOrder[b.Severity]; sa != sb {
		return sa < sb
	}
	if a.File != b.File {
		return a.File < b.File
	}
	return a.Line < b.Line
}

func oneLine(s string, n int) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r", " "), "\n", " ")
	if len(s) > n {
		s = s[:n]
	}
	return s
}

// wrap splits text into lines of at most width runes, breaking on spaces.
func wrap(text string, width int) []string {
	var out []string
	var cur strings.Builder
	count := 0
	for _, word := range strings.Fields(text) {
		wl := utf8.RuneCountInString(word)
		if count > 0 && count+1+wl > width {
			out = append(out, cur.String())
			cur.Reset()
			count = 0
		}
		if count > 0 {
			cur.WriteString(" ")
			count++
		}
		cur.WriteString(word)
		count += wl
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	if len(out) == 0 {
		out = []string{text}
	}
	return out
}
