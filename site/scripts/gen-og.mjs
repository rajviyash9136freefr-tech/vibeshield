import fs from 'node:fs';
import path from 'node:path';
import sharp from 'sharp';

const publicDir = path.resolve('public');
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Strict Design Tokens: East Bay (#474C80) & Rum Swizzle (#F8F7E2)
const C = {
  bg: '#F8F7E2',
  surface: '#EFEDD2',
  ink: '#474C80',
  inkStrong: '#2F3359',
  inkSoft: '#5D6295',
  danger: '#A63D2F',
  warn: '#8A5A00',
  ok: '#2F6B4F',
  border: 'rgba(71, 76, 128, 0.22)',
};

const MONO = "'Geist Mono', ui-monospace, SFMono-Regular, Menlo, monospace";
const SANS = "'Outfit', 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif";

/** The share card. Drawn on a 1200x630 grid; other ratios scale-and-crop. */
function card(w = 1200, h = 630) {
  return `<svg width="${w}" height="${h}" viewBox="0 0 1200 630" preserveAspectRatio="xMidYMid slice" xmlns="http://www.w3.org/2000/svg">
  <rect width="1200" height="630" fill="${C.bg}"/>

  <!-- Wordmark -->
  <g transform="translate(80, 62)">
    <rect width="206" height="36" rx="18" fill="${C.surface}" stroke="${C.border}" stroke-width="1.5"/>
    <text x="20" y="24" fill="${C.ink}" font-family="${SANS}" font-size="14" font-weight="800" letter-spacing="1.5">VIBESHIELD</text>
  </g>

  <!-- Headline -->
  <text x="80" y="152" fill="${C.ink}" font-family="${SANS}" font-size="44" font-weight="800" letter-spacing="-1.2">AUDIT AI CODE BEFORE MERGE</text>
  <text x="80" y="192" fill="${C.inkSoft}" font-family="${SANS}" font-size="20">Catches hallucinated packages (slopsquatting), leaked API keys &amp; insecure scaffolding.</text>

  <!-- Terminal card in surface -->
  <g transform="translate(80, 224)">
    <rect width="1040" height="322" rx="16" fill="${C.surface}" stroke="${C.border}" stroke-width="1.5"/>
    <circle cx="28" cy="24" r="5" fill="${C.inkSoft}"/>
    <circle cx="46" cy="24" r="5" fill="${C.inkSoft}"/>
    <circle cx="64" cy="24" r="5" fill="${C.inkSoft}"/>
    <text x="92" y="28" fill="${C.inkSoft}" font-family="${MONO}" font-size="13">terminal — vibeshield scan .</text>
    <line x1="0" y1="46" x2="1040" y2="46" stroke="${C.border}" stroke-width="1"/>

    <text x="28" y="80" fill="${C.ink}" font-family="${MONO}" font-size="14" font-weight="700">$ vibeshield scan .</text>
    <text x="28" y="108" fill="${C.inkSoft}" font-family="${MONO}" font-size="13.5">  Scanning 18 files (staged diff mode)… done in 0.38s</text>

    <text x="28" y="146" fill="${C.danger}" font-family="${MONO}" font-size="13.5" font-weight="700">  CRITICAL  VS-PKG-001  hallucinated-package</text>
    <text x="28" y="170" fill="${C.ink}" font-family="${MONO}" font-size="13.5">     package.json:14 — fast-parse-utils-v3@2.1.4 (package does not exist in registry)</text>
    <text x="28" y="194" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: replace with native node:util (zero extra dependencies).</text>

    <text x="28" y="232" fill="${C.warn}" font-family="${MONO}" font-size="13.5" font-weight="700">  HIGH      VS-SEC-017  hardcoded-secret</text>
    <text x="28" y="256" fill="${C.ink}" font-family="${MONO}" font-size="13.5">     src/agent.ts:42 — OPENAI_API_KEY pasted directly from chat context</text>
    <text x="28" y="280" fill="${C.ok}" font-family="${MONO}" font-size="13.5">     → Fix: read from process.env.OPENAI_API_KEY and rotate the key now.</text>

    <text x="28" y="308" fill="${C.ok}" font-family="${MONO}" font-size="13">  ✓ 16 files clean · 2 atomic fixes generated · 0 unverified dependencies merged</text>
  </g>

  <!-- Footer facts -->
  <g transform="translate(80, 578)">
    <text x="0" y="0" fill="${C.inkSoft}" font-family="${SANS}" font-size="15">100% local · MIT Open Source · Zero Network Egress · Cursor, Claude Code, Copilot, &amp; Windsurf</text>
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
