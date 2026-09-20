---
title: Quickstart
description: "From zero to your first VibeCheck report in about three minutes — the CLI, the pre-commit hook, the GitHub Action, or your coding agent."
order: 1
---

VibeShield audits AI-generated code where it enters your pipeline: your
terminal, the commit, or the pull request. This guide gets you to your first
report in about three minutes. No account is required.

Not installed yet? [Installation](/docs/installation) covers every platform.

## Option A — the CLI (fastest)

```bash
vibeshield init --dry-run    # see exactly what it would write
vibeshield init              # write it, then run the first scan
```

`init` detects your stack, writes a `vibeshield.yml`, a pull-request gate
workflow and a pre-commit hook, then prints the findings it already has — the
instant aha.

From then on:

```bash
vibeshield scan .            # the whole project
vibeshield scan --staged     # only what you are about to commit
vibeshield scan --diff main  # only the lines this branch added
vibeshield doctor            # is everything still wired up?
```

Every command and flag is on the [CLI reference](/docs/cli).

## Option B — the GitHub Action

Create `.github/workflows/vibeshield.yml` in your repo:

```yaml
name: VibeShield
on: [pull_request]
jobs:
  vibeshield:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # diff mode needs the base branch
      - uses: rajviyash9136freefr-tech/vibeshield/action@v3.0.1
```

Push the branch and open a PR. Within seconds VibeShield posts a **VibeCheck
report** on the PR: findings with severity, AI-origin tags, the dependency
delta, and accept/dismiss commands. The default mode is `warn` — it comments
but never blocks your merge until you ask it to.

Full options: [GitHub Action](/docs/github-action).

## Option C — the pre-commit hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v3.0.1
    hooks:
      - id: vibeshield
```

Then `pre-commit install`. Every `git commit` scans staged files locally in
under 1.5 seconds. Rules run fully offline, and `git commit --no-verify` always
works — VibeShield is a gate, never a hostage.

No pre-commit framework? `vibeshield init` writes a plain
`.git/hooks/pre-commit` for you. Details: [pre-commit hook](/docs/pre-commit).

## Option D — your coding agent

Drop the rule block into whichever agent writes your code, so it stops making
the mistakes in the first place:

```bash
vibeshield agents --body > AGENTS.md          # Codex, Cline, Amp, Zed, Aider…
vibeshield agents cursor                      # the Cursor recipe
vibeshield agents                             # all of them
```

For **Claude Code**, install the plugin, which also bundles the parallel
bug-hunting skill:

```bash
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield
claude plugin install vibeshield@vibeshield
```

Then, in any project: `/vibeshield:vibeshield-audit --quick`. See
[the bug-hunting skill](/docs/skill).

## What you get per finding

Every finding carries the same five things, on every surface:

1. **Severity** — CRITICAL / HIGH / MEDIUM / LOW / INFO
2. **Rule ID** — e.g. `VS-PKG-001`, `VS-SEC-017`
3. **Why this matters for AI code** — 1–3 sentences, blaming the pattern, never the person
4. **A one-line fix** — imperative voice
5. **Accept / dismiss** — a dismissal writes an auditable ignore with a required reason

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

## Fixing what it finds

```bash
vibeshield fix . --dry-run    # preview the −/+ diff
vibeshield fix .              # apply, with a prompt per file
vibeshield fix . --yes        # apply everything (agents and CI)
```

Every applied patch is logged to `vibeshield-fixes.log`. Secrets and
prompt-injection findings are never patched mechanically — those need a human.

## Next steps

- Tune the gate: [configuration](/docs/config)
- Understand the checks: [rules & detection](/docs/rules)
- Wire it into CI: [GitHub Action](/docs/github-action)
- Something not working: [troubleshooting](/docs/troubleshooting)
