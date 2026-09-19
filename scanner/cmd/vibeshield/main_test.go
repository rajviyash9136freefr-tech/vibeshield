package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// newFlagSet builds a FlagSet with the config flag and parses args, so that
// fl.Visit can tell "the user passed --config" from "the default applied".
func newFlagSet(t *testing.T, args ...string) *flag.FlagSet {
	t.Helper()
	fl := flag.NewFlagSet("test", flag.ContinueOnError)
	fl.SetOutput(io.Discard)
	fl.String("config", "vibeshield.yml", "config file path")
	fl.String("mode", "", "mode")
	fl.Bool("verbose", false, "verbose")
	if err := fl.Parse(args); err != nil {
		t.Fatal(err)
	}
	return fl
}

func TestResolveConfigPathPrefersTheScannedProject(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "vibeshield.yml")
	if err := os.WriteFile(project, []byte("mode: warn\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Scanning another project from somewhere else must read THAT project's
	// config. Resolving from the working directory silently ignored it.
	got := resolveConfigPath(newFlagSet(t), "vibeshield.yml", dir)
	if got != project {
		t.Errorf("resolved %q, want the scanned project's %q", got, project)
	}
}

func TestResolveConfigPathFallsBackToTheDefault(t *testing.T) {
	dir := t.TempDir() // no vibeshield.yml here
	got := resolveConfigPath(newFlagSet(t), "vibeshield.yml", dir)
	if got != "vibeshield.yml" {
		t.Errorf("resolved %q, want the default %q", got, "vibeshield.yml")
	}
}

func TestResolveConfigPathHonoursExplicitConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vibeshield.yml"), []byte("mode: warn\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// An explicit --config wins even when the scanned project has its own.
	got := resolveConfigPath(newFlagSet(t, "--config", "custom.yml"), "custom.yml", dir)
	if got != "custom.yml" {
		t.Errorf("resolved %q, want the explicit %q", got, "custom.yml")
	}
}

func TestResolveConfigPathIgnoresADirectory(t *testing.T) {
	dir := t.TempDir()
	// A directory named vibeshield.yml is not a config file.
	if err := os.MkdirAll(filepath.Join(dir, "vibeshield.yml"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := resolveConfigPath(newFlagSet(t), "vibeshield.yml", dir)
	if got != "vibeshield.yml" {
		t.Errorf("resolved %q, want the default when the candidate is a directory", got)
	}
}

func TestNormalizeScanArgsMovesFlagsFirst(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"already ordered", []string{"--format", "json", "."}, []string{"--format", "json", "."}},
		{"path first", []string{".", "--format", "json"}, []string{"--format", "json", "."}},
		{"bool after path", []string{".", "--staged"}, []string{"--staged", "."}},
		{"equals form is not split", []string{".", "--format=json"}, []string{"--format=json", "."}},
		{"search query with a trailing flag", []string{"aws", "--limit", "5"}, []string{"--limit", "5", "aws"}},
		{"multi-word query survives", []string{"prompt", "injection", "--limit", "5"}, []string{"--limit", "5", "prompt", "injection"}},
		{"no args", nil, nil},
	}
	for _, tc := range cases {
		got := normalizeScanArgs(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: normalizeScanArgs(%v) = %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

// A lone "-" is the stdin diff spec, not a flag, and must survive reordering
// as a positional rather than being swallowed as a flag value.
func TestNormalizeScanArgsKeepsStdinSpec(t *testing.T) {
	got := normalizeScanArgs([]string{"--diff", "-", "--format", "json"})
	want := []string{"--diff", "-", "--format", "json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFlagTakesValueCoversEveryValueFlag(t *testing.T) {
	// Getting this wrong swallows the following positional argument, which is
	// how `search "prompt injection" --limit 5` once treated "--limit 5" as
	// part of the query.
	for _, name := range []string{"diff", "format", "config", "mode", "rules", "max-cols", "report", "limit"} {
		if !flagTakesValue(name) {
			t.Errorf("flagTakesValue(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"staged", "dry-run", "yes", "no-color", "verbose", "online", "force"} {
		if flagTakesValue(name) {
			t.Errorf("flagTakesValue(%q) = true, want false — it takes no value", name)
		}
	}
}
