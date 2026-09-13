---
title: Rules & detection
description: "How the ~120-rule AI pack works: the 7 categories, rule IDs, YAML-as-data packs, AI-origin weighting, and the precision bar."
order: 6
---

VibeShield's ruleset is deliberately narrow: **AI failure modes, not general
linting** — ~120 rules versus the thousands in general SAST tools. Every rule
fires at a known way LLM-generated code breaks, and every finding says so.

## Categories & ID ranges

| Prefix | Category | Hunts |
|---|---|---|
| `VS-PKG-###` | hallucinated-package | generated-shape names, remote-code install hooks, unpinned install paths |
| `VS-SEC-###` | hardcoded-secret / insecure-api / insecure-default | training-data echo keys; md5/eval/SQL-concat/ECB; `debug=True`, CORS `*`, `algorithms:["none"]` |
| `VS-LIC-###` | license-missing | stripped headers, copyleft-in-permissive drift |
| `VS-DEP-###` | dependency-risk | transitive blowout, abandoned/forked additions, unpinned ranges |
| `VS-INJ-###` | prompt-injection | agent-instruction payloads in READMEs, issues, CI text, tool configs |

Severity tokens: `critical · high · medium · low · info` — the same on every
surface (CLI, PR comment, JSON, dashboard).

## Rules are data, not code

Packs are versioned YAML ([format](https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/contracts/rulepack.md))
loaded at startup; the core pack ships embedded in the binary under MIT, so the
scanner is fully functional offline. Every published pack is MIT — the
AI-hardening pack lands in this repo as a community-maintained add-on, fetched
only in `--online` mode.

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

Regexes run under Go RE2 — no lookahead/backref tricks — and pack validation
is strict: a malformed rule is a startup error (exit 2), never a silent skip.

## Ranking by AI-likelihood

Findings are ordered by severity × rule confidence × AI-origin attribution.
A confirmed-AI hunk outranks a human-written one at equal severity, because
that's where review attention is scarcest. Attribution never blames a person:
tags name hunks and bots (`copilot-swe-agent`), never humans.

## The precision bar

The golden set: 200 seeded AI PRs (public `fixtures/`) with known answers.
CI runs the whole pack against it; release is gated on **precision ≥ 85 % on
critical/high** and ≤ 5 false positives per 1k diff lines. A rule that can't
clear the bar gets its confidence lowered or it doesn't ship. Measure, then
market.

Browse the live core pack in the repo: `rules/core/` — six files, every rule
with its "why this matters for AI code" text and a one-line fix.
