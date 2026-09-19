/**
 * WCAG 2.1 Contrast Ratio Calculator for VibeShield Design Tokens
 */
function luminance(hex) {
  const rgb = hex.replace('#', '').match(/.{2}/g).map((c) => {
    const s = parseInt(c, 16) / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2];
}

function contrast(hex1, hex2) {
  const l1 = luminance(hex1);
  const l2 = luminance(hex2);
  const ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
  return Number(ratio.toFixed(2));
}

const tokens = {
  bg: '#F8F7E2',
  surface: '#EFEDD2',
  ink: '#474C80',
  inkStrong: '#2F3359',
  inkSoft: '#5D6295',
  onInk: '#F8F7E2',
  onInkSoft: '#C8CAE0',
  danger: '#A63D2F',
  warn: '#8A5A00',
  ok: '#2F6B4F',
  dangerOnInk: '#FFB3A7',
  warnOnInk: '#FFD580',
  okOnInk: '#96E6B8',
};

const checks = [
  { fg: 'ink', bg: 'bg', min: 4.5, label: 'Primary text on cream bg' },
  { fg: 'ink', bg: 'surface', min: 4.5, label: 'Primary text on cream surface' },
  { fg: 'inkStrong', bg: 'bg', min: 4.5, label: 'Headings / emphasis on cream bg' },
  { fg: 'inkStrong', bg: 'surface', min: 4.5, label: 'Headings on cream surface' },
  { fg: 'inkSoft', bg: 'bg', min: 4.5, label: 'Secondary text on cream bg' },
  { fg: 'inkSoft', bg: 'surface', min: 4.5, label: 'Secondary text on cream surface' },
  { fg: 'onInk', bg: 'ink', min: 4.5, label: 'Primary text on indigo block' },
  { fg: 'onInkSoft', bg: 'ink', min: 4.5, label: 'Secondary text on indigo block' },
  { fg: 'onInk', bg: 'inkStrong', min: 4.5, label: 'Text on dark footer block' },
  { fg: 'danger', bg: 'bg', min: 4.5, label: 'Critical / error on cream bg' },
  { fg: 'danger', bg: 'surface', min: 4.5, label: 'Critical / error on cream surface' },
  { fg: 'warn', bg: 'bg', min: 4.5, label: 'Warning on cream bg' },
  { fg: 'warn', bg: 'surface', min: 4.5, label: 'Warning on cream surface' },
  { fg: 'ok', bg: 'bg', min: 4.5, label: 'Success on cream bg' },
  { fg: 'ok', bg: 'surface', min: 4.5, label: 'Success on cream surface' },
  { fg: 'dangerOnInk', bg: 'ink', min: 4.5, label: 'Critical on indigo block' },
  { fg: 'warnOnInk', bg: 'ink', min: 4.5, label: 'Warning on indigo block' },
  { fg: 'okOnInk', bg: 'ink', min: 4.5, label: 'Success on indigo block' },
];

console.log('| Pair | Foreground | Background | Ratio | Target | Result |');
console.log('| :--- | :--- | :--- | :--- | :--- | :--- |');

let failed = false;
for (const c of checks) {
  const fgHex = tokens[c.fg];
  const bgHex = tokens[c.bg];
  const ratio = contrast(fgHex, bgHex);
  const pass = ratio >= c.min;
  if (!pass) failed = true;
  console.log(`| ${c.label} | \`${fgHex}\` (${c.fg}) | \`${bgHex}\` (${c.bg}) | **${ratio}:1** | >= ${c.min}:1 | ${pass ? 'PASS' : 'FAIL'} |`);
}

if (failed) {
  console.error('\nContrast check failed!');
  process.exit(1);
} else {
  console.log('\nAll token pairs pass WCAG AA requirements!');
}
