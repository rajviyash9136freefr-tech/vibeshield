package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFile, WorkflowFile and HookFile are the paths init writes, relative to
// the project root (the hook is relative to the git dir).
const (
	ConfigFile   = "vibeshield.yml"
	WorkflowFile = ".github/workflows/vibeshield.yml"
	HookFile     = "hooks/pre-commit"
)

// Options tune what init writes.
type Options struct {
	Dir        string // project root
	Mode       string // initial gate mode
	Version    string // this binary's version, used to pin the Action tag
	Force      bool   // overwrite files that already exist
	Hook       bool   // install the pre-commit hook
	Workflow   bool   // write the GitHub Action workflow
	BinaryPath string // absolute path to this binary, baked into the hook
}

// File is one file init will write.
type File struct {
	Path    string
	Rel     string
	Content string
	Mode    os.FileMode
}

// Skip is a file init deliberately did not write, and why.
type Skip struct {
	Rel    string
	Reason string
}

// Plan is everything init intends to do, computed before anything touches disk
// so --dry-run and the real run can never disagree.
type Plan struct {
	Stack   Stack
	Root    string
	GitDir  string
	Files   []File
	Skips   []Skip
	NoRepo  bool
	Scanned bool
}

// BuildPlan detects the stack and decides which files to write.
func BuildPlan(opts Options) (*Plan, error) {
	root, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(root)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a readable directory", opts.Dir)
	}

	stack, err := Detect(root)
	if err != nil {
		return nil, err
	}
	p := &Plan{Stack: stack, Root: root}

	// vibeshield.yml
	cfgPath := filepath.Join(root, ConfigFile)
	if exists(cfgPath) && !opts.Force {
		p.Skips = append(p.Skips, Skip{ConfigFile, "already exists (use --force to overwrite)"})
	} else {
		p.Files = append(p.Files, File{
			Path: cfgPath, Rel: ConfigFile,
			Content: renderConfig(stack, opts.Mode), Mode: 0o644,
		})
	}

	// GitHub Action workflow
	if opts.Workflow {
		wfPath := filepath.Join(root, filepath.FromSlash(WorkflowFile))
		if exists(wfPath) && !opts.Force {
			p.Skips = append(p.Skips, Skip{WorkflowFile, "already exists (use --force to overwrite)"})
		} else {
			p.Files = append(p.Files, File{
				Path: wfPath, Rel: WorkflowFile,
				Content: renderWorkflow(opts.Mode, opts.Version), Mode: 0o644,
			})
		}
	}

	// Pre-commit hook. This is the one file that can destroy work if it
	// clobbers an existing hook, so it is opt-in and never overwritten
	// without --force.
	if opts.Hook {
		gitDir, ok := findGitDir(root)
		if !ok {
			p.NoRepo = true
		} else {
			p.GitDir = gitDir
			hookPath := filepath.Join(gitDir, filepath.FromSlash(HookFile))
			if exists(hookPath) && !opts.Force {
				p.Skips = append(p.Skips, Skip{".git/" + HookFile, "already exists — not touching your hook (use --force to overwrite)"})
			} else {
				p.Files = append(p.Files, File{
					Path: hookPath, Rel: ".git/" + HookFile,
					Content: renderHook(opts.BinaryPath), Mode: 0o755,
				})
			}
		}
	}
	return p, nil
}

// Apply writes the plan. It returns the files written, in order.
func (p *Plan) Apply() ([]File, error) {
	var written []File
	for _, f := range p.Files {
		if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(f.Path, []byte(f.Content), f.Mode); err != nil {
			return written, err
		}
		// WriteFile honours umask, so re-assert the mode the hook needs.
		if f.Mode&0o111 != 0 {
			if err := os.Chmod(f.Path, f.Mode); err != nil {
				return written, err
			}
		}
		written = append(written, f)
	}
	return written, nil
}

func exists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// findGitDir walks up from dir looking for a .git directory or a .git file
// (worktrees and submodules use the file form with a "gitdir:" pointer).
func findGitDir(dir string) (string, bool) {
	cur := dir
	for {
		dotGit := filepath.Join(cur, ".git")
		if fi, err := os.Stat(dotGit); err == nil {
			if fi.IsDir() {
				return dotGit, true
			}
			data, err := os.ReadFile(dotGit)
			if err == nil {
				line := strings.TrimSpace(string(data))
				if target, ok := strings.CutPrefix(line, "gitdir:"); ok {
					target = strings.TrimSpace(target)
					if !filepath.IsAbs(target) {
						target = filepath.Join(cur, target)
					}
					if fi, err := os.Stat(target); err == nil && fi.IsDir() {
						return target, true
					}
				}
			}
			return "", false
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false
		}
		cur = parent
	}
}

// --- renderers --------------------------------------------------------------

// renderConfig writes a vibeshield.yml that only sets what init actually
// knows. Everything else is commented so a reader can see the options without
// the file claiming values nobody chose.
func renderConfig(s Stack, mode string) string {
	if mode == "" {
		mode = "warn"
	}
	var b strings.Builder
	b.WriteString("# VibeShield configuration — written by `vibeshield init`.\n")
	b.WriteString("# Schema: https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/contracts/cli.md\n")
	b.WriteString("# Every key is optional; delete one to get the scanner default back.\n\n")

	fmt.Fprintf(&b, "# Gate policy: off | warn | block-on-critical | block-on-high+\n")
	fmt.Fprintf(&b, "# Start on warn, then tighten once your first scans come back clean.\n")
	fmt.Fprintf(&b, "mode: %s\n\n", mode)

	if !s.Empty() {
		b.WriteString("# Languages the scanner looks at. Detected from this project:\n")
		if len(s.Manifests) > 0 {
			fmt.Fprintf(&b, "#   %s\n", strings.Join(s.Manifests, ", "))
		}
		b.WriteString("# Trim the list to narrow the scan; an empty list means every language.\n")
		b.WriteString("languages:\n")
		for _, l := range s.Languages {
			fmt.Fprintf(&b, "  - %s\n", l)
		}
		b.WriteString("\n")
	} else {
		b.WriteString("# No manifest was recognised at the root, so every language is scanned.\n")
		b.WriteString("# Uncomment and trim to narrow it:\n")
		b.WriteString("# languages: [javascript, typescript, python, go]\n\n")
	}

	b.WriteString("# Auditable ignores. Every entry needs a reason — it is echoed in reports.\n")
	b.WriteString("ignore: []\n")
	b.WriteString("#  - rule: VS-SEC-014\n")
	b.WriteString("#    paths: [\"tests/**\"]\n")
	b.WriteString("#    reason: \"intentional insecure fixture\"\n\n")

	b.WriteString("thresholds:\n")
	b.WriteString("  # A dependency younger than this is treated as a slopsquatting risk.\n")
	b.WriteString("  new_dependency_max_age_days: 30\n")
	return b.String()
}

// renderWorkflow writes the pull-request gate. It pins the Action to the
// release this binary came from rather than a moving major tag, so a scan
// cannot change behaviour without a commit. A local build reports a version
// like "2.0.0" or "dev"; only a semver-looking one is turned into a tag.
func renderWorkflow(mode, version string) string {
	if mode == "" {
		mode = "warn"
	}
	return fmt.Sprintf(`# VibeShield pull-request gate — written by `+"`vibeshield init`"+`.
# Scans the diff on every PR and posts one consolidated VibeCheck report.
# Docs: https://github.com/rajviyash9136freefr-tech/vibeshield/tree/main/action
name: VibeShield

on:
  pull_request:

permissions:
  contents: read
  pull-requests: write

jobs:
  vibeshield:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          # Diff mode needs the base branch, so a shallow clone is not enough.
          fetch-depth: 0
      - uses: rajviyash9136freefr-tech/vibeshield/action@%s
        with:
          # off | warn | block-on-critical | block-on-high+
          mode: %s
`, actionRef(version), mode)
}

// defaultActionTag is the Action ref used when the running binary's version is
// not a plain semver (a local `go build` reports "dev"). It is a release
// constant, so scripts/bump-version.mjs keeps it current.
const defaultActionTag = "v2.0.1"

// actionRef turns a binary version into a usable Action ref. Anything that is
// not a plain semver is replaced by defaultActionTag, so a locally built
// binary still generates a workflow that resolves.
func actionRef(version string) string {
	v := strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return defaultActionTag
	}
	for _, p := range parts {
		if p == "" {
			return defaultActionTag
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return defaultActionTag
			}
		}
	}
	return "v" + v
}

// renderHook writes a POSIX pre-commit hook. It prefers vibeshield from PATH
// and falls back to the absolute path of the binary that ran init, so the hook
// keeps working whether or not the user has it on PATH.
//
// The path is normalised to forward slashes: the hook runs under sh, where a
// Windows path like C:\Users\... would be read literally (backslashes are not
// separators), while C:/Users/... resolves in Git Bash, MSYS and every POSIX
// shell. The replacement is explicit rather than filepath.ToSlash, which is a
// no-op on Unix — there the separator is already "/", so backslashes survive
// and the behaviour would differ by platform.
func renderHook(binaryPath string) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString("# VibeShield pre-commit hook — installed by `vibeshield init`.\n")
	b.WriteString("# Scans staged changes only, fully offline. Bypass with: git commit --no-verify\n")
	b.WriteString("if command -v vibeshield >/dev/null 2>&1; then\n")
	b.WriteString("  exec vibeshield scan --staged\n")
	b.WriteString("fi\n")
	if binaryPath != "" {
		posix := strings.ReplaceAll(binaryPath, `\`, "/")
		fmt.Fprintf(&b, "exec %s scan --staged\n", shellQuote(posix))
	} else {
		b.WriteString("echo 'vibeshield: not found on PATH — install it or edit this hook' >&2\n")
		b.WriteString("exit 0\n")
	}
	return b.String()
}

// shellQuote single-quotes a path for /bin/sh, escaping embedded quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
