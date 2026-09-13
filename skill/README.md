# VibeShield — bug-hunting audit skill

A coding-agent **skill** that audits *your local project* for the failure modes
of AI-generated code, hunting bugs with **parallel subagents** and writing
findings in the VibeShield contract format
([`contracts/finding/schema.json`](../contracts/finding/schema.json)).
It works with or without the VibeShield binary, fully offline, static analysis
only — your code never leaves your machine.

Categories hunted: hallucinated packages · hardcoded secrets · insecure API
usage · license stripping · insecure defaults · dependency risk ·
prompt-injection traps.

## Install — one command

**Claude Code (recommended).** Paste into a Claude Code session:

```
/plugin marketplace add vibeshield/vibeshield
/plugin install vibeshield@vibeshield
```

Or from your terminal:

```bash
claude plugin marketplace add vibeshield/vibeshield && claude plugin install vibeshield@vibeshield
```

(The marketplace manifest lives at the repo root; the plugin itself in
`skill/` — one repo, both the scanner and the skill.)

**Any coding IDE that reads Agent Skills** (Cursor, GitHub Copilot / VS Code,
Gemini CLI, Codex, Amp, …): copy this repo's `skill/skills/vibeshield-audit/`
folder into your client's skills directory, e.g.

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/vibeshield/vibeshield.git /tmp/vibeshield
cd /tmp/vibeshield && git sparse-checkout set skill
# Claude Code (personal):
mkdir -p ~/.claude/skills && cp -r skill/skills/vibeshield-audit ~/.claude/skills/
# Cursor / VS Code (per project):
mkdir -p .cursor/skills && cp -r skill/skills/vibeshield-audit .cursor/skills/
```

Windows (Git Bash): same commands work; `~/.claude/skills` expands to
`%USERPROFILE%\.claude\skills`.

## Use

Open your project, then:

```
/vibeshield:vibeshield-audit
```

(plain `/vibeshield-audit` when installed as a personal/project skill).
Options: `--quick` (no subagent fan-out), `--diff main` (audit only your
changes vs `main`), or free-text scope: `/vibeshield:vibeshield-audit --quick focus on src/api`.

What happens:

1. **Scope** — the repo is split into work units (modules, manifests, CI, docs).
2. **Hunt** — one bug-hunter subagent per unit runs in parallel, grepping the
   playbook patterns and *reading* the code to confirm each hit.
3. **Verify** — every CRITICAL/HIGH candidate faces an adversarial skeptic that
   tries to refute it; refuted findings are dropped, not softened.
4. **Report** — `vibeshield-findings.json` (contract format, secrets redacted,
   validated) + a VibeCheck summary in your terminal: severity, rule ID, why it
   matters for AI code, one-line fix.

Zero findings is an honest answer; the skill is barred from padding.

## Trust

- Read-only on your code by design (hunters get `Read/Grep/Glob` only; the
  orchestrator writes exactly one file: the findings JSON).
- Never executes your project's code, never installs dependencies, no network.
- Core skill is MIT; the rule IDs match the open core pack in `rules/`.

The hosted gate — GitHub Action + pre-commit + CLI — lives at
[docs](https://vibeshield.dev/docs); this skill is the in-agent audit, sharing
the same finding contract so both roll up into one report.
