// Package scandiff implements diff mode: scan only the lines a change
// touched. It shells out to git (no network), parses unified diffs with
// -U0 context, and maps file → set of added line numbers.
package scandiff

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Diff is a parsed unified diff: per file, the added line numbers.
type Diff struct {
	Ref     string // ref scanned against, or "staged"/"stdin"
	Added   map[string]map[int]bool
	Files   []string // stable order
}

var (
	hunkRe  = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)
	fileRe  = regexp.MustCompile(`^\+\+\+ b/(.+)$`)
	devRe   = regexp.MustCompile(`^\+\+\+ /dev/null`)
)

// FromGit runs `git diff` for the given spec args and parses the output.
func FromGit(ref string, extraArgs ...string) (*Diff, error) {
	args := append([]string{"diff", "--unified=0", "--no-color"}, extraArgs...)
	if ref != "" {
		args = append(args, ref)
	}
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	d := newDiff(ref)
	d.parse(string(out))
	return d, nil
}

// FromStdinText parses a unified diff read from stdin (ref recorded as "stdin").
func FromStdinText(text string) *Diff {
	d := newDiff("stdin")
	d.parse(text)
	return d
}

func newDiff(ref string) *Diff {
	return &Diff{Ref: ref, Added: map[string]map[int]bool{}}
}

func (d *Diff) parse(text string) {
	var cur string
	var newLine int
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "+++ "):
			cur = ""
			if devRe.MatchString(line) {
				// new file: all added lines are counted by hunk headers anyway
			} else if m := fileRe.FindStringSubmatch(line); m != nil {
				cur = m[1]
				if cur == "/dev/null" {
					cur = ""
				}
			}
			newLine = 0
		case strings.HasPrefix(line, "@@"):
			m := hunkRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			n, _ := strconv.Atoi(m[1])
			newLine = n
		case strings.HasPrefix(line, "+") && cur != "":
			if newLine > 0 {
				d.mark(cur, newLine)
			}
			newLine++
		case strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\\"):
			newLine++
		}
	}
}

func (d *Diff) mark(file string, line int) {
	if d.Added[file] == nil {
		d.Added[file] = map[int]bool{}
		d.Files = append(d.Files, file)
	}
	d.Added[file][line] = true
}

// Has reports whether file:line is an added line.
func (d *Diff) Has(file string, line int) bool {
	return d.Added[file][line]
}

// InRepo reports the git working tree root (for resolving diff paths).
func InRepo() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}
