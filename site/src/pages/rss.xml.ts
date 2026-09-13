// RSS feed for /blog (Astro static endpoint; content collections at build time).
import rss from '@astrojs/rss';
import { getCollection } from 'astro:content';
import { SITE } from '../data/site';
import { baseHref } from '../lib/base';

export async function GET(context) {
  const posts = (await getCollection('blog', ({ data }) => !data.draft)).sort(
    (a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf(),
  );
  return rss({
    title: 'VibeShield blog',
    description: 'Notes on AI-code security: slopsquatting, hallucinated packages, threat models.',
    site: new URL((import.meta.env.BASE_URL ?? '/').replace(/\/+$/, ''), SITE.url).href + '/',
    items: posts.map((p) => ({
      title: p.data.title,
      description: p.data.description,
      pubDate: p.data.pubDate,
      link: baseHref(`/blog/${p.id}/`),
    })),
  });
}
