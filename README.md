<div align="center">

# 🛡️ VibeShield
### The Bug Hunter & Tester for AI Vibe Coding

**Test your vibe-coded apps, hunt down hallucinated packages, leaked API keys, and insecure AI defaults — with instant 1-line verified fixes.**

[![Website](https://img.shields.io/badge/🌐_Website-Live_Simulator-white?style=for-the-badge&logo=googlechrome&logoColor=black)](https://rajviyash9136freefr-tech.github.io/vibeshield/)
[![GitHub Stars](https://img.shields.io/github/stars/rajviyash9136freefr-tech/vibeshield?style=for-the-badge&logo=github&color=white&labelColor=black)](https://github.com/rajviyash9136freefr-tech/vibeshield/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-white?style=for-the-badge)](LICENSE)
[![GitHub Action](https://img.shields.io/badge/PR_Gate-v1.0.0-white?style=for-the-badge&logo=githubactions&logoColor=black)](action/)

[![Agents Supported](https://img.shields.io/badge/Supported_Agents-Cursor_·_Claude_Code_·_Windsurf_·_Copilot-111111?style=flat-square&logo=visualstudiocode)](https://rajviyash9136freefr-tech.github.io/vibeshield/#agents)
[![Offline & Local](https://img.shields.io/badge/100%25_Local_&_Offline-Zero_Data_Sent_Outside-success?style=flat-square)](#-frequently-asked-questions)
[![Audit Speed](https://img.shields.io/badge/Audit_Speed-1.2s_Sub--Second-blue?style=flat-square)](#-sample-bug-hunter-output)

<br/>

[⚡ **Try Live Web Simulator**](https://rajviyash9136freefr-tech.github.io/vibeshield/#scanner) · [🤖 **How to Setup Agents**](#-ai-agent-prompts--setup) · [📦 **How to Install**](#-1-command-quick-install) · [🔍 **How to Scan**](#-how-to-scan--hunt-bugs) · [❓ **FAQ**](#-frequently-asked-questions) · [🌟 **Community**](#-welcome-to-the-community)

<br/>

```text
  $ vibeshield scan .
  Scanning 14 vibe-coded files… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     package.json:8 — fast-parse-utils-v3@2.1.4 (Registered 9d ago on npm)
     → Fix: Replace with native node:util (clean, 12 lines, zero dependencies).

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     src/agent.ts:41 — OPENAI_API_KEY echoed from chat context
     → Fix: Move to process.env.OPENAI_API_KEY and rotate key now.

  ✓ 12 files clean · 2 findings · 0 leaks merged to main
```

</div>

---

## 🧩 The VibeShield Ecosystem: How Everything Works Together

To test and secure vibe-coded applications, VibeShield consists of three interconnected components:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                           1. THE WEBSITE & SIMULATOR                        │
│                https://rajviyash9136freefr-tech.github.io/vibeshield/        │
│  • Test real AI bug scenarios live in the browser                           │
│  • 1-Click copy prompts for Cursor, Claude Code, Windsurf & Copilot         │
│  • 1-Line automated terminal install commands & 5 curated FAQs              │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          2. THE IN-AGENT SKILL                              │
│       Cursor (.cursorrules) · Claude Code (/plugin) · Windsurf (.windsurfrules)│
│  • Sits directly inside your coding agent                                    │
│  • Intercepts hallucinated dependencies & leaked keys in real time          │
│  • Suggests verified 1-line standard library fixes before code is saved     │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         3. THE SCANNER & PR GATE                            │
│                 CLI (vibeshield scan .) · GitHub Action (CI/CD)             │
│  • 100% offline, local Go static analysis binary (runs in 1.2s)             │
│  • Pre-commit hooks block secrets from ever entering git history            │
│  • Posts clear VibeCheck review reports directly on pull requests           │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Component | What It Does | Where It Runs | How to Use |
|:---|:---|:---|:---|
| **🌐 The Website** | Interactive visual simulator, agent prompts hub, install guide, and FAQ | Web Browser | [Open Live Simulator](https://rajviyash9136freefr-tech.github.io/vibeshield/) |
| **🤖 The Agent Skill** | Intercepts hallucinations & leaked secrets inside the AI agent before code is saved | Cursor, Claude, Windsurf, Copilot | Copy prompts into `.cursorrules` or install plugin |
| **⚡ The CLI Scanner** | Sub-second offline static analysis engine (1.2s) hunting supply-chain flaws | Local Terminal (macOS/Linux/Windows) | `vibeshield scan .` or `npx vibeshield scan` |
| **🛡️ GitHub Action** | Automated Pull Request gate blocking dangerous diffs before merge | GitHub CI/CD | `uses: rajviyash9136freefr-tech/vibeshield/action@v1` |

---

## ⚡ Why VibeShield for AI Vibe Coding?

When vibe coding, developers ship hundreds of lines of AI-generated code every session without reading every single character. Classic security linters were built for human code and known CVE databases — they completely miss **AI-specific failure modes**:

| Vibe Coding Bug | Real-World Attack / Flaw | Classic Linters | VibeShield Bug Hunter |
|:---|:---|:---:|:---:|
| **Hallucinated Packages (Slopsquatting)** | LLM invents `fast-parse-utils-v3`; threat actors register it on npm with malicious payloads | ❌ Ignored (No CVE yet) | 🚨 **BLOCKED** at PR diff |
| **Pasted API Keys & Secret Leaks** | AI pastes live `OPENAI_API_KEY` or AWS token from chat context into client components | ⚠️ Generic / noisy | 🔒 **QUARANTINED** pre-commit |
| **Broken Logic & Component States** | AI writes infinite render loops, missing error boundaries, or race conditions | ❌ Ignored | 🎯 **TESTED** & flagged |
| **Insecure AI Boilerplate** | AI scaffolds wildcard CORS (`*`), `debug=True`, unhashed passwords, or permissive JWTs | ⚠️ Buried in lint noise | ⚡ **FLAGGED** with 1-line fix |
| **Prompt-Injection Traps** | Malicious README or third-party context tricks your coding agent into opening a backdoor | ❌ Zero coverage | 🛡️ **DETECTED** by AST heuristics |

---

## 🤖 AI Agent Prompts & Setup

Copy and paste these dedicated rules directly into your coding agents to turn them into adversarial bug hunters:

### 🟦 Cursor (`.cursorrules`)
Create a `.cursorrules` file in your repository root and paste:
```markdown
# VibeShield Bug Hunter Rule for Cursor
You are an adversarial code auditor and bug hunter for vibe-coded applications.
Before finalizing, applying, or committing any code in this project:
1. HALLUCINATION CHECK: Audit every imported package, module, or API. If a library is not well-established, flag it and replace with standard library or verified packages.
2. SECRET SCAN: Ensure NO OpenAI, Anthropic, Stripe, AWS, or database credentials are leaked in client-side code or git diffs.
3. LOGIC & COMPONENT TESTING: Inspect state updates, race conditions, async error handling, and broken edge cases in UI components.
4. INSECURE DEFAULTS: Never scaffold wildcard CORS (*), debug=True, weak session secrets, or JWT algorithms:["none"].
5. ATOMIC FIX: For every bug found, provide an immediate one-line diff replacement.
```

### 🟣 Claude Code (`/plugin` & Skill)
Install the official VibeShield skill into your Claude Code session:
```bash
# Add marketplace & install plugin
/plugin marketplace add rajviyash9136freefr-tech/vibeshield
/plugin install vibeshield@vibeshield

# Run an audit on your project
/vibeshield:vibeshield-audit --quick
```

### 🟩 Windsurf Cascade (`.windsurfrules`)
Add to `.windsurfrules` in your workspace:
```markdown
# VibeShield Security & Bug Hunter for Windsurf Cascade
When generating or refactoring code in this repository:
1. Verify all dependencies against hallucinated package names (slopsquatting).
2. Scan all generated code for credentials, API tokens, and private environment variables.
3. Test edge-case logic, error handling, and component state transitions.
4. Reject insecure boilerplate: wildcard CORS, disabled auth, and unsafe evals.
5. Provide safe, production-grade replacements for any identified vulnerability.
```

### 🔲 GitHub Copilot (`.github/copilot-instructions.md`)
Add to `.github/copilot-instructions.md`:
```markdown
# VibeShield Bug Hunter Instructions for GitHub Copilot
Act as VibeShield Bug Hunter and adversarial code reviewer for AI-generated diffs:
1. Identify hallucinated packages and unpinned supply chain risks in package manifests.
2. Quarantine any hardcoded secrets, API tokens, or credentials echoed from context.
3. Scrutinize component state, edge cases, and asynchronous error boundaries.
4. Flag insecure defaults (CORS *, debug flags, permissive JWTs) and suggest clean diff fixes.
```

---

## 🚀 1-Command Quick Install

Install the standalone VibeShield binary on any OS in seconds (zero dependencies, no sudo required):

### 🍏 macOS & 🐧 Linux (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
```

### 🪟 Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex
```

### 📦 Node.js / NPX (Zero-Install Scan)
Scan your repository immediately without installing any permanent binary:
```bash
npx vibeshield scan .
```

### 🐹 Go Developers
```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@latest
```

---

## 🔍 How to Scan & Hunt Bugs

Run sub-second audits from your terminal or directly inside your agent:

```bash
# 1. Audit your entire project
vibeshield scan .

# 2. Audit only current vibe-coding changes against main (fast PR mode)
vibeshield scan --diff main

# 3. Fast pre-commit hook scan before git commit
vibeshield scan --staged

# 4. Interactively preview and apply verified 1-line fixes
vibeshield fix .
```

---

## 🛠️ Automated CI/CD Setup

### GitHub Action (Pull Request Gate)
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
*Automatically posts a **VibeCheck Report** directly to your pull request with severity tags and 1-click suggested diff fixes in under 1.2 seconds.*

### Local Git Pre-Commit Hook
Prevent secrets and hallucinated packages from ever being committed to git:
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v1.0.0
    hooks:
      - id: vibeshield
```

---

## ❓ Frequently Asked Questions

### 1. What is VibeShield and how does it test vibe-coded apps?
VibeShield is an adversarial bug hunter and security tester built specifically for AI vibe coding. When AI agents (Cursor, Claude Code, GitHub Copilot, Windsurf) generate hundreds of lines of code, they introduce hallucinated packages, leaked keys, and broken component logic. VibeShield inspects code the moment it is written, hunts these specific bugs, and gives you instant 1-click fixes before you commit or merge.

### 2. How do I install and use VibeShield with my AI coding agent?
You can use VibeShield in three ways: (1) Copy our dedicated agent prompt into your `.cursorrules`, `.windsurfrules`, or Claude Code skill; (2) Run our 1-command installer on macOS/Linux (`curl -fsSL ... | bash`) or Windows (`irm ... | iex`); or (3) Add our automated GitHub Action to audit every Pull Request in under 1.2 seconds.

### 3. What specific bugs and vulnerabilities does VibeShield hunt down?
VibeShield targets failure modes unique to LLM-generated code: hallucinated npm/pip packages (slopsquatting attacks where attackers pre-register names invented by LLMs), leaked API keys and passwords echoed in boilerplate, insecure wildcard CORS (*), broken JWT auth, unhandled UI state bugs, and prompt-injection backdoors.

### 4. Does my source code or prompt data leave my machine?
No. Never. VibeShield is designed for zero data exfiltration. The CLI, agent skills, and pre-commit hooks execute 100% locally and offline on your machine using static analysis. Your code, API keys, and prompts are never sent to external servers and are never used for model training.

### 5. Is VibeShield free, and can I use it for private repositories?
Yes. VibeShield is 100% free and open-source under the MIT license. There are no paid tiers, no seat limits, no tokens, and no account required. You can freely use it on unlimited personal, public, and private commercial repositories forever.

---

## 🌟 Welcome to the Community!

We welcome every developer, vibe coder, and AI enthusiast! Here is how you can get involved:

* ⭐ **Star this repository** to support open-source AI security tools.
* 🧪 **Try the Live Simulator**: Visit [https://rajviyash9136freefr-tech.github.io/vibeshield/](https://rajviyash9136freefr-tech.github.io/vibeshield/) to test real AI bugs live.
* 🐛 **Report a Bug / Suggest a Rule**: Encountered a new hallucinated package or LLM failure mode? [Open an issue](https://github.com/rajviyash9136freefr-tech/vibeshield/issues).
* 🤝 **Contribute**: Check out [CONTRIBUTING.md](CONTRIBUTING.md) to add new agent skills or detection heuristics.

---

## 📂 Repository Layout

```text
├── .github/          # CI/CD workflows (deploy-site.yml, release.yml)
├── action/           # GitHub Action composite definition (action.yml)
├── contracts/        # Cross-component JSON schemas and interfaces
├── fixtures/         # Golden test fixtures with verified test diffs
├── npm/              # Node launcher package published to npm
├── rules/            # Core rule packs (versioned YAML rules, MIT)
├── scanner/          # Core scanner engine (pure Go, zero CGO)
│   ├── cmd/          # vibeshield CLI entrypoint
│   └── pkg/          # AST parsers, entropy check, slopsquatting heuristics
├── scripts/          # Direct installers: install.sh (macOS/Linux), install.ps1 (Windows)
├── site/             # Astro 5 + Tailwind v4 Apple Pro landing page & docs
└── skill/            # Coding agent skill (Claude Code, Cursor, Windsurf, Copilot)
```

---

## 📄 License & Links

- **License**: [MIT License](LICENSE) — 100% Free and Open Source.
- **Website & Interactive Simulator**: [https://rajviyash9136freefr-tech.github.io/vibeshield/](https://rajviyash9136freefr-tech.github.io/vibeshield/)
- **Documentation**: [https://rajviyash9136freefr-tech.github.io/vibeshield/docs/](https://rajviyash9136freefr-tech.github.io/vibeshield/docs/)

---

<!-- GitHub Topic Keywords for Search Ranking -->
<!-- vibe-coding, ai-coding, cursor, claude-code, github-copilot, windsurf, bug-hunter, security-audit, slopsquatting, hallucinated-packages, pre-commit, github-action, cursorrules, ai-code-tester, devsecops, vibe-code-tester -->
