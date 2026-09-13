# VibeShield Rule Pack format (v1)

Rule packs are **data, not code**: versioned YAML loaded by the Go scanner at startup.
Core pack ships embedded (MIT). A pack file looks like:

```yaml
schema: vibeshield.rules/v1
id: core
version: 1.0.0
license: MIT
rules:
  - id: VS-SEC-014
    category: insecure-default
    severity: high
    title: "Debug mode enabled in Flask app"
    message: >-
      LLM-generated Flask scaffolding frequently ships app.run(debug=True).
      Debug mode exposes the interactive debugger (remote code execution
      when reachable) and leaks environment contents in error pages.
    fix: "Set debug from an env var: app.run(debug=os.environ.get('FLASK_DEBUG') == '1')"
    languages: [python]
    pattern:
      kind: regex          # regex | literal | structural
      match: 'app\.run\s*\([^)]*debug\s*=\s*True'
      flags: [multiline]
    paths:
      include: ["**/*.py"]
      exclude: ["tests/**", "**/test_*"]
    confidence: 0.9
    references:
      - "https://flask.palletsprojects.com/en/stable/api/#flask.Flask.run"
```

## Field law

| Field | Required | Notes |
|---|---|---|
| `id` | yes | `VS-(PKG\|SEC\|LIC\|DEP\|INJ)-###`, unique across all loaded packs |
| `category` | yes | one of the 7 finding categories (see finding schema) |
| `severity` | yes | critical/high/medium/low/info |
| `title` | yes | ≤80 chars, sentence case, names the pattern not the person |
| `message` | yes | the "why this matters for AI code" text — 1–3 sentences |
| `fix` | yes | one imperative line |
| `autofix` | no | mechanical same-line rewrite (VibePatch) — see below |
| `languages` | yes | from: javascript, typescript, python, go, java, ruby, php, rust, csharp, yaml, generic |
| `pattern.kind` | yes | `regex` (RE2 syntax, no backrefs), `literal`, or `structural` (v1.1 reserved) |
| `pattern.match` | yes for regex/literal | RE2 regex or exact literal |
| `pattern.flags` | no | subset of: multiline, caseless, dotall |
| `paths.include/exclude` | no | doublestar globs, applied before matching |
| `confidence` | no | default 0.8 |
| `references` | no | URLs shown in reports |

## Engine semantics (scanner contract)

### `autofix` (VibePatch) semantics

```yaml
    autofix:
      match: '^(\s*)DEBUG\s*=\s*True\s*$'
      replace: '${1}DEBUG = False'
```

- Applied by `vibeshield fix` only — never by `scan`, the Action, or the hook.
- RE2 `match` runs against the **finding's line**; `replace` is a Go
  `Regexp.ReplaceAllString` template. Group references MUST use the braced
  form `${1}` (a bare `$1DEBUG` expands as group `1DEBUG` → empty, silently
  deleting code; the loader rejects unknown/dangling group names).
- `replace` must keep the rewrite on one line — no `\n`/`\r` (load error).
- The human gate (`y/N` per file, or explicit `--yes` for agents) is part of
  the contract; every applied patch is appended to `vibeshield-fixes.log`.
- Categories `hardcoded-secret` and `prompt-injection` are never patched
  mechanically regardless of rule data.
- `autofix` is optional; a rule without it is "explain + suggest" only.

- A rule **fires** when `pattern.match` matches any line within an included path,
  and the matched region lies inside the scanned diff hunk (diff mode) or file (full mode).
- Each fire produces one Finding with `snippet` = the matched line (secrets redacted by
  the scanner, never by the rule author).
- Multiple matches of one rule in one file = multiple findings (deduped by dismiss_hash).
- Regexes MUST compile under Go RE2: no lookahead, no backreferences, no possessive quantifiers.
- Pack validation is strict: unknown fields, bad globs, non-RE2 patterns = startup error (exit code 2).
