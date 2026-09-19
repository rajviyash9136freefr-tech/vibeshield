package initcmd

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// gitInit creates a minimal .git directory so hook installation has a target.
func gitInit(t *testing.T, dir string) string {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(filepath.Join(gitDir, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	return gitDir
}

func rels(files []File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Rel
	}
	return out
}

func TestBuildPlanWritesAllThreeFiles(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	write(t, dir, "package.json", `{"dependencies":{"express":"5"}}`)

	plan, err := BuildPlan(Options{Dir: dir, Mode: "warn", Version: "2.0.0", Hook: true, Workflow: true, BinaryPath: "/usr/local/bin/vibeshield"})
	if err != nil {
		t.Fatal(err)
	}
	got := rels(plan.Files)
	want := []string{ConfigFile, WorkflowFile, ".git/" + HookFile}
	if len(got) != len(want) {
		t.Fatalf("planned %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("file %d = %q, want %q", i, got[i], want[i])
		}
	}
	if len(plan.Skips) != 0 {
		t.Errorf("nothing should be skipped in a clean dir, got %+v", plan.Skips)
	}
}

func TestBuildPlanNeverClobbersExistingFiles(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	write(t, dir, ConfigFile, "mode: off\n")
	write(t, dir, WorkflowFile, "# my own workflow\n")
	write(t, dir, ".git/"+HookFile, "#!/bin/sh\necho mine\n")

	plan, err := BuildPlan(Options{Dir: dir, Mode: "warn", Hook: true, Workflow: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 0 {
		t.Fatalf("nothing may be overwritten without --force, planned %v", rels(plan.Files))
	}
	if len(plan.Skips) != 3 {
		t.Fatalf("expected 3 skips, got %+v", plan.Skips)
	}
	// The user's own hook must still be intact after a no-op plan.
	body, _ := os.ReadFile(filepath.Join(dir, ".git", HookFile))
	if !strings.Contains(string(body), "echo mine") {
		t.Error("an existing hook was modified by planning")
	}
}

func TestForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	write(t, dir, ConfigFile, "mode: off\n")

	plan, err := BuildPlan(Options{Dir: dir, Mode: "block-on-high+", Force: true, Hook: false, Workflow: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].Rel != ConfigFile {
		t.Fatalf("--force should overwrite the config, planned %v", rels(plan.Files))
	}
	if !strings.Contains(plan.Files[0].Content, "mode: block-on-high+") {
		t.Error("the overwriting config does not carry the requested mode")
	}
}

func TestPlanWithoutGitRepoSkipsHookOnly(t *testing.T) {
	dir := t.TempDir()
	plan, err := BuildPlan(Options{Dir: dir, Mode: "warn", Hook: true, Workflow: true})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.NoRepo {
		t.Error("NoRepo should be set when there is no git repository")
	}
	got := rels(plan.Files)
	if len(got) != 2 {
		t.Fatalf("config + workflow should still be planned, got %v", got)
	}
	for _, r := range got {
		if strings.HasPrefix(r, ".git/") {
			t.Errorf("no hook may be planned without a repository, got %q", r)
		}
	}
}

func TestFlagsCanSuppressFiles(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	plan, err := BuildPlan(Options{Dir: dir, Mode: "warn", Hook: false, Workflow: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 1 || plan.Files[0].Rel != ConfigFile {
		t.Fatalf("--no-hook --no-workflow should leave only the config, got %v", rels(plan.Files))
	}
}

func TestApplyWritesFilesWithModes(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	plan, err := BuildPlan(Options{Dir: dir, Mode: "warn", Version: "2.0.0", Hook: true, Workflow: true, BinaryPath: "/opt/vibeshield"})
	if err != nil {
		t.Fatal(err)
	}
	written, err := plan.Apply()
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 3 {
		t.Fatalf("wrote %d files, want 3", len(written))
	}
	for _, f := range written {
		if _, err := os.Stat(f.Path); err != nil {
			t.Errorf("%s was not written: %v", f.Rel, err)
		}
	}
	fi, err := os.Stat(filepath.Join(dir, ".git", HookFile))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		// Windows has no POSIX permission bits: os.Chmod only toggles the
		// read-only flag, and git for Windows runs hooks via its bundled sh
		// regardless of mode. The exec bit is asserted on the Unix builds in CI.
		t.Log("skipping the executable-bit assertion on windows")
		return
	}
	if fi.Mode().Perm()&0o111 == 0 {
		t.Errorf("the hook is not executable (mode %v) — git would ignore it", fi.Mode().Perm())
	}
}

func TestBuildPlanRejectsMissingDir(t *testing.T) {
	if _, err := BuildPlan(Options{Dir: filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Error("a missing directory should be an error")
	}
}

func TestRenderedConfigIsLoadableShape(t *testing.T) {
	stack := Stack{Languages: []string{"javascript", "typescript"}, Manifests: []string{"package.json"}}
	out := renderConfig(stack, "block-on-critical")

	for _, want := range []string{"mode: block-on-critical", "languages:", "  - javascript", "  - typescript", "ignore: []", "new_dependency_max_age_days: 30"} {
		if !strings.Contains(out, want) {
			t.Errorf("config missing %q\n%s", want, out)
		}
	}
	// The ignore block must stay commented — an active ignore with no reason
	// would be a contract violation (contracts/cli.md requires a reason).
	if strings.Contains(out, "\n  - rule:") {
		t.Error("the ignore example must remain commented out")
	}
}

func TestRenderedConfigWithNoDetectedStack(t *testing.T) {
	out := renderConfig(Stack{}, "warn")
	if strings.Contains(out, "\nlanguages:") {
		t.Error("with nothing detected the languages key should stay commented")
	}
	if !strings.Contains(out, "mode: warn") {
		t.Error("mode must always be written")
	}
}

func TestRenderedWorkflowPinsTheActionTag(t *testing.T) {
	out := renderWorkflow("warn", "2.0.0")
	for _, want := range []string{"action@v2.0.0", "mode: warn", "pull_request:", "fetch-depth: 0", "pull-requests: write"} {
		if !strings.Contains(out, want) {
			t.Errorf("workflow missing %q\n%s", want, out)
		}
	}
}

// The fallback tag is a release constant, not a literal sprinkled through the
// function — a stale one generates a workflow that 404s.
func TestDefaultActionTagIsSemver(t *testing.T) {
	if !regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(defaultActionTag) {
		t.Errorf("defaultActionTag = %q, want a vX.Y.Z tag", defaultActionTag)
	}
	if got := actionRef("dev"); got != defaultActionTag {
		t.Errorf("a non-semver version should fall back to %q, got %q", defaultActionTag, got)
	}
}

func TestActionRefRejectsNonSemver(t *testing.T) {
	cases := map[string]string{
		"2.0.1":     "v2.0.1",
		"v2.0.1":    "v2.0.1",
		"2.1.10":    "v2.1.10",
		"dev":       "v2.0.1",
		"":          "v2.0.1",
		"2.0":       "v2.0.1",
		"2.0.0-rc1": "v2.0.1",
		"main":      "v2.0.1",
	}
	for in, want := range cases {
		if got := actionRef(in); got != want {
			t.Errorf("actionRef(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderedHookPrefersPathThenFallsBack(t *testing.T) {
	out := renderHook("/opt/vibe shield/vibeshield")
	if !strings.HasPrefix(out, "#!/bin/sh\n") {
		t.Error("the hook must start with a shebang")
	}
	if !strings.Contains(out, "command -v vibeshield") {
		t.Error("the hook should prefer vibeshield from PATH")
	}
	if !strings.Contains(out, "'/opt/vibe shield/vibeshield' scan --staged") {
		t.Errorf("a path with a space must be quoted\n%s", out)
	}
}

func TestRenderedHookNormalisesWindowsPaths(t *testing.T) {
	// The hook is run by sh, where backslashes are not path separators — a
	// C:\Users\... path would be read literally and never resolve.
	out := renderHook(`C:\Users\dev\bin\vibeshield.exe`)
	if strings.Contains(out, `\`) {
		t.Errorf("the hook must not contain backslashes\n%s", out)
	}
	if !strings.Contains(out, "'C:/Users/dev/bin/vibeshield.exe' scan --staged") {
		t.Errorf("expected a forward-slash absolute path\n%s", out)
	}
}

func TestRenderedHookWithoutBinaryExitsZero(t *testing.T) {
	out := renderHook("")
	if !strings.Contains(out, "exit 0") {
		t.Error("without a binary the hook must not block commits")
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"/usr/bin/vibeshield": "'/usr/bin/vibeshield'",
		"/opt/my tools/vb":    "'/opt/my tools/vb'",
		"/it's/here":          `'/it'\''s/here'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFindGitDirWalksUp(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := findGitDir(nested)
	if !ok {
		t.Fatal("expected to find the repository root from a nested directory")
	}
	if filepath.Base(got) != ".git" {
		t.Errorf("found %q, want a .git directory", got)
	}
}

func TestFindGitDirHandlesWorktreeFile(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real-git-dir")
	if err := os.MkdirAll(filepath.Join(real, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, "worktree")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, work, ".git", "gitdir: "+real+"\n")

	got, ok := findGitDir(work)
	if !ok {
		t.Fatal("a .git file with a gitdir pointer should resolve")
	}
	if got != real {
		t.Errorf("resolved to %q, want %q", got, real)
	}
}

func TestFindGitDirAbsent(t *testing.T) {
	if _, ok := findGitDir(t.TempDir()); ok {
		t.Error("a plain temp dir should not resolve to a repository")
	}
}
