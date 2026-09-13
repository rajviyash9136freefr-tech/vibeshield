/**
 * Base-aware href helper. The site is hosted on GitHub Pages under
 * /vibeshield/ (Astro `base`), so every internal root-relative link must be
 * prefixed with import.meta.env.BASE_URL at build time. External URLs,
 * mailto:, and bare #hashes pass through untouched.
 */
export function baseHref(path: string): string {
  if (!path || /^(https?:|mailto:|#)/.test(path)) return path;
  if (!path.startsWith('/')) return path;
  const base = (import.meta.env.BASE_URL ?? '/').replace(/\/+$/, '');
  // Idempotent: Astro.url.pathname already carries the base when hosted under
  // one, so skip re-prefixing paths that start with it.
  if (base && (path === base || path.startsWith(`${base}/`))) return path;
  const [p, hash] = path.split('#');
  // GitHub Pages serves directories: /docs → /docs/index.html needs the slash.
  let out = p.replace(/\/+$/, '');
  if (out !== '' && !/\.[a-z0-9]+$/i.test(out)) out += '/';
  return `${base}${out === '' ? '/' : out}${hash ? `#${hash}` : ''}`;
}
