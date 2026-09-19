---
title: Rules & detection
description: "How the 122-rule AI pack works: the 6 categories, rule IDs, YAML-as-data packs, AI-origin weighting, and what the engine can actually evaluate."
order: 6
---

VibeShield's ruleset is deliberately narrow: **AI failure modes, not general
linting** — 122 rules versus the thousands in general SAST tools. Every rule
fires at a known way LLM-generated code breaks, and every finding says so.

## Categories & ID ranges

| Prefix | Category | Hunts |
|---|---|---|
| `VS-PKG-###` | hallucinated-package | generated-shape names, remote-code install hooks, unpinned install paths |
| `VS-SEC-###` | hardcoded-secret | keys and tokens echoed from training data or chat |
| `VS-SEC-###` | insecure-api | `eval`, md5 passwords, SQL string concatenation, ECB mode |
| `VS-SEC-###` | insecure-default | `debug=True`, CORS `*`, `algorithms:["none"]`, TLS verification off |
| `VS-DEP-###` | dependency-risk | unpinned ranges, single-maintainer additions, install hooks that `curl \| bash` |
| `VS-LIC-###` | license-missing | stripped SPDX headers, copyleft-in-permissive drift |
| `VS-INJ-###` | prompt-injection | agent-instruction payloads in READMEs, issues, CI text, tool configs |

Severity tokens: `critical · high · medium · low · info` — the same on every
surface (CLI, console, PR comment, JSON, SARIF).

Read the whole pack from the terminal:

```bash
vibeshield rules                  # grouped by category
vibeshield rules VS-SEC-017       # one rule, in full
vibeshield rules --format json    # vibeshield.search/v1
```

## Coverage

The pack is deliberately larger than the engine:

```console
$ vibeshield version
vibeshield 3.0.0
rule packs: core 2.0.0 (MIT, 122 rules)
engine:     117 active · 5 reserved (structural — pending the package-intel model)
```

The five reserved rules load and validate so packs stay portable across
versions, but the matcher skips them — they depend on the package-intel model
that `--online` will eventually provide. `vibeshield rules` marks each one
`RESERVED` rather than letting you search for a rule that cannot fire, and
`scripts/audit-contract.mjs` surfaces the split in CI so it cannot quietly
widen.

This matters in practice: a fixture built around a hallucinated package will
come back clean, because that is a reserved rule. The scan is telling the truth
about what it checked.

## Rules are data, not code

Packs are versioned YAML ([format](https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/contracts/rulepack.md))
loaded at startup; the core pack ships embedded in the binary under MIT, so the
scanner is fully functional offline. Adding a rule needs no Go change.

```yaml
- id: VS-SEC-014
  category: insecure-default
  severity: high
  title: "Debug mode enabled in Flask app"
  message: >-
    LLM-generated Flask scaffolding frequently ships app.run(debug=True).
    Debug mode exposes the interactive debugger and leaks environment
    contents in error pages.
  fix: "Set debug from an env var: app.run(debug=os.environ.get('FLASK_DEBUG') == '1')"
  languages: [python]
  pattern:
    kind: regex
    match: 'app\.run\s*\([^)]*debug\s*=\s*True'
    flags: [multiline]
  confidence: 0.9
```

Regexes run under Go RE2 — no lookahead or backreference tricks — and pack
validation is strict: a malformed rule is a startup error (exit 2), never a
silent skip. A pack is data, never something the scanner executes.

## Ranking by AI-likelihood

Findings are ordered by severity × rule confidence × AI-origin attribution. A
confirmed-AI hunk outranks a human-written one at equal severity, because that
is where review attention is scarcest. Attribution never blames a person: tags
name hunks and bots (`copilot-swe-agent`), never humans.

## The precision bar

The golden set is the `fixtures/` corpus with known answers. CI runs the pack
against it, and a rule that cannot clear the bar gets its confidence lowered or
it does not ship. Measure, then market.

Browse the live core pack in the repo: `rules/core/` — six files, every rule
with its "why this matters for AI code" text and a one-line fix.

## Writing a rule

Rules are the easiest thing to contribute to this project — they are YAML, and
they need no Go. See [CONTRIBUTING.md](https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/CONTRIBUTING.md)
for the workflow, and remember to run `node scripts/sync-rules.mjs` afterwards:
the scanner embeds a copy of the packs, and CI fails if the two drift.
