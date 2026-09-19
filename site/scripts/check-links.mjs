import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs';
import { join } from 'node:path';

const distDir = join(process.cwd(), 'dist');

function getAllHtmlFiles(dir) {
  let results = [];
  const list = readdirSync(dir);
  for (const file of list) {
    const fullPath = join(dir, file);
    const stat = statSync(fullPath);
    if (stat && stat.isDirectory()) {
      results = results.concat(getAllHtmlFiles(fullPath));
    } else if (file.endsWith('.html')) {
      results.push(fullPath);
    }
  }
  return results;
}

const htmlFiles = getAllHtmlFiles(distDir);
console.log(`Checking internal links across ${htmlFiles.length} HTML pages in dist...`);

let brokenCount = 0;
const checkedLinks = new Set();

for (const file of htmlFiles) {
  const content = readFileSync(file, 'utf8');
  const hrefMatches = content.matchAll(/href=["']([^"']+)["']/g);

  for (const match of hrefMatches) {
    const href = match[1];
    if (href.startsWith('http://') || href.startsWith('https://') || href.startsWith('mailto:') || href.startsWith('#')) {
      continue;
    }

    const cleanHref = href.split('#')[0].split('?')[0];
    if (!cleanHref) continue;

    const key = `${file} -> ${cleanHref}`;
    if (checkedLinks.has(key)) continue;
    checkedLinks.add(key);

    let targetPath;
    if (cleanHref.startsWith('/vibeshield/')) {
      const rel = cleanHref.slice('/vibeshield/'.length);
      targetPath = rel === '' ? join(distDir, 'index.html') : join(distDir, rel);
    } else if (cleanHref.startsWith('/')) {
      targetPath = join(distDir, cleanHref);
    } else {
      targetPath = join(file, '..', cleanHref);
    }

    // Check direct file, directory with index.html, or file with .html
    let exists = existsSync(targetPath) && statSync(targetPath).isFile();
    if (!exists) {
      const asIndex = join(targetPath, 'index.html');
      if (existsSync(asIndex) && statSync(asIndex).isFile()) {
        exists = true;
      }
    }
    if (!exists) {
      const withHtml = `${targetPath.replace(/\/$/, '')}.html`;
      if (existsSync(withHtml) && statSync(withHtml).isFile()) {
        exists = true;
      }
    }

    if (!exists) {
      console.error(`BROKEN LINK in ${file.slice(distDir.length)}: ${href} (resolved to: ${targetPath})`);
      brokenCount++;
    }
  }
}

if (brokenCount === 0) {
  console.log(`All internal links verified successfully! 0 broken links.`);
} else {
  console.error(`Found ${brokenCount} broken links!`);
  process.exit(1);
}
