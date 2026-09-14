# 🛡️ VibeShield

> **Security & Dependency Auditor for AI-Generated Code**  
> Protect your codebase from hallucinated packages, leaked secrets, and insecure AI boilerplate — in CI, at pre-commit, and directly inside your coding agents. Before it reaches `main`.

[![Live Website](https://img.shields.io/badge/Website-vibeshield.dev-black?style=flat&logo=safari)](https://rajviyash9136freefr-tech.github.io/vibeshield/)
[![License: MIT](https://img.shields.io/badge/License-MIT-white.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8.svg)](scanner/go.mod)
[![GitHub Action](https://img.shields.io/badge/GitHub_Action-v1.0.0-2088FF.svg?logo=githubactions&logoColor=white)](action/)
[![Platforms](https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)](#-quick-install--all-platforms)
[![Agent Ready](https://img.shields.io/badge/Agents-Cursor%20%7C%20Copilot%20%7C%20Claude%20Code%20%7C%20Windsurf-purple.svg)](#-ai-coding-agent-skills)

---

## ⚡ What is VibeShield?

Your AI agents (Cursor, Copilot, Claude Code, Windsurf, Aider) ship hundreds of lines of code every day. Classic security scanners were built for human code and known CVEs. They miss **AI-specific failure modes**:

| AI Failure Mode | Real-World Scenario | Classic Scanners | VibeShield Gate |
|:---|:---|:---:|:---:|
| **Slopsquatting (Hallucinated Packages)** | LLM invents `fast-parse-utils-v3`; attacker registers it on npm within hours | ❌ Ignored (no CVE yet) | 🚨 **BLOCKED** at PR diff |
| **Secrets & Memory Echo** | Model pastes an active `OPENAI_API_KEY` or AWS token from chat context memory | ⚠️ Generic / noisy | 🔒 **QUARANTINED** pre-commit |
| **AI-Origin Attribution Drift** | Unvetted AI agent diffs merged without human review or accountability | ❌ Not tracked | 🎯 **ATTRIBUTED** per diff hunk |
| **Insecure AI Boilerplate** | Wildcard CORS `*`, `eval()`, hardcoded debug flags, non-expiring JWTs | ⚠️ Buried in lint noise | ⚡ **FLAGGED** with 1-line auto-fix |
| **Prompt-Injection Backdoors** | Malicious README or third-party context instructs agent to open backdoors | ❌ Zero coverage | 🛡️ **DETECTED** by AST heuristic rules |

---

## 🚀 Quick Install — All Platforms

Install VibeShield directly on any PC, laptop, server, or CI runner in seconds.

### 🍏 macOS & 🐧 Linux (Terminal / Bash)
Run the automated one-command installer (checksum-verified static binary + automatic agent skill setup):
```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

### 🪟 Windows (PC / Laptop via PowerShell)
Run directly in PowerShell (no admin privileges required):
```powershell
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex
```

### 📦 Node.js / npm / npx (Zero-Install Instant Run)
Scan your repository immediately without persistent installation:
```bash
# Instant scan via npx
npx vibeshield scan .

# Or install globally
npm install -g vibeshield
```

### 🐹 Go Developers
Install the single static compiled Go binary directly:
```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@latest
```

### 🤖 AI Coding Agent Skills (Claude Code, Cursor, Codex, Antigravity)
Equip your AI assistant with the self-auditing security skill:
```bash
# For Claude Code
/plugin marketplace add rajviyash9136freefr-tech/vibeshield
/plugin install vibeshield@vibeshield
```
*(In Cursor, Codex, or Antigravity, the `install.sh` and `install.ps1` scripts automatically register the `.cursor/skills` or `.gemini/config` skills).*

---

## 🛠️ Integration & CI/CD Setup

### 1. GitHub Action (Recommended for Pull Request Gates)
Add `.github/workflows/vibeshield.yml` to your repository:
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
      - uses: rajviyash9136freefr-tech/vibeshield/action@v1
```
*Posts a detailed **VibeCheck Report** directly to your pull request: severity breakdown, AI-origin markers, and 1-click suggested fixes in under 1.2 seconds.*

### 2. Local Git Pre-Commit Hook
Prevent secrets and hallucinated imports from ever being committed:
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v1.0.0
    hooks:
      - id: vibeshield
```

---

## 💻 CLI Usage Guide

### Basic Scanning
```bash
# Scan entire project directory
vibeshield scan .

# Scan only staged Git changes (pre-commit speed)
vibeshield scan --staged

# Scan diff against main branch
vibeshield scan --diff main

# Output machine-readable JSON or SARIF for CI pipelines
vibeshield scan . --format json
vibeshield scan . --format sarif
```

### Sample Terminal Output
```text
$ vibeshield scan .

  Scanning 14 changed files (diff mode)… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     package.json:8 — fast-parse-utils-v3@2.1.4
     Why: LLM hallucinated package name. Registered 9 days ago on npm with curl|sh install payload.
     → Fix: Replace with native node:util (clean, 12 lines, zero dependencies).

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     src/agent.ts:41 — OPENAI_API_KEY = "sk-proj-••••••••••••••4a2f"
     Why: Copilot pasted credential echoed from local context memory.
     → Fix: Move to process.env.OPENAI_API_KEY and rotate the leaked key immediately.

  ✓ 12 files clean · 2 findings · 0 leaks merged
```

### Automated Remediation (`vibeshield fix`)
```bash
# Preview proposed fixes without modifying files
vibeshield fix . --dry-run

# Interactively review and apply fixes per file
vibeshield fix .

# Non-interactive agent mode (logs all edits to vibeshield-fixes.log)
vibeshield fix . --yes
```

---

## 🛡️ Core Security Capabilities

1. **Slopsquatting Shield**: Evaluates npm, PyPI, Crates.io, and Go module additions against domain age, author reputation, install scripts, and hallucination likelihood.
2. **Secrets Zero-Echo Quarantine**: High-entropy scanners tuned for OpenAI, Anthropic, AWS, Stripe, GitHub, and private PEM keys.
3. **AI-Origin Attribution**: Distinguishes diff hunks authored by AI agents (`copilot-swe-agent`, `claude-code`, `cursor`) from human engineers.
4. **Sub-Second CI Gate**: 1.2s median PR execution time; runs locally without network bottlenecks.
5. **120+ AI-Specific Rules across 8 Languages**: JavaScript, TypeScript, Python, Go, Java, Ruby, PHP, Rust, C#.
6. **100% Local & Privacy-Guaranteed**: Analysis runs on your local machine or private GitHub runner. Source code is never persisted, collected, or used for model training.

---

## 📂 Repository Layout

```text
├── .github/          # CI/CD workflows (deploy-site.yml, release.yml)
├── action/           # GitHub Action composite definition (action.yml)
├── api/              # Fastify package-intel API service
├── contracts/        # Cross-component JSON schemas and interfaces
├── fixtures/         # Golden test fixtures with verified test keys
├── npm/              # Node launcher package published to npm
├── rules/            # Core rule packs (versioned YAML rules, MIT)
├── scanner/          # Core scanner engine (pure Go, zero CGO)
│   ├── cmd/          # vibeshield CLI entrypoint
│   └── pkg/          # AST parsers, entropy check, slopsquatting heuristics
├── scripts/          # Direct installers: install.sh (macOS/Linux), install.ps1 (Windows)
├── site/             # Astro 5 + Tailwind v4 Apple Pro landing page & docs
└── skill/            # Coding agent skill (Claude Code, Cursor, Codex, Antigravity)
```

---

## 🏗️ Building From Source

```bash
# Build Scanner (Go 1.24+)
cd scanner
go build ./cmd/vibeshield
go test ./...

# Build Marketing & Docs Site (Node 22+)
cd site
npm install
npm run build
```

---

## 📄 License & Community

- **License**: [MIT License](LICENSE) — free forever for public and private repositories.
- **Documentation**: [https://rajviyash9136freefr-tech.github.io/vibeshield/docs/](https://rajviyash9136freefr-tech.github.io/vibeshield/docs/)
- **Live Demo & Simulator**: [https://rajviyash9136freefr-tech.github.io/vibeshield/](https://rajviyash9136freefr-tech.github.io/vibeshield/)
- **Contributions**: Pull requests, new rulepacks, and agent integrations are welcome!
