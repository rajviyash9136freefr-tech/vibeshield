---
title: Quickstart
description: "Install VibeShield in 3 minutes: GitHub Action on a PR, pre-commit hook, or CLI scan of a local project."
order: 1
---

VibeShield audits AI-generated code where it enters your pipeline: the commit,
the pull request, or your terminal. This guide gets you from zero to your first
VibeCheck report in about three minutes. No account is required for public repos.

## Option A — GitHub Action (recommended first gate)

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
      - uses: vibeshield/action@v2.0.1
```

Push the branch and open a PR. Within seconds VibeShield posts a
**VibeCheck report** on the PR: findings with severity, AI-origin tags,
dependency delta, and accept/dismiss commands. Default mode is `warn` — it
comments but never blocks your merge until you ask it to.

Full options live on the [GitHub Action page](/docs/github-action).

## Option B — pre-commit hook

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v2.0.1
    hooks:
      - id: vibeshield
```

Then `pre-commit install`. Every `git commit` scans staged files locally in
under 1.5 seconds. Rules run fully offline; `git commit --no-verify` always
works — VibeShield is a gate, never a hostage.

## Option C — CLI

```bash
npx vibeshield scan .        # audit the current project
npx vibeshield init          # write vibeshield.yml + hook + workflow, run first scan
```

`init` detects your frameworks, writes the config, installs the hook and the
workflow file, and immediately prints the findings it already has — the
instant aha. See the [CLI reference](/docs/cli) for `--format json|sarif|github`
and exit-code semantics.

## What you get per finding

Every finding carries the same five things, on every surface:

1. **Severity** — CRITICAL / HIGH / MEDIUM / LOW / INFO
2. **Rule ID** — e.g. `VS-PKG-001` (hallucinated package), `VS-SEC-017` (hardcoded secret)
3. **Why this matters for AI code** — 1–3 sentences, blaming the pattern, never the person
4. **A one-line fix** — imperative voice
5. **Accept / dismiss** — dismissals write an auditable ignore with a required reason

## Next steps

- Tune the gate: [configuration](/docs/config)
- Understand the checks: [rules](/docs/rules)
- Hunt bugs inside your coding agent: [the VibeShield skill](/docs/skill)
