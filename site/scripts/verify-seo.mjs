import { readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';

const distDir = join(process.cwd(), 'dist');

const pages = [
  { path: 'index.html', route: '/' },
  { path: 'privacy/index.html', route: '/privacy' },
  { path: 'terms/index.html', route: '/terms' },
  { path: 'security/index.html', route: '/security' },
  { path: 'install/index.html', route: '/install' },
  { path: '404.html', route: '/404' },
];

let allPassed = true;

for (const p of pages) {
  const filePath = join(distDir, p.path);
  if (!existsSync(filePath)) {
    console.error(`Missing file: ${p.path}`);
    allPassed = false;
    continue;
  }
  const html = readFileSync(filePath, 'utf8');

  // Title
  const titleMatch = html.match(/<title>([^<]+)<\/title>/);
  const title = titleMatch ? titleMatch[1] : null;

  // Description
  const descMatch = html.match(/<meta name="description" content="([^"]+)"/);
  const description = descMatch ? descMatch[1] : null;

  // Canonical
  const canonMatch = html.match(/<link rel="canonical" href="([^"]+)"/);
  const canonical = canonMatch ? canonMatch[1] : null;

  // H1 count
  const h1Matches = html.match(/<h1[\s>]/g) || [];

  // Robots
  const robotsMatch = html.match(/<meta name="robots" content="([^"]+)"/);
  const robots = robotsMatch ? robotsMatch[1] : null;

  // JSON-LD
  const jsonLdMatch = html.match(/<script type="application\/ld\+json">([\s\S]*?)<\/script>/);
  let jsonLdValid = false;
  let schemaCount = 0;
  if (jsonLdMatch) {
    try {
      const parsed = JSON.parse(jsonLdMatch[1]);
      jsonLdValid = true;
      schemaCount = Array.isArray(parsed) ? parsed.length : 1;
    } catch (e) {
      console.error(`Invalid JSON-LD on ${p.route}:`, e.message);
      allPassed = false;
    }
  }

  console.log(`\n=== Route: ${p.route} ===`);
  console.log(`Title (${title ? title.length : 0} chars): ${title}`);
  console.log(`Description (${description ? description.length : 0} chars): ${description}`);
  console.log(`Canonical: ${canonical}`);
  console.log(`Robots: ${robots}`);
  console.log(`H1 count: ${h1Matches.length} (${h1Matches.length === 1 ? 'PASS' : 'FAIL'})`);
  console.log(`JSON-LD valid: ${jsonLdValid ? `YES (${schemaCount} schemas)` : 'NO'}`);

  if (h1Matches.length !== 1) allPassed = false;
  if (!title || !description || !canonical) allPassed = false;
  if (p.route === '/404' && (!robots || !robots.includes('noindex'))) allPassed = false;
}

// Check sitemap and robots
const sitemapExists = existsSync(join(distDir, 'sitemap-index.xml')) && existsSync(join(distDir, 'sitemap-0.xml'));
const robotsExists = existsSync(join(distDir, 'robots.txt'));
const securityTxtExists = existsSync(join(distDir, '.well-known/security.txt'));

console.log(`\n=== Static Assets in dist ===`);
console.log(`sitemap-index.xml + sitemap-0.xml: ${sitemapExists ? 'PASS' : 'FAIL'}`);
console.log(`robots.txt: ${robotsExists ? 'PASS' : 'FAIL'}`);
console.log(`.well-known/security.txt: ${securityTxtExists ? 'PASS' : 'FAIL'}`);

if (!allPassed || !sitemapExists || !robotsExists || !securityTxtExists) {
  process.exit(1);
} else {
  console.log('\nAll SEO and metadata verifications PASSED!');
}
