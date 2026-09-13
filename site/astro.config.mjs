import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import sitemap from '@astrojs/sitemap';

// GitHub Pages project-page hosting: origin-only `site`, path under `base`.
// If vibeshield.dev is ever connected (CNAME + Pages custom domain), move the
// full URL into `site` and keep `base` only while it's a project page.
const BASE = '/vibeshield/';

/**
 * Astro does not apply `base` to links authored in markdown content, so
 * root-relative hrefs in docs/blog posts are prefixed here (idempotent —
 * anything already prefixed, external, or a bare #hash passes through).
 */
function rehypeBaseLinks() {
  return (tree) => {
    const walk = (node) => {
      if (node?.type === 'element' && node.properties) {
        for (const key of ['href', 'src']) {
          const v = node.properties[key];
          if (typeof v !== 'string' || !v.startsWith('/') || v.startsWith('//')) continue;
          const [p, hash] = v.split('#');
          let out = p.replace(/\/+$/, '');
          // extensionless paths are directories on Pages; keep bare "/" bare
          if (out !== '' && !/\.[a-z0-9]+$/i.test(out) && !out.startsWith(BASE)) out += '/';
          if (!out.startsWith(BASE)) out = BASE + out.replace(/^\/+/, '');
          node.properties[key] = out + (hash ? `#${hash}` : '');
        }
      }
      (node?.children || []).forEach(walk);
    };
    walk(tree);
  };
}

export default defineConfig({
  site: 'https://rajviyash9136freefr-tech.github.io',
  base: BASE,
  trailingSlash: 'ignore',
  markdown: {
    rehypePlugins: [rehypeBaseLinks],
  },
  vite: {
    plugins: [tailwindcss()],
    server: {
      fs: {
        allow: ['..'],
      },
    },
  },
  integrations: [sitemap()],
});
