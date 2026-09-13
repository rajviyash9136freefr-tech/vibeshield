# 🛡 VibeShield

**Security & dependency auditing for AI-generated code.**
Your AI agent ships code you didn't write and can't be bothered to read. VibeShield reads it for you — in CI, at the commit hook, before it reaches `main`.

MIT-licensed · free forever · no account, no tiers, no telemetry · your code never leaves your machine.

[![License: MIT](https://img.shields.io/badge/License-MIT-teal.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24-00ADD8.svg)](scanner/go.mod)
[![Scanner](https://img.shields.io/badge/CLI-npx%20vibeshield-CB3837.svg)](#cli)

---

## What it catches

LLM-generated code fails in specific, enumerable ways that classic scanners weren't built to weight:

| Failure mode | Example | Classic scanners |
|---|---|---|
| **Hallucinated packages** | Model invents `fast-parse-utils-v3`; attacker pre-registers it | ❌ No CVE to match — yet |
| **Pasted secrets** | `API_KEY = "sk-..."` echoed from training data | ⚠️ Generic-only |
| **Insecure boilerplate** | `md5` passwords, `eval()` on input, SQL string concat | ⚠️ Buried in lint noise |
| **License stripping** | GPL snippet arrives with the header removed | ❌ Invisible |
| **Insecure defaults** | CORS `*`, `debug=True`, JWT `algorithms:["none"]` | ⚠️ Not weighted for AI code |
| **Prompt-injection changes** | Malicious README tells the agent to open a backdoor | ❌ Nobody covers this |

VibeShield ships 120+ AI-specific rules (the `core` pack, MIT) across 8 languages, and runs three ways: GitHub Action, pre-commit hook, or CLI.

## Install

### GitHub Action (3 minutes, no account)

```yaml
# .github/workflows/vibeshield.yml
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
      - uses: rajviyash9136freefr-tech/vibeshield/action@v1
```

The next PR gets a **VibeCheck report** comment: findings by severity, the AI-origin tag on each hunk, and a suggested fix per finding.

### CLI

```bash
npm install -g vibeshield        # or: npx vibeshield scan .
vibeshield scan .                # scan a project
vibeshield scan --staged         # pre-commit style: only staged changes
```

Go users can skip npm entirely:

```bash
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@latest
```

Or grab a static binary from [Releases](../../releases) — pure Go, no CGO, single file, works offline.

### pre-commit

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v1.0.0
    hooks:
      - id: vibeshield
```

## What output looks like

```
$ vibeshield scan .

  Scanning 31 files (full mode)… done in 0.1s

  🔴 CRITICAL  VS-SEC-001  hardcoded-secret
     AKIA/ASIA-style access key IDs show up in AI output when a model
     echoes an example it memorized — and the example is often a key
     that was real.
     → Fix: Remove the key, rotate it in IAM, and load credentials from
       the environment or an instance profile

  🟠 HIGH      VS-SEC-036  insecure-api
     "SELECT ... WHERE x = '" + value is the classic AI completion for
     parameterized-looking code that is not parameterized at all.
     → Fix: Bind every value with placeholders handled by the driver

  ✓ 28 files clean · 3 findings
```

Secrets in snippets are redacted by the scanner before anything prints or serializes. `--format json` emits the [`contracts/finding/schema.json`](contracts/finding/schema.json) envelope; `--format github` emits PR annotations; exit codes follow [`contracts/cli.md`](contracts/cli.md) (0 = pass, 1 = block threshold met, 2 = config error).

## Privacy

This is the whole design, not a policy page:

- Static analysis only. **No code or file content is ever uploaded** — from the CLI, the hook, or the Action runner.
- Network calls are strictly opt-in (`--online`, package-intel lookups) and the scanner degrades gracefully offline.
- The core rule packs are versioned YAML you can read and diff in [`rules/`](rules/).

## Repo layout

```
scanner/   Go scanner binary (CLI + pre-commit engine)
rules/     Versioned YAML rule packs (core pack = MIT, embedded in the binary)
action/    GitHub Action composite wrapper (action.yml + entrypoint)
api/       Fastify package-intel + findings API (optional, for dashboards)
site/      Astro marketing site
contracts/ Cross-component JSON schemas — the interface law
fixtures/  Golden test repos with obviously-fake seeded keys
docs/      Product docs
```

## Build from source

```bash
# Scanner (needs Go 1.24+; pure Go, no CGO)
cd scanner && go build ./cmd/vibeshield && go test ./...

# Site (needs Node 22+)
cd site && npm install && npm run build
```

## Contributing

Issues and PRs welcome — especially new rules. A rule is ~15 lines of YAML (see [`contracts/rulepack.md`](contracts/rulepack.md)); the bar is a real AI-code failure mode with a low false-positive rate, blamed on the pattern, never on a person or a vendor. Run `go test ./...` and open a PR against `main`.

## License

[MIT](LICENSE) — the scanner, the action, the hooks, and every rules pack. This is open source first: security buyers and weekend shippers alike should run what they can read.
