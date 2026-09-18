#!/usr/bin/env bash
# VibeShield GitHub Action entrypoint (composite step).
#
# Pipeline:
#   1. Resolve a vibeshield binary (scanner_bin -> pinned release -> go build).
#   2. Scan the PR diff (full tree on push without a usable base) as JSON, and
#      emit :error/:warning annotations via --format github.
#   3. Build ONE consolidated VibeCheck PR comment (UIUX.md §4.1 anatomy) and
#      create-or-update it via the GitHub API (curl + jq, both runner-preinstalled).
#   4. Print the README badge snippet + summary to $GITHUB_STEP_SUMMARY.
#   5. Exit per mode (contracts/cli.md): block modes fail with 1 at threshold;
#      config/usage errors fail with 2.
#
# Privacy (PRD §6 decision 5): static analysis on the runner only. This script
# uploads nothing beyond the PR comment body itself; file contents never leave
# the machine.
#
# Local smoke test (Git Bash on Windows): see action/README.md "Testing locally".

set -uo pipefail

# ---------------------------------------------------------------------------
# Small utilities
# ---------------------------------------------------------------------------

err()  { echo "::error::$*" >&2; }
warn() { echo "::warning::$*" >&2; }
note() { echo "::notice::$*"; }
dbg()  { if [ "${RUNNER_DEBUG:-}" = "1" ]; then echo "::debug::$*" >&2; fi; }

# RUNNER_TEMP / GITHUB_* are absent when running the script by hand.
RUNNER_TEMP="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
mkdir -p "$RUNNER_TEMP" 2>/dev/null || RUNNER_TEMP=/tmp

GITHUB_OUTPUT="${GITHUB_OUTPUT:-$RUNNER_TEMP/vs-output.env}"
GITHUB_STEP_SUMMARY="${GITHUB_STEP_SUMMARY:-$RUNNER_TEMP/vs-summary.md}"
GITHUB_WORKSPACE="${GITHUB_WORKSPACE:-$PWD}"
GITHUB_SERVER_URL="${GITHUB_SERVER_URL:-https://github.com}"
GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"
GITHUB_EVENT_NAME="${GITHUB_EVENT_NAME:-}"
GITHUB_REPOSITORY="${GITHUB_REPOSITORY:-}"
GITHUB_SHA="${GITHUB_SHA:-}"

set_output() {
  # set_output <name> <value> — appends to $GITHUB_OUTPUT (created by the
  # runner in CI; falls back to a temp file for local runs).
  printf '%s=%s\n' "$1" "$2" >>"$GITHUB_OUTPUT"
}

is_true() {
  case "$(printf '%s' "${1:-}" | tr '[:upper:]' '[:lower:]')" in
    1|true|yes|on) return 0 ;;
    *) return 1 ;;
  esac
}

# ---------------------------------------------------------------------------
# 0. Inputs + validation
# ---------------------------------------------------------------------------

MODE="${INPUT_MODE:-warn}"
CONFIG="${INPUT_CONFIG:-vibeshield.yml}"
ONLINE="${INPUT_ONLINE:-false}"
SCANNER_BIN="${INPUT_SCANNER_BIN:-}"
VERSION="${INPUT_VERSION:-v2.0.0}"
GH_TOKEN="${INPUT_GITHUB_TOKEN:-${GITHUB_TOKEN:-}}"

case "$MODE" in
  off|warn|block-on-critical|block-on-high+) : ;;
  *)
    err "invalid mode '$MODE' — expected off | warn | block-on-critical | block-on-high+"
    exit 2
    ;;
esac

command -v jq   >/dev/null 2>&1 || { err "jq not found on PATH (preinstalled on runners; locally: winget/brew/apt install jq)"; exit 2; }
command -v curl >/dev/null 2>&1 || { err "curl not found on PATH (preinstalled on runners)"; exit 2; }

cd "$GITHUB_WORKSPACE" || { err "cannot cd to workspace '$GITHUB_WORKSPACE'"; exit 2; }

if is_true "${RUNNER_DEBUG:-}"; then
  # Smoke path: dump the resolved config without leaking the token.
  echo "── VibeShield resolved config (RUNNER_DEBUG) ──────────────────"
  echo "mode=$MODE  online=$ONLINE  version=$VERSION  config=$CONFIG"
  echo "scanner_bin=${SCANNER_BIN:-<auto>}"
  echo "event=${GITHUB_EVENT_NAME:-<none>} repo=${GITHUB_REPOSITORY:-<none>} sha=${GITHUB_SHA:-<none>}"
  echo "workspace=$GITHUB_WORKSPACE  server=$GITHUB_SERVER_URL  api=$GITHUB_API_URL"
  echo "github_token: $([ -n "$GH_TOKEN" ] && echo present || echo 'ABSENT (PR comments skipped; scan + annotations still run)')"
  echo "------------------------------------------------------------------"
fi

# ---------------------------------------------------------------------------
# 1. Binary resolution: scanner_bin -> pinned GitHub release -> go build
# ---------------------------------------------------------------------------

OS="" ARCH=""
case "$(uname -s 2>/dev/null || echo unknown)" in
  Linux*)               OS=linux ;;
  Darwin*)              OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
esac
case "$(uname -m 2>/dev/null || echo unknown)" in
  x86_64|amd64)  ARCH=x86_64 ;;
  arm64|aarch64) ARCH=arm64 ;;
esac

resolve_binary() {
  if [ -n "$SCANNER_BIN" ]; then
    [ -x "$SCANNER_BIN" ] || [ -x "$SCANNER_BIN.exe" ] \
      || { err "scanner_bin '$SCANNER_BIN' is not an executable"; exit 2; }
    printf '%s\n' "$SCANNER_BIN"
    return
  fi

  if [ -n "$OS" ] && [ -n "$ARCH" ]; then
    # Release asset naming: vibeshield-<os>-<arch>.tar.gz (windows: .zip).
    # INTEGRATION NOTE: must match the release-asset workflow in .github/workflows.
    local ext=tar.gz url dl found
    [ "$OS" = windows ] && ext=zip
    url="$GITHUB_SERVER_URL/rajviyash9136freefr-tech/vibeshield/releases/download/$VERSION/vibeshield-$OS-$ARCH.$ext"
    dl="$RUNNER_TEMP/vibeshield-dl.$ext"
    dbg "downloading $url"
    local auth=()
    [ -n "$GH_TOKEN" ] && auth=(-H "Authorization: Bearer $GH_TOKEN")
    if curl -fsSL --max-time 120 "${auth[@]}" -o "$dl" "$url" 2>/dev/null; then
      # Verify the archive against the release's sha256sums.txt before
      # extracting it. Downloading and executing an unverified binary on a
      # runner is the exact supply-chain failure mode this tool exists to catch.
      # A missing sums file is a warning (older releases predate it); a
      # mismatched hash is always fatal.
      local asset="vibeshield-$OS-$ARCH.$ext"
      if curl -fsSL --max-time 60 "${auth[@]}" \
        -o "$RUNNER_TEMP/vs-sums.txt" \
        "$GITHUB_SERVER_URL/rajviyash9136freefr-tech/vibeshield/releases/download/$VERSION/sha256sums.txt" 2>/dev/null; then
        local want got
        want="$(awk -v f="$asset" '{ n=$2; sub(/^\*/, "", n); if (n == f) { print $1; exit } }' "$RUNNER_TEMP/vs-sums.txt")"
        if [ -z "$want" ]; then
          err "sha256sums.txt for $VERSION has no entry for $asset — refusing to run an unverified binary"
          exit 2
        fi
        got="$(sha256sum "$dl" 2>/dev/null | awk '{print $1}')"
        [ -n "$got" ] || got="$(shasum -a 256 "$dl" 2>/dev/null | awk '{print $1}')"
        if [ "$want" != "$got" ]; then
          err "checksum mismatch for $asset (expected $want, got ${got:-none}) — refusing to run"
          exit 2
        fi
        dbg "checksum verified for $asset"
      else
        warn "no sha256sums.txt for $VERSION — continuing without checksum verification"
      fi
      mkdir -p "$RUNNER_TEMP/vibeshield-bin"
      if [ "$ext" = zip ]; then unzip -qo "$dl" -d "$RUNNER_TEMP/vibeshield-bin" 2>/dev/null
      else tar -xzf "$dl" -C "$RUNNER_TEMP/vibeshield-bin" 2>/dev/null; fi
      found="$(find "$RUNNER_TEMP/vibeshield-bin" -type f \( -name 'vibeshield' -o -name 'vibeshield.exe' \) 2>/dev/null | head -1)"
      if [ -n "$found" ]; then
        chmod +x "$found" 2>/dev/null || true
        printf '%s\n' "$found"
        return
      fi
      warn "downloaded $VERSION but the archive contained no vibeshield binary"
    else
      warn "could not download $url — falling back to go build if Go is installed"
    fi
  else
    dbg "no release asset mapping for $(uname -s)/$(uname -m) — trying go build fallback"
  fi

  # Fallback: build from source. Works on a monorepo checkout (scanner/ module)
  # and in local Git Bash once /w/tools/go/bin is on PATH.
  if command -v go >/dev/null 2>&1; then
    if [ -f scanner/go.mod ]; then
      local out="$RUNNER_TEMP/vibeshield"
      [ "$OS" = windows ] && out="$out.exe"
      if go build -o "$out" ./scanner 2>"$RUNNER_TEMP/vs-gobuild.err"; then
        printf '%s\n' "$out"
        return
      fi
      warn "go build ./scanner failed: $(head -3 "$RUNNER_TEMP/vs-gobuild.err" 2>/dev/null | tr '\n' ' ')"
    else
      warn "go build fallback needs a monorepo checkout with scanner/ (not found here)"
    fi
  fi

  err "no vibeshield binary available — set scanner_bin, publish release $VERSION, or run on a Go checkout (see action/README.md)"
  exit 2
}

BIN="$(resolve_binary)"
[ -n "$BIN" ] || { err "binary resolution produced no path"; exit 2; }
dbg "using binary: $BIN"
"$BIN" version >/dev/null 2>&1 || dbg "binary did not answer 'version'"

# ---------------------------------------------------------------------------
# 2. Base ref + scan
# ---------------------------------------------------------------------------

# Diff base: PR head's base sha from the event JSON; push events use
# event.before. Neither present -> full scan. A base that is absent from a
# shallow checkout makes the diff fail; we retry once as a full scan.
BASE=""
EVENT_PATH="${GITHUB_EVENT_PATH:-}"
if [ -n "$EVENT_PATH" ] && [ -f "$EVENT_PATH" ]; then
  BASE="$(jq -r '(.pull_request.base.sha // .number.base.sha // .before // "") | select(. != "")' "$EVENT_PATH" 2>/dev/null || true)"
  [ -z "$BASE" ] && BASE="$(jq -r '.base.sha // empty' "$EVENT_PATH" 2>/dev/null || true)"
fi
[ -z "$BASE" ] && [ -n "${GITHUB_BASE_REF:-}" ] && BASE="$GITHUB_BASE_REF"
dbg "diff base: ${BASE:-<full scan>}"

SCAN_PRE_ARGS=(scan)
SCAN_POST_ARGS=()
if [ -n "$BASE" ]; then SCAN_PRE_ARGS+=(--diff "$BASE"); fi
# Pass --config only when the file exists, or when the user asked for a
# non-default path (a missing explicit config is a contract exit-2 error).
if [ -f "$CONFIG" ] || [ "$CONFIG" != "vibeshield.yml" ]; then
  SCAN_PRE_ARGS+=(--config "$CONFIG")
fi
is_true "$ONLINE" && SCAN_PRE_ARGS+=(--online)
SCAN_PRE_ARGS+=(--mode "$MODE")
SCAN_POST_ARGS+=(--no-color)

# run_scan <format> <output-file> <err-file> — exit code via $? (set with ||).
run_scan() {
  local fmt="$1" out="$2" err="$3"
  shift 3
  "$BIN" "${SCAN_PRE_ARGS[@]}" "$@" "${SCAN_POST_ARGS[@]}" --format "$fmt" >"$out" 2>"$err"
}

REPORT_JSON="$RUNNER_TEMP/vibeshield-report.json"
SCAN_ERR="$RUNNER_TEMP/vs-scan.err"

echo "Running: $BIN ${SCAN_PRE_ARGS[*]} ${SCAN_POST_ARGS[*]} --format json"
scan_exit=0
run_scan json "$REPORT_JSON" "$SCAN_ERR" || scan_exit=$?
[ -s "$SCAN_ERR" ] && cat "$SCAN_ERR" >&2

if [ "$scan_exit" -eq 2 ]; then
  err "vibeshield reported a config/usage error (exit 2) — fix vibeshield.yml or the action inputs and re-run"
  exit 2
fi
# Diff failed for a non-block reason (exit >1, or no JSON) with a base set —
# a shallow checkout missing the base sha is the usual cause: retry as full scan.
if { [ "$scan_exit" -gt 1 ] || [ ! -s "$REPORT_JSON" ]; } && [ -n "$BASE" ]; then
  warn "diff scan against ${BASE:0:10} failed (shallow checkout?); retrying as full scan"
  SCAN_PRE_ARGS=(scan)
  if [ -f "$CONFIG" ] || [ "$CONFIG" != "vibeshield.yml" ]; then
    SCAN_PRE_ARGS+=(--config "$CONFIG")
  fi
  is_true "$ONLINE" && SCAN_PRE_ARGS+=(--online)
  SCAN_PRE_ARGS+=(--mode "$MODE")
  BASE=""
  scan_exit=0
  run_scan json "$REPORT_JSON" "$SCAN_ERR" || scan_exit=$?
  [ -s "$SCAN_ERR" ] && cat "$SCAN_ERR" >&2
  if [ "$scan_exit" -eq 2 ]; then exit 2; fi
fi
if [ ! -s "$REPORT_JSON" ] || ! jq -e . "$REPORT_JSON" >/dev/null 2>&1; then
  err "scanner produced no valid JSON (exit $scan_exit) — see scanner output above"
  exit 1
fi

# Annotations: second pass in github format (:error/:warning per finding).
# Blocking decision still comes from the JSON summary + mode, not this exit.
run_scan github /dev/null "$RUNNER_TEMP/vs-annotations.err" || dbg "annotation pass exited $? (informational only)"

# ---------------------------------------------------------------------------
# 3. Summary + mode -> block decision (contracts/cli.md exit codes)
# ---------------------------------------------------------------------------

getsum() { jq -r ".summary.$1 // 0" "$REPORT_JSON"; }
CRIT="$(getsum critical)" HIGH="$(getsum high)" MED="$(getsum medium)"
LOW="$(getsum low)" INFO="$(getsum info)" CLEAN="$(getsum clean_files)"
TOTAL=$((CRIT + HIGH + MED + LOW + INFO))
DUR_MS="$(jq -r '.scan.duration_ms // 0' "$REPORT_JSON")"
DUR="$(awk -v ms="$DUR_MS" 'BEGIN { printf "%.2fs", ms/1000 }')"

BLOCKED=false
case "$MODE" in
  block-on-critical) [ "$CRIT" -gt 0 ] && BLOCKED=true ;;
  block-on-high+)    [ $((CRIT + HIGH)) -gt 0 ] && BLOCKED=true ;;
esac
[ "$scan_exit" -eq 1 ] && BLOCKED=true   # scanner already blocked at its own threshold

set_output critical "$CRIT"; set_output high "$HIGH"; set_output medium "$MED"
set_output low "$LOW";       set_output info "$INFO"; set_output total "$TOTAL"
set_output blocked "$BLOCKED"

# ---------------------------------------------------------------------------
# 4. Badge (UIUX §4.4) — computed before rendering so the comment footer can
#    embed the README snippet (PRD §5.1 badge = viral loop).
# ---------------------------------------------------------------------------

if   [ "$CRIT" -gt 0 ]; then BCOLOR=F87171; BMSG="$TOTAL findings | $CRIT critical"
elif [ "$HIGH" -gt 0 ]; then BCOLOR=FB923C; BMSG="$TOTAL findings | $HIGH high"
elif [ "$TOTAL" -gt 0 ]; then BCOLOR=FBBF24; BMSG="$TOTAL findings"
else BCOLOR=34D399; BMSG="passing | 0 findings"
fi
BADGE_LABEL_V="$(printf '%s' "VibeShield" | sed 's/ /%20/g')"
BADGE_MSG_V="$(printf '%s' "$BMSG" | sed -e 's/ /%20/g' -e 's/|/%7C/g')"
BADGE_URL="https://img.shields.io/badge/$BADGE_LABEL_V-$BADGE_MSG_V-$BCOLOR"
BADGE_MD="[![VibeShield]($BADGE_URL)](https://vibeshield.dev)"

# ---------------------------------------------------------------------------
# 5. VibeCheck comment body (UIUX §4.1) rendered from report JSON via jq
# ---------------------------------------------------------------------------

MARKER='<!-- vibeshield-report:v1 -->'
REPORT_MD="$RUNNER_TEMP/vibeshield-report.md"

write_renderer() {
  cat >"$RUNNER_TEMP/vibeshield-report.jq" <<'JQEOF'
# VibeShield report renderer — contracts/cli.md JSON -> UIUX.md §4.1 comment.
# Severity is conveyed as TEXT (screen-reader rule UIUX §6); the leading emoji
# are decorative. Tone: blame patterns, never people (PRD §5.4).
def emoji: {critical:"🔴",high:"🟠",medium:"🟡",low:"🔵",info:"⚪"}[.] // "⚪";
def order: {critical:0,high:1,medium:2,low:3,info:4}[.] // 5;
def chip($n; $k): if $n > 0 then "\($n) \($k)" else empty end;
def cell: (. // "") | tostring | gsub("|"; "\\|") | gsub("\n"; " ");
def loc: .file + (if .line then ":" + (.line|tostring) else "" end);

. as $r
| ($r.summary // {}) as $s
| ($r.scan // {}) as $sc
| ([chip(($s.critical // 0); "critical"),
    chip(($s.high // 0); "high"),
    chip(($s.medium // 0); "medium"),
    chip(($s.low // 0); "low"),
    chip(($s.info // 0); "info")]
   | if length == 0 then "✅ clean" else join(" · ") end) as $cluster
| ($ENV.VS_BADGE // "") as $badge
| ($ENV.VS_MODE // "warn") as $mode
| [ "🛡 **VibeCheck Report** — \($cluster) · \($s.clean_files // 0) clean · ⏱ \(($sc.duration_ms // 0)/1000)s",
    "",
    "<sub>Severity is shown as text; the emoji are decorative. Analysis ran on this runner — source code never left it.</sub>",
    "",
    ( if ($r.findings // [] | length) == 0 then
        if ($r.dependencies // [] | length) == 0 then
          "_No findings in this change._"
        else
          "_No findings — review the dependency delta below._"
        end
      else
        ($r.findings | sort_by(order(.severity), .rule_id, loc)[]
         | . as $f
         , "**\($f.severity | ascii_upcase)** \($f.severity | emoji) · `\($f.rule_id)` · \($f.category) — \($f.title)",
           "",
           "<details>",
           "<summary><code>\($f | loc)</code> — \($f.message // $f.title | split("\n")[0] | cell)</summary>",
           "",
           "- **Why this matters for AI code:** \($f.message // "—")",
           "- **Location:** `\($f | loc)` · AI-origin: \($f.ai_origin // "unknown")",
           (if $f.fix then "- **Fix:** \($f.fix)" else empty end),
           (if $f.dismiss_hash != null and ($f.dismissable // true) != false
            then "- **Triage:** `/vibeshield accept \($f.dismiss_hash) --reason \"…\"`"
            else empty end),
           (if $f.snippet then "", "```", ($f.snippet | rtrimstr("\n")), "```" else empty end),
           "</details>",
           ""
        )
      end ),
    ( if ($r.dependencies // [] | length) > 0 then
        "**Dependency delta** — what this change added to the tree:",
        "",
        "| Package | Version | Age | Maintainers | Risk | Why |",
        "| --- | --- | --- | --- | --- | --- |",
        ($r.dependencies[]
         | "| `\(.name | cell)` | \(.version | cell) | \((.age_days // "?") | cell)d | \((.maintainers // "?") | cell) | \(.risk // "unknown" | ascii_downcase) \(.risk // "unknown" | emoji) | \(.why | cell) |"),
        ""
      else empty end ),
    "---",
    "",
    "Reply here to triage: `/vibeshield accept <dismiss_hash> --reason \"intentional test fixture\"` — acceptances are audit-logged.",
    "",
    "Docs: https://vibeshield.dev/docs/github-action · Mode: `\($mode)` · Add the README badge: \($badge)",
    "",
    "<!-- vibeshield:blame-patterns-not-people -->"
  ]
| join("\n")
JQEOF
}

render_report() {
  write_renderer
  VS_MODE="$MODE" VS_BADGE="$BADGE_MD" jq -r -f "$RUNNER_TEMP/vibeshield-report.jq" "$REPORT_JSON"
}

if ! render_report >"$REPORT_MD" 2>"$RUNNER_TEMP/vs-render.err"; then
  warn "report renderer failed (summary-only comment fallback): $(head -2 "$RUNNER_TEMP/vs-render.err" 2>/dev/null | tr '\n' ' ')"
  {
    echo "$MARKER"
    echo "🛡 **VibeCheck Report** — $TOTAL finding(s): $CRIT critical · $HIGH high · $MED medium · $LOW low · $INFO info · $CLEAN clean · ⏱ $DUR"
    echo ""
    echo "Docs: https://vibeshield.dev/docs/github-action · Mode: \`$MODE\` · Badge: $BADGE_MD"
  } >"$REPORT_MD"
fi
# Marker first, exactly once.
if ! head -1 "$REPORT_MD" | grep -qF "$MARKER"; then
  { printf '%s\n' "$MARKER"; cat "$REPORT_MD"; } >"$REPORT_MD.tmp" && mv "$REPORT_MD.tmp" "$REPORT_MD"
fi
# GitHub comment body cap is 65536 chars.
if [ "$(wc -c <"$REPORT_MD")" -gt 65000 ]; then
  { head -c 63500 "$REPORT_MD"
    printf '\n\n…truncated — full detail in the step annotations, JSON at `%s`.\n' "$REPORT_JSON"; } >"$REPORT_MD.tmp"
  mv "$REPORT_MD.tmp" "$REPORT_MD"
fi

set_output report_path "$REPORT_MD"
set_output report_json "$REPORT_JSON"

# ---------------------------------------------------------------------------
# 6. Create-or-update the single PR comment (curl + jq)
# ---------------------------------------------------------------------------

post_comment() {
  local pr api auth payload cid http
  pr="$(jq -r '.number // .pull_request.number // empty' "$EVENT_PATH" 2>/dev/null || true)"
  [ -z "$pr" ] && pr="${PR_NUMBER:-}"
  if [ -z "$pr" ]; then dbg "no PR number — skipping comment"; return 0; fi
  if [ -z "$GH_TOKEN" ]; then warn "github_token empty — VibeCheck comment skipped"; return 0; fi
  if [ -z "$GITHUB_REPOSITORY" ]; then warn "GITHUB_REPOSITORY unset — skipping comment"; return 0; fi

  api="$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$pr/comments"
  auth=(-H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json")
  payload="$RUNNER_TEMP/vs-comment.json"
  jq -Rs '{body:.}' "$REPORT_MD" >"$payload" || { warn "could not build comment payload"; return 0; }

  cid="$(curl -sS "${auth[@]}" "$api?per_page=100" 2>/dev/null \
          | jq -r --arg m "$MARKER" '.[] | select((.body // "") | contains($m)) | .id' 2>/dev/null \
          | head -1 || true)"

  if [ -n "$cid" ]; then
    dbg "updating existing comment $cid on PR #$pr"
    http="$(curl -sS -o "$RUNNER_TEMP/vs-api.out" -w '%{http_code}' -X PATCH "${auth[@]}" \
              -H "Content-Type: application/json" -d "@$payload" "$api/$cid" 2>>"$RUNNER_TEMP/vs-api.err" || echo 000)"
  else
    dbg "creating comment on PR #$pr"
    http="$(curl -sS -o "$RUNNER_TEMP/vs-api.out" -w '%{http_code}' -X POST "${auth[@]}" \
              -H "Content-Type: application/json" -d "@$payload" "$api" 2>>"$RUNNER_TEMP/vs-api.err" || echo 000)"
  fi
  case "$http" in
    200|201) note "VibeCheck report posted on PR #$pr" ;;
    403|401) warn "comment API auth/permission HTTP $http — add 'permissions: pull-requests: write' to the workflow token" ;;
    000)     warn "comment API unreachable — scan results in the step log stand" ;;
    *)       warn "comment API returned HTTP $http — scan results in the step log stand" ;;
  esac
}

case "$GITHUB_EVENT_NAME" in
  pull_request|pull_request_target) post_comment ;;
  *) dbg "event '${GITHUB_EVENT_NAME:-<none>}' is not a PR — comment skipped (push runs rely on annotations + summary)" ;;
esac

# ---------------------------------------------------------------------------
# 7. Step summary: VibeCheck recap + README badge (markdown + SVG URL)
# ---------------------------------------------------------------------------

{
  echo "## 🛡 VibeShield — VibeCheck $([ "$BLOCKED" = true ] && echo "⛔ blocked" || echo "✅ passing")"
  echo ""
  echo "- Mode \`$MODE\` · ${BASE:+diff vs \`${BASE:0:12}\`}${BASE:-full-tree scan} · runtime $DUR"
  echo "- **$CRIT critical · $HIGH high · $MED medium · $LOW low · $INFO info** · $CLEAN files clean"
  echo "- New dependencies: $(jq -r '(.dependencies // []) | length' "$REPORT_JSON") flagged"
  echo ""
  echo "### README badge"
  echo ""
  echo '```markdown'
  echo "$BADGE_MD"
  echo '```'
  echo ""
  echo "SVG: \`$BADGE_URL\`"
} >>"$GITHUB_STEP_SUMMARY"

# ---------------------------------------------------------------------------
# 8. Exit semantics (contracts/cli.md)
#    0 = clean or findings under warn/off · 1 = block threshold met · 2 = config
# ---------------------------------------------------------------------------

if [ "$BLOCKED" = true ]; then
  echo "::error::VibeShield blocked this change (mode=$MODE): $CRIT critical / $HIGH high finding(s). See the VibeCheck report — every finding includes a fix and a dismiss path with an audit trail."
  exit 1
fi
[ "$MODE" = "off" ] && note "mode=off — findings reported for visibility only; step passes regardless"
exit 0
