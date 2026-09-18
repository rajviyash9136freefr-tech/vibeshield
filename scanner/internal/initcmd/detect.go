// Package initcmd implements `vibeshield init`: detect what a project is
// built from, write a vibeshield.yml, a GitHub Action workflow and a
// pre-commit hook, then run a first scan.
//
// The command is documented in contracts/cli.md as
// "Detect frameworks, write vibeshield.yml + hook + workflow, run first scan".
// Everything here is offline static inspection — it reads manifests, never
// installs anything and never executes project code.
package initcmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Stack is what init could tell about a project by looking at it.
type Stack struct {
	// Languages are scanner language tokens (rules.AllowedLanguages), sorted.
	Languages []string
	// Manifests are the files that proved each language, sorted.
	Manifests []string
	// Frameworks are recognised app frameworks, for the printed summary.
	Frameworks []string
	// Ecosystems are package ecosystems found, e.g. npm, pip, go.
	Ecosystems []string
}

// Empty reports whether nothing recognisable was found.
func (s Stack) Empty() bool { return len(s.Languages) == 0 }

// manifest describes one marker file init looks for.
type manifest struct {
	file      string
	languages []string
	ecosystem string
	parseDeps bool // read dependencies out of it (package.json only)
}

// manifests is the detection table. Order only affects nothing — results are
// sorted before they are returned, so the generated config is deterministic.
var manifests = []manifest{
	{"package.json", []string{"javascript"}, "npm", true},
	{"tsconfig.json", []string{"typescript"}, "", false},
	{"requirements.txt", []string{"python"}, "pip", false},
	{"pyproject.toml", []string{"python"}, "pip", false},
	{"Pipfile", []string{"python"}, "pip", false},
	{"setup.py", []string{"python"}, "pip", false},
	{"go.mod", []string{"go"}, "go", false},
	{"Gemfile", []string{"ruby"}, "bundler", false},
	{"composer.json", []string{"php"}, "composer", false},
	{"pom.xml", []string{"java"}, "maven", false},
	{"build.gradle", []string{"java"}, "gradle", false},
	{"build.gradle.kts", []string{"java"}, "gradle", false},
	{"Cargo.toml", []string{"rust"}, "cargo", false},
}

// frameworkSignals maps a dependency name (lowercased) to the framework it
// implies. Only used for the human summary — nothing gates on it.
var frameworkSignals = map[string]string{
	"next": "Next.js", "react": "React", "vue": "Vue", "svelte": "Svelte",
	"@angular/core": "Angular", "express": "Express", "fastify": "Fastify",
	"@nestjs/core": "NestJS", "astro": "Astro", "nuxt": "Nuxt",
	"django": "Django", "flask": "Flask", "fastapi": "FastAPI",
	"rails": "Rails", "sinatra": "Sinatra",
	"gin-gonic/gin": "Gin", "labstack/echo": "Echo", "gofiber/fiber": "Fiber",
	"laravel/framework": "Laravel", "symfony/framework-bundle": "Symfony",
	"actix-web": "Actix", "rocket": "Rocket", "axum": "Axum",
}

// Detect inspects dir (top level only — a monorepo's nested packages are left
// alone so the generated config never claims languages the root does not have).
func Detect(dir string) (Stack, error) {
	var s Stack
	langs := map[string]bool{}
	ecos := map[string]bool{}
	seen := map[string]bool{}

	for _, m := range manifests {
		path := filepath.Join(dir, m.file)
		if fi, err := os.Stat(path); err != nil || fi.IsDir() {
			continue
		}
		s.Manifests = append(s.Manifests, m.file)
		for _, l := range m.languages {
			langs[l] = true
		}
		if m.ecosystem != "" {
			ecos[ecosystemName(m.ecosystem)] = true
		}
		if m.parseDeps {
			for _, fw := range frameworksFromPackageJSON(path) {
				if !seen[fw] {
					seen[fw] = true
					s.Frameworks = append(s.Frameworks, fw)
				}
			}
		}
	}

	// A .ts/.tsx file anywhere in the tree means the project really is
	// TypeScript, even without a tsconfig at the root.
	if !langs["typescript"] && hasAnyFile(dir, ".ts", ".tsx") {
		langs["typescript"] = true
	}
	// A Dockerfile is where several dependency-risk rules live.
	if fi, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil && !fi.IsDir() {
		langs["yaml"] = true
		s.Manifests = append(s.Manifests, "Dockerfile")
	}

	s.Languages = sortedKeys(langs)
	s.Ecosystems = sortedKeys(ecos)
	sort.Strings(s.Manifests)
	sort.Strings(s.Frameworks)
	return s, nil
}

// ecosystemName normalises the ecosystem label used in the summary.
func ecosystemName(raw string) string {
	switch raw {
	case "pip":
		return "python (pip)"
	case "bundler":
		return "ruby (bundler)"
	default:
		return raw
	}
}

// frameworksFromPackageJSON reads dependencies + devDependencies from a
// package.json and returns the frameworks it recognises. A malformed file is
// not an error: init should still be able to configure the project.
func frameworksFromPackageJSON(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var doc struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}
	var out []string
	for name := range doc.Dependencies {
		if fw, ok := frameworkSignals[strings.ToLower(name)]; ok {
			out = append(out, fw)
		}
	}
	for name := range doc.DevDependencies {
		if fw, ok := frameworkSignals[strings.ToLower(name)]; ok {
			out = append(out, fw)
		}
	}
	sort.Strings(out)
	return out
}

// hasAnyFile reports whether the tree under dir contains a file with one of
// the given extensions. It stops at the first hit and skips the directories
// that never hold first-party source.
func hasAnyFile(dir string, exts ...string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".git", "dist", "build", "vendor", ".venv", "venv", "__pycache__":
				return filepath.SkipDir
			}
			return nil
		}
		for _, ext := range exts {
			if strings.HasSuffix(d.Name(), ext) {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
