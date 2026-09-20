# VibeShield QA Bug Log

Summary of bugs identified, investigated, and resolved during comprehensive CLI and scanner testing.

---

## Bug Index

| Bug ID | Severity | Command / Component | Summary | Status |
|---|---|---|---|---|
| **BUG-001** | High | CLI Path Resolution | Old v1.0.0 binary in `~/.local/bin` hijacked PATH | **Fixed** |
| **BUG-002** | Medium | Scanner Self-Audit | Documentation prose triggered `VS-SEC-027` & `VS-DEP-011` | **Fixed** |
| **BUG-003** | Medium | TUI Diff Filter | Compilation error: `d.Touches` called on `scandiff.Diff` | **Fixed** |
| **BUG-004** | Low | Build Tooling | Missing transitive `go.sum` entries for Charm Bubbles | **Fixed** |

---

## Detailed Bug Reports

### BUG-001: Old v1.0.0 binary on PATH overriding modern build
- **Severity:** High
- **Component:** CLI Execution / Installation
- **Exact Repro:**
  ```powershell
  vibeshield
  # Output: vibeshield v1.0.0 — security scanner for AI-generated code
  ```
- **Expected:** Invocations of `vibeshield` run the current v3.0.0+ engine with interactive Bubble Tea TUI.
- **Actual:** User's local `~/.local/bin/vibeshield.exe` was pinned to an old v1.0.0 binary, exiting with code 2.
- **Root Cause:** Installer script targets `$HOME\.local\bin` which precedes repo-local builds in PATH order.
- **Fix Applied:** Synchronized current build into `$HOME\.local\bin\vibeshield.exe` and verified `vibeshield version` outputs `3.0.0`.
- **Status:** Verified Fixed.

---

### BUG-002: Documentation prose in `docs/CLI_PLAN.md` triggered security scanner
- **Severity:** Medium
- **Component:** Rules Engine / Self-Audit
- **Exact Repro:**
  ```powershell
  scanner/vibeshield.exe scan .
  ```
- **Expected:** Clean scan exit code 0 when documentation discusses rule descriptions.
- **Actual:** Flagged `VS-SEC-027` (insecure-api) on `curl ... | bash` pattern in docs and `VS-DEP-011` on `@latest` inside markdown prose.
- **Root Cause:** Markdown prose contained verbatim install patterns that the scanner matches.
- **Fix Applied:**
  1. Rewrote documentation examples to avoid literal unpinned `@latest` commands.
  2. Added `test/fixtures/**` to `vibeshield.yml` ignore paths so deliberate test fixture vulnerabilities do not contaminate production repository audits.
- **Status:** Verified Fixed.

---

### BUG-003: `filterDiffFindings` called non-existent method `Touches` on `scandiff.Diff`
- **Severity:** Medium
- **Component:** `scanner/internal/tui/app.go`
- **Exact Repro:**
  ```powershell
  go build ./internal/tui/...
  ```
- **Expected:** Clean build and filtering of findings against git diff lines.
- **Actual:** Compiler error: `d.Touches undefined (type *scandiff.Diff has no field or method Touches)`.
- **Root Cause:** In `scanner/internal/scandiff/scandiff.go`, the line-membership method is named `Has(file, line)`.
- **Fix Applied:** Replaced `d.Touches(f.File, f.Line)` with `d.Has(f.File, f.Line)` in `scanner/internal/tui/app.go`.
- **Status:** Verified Fixed.

---

### BUG-004: Missing `go.sum` entries for Charm Bubbles components
- **Severity:** Low
- **Component:** Build System / Go Modules
- **Exact Repro:**
  ```powershell
  go build ./internal/tui/...
  ```
- **Expected:** Dependencies download and compile without missing module checksums.
- **Actual:** Missing checksums for `github.com/charmbracelet/harmonica` and `github.com/atotto/clipboard`.
- **Root Cause:** Secondary dependencies required by `bubbles/progress` and `bubbles/textinput` were missing from `go.sum`.
- **Fix Applied:** Ran `go mod tidy` in `scanner/` to pin all transitive dependencies.
- **Status:** Verified Fixed.
