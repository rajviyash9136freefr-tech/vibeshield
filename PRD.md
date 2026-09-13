# VibeShield — Product Requirements Document (PRD)

**Product:** VibeShield — the security & dependency auditor built for AI-generated code
**Form factor:** GitHub Action + pre-commit hook + CLI, packaged as a CI plugin
**Version:** 1.1 (MVP) · **Author:** Jai · **Date:** 2026-09-08 · **Amended 2026-09-13:** fully free & open source (MIT), no paid tiers
**Status:** Draft for build
**Target launch:** GitHub Marketplace public listing + `vibeshield.dev` landing site

---

## 1. Executive summary

VibeShield is a static-analysis and dependency-audit tool that runs on code the moment it is written by an AI coding agent — Cursor, GitHub Copilot, Claude Code, Windsurf, Cody — and before it is committed or merged. Unlike general-purpose scanners (Snyk, Dependabot, Semgrep, Trivy), VibeShield's entire detection model is tuned to the *failure modes of LLM-generated code*: hallucinated package names (slopsquatting), pasted-out-of-context secrets, outdated API patterns from training data, subtly insecure boilerplate, license-dropped snippets, and "it works but nobody read it" dependency trees.

The developer does not review AI code the way they review a teammate's PR. VibeShield is the reviewer that never gets tired of reading what the machine wrote.

**One-line pitch:** *"Your AI agent ships code you didn't write and can't be bothered to read. VibeShield reads it for you — in CI, at the commit hook, before it reaches main."*

---

## 2. The problem

### 2.1 What is actually happening

- AI agents now write a large and growing share of merged code. Industry surveys through 2025–2026 consistently place AI-assisted code at roughly a third to a half of new code in adopting teams (verify latest figures at publish time — cite Stack Overflow Developer Survey, GitHub Copilot research, DORA report as the public sources on the landing page).
- Developers using AI agents report **reviewing AI code less rigorously** than human code — the "vibe check" workflow: accept, run, ship.
- LLMs reliably reproduce a specific and *enumerable* class of defects that human reviewers miss and classic scanners weren't designed to weight:

| # | AI-code failure mode | Example | Classic scanner behavior |
|---|---|---|---|
| 1 | **Hallucinated packages (slopsquatting / package hallucination)** | Model invents `fast-parse-utils-v3`, attacker pre-registers it on npm/PyPI with a dropper | ❌ Not flagged — package is "new," no CVE yet |
| 2 | **Secrets in generated boilerplate** | `API_KEY = "sk-..."` pasted from training data or dev's context | ⚠️ Partially (gitleaks/truffleHog, generic) |
| 3 | **Outdated/insecure API usage** | `md5` for passwords, `eval()` on user input, SQL string concat, ECB mode | ⚠️ Rule-dependent, noisy |
| 4 | **Copy-paste license stripping** | Snippet with GPL header arrives, header removed by model | ❌ Not detected |
| 5 | **Insecure default scaffolding** | CORS `*`, `debug=True`, JWT `algorithms:["none"]`, weak Flask session secret | ⚠️ Buried in noise |
| 6 | **Unpinned transitive blowout** | Agent adds 1 library, tree gains 200 packages nobody saw | ⚠️ Shows up in `npm audit` later, not at PR time |
| 7 | **Prompt-injection-driven changes** | Malicious README/issue content instructs the agent to open a backdoor | ❌ Nobody covers this |

### 2.2 Why existing tools don't solve it

Snyk, Dependabot, Semgrep, SonarQube, and Trivy assume (a) a human wrote or at least read the code, (b) risk = known CVEs + generic bad patterns, and (c) the unit of concern is the repo. None of them answer the question a team actually has in 2026: **"What did the AI just change, and is it safe?"** Their findings also arrive as thousands of unlabeled issues; a vibe-coding dev ignores all of them. VibeShield's differentiator is not "more scanning" — it is **AI-specific detections + AI-attributed findings surfaced exactly where the AI code enters the pipeline** (the agent's own commit/PR).

### 2.3 Competitive snapshot

| Tool | AI-code-specific? | Pre-commit? | GitHub Action? | Hallucinated-package detection | Pricing wedge |
|---|---|---|---|---|---|
| Snyk | ❌ | ❌ | ✅ | ❌ | $$$ per developer |
| Dependabot | ❌ | ❌ | ✅ (native) | ❌ | Free, GitHub-only |
| Semgrep CE | ❌ | ✅ | ✅ | ❌ | Free + $$$ Pro |
| Trivy | ❌ | ⚠️ community | ✅ | ❌ | Free |
| Gitleaks / SecretScanning | ⚠️ (secrets only) | ✅ | ✅ | ❌ | Free / GitHub native |
| Socket.dev | ⚠️ (supply chain) | ✅ | ✅ | ⚠️ heuristic | $$ |
| **VibeShield** | ✅ **core design** | ✅ | ✅ | ✅ **dedicated model** | Free, MIT open source |

Positioning statement: *"The first security gate designed for the vibe-coding workflow."*

---

## 3. Target market & personas

### 3.1 Market

- **Primary:** USA-first English-speaking developers and small teams (1–50 devs) who use AI coding agents daily. Secondary: EU/UK/India pro-devs.
- **Buyer:** the staff/principal engineer or indie founder who owns CI; **user:** the same person, who is also doing the vibe-coding.
- **Distribution-native channels:** GitHub Marketplace, pre-commit.framework registry, Homebrew, npm (`npx vibeshield`), dev.to / HN / X launch.

### 3.2 Persona A — "Solo Viber" (indie hacker)

Ships weekend products with Cursor + Claude Code. Doesn't read diffs. Fear: wakes up to a banned npm package or a leaked key in a public repo. Needs: zero-config, free & open source, one-line install, badge for the README. **This persona is the viral engine** (badges + GitHub stars).

### 3.3 Persona B — "Skeptical Staff Eng" (team of 5–40)

Team has an AI-agent mandate; she has to sign off on the security story. Doesn't want to ban AI, wants a gate that makes AI output reviewable. Needs: PR comments that *explain* findings, policy config (`vibeshield.yml`), org-wide baseline, ignore-audit trail, SOC2-friendly logs. **This persona pays.**

### 3.4 Persona C — "OSS Maintainer"

Popular repo getting AI-generated drive-by PRs. Needs: GitHub Action that litters AI slop PRs with precise, kind comments and auto-labels `ai-generated-risk`. Free team tier. **This persona gives credibility and SEO backlinks.**

---

## 4. Product goals & non-goals

### Goals (MVP, next 90 days)

1. Install-to-first-scan in **under 3 minutes** for a public GitHub repo (one file, no account needed).
2. Detect the 7 failure modes in §2.1 with a curated rules engine + package-hallucination model, **precision ≥ 85 % on Critical/High** (measured against a golden set of 200 seeded AI PRs).
3. Every finding carries: **why it matters for AI code**, the vulnerable pattern, a one-line suggested fix, and a "accept/dismiss" flow that writes an auditable ignore.
4. Fully free and open source: every feature ships to everyone — the product goal is to be a default install, not a conversion funnel.

### Non-goals (MVP — explicitly out)

- Full SAST parity with Semgrep/Sonar (we are AI-behavior-focused, not a general scanner).
- Runtime/appsec, DAST, container scanning, IaC scanning (roadmap).
- GitLab/Azure DevOps/Bitbucket (roadmap; architecture keeps hosts pluggable).
- Fixing code automatically is NOT silent: `vibeshield fix` (VibePatch) applies mechanical, rule-authored same-line rewrites only behind an explicit preview + per-file y/N gate, or an explicit `--yes` for coding agents — where every patch still lands in `vibeshield-fixes.log` (see contracts/cli.md; secrets and prompt-injection findings are never auto-patched).
- IDE extensions (Cursor plugin is roadmap P1 after launch).

---

## 5. MVP feature set

### 5.1 GitHub Action — `vibeshield/action` (flagship)

- Triggered on `pull_request` and `push`. Scans the **diff**, not the whole repo (incremental-first = fast = adopted).
- Detects AI attribution signals: commits from `*[bot]` authors, Cursor/Copilot/Claude Code co-author trailers, `generate_by` markers, "looks LLM-written" heuristics — findings get an `AI-origin: likely|confirmed` tag.
- Posts one consolidated PR comment (the "VibeCheck Report"): severity card, per-file findings with anchors, dependency-delta table (what the AI added to the tree), and dismiss UI (`/vibeshield accept #hash` reply commands).
- Fail-the-check modes: `off | warn | block-on-critical | block-on-high+` (default: warn).
- README **VibeShield badge** (shield.io format) → drives organic installs and backlinks.

### 5.2 Pre-commit hook — `.pre-commit-config.yaml`

```yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v1.0.0
    hooks:
      - id: vibeshield        # staged-files scan, < 1.5 s on typical diffs
```

- Runs the same rules engine locally on staged files; offline for rules, online (optional) for package-hallucination lookups.
- Prints findings in the "explain like I reviewed this PR" format (see 5.4). `git commit --no-verify` always works (never a hostage).

### 5.3 CLI — `npx vibeshield scan .` (and `brew install vibeshield`)

- Local scans, CI-agnostic; JSON/SARIF/GitHub annotations output for power users.
- `vibeshield init` → detects frameworks, writes `vibeshield.yml`, adds the hook + workflow files, prints the 3 findings it found right now (instant aha).

### 5.4 The VibeCheck findings format (the actual product)

Every finding renders as:

```
🔴 CRITICAL · hallucinated-package · VS-PKG-001
package: fast-parse-utils-v3@2.1.4  (added in this PR by copilot-swe-agent)
Why this matters for AI code: LLMs invent plausible package names; attackers
register them within hours ("slopsquatting"). This package is 9 days old,
has 1 maintainer, no README, and an install-script that curls a remote host.
Fix: remove it — the code it's used for is 12 lines; replace with node:util
or a vetted package (e.g. yargs).      [ accept ]  [ dismiss w/ reason ]
```

Tone rule: precise, zero fear-mongering, never snarky toward the developer. Blame the pattern, not the person.

### 5.5 Rules & detection engine (MVP scope)

- **A. AI-package risk model:** for every newly added dependency — age, maintainer count, download-count-vs-tree-position, install scripts, typosquat distance, "did this package exist before the model's training cutoff" heuristic, and an LLM-namescore (does the name look *generated*). Backed by npm/PyPI/ crates/ Go module indexes refreshed hourly.
- **B. AI-pattern ruleset v1:** ~120 rules across 8 languages (JS/TS, Python, Go, Java, Ruby, PHP, Rust, C#) in the 7 categories of §2.1 — weighted and *ranked by AI-likelihood*, not generic severity alone.
- **C. Secret re-scan:** entropy + allow-list + known-format scanning, tuned to training-data-echo patterns (e.g. `sk-proj-`, `ghp_`, pasted test keys that look fake but are real).
- **D. Diff-aware dependency tree delta:** what the PR added transitively, with the risk of each new node.
- **E. Attribution:** which findings sit inside hunks the AI most likely wrote.

### 5.6 Config — `vibeshield.yml`

```yaml
mode: warn                 # off | warn | block-on-critical | block-on-high+
languages: [typescript, python]
ignore:
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"   # stored in audit log
thresholds:
  new_dependency_max_age_days: 30
notifications:
  slack: ${{ secrets.SLACK_WEBHOOK }}
```

### 5.7 Dashboard (thin web app, launch-adjacent, not launch-blocking)

- `app.vibeshield.dev`: org settings, findings history, ignore audit log, weekly "AI risk digest" email. MVP = email digest + read-only results; edits minimal.

---

## 6. Architecture (build view)

```
┌────────────┐   ┌──────────────────────┐   ┌──────────────────────────┐
│ GH Action  │──▶│ Scanner binary (Go)  │   │ vibeshield.dev API       │
│ (composite)│   │ rules engine, diff   │   │ (Fastify on Node / Bun)  │
├────────────┤   │ parser (tree-sitter) │   │  ├ package-intel svc     │
│pre-commit  │──▶│ secret scan          │   │  │  (npm/PyPI crawlers)  │
│hook (Go)   │   └──────────┬───────────┘   │  ├ findings store (PG)   │
├────────────┤              │ SARIF/JSON    │  └ auth (GitHub OAuth)   │
│ CLI (same  │──────────────┘               └──────────────────────────┘
│ Go binary) │   Rules ship as signed, versioned packs (OSS core pack;
└────────────┘   "AI-hardening pack" = MIT OSS, updated weekly)
```

Key decisions:

1. **Scanner = single static Go binary.** Installs everywhere (brew/npm-release/go-get), no runtime deps, boots <50 ms, works offline for rules. Critical for pre-commit UX.
2. **tree-sitter parsing** per language → robust to weird formatting that LLMs emit; rules are queries + dataflow-lite taint tags, not regex.
3. **Rules as data, not code.** Versioned YAML packs → weekly updates without binary releases; the AI-hardening pack is MIT-licensed and community-maintained alongside the core pack (open source everywhere = trust + distribution; the moat is the intelligence, not the license).
4. **Package-intelligence service is the server-side moat** — the hourly-rebuilt registry snapshot (age, maintainers, scripts, download velocity, typosquat graph) is something a local-only tool can't replicate.
5. **Privacy by default:** code never leaves the machine in CLI/hook mode; in Action mode only diff *analysis results* (file paths, rule IDs, symbol names) go to the dashboard — file contents never persist. State this on the landing page in one sentence and in the marketplace listing.

---

## 7. SEO & growth (the landing-site brief)

The `vibeshield.dev` site is an Astro static build (same stack as our other properties), English (en) first with hreflang ready for es/de/fr, JSON-LD `FAQPage` + `SoftwareApplication` + `HowTo` schema, and the FAQ below rendered as real HTML (not accordion-hidden text; use `<details>` which Google indexes).

### 7.1 Searchable keyword map (USA-first)

**Primary (money) keywords — home + GitHub Action page:**
`vibe coding security` · `AI generated code security scanner` · `security scanner for AI code` · `github action security scan` · `pre-commit security hook` · `AI code review tool`

**Secondary (high-intent, low competition today):**
`hallucinated npm packages` · `slopsquatting` · `slopsquatting meaning` · `package hallucination attacks` · `prompt injection code example` · `is AI generated code safe` · `cursor security scan` · `copilot security scanner` · `claude code security check` · `AI coding agent security` · `secure vibe coding` · `LLM code injection`

**Long-tail FAQ-target queries (one page-section per question, each with its own H3 + JSON-LD entry):**
`who is responsible for security of AI generated code` · `can AI write insecure code` · `do I need a security scanner for AI code` · `how to scan AI generated code before committing` · `github action to check for secrets` · `detect ai generated pull request` · `npm package hallucination fix` · `is dependabot enough` · `semgrep vs snyk for AI code`

**Comparison-page keywords (roadmap content, each = its own page):**
`snyk alternative` · `socket.dev alternative` · `vibeshield vs snyk` · `vibeshield vs dependabot` · `vibeshield vs semgrep` · `best security tools for AI coding 2026`

### 7.2 FAQ section (ship this verbatim on the landing page)

**Q1. What is VibeShield?**
VibeShield is a security and dependency auditor built specifically for AI-generated code. It runs as a GitHub Action, a pre-commit hook, and a CLI, catching hallucinated packages, pasted secrets, insecure AI boilerplate, and risky new dependencies the moment your AI agent writes them — before they reach `main`.

**Q2. Why isn't Dependabot or Snyk enough for AI-generated code?**
Because they look for known vulnerabilities, and most AI-code risk isn't "known" yet. When an LLM invents a package name and an attacker registers it the same week, there is no CVE to match — the package *is* brand-new. VibeShield scores every newly-added dependency on supply-chain signals (age, maintainers, install scripts, name-hallucination likelihood) instead of waiting for someone to report abuse. It complements Dependabot; it doesn't replace it.

**Q3. What is slopsquatting (package hallucination)?**
LLMs confidently cite packages that don't exist. Attackers harvest those invented names, publish malicious versions to npm/PyPI/crates, and wait for a developer's AI agent to `npm install` them — a documented attack class (see the Boston University and Veracode studies on package hallucination). VibeShield's model flags names that look generated and packages that appeared suspiciously recently, and blocks the install path at PR time.

**Q4. Does VibeShield work with Cursor, Copilot, and Claude Code?**
Yes. It doesn't plug into the agents themselves — it gates what they output. Anything written by Cursor, GitHub Copilot, Copilot Workspace, Claude Code, Windsurf, Aider, Devin, or an agent-authored PR (`copilot-swe-agent`, custom bots) is scanned at commit and at PR. Findings are attributed to likely-AI hunks using bot-author metadata and writing-pattern heuristics.

**Q5. Will it slow down my CI or flood me with false positives?**
Diff-only scanning takes seconds, not minutes — most PRs finish under 10s in the Action and under 1.5s in the pre-commit hook. The ruleset is deliberately narrow (AI failure modes, not general linting): ~120 rules versus thousands in general SAST tools. Default mode is `warn`; you choose when it can block a merge.

**Q6. Does my source code leave my machine?**
No. The CLI and pre-commit hook run fully local analysis. In GitHub Action mode, only findings metadata (file paths, rule IDs, symbol names) is sent to your dashboard — file contents are never persisted, and private-repo analysis results are never used for model training. Ever.

**Q7. How much does VibeShield cost?**
Nothing. VibeShield is fully open source under the MIT license — the scanner, the GitHub Action, the pre-commit hook, and every rules pack. No tiers, no seats, no credit card, no account. Install takes one YAML file, and it works offline forever.

**Q8. How do I get started in 3 minutes?**
Add `uses: rajviyash9136freefr-tech/vibeshield/action@v1` to your workflow (or run `npx vibeshield scan .` / `go install` for the CLI). VibeShield scans the next PR, comments a VibeCheck report, and you can enable the badge on your README. No account, no signup — MIT-licensed and free for every repo.

**Q9. Can VibeShield fix the code automatically?**
Yes — with a human gate, never silently. `vibeshield fix` previews every change (which file was read, −/+ per line), asks `[y/N]` per file, and writes an audit trail to `vibeshield-fixes.log`. Coding agents can run `vibeshield fix --yes` to apply the mechanical patches directly, but only mechanical, rule-authored, same-line rewrites are ever eligible — and secrets or prompt-injection findings are excluded by contract, because a key needs rotation and an agent editing its own instruction file is exactly the problem we're here to reduce, not reinvent.

**Q10. Who is responsible for the security of AI-generated code?**
Legally and practically, still you — the shipping team. Standards bodies and enterprise procurement are converging on "AI-authored changes need the same gate as human-authored changes"; VibeShield gives you that gate plus the audit trail (who shipped what, when, with what risk) that questions like this require.

### 7.3 Launch content plan (first 10 posts, all keyword-mapped)

1. HN + dev.to: "We scanned 500 AI-generated PRs. Here's what the bots shipped." (data study, linkbait with substance)
2. `slopsquatting explained` pillar post → Q3 above becomes the TL;DR.
3. "Cursor doesn't check what it installs: a 3-minute gate" → install walkthrough.
4. Comparison: VibeShield vs Snyk/Socket/Dependabot for AI code (honest table).
5. CVE-style disclosure write-ups of real hallucination incidents (anonymized with vendors).
6. "The vibe-coding threat model" — the doc teams forward to their security lead.
7–10. One per major agent release cycle ("What to gate when Claude Code does X").

---

## 8. Pricing & packaging

**Everything is free.** VibeShield is MIT-licensed open source end to end — scanner, GitHub Action, pre-commit hook, all rules packs (core now, AI-hardening as a community pack). No tiers, no seats, no SSO upcharge, no account.

The packaging question was reframed on 2026-09-13: the moat for an AI-code auditor is *trust + rule quality + distribution*, not a paywall. Security buyers run what they can read; a fully public codebase and rule set is the strongest argument to run it. Support stays community-first (GitHub issues); anything commercial (hosted dashboard, SLAs) is explicitly out of scope for the OSS project and would live, if ever, in a separate repo.

---

## 9. Metrics

**Activation:** installs → 2+ scans within 7 days ≥ 40%. Time-to-first-VibeCheck < 3 min (median).
**Quality:** precision ≥ 85% C/H findings (golden set, refreshed monthly); ≤ 5 false positives per 1k diff-lines on median PR; noise complaints ("turned it off") < 8% of installs/mo.
**Growth:** GitHub stars 1k in 90 days (launch bar); 30% of installs from badge/README links; organic signups from SEO ≥ 50% of total by month 6.
**Reach (no revenue targets — the tool is free):** installs tracked by release-asset downloads + Action usage; ≥ 500 weekly active repos by month 6; CAC ≈ 0 (SEO/Marketplace-native distribution).

---

## 10. Roadmap

**v1.0 (launch, ~6 weeks):** Action + pre-commit + CLI, rules engine + 8-language AI pack, hallucinated-package model, PR comment report, badge, landing site — all free. VibePatch ships here as the gated `vibeshield fix` command (mechanical autofixes, preview + y/N or audited `--yes`).
**v1.1 (+6 weeks):** ignore-audit UI, org config inheritance, Slack digest, SARIF export (defensible on a GitHub security tab), VS Code / Cursor companion extension.
**v1.2:** broader autofix coverage per rule pack, prompt-injection-change detection for agent-driven repos, GitLab CI support.
**v2.0 (next year):** Business tier (SSO, SBOM export, audit reports), IDE pre-generation warnings, registry-intel API licensing to other scanners, "AI change provenance" attestation (signed AI-origin metadata per hunk).

---

## 11. Risks & mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Snyk/GitHub ships "AI-code mode" and bundles it | High (12–18 mo) | Speed + niche truth: our ruleset, terminology, and tone stay AI-workflow-native; they will generalize, we specialize. Ship the community rules pack early to own the category vocabulary. |
| False-positive fatigue → churn | Medium | Narrow ruleset by design, `warn` default, one-click dismiss-with-reason feeding a feedback loop, golden-set precision bar in CI. |
| slopsquatting news cycle fades | Medium | Category already broader than one attack: we're "security for AI-authored code," with the whole §2.1 table. |
| Legal/IP: findings-based blame ("the AI did it") | Low | Language policy: blame patterns, not people/agents; never name-and-shame model vendors in public outputs. |
| Registry APIs (npm/PyPI) rate-limit the intel service | Medium | Multi-source snapshots (deps.dev, libraries.io) + nightly bulk + graceful degradation to offline heuristics. |

---

## 12. Open questions (decide before build week 2)

1. Where does the "AI-origin heuristics" model run — bundled in the Go binary (offline) or server-side (fresher)? *Leaning: tiny ONNX/standalone classifier embedded + server-side reinforcement.*
2. Public-repo results as an open "AI code risk index" (data marketing) — ship in v1.0 or v1.1?
3. Logo/name check: "VibeShield" is descriptive and meme-adjacent — verify trademark + GitHub org availability before the Marketplace listing.

---

*End of PRD. UI/UX specification lives in `UIUX.md`.*
