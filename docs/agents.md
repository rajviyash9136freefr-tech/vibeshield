# AI agent setup

VibeShield ships a rules file for every major coding agent. All seven files are
**generated from a single source of truth** — the `AgentRulesBody` constant in
[`scanner/internal/cli/catalog.go`](../scanner/internal/cli/catalog.go) — so the
console, `vibeshield agents --body`, and the checked-in files can never disagree.

```bash
node scripts/sync-agent-rules.mjs           # regenerate all seven files
node scripts/sync-agent-rules.mjs --check   # CI: fail on drift
```

## The matrix

| Agent | File | Scope | Notes |
|:---|:---|:---|:---|
| **Codex** (OpenAI) | [`AGENTS.md`](../AGENTS.md) | repo root, read before every task | `~/.codex/AGENTS.md` applies globally |
| **Claude Code** (CLI + Desktop) | [`CLAUDE.md`](../CLAUDE.md) + the `vibeshield` plugin | repo root; plugin is per user or per project | the plugin also ships the audit skill |
| **Google Antigravity** | [`.agents/rules/vibeshield.md`](../.agents/rules/vibeshield.md) | workspace / git root | older builds read `.agent/rules/`; `~/.gemini/GEMINI.md` is global |
| **Cursor** | [`.cursor/rules/vibeshield.mdc`](../.cursor/rules/vibeshield.mdc) | repo root, `alwaysApply: true` | [`.cursorrules`](../.cursorrules) is the legacy fallback |
| **Windsurf** | [`.windsurf/rules/vibeshield.md`](../.windsurf/rules/vibeshield.md) | repo root, applies to Cascade | [`.windsurfrules`](../.windsurfrules) is the legacy fallback |
| **GitHub Copilot** | [`.github/copilot-instructions.md`](../.github/copilot-instructions.md) | every Copilot chat in the repo | pair with `.github/instructions/*.instructions.md` for path scoping |
| **Any `AGENTS.md` client** | [`AGENTS.md`](../AGENTS.md) | repo root | Cline · Roo Code · Amp · Zed · Aider · Gemini CLI · OpenCode |

## Install

The rules files are already in this repository — if you cloned it, you are done.
To drop the rule block into **another** project:

```bash
vibeshield agents --body >> /path/to/your/project/AGENTS.md
```

Then, per agent:

**Codex**

```bash
vibeshield agents --body > AGENTS.md          # project scope
mkdir -p ~/.codex && vibeshield agents --body > ~/.codex/AGENTS.md   # global
```

**Claude Code** (CLI and Desktop)

```bash
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield
claude plugin install vibeshield@vibeshield
```

Then in any project: `/vibeshield:vibeshield-audit --quick`. The plugin bundles
the parallel bug-hunter skill, which is the only surface that hunts `VS-INJ`
prompt-injection traps today.

**Google Antigravity**

```bash
mkdir -p .agents/rules
vibeshield agents --body > .agents/rules/vibeshield.md
```

In the Customizations panel, set the rule's activation to **Always On**, or use a
**Glob** trigger such as `package.json` / `**/*.ts` to keep it scoped.

**Cursor**

```bash
mkdir -p .cursor/rules
vibeshield agents --body > .cursor/rules/vibeshield.mdc
```

Prepend the MDC frontmatter (`alwaysApply: true`) — the file shipped in this repo
already has it.

**Windsurf**

```bash
mkdir -p .windsurf/rules
vibeshield agents --body > .windsurf/rules/vibeshield.md
```

**GitHub Copilot**

```bash
vibeshield agents --body > .github/copilot-instructions.md
```

## What the rule block asks for

Five constraints, deliberately short so agents that truncate long context keep
the important ones:

1. **Hallucinated packages** — every new dependency must exist before the model's
   training cutoff; otherwise say so and use the standard library.
2. **Secrets** — never write keys, tokens or passwords into source, tests,
   fixtures or client bundles; flag existing ones for rotation.
3. **Insecure defaults** — refuse wildcard CORS, `debug=true`, unhashed
   passwords, `JWT alg: "none"`, `eval()` on model output, TLS verification off.
4. **Verify, then ship** — run `vibeshield scan --staged` before every commit and
   `vibeshield scan . --format json` when machine-readable findings are needed.
5. **Atomic fixes** — one-line diffs, not lectures.

## Why the rule file and the scanner both exist

The rule file shapes what the agent *writes*; the scanner checks what actually
*landed*. Rules are advisory and can be ignored by a model; `vibeshield scan
--staged` and the [GitHub Action](../action/) are the parts that cannot be.
Use both: the agent stops most bad code, the gate catches the rest.
