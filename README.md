<div align="center">

# 🛡️ VibeShield

### The bug hunter for AI-generated code

**Scan a vibe-coded project for hallucinated packages, secrets pasted from chat, and insecure AI boilerplate — then fix them from the terminal. One static binary. No account. No network.**

[![Version](https://img.shields.io/badge/version-3.0.0-white?style=for-the-badge&logo=git&logoColor=black)](CHANGELOG.md)
[![Website](https://img.shields.io/badge/🌐_Website-Live_Simulator-white?style=for-the-badge&logo=googlechrome&logoColor=black)](https://rajviyash9136freefr-tech.github.io/vibeshield/)
[![GitHub Stars](https://img.shields.io/github/stars/rajviyash9136freefr-tech/vibeshield?style=for-the-badge&logo=github&color=white&labelColor=black)](https://github.com/rajviyash9136freefr-tech/vibeshield/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-white?style=for-the-badge)](LICENSE)

[![Agents](https://img.shields.io/badge/Agents-Codex_·_Claude_Code_·_Antigravity_·_Cursor_·_Windsurf_·_Copilot-111111?style=flat-square)](docs/agents.md)
[![Offline](https://img.shields.io/badge/100%25_Local_&_Offline-Zero_Data_Sent_Outside-success?style=flat-square)](#does-my-code-leave-my-machine)
[![Rules](https://img.shields.io/badge/Core_rules-122_across_6_categories-blue?style=flat-square)](rules/core)
[![Scan speed](https://img.shields.io/badge/Scan_speed-168_files_in_1.7s-blue?style=flat-square)](#how-fast-is-it)

<br/>

[⚡ **Try the live simulator**](https://rajviyash9136freefr-tech.github.io/vibeshield/#scanner) · [📦 **Install**](#install) · [🚀 **60-second quickstart**](#60-second-quickstart) · [📖 **Command reference**](#command-reference) · [📚 **Docs site**](https://rajviyash9136freefr-tech.github.io/vibeshield/docs) · [❓ **FAQ**](#frequently-asked-questions)

<br/>

```console
$ vibeshield scan .
Scanning 14 files (full mode)… done in 1.2s

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

## What is VibeShield?

AI coding agents write code fast, and they make a specific set of mistakes over and over: they **invent package names** that an attacker may already have registered, they **paste API keys** straight from the chat window, and they **scaffold insecure defaults** like wildcard CORS and `debug=True`.

Classic linters and CVE scanners miss almost all of it — an invented package has no CVE, because it did not exist yesterday.

VibeShield is a **single static Go binary** that hunts exactly those failure modes. It runs offline, in about a second, and it works in three places:

| Where | How |
|:---|:---|
| 🧑‍💻 **Your terminal** | `vibeshield scan .` — an audit you can run any time |
| 🪝 **Your commits** | `vibeshield init` writes a pre-commit hook that scans staged changes |
| 🔀 **Your pull requests** | The bundled GitHub Action comments a VibeCheck report on every PR |

**Why people use it**

- 🔒 **Nothing leaves your machine.** Static analysis only, no telemetry, no account, no upload. `--online` is opt-in and reserved.
- ⚡ **Fast enough to actually run.** 168 files in 1.7 s; a staged scan is well under 1.5 s.
- 🎯 **Narrow on purpose.** 122 rules about AI failure modes, not 5,000 general lint rules.
- 🔧 **It fixes things.** `vibeshield fix` applies the mechanical fixes and logs every patch.
- 🤖 **Speaks to your agent.** Ships rule files for Codex, Claude Code, Cursor, Windsurf, Copilot and Antigravity.
- 🆓 **MIT, end to end.** No tiers, no seat limits, unlimited private repos.

---

## Install

Pick whichever fits your setup. Every option installs the same static binary — no runtime, no dependencies, no sudo.

### 🍏 macOS & 🐧 Linux — install script

```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

### 🪟 Windows — PowerShell

```powershell
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex
```

### 📦 Zero-install — npx

```bash
npx vibeshield scan .
```

Downloads the right release binary on first run and verifies it against the release `sha256sums.txt`. If the npm package is not published to your registry yet, use the install script or `go install` below — the launcher and the binary are the same thing either way.

### 🐹 Go developers

```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@v3.0.0
```

### 🧰 Build from source

```bash
git clone https://github.com/rajviyash9136freefr-tech/vibeshield.git
cd vibeshield/scanner
go build -o vibeshield ./cmd/vibeshield
```

Pure Go, `CGO_ENABLED=0`. Cross-compiles clean for `linux`, `darwin` and `windows` on `amd64` and `arm64`.

### ⌨️ Turn on Tab completion

```bash
vibeshield completion bash   >> ~/.bashrc                      # bash
vibeshield completion zsh    >  "${fpath[1]}/_vibeshield"      # zsh
vibeshield completion fish   >  ~/.config/fish/completions/vibeshield.fish
vibeshield completion powershell | Out-String | Invoke-Expression
```

### ✅ Check it worked

```bash
vibeshield version     # binary + rule pack versions
vibeshield doctor      # is the config, the git hook and the PR gate wired up?
```

---

## 60-second quickstart

```bash
# 1. Go to a project and let VibeShield set itself up.
#    It detects your stack, writes vibeshield.yml, a PR-gate workflow and a
#    pre-commit hook, then runs the first scan. Preview it first if you like:
cd your-project
vibeshield init --dry-run
vibeshield init

# 2. From then on, audit whenever you want.
vibeshield scan .                    # everything
vibeshield scan --staged             # only what you are about to commit
vibeshield scan --diff main          # only the lines this branch added

# 3. Read the findings and let it fix the mechanical ones.
vibeshield fix . --dry-run           # preview the −/+ diff
vibeshield fix .                     # apply, with a prompt per file
```

That is the whole tool. Everything below is detail.

---

## Command reference

Run `vibeshield` with no arguments for the interactive console, or `vibeshield help <command>` for any command's full help.

| Command | What it does |
|:---|:---|
| `vibeshield init [path]` | Set a project up: config, PR gate, pre-commit hook, first scan |
| `vibeshield scan [path]` | Audit a directory, a git diff, or the staged changes |
| `vibeshield fix [path]` | Preview and apply the mechanical fixes (VibePatch) |
| `vibeshield doctor [path]` | Report what is wired up and what is not — changes nothing |
| `vibeshield search [query]` | Search every rule, action and agent recipe |
| `vibeshield rules [id]` | List the rule packs, or read one rule in full |
| `vibeshield agents [name]` | Per-agent setup recipes |
| `vibeshield completion <shell>` | Print a bash / zsh / fish / PowerShell completion script |
| `vibeshield version` | Version, rule packs, engine coverage |
| `vibeshield help [command]` | Help — also `vibeshield <command> --help` |
| `vibeshield ui` | Open the interactive console (same as a bare `vibeshield`) |

Every command takes `-h` / `--help`, and `vibeshield --version` prints the version.

### `vibeshield init` — set a project up

Detects the stack from your manifests, then writes **at most three files**:

```console
$ vibeshield init --dry-run
VibeShield 3.0.0 — project setup
Detected   javascript, typescript
Ecosystem  npm
Framework  Next.js, React

would write  vibeshield.yml
would write  .github/workflows/vibeshield.yml
would write  .git/hooks/pre-commit

--dry-run: nothing was written. Re-run without it to apply.
```

It is deliberately conservative: `--dry-run` prints the whole plan, an existing file is **never** overwritten without `--force`, and an existing git hook is never touched at all.

| Flag | Meaning |
|:---|:---|
| `--mode <mode>` | Gate mode written to the config (default `warn`) |
| `--dry-run` | Show the plan, write nothing |
| `--force` | Overwrite files that already exist |
| `--no-hook` | Skip the git pre-commit hook |
| `--no-workflow` | Skip the GitHub Action workflow |
| `--no-scan` | Skip the first scan |
| `--no-color` | Disable colour |

### `vibeshield scan` — the audit

```bash
vibeshield scan .                     # the whole project
vibeshield scan ../api                # another project, using its own config
vibeshield scan --diff main           # only findings on lines added vs main
vibeshield scan --staged              # what the pre-commit hook runs
vibeshield scan --diff -  < patch.diff   # a unified diff from stdin
vibeshield scan . --format json       # machine-readable
vibeshield scan . --format sarif      # GitHub code scanning
vibeshield scan . --mode block-on-critical   # exit 1 on a critical finding
vibeshield scan . -v                  # explain what was scanned, on stderr
```

| Flag | Meaning |
|:---|:---|
| `--diff <ref\|->` | Diff mode: scan changes vs a git ref, or `-` to read a diff from stdin |
| `--staged` | Pre-commit mode: scan git staged changes |
| `--format <fmt>` | `pretty` (default) · `json` · `github` · `sarif` |
| `--config <file>` | Config path. Default: `<path>/vibeshield.yml`, then `./vibeshield.yml` |
| `--mode <mode>` | `off` · `warn` · `block-on-critical` · `block-on-high+` |
| `--rules <dir>` | Load extra rule packs from a directory |
| `--online` | (reserved) allow package-intel network lookups |
| `--max-cols <n>` | Output width cap (default 88) |
| `--no-color` | Disable colour (also `NO_COLOR`, and automatic on a non-TTY) |
| `-v`, `--verbose` | Explain what was scanned, on stderr |

**Exit codes** — `0` clean or warn-mode findings · `1` block threshold met · `2` usage or config error.

### `vibeshield doctor` — check the setup

The first thing to run when a scan behaves unexpectedly, or in CI to assert the gate is really wired up. It reads, reports, and prints the exact command that fixes each gap. Nothing is modified.

```console
$ vibeshield doctor
VibeShield 3.0.0 — doctor

ok   binary       vibeshield 3.0.0 · pack core 2.0.0 · 117 active rules (+5 reserved)
ok   project      /home/you/api
ok   config       vibeshield.yml · mode warn
ok   git          repository detected
warn git hook     no .git/hooks/pre-commit — commits are not scanned
                  → vibeshield init
warn pr gate      .github/workflows/vibeshield.yml is missing — pull requests are not gated
                  → vibeshield init
ok   agent rules  1 of 8 present: AGENTS.md

2 thing(s) to fix.
Next: vibeshield init
```

Exit codes: `0` healthy · `1` something needs fixing · `2` usage error. Add `--format json` for the machine-readable checklist.

### `vibeshield fix` — VibePatch

Applies the one-line fixes the rules already know how to make. Opt-in and human-gated: nothing is written until you say so, and every patch is appended to `vibeshield-fixes.log` (JSONL) so the change trail stays auditable.

```bash
vibeshield fix . --dry-run            # show the −/+ diff, change nothing
vibeshield fix .                      # prompt per file: [y/N/a/s]
vibeshield fix . --yes                # apply everything, no prompts (agents / CI)
vibeshield fix . --report scan.json   # reuse a previous --format json scan
```

Two categories are **never** patched mechanically: a leaked secret needs rotation, not a rewrite, and a prompt-injection finding lives in the very instruction file an agent would be editing.

A non-interactive stdin must pass `--dry-run` or `--yes` — the gate never guesses at an answer.

### `vibeshield search` and `vibeshield rules` — explore the ruleset

```bash
vibeshield rules                      # every rule, grouped by category
vibeshield rules VS-SEC-017           # one rule, in full
vibeshield rules cors                 # rules matching a word
vibeshield search "prompt injection"  # every token must match
vibeshield search --agents cursor     # agent setup recipes only
vibeshield search --rules --list      # every shipped rule
vibeshield rules --format json        # machine-readable (vibeshield.search/v1)
```

The ranker folds separators (`vs-sec-017` ≡ `vs sec 017` ≡ `vssec017`), prefers word-boundary hits, and only falls back to fuzzy matching for queries long enough to be meaningful — so a three-letter search returns the AWS rule, not a wall of coincidences.

### `vibeshield agents` — set up your coding agent

```bash
vibeshield agents                     # every recipe
vibeshield agents cursor              # one agent
vibeshield agents --body > AGENTS.md  # just the pasteable rule block
vibeshield agents --markdown          # the support matrix as a table
```

### The interactive console

A bare `vibeshield` on a terminal opens one search box over **every action, every rule and every agent recipe**:

```console
VibeShield 3.0.0                                    140 entries · fully offline
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

`↑`/`↓` move · `Tab` jumps group · `Enter` opens · `Esc` clears then quits · `Ctrl+U` resets · `Ctrl+C` quits.

Selecting an action hands the terminal back and runs the **real** command, so the menu can never drift from the documented flags. Piping into `vibeshield` skips the console and prints help instead, so scripts are unaffected. No TTY at all? `vibeshield search` is the same index on stdout.

---

## Output formats

| Format | Use it for |
|:---|:---|
| `pretty` | Humans, on a terminal (the default) |
| `json` | Agents, dashboards, `jq`. Schema in [contracts/cli.md](contracts/cli.md) |
| `github` | `::error` / `::warning` annotations in an Actions log |
| `sarif` | SARIF 2.1.0 → GitHub code scanning (the Security tab) |

```json
{
  "schema_version": 1,
  "tool": "vibeshield",
  "version": "3.0.0",
  "scan": { "mode": "diff", "ref": "HEAD~1", "files_scanned": 14, "duration_ms": 1234 },
  "summary": { "critical": 1, "high": 1, "medium": 0, "low": 0, "info": 0, "clean_files": 11 },
  "dependencies": [
    { "name": "fast-parse-utils-v3", "version": "2.1.4", "ecosystem": "npm",
      "age_days": 9, "maintainers": 1, "risk": "critical",
      "why": "registered 9 days ago; 1 maintainer; post-install fetches remote binary" }
  ],
  "findings": [ /* contracts/finding/schema.json objects */ ]
}
```

Findings keep a stable fingerprint, so a GitHub code-scanning alert survives reformatting instead of reopening on every push.

---

## Configuration

Drop a `vibeshield.yml` at the project root, or point `--config` at one. Every key is optional.

```yaml
mode: warn                        # off | warn | block-on-critical | block-on-high+
languages: [typescript, python]   # filter; omit or leave empty for all
ignore:
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"   # REQUIRED — echoed in every report
thresholds:
  new_dependency_max_age_days: 30
notifications:
  slack: ${SLACK_WEBHOOK}         # env reference only — never a literal secret
```

**Gate modes**

| Value | PR check | Pre-commit | Findings reported? |
|:---|:---|:---|:---|
| `off` | passes | proceeds | no |
| `warn` (default) | passes | proceeds | yes |
| `block-on-critical` | fails on ≥ 1 critical | blocks those commits | yes |
| `block-on-high+` | fails on high or critical | blocks | yes |

Start on `warn`, clear the noise, then tighten. Ratcheting is a one-word edit.

**Ignore law** — every ignore needs a `reason`. It is written to the audit log and shown in the PR report, so dismissals leave a trail instead of a mystery.

**Secret law** — `notifications.slack` must be an environment reference. A literal webhook URL is a config error (exit `2`), because the path *is* the whole secret and a config file in git is a bad home for it.

**Config resolution** — an explicit `--config` always wins; otherwise the scanned project's own `vibeshield.yml` is used, so `vibeshield scan ../other-project` honours *that* project's settings rather than silently ignoring them.

---

## CI integration

### GitHub Action — the pull-request gate

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
          fetch-depth: 0        # diff mode needs the base branch
      - uses: rajviyash9136freefr-tech/vibeshield/action@v3.0.0
        with:
          mode: block-on-critical
```

The Action posts **one consolidated VibeCheck report** per PR — severity chips, `file:line` links, the dependency delta, and dismiss commands — and updates it in place on every push.

| Input | Default | Meaning |
|:---|:---|:---|
| `mode` | `warn` | `off` · `warn` · `block-on-critical` · `block-on-high+` |
| `github_token` | `${{ github.token }}` | Used to create/update the PR comment |
| `config` | `vibeshield.yml` | Config path |
| `online` | `false` | Allow package-intel lookups |
| `scanner_bin` | — | A prebuilt binary; skips the download |
| `version` | `v3.0.0` | Pinned scanner release to download |

Outputs: `critical` `high` `medium` `low` `info` `total` `blocked` `report_path` `report_json`.

### GitHub code scanning (SARIF)

```yaml
- run: vibeshield scan . --format sarif > vibeshield.sarif
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: vibeshield.sarif
```

Findings then appear as native alerts in the repository's **Security** tab, each carrying the rule text and the one-line fix.

### Pre-commit hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v3.0.0
    hooks:
      - id: vibeshield
```

Then `pre-commit install`. No pre-commit framework? `vibeshield init` writes a plain `.git/hooks/pre-commit` for you. It scans **staged changes only**, so the cost tracks your edit size, not your repository size.

A broken config never holds your repo hostage — the hook warns loudly and lets the commit through, and `git commit --no-verify` always works.

---

## AI agent setup

VibeShield ships a rule file for every major coding agent, **generated from one source of truth** and checked in CI, so the docs cannot claim support the binary does not ship.

| Agent | File VibeShield writes | Scope |
|:---|:---|:---|
| **Codex** (OpenAI) | [`AGENTS.md`](AGENTS.md) | repo root · `~/.codex/AGENTS.md` for global |
| **Claude Code** (CLI + Desktop) | [`CLAUDE.md`](CLAUDE.md) + plugin | per project / per user |
| **Google Antigravity** | [`.agents/rules/vibeshield.md`](.agents/rules/vibeshield.md) | workspace root · `~/.gemini/GEMINI.md` for global |
| **Cursor** | [`.cursor/rules/vibeshield.mdc`](.cursor/rules/vibeshield.mdc) · [`.cursorrules`](.cursorrules) | repo root, `alwaysApply: true` |
| **Windsurf** | [`.windsurf/rules/vibeshield.md`](.windsurf/rules/vibeshield.md) · [`.windsurfrules`](.windsurfrules) | repo root, applies to Cascade |
| **GitHub Copilot** | [`.github/copilot-instructions.md`](.github/copilot-instructions.md) | every Copilot chat in the repo |
| **Any `AGENTS.md` client** | `AGENTS.md` | Cline · Roo Code · Amp · Zed · Aider · Gemini CLI |

```bash
vibeshield agents --body > AGENTS.md        # paste the rule block anywhere
vibeshield agents cursor                    # the Cursor recipe
```

**Claude Code plugin** (also bundles the bug-hunting skill):

```bash
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield
claude plugin install vibeshield@vibeshield
# then, in any project:
/vibeshield:vibeshield-audit --quick
```

The skill spawns parallel bug-hunter subagents, puts every critical/high candidate in front of an adversarial skeptic, and writes `vibeshield-findings.json` in the [finding contract](contracts/finding/schema.json). It is the only surface that hunts `VS-INJ` prompt-injection traps today — malicious READMEs, issue text and `AGENTS.md` diffs that try to steer your agent.

Full matrix and install steps: [docs/agents.md](docs/agents.md).

---

## Rules

The core pack is **122 rules across 6 categories** — deliberately narrow, because every rule fires at a known way LLM-generated code breaks:

| Prefix | Category | Hunts |
|:---|:---|:---|
| `VS-PKG-###` | hallucinated-package | generated-shape names, remote-code install hooks, unpinned install paths |
| `VS-SEC-###` | hardcoded-secret | keys and tokens echoed from training data or chat |
| `VS-SEC-###` | insecure-api | `eval`, md5 passwords, SQL string concatenation, ECB mode |
| `VS-SEC-###` | insecure-default | wildcard CORS, `debug=True`, `JWT alg: none`, TLS verification off |
| `VS-DEP-###` | dependency-risk | unpinned ranges, single-maintainer additions, install hooks that `curl \| bash` |
| `VS-LIC-###` | license-missing | stripped SPDX headers, copyleft-in-permissive drift |
| `VS-INJ-###` | prompt-injection | agent-instruction payloads in READMEs, issues, CI text, tool configs |

```bash
vibeshield rules                  # read all of them
vibeshield rules VS-PKG-001       # read one
```

Rules are **data, not code** — versioned YAML loaded at startup, format in [contracts/rulepack.md](contracts/rulepack.md). Adding a rule needs no Go change.

> **Honest note on coverage.** `vibeshield version` reports `117 active · 5 reserved`. The five reserved rules load and validate so packs stay portable across versions, but the matcher cannot evaluate them yet — they depend on the package-intel model that `--online` will provide. `vibeshield rules` marks each one `RESERVED` rather than letting you search for a rule that cannot fire. This is visible in `vibeshield version`, in the console, and in CI via `scripts/audit-contract.mjs`.

### Why not just use a linter?

| AI failure mode | Why it bites | Classic linters | VibeShield |
|:---|:---|:---:|:---:|
| **Hallucinated packages** (slopsquatting) | The model invents `fast-parse-utils-v3`; an attacker registers it within hours | ❌ no CVE exists | 🚨 flagged in the diff |
| **Secrets pasted from chat** | `OPENAI_API_KEY` lands in a client bundle or a test fixture | ⚠️ noisy generics | 🔒 caught pre-commit |
| **Insecure AI boilerplate** | Wildcard CORS, `debug=True`, unhashed passwords, `JWT alg: none`, `eval()` | ⚠️ buried in lint noise | ⚡ flagged with a one-line fix |
| **License stripping** | A model rewrites vendored code and drops the SPDX header | ❌ no coverage | 📄 flagged (`VS-LIC`) |
| **Dependency drift** | Unpinned ranges, install hooks that `curl \| bash` | ⚠️ partial | 📦 flagged (`VS-DEP`) |
| **Prompt-injection traps** | A README or `AGENTS.md` diff tells your agent to open a backdoor | ❌ zero coverage | 🛡️ hunted by the agent skill (`VS-INJ`) |

A conventional scanner asks *"does this package have a known CVE?"*. A slopsquatted package has none — it was registered nine days ago and its only release is the attack. VibeShield flags the *pattern*.

---

## Frequently asked questions

**Does my code leave my machine?**
No. Never. The CLI, console, hooks and skill run 100% locally using static analysis. `--online` is reserved for opt-in package-intel lookups and degrades gracefully offline. Your code, keys and prompts are never uploaded and never used for training.

**How fast is it?**
168 files in 1.7 s. A staged pre-commit scan is under 1.5 s on a typical diff; the Action's median PR scan is under 10 s. Binary boot is under 50 ms.

**Which AI agents does it work with?**
Codex, Claude Code (CLI and Desktop), Google Antigravity, Cursor, Windsurf, GitHub Copilot, and anything reading the `AGENTS.md` convention (Cline, Roo Code, Amp, Zed, Aider, Gemini CLI). See [docs/agents.md](docs/agents.md).

**Does the console work over SSH, in tmux, and on Windows?**
Yes. Raw-terminal handling is per-platform `syscall` with no CGO and no third-party dependency, so the same static binary drives Windows Terminal, `cmd.exe`, macOS Terminal and any POSIX terminal. If stdin is not a terminal the console steps aside and prints help.

**Is it really free?**
Yes — MIT, end to end: scanner, Action, hooks, console, all rule packs. No tiers, no seat limits, no tokens, no account. Unlimited private and commercial repositories.

**Something is not working. Where do I look?**
`vibeshield doctor`. It reports the binary, the config, the git hook, the PR gate and your agent rule files, and prints the command that fixes each gap.

---

## Repository layout

```text
├── .agents/rules/        # Antigravity workspace rules (generated)
├── .cursor/rules/        # Cursor MDC rules (generated)
├── .github/              # CI workflows, issue & PR templates, Copilot instructions
├── action/               # GitHub Action composite wrapper + entrypoint
├── contracts/            # Cross-component JSON schemas — the interface law
├── docs/                 # Product docs: agents.md, growth.md
├── fixtures/             # Golden test repos & seeded AI-PR corpus
├── npm/                  # npm launcher (fetch-on-first-run, checksum-verified)
├── packaging/            # Homebrew formula generator (see packaging/homebrew/)
├── rules/core/           # Versioned YAML rule packs (MIT) — the source of truth
├── scanner/              # Go scanner: CLI, console, rules engine, VibePatch
│   ├── cmd/vibeshield/   # CLI entrypoint, help text, doctor, completions
│   └── internal/
│       ├── cli/          # console, search index, agent catalog
│       ├── scan/         # file walker, language detection, matching
│       ├── rules/        # pack loader (embedded core pack)
│       ├── fix/          # VibePatch planner + human gate
│       └── output/       # pretty / json / github / sarif renderers
├── scripts/              # installers + sync-rules / sync-agent-rules / bump-version
├── site/                 # Astro 5 marketing site & docs
└── skill/                # Claude Code audit skill (parallel bug-hunter subagents)
```

---

## Contributing

Contributions are welcome — especially **new rules**, which are data and need no Go change.

```bash
git clone https://github.com/rajviyash9136freefr-tech/vibeshield.git
cd vibeshield/scanner && go test ./...
```

Before you open a PR, three gates must pass:

```bash
cd scanner && go test ./... && go vet ./... && gofmt -l .
node scripts/sync-rules.mjs --check          # rules/core ↔ embedded pack
node scripts/sync-agent-rules.mjs --check    # generated agent rule files
node scripts/audit-contract.mjs scanner/vibeshield   # docs ↔ binary
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the rule-authoring guide and the finding contract, and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for how we treat each other. First time here? Issues labelled [`good first issue`](https://github.com/rajviyash9136freefr-tech/vibeshield/labels/good%20first%20issue) are a good entry point.

---

## Community

- ⭐ **Star the repo** if VibeShield caught something your linter missed — it is how other people find it.
- 🧪 **Try the live simulator** — [test real AI bug scenarios in the browser](https://rajviyash9136freefr-tech.github.io/vibeshield/).
- 🐛 **Found a new AI failure mode?** [Open an issue](https://github.com/rajviyash9136freefr-tech/vibeshield/issues) — rule packs are data, so new rules ship without a code change.
- 💬 **Questions and ideas** — [Discussions](https://github.com/rajviyash9136freefr-tech/vibeshield/discussions).

---

## License & links

- **License**: [MIT](LICENSE) — free and open source, no tiers.
- **Website & simulator**: <https://rajviyash9136freefr-tech.github.io/vibeshield/>
- **Documentation**: <https://rajviyash9136freefr-tech.github.io/vibeshield/docs>
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **Agent setup guide**: [docs/agents.md](docs/agents.md)
- **Security policy & threat model**: [SECURITY.md](SECURITY.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md) · [Code of conduct](CODE_OF_CONDUCT.md)
- **Project growth playbook**: [docs/growth.md](docs/growth.md)

---

<!--
Search keywords. These are live as GitHub topics on this repo; keep the two in
sync if you fork it (Settings → Topics):
vibe-coding, ai-coding, ai-agents, ai-security, claude-code, codex, cursor, antigravity,
github-copilot, windsurf, agents-md, slopsquatting, hallucinated-packages,
secret-scanning, supply-chain-security, pre-commit, github-action, devsecops, sast,
static-analysis, cli, golang, developer-tools, security-tools
-->
