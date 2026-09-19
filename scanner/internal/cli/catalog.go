package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

// Agent is one AI coding agent VibeShield drops first-class rules into.
type Agent struct {
	Name    string   // display name
	Slug    string   // stable id used by `vibeshield agents <slug>`
	File    string   // the file VibeShield writes into the user's project
	Scope   string   // where that file is read from
	Aliases []string // search aliases
	Snippet string   // ready-to-paste rule body
	Steps   []string // human install steps
}

// Agents is the v2 agent matrix. Paths follow each vendor's current rules
// convention (verified 2026-09); legacy filenames are kept in the steps so
// older client versions still work.
func Agents() []Agent {
	return []Agent{
		{
			Name:    "Codex (OpenAI)",
			Slug:    "codex",
			File:    "AGENTS.md",
			Scope:   "repo root — read before every task; ~/.codex/AGENTS.md applies globally",
			Aliases: []string{"codex", "openai", "agents.md", "agentsmd", "codex cli"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Add AGENTS.md to the repo root (VibeShield already ships one).",
				"Ask Codex to run `vibeshield scan --staged` before it proposes a commit.",
				"For a global default, copy the same block to ~/.codex/AGENTS.md.",
			},
		},
		{
			Name:    "Claude Code (CLI + Desktop)",
			Slug:    "claude-code",
			File:    "CLAUDE.md + the vibeshield plugin",
			Scope:   "repo root CLAUDE.md; plugin installed per user or per project",
			Aliases: []string{"claude", "claude code", "claude desktop", "anthropic", "plugin", "skill"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Run: claude plugin marketplace add rajviyash9136freefr-tech/vibeshield",
				"Then: claude plugin install vibeshield@vibeshield",
				"Audit any project with /vibeshield:vibeshield-audit --quick.",
				"Desktop app: the same plugin works once the marketplace is added.",
			},
		},
		{
			Name:    "Google Antigravity",
			Slug:    "antigravity",
			File:    ".agents/rules/vibeshield.md",
			Scope:   "workspace/git root; ~/.gemini/GEMINI.md applies globally",
			Aliases: []string{"antigravity", "google", "gemini", "agy", ".agents", "rules"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Create .agents/rules/vibeshield.md in the workspace root (VibeShield ships this).",
				"Set the rule's activation to Always On (or Glob: package.json, **/*.ts) in the Customizations panel.",
				"Legacy Antigravity builds read .agent/rules/ instead — keep that folder if you are on an older release.",
				"For every workspace, append the same block to ~/.gemini/GEMINI.md.",
			},
		},
		{
			Name:    "Cursor",
			Slug:    "cursor",
			File:    ".cursor/rules/vibeshield.mdc",
			Scope:   "repo root; alwaysApply: true",
			Aliases: []string{"cursor", "cursorrules", "mdc", "cursor ide"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Create .cursor/rules/vibeshield.mdc (VibeShield ships this).",
				"Set alwaysApply: true so the rule survives every chat.",
				"Older Cursor builds only read .cursorrules — VibeShield ships that file too.",
			},
		},
		{
			Name:    "Windsurf",
			Slug:    "windsurf",
			File:    ".windsurf/rules/vibeshield.md",
			Scope:   "repo root, applied to Cascade",
			Aliases: []string{"windsurf", "cascade", "codeium", "windsurfrules"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Create .windsurf/rules/vibeshield.md (VibeShield ships this).",
				"Older Cascade builds read .windsurfrules — VibeShield ships that file too.",
			},
		},
		{
			Name:    "GitHub Copilot",
			Slug:    "copilot",
			File:    ".github/copilot-instructions.md",
			Scope:   "repo root, applied to every Copilot chat in the repo",
			Aliases: []string{"copilot", "github copilot", "vscode", "vs code", "coding agent"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Create .github/copilot-instructions.md (VibeShield ships this).",
				"Optionally add .github/instructions/*.instructions.md for path-scoped rules.",
			},
		},
		{
			Name:    "Any AGENTS.md agent",
			Slug:    "generic",
			File:    "AGENTS.md",
			Scope:   "repo root — Cline, Roo Code, Amp, Zed, Aider, Gemini CLI, OpenCode",
			Aliases: []string{"cline", "roo", "amp", "zed", "aider", "opencode", "generic", "agents"},
			Snippet: AgentRulesBody,
			Steps: []string{
				"Any client that reads the AGENTS.md convention picks the rule up for free.",
				"Keep the block under ~2 KB; agents that truncate long context keep the critical rules.",
			},
		},
	}
}

// AgentRulesBody is the shared, agent-agnostic rule block. It is deliberately
// short: every client truncates long context, and the scanner covers the rest.
// Kept free of backticks so it can stay a raw string literal — the same text
// ships as AGENTS.md, .cursor/rules/vibeshield.mdc and friends.
const AgentRulesBody = `# VibeShield — bug hunter for AI-generated code

Act as an adversarial reviewer for code this agent writes. Before proposing,
applying, or committing a change:

1. HALLUCINATED PACKAGES — every new dependency must exist before the model's
   training cutoff. If you are not certain, say so and use the standard library.
   A plausible-but-invented name ("fast-parse-utils-v3") is a slopsquatting trap.
2. SECRETS — never write API keys, tokens, or passwords into source, tests,
   fixtures, or client bundles. Read them from the environment; if one is
   already in the diff, flag it for rotation.
3. INSECURE DEFAULTS — refuse to scaffold wildcard CORS, debug=true, unhashed
   passwords, JWT alg "none", eval() on model output, or TLS verification off.
4. VERIFY, THEN SHIP — run "vibeshield scan --staged" before every commit and
   "vibeshield scan . --format json" when you need machine-readable findings.
5. ATOMIC FIXES — for each finding, propose a one-line diff, not a lecture.
`

// Catalog is the searchable entry set the console shows: the CLI's own
// actions, the per-agent setup recipes, and every rule in the loaded packs.
func Catalog(pack *rules.Pack, version string) []Item {
	items := actionItems(version)
	items = append(items, agentItems()...)
	items = append(items, ruleItems(pack)...)
	return items
}

// actionItems are the CLI verbs, phrased as the task the user wants done —
// people search for "check my code", not for "scan".
func actionItems(version string) []Item {
	return []Item{
		{
			Kind: KindAction, Group: "Setup",
			Title:   "Set up this project",
			Summary: "Detect the stack, write vibeshield.yml, a PR gate and a hook.",
			Keywords: []string{
				"init", "setup", "install", "configure", "configure", "start", "onboard",
				"workflow", "hook", "pre commit", "scaffold", "bootstrap", "ci", "yml",
			},
			Args: []string{"init", ".", "--dry-run"},
		},
		{
			Kind: KindAction, Group: "Setup",
			Title:   "Check what is set up (and what is not)",
			Summary: "Read-only report: config, git hook, PR gate, agent rules.",
			Keywords: []string{
				"doctor", "check", "health", "diagnose", "status", "setup", "configure",
				"why", "not working", "hook", "workflow", "debug", "verify", "install",
			},
			Args: []string{"doctor"},
		},
		{
			Kind: KindAction, Group: "Setup",
			Title:    "Set up this project (write the files)",
			Summary:  "Same as above, without the dry-run preview.",
			Keywords: []string{"init", "setup", "write", "apply", "configure", "workflow", "hook", "ci"},
			Args:     []string{"init", "."},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:   "Scan this project",
			Summary: "Full offline audit of the current directory.",
			Keywords: []string{
				"scan", "audit", "check", "test", "review", "bug hunt", "full", "local",
				"offline", "run", "start", "analyse", "analyze", "vibe check", "security",
			},
			Args: []string{"scan", "."},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:    "Scan only my changes vs main",
			Summary:  "Diff mode — reports findings on added lines only.",
			Keywords: []string{"scan", "diff", "changes", "pr", "pull request", "main", "added", "fast", "incremental"},
			Args:     []string{"scan", "--diff", "main"},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:    "Scan staged changes (pre-commit)",
			Summary:  "The same gate the pre-commit hook runs before every commit.",
			Keywords: []string{"scan", "staged", "pre commit", "precommit", "hook", "git", "commit", "before commit"},
			Args:     []string{"scan", "--staged"},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:    "Scan as JSON",
			Summary:  "contracts/cli.md report — feed it to agents and CI.",
			Keywords: []string{"scan", "json", "machine readable", "output", "agent", "ci", "pipe", "report"},
			Args:     []string{"scan", ".", "--format", "json"},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:    "Scan as GitHub annotations",
			Summary:  "::error / ::warning lines for GitHub Actions logs.",
			Keywords: []string{"scan", "github", "annotations", "actions", "ci", "workflow", "pipeline"},
			Args:     []string{"scan", ".", "--format", "github"},
		},
		{
			Kind: KindAction, Group: "Scan",
			Title:    "Scan as SARIF",
			Summary:  "SARIF 2.1.0 for GitHub code scanning (Security tab).",
			Keywords: []string{"scan", "sarif", "code scanning", "security tab", "codeql", "upload", "github", "alerts", "ci", "sast", "integration"},
			Args:     []string{"scan", ".", "--format", "sarif"},
		},
		{
			Kind: KindAction, Group: "Fix",
			Title:    "Preview fixes (VibePatch dry run)",
			Summary:  "Show the exact −/+ diff without changing a single file.",
			Keywords: []string{"fix", "patch", "vibepatch", "preview", "dry run", "diff", "autofix", "repair"},
			Args:     []string{"fix", ".", "--dry-run"},
		},
		{
			Kind: KindAction, Group: "Fix",
			Title:    "Apply fixes",
			Summary:  "Write the mechanical fixes and log every patch.",
			Keywords: []string{"fix", "patch", "vibepatch", "apply", "write", "autofix", "repair", "yes"},
			Args:     []string{"fix", ".", "--yes"},
		},
		{
			Kind: KindAction, Group: "Search",
			Title:    "Search rule packs from the shell",
			Summary:  "Same ranking as this console, scriptable — vibeshield search <query>.",
			Keywords: []string{"search", "find", "grep", "query", "rules", "keywords", "lookup", "index"},
			Args:     []string{"search", "--list"},
		},
		{
			Kind: KindAction, Group: "Search",
			Title:    "List every agent setup recipe",
			Summary:  "Codex, Claude Code, Antigravity, Cursor, Windsurf, Copilot.",
			Keywords: []string{"agents", "agent", "setup", "install", "codex", "claude", "antigravity", "cursor", "windsurf", "copilot", "rules"},
			Args:     []string{"agents"},
		},
		{
			Kind: KindAction, Group: "Search",
			Title:    "List every rule",
			Summary:  "The whole core pack, grouped by category — vibeshield rules.",
			Keywords: []string{"rules", "rule", "list", "pack", "all", "catalog", "reference", "what does it check"},
			Args:     []string{"rules"},
		},
		{
			Kind: KindAction, Group: "Setup",
			Title:    "Set up shell completion",
			Summary:  "Tab-complete every command and flag: bash, zsh, fish, PowerShell.",
			Keywords: []string{"completion", "complete", "tab", "autocomplete", "shell", "bash", "zsh", "fish", "powershell", "install"},
			Args:     []string{"completion", "bash"},
		},
		{
			Kind: KindAction, Group: "Meta",
			Title:    "Show version and rule packs",
			Summary:  "Binary version, pack ids, licence, rule count.",
			Keywords: []string{"version", "about", "info", "build", "packs", "license", "licence"},
			Args:     []string{"version"},
		},
		{
			Kind: KindAction, Group: "Meta",
			Title:    "Show all commands",
			Summary:  "Full flag reference for scan and fix.",
			Keywords: []string{"help", "usage", "commands", "flags", "options", "manual", "docs", "reference"},
			Args:     []string{"help"},
		},
	}
}

// agentItems turn each agent recipe into a searchable, readable entry.
func agentItems() []Item {
	ags := Agents()
	out := make([]Item, 0, len(ags))
	for _, a := range ags {
		kws := append([]string{"agent", "setup", "rules", a.Slug, a.File}, a.Aliases...)
		out = append(out, Item{
			Kind:     KindInfo,
			Group:    "Agent setup",
			Title:    a.Name,
			Summary:  "Write " + a.File + " — " + a.Scope,
			Keywords: kws,
			Body:     agentBody(a),
		})
	}
	return out
}

func agentBody(a Agent) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n\n", a.Name, strings.Repeat("─", len(a.Name)+8))
	fmt.Fprintf(&b, "File    %s\n", a.File)
	fmt.Fprintf(&b, "Scope   %s\n", a.Scope)
	fmt.Fprintf(&b, "Shell   vibeshield agents %s\n\n", a.Slug)
	b.WriteString("Install\n")
	for i, s := range a.Steps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, s)
	}
	b.WriteString("\nRule body (paste into " + a.File + ")\n\n")
	b.WriteString(a.Snippet)
	return b.String()
}

// ruleItems exposes every rule in the loaded packs as a searchable entry, so
// "aws key", "vs-sec-017" and "cors" all reach the same document.
func ruleItems(pack *rules.Pack) []Item {
	if pack == nil || len(pack.Rules) == 0 {
		return nil
	}
	rs := append([]rules.Rule(nil), pack.Rules...)
	sort.SliceStable(rs, func(i, j int) bool { return rs[i].ID < rs[j].ID })

	out := make([]Item, 0, len(rs))
	for _, r := range rs {
		kws := make([]string, 0, 4+len(r.Languages))
		kws = append(kws, r.ID, r.Category, r.Severity, r.Fix)
		kws = append(kws, r.Languages...)
		out = append(out, Item{
			Kind:     KindInfo,
			Group:    "Rule · " + r.Category,
			Title:    r.ID + "  " + r.Title,
			Summary:  severityLabel(r.Severity) + " · " + r.Category,
			Keywords: kws,
			Body:     ruleBody(r),
		})
	}
	return out
}

func ruleBody(r rules.Rule) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %s\n%s\n\n", r.ID, r.Title, strings.Repeat("─", len(r.ID)+len(r.Title)+3))
	fmt.Fprintf(&b, "Severity    %s\n", severityLabel(r.Severity))
	fmt.Fprintf(&b, "Category    %s\n", r.Category)
	fmt.Fprintf(&b, "Languages   %s\n", strings.Join(r.Languages, ", "))
	fmt.Fprintf(&b, "Pack        %s\n", r.Pack())
	if r.Confidence > 0 {
		fmt.Fprintf(&b, "Confidence  %.2f\n", r.Confidence)
	}
	if !r.Evaluable() {
		// A reserved rule loads and validates but the matcher skips it, so a
		// scan will never report it. Saying so here beats letting someone
		// search for a rule that cannot fire.
		fmt.Fprintf(&b, "\n%s\n", "Status      RESERVED — this build cannot evaluate it yet.\n"+
			"            The pack ships it so packs stay portable across versions;\n"+
			"            see contracts/rulepack.md (pattern.kind: structural).")
	}
	fmt.Fprintf(&b, "\nWhy it matters\n\n%s\n", r.Message)
	fmt.Fprintf(&b, "\nFix\n\n%s\n", r.Fix)
	if r.Autofix != nil {
		fmt.Fprintf(&b, "\nAutofix (VibePatch)\n\n  match:   %s\n  replace: %s\n", r.Autofix.Match, r.Autofix.Replace)
	}
	if len(r.References) > 0 {
		fmt.Fprintf(&b, "\nReferences\n\n")
		for _, ref := range r.References {
			fmt.Fprintf(&b, "  %s\n", ref)
		}
	}
	return b.String()
}

func severityLabel(s string) string {
	switch s {
	case "critical":
		return "🔴 CRITICAL"
	case "high":
		return "🟠 HIGH"
	case "medium":
		return "🟡 MEDIUM"
	case "low":
		return "🔵 LOW"
	case "info":
		return "⚪ INFO"
	}
	return strings.ToUpper(s)
}
