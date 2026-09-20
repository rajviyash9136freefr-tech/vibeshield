# VibeShield CLI — Product Analysis & Architecture Plan

> **Document Version:** 1.0.0  
> **Target Release:** VibeShield CLI v3.1.0 → v4.0.0  
> **Status:** Draft for Review & Approval  
> **Primary Author:** Senior Developer Tools Engineering Team  

---

## 1. Executive Summary

VibeShield today operates as a fast, air-gapped Go 1.24 static analysis binary (`v3.0.0`) alongside an Astro 5 web simulator, pre-commit hook, and GitHub Action. This document plans VibeShield's evolution into a **first-class, interactive terminal product** inspired by Claude Code: developers install a single binary, type `vibeshield`, and immediately enter an interactive, keyboard-driven terminal application that audits their vibe-coded repository, streams findings in real time, inspects syntax-highlighted code snippets, and applies verified atomic diff fixes. 

The terminal UI is built natively in **Go using Charm's Bubble Tea, Lip Gloss, and Bubbles ecosystem**, preserving VibeShield’s zero-dependency, sub-15ms startup, cross-platform (Windows, macOS, Linux) guarantee without Node or Python runtime friction. The existing deterministic scanner core and 122 rules remain 100% offline and air-gapped by default, with an optional Bring-Your-Own-Key (BYOK) layer for deep AI-driven explanations and complex multi-line remediation, safeguarded by automated secret redaction and prompt-injection defenses.

---

## 2. Codebase & Repository Analysis Findings

A comprehensive audit of the workspace files and repository contracts reveals a solid, highly disciplined foundation:

### 2.1 Project Structure & Monorepo Layout
The repository is structured as a clean multi-component workspace:
- **`scanner/`** (`scanner/cmd/vibeshield/main.go`): Pure Go static binary. Houses the CLI entrypoint, argument dispatch, command routing, help generator, shell completion, doctor diagnostics, and internal engine packages.
- **`rules/`** (`rules/core/`): 6 core YAML rule packs (`dependencies.yaml`, `license.yaml`, `packages.yaml`, `security-api.yaml`, `security-defaults.yaml`, `security-secrets.yaml`). Embedded directly into the compiled binary via `scanner/internal/rules/embed.go` using `//go:embed packs/*.yaml`.
- **`npm/vibeshield/`** (`npm/vibeshield/package.json`): Zero-dependency npm launcher (`bin/vibeshield.js`, `lib/run.js`). Does not compile or bundle code; fetches the platform-specific release binary (`vibeshield-{os}-{arch}.tar.gz` / `.zip`) from GitHub Releases on first execution, verifies SHA-256 checksums from `sha256sums.txt`, and caches it in `~/.cache/vibeshield/` or `%LOCALAPPDATA%\vibeshield\`.
- **`site/`** (`site/src/pages/index.astro`): Astro 5 marketing site and documentation hosted on GitHub Pages (`/vibeshield/`). Features client-side code simulator (`QuickScanner.astro`) illustrating threat detection scenarios.
- **`contracts/`** (`contracts/cli.md`, `contracts/finding/schema.json`): Formal interface laws defining command behavior, exit codes, finding schemas, rule pack syntax, and future API endpoints.
- **`fixtures/golden/`** (`fixtures/golden/`): 15 golden test repositories (`cors-star`, `fake-keys`, `flask-debug`, `hallucinated-package`, `postinstall-package`, etc.) verifying true positives and clean negatives.
- **`action/`** (`action/action.yml`): Composite GitHub Action wrapper executing diff scans and emitting PR comments/annotations.
- **`scripts/`** (`scripts/audit-contract.mjs`): Strict test and sync gates (`audit-contract.mjs`, `sync-rules.mjs`, `sync-agent-rules.mjs`, `bump-version.mjs`).

### 2.2 Tech Stack & Tooling
- **Language:** Go 1.24+ (pure Go, strictly `CGO_ENABLED=0`), Node 18+ (ESM scripts only).
- **Dependencies:** Minimalist Go module (`gopkg.in/yaml.v3` for rule pack parsing). Zero network client or telemetry libraries linked into the binary.
- **OS Support:** First-class parity across Windows (`x86_64`, `arm64`), macOS (`x86_64`, `arm64` Apple Silicon), and Linux (`x86_64`, `arm64`).

### 2.3 Scanner Core
- **Location:** `scanner/internal/scan/scan.go`.
- **Invocation:** `scan.Dir(root, pack, version, opts)` traverses directories via `filepath.WalkDir`.
- **Inputs:** Local directory path, git ref comparison (`--diff <ref>`), staged git index (`--staged`), or unified diff from standard input (`--diff -`).
- **Filtering & Performance:** Automatically skips ignore directories (`skipDirs`: `.git`, `node_modules`, `dist`, `build`, `.venv`, `.next`, `.astro`, etc.) and files exceeding 1 MiB (`MaxFileBytes`). Scans a typical 200-file project in ~1.5s and 30-file fixtures in <20ms.
- **Outputs:** `scan.Report` struct with `ScanInfo`, `Summary` (severity counts), and `Findings []*finding.Finding`.
- **Decoupling:** Pure Go library; 100% decoupled from the Astro web front-end.

### 2.4 AI Integration (Current State vs. Vision)
- **Current State:** The compiled scanner binary contains **no AI model integration or external LLM calls**. Detection is deterministic static analysis via RE2 regular expressions and AST patterns. AI awareness exists strictly as external agent prompt files (`AGENTS.md`, `.cursorrules`, `CLAUDE.md`, `.windsurfrules`) and a Claude Code plugin in `skill/`.
- **Gap for CLI:** To deliver a Claude Code-like experience, the CLI must optionally provide natural-language vulnerability explanations and context-aware code refactoring (beyond simple single-line regex rewrites).

### 2.5 Rules & Detections
- **Active Rules:** 117 active rules + 5 reserved structural rules across 7 categories:
  1. `hallucinated-package` (`VS-PKG-###`): Non-existent or freshly registered npm/PyPI packages.
  2. `hardcoded-secret` (`VS-SEC-###`): Leaked keys (OpenAI, Anthropic, AWS, Stripe, GitHub tokens, Slack webhooks).
  3. `insecure-default` (`VS-SEC-###`): Flask `debug=True`, wildcard CORS (`*`), disabled TLS verification, unhashed passwords.
  4. `insecure-api` (`VS-SEC-###`): `eval()` on untrusted input, SQL string concatenation, unverified remote script execution.
  5. `dependency-risk` (`VS-DEP-###` / `VS-PKG-###`): Untrusted postinstall scripts, wildcard versions (`*`).
  6. `license-missing` (`VS-LIC-###`): Stripped copyleft (GPL) code pasted without attribution.
  7. `prompt-injection` (`VS-INJ-###`): Instructions attempting to trick coding agents.

### 2.6 Auth, Accounts, Plans, & Limits
- **Current Model:** Free and open source (MIT license end-to-end). No accounts, logins, telemetry, subscriptions, or paywalls exist in the codebase (`docs/site-facts.md`).
- **Unimplemented Design:** `contracts/api.md` drafts an unreleased `api.vibeshield.dev` specification for package intelligence caching and PR comment ingestion.

### 2.7 Reusability & Standalone Extraction
The existing packages in `scanner/internal/` are exceptionally modular:
- `rules.LoadCore()`: Readily provides all embedded rules.
- `scan.Dir()` / `scandiff`: Standalone scanning and diff filtering.
- `fix.BuildPlan()` / `fix.Apply()`: Clean mechanical patch engine (`VibePatch`).
- `finding.*`: Robust data model with built-in secret redaction (`finding.go`).

---

## 3. Gaps, Risks, and Required Refactors

Before implementing a full-scale interactive TUI, several architectural constraints must be resolved:

| Area | Current State | Target State for Interactive CLI | Required Refactor |
|---|---|---|---|
| **Console UI** | `scanner/internal/cli/console.go` is an ad-hoc search menu over help items, not a persistent interactive app. | Rich, multi-view TUI with tabs, real-time streaming, diff viewer, and slash commands. | Replace `console.go` with a modular Charm Bubble Tea architecture. |
| **Scan Execution** | `scan.Dir` is completely synchronous and blocking; returns only when the entire tree walk completes. | Streaming progress events: emits `FileScanned`, `FindingDiscovered`, `ProgressTick` over Go channels. | Refactor `scan.Dir` into a channel-based `scan.StreamDir(ctx, opts) <-chan ScanEvent`. |
| **Fix Engine** | `fix.go` only supports single-line regex rewrites (`VibePatch`). Multi-line fixes or non-regex suggestions fail. | Multi-line unified diff patches, structural AST rewrites, and optional AI-generated fixes. | Extend `fix.Plan` and `FilePatch` to support multi-line hunk replacements and interactive diff confirmation. |
| **AI Capabilities** | None in binary; `--online` is reserved and prints a notice. | Optional BYOK (Anthropic, OpenAI, Gemini, Ollama) for `/explain` and complex `/fix`. | Create an isolated `internal/ai` package that is strictly opt-in, client-side only, and never activates without user consent. |
| **Secret Redaction** | Redacts matched regex spans for `hardcoded-secret` findings. | Redact secrets across entire files before any context is passed to the TUI renderer or BYOK AI prompts. | Centralize automated redaction in `internal/finding/redact.go` with entropy-based secret masking. |
| **Windows TTY Quirks** | Custom raw-mode handling in `term_windows.go`. | Robust cross-platform terminal control (conhost, Windows Terminal, PowerShell, Git Bash). | Bubble Tea natively leverages `golang.org/x/term` and Windows Virtual Terminal Processing. |

---

## 4. CLI Product Specification

### 4.1 Naming & Installation Experience

- **Binary Name:** `vibeshield` (executable: `vibeshield` on POSIX, `vibeshield.exe` on Windows).
- **Distribution Strategy (Ranked by Priority):**
  1. **NPM / NPX Launcher (`npx vibeshield`):** Primary zero-friction path for web & vibe developers. The existing `npm/vibeshield` package already downloads and checksum-verifies the native binary.
  2. **Shell One-Liners (`install.sh` / `install.ps1`):** Direct curl/irm installers for standalone terminal users without Node installed.
  3. **Homebrew Tap:** `brew install rajviyash9136freefr-tech/tap/vibeshield`.
  4. **Scoop / Winget:** Windows native package management.
  5. **Go Install:** `go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@v3.0.0`.

### 4.2 Two Operating Modes

```
                               ┌───────────────────────────┐
                               │     vibeshield [args]     │
                               └─────────────┬─────────────┘
                                             │
                       ┌─────────────────────┴─────────────────────┐
                       ▼                                           ▼
             [ TTY && No Subcommand ]                    [ Subcommand or CI Pipe ]
                       │                                           │
                       ▼                                           ▼
         ┌───────────────────────────┐               ┌───────────────────────────┐
         │     INTERACTIVE MODE      │               │   NON-INTERACTIVE / CI    │
         │  • Claude Code Experience │               │  • `vibeshield scan .`    │
         │  • Bubble Tea TUI         │               │  • `--format json/sarif`  │
         │  • Live Streaming Scanner │               │  • `--staged` Pre-commit  │
         │  • Slash Commands         │               │  • Strict Exit Codes      │
         │  • Interactive Diff Fixes │               │  • Machine Parsable       │
         └───────────────────────────┘               └───────────────────────────┘
```

#### Mode 1: Interactive Mode (Run `vibeshield` with no arguments in a terminal)
- Launches an interactive, full-screen terminal application (Claude Code aesthetic).
- Detects the project stack (`package.json`, `requirements.txt`, `go.mod`, Astro, Supabase, Next.js).
- Prompts the user with a clean input bar supporting slash commands:
  - `/scan [path]` — Run or re-run a security audit.
  - `/diff [ref]` — Audit only modified/uncommitted code.
  - `/fix [finding-id]` — Interactively review and apply atomic diff patches.
  - `/explain <finding-id>` — Deep-dive security analysis and attack scenario.
  - `/doctor` — Run environment, git hook, and rule health checks.
  - `/baseline` — Snapshot existing findings to ignore legacy issues and only flag regressions.
  - `/rules [query]` — Fuzzy-search active security rules.
  - `/config` — Inspect or update `vibeshield.yml`.
  - `/help` — Command guide and keyboard shortcuts.
  - `/exit` or `Ctrl+C` — Gracefully exit back to shell.

#### Mode 2: Non-Interactive / CI Mode (`vibeshield scan`, `vibeshield fix`, etc.)
- Headless, ultra-fast execution for CI/CD pipelines, git hooks, and background automation.
- Output flags: `--format pretty` (default ANSI), `--format json`, `--format sarif` (GitHub Code Scanning), `--format github` (`::error::` workflow annotations).
- Gating: `--mode block-on-critical` (exit 1 on critical), `--mode block-on-high+` (exit 1 on critical or high), `--mode warn` (exit 0, report only).
- Diff flags: `--staged` (git staged changes), `--diff <branch|commit>`, `--diff -` (read unified diff from stdin).

### 4.3 Command Reference Matrix

| Command | Arguments / Flags | Description |
|---|---|---|
| `vibeshield` | `[path]` | Launch interactive terminal experience (defaults to `.`). |
| `vibeshield scan` | `[path] [--diff <ref>\|--staged] [--format pretty\|json\|sarif\|github] [--mode <mode>]` | Perform standalone security scan and exit with code. |
| `vibeshield fix` | `[path] [--dry-run] [--yes] [--report <file>]` | Review and apply mechanical or AI patches. |
| `vibeshield explain` | `<rule-or-finding-id> [--byok]` | Provide deep context and risk breakdown for a finding. |
| `vibeshield init` | `[path] [--mode <mode>] [--force] [--dry-run]` | Onboard project: scaffold `vibeshield.yml`, git hook, and GitHub Action. |
| `vibeshield doctor` | `[path] [--format pretty\|json] [-v]` | Diagnostic check of binary, rule packs, config, and git hooks. |
| `vibeshield baseline` | `[path] [--output <file>]` | Record current findings into `.vibeshield-baseline.json`. |
| `vibeshield watch` | `[path] [--debounce <ms>]` | Daemon mode watching filesystem and re-auditing on file save. |
| `vibeshield config` | `[get\|set\|list] [key] [value]` | Manage project and global CLI preferences. |
| `vibeshield auth` | `[login\|logout\|status]` | Configure optional local BYOK credentials securely in OS keychain. |
| `vibeshield update` | `[--check]` | Check for newer binary release and self-update. |

---

### 4.4 Step-by-Step Core User Journeys

#### Journey 1: First Install & Interactive Welcome
```
$ npm i -g vibeshield
$ cd my-vibe-app
$ vibeshield

  ┌────────────────────────────────────────────────────────────────────────┐
  │  🛡️  VIBESHIELD  v3.1.0 — AI Code Security Scanner                     │
  │  Local-first • Air-gapped • Zero telemetry • MIT License               │
  └────────────────────────────────────────────────────────────────────────┘

  Project:    my-vibe-app (git: main)
  Detected:   TypeScript 5.4, Next.js 14, Supabase Auth
  Config:     Using default rules (no vibeshield.yml found)
  Engine:     117 active security rules loaded

  Ready. Type a slash command or press Enter to scan.

  > /scan .                                                    [Tab: Commands]
```

#### Journey 2: Interactive Scan with Live Streaming Counter
```
  Scanning repository: my-vibe-app
  [████████████████████░░░░░░░░░░░░] 64% • 142/220 files (0.8s)
  Current: src/app/api/auth/route.ts

  Live Findings:
    🔴 CRITICAL   1   (VS-SEC-001: hardcoded-secret)
    🟠 HIGH       2   (VS-SEC-014: insecure-default, VS-PKG-001: hallucinated-pkg)
    🟡 MEDIUM     1   (VS-SEC-036: sql-concat)

  [Esc] Cancel scan
```

#### Journey 3: Interactive Findings Review & Atomic Fix
```
  Scan Complete: 220 files scanned in 1.2s • 4 findings found

  SEV       RULE ID     CATEGORY          FILE:LINE            TITLE
  ─────────────────────────────────────────────────────────────────────────────
❯ 🔴 CRIT   VS-SEC-001  hardcoded-secret  src/lib/ai.ts:14     OpenAI API Key
  🟠 HIGH   VS-SEC-014  insecure-default  server.py:42         Flask debug=True
  🟠 HIGH   VS-PKG-001  hallucinated-pkg  package.json:18      fast-parse-utils-v3
  🟡 MED    VS-SEC-036  insecure-api      src/db/users.ts:28   SQL string concat

  [↑/↓] Navigate  [Enter] Inspect  [f] Apply Fix  [e] Explain  [q] Quit

  ─────────────────────────────────────────────────────────────────────────────
  Finding: VS-SEC-001 — Hardcoded OpenAI API Key
  File:    src/lib/ai.ts:14:21

   12 |  // Client initialization
   13 |  import OpenAI from 'openai';
   14 |  const apiKey = "sk-proj-9A82B••••••••••••••••••••••••••••••••3F9A";
   15 |  export const openai = new OpenAI({ apiKey });

  Why this matters:
  Inline API keys copied from AI chat responses are committed to version control
  and visible to anyone with repository read access.

  Suggested Fix: Read credential from environment variable.

  Proposed Diff:
  src/lib/ai.ts
  - const apiKey = "sk-proj-9A82B••••••••••••••••••••••••••••••••3F9A";
  + const apiKey = process.env.OPENAI_API_KEY;

  [y] Apply patch now    [n] Skip    [a] Apply all    [d] Open in VS Code
  Choice (y/n/a/d): y
  ✓ Applied patch to src/lib/ai.ts (audit logged to vibeshield-fixes.log)
```

#### Journey 4: Non-Interactive Pre-Commit Hook & CI Gate
```
$ git commit -m "feat: add user parser"
[pre-commit] Running VibeShield diff audit...

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     package.json:18 — "fast-parse-utils-v3": "^2.1.4"
     → Fix: Replace with standard node:util or vetted package

vibeshield: findings at or above threshold — blocked (mode block-on-critical)
Commit aborted. Run `vibeshield fix` to resolve.
```

---

## 5. Terminal UI / UX Design

### 5.1 Tech Stack Comparison & Final Recommendation

| Framework / Stack | Language | Startup Latency | Binary Footprint | Cross-Platform Parity | Direct Core Reuse | Dev Friction | Verdict |
|---|---|---|---|---|---|---|---|
| **Bubble Tea + Lip Gloss** (Charm) | Go | **< 15ms** | Single static **~14 MB** | **Flawless** (Windows Terminal, Git Bash, macOS, Linux) | **100% Native** (Imports `scanner/internal/*` directly) | Low (Single language) | **RECOMMENDED (WINNER)** |
| **Ink (React for CLI)** | TypeScript / Node | ~250–400ms | Requires Node runtime or ~80MB SEA bundle | Good, but CMD/PowerShell ANSI color glitches common | Poor (Requires IPC/JSON bridge to Go binary) | High (Two ecosystems) | Rejected |
| **Textual / Rich** | Python | ~300–500ms | Requires Python 3.10+ or PyInstaller bloat | Moderate | Poor (Requires Python/Go C-shared bindings) | High | Rejected |
| **Ratatui** | Rust | < 10ms | ~10 MB | Flawless | Zero (Would require rewriting Go scanner in Rust) | Extreme | Rejected |

**Decision:** **Go with Bubble Tea (`github.com/charmbracelet/bubbletea`), Lip Gloss (`github.com/charmbracelet/lipgloss`), and Bubbles (`github.com/charmbracelet/bubbles`)**.
- **Reasoning:** VibeShield’s scanner is already written in clean, robust Go 1.24. Building the TUI in Bubble Tea allows direct in-process access to `scan.Dir`, `rules.Pack`, and `fix.Apply` without IPC serialization, child processes, or foreign runtime dependencies. It compiles into a single, self-contained executable that starts instantly on all operating systems.

---

### 5.2 ASCII Wireframes of Key Terminal Screens

#### Screen 1: Welcome & Command Hub
```
╭──────────────────────────────────────────────────────────────────────────────╮
│  🛡️  VIBESHIELD  v3.1.0 — AI Code Security Scanner                           │
│  The Bug Hunter for AI-Generated & Vibe-Coded Software                       │
╰──────────────────────────────────────────────────────────────────────────────╯
   Project  › my-awesome-project  (Git: feature/auth-fix*)
   Stack    › TypeScript 5.4 • React 18 • Supabase JS • Express
   Engine   › 117 active rules loaded • Local & Air-gapped (0 telemetry)

 ┌─ Quick Actions ─────────────────────────────────────────────────────────────┐
 │  [s]  Full Scan (audit entire repo)      [d]  Diff Scan (git changes only)   │
 │  [f]  Fix Wizard (review pending diffs)  [h]  Health Doctor Check           │
 └─────────────────────────────────────────────────────────────────────────────┘

 › Type /scan, /fix, /explain, or press Enter to scan current workspace...
 ───────────────────────────────────────────────────────────────────────────────
 [Tab] Command Autocomplete    [Ctrl+C] Quit                        [?] Help
```

#### Screen 2: Real-Time Streaming Scan Screen
```
╭─ AUDITING WORKSPACE ─────────────────────────────────────────────────────────╮
│  Target: c:/Projects/my-awesome-project                                      │
╰──────────────────────────────────────────────────────────────────────────────╯

  Scanning Files:
  [████████████████████████████░░░░░░░░░░░░] 72% (158 / 220 files)
  Time: 0.9s • Speed: 175 files/sec • Current: src/controllers/billing.ts

 ╭─ Live Telemetry ────────────────────────────────────────────────────────────╮
 │  🔴  1 Critical   VS-SEC-001: OpenAI API Key in src/lib/ai.ts:14            │
 │  🟠  2 High       VS-PKG-001: Slopsquatted pkg in package.json:8           │
 │  🟡  1 Medium     VS-SEC-014: Flask debug=True in tests/fixtures/app.py:12  │
 │  🟢  154 Clean files                                                        │
 ╰─────────────────────────────────────────────────────────────────────────────╯

 [Esc] Abort Scan
```

#### Screen 3: Findings Explorer (List View)
```
╭─ AUDIT RESULTS ──────────────────────────────────── 4 Findings (1.2s) ───────╮
│ Filter: [All]  Critical (1)  High (2)  Medium (1)     Sort: Severity (High→Low)│
╰──────────────────────────────────────────────────────────────────────────────╯
   SEV     ID          CATEGORY          LOCATION             TITLE
 ───────────────────────────────────────────────────────────────────────────────
❯ 🔴 CRIT  VS-SEC-001  hardcoded-secret  src/lib/ai.ts:14     Leaked OpenAI Key
  🟠 HIGH  VS-PKG-001  hallucinated-pkg  package.json:8       fast-parse-utils-v3
  🟠 HIGH  VS-SEC-014  insecure-default  server/app.py:42     Flask debug=True
  🟡 MED   VS-SEC-036  insecure-api      src/db/users.ts:89   SQL String Concat

 ┌─ Quick Preview: VS-SEC-001 ─────────────────────────────────────────────────┐
 │ File: src/lib/ai.ts:14                                                      │
 │ Snippet: const apiKey = "sk-proj-9A82B••••••••••••••••••••••••3F9A";        │
 │ Suggested Fix: process.env.OPENAI_API_KEY                                   │
 └─────────────────────────────────────────────────────────────────────────────┘
 [↑/↓] Select  [Enter] Detailed View  [f] Auto-Fix  [e] AI Explain  [/] Filter
```

#### Screen 4: Single Finding Detail & Fix Preview
```
╭─ FINDING DETAIL: VS-SEC-001 ─────────────────────────────────────────────────╮
│ 🔴 CRITICAL • hardcoded-secret • Confidence: 95%                             │
╰──────────────────────────────────────────────────────────────────────────────╯
  File: src/lib/ai.ts:14:19
  Rule: OpenAI API key detected in source code

 ┌─ Source Context ────────────────────────────────────────────────────────────┐
 │  12 │ import OpenAI from 'openai';                                          │
 │  13 │                                                                       │
 │  14 │ const client = new OpenAI({ apiKey: "sk-proj-9A82••••••••••••••" });  │
 │  15 │                                                                       │
 │  16 │ export async function generateText(prompt: string) {                  │
 └─────────────────────────────────────────────────────────────────────────────┘

 ┌─ Risk Explanation ──────────────────────────────────────────────────────────┐
 │ Inline API keys are frequently committed into public git repositories after │
 │ being copied directly from ChatGPT or Claude code generation sessions.      │
 └─────────────────────────────────────────────────────────────────────────────┘

 ┌─ Atomic Diff Patch (VibePatch) ─────────────────────────────────────────────┐
 │ src/lib/ai.ts                                                               │
 │ - const client = new OpenAI({ apiKey: "sk-proj-9A82••••••••••••••" });      │
 │ + const client = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });        │
 └─────────────────────────────────────────────────────────────────────────────┘

 [y] Apply Patch    [o] Open File in Editor    [b] Add to Baseline    [Esc] Back
```

---

### 5.3 Design System & Terminal Styling Tokens

Adhering strictly to `CLAUDE.md` §8 and `contracts/cli.md`:

- **Severity Color Palette (24-bit TrueColor with ANSI-16 fallbacks):**
  - **CRITICAL:** `#F87171` (ANSI Red, Bold) • Icon: `🔴` (ASCII: `[CRIT]`)
  - **HIGH:** `#FB923C` (ANSI Bright Yellow/Orange) • Icon: `🟠` (ASCII: `[HIGH]`)
  - **MEDIUM:** `#FBBF24` (ANSI Yellow) • Icon: `🟡` (ASCII: `[MED]`)
  - **LOW:** `#60A5FA` (ANSI Blue) • Icon: `🔵` (ASCII: `[LOW]`)
  - **INFO:** `#71717A` (ANSI Gray/Muted) • Icon: `⚪` (ASCII: `[INFO]`)
  - **CLEAN / SUCCESS:** `#34D399` (ANSI Green) • Icon: `✓` (ASCII: `[OK]`)
  - **BRAND ACCENT:** `#818CF8` (ANSI Magenta/Cyan) • Used for borders and headers.

- **Compatibility Fallbacks:**
  - **`NO_COLOR` Environment Variable:** When detected or stdout is non-TTY, color escape codes are completely stripped.
  - **Windows CMD / Legacy Consoles:** Falls back to 16 ANSI colors and pure ASCII box drawing (`+--+`, `|`) instead of Unicode rounded corners (`╭─╮`, `╰─╯`).

---

## 6. Architecture & Implementation Design

### 6.1 Monorepo Package Layout
The TUI lives directly inside `scanner/` to share types and compile into the single binary:

```
vibeshield/
├── scanner/
│   ├── cmd/
│   │   └── vibeshield/
│   │       ├── main.go               # Top-level dispatcher (TTY vs Non-TTY router)
│   │       ├── doctor.go             # Diagnostics
│   │       ├── completion.go         # Shell completions
│   │       └── help.go               # Contract help definitions
│   ├── internal/
│   │   ├── app/                      # Non-interactive CLI command handlers
│   │   │   ├── scan.go
│   │   │   ├── fix.go
│   │   │   └── doctor.go
│   │   ├── tui/                      # Bubble Tea Interactive Terminal Application
│   │   │   ├── app.go                # Main Tea Model & event loop
│   │   │   ├── router.go             # Screen routing (home, scan, list, detail, diff)
│   │   │   ├── styles/               # Lip Gloss style definitions & design tokens
│   │   │   ├── views/                # Individual UI screens
│   │   │   │   ├── welcome.go        # Banner & prompt view
│   │   │   │   ├── scanning.go       # Live progress & spinner view
│   │   │   │   ├── findings.go       # List explorer & filter table
│   │   │   │   ├── detail.go         # Single finding deep dive
│   │   │   │   ├── diff.go           # Side-by-side / unified patch viewer
│   │   │   │   └── doctor.go         # System health screen
│   │   │   └── components/           # Reusable UI widgets
│   │   │       ├── prompt.go         # Slash command input bar
│   │   │       ├── progress.go       # Smooth progress bar
│   │   │       └── codeview.go       # Syntax-highlighted code viewport
│   │   ├── scan/
│   │   │   ├── scan.go               # Synchronous scan
│   │   │   └── stream.go             # NEW: Event-driven streaming scan channels
│   │   ├── ai/                       # NEW: BYOK AI remediation & explanation layer
│   │   │   ├── provider.go           # Provider interface (Anthropic, OpenAI, Gemini, Ollama)
│   │   │   ├── client.go             # HTTP client with strict timeouts & redaction
│   │   │   └── prompts.go            # Untrusted input sandboxed prompts
│   │   ├── baseline/                 # NEW: Baseline snapshotting engine
│   │   │   └── baseline.go           # .vibeshield-baseline.json serializer/matcher
│   │   ├── fix/                      # VibePatch mechanical patch engine
│   │   ├── rules/                    # Embedded rule packs
│   │   ├── finding/                  # Finding data models & automated redaction
│   │   └── config/                   # vibeshield.yml loader
```

### 6.2 Streaming Scan Design
`scanner/internal/scan/stream.go` decouples filesystem discovery from UI rendering:

```go
type ScanEventKind int

const (
    EventFileWalked ScanEventKind = iota
    EventFindingDiscovered
    EventScanCompleted
    EventScanFailed
)

type ScanEvent struct {
    Kind        ScanEventKind
    TotalFiles  int
    FilesDone   int
    CurrentFile string
    Finding     *finding.Finding
    Report      *scan.Report
    Error       error
}

// StreamDir starts an asynchronous walk and pushes events to the channel.
func StreamDir(ctx context.Context, root string, pack *rules.Pack, opts Options) <-chan ScanEvent {
    out := make(chan ScanEvent, 64)
    go func() {
        defer close(out)
        // Traverses directory, emitting EventFileWalked and EventFindingDiscovered
    }()
    return out
}
```

### 6.3 Bring-Your-Own-Key (BYOK) AI Architecture
To preserve VibeShield’s strict privacy contract (`CLAUDE.md` § hard rule #2: *"No code leaves the machine"*):
- **100% Opt-In:** The AI module is completely dormant unless the user explicitly runs `/explain --byok` or configures an API key via `vibeshield auth login`.
- **Supported Providers:** Anthropic Claude (default for coding intelligence), OpenAI, Google Gemini, and local Ollama (`http://localhost:11434` for 100% offline air-gapped AI).
- **Automated Secret Redaction:** Code snippets undergo multi-stage redaction before being sent to an external provider:
  - Regex secret matching (`VS-SEC-###`).
  - High-entropy string masking.
  - Line bounds: Only the minimal relevant context (maximum 15 lines surrounding the finding) is transmitted.
- **Prompt Injection Defense:** Scanned code is treated as untrusted user input wrapped inside fenced delimitations with explicit system instructions prohibiting instruction override.

---

## 7. Security and Privacy Model

1. **Zero Outbound Telemetry:** No analytics, tracking pixels, or phone-home requests exist anywhere in the CLI.
2. **Deterministic Offline Default:** The default scan and VibePatch fix mechanisms run 100% offline against embedded compiled YAML rule packs.
3. **OS Keychain Credential Storage:** BYOK API keys are stored securely using platform-native credential managers (Windows Credential Manager, macOS Keychain, Linux Secret Service / DBus) via `zalando/go-keyring`. No plaintext tokens in dotfiles.
4. **Audit Log:** Every applied fix is appended to `vibeshield-fixes.log` in JSONL format with timestamp, rule ID, file, line number, before, and after text.

---

## 8. Testing, Packaging & Release Strategy

### 8.1 Automated Test Suites
- **Unit Testing:** Table-driven tests for all scanner components (`scan_test.go`, `fix_test.go`, `rules_test.go`).
- **TUI Snapshot Testing:** Headless Bubble Tea testing using Charm's `teatest` to verify that UI views render exact expected character grids for sample view models.
- **Fixture Verification:** Continuous validation against the 15 golden repositories in `fixtures/golden/`.
- **Contract Verification Gate:** Automated CI execution of `scripts/audit-contract.mjs` ensuring that all documented commands, flags, and formats exist in the binary before release.

### 8.2 Distribution Pipeline
Cross-compilation automated via GitHub Actions (`.github/workflows/release.yml`):
- Matrix: `windows-amd64`, `windows-arm64`, `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`.
- Artifacts: Compressed archives signed with SHA-256 checksums published to GitHub Releases.
- Automated npm publish triggering `npm/vibeshield` version bump.

---

## 9. Phased Implementation Roadmap

```
  Phase 1 (v3.1.0)           Phase 2 (v3.2.0)           Phase 3 (v3.3.0)           Phase 4 (v3.4.0)
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│  TUI FOUNDATION  │  ───► │  DIFF & FIXES    │  ───► │     BYOK AI      │  ───► │ WATCH & BASELINE │
│ • Bubble Tea app │       │ • Unified diff   │       │ • /explain deep  │       │ • Daemon mode    │
│ • Stream scanner │       │ • Multi-line fix │       │ • Claude/OpenAI  │       │ • .vibeshield-   │
│ • List & detail  │       │ • Editor launch  │       │ • Ollama local   │       │   baseline.json  │
└──────────────────┘       └──────────────────┘       └──────────────────┘       └──────────────────┘
```

### Milestone 1: Interactive Foundation (v3.1.0) — Target: 2 Weeks
- Implement Bubble Tea application entrypoint in `scanner/internal/tui`.
- Implement `StreamDir` streaming scanner channel with live spinner and progress bar.
- Deliver Screen 1 (Welcome banner & slash prompt), Screen 2 (Live scanning), and Screen 3 (Findings explorer table).
- Preserve existing non-interactive `vibeshield scan` behavior and pass all contract audits.

### Milestone 2: Interactive Fixes & Diff Viewer (v3.2.0) — Target: 2 Weeks
- Deliver Screen 4 (Finding detail view) and Screen 5 (Diff inspection view).
- Implement syntax-highlighted `-`/`+` patch preview using Lip Gloss.
- Wire interactive `y/n/a` patch confirmation directly into `fix.Apply`.
- Add editor opening integration (`o` key opens default `$EDITOR` or `code -g file:line`).

### Milestone 3: BYOK AI Intelligence (v3.3.0) — Target: 2 Weeks
- Implement `internal/ai` provider interface with Anthropic Claude, OpenAI, and Ollama support.
- Implement `/explain <finding-id>` interactive pane explaining vulnerabilities and attack vectors.
- Add secure token storage using OS Keychain via `zalando/go-keyring`.
- Strict automated secret redaction prior to prompt assembly.

### Milestone 4: Developer Ergonomics & Baseline (v3.4.0) — Target: 1.5 Weeks
- Implement `vibeshield baseline` generating `.vibeshield-baseline.json` to suppress legacy technical debt.
- Implement `vibeshield watch` leveraging `fsnotify` for real-time background scans on file save.
- Update documentation and marketing site with interactive terminal demo animations.

---

## 10. Open Decisions for User Review

The following key product decisions are ready for your guidance before we proceed to implementation:

1. **Default Invocation Behavior:**
   - *Option A (Recommended):* Running bare `vibeshield` in an interactive terminal (TTY) launches the interactive Claude Code-style TUI. If piped or passed arguments (e.g. `vibeshield scan .`), it runs headless.
   - *Option B:* Require an explicit `vibeshield ui` or `vibeshield interactive` command, keeping bare `vibeshield` as a help printout.

2. **AI Engine Integration Scope:**
   - *Option A (Recommended):* Support direct BYOK (users paste an Anthropic, OpenAI, or local Ollama key stored in OS keychain), keeping VibeShield 100% serverless and free.
   - *Option B:* Keep VibeShield strictly deterministic (regex/AST only) with no AI model calls inside the binary, delegating all AI remediation to external coding agents (Cursor, Claude Code, Windsurf) via rules files.

3. **Baseline File Convention:**
   - *Recommendation:* Commit `.vibeshield-baseline.json` to git so teams can suppress legacy findings across CI and local developer machines alike.
