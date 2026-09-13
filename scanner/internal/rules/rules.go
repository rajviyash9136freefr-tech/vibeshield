// Package rules implements the VibeShield rule-pack engine per
// contracts/rulepack.md: strict YAML validation, Go RE2 regexes only
// (lookaheads / backreferences / possessive quantifiers are rejected at
// load), doublestar include/exclude globs, and line matching.
package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

// SchemaID is the only pack schema this scanner accepts.
const SchemaID = "vibeshield.rules/v1"

var ruleIDRe = regexp.MustCompile(`^VS-(PKG|SEC|LIC|DEP|INJ)-\d{3}$`)

// AllowedLanguages is contracts/rulepack.md field law.
var AllowedLanguages = map[string]bool{
	"javascript": true, "typescript": true, "python": true, "go": true,
	"java": true, "ruby": true, "php": true, "rust": true,
	"csharp": true, "yaml": true, "generic": true,
}

var allowedSeverities = map[string]bool{
	"critical": true, "high": true, "medium": true, "low": true, "info": true,
}

var allowedCategories = map[string]bool{
	"hallucinated-package": true, "hardcoded-secret": true, "insecure-api": true,
	"license-missing": true, "insecure-default": true, "dependency-risk": true,
	"prompt-injection": true,
}

var allowedFlags = map[string]bool{"multiline": true, "caseless": true, "dotall": true}

// Pattern is the rule's matcher.
type Pattern struct {
	Kind  string   `yaml:"kind"`
	Match string   `yaml:"match"`
	Flags []string `yaml:"flags,omitempty"`

	re *regexp.Regexp // compiled form (regex kind)
	// captureGroup is the submatch index holding the secret value for
	// redaction when the full match is a key=value assignment.
	captureGroup int
}

// Paths holds optional doublestar filters applied before matching.
type Paths struct {
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// Autofix is an optional, mechanical line rewrite (VibePatch). match is an
// RE2 regex applied per finding line; replace is the Go expansion template
// ($1 groups). Replace must not add or remove lines — VibeShield patches
// within the matched line only (contracts/rulepack.md).
type Autofix struct {
	Match   string `yaml:"match"`
	Replace string `yaml:"replace"`

	re *regexp.Regexp
}

// Re exposes the compiled autofix matcher.
func (a *Autofix) Re() *regexp.Regexp { return a.re }

// Rule is one validated rule with its compiled matcher.
type Rule struct {
	ID         string   `yaml:"id"`
	Category   string   `yaml:"category"`
	Severity   string   `yaml:"severity"`
	Title      string   `yaml:"title"`
	Message    string   `yaml:"message"`
	Fix        string   `yaml:"fix"`
	Autofix    *Autofix `yaml:"autofix,omitempty"`
	Languages  []string `yaml:"languages"`
	Pattern    Pattern  `yaml:"pattern"`
	Paths      Paths    `yaml:"paths,omitempty"`
	Confidence float64  `yaml:"confidence,omitempty"`
	References []string `yaml:"references,omitempty"`

	pack string // owning pack id:version (for reports)
}

// Pack is one pack document (one YAML file inside a pack directory).
type Pack struct {
	Schema  string `yaml:"schema"`
	ID      string `yaml:"id"`
	Version string `yaml:"version"`
	License string `yaml:"license"`
	Rules   []Rule `yaml:"rules"`
}

// LoadOptions tune pack loading. Reserved for future options; empty in v1.
type LoadOptions struct{}

// strictKeys decodes m into out and returns the mapping's raw key list.
func strictKeys(node *yaml.Node, out any, path string) ([]string, error) {
	keys := make([]string, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	if err := node.Decode(out); err != nil {
		return keys, fmt.Errorf("%s: %v", path, err)
	}
	return keys, nil
}

func checkKeys(keys []string, allowed map[string]bool, kind, path string) error {
	for _, k := range keys {
		if !allowed[k] {
			return fmt.Errorf("%s %q: unknown field %q", kind, path, k)
		}
	}
	return nil
}

var (
	packFields    = keyset("schema", "id", "version", "license", "rules")
	ruleFields    = keyset("id", "category", "severity", "title", "message", "fix", "autofix", "languages", "pattern", "paths", "confidence", "references")
	patternFields = keyset("kind", "match", "flags")
	autofixFields = keyset("match", "replace")
	pathsFields   = keyset("include", "exclude")
)

func keyset(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

// ParsePackBytes strictly validates and compiles one pack document.
// Any violation (unknown field, bad enum, non-RE2 regex, bad glob) is an
// error; callers map errors to exit code 2.
func ParsePackBytes(data []byte, name string, opts LoadOptions) (*Pack, []error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, []error{fmt.Errorf("%s: invalid YAML: %v", name, err)}
	}
	if root.Kind == 0 {
		return nil, []error{fmt.Errorf("%s: empty pack document", name)}
	}
	var doc *yaml.Node
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return nil, []error{fmt.Errorf("%s: empty pack document", name)}
		}
		doc = root.Content[0]
	} else {
		doc = &root
	}
	if doc.Kind != yaml.MappingNode {
		return nil, []error{fmt.Errorf("%s: pack must be a mapping", name)}
	}
	var errs []error

	keys, err := strictKeys(doc, &Pack{}, name)
	if err != nil {
		return nil, []error{err}
	}
	if err := checkKeys(keys, packFields, "pack", name); err != nil {
		errs = append(errs, err)
	}

	var pack Pack
	for i := 0; i+1 < len(doc.Content); i += 2 {
		k := doc.Content[i].Value
		v := doc.Content[i+1]
		switch k {
		case "schema":
			pack.Schema = v.Value
		case "id":
			pack.ID = v.Value
		case "version":
			pack.Version = v.Value
		case "license":
			pack.License = v.Value
		case "rules":
			if v.Kind != yaml.SequenceNode {
				errs = append(errs, fmt.Errorf("%s: rules must be a sequence", name))
				continue
			}
			for _, rn := range v.Content {
				r, rerrs := parseRule(rn, name, ruleFields, opts)
				errs = append(errs, rerrs...)
				if r != nil {
					pack.Rules = append(pack.Rules, *r)
				}
			}
		}
	}

	if pack.Schema != SchemaID {
		errs = append(errs, fmt.Errorf("%s: unsupported schema %q (want %q)", name, pack.Schema, SchemaID))
	}
	if pack.ID == "" {
		errs = append(errs, fmt.Errorf("%s: pack id required", name))
	}
	if pack.Version == "" {
		errs = append(errs, fmt.Errorf("%s: pack version required", name))
	}
	if pack.License == "" {
		errs = append(errs, fmt.Errorf("%s: pack license required", name))
	}
	if len(pack.Rules) == 0 && len(errs) == 0 {
		// A pack with zero rules is legal data but usually a mistake; keep it
		// non-fatal (the merge of many files can legitimately skip one).
		_ = 0
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return &pack, nil
}

func parseRule(node *yaml.Node, packName string, allowed map[string]bool, opts LoadOptions) (*Rule, []error) {
	if node.Kind != yaml.MappingNode {
		return nil, []error{fmt.Errorf("%s: rule must be a mapping", packName)}
	}
	var errs []error
	keys := make([]string, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	if err := checkKeys(keys, allowed, "rule", packName); err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	var rule Rule
	var patternNode, pathsNode, autofixNode *yaml.Node
	for i := 0; i+1 < len(node.Content); i += 2 {
		k := node.Content[i].Value
		v := node.Content[i+1]
		switch k {
		case "id":
			rule.ID = v.Value
		case "category":
			rule.Category = v.Value
		case "severity":
			rule.Severity = v.Value
		case "title":
			rule.Title = v.Value
		case "message":
			// yaml block scalars end with a newline; normalize to trimmed text.
			rule.Message = strings.TrimSpace(v.Value)
		case "fix":
			rule.Fix = strings.TrimSpace(v.Value)
		case "autofix":
			autofixNode = v
		case "languages":
			if err := v.Decode(&rule.Languages); err != nil {
				errs = append(errs, fmt.Errorf("rule %s: languages: %v", rule.ID, err))
			}
		case "pattern":
			patternNode = v
		case "paths":
			pathsNode = v
		case "confidence":
			var c float64
			if err := v.Decode(&c); err != nil {
				errs = append(errs, fmt.Errorf("rule %s: confidence: %v", rule.ID, err))
			} else {
				rule.Confidence = c
			}
		case "references":
			if err := v.Decode(&rule.References); err != nil {
				errs = append(errs, fmt.Errorf("rule %s: references: %v", rule.ID, err))
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}

	idCtx := rule.ID
	if idCtx == "" {
		idCtx = "<no id>"
	}
	if !ruleIDRe.MatchString(rule.ID) {
		errs = append(errs, fmt.Errorf("rule %q in %s: id must match VS-(PKG|SEC|LIC|DEP|INJ)-###", idCtx, packName))
	}
	if !allowedCategories[rule.Category] {
		errs = append(errs, fmt.Errorf("rule %s: unknown category %q", idCtx, rule.Category))
	}
	if !allowedSeverities[rule.Severity] {
		errs = append(errs, fmt.Errorf("rule %s: unknown severity %q", idCtx, rule.Severity))
	}
	if rule.Title == "" {
		errs = append(errs, fmt.Errorf("rule %s: title required", idCtx))
	} else if len([]rune(rule.Title)) > 80 {
		errs = append(errs, fmt.Errorf("rule %s: title exceeds 80 chars", idCtx))
	}
	if rule.Message == "" {
		errs = append(errs, fmt.Errorf("rule %s: message required", idCtx))
	}
	if rule.Fix == "" {
		errs = append(errs, fmt.Errorf("rule %s: fix required", idCtx))
	}
	if len(rule.Languages) == 0 {
		errs = append(errs, fmt.Errorf("rule %s: languages required", idCtx))
	}
	for _, l := range rule.Languages {
		if !AllowedLanguages[l] {
			errs = append(errs, fmt.Errorf("rule %s: unknown language %q", idCtx, l))
		}
	}
	if rule.Confidence == 0 {
		rule.Confidence = 0.8
	}
	if rule.Confidence < 0 || rule.Confidence > 1 {
		errs = append(errs, fmt.Errorf("rule %s: confidence out of [0,1]", idCtx))
	}

	// pattern
	if patternNode == nil || patternNode.Kind != yaml.MappingNode {
		errs = append(errs, fmt.Errorf("rule %s: pattern mapping required", idCtx))
	} else {
		pkeys := make([]string, 0, 3)
		for i := 0; i+1 < len(patternNode.Content); i += 2 {
			pkeys = append(pkeys, patternNode.Content[i].Value)
		}
		if err := checkKeys(pkeys, patternFields, "pattern", idCtx); err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		var p Pattern
		for i := 0; i+1 < len(patternNode.Content); i += 2 {
			k := patternNode.Content[i].Value
			v := patternNode.Content[i+1]
			switch k {
			case "kind":
				p.Kind = v.Value
			case "match":
				p.Match = v.Value
			case "flags":
				if err := v.Decode(&p.Flags); err != nil {
					errs = append(errs, fmt.Errorf("rule %s: flags: %v", idCtx, err))
				}
			}
		}
		switch p.Kind {
		case "regex", "literal":
			if p.Match == "" {
				errs = append(errs, fmt.Errorf("rule %s: pattern.match required for kind %s", idCtx, p.Kind))
			}
		case "structural":
			// Reserved for the v1.1 dependency-risk model (namescore/age).
			// Accepted at load so packs validate everywhere; the engine
			// skips structural rules (Re() stays nil).
			if p.Match == "" {
				errs = append(errs, fmt.Errorf("rule %s: pattern.match required for kind structural", idCtx))
			}
		default:
			errs = append(errs, fmt.Errorf("rule %s: pattern.kind must be regex|literal|structural (got %q)", idCtx, p.Kind))
		}
		for _, f := range p.Flags {
			if !allowedFlags[f] {
				errs = append(errs, fmt.Errorf("rule %s: unsupported pattern flag %q", idCtx, f))
			}
		}
		if len(errs) == 0 && (p.Kind == "regex" || p.Kind == "literal") {
			if err := compilePattern(&p, idCtx, errs == nil); err != nil {
				errs = append(errs, err)
			}
		}
		rule.Pattern = p
	}

	// autofix (optional, mechanical line rewrite — VibePatch)
	if autofixNode != nil {
		if autofixNode.Kind != yaml.MappingNode {
			errs = append(errs, fmt.Errorf("rule %s: autofix must be a mapping", idCtx))
		} else {
			ak := make([]string, 0, 2)
			for i := 0; i+1 < len(autofixNode.Content); i += 2 {
				ak = append(ak, autofixNode.Content[i].Value)
			}
			if err := checkKeys(ak, autofixFields, "autofix", idCtx); err != nil {
				errs = append(errs, err)
				return nil, errs
			}
			var a Autofix
			for i := 0; i+1 < len(autofixNode.Content); i += 2 {
				k := autofixNode.Content[i].Value
				v := autofixNode.Content[i+1]
				switch k {
				case "match":
					a.Match = v.Value
				case "replace":
					a.Replace = v.Value
				}
			}
			if a.Match == "" {
				errs = append(errs, fmt.Errorf("rule %s: autofix.match required", idCtx))
			} else if a.Replace == "" {
				errs = append(errs, fmt.Errorf("rule %s: autofix.replace required", idCtx))
			} else {
				if loc := unsupportedRe.FindStringIndex(a.Match); loc != nil {
					errs = append(errs, fmt.Errorf("rule %s: autofix.match uses a construct RE2 does not support (%q)", idCtx, a.Match[loc[0]:loc[1]]))
				} else if re, err := regexp.Compile(a.Match); err != nil {
					errs = append(errs, fmt.Errorf("rule %s: autofix.match does not compile under Go RE2: %v", idCtx, err))
				} else {
					a.re = re
					// A replace that adds a newline would silently
					// renumber every finding below it — refuse.
					if strings.ContainsAny(a.Replace, "\n\r") {
						errs = append(errs, fmt.Errorf("rule %s: autofix.replace must stay on one line", idCtx))
					}
					// Go expands `$1abc` as group "1abc" (invalid → empty).
					// Group references in autofix templates must use the
					// braced form `${1}`; reject dangling names so a typo
					// can never delete code silently.
					if bad := badGroupRefs(re, a.Replace); bad != "" {
						errs = append(errs, fmt.Errorf("rule %s: autofix.replace references unknown group %q (match has %d groups; use ${1}…${%d})", idCtx, bad, re.NumSubexp(), re.NumSubexp()))
					}
					rule.Autofix = &a
				}
			}
		}
	}

	// paths
	if pathsNode != nil {
		if pathsNode.Kind != yaml.MappingNode {
			errs = append(errs, fmt.Errorf("rule %s: paths must be a mapping", idCtx))
		} else {
			pk := make([]string, 0, 2)
			for i := 0; i+1 < len(pathsNode.Content); i += 2 {
				pk = append(pk, pathsNode.Content[i].Value)
			}
			if err := checkKeys(pk, pathsFields, "paths", idCtx); err != nil {
				errs = append(errs, err)
				return nil, errs
			}
			for i := 0; i+1 < len(pathsNode.Content); i += 2 {
				k := pathsNode.Content[i].Value
				v := pathsNode.Content[i+1]
				var g []string
				if err := v.Decode(&g); err != nil {
					errs = append(errs, fmt.Errorf("rule %s: paths.%s: %v", idCtx, k, err))
					continue
				}
				for _, pat := range g {
					if _, err := doublestar.Match(pat, "probe/x.y"); err != nil {
						errs = append(errs, fmt.Errorf("rule %s: bad glob %q in paths.%s", idCtx, pat, k))
					}
				}
				if k == "include" {
					rule.Paths.Include = g
				} else {
					rule.Paths.Exclude = g
				}
			}
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return &rule, nil
}

var (
	// RE2-excluded constructs (Go's regexp rejects most at compile time, but
	// we also want an explicit, spec-worded diagnostic for the common ones):
	// lookahead/lookbehind (?= (?! (?<=, atomic groups (?>, backreferences \1,
	// and possessive quantifiers x++ x*+ x?+.
	unsupportedRe = regexp.MustCompile(`\(\?[=!<]|\(\?>|\\[1-9]|[+*?]\+`)
)

// compilePattern builds the RE2 regexp enforcing the no-lookahead /
// no-backreference / no-possessive rule and applying flags.
func compilePattern(p *Pattern, ruleID string, _ bool) error {
	expr := p.Match
	if p.Kind == "literal" {
		expr = regexp.QuoteMeta(expr)
	} else {
		if loc := unsupportedRe.FindStringIndex(expr); loc != nil {
			return fmt.Errorf("rule %s: pattern uses a construct RE2 does not support (%q): no lookahead/backreference/possessive quantifiers allowed", ruleID, expr[loc[0]:loc[1]])
		}
	}
	prefix := ""
	for _, f := range p.Flags {
		switch f {
		case "multiline":
			prefix += "(?m)"
		case "caseless":
			prefix += "(?i)"
		case "dotall":
			prefix += "(?s)"
		}
	}
	re, err := regexp.Compile(prefix + expr)
	if err != nil {
		return fmt.Errorf("rule %s: pattern does not compile under Go RE2: %v", ruleID, err)
	}
	p.re = re
	p.captureGroup = 0
	if p.Kind == "regex" && re.NumSubexp() >= 1 {
		p.captureGroup = 1
	}
	return nil
}

// Re exposes the compiled matcher for the engine.
func (p *Pattern) Re() *regexp.Regexp { return p.re }

// badGroupRefs validates a Replace template against re's groups: every $name
// / ${name} must be an existing numeric group (1..NumSubexp) or a declared
// named group. Go's Expand silently drops unknown names, which in an autofix
// would delete the wrong span — `$1DEBUG` (meant `${1}DEBUG`) is exactly the
// class of typo that must be a load error, not a corrupted file.
func badGroupRefs(re *regexp.Regexp, tmpl string) string {
	names := map[string]bool{}
	for i, n := range re.SubexpNames() {
		if i > 0 && n != "" {
			names[n] = true
		}
	}
	isWord := func(c byte) bool {
		return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	}
	for i := 0; i < len(tmpl); i++ {
		if tmpl[i] != '$' {
			continue
		}
		j := i + 1
		braced := false
		if j < len(tmpl) && tmpl[j] == '{' {
			braced = true
			j++
		}
		k := j
		for k < len(tmpl) && isWord(tmpl[k]) {
			k++
		}
		name := tmpl[j:k]
		if braced && (k >= len(tmpl) || tmpl[k] != '}') {
			return name
		}
		if name == "" {
			continue // literal $
		}
		if n, err := strconv.Atoi(name); err == nil {
			if n < 1 || n > re.NumSubexp() {
				return name
			}
			continue
		}
		if !names[name] {
			return name
		}
	}
	return ""
}

// CaptureGroup returns the submatch index holding the sensitive value
// (1 when the pattern has a capture group, else 0 = whole match).
func (p *Pattern) CaptureGroup() int { return p.captureGroup }

// MatchesPath applies include/exclude doublestar globs.
func (r *Rule) MatchesPath(rel string) bool {
	for _, g := range r.Paths.Exclude {
		if ok, _ := doublestar.Match(g, rel); ok {
			return false
		}
	}
	if len(r.Paths.Include) == 0 {
		return true
	}
	for _, g := range r.Paths.Include {
		if ok, _ := doublestar.Match(g, rel); ok {
			return true
		}
	}
	return false
}

// AppliesToLanguage reports whether the rule runs against a language, or
// "generic" (used by file-language filter).
func (r *Rule) AppliesToLanguage(lang string) bool {
	for _, l := range r.Languages {
		if l == lang || l == "generic" {
			return true
		}
	}
	return false
}

// PackInfo is pack metadata for `vibeshield version`.
type PackInfo struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	License string `json:"license"`
	Rules   int    `json:"rules"`
}

// SetPack stamps owning pack info.
func (r *Rule) SetPack(id, version string) { r.pack = id + ":" + version }

// Pack returns the owning pack stamp.
func (r *Rule) Pack() string { return r.pack }
