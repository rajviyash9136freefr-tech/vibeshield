<div align="center">

# 🛡️ VibeShield

### The bug hunter & terminal workspace for AI-generated code

**Scan your vibe-coded app for hallucinated packages, leaked API keys and insecure AI defaults — then browse, search and fix every finding without leaving the terminal.**

[![Version](https://img.shields.io/badge/version-2.0.1-white?style=for-the-badge&logo=git&logoColor=black)](CHANGELOG.md)
[![Website](https://img.shields.io/badge/🌐_Website-Live_Simulator-white?style=for-the-badge&logo=googlechrome&logoColor=black)](https://rajviyash9136freefr-tech.github.io/vibeshield/)
[![GitHub Stars](https://img.shields.io/github/stars/rajviyash9136freefr-tech/vibeshield?style=for-the-badge&logo=github&color=white&labelColor=black)](https://github.com/rajviyash9136freefr-tech/vibeshield/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-white?style=for-the-badge)](LICENSE)

[![Agents](https://img.shields.io/badge/Agents-Codex_·_Claude_Code_·_Antigravity_·_Cursor_·_Windsurf_·_Copilot-111111?style=flat-square)](docs/agents.md)
[![Offline](https://img.shields.io/badge/100%25_Local_&_Offline-Zero_Data_Sent_Outside-success?style=flat-square)](#-frequently-asked-questions)
[![Rules](https://img.shields.io/badge/Core_rules-122_across_6_categories-blue?style=flat-square)](rules/core)
[![Scan speed](https://img.shields.io/badge/Scan_speed-168_files_in_1.7s-blue?style=flat-square)](#-how-to-scan--hunt-bugs)

<br/>

[⚡ **Try the live simulator**](https://rajviyash9136freefr-tech.github.io/vibeshield/#scanner) · [🖥️ **The v2 console**](#-the-v2-console--vibeshield-with-no-arguments) · [🔍 **Search**](#-search-everything-vibeshield-search) · [🤖 **Agent setup**](#-ai-agent-setup--codex-claude-code-antigravity-cursor-windsurf-copilot) · [📦 **Install**](#-1-command-quick-install) · [❓ **FAQ**](#-frequently-asked-questions)

<br/>

```text
  $ vibeshield scan .
  Scanning 14 vibe-coded files… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     package.json:8 — fast-parse-utils-v3@2.1.4 (registered 9 days ago on npm)
     → Fix: replace with node:util (12 lines, zero dependencies).

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     src/agent.ts:41 — OPENAI_API_KEY pasted from chat context
     → Fix: read process.env.OPENAI_API_KEY and rotate the key now.

  ✓ 12 files clean · 2 findings · 0 leaks merged to main
```

</div>

---

## 🆕 What's new in v2

v1 was a scanner you piped into CI. **v2 is a workspace you sit in.**

| | v1.0.0 | **v2.0.0** |
|:---|:---|:---|
| Bare `vibeshield` | printed usage | **opens a searchable interactive console** |
| Project setup | wire it up by hand | **`vibeshield init` — detects the stack, writes config + PR gate + hook** |
| Finding a rule | read the YAML | **`vibeshield search "aws key"`** |
| Agent setup | one prompt in the README | **7 generated, drift-checked rule files** |
| Agents covered | Cursor · Claude Code · Windsurf · Copilot | **+ Codex · Google Antigravity · Claude Code Desktop · any `AGENTS.md` client** |
| Action downloads | unverified | **verified against the release `sha256sums.txt`** |
| Rule packs | `core 1.0.0` | `core 2.0.0` |

Full details in [CHANGELOG.md](CHANGELOG.md).

---

## 🖥️ The v2 console — `vibeshield` with no arguments

Run `vibeshield` in any project and you get one search box over **every action, every rule, and every agent setup recipe**:

```text
  VibeShield 2.0.1                                    140 entries · fully offline
  ────────────────────────────────────────────────────────────────────────────────
  ❯ scan▏
  ────────────────────────────────────────────────────────────────────────────────
  SCAN
    ▸ Scan this project                        Full offline audit of .
      Scan only my changes vs main             Diff mode — added lines only
      Scan staged changes (pre-commit)         The gate the hook runs
      Scan as JSON                             contracts/cli.md report
  FIX
      Preview fixes (VibePatch dry run)        Show the −/+ diff, change nothing
  ────────────────────────────────────────────────────────────────────────────────
  ↑↓ move · ⏎ open · tab next group · esc clear · ctrl+c quit          1/140
```

**Keys** — `↑`/`↓` (or `Ctrl+P`/`Ctrl+N`) move · `Tab` jumps to the next group · `Enter` opens · `Esc` clears the query, then quits · `Ctrl+U` resets · `Ctrl+C` quits.

Selecting an action hands the terminal back and runs the **real** command — the menu is wired to the same dispatcher as the CLI, so it can never drift from the documented flags. `vibeshield ui` opens the console explicitly; when stdin is a pipe it falls back to plain usage text and exit `2`, so scripts are unaffected.

No TTY? No problem — `vibeshield search` is the same index on stdout.

---

## 🔍 Search everything: `vibeshield search`

The console's ranking, scriptable:

```bash
vibeshield search aws                    # AWS credential rules
vibeshield search vs-sec-017             # one rule, by id
vibeshield search "prompt injection"     # multi-word: every token must match
vibeshield search --agents cursor        # agent setup recipes only
vibeshield search --rules --list         # every shipped rule
vibeshield search aws --format json      # vibeshield.search/v1 for agents and CI
```

The ranker folds separators (`vs-sec-017` ≡ `vs sec 017` ≡ `vssec017`), prefers word-boundary hits, lets keywords reinforce a title match, and only falls back to fuzzy subsequence matching for queries long enough to be meaningful — so a three-letter search returns the AWS rule, not a wall of coincidences.

---

## 🤖 AI agent setup — Codex, Claude Code, Antigravity, Cursor, Windsurf, Copilot

VibeShield ships a rule file for every major agent, **generated from one source of truth** and checked in CI so the docs can never lie:

| Agent | File VibeShield writes | Scope |
|:---|:---|:---|
| **Codex** (OpenAI) | [`AGENTS.md`](AGENTS.md) | repo root · `~/.codex/AGENTS.md` for global |
| **Claude Code** (CLI + Desktop) | [`CLAUDE.md`](CLAUDE.md) + plugin | per project / per user |
| **Google Antigravity** | [`.agents/rules/vibeshield.md`](.agents/rules/vibeshield.md) | workspace root · `~/.gemini/GEMINI.md` for global |
| **Cursor** | [`.cursor/rules/vibeshield.mdc`](.cursor/rules/vibeshield.mdc) · [`.cursorrules`](.cursorrules) | repo root, `alwaysApply: true` |
| **Windsurf** | [`.windsurf/rules/vibeshield.md`](.windsurf/rules/vibeshield.md) · [`.windsurfrules`](.windsurfrules) | repo root, applies to Cascade |
| **GitHub Copilot** | [`.github/copilot-instructions.md`](.github/copilot-instructions.md) | every Copilot chat in the repo |
| **Any `AGENTS.md` client** | `AGENTS.md` | Cline · Roo Code · Amp · Zed · Aider · Gemini CLI |

Print any recipe, or the pasteable rule block:

```bash
vibeshield agents                        # every recipe
vibeshield agents cursor                 # one agent
vibeshield agents --body > AGENTS.md     # just the shared rule block
vibeshield agents --markdown             # the matrix as a table
```

### Install the audit skill

**Claude Code** (CLI and Desktop):

```bash
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield
claude plugin install vibeshield@vibeshield
# then, in any project:
/vibeshield:vibeshield-audit --quick
```

The skill spawns parallel bug-hunter subagents, verifies every CRITICAL/HIGH candidate against an adversarial skeptic, and writes `vibeshield-findings.json` in the [finding contract](contracts/finding/schema.json). It is the only surface that hunts **`VS-INJ` prompt-injection traps** today — malicious READMEs, issue text and `AGENTS.md` diffs that try to steer your agent.

---

## ⚡ Why VibeShield for AI-generated code

Classic linters were built for human code and known CVE databases. They miss the failure modes LLMs actually produce:

| AI failure mode | Why it bites | Classic linters | VibeShield |
|:---|:---|:---:|:---:|
| **Hallucinated packages** (slopsquatting) | The model invents `fast-parse-utils-v3`; an attacker registers it within hours | ❌ no CVE exists | 🚨 **blocked** at the PR diff |
| **Secrets pasted from chat** | `OPENAI_API_KEY` lands in a client bundle or a test fixture | ⚠️ noisy generics | 🔒 **quarantined** pre-commit |
| **Insecure AI boilerplate** | Wildcard CORS, `debug=True`, unhashed passwords, `JWT alg: none`, `eval()` | ⚠️ buried in lint noise | ⚡ **flagged** with a one-line fix |
| **License stripping** | A model rewrites vendored code and drops the SPDX header | ❌ no coverage | 📄 **flagged** (`VS-LIC`) |
| **Dependency drift** | Unpinned ranges, install hooks that `curl \| bash`, single-maintainer packages | ⚠️ partial | 📦 **flagged** (`VS-DEP`, `VS-PKG`) |
| **Prompt-injection traps** | A README or `AGENTS.md` diff tells your agent to open a backdoor | ❌ zero coverage | 🛡️ **hunted by the agent skill** (`VS-INJ`) |

The shipped core pack is **122 rules across 6 categories**: `hallucinated-package` · `hardcoded-secret` · `insecure-api` · `insecure-default` · `dependency-risk` · `license-missing`. Run `vibeshield search --rules --list` to read every one.

---

## 🚀 1-command quick install

Zero dependencies, no sudo.

### 🍏 macOS & 🐧 Linux

```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

### 🪟 Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex
```

### 📦 Node / npx (zero-install scan)

```bash
npx vibeshield scan .
```

### 🐹 Go developers

```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@v2.0.1
```

### 🧰 Build from source

```bash
git clone https://github.com/rajviyash9136freefr-tech/vibeshield.git
cd vibeshield/scanner && go build -o vibeshield ./cmd/vibeshield
```

Pure Go, `CGO_ENABLED=0`, no third-party runtime dependencies.

---

## 🔍 How to scan & hunt bugs

```bash
# 0. Set the project up — detects the stack, writes vibeshield.yml,
#    a PR-gate workflow and a pre-commit hook, then runs the first scan
vibeshield init

# 1. Audit the whole project
vibeshield scan .

# 2. Audit only what changed against main (fast PR mode)
vibeshield scan --diff main

# 3. Pre-commit mode, before anything enters git history
vibeshield scan --staged

# 4. Machine-readable, for agents and dashboards
vibeshield scan . --format json

# 5. Preview the verified one-line fixes
vibeshield fix . --dry-run

# 6. Apply them, with every patch written to vibeshield-fixes.log
vibeshield fix . --yes
```

`vibeshield init` is deliberately conservative — `--dry-run` prints the whole
plan, nothing is overwritten without `--force`, and an existing git hook is
never touched:

```text
  $ vibeshield init --dry-run
  VibeShield 2.0.1 — project setup
  Detected   javascript, typescript
  Ecosystem  npm
  Framework  Next.js, React

  would write  vibeshield.yml
  would write  .github/workflows/vibeshield.yml
  would write  .git/hooks/pre-commit
```

| Flag | Meaning |
|:---|:---|
| `--diff <ref\|->` | Scan changes vs a git ref, or a unified diff on stdin |
| `--staged` | Scan staged changes (what the pre-commit hook runs) |
| `--format <fmt>` | `pretty` · `json` · `github` · `sarif` |
| `--mode <mode>` | `off` · `warn` · `block-on-critical` · `block-on-high+` |
| `--rules <dir>` | Load extra rule packs |
| `--config <file>` | Config path (default `vibeshield.yml`) |
| `-v`, `--verbose` | Explain what was scanned, on stderr |
| `--no-color` | Disable colour (also `NO_COLOR`, auto on non-TTY) |

**Exit codes** — `0` clean or warn-mode findings · `1` block threshold met · `2` config or usage error.

### GitHub code scanning (Security tab)

`--format sarif` emits SARIF 2.1.0, so findings show up as native code-scanning
alerts — with the rule text, the one-line fix, and a stable fingerprint that
survives reformatting:

```yaml
- run: vibeshield scan . --format sarif > vibeshield.sarif
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: vibeshield.sarif
```

---

## 🛠️ Automated CI/CD

### GitHub Action (pull-request gate)

```yaml
name: VibeShield PR Gate
on: [pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: rajviyash9136freefr-tech/vibeshield/action@v2.0.1
        with:
          mode: block-on-critical
```

Posts one consolidated **VibeCheck report** per PR — severity chips, `file:line` links, the dependency delta, and dismiss commands — updated in place on every push.

### Local pre-commit hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v2.0.1
    hooks:
      - id: vibeshield
```

---

## 📂 Repository layout

```text
├── .agents/rules/        # Antigravity workspace rules (generated)
├── .cursor/rules/        # Cursor MDC rules (generated)
├── .github/              # CI workflows + Copilot instructions
├── action/               # GitHub Action composite wrapper
├── contracts/            # Cross-component JSON schemas — the interface law
├── docs/                 # Product docs, incl. docs/agents.md
├── fixtures/             # Golden test repos & seeded AI-PR corpus
├── npm/                  # npm launcher (fetch-on-first-run)
├── rules/core/           # Versioned YAML rule packs (MIT) — the source of truth
├── scanner/              # Go scanner: CLI, console, rules engine, VibePatch
│   ├── cmd/vibeshield/   # CLI entrypoint + command dispatch
│   └── internal/
│       ├── cli/          # v2 console, search index, agent catalog
│       ├── scan/         # file walker, language detection, matching
│       ├── rules/        # pack loader (embedded core pack)
│       ├── fix/          # VibePatch planner + human gate
│       └── output/       # pretty / json / github renderers
├── scripts/              # installers + sync-rules / sync-agent-rules / bump-version
├── site/                 # Astro 5 marketing site & docs
└── skill/                # Claude Code audit skill (parallel bug-hunter subagents)
```

---

## ❓ Frequently asked questions

**What is VibeShield?**
An offline security and dependency auditor built for AI-generated code. It hunts the failure modes LLMs actually produce — invented package names, keys pasted from chat, insecure scaffolding, license stripping — and gives you a one-line fix for each. In v2 it also gives you a searchable terminal console over every action, rule and agent recipe.

**Does my code or my prompts leave my machine?**
No. Never. The CLI, console, hooks and skill run 100% locally using static analysis. `--online` is reserved for opt-in package-intel lookups and degrades gracefully offline. Your code, keys and prompts are never uploaded and never used for training.

**Which AI agents does it work with?**
Codex, Claude Code (CLI and Desktop), Google Antigravity, Cursor, Windsurf, GitHub Copilot, and anything that reads the `AGENTS.md` convention (Cline, Roo Code, Amp, Zed, Aider, Gemini CLI). See [docs/agents.md](docs/agents.md).

**How is this different from a normal dependency scanner?**
A conventional scanner asks "does this package have a known CVE?" A slopsquatted package has none — it was registered nine days ago and its only release is the attack. VibeShield flags the *pattern*: a name that scores as model-generated, paired with weak supply-chain signals, arriving in a diff.

**Does the console work over SSH, in tmux, and on Windows?**
Yes. Raw-terminal handling is per-platform `syscall` with no CGO and no third-party dependency, so the same static binary drives Windows Terminal, cmd.exe, macOS Terminal, and any POSIX terminal. If stdin is not a terminal the console steps aside and prints usage instead.

**Is it really free?**
Yes — MIT, end to end: scanner, action, hooks, console, all rule packs. No tiers, no seat limits, no tokens, no account. Unlimited private and commercial repositories.

---

## 🌟 Community

* ⭐ **Star the repo** if VibeShield caught something your linter missed.
* 🧪 **Try the live simulator** — [test real AI bug scenarios in the browser](https://rajviyash9136freefr-tech.github.io/vibeshield/).
* 🐛 **Found a new AI failure mode?** [Open an issue](https://github.com/rajviyash9136freefr-tech/vibeshield/issues) — rule packs are data, so new rules ship without a code change.
* 🤝 **Contributing** — see [CONTRIBUTING.md](CONTRIBUTING.md). `go test ./...` and `node scripts/sync-agent-rules.mjs --check` must pass.

---

## 📄 License & links

- **License**: [MIT](LICENSE) — free and open source, no tiers.
- **Website & simulator**: <https://rajviyash9136freefr-tech.github.io/vibeshield/>
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **Agent setup guide**: [docs/agents.md](docs/agents.md)

---

<!--
Search keywords. These are live as GitHub topics on this repo; keep the two in
sync if you fork it (Settings → Topics):
vibe-coding, ai-coding, ai-agents, ai-security, claude-code, codex, cursor, antigravity,
github-copilot, windsurf, agents-md, slopsquatting, hallucinated-packages,
secret-scanning, supply-chain-security, pre-commit, github-action, devsecops, sast,
static-analysis
-->
