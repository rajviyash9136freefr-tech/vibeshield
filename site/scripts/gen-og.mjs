import fs from 'node:fs';
import path from 'node:path';
import sharp from 'sharp';

const publicDir = path.resolve('public');
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

const svg = `<svg width="1200" height="630" viewBox="0 0 1200 630" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <radialGradient id="heroGlow" cx="50%" cy="25%" r="65%">
      <stop offset="0%" stop-color="#FFFFFF" stop-opacity="0.08"/>
      <stop offset="100%" stop-color="#000000" stop-opacity="0"/>
    </radialGradient>
  </defs>

  <rect width="1200" height="630" fill="#000000"/>
  <rect width="1200" height="630" fill="url(#heroGlow)"/>

  <!-- Border hairline -->
  <rect x="1" y="1" width="1198" height="628" fill="none" stroke="rgba(255,255,255,0.12)" stroke-width="2"/>

  <!-- Brand badge -->
  <g transform="translate(80, 70)">
    <rect width="150" height="34" rx="17" fill="rgba(255,255,255,0.06)" stroke="rgba(255,255,255,0.2)" stroke-width="1"/>
    <text x="24" y="22" fill="#FFFFFF" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="13" font-weight="600" letter-spacing="1">🛡 VIBESHIELD</text>
  </g>

  <!-- Title -->
  <text x="80" y="155" fill="#FFFFFF" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="44" font-weight="700" letter-spacing="-1.5">Security for the code your AI writes</text>
  <text x="80" y="200" fill="rgba(255,255,255,0.65)" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="20" font-weight="400">Audits Cursor, Copilot &amp; Claude Code commits before they hit main.</text>

  <!-- Terminal card -->
  <g transform="translate(80, 240)">
    <rect width="1040" height="320" rx="16" fill="#0A0A0C" stroke="rgba(255,255,255,0.15)" stroke-width="1"/>

    <!-- Terminal header -->
    <circle cx="28" cy="24" r="5" fill="rgba(255,255,255,0.4)"/>
    <circle cx="46" cy="24" r="5" fill="rgba(255,255,255,0.25)"/>
    <circle cx="64" cy="24" r="5" fill="rgba(255,255,255,0.15)"/>
    <text x="92" y="28" fill="rgba(255,255,255,0.5)" font-family="'Geist Mono', monospace" font-size="13">bash — vibeshield scan</text>
    <line x1="0" y1="44" x2="1040" y2="44" stroke="rgba(255,255,255,0.1)" stroke-width="1"/>

    <!-- Terminal body -->
    <text x="28" y="78" fill="#FFFFFF" font-family="'Geist Mono', monospace" font-size="14" font-weight="700">$ npx vibeshield scan</text>
    <text x="28" y="106" fill="rgba(255,255,255,0.65)" font-family="'Geist Mono', monospace" font-size="13.5">  Scanning 14 changed files (diff mode)… done in 1.2s</text>

    <!-- Finding 1 -->
    <text x="28" y="142" fill="#FFFFFF" font-family="'Geist Mono', monospace" font-size="13.5" font-weight="700">  [CRITICAL] VS-PKG-001  hallucinated-package</text>
    <text x="28" y="166" fill="rgba(255,255,255,0.85)" font-family="'Geist Mono', monospace" font-size="13.5">     fast-parse-utils-v3@2.1.4 — registered 9 days ago, 1 maintainer, post-install remote fetch</text>
    <text x="28" y="190" fill="#FFFFFF" font-family="'Geist Mono', monospace" font-size="13.5">     → Fix: replace with node:util (12-line change in src/parse.ts)</text>

    <!-- Finding 2 -->
    <text x="28" y="224" fill="rgba(255,255,255,0.85)" font-family="'Geist Mono', monospace" font-size="13.5" font-weight="700">  [HIGH]     VS-SEC-017  hardcoded-secret</text>
    <text x="28" y="248" fill="rgba(255,255,255,0.85)" font-family="'Geist Mono', monospace" font-size="13.5">     OPENAI_API_KEY echoed in src/lib/agent.ts:41 — likely pasted from AI chat</text>
    <text x="28" y="272" fill="#FFFFFF" font-family="'Geist Mono', monospace" font-size="13.5">     → Fix: move to env, rotate the key now</text>

    <!-- Summary -->
    <text x="28" y="304" fill="#FFFFFF" font-family="'Geist Mono', monospace" font-size="13">  ✓ 11 files clean · 2 findings · 1 dependency added (risky)</text>
  </g>
</svg>`;

async function main() {
  const outputPath = path.join(publicDir, 'og.png');
  await sharp(Buffer.from(svg))
    .png({ quality: 90 })
    .toFile(outputPath);
  console.log(`[gen-og] Generated ${outputPath}`);
}

main().catch((err) => {
  console.error('[gen-og] Error:', err);
  process.exit(1);
});
