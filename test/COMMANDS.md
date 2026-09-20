# VibeShield Command & Feature Inventory

Comprehensive inventory checklist of all commands, subcommands, flags, environment variables, config options, and interactive actions.

---

## 1. Top-Level Commands & Aliases

| Command | Aliases | Description | TTY Behavior | Non-TTY Behavior |
|---|---|---|---|---|
| `vibeshield` | — | Bare invocation | Opens Interactive Bubble Tea TUI | Prints usage text to stderr, exits 2 |
| `vibeshield scan` | — | Security & dependency audit | Runs scan, formats output | Runs scan, formats output |
| `vibeshield fix` | — | VibePatch mechanical patch engine | Prompts per-file [y/N/a/s] | Refuses to guess without `--dry-run` or `--yes`, exits 2 |
| `vibeshield init` | — | Project onboarding | Detects stack, writes config/hook/workflow | Writes config/hook/workflow |
| `vibeshield doctor`| — | Diagnostic health check | Formatted checklist | Formatted or JSON checklist |
| `vibeshield search`| `find` | Search rules & catalog | Ranked text or JSON | Ranked text or JSON |
| `vibeshield rules` | `rule` | List rules or read rule | Grouped rules or single rule | Grouped rules or single rule |
| `vibeshield agents`| `agent`| Agent setup recipes | Human guide, table or rule body | Human guide, table or rule body |
| `vibeshield completion` | `completions` | Shell completions | Prints script for bash/zsh/fish/powershell | Prints script |
| `vibeshield ui` | `menu`, `console` | Interactive TUI | Opens Interactive TUI | Opens Interactive TUI (or fallback) |
| `vibeshield version` | `-v`, `-V`, `--version` | Print version info | Prints version, pack info, engine count | Prints version info |
| `vibeshield help` | `-h`, `--help` | Help documentation | Overview or command-specific manual | Overview or command-specific manual |

---

## 2. Command Flags Inventory

### `vibeshield scan [path]`
- `--diff <ref|->`: Diff mode against a git ref or unified diff via stdin `-`
- `--staged`: Pre-commit mode, audits git index staged lines
- `--format <fmt>`: Output format (`pretty` default, `json`, `github`, `sarif`)
- `--config <file>`: Explicit config file path (default: `<path>/vibeshield.yml`, fallback `./vibeshield.yml`)
- `--mode <mode>`: Gate threshold (`off`, `warn`, `block-on-critical`, `block-on-high+`)
- `--rules <dir>`: Load additional YAML rule packs from directory
- `--online`: (Reserved) package intelligence network lookup
- `--max-cols <n>`: Output column width limit (default: 88 in TTY)
- `--no-color`: Strip ANSI colors
- `-v, --verbose`: Print debug and walk details on stderr

### `vibeshield fix [path]`
- `--dry-run`: Preview patch diff, modify zero files
- `--yes`: Apply patches non-interactively (audit-logged to `vibeshield-fixes.log`)
- `--report <file>`: Ingest an existing `--format json` report instead of re-scanning
- `--config <file>`: Config path
- `--rules <dir>`: Extra rule packs
- `--no-color`: Disable colors
- `-v, --verbose`: Verbose stderr logs

### `vibeshield init [path]`
- `--mode <mode>`: Gate mode written to `vibeshield.yml` (default: `warn`)
- `--dry-run`: Preview files to be created
- `--force`: Overwrite existing files
- `--no-hook`: Skip creating `.git/hooks/pre-commit`
- `--no-workflow`: Skip creating `.github/workflows/vibeshield.yml`
- `--no-scan`: Skip first scan execution
- `--no-color`: Disable color

### `vibeshield doctor [path]`
- `--config <file>`: Config path to inspect
- `--format <fmt>`: `pretty` (default) or `json` (`vibeshield.doctor/v1`)
- `--no-color`: Disable color
- `-v, --verbose`: Print all checks including passing checks

### `vibeshield search [query]`
- `--list`: List all catalog entries without filtering
- `--rules`: Filter to rule packs only
- `--agents`: Filter to agent recipes only
- `--limit <n>`: Maximum results (default 20)
- `--format <fmt>`: `pretty` (default) or `json`
- `--no-color`: Disable color

### `vibeshield rules [id]`
- `--limit <n>`: Limit output (default: 0, no limit)
- `--format <fmt>`: `pretty` (default) or `json`
- `--no-color`: Disable color

### `vibeshield agents [name]`
- `--body`: Print only the 5-rule block
- `--markdown`: Print Markdown table of supported agent files

### `vibeshield completion <shell>`
- Shell argument: `bash`, `zsh`, `fish`, `powershell`

---

## 3. Environment Variables
- `NO_COLOR`: Disables all ANSI terminal color sequences
- `VIBESHIELD_VERSION`: Overrides binary version in npm launcher
- `VIBESHIELD_BIN`: Overrides binary path in npm launcher
- `VIBESHIELD_HOME`: Overrides binary cache directory in npm launcher
- `XDG_CACHE_HOME` / `LOCALAPPDATA`: Base cache directories

---

## 4. Configuration Schema (`vibeshield.yml`)
- `mode`: `off` | `warn` | `block-on-critical` | `block-on-high+`
- `languages`: Optional list of languages to restrict
- `ignore`: Array of rule ignores with required `reason` and optional `paths`
- `thresholds`: `new_dependency_max_age_days`
- `notifications`: `slack` (must be environment variable reference)

---

## 5. Interactive UI Actions
- Slash commands: `/scan`, `/diff`, `/fix`, `/help`, `/quit`
- Keybindings:
  - `Enter` / `s`: Trigger full scan
  - `d`: Trigger diff scan
  - `f`: Enter VibePatch diff preview & apply
  - `j` / `k` / `↑` / `↓`: Navigate findings
  - `1`, `2`, `3`, `0`: Filter Critical, High, Medium, All
  - `Enter` on finding: Open single finding detail view
  - `y` on diff preview: Apply patches
  - `n` / `Esc`: Cancel patch or return to previous screen
  - `Ctrl+C`: Graceful immediate exit
