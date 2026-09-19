---
title: "Bug-hunting skill"
description: "Paste one command into your coding agent and it audits your project for AI-code bugs with parallel hunter subagents. Claude Code plugin + open Agent Skills format."
order: 7
---

VibeShield also ships as an **agent skill**: paste one install command into
Claude Code (or your Agent-Skills-compatible coding IDE), then run one command
and the agent hunts bugs in *your* project the way the CI gate does — same
categories, same finding contract — and writes `vibeshield-findings.json`
locally. Nothing leaves your machine; the skill is read-only plus one output
file.

## Install (Claude Code)

In a Claude Code session:

```
/plugin marketplace add rajviyash9136freefr-tech/vibeshield
/plugin install vibeshield@vibeshield
```

From your terminal:

```bash
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield && claude plugin install vibeshield@vibeshield
```

## Install (other coding agents)

The skill core is the open [Agent Skills](https://agentskills.io) format —
portable to Cursor, GitHub Copilot / VS Code, Gemini CLI, Codex and ~40 other
clients: copy `skill/skills/vibeshield-audit/` from the repo into your client's
skills directory (`~/.claude/skills/`, `.cursor/skills/`, …). Claude-specific
pieces (the plugin's bundled bug-hunter subagent, `${CLAUDE_SKILL_DIR}` paths)
degrade to plain-skill mode elsewhere.

## Run

```
/vibeshield:vibeshield-audit              # whole repo
/vibeshield:vibeshield-audit --diff main  # only my changes vs main
/vibeshield:vibeshield-audit --quick      # no subagent fan-out
```

## How the hunt works

1. **Scope** — modules, manifests/lockfiles, CI/agent configs, and docs (the
   injection surface) become work units. If the `vibeshield` binary is present,
   its scan seeds the baseline; otherwise the skill runs fully standalone.
2. **Fan-out** — one bug-hunter subagent per unit, in parallel, grepping the
   playbook (the same failure-mode knowledge as the
   [rules](/docs/rules)) and reading code to confirm each hit.
3. **Verify** — every critical/high candidate faces an adversarial skeptic
   that tries to refute it. Refuted findings are dropped, not softened.
4. **Report** — findings JSON (validated against
   [contracts/finding/schema.json](https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/contracts/finding/schema.json),
   secrets redacted) + a VibeCheck summary in your terminal with one-line fixes.

## Why a skill *and* a CI gate

The skill is the audit you run **while you work**, inside the agent that wrote
the code. The [Action](/docs/github-action) is the gate that runs **whether or
not you remembered** to audit. Both emit the same finding format, so a
dismissed skill finding and an accepted PR finding are the same object —
triage once, trust everywhere.

The skill is MIT, in the main repo under `skill/`. Zero findings is an honest
answer; padding is explicitly banned in its instructions.
