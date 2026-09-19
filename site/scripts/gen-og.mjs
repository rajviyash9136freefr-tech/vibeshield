// Generates the two share images from one source of truth:
//
//   public/og.png             1200x630  — og:image / twitter:image for the site
//   public/social-preview.png 1280x640  — upload to GitHub → Settings → Social preview
//
// The GitHub one is the single highest-value image in the project: without it
// every Slack, Discord, X and LinkedIn link to the repository renders as a grey
// rectangle with a tiny avatar. It cannot be set from a file in the repo — it
// has to be uploaded in repository settings — so this script exists to make
// that upload a two-click job instead of a design task.
//
// The terminal transcript is kept byte-compatible with what `vibeshield scan`
// actually prints, so the share image cannot promise an output format the
// binary does not produce. Severity labels are ASCII ([CRITICAL], not 🔴)
// because emoji rendering through SVG → librsvg → PNG is not reliable.
import fs from 'node:fs';
import path from 'node:path';
import sharp from 'sharp';

const publicDir = path.resolve('public');
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Palette — East Bay and Rum Swizzle theme
const C = {
  critical: '#FF7B72',
  high: '#FFA657',
  ok: '#7EE787',
  dim: 'rgba(248, 247, 226, 0.65)',
  fg: '#F8F7E2',
  hairline: 'rgba(248, 247, 226, 0.15)',
  surface: '#232747',
  bg: '#181b30',
  eastBay: '#474C80',
  rumSwizzle: '#F8F7E2',
};

const MONO = "'Geist Mono', ui-monospace, SFMono-Regular, Menlo, monospace";
const SANS = "'Outfit', 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif";

/** The share card. Drawn on a 1200x630 grid; other ratios scale-and-crop. */
function card(w = 1200, h = 630) {
  return `<svg width="${w}" height="${h}" viewBox="0 0 1200 630" preserveAspectRatio="xMidYMid slice" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <radialGradient id="glow" cx="50%" cy="15%" r="75%">
      <stop offset="0%" stop-color="#474C80" stop-opacity="0.45"/>
      <stop offset="60%" stop-color="#232747" stop-opacity="0.2"/>
      <stop offset="100%" stop-color="#181b30" stop-opacity="0"/>
    </radialGradient>
  </defs>

  <rect width="1200" height="630" fill="#181b30"/>
  <rect width="1200" height="630" fill="url(#glow)"/>

  <!-- Wordmark -->
  <g transform="translate(80, 62)">
    <rect width="206" height="36" rx="18" fill="rgba(71,76,128,0.35)" stroke="rgba(248,247,226,0.3)" stroke-width="1"/>
    <text x="20" y="24" fill="#F8F7E2" font-family="${SANS}" font-size="14" font-weight="800" letter-spacing="1.5">VIBESHIELD</text>
  </g>

  <!-- Headline -->
  <text x="80" y="152" fill="#F8F7E2" font-family="${SANS}" font-size="46" font-weight="800" letter-spacing="-1.2">The bug hunter for AI-generated code</text>
  <text x="80" y="192" fill="${C.dim}" font-family="${SANS}" font-size="20">Hallucinated packages · leaked secrets · insecure defaults — caught before they merge.</text>

  <!-- Terminal card -->
  <g transform="translate(80, 224)">
    <rect width="1040" height="322" rx="16" fill="#232747" stroke="${C.hairline}" stroke-width="1"/>
    <circle cx="28" cy="24" r="5" fill="rgba(248,247,226,0.40)"/>
    <circle cx="46" cy="24" r="5" fill="rgba(248,247,226,0.25)"/>
    <circle cx="64" cy="24" r="5" fill="rgba(248,247,226,0.15)"/>
    <text x="92" y="28" fill="${C.dim}" font-family="${MONO}" font-size="13">bash — vibeshield</text>
    <line x1="0" y1="46" x2="1040" y2="46" stroke="${C.hairline}" stroke-width="1"/>

    <text x="28" y="80" fill="#F8F7E2" font-family="${MONO}" font-size="14" font-weight="700">$ vibeshield scan .</text>
    <text x="28" y="108" fill="${C.dim}" font-family="${MONO}" font-size="13.5">  Scanning 14 files (full mode)… done in 1.2s</text>

    <text x="28" y="146" fill="${C.critical}" font-family="${MONO}" font-size="13.5" font-weight="700">  CRITICAL  VS-PKG-001  hallucinated-package</text>
    <text x="28" y="170" fill="${C.fg}" font-family="${MONO}" font-size="13.5">     package.json:8 — fast-parse-utils-v3@2.1.4 (9 days old, 1 maintainer)</text>
    <text x="28" y="194" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: replace with node:util (12 lines, zero dependencies).</text>

    <text x="28" y="232" fill="${C.high}" font-family="${MONO}" font-size="13.5" font-weight="700">  HIGH      VS-SEC-017  hardcoded-secret</text>
    <text x="28" y="256" fill="${C.fg}" font-family="${MONO}" font-size="13.5">     src/agent.ts:41 — OPENAI_API_KEY pasted from chat context</text>
    <text x="28" y="280" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: read process.env.OPENAI_API_KEY and rotate the key now.</text>

    <text x="28" y="308" fill="${C.ok}" font-family="${MONO}" font-size="13">  ✓ 12 files clean · 2 findings · 0 leaks merged to main</text>
  </g>

  <!-- Footer facts -->
  <g transform="translate(80, 578)">
    <text x="0" y="0" fill="${C.dim}" font-family="${SANS}" font-size="15">100% local · MIT · one static binary · GitHub Action, pre-commit hook and CLI</text>
  </g>
</svg>`;
}


async function main() {
  // og:image — the 1200x630 the site's <meta> tags point at.
  await sharp(Buffer.from(card(1200, 630))).png({ quality: 92 }).toFile(path.join(publicDir, 'og.png'));
  console.log('[gen-og] public/og.png (1200x630)');

  // GitHub social preview — 1280x640. The card is drawn on a 1200x630 grid and
  // scaled to fill, so there are no letterbox bars: the extra 10px of height is
  // cropped from the top and bottom margins, which are empty by design.
  await sharp(Buffer.from(card(1280, 640))).png({ quality: 92 }).toFile(path.join(publicDir, 'social-preview.png'));
  console.log('[gen-og] public/social-preview.png (1280x640) → upload to GitHub → Settings → Social preview');
}

main().catch((err) => {
  console.error('[gen-og] Error:', err);
  process.exit(1);
});
