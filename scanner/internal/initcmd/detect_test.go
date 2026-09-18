package initcmd

import (
	"os"
	"path/filepath"
	"testing"
)

// write creates a file (and its parents) inside a temp dir.
func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestDetectEmptyDir(t *testing.T) {
	stack, err := Detect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !stack.Empty() {
		t.Errorf("empty dir detected %v, want nothing", stack.Languages)
	}
	if len(stack.Manifests) != 0 || len(stack.Frameworks) != 0 {
		t.Errorf("empty dir produced manifests/frameworks: %+v", stack)
	}
}

func TestDetectNodeProject(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"dependencies":{"next":"15.0.0","react":"19.0.0"}}`)
	write(t, dir, "src/app.ts", "export const x = 1;")

	stack, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"javascript", "typescript"} {
		if !contains(stack.Languages, want) {
			t.Errorf("missing language %q in %v", want, stack.Languages)
		}
	}
	if !contains(stack.Ecosystems, "npm") {
		t.Errorf("missing npm ecosystem in %v", stack.Ecosystems)
	}
	for _, fw := range []string{"Next.js", "React"} {
		if !contains(stack.Frameworks, fw) {
			t.Errorf("missing framework %q in %v", fw, stack.Frameworks)
		}
	}
}

func TestDetectTypeScriptWithoutTsconfig(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"dependencies":{}}`)
	write(t, dir, "lib/deep/thing.tsx", "export default () => null;")

	stack, _ := Detect(dir)
	if !contains(stack.Languages, "typescript") {
		t.Errorf("a .tsx in the tree should imply typescript, got %v", stack.Languages)
	}
}

func TestDetectSkipsNodeModulesForTypeScriptProbe(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"dependencies":{}}`)
	write(t, dir, "node_modules/some-dep/index.ts", "export const x = 1;")

	stack, _ := Detect(dir)
	if contains(stack.Languages, "typescript") {
		t.Error("typescript must not be inferred from a dependency inside node_modules")
	}
}

func TestDetectMultiLanguageProject(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"dependencies":{}}`)
	write(t, dir, "requirements.txt", "flask\n")
	write(t, dir, "go.mod", "module example.com/x\n")
	write(t, dir, "Gemfile", "gem 'rails'\n")
	write(t, dir, "Cargo.toml", "[package]\nname = \"x\"\n")
	write(t, dir, "Dockerfile", "FROM scratch\n")

	stack, _ := Detect(dir)
	for _, want := range []string{"javascript", "python", "go", "ruby", "rust", "yaml"} {
		if !contains(stack.Languages, want) {
			t.Errorf("missing language %q in %v", want, stack.Languages)
		}
	}
	if !contains(stack.Manifests, "Dockerfile") {
		t.Errorf("Dockerfile should be listed as a manifest, got %v", stack.Manifests)
	}
	if contains(stack.Languages, "typescript") {
		t.Error("no tsconfig or .ts file, typescript should not be inferred")
	}
}

func TestDetectIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"dependencies":{"react":"19"}}`)
	write(t, dir, "requirements.txt", "flask\n")

	first, _ := Detect(dir)
	for i := 0; i < 20; i++ {
		got, _ := Detect(dir)
		if len(got.Languages) != len(first.Languages) {
			t.Fatalf("run %d produced a different language count", i)
		}
		for j := range first.Languages {
			if got.Languages[j] != first.Languages[j] {
				t.Fatalf("run %d produced unsorted/unstable languages: %v vs %v", i, got.Languages, first.Languages)
			}
		}
	}
}

func TestDetectToleratesMalformedPackageJSON(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", "{ this is not json")

	stack, err := Detect(dir)
	if err != nil {
		t.Fatalf("a malformed manifest must not fail detection: %v", err)
	}
	if !contains(stack.Languages, "javascript") {
		t.Error("package.json should still imply javascript")
	}
	if len(stack.Frameworks) != 0 {
		t.Errorf("no frameworks can be known from invalid JSON, got %v", stack.Frameworks)
	}
}

func TestDetectIgnoresDirectoriesNamedLikeManifests(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "package.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	stack, _ := Detect(dir)
	if !stack.Empty() {
		t.Errorf("a directory named package.json is not a manifest, got %v", stack.Languages)
	}
}
