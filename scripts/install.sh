#!/usr/bin/env bash
# VibeShield one-command installer (Git Bash / Linux / macOS).
#
#   curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash
#
# What it does, in order:
#   1. Downloads the right static binary from GitHub Releases (checksum
#      verified against sha256sums.txt from the same release).
#   2. Installs it to ~/.local/bin/vibeshield (override: VIBESHIELD_INSTALL_DIR).
#   3. Installs the vibeshield-audit skill into every coding agent it detects
#      (Claude Code, Cursor, Codex CLI, Antigravity/Gemini) — or the one you
#      name with --agent <claude|cursor|codex|antigravity|none>.
#
# Flags:  --agent <name>   --version <tag>   --no-skill   --help
# No sudo, no PATH edits written behind your back, and nothing here sends
# your code anywhere: the only network is GitHub Releases over HTTPS.
set -euo pipefail

REPO="rajviyash9136freefr-tech/vibeshield"
VERSION="${VIBESHIELD_VERSION:-v1.0.0}"
AGENT="auto"
WANT_SKILL=1
INSTALL_DIR="${VIBESHIELD_INSTALL_DIR:-$HOME/.local/bin}"

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [ $# -gt 0 ]; do
  case "$1" in
    --agent)   AGENT="$2"; shift 2 ;;
    --version) VERSION="$2"; shift 2 ;;
    --no-skill) WANT_SKILL=0; shift ;;
    -h|--help) usage ;;
    *) echo "vibeshield-install: unknown flag $1 (see --help)" >&2; exit 2 ;;
  esac
done

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  vs_os=linux ;;
  darwin) vs_os=darwin ;;
  msys*|mingw*|cygwin*) vs_os=windows ;;
  *) echo "vibeshield-install: unsupported OS $OS — build from source: go install github.com/$REPO/scanner/cmd/vibeshield@latest" >&2; exit 1 ;;
esac
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) vs_arch=x86_64 ;;
  arm64|aarch64) vs_arch=aarch64 ;;
  *) echo "vibeshield-install: unsupported architecture $ARCH" >&2; exit 1 ;;
esac

EXT=tar.gz; BIN=vibeshield
[ "$vs_os" = windows ] && { EXT=zip; BIN=vibeshield.exe; }
ASSET="vibeshield-${vs_os}-${vs_arch}.${EXT}"
BASE="https://github.com/$REPO/releases/download/$VERSION"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "→ fetching $ASSET ($VERSION)"
if ! curl -fsSL "$BASE/$ASSET" -o "$TMP/$ASSET"; then
  echo "vibeshield-install: download failed. Is $VERSION released? List: https://github.com/$REPO/releases" >&2
  echo "  Offline / from source:  go install github.com/$REPO/scanner/cmd/vibeshield@latest" >&2
  exit 1
fi
curl -fsSL "$BASE/sha256sums.txt" -o "$TMP/sha256sums.txt" || {
  echo "vibeshield-install: could not fetch checksums — refusing to run an unverified binary" >&2; exit 1; }

WANT=$(awk -v a="$ASSET" '$2==a || $2=="*"a {print $1; exit}' "$TMP/sha256sums.txt")
if [ -n "$WANT" ]; then
  GOT=$( (sha256sum "$TMP/$ASSET" 2>/dev/null || shasum -a 256 "$TMP/$ASSET") | awk '{print $1}')
  [ "$GOT" = "$WANT" ] || { echo "vibeshield-install: checksum mismatch for $ASSET — refusing" >&2; exit 1; }
  echo "→ checksum verified ✓"
else
  echo "vibeshield-install: $ASSET not in sha256sums.txt — refusing" >&2; exit 1
fi

mkdir -p "$TMP/x"
if [ "$EXT" = zip ]; then
  unzip -o -q "$TMP/$ASSET" -d "$TMP/x"
else
  tar -xzf "$TMP/$ASSET" -C "$TMP/x"
fi
SRC=$(find "$TMP/x" -name "$BIN" -type f | head -1)
[ -n "$SRC" ] || { echo "vibeshield-install: binary missing from archive" >&2; exit 1; }

mkdir -p "$INSTALL_DIR"
mv "$SRC" "$INSTALL_DIR/$BIN"
chmod +x "$INSTALL_DIR/$BIN"
echo "→ installed: $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) : ;;
  *) echo "  (add it to PATH:  export PATH=\"$INSTALL_DIR:\$PATH\")" ;;
esac
"$INSTALL_DIR/$BIN" version | sed 's/^/  /'

# ---------- skill install ----------
install_skill() { # $1=dest dir, $2=label
  local dest="$1" label="$2"
  mkdir -p "$dest"
  rm -rf "$dest/vibeshield-audit"
  cp -r "$TMP/skill/skills/vibeshield-audit" "$dest/vibeshield-audit"
  echo "→ skill installed for $label → $dest/vibeshield-audit"
}

if [ "$WANT_SKILL" = 1 ] && [ "$AGENT" != none ]; then
  echo "→ fetching the audit skill (SKILL.md)"
  SK="https://raw.githubusercontent.com/$REPO/$VERSION"
  # GitHub release pages don't carry the repo tree; pull the two files from
  # the tag ref instead (raw content — the skill is plain markdown + one
  # dependency-free validator).
  mkdir -p "$TMP/skill/skills/vibeshield-audit/scripts"
  for f in SKILL.md patterns.md; do
    curl -fsSL "$SK/skill/skills/vibeshield-audit/$f" -o "$TMP/skill/skills/vibeshield-audit/$f" ||
      { echo "vibeshield-install: could not fetch skill/$f (installed binary anyway; add the skill later — see README)"; WANT_SKILL=0; };
  done
  if [ "$WANT_SKILL" = 1 ]; then
    curl -fsSL "$SK/skill/skills/vibeshield-audit/scripts/validate.mjs" -o "$TMP/skill/skills/vibeshield-audit/scripts/validate.mjs" || true
  fi
fi

if [ "$WANT_SKILL" = 1 ] && [ "$AGENT" != none ]; then
  detect_and_install() {
    case "$AGENT" in
      claude) install_skill "$HOME/.claude/skills" "Claude Code" ;;
      cursor) install_skill "$PWD/.cursor/skills" "Cursor (this project)" ;;
      codex)  install_skill "$HOME/.codex/skills" "Codex" ;;
      antigravity) install_skill "$HOME/.gemini/antigravity/skills" "Antigravity (global)" ;;
      auto)
        done_any=0
        [ -d "$HOME/.claude" ] && { install_skill "$HOME/.claude/skills" "Claude Code"; done_any=1; }
        [ -d "$HOME/.codex" ]  && { install_skill "$HOME/.codex/skills" "Codex"; done_any=1; }
        [ -d "$HOME/.gemini/antigravity" ] && { install_skill "$HOME/.gemini/antigravity/skills" "Antigravity"; done_any=1; }
        [ -d "$PWD/.cursor" ]  && { install_skill "$PWD/.cursor/skills" "Cursor"; done_any=1; }
        [ -d "$PWD/.agents" ]  && { install_skill "$PWD/.agents/skills" "cross-agent (.agents)"; done_any=1; }
        [ "$done_any" = 1 ] || echo "→ no coding agent detected — skill skipped; see README \"Install into your coding agent\" for one-line commands"
        ;;
      *) echo "vibeshield-install: --agent must be claude|cursor|codex|antigravity|auto|none" >&2; exit 2 ;;
    esac
  }
  detect_and_install
fi

echo
echo "✓ done. Next:  vibeshield scan .   ·   vibeshield fix . --dry-run"
