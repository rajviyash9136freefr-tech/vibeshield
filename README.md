<div align="center">

# 🛡️ VibeShield

### The Bug Hunter for AI-Generated & Vibe-Coded Code

**Instantly catch hallucinated packages, leaked API keys, and insecure defaults before they reach production.**

[![Version](https://img.shields.io/badge/version-v3.0.0-blue?style=for-the-badge&logo=git&logoColor=white)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=for-the-badge)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases)
[![Website](https://img.shields.io/badge/🌐_Website-Live_Demo-purple?style=for-the-badge)](https://rajviyash9136freefr-tech.github.io/vibeshield/)

<br/>

[⚡ **Live Web Simulator**](https://rajviyash9136freefr-tech.github.io/vibeshield/) • [📦 **Downloads**](#-downloads--installation) • [🚀 **Quick Start**](#-quick-start) • [📋 **Releases**](https://github.com/rajviyash9136freefr-tech/vibeshield/releases)

<br/>

</div>

---

## 📥 Downloads & Installation

Choose your operating system below for one-command install or direct binary download:

### 🪟 Windows

**Option 1: One-Line Install (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex
```

**Option 2: Direct Binary Download (.zip)**
- ⬇️ **[Download for Windows 64-bit (x86_64)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-windows-x86_64.zip)**
- ⬇️ **[Download for Windows ARM64](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-windows-aarch64.zip)**

*(Extract `vibeshield.exe` and add it to your PATH or copy to your project folder).*

---

### 🍏 macOS

**Option 1: One-Line Install (Terminal)**
```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

**Option 2: Direct Binary Download (.tar.gz)**
- ⬇️ **[Download for Apple Silicon (M1/M2/M3/M4 - ARM64)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-darwin-aarch64.tar.gz)**
- ⬇️ **[Download for Intel Mac (x86_64)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-darwin-x86_64.tar.gz)**

---

### 🐧 Linux

**Option 1: One-Line Install (Terminal)**
```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

**Option 2: Direct Binary Download (.tar.gz)**
- ⬇️ **[Download for Linux 64-bit (x86_64)](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-linux-x86_64.tar.gz)**
- ⬇️ **[Download for Linux ARM64](https://github.com/rajviyash9136freefr-tech/vibeshield/releases/download/v3.0.0/vibeshield-linux-aarch64.tar.gz)**

---

### ⚡ Other Ways to Run

#### Run instantly with NPM / npx (No install needed)
```bash
npx vibeshield scan .
```

#### Install from source with Go
```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@v3.0.0
```

---

## 🚀 Quick Start & CLI Usage

### 🖥️ 1. Interactive Terminal Mode (Claude Code Experience)

Just type `vibeshield` in any project folder to enter the full interactive terminal application:

```bash
vibeshield
```

Get a live project detection banner, real-time scanning progress, interactive findings explorer, and atomic diff patch confirmation without leaving your terminal:

```text
╭──────────────────────────────────────────────────────────────────────────────╮
│  🛡️  VIBESHIELD v3.1.0 — AI Code Security Scanner                           │
│  The Bug Hunter for AI-Generated & Vibe-Coded Software                       │
╰──────────────────────────────────────────────────────────────────────────────╯
   Project  › my-vibe-app (git: main)
   Stack    › TypeScript 5.4 • React 18 • Supabase JS
   Engine   › 117 active rules loaded • 100% Local & Offline

 Quick Keys:  [Enter / s] Full Scan    [d] Diff Scan    [f] Fix Wizard    [q] Quit

 > /scan .
```

---

### ⚡ 2. Command Line (Headless / CI Mode)

Use subcommands directly for fast audits, git hooks, and CI/CD pipelines:

```bash
# Audit the whole repository
vibeshield scan .

# Audit only uncommitted/staged changes (used in pre-commit hook)
vibeshield scan --staged

# Machine-readable JSON output
vibeshield scan . --format json

# Preview and apply mechanical fixes (VibePatch)
vibeshield fix .

# Check environment, git hooks, and rule engine health
vibeshield doctor

# One-command project onboarding (writes config, git hook & CI workflow)
vibeshield init
```

### Example Scan Output

```text
$ vibeshield scan .
Scanning 14 files (full mode)… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     package.json:8 — fast-parse-utils-v3@2.1.4 (registered 9 days ago on npm)
     → Fix: replace with node:util

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     src/agent.ts:41 — OPENAI_API_KEY pasted from chat context
     → Fix: read process.env.OPENAI_API_KEY

  ✓ 12 files clean · 2 findings fixed
```

---

## 🔍 What VibeShield Checks

| Category | What it catches |
| :--- | :--- |
| 📦 **Hallucinated Packages** | Non-existent or newly-registered npm/PyPI/Go libraries invented by AI models |
| 🔑 **Hardcoded Secrets** | API keys (OpenAI, Anthropic, AWS, Stripe) pasted directly into code |
| 🛡️ **Insecure Defaults** | Wildcard CORS (`*`), `debug=True`, unhashed passwords, disabled SSL verification |
| 🤖 **Agent Rules** | Automatic rule injection for Cursor, Claude Code, Copilot, Windsurf & Antigravity |

---

## 🔗 Useful Links

- 🌐 **[Live Demo & Web Simulator](https://rajviyash9136freefr-tech.github.io/vibeshield/)**
- 📦 **[All GitHub Releases](https://github.com/rajviyash9136freefr-tech/vibeshield/releases)**
- 📖 **[Full Documentation & Guides](https://rajviyash9136freefr-tech.github.io/vibeshield/docs/)**
- 🐛 **[Report an Issue](https://github.com/rajviyash9136freefr-tech/vibeshield/issues)**

---

<div align="center">

Distributed under the **MIT License**. Free and open source for everyone.

</div>
