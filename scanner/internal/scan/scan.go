// Package scan walks a project tree and applies a loaded rule pack,
// emitting contracts/finding/schema.json findings. It is the full-tree
// engine behind `vibeshield scan`; diff mode is wired separately.
//
// Privacy (PRD §6 decision 5): this package never opens a network
// connection. Everything it knows comes from the embedded rule packs.
package scan

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

// MaxFileBytes bounds the size of a single scannable file (bigger files are
// skipped — minified bundles are not where AI findings live).
const MaxFileBytes = 1 << 20 // 1 MiB

// skipDirs are never descended into.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
	"build": true, "target": true, ".venv": true, "venv": true,
	"__pycache__": true, ".astro": true, ".next": true, ".cache": true,
	"coverage": true, ".idea": true, ".vscode": true, ".vite": true,
}

// extLangs maps file extension → rule language (contracts/rulepack.md enum).
var extLangs = map[string]string{
	".js": "javascript", ".jsx": "javascript", ".mjs": "javascript", ".cjs": "javascript",
	".ts": "typescript", ".tsx": "typescript",
	".py":   "python",
	".go":   "go",
	".java": "java",
	".rb":   "ruby",
	".php":  "php",
	".rs":   "rust",
	".cs":   "csharp",
	".yaml": "yaml", ".yml": "yaml",
}

// LangForFile reports the rule language for a path ("generic" when unknown;
// unknown extensions still get generic-category rules).
func LangForFile(path string) string {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".env") {
		return "generic"
	}
	if l, ok := extLangs[strings.ToLower(filepath.Ext(base))]; ok {
		return l
	}
	return "generic"
}

// Ignore is one vibeshield.yml ignore entry (rule + optional path globs).
type Ignore struct {
	Rule  string
	Globs []string
}

func (ig Ignore) matches(ruleID, rel string) bool {
	if ig.Rule != "" && ig.Rule != ruleID {
		return false
	}
	if len(ig.Globs) == 0 {
		return true
	}
	for _, g := range ig.Globs {
		if ok, _ := MatchGlob(g, rel); ok {
			return true
		}
	}
	return false
}

// MatchGlob compiles a * / ? / ** glob (the vibeshield.yml ignore dialect)
// and matches it against a slash-separated relative path.
func MatchGlob(g, path string) (bool, error) {
	var sb strings.Builder
	sb.WriteString("^")
	for i := 0; i < len(g); i++ {
		switch c := g[i]; c {
		case '*':
			if i+1 < len(g) && g[i+1] == '*' {
				i++
				if i+1 < len(g) && g[i+1] == '/' {
					i++
					sb.WriteString("(?:[^/]+/)*")
				} else {
					sb.WriteString(".*")
				}
			} else {
				sb.WriteString("[^/]*")
			}
		case '?':
			sb.WriteString("[^/]")
		default:
			sb.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	sb.WriteString("$")
	re, err := regexp.Compile(sb.String())
	if err != nil {
		return false, err
	}
	return re.MatchString(path), nil
}

// Options configure one scan.
type Options struct {
	Languages []string // optional filter; empty = all
	Ignores   []Ignore // auditable ignores from config
	MaxFiles  int      // safety cap (0 = default 20000)
}

// Summary counts by severity plus clean-file count.
type Summary struct {
	Critical   int `json:"critical"`
	High       int `json:"high"`
	Medium     int `json:"medium"`
	Low        int `json:"low"`
	Info       int `json:"info"`
	CleanFiles int `json:"clean_files"`
}

// Add folds one finding's severity into the counts.
func (s *Summary) Add(sev string) {
	switch sev {
	case finding.SeverityCritical:
		s.Critical++
	case finding.SeverityHigh:
		s.High++
	case finding.SeverityMedium:
		s.Medium++
	case finding.SeverityLow:
		s.Low++
	default:
		s.Info++
	}
}

// Blocks reports whether the summary triggers the exit-1 gate for a mode
// (contracts/cli.md exit-code law).
func (s Summary) Blocks(mode string) bool {
	switch mode {
	case "block-on-critical":
		return s.Critical > 0
	case "block-on-high+":
		return s.Critical+s.High > 0
	default: // off, warn
		return false
	}
}

// ScanInfo describes the scan run (contracts/cli.md JSON envelope).
type ScanInfo struct {
	Mode         string `json:"mode"`
	Ref          string `json:"ref,omitempty"`
	FilesScanned int    `json:"files_scanned"`
	DurationMs   int64  `json:"duration_ms"`
}

// Report is the contracts/cli.md JSON output document.
type Report struct {
	SchemaVersion int                `json:"schema_version"`
	Tool          string             `json:"tool"`
	Version       string             `json:"version"`
	Scan          ScanInfo           `json:"scan"`
	Summary       Summary            `json:"summary"`
	Findings      []*finding.Finding `json:"findings"`
}

// errTooManyFiles aborts the walk once the safety cap is hit.
var errTooManyFiles = errors.New("scan: file cap reached")

// Dir walks root and scans every non-skipped file against the pack.
func Dir(root string, pack *rules.Pack, version string, opts Options) (*Report, error) {
	if pack == nil {
		return nil, errors.New("scan: nil rule pack")
	}
	st := time.Now()
	rep := &Report{
		SchemaVersion: 1,
		Tool:          "vibeshield",
		Version:       version,
		Scan:          ScanInfo{Mode: "full"},
		Findings:      []*finding.Finding{},
	}
	maxFiles := opts.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 20000
	}

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entries are skipped, not fatal
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if rep.Scan.FilesScanned >= maxFiles {
			return errTooManyFiles
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		info, serr := d.Info()
		if serr != nil || info.Size() > MaxFileBytes {
			return nil
		}
		lang := LangForFile(path)
		if len(opts.Languages) > 0 && !langIn(opts.Languages, lang) {
			return nil
		}
		data, oerr := os.ReadFile(path)
		if oerr != nil || strings.ContainsRune(string(data), 0) {
			return nil // binary or unreadable
		}
		rep.Scan.FilesScanned++
		before := len(rep.Findings)
		scanFile(rel, string(data), lang, pack, opts, rep)
		if len(rep.Findings) == before {
			rep.Summary.CleanFiles++
		}
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, errTooManyFiles) {
		return nil, walkErr
	}
	rep.Scan.DurationMs = time.Since(st).Milliseconds()
	return rep, nil
}

func langIn(languages []string, lang string) bool {
	for _, l := range languages {
		if l == lang || l == "generic" {
			return true
		}
	}
	return false
}

// scanFile applies every path/language-applicable rule to the file's lines.
// The matched region of secret findings is redacted here — rule authors
// never redact (contracts/rulepack.md engine semantics).
func scanFile(rel, content, lang string, pack *rules.Pack, opts Options, rep *Report) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i := range pack.Rules {
		r := &pack.Rules[i]
		if !r.AppliesToLanguage(lang) || !r.MatchesPath(rel) {
			continue
		}
		re := r.Pattern.Re()
		if re == nil {
			continue
		}
		for ln, line := range lines {
			loc := re.FindStringSubmatchIndex(line)
			if loc == nil {
				continue
			}
			if ignored(opts.Ignores, r.ID, rel) {
				continue
			}
			f := finding.New(r.ID, r.Severity, r.Category, r.Title, r.Message, rel)
			f.Line = ln + 1
			f.Column = loc[0] + 1
			f.Fix = r.Fix
			f.Confidence = r.Confidence
			// Redact on the raw line (offsets are line-relative), then trim.
			// Redaction covers the whole match — the match is the dangerous
			// span. (Capture-group refinement for key="value" shapes lands
			// with the v1.1 secret ruleset upgrade.)
			snip := line
			if r.Category == finding.CategoryHardcodedSecret {
				snip = finding.RedactRegion(snip, loc[0], loc[1])
			}
			snip = strings.TrimSpace(snip)
			if len(snip) > 200 {
				snip = snip[:200] + "…"
			}
			f.Snippet = snip
			f.Finalize()
			rep.Findings = append(rep.Findings, f)
			rep.Summary.Add(r.Severity)
		}
	}
}

func ignored(ignores []Ignore, ruleID, rel string) bool {
	for _, ig := range ignores {
		if ig.matches(ruleID, rel) {
			return true
		}
	}
	return false
}
