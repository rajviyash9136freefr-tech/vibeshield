import fs from 'node:fs';
import path from 'node:path';
import sharp from 'sharp';

const publicDir = path.resolve('public');
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Light Palette — East Bay (#474C80) and Rum Swizzle (#F8F7E2)
const C = {
  critical: '#b91c1c',
  high: '#c2410c',
  ok: '#15803d',
  dim: 'rgba(71, 76, 128, 0.75)',
  fg: '#474c80',
  hairline: 'rgba(71, 76, 128, 0.20)',
  surface: '#ffffff',
  bg: '#f8f7e2',
  eastBay: '#474C80',
  rumSwizzle: '#F8F7E2',
};

const MONO = "'Geist Mono', ui-monospace, SFMono-Regular, Menlo, monospace";
const SANS = "'Outfit', 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif";

/** The share card. Drawn on a 1200x630 grid; other ratios scale-and-crop. */
function card(w = 1200, h = 630) {
  return `<svg width="${w}" height="${h}" viewBox="0 0 1200 630" preserveAspectRatio="xMidYMid slice" xmlns="http://www.w3.org/2000/svg">
  <rect width="1200" height="630" fill="#f8f7e2"/>

  <!-- Wordmark -->
  <g transform="translate(80, 62)">
    <rect width="206" height="36" rx="18" fill="rgba(71, 76, 128, 0.08)" stroke="rgba(71, 76, 128, 0.25)" stroke-width="1"/>
    <text x="20" y="24" fill="#474c80" font-family="${SANS}" font-size="14" font-weight="800" letter-spacing="1.5">VIBESHIELD</text>
  </g>

  <!-- Headline -->
  <text x="80" y="152" fill="#474c80" font-family="${SANS}" font-size="46" font-weight="800" letter-spacing="-1.2">Stop AI Hallucinations &amp; Leaks Before Merge</text>
  <text x="80" y="192" fill="${C.dim}" font-family="${SANS}" font-size="20">Hallucinated packages · leaked secrets · insecure defaults — caught in sub-second local diffs.</text>

  <!-- Terminal card -->
  <g transform="translate(80, 224)">
    <rect width="1040" height="322" rx="16" fill="#ffffff" stroke="${C.hairline}" stroke-width="1.5"/>
    <circle cx="28" cy="24" r="5" fill="rgba(71, 76, 128, 0.25)"/>
    <circle cx="46" cy="24" r="5" fill="rgba(71, 76, 128, 0.25)"/>
    <circle cx="64" cy="24" r="5" fill="rgba(71, 76, 128, 0.25)"/>
    <text x="92" y="28" fill="${C.dim}" font-family="${MONO}" font-size="13">terminal — vibeshield scan</text>
    <line x1="0" y1="46" x2="1040" y2="46" stroke="${C.hairline}" stroke-width="1"/>

    <text x="28" y="80" fill="#474c80" font-family="${MONO}" font-size="14" font-weight="700">$ vibeshield scan .</text>
    <text x="28" y="108" fill="${C.dim}" font-family="${MONO}" font-size="13.5">  Scanning 18 files (staged diff mode)… done in 0.38s</text>

    <text x="28" y="146" fill="${C.critical}" font-family="${MONO}" font-size="13.5" font-weight="700">  CRITICAL  VS-PKG-001  hallucinated-package</text>
    <text x="28" y="170" fill="${C.fg}" font-family="${MONO}" font-size="13.5">     package.json:14 — fast-parse-utils-v3@2.1.4 (package does not exist in registry)</text>
    <text x="28" y="194" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: replace with native node:util (zero extra dependencies).</text>

    <text x="28" y="232" fill="${C.high}" font-family="${MONO}" font-size="13.5" font-weight="700">  HIGH      VS-SEC-017  hardcoded-secret</text>
    <text x="28" y="256" fill="${C.fg}" font-family="${MONO}" font-size="13.5">     src/agent.ts:42 — OPENAI_API_KEY pasted directly from chat context</text>
    <text x="28" y="280" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: read from process.env.OPENAI_API_KEY and rotate the key now.</text>

    <text x="28" y="308" fill="${C.ok}" font-family="${MONO}" font-size="13">  ✓ 16 files clean · 2 atomic fixes generated · 0 unverified dependencies merged</text>
  </g>

  <!-- Footer facts -->
  <g transform="translate(80, 578)">
    <text x="0" y="0" fill="${C.dim}" font-family="${SANS}" font-size="15">100% local · MIT Open Source · Zero Telemetry · Cursor, Claude Code, Copilot, &amp; Windsurf</text>
  </g>
</svg>`;
}

async function main() {
  await sharp(Buffer.from(card(1200, 630))).png({ quality: 92 }).toFile(path.join(publicDir, 'og.png'));
  console.log('[gen-og] public/og.png (1200x630)');

  await sharp(Buffer.from(card(1280, 640))).png({ quality: 92 }).toFile(path.join(publicDir, 'social-preview.png'));
  console.log('[gen-og] public/social-preview.png (1280x640) → upload to GitHub → Settings → Social preview');
}

main().catch((err) => {
  console.error('[gen-og] Error:', err);
  process.exit(1);
});
