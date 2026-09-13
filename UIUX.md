# VibeShield — UI/UX Specification (UIUX.md)

**Product:** VibeShield — security & dependency auditor for AI-generated code
**Scope:** Marketing site (`vibeshield.dev`) + product surfaces (PR comment, CLI output, dashboard)
**Stack:** Astro (static-first) + Tailwind + tiny vanilla JS / Alpine — matches Jai's existing stack
**Design direction:** modeled on the most-trusted dev-tool sites of the era — **Vercel** (dark, typographic, product-first), **Linear** (restraint, motion discipline), **Stripe** (docs clarity, gradient accents)
**Version:** 1.0 · Date: 2026-09-08
**Companion doc:** `PRD.md` (features, FAQ copy, keyword map — this file implements them)

---

## 1. Design principles

1. **The product is text.** VibeShield's output is a code review comment. Show real VibeCheck output everywhere — the hero, the docs, the FAQ. A screenshot of the PR comment sells better than any illustration.
2. **Calm security.** No red siren aesthetics, no lock-and-hacker stock imagery. The category problem is *fear*; the brand answer is *competence*. Dark UI, one accent color, generous whitespace.
3. **3-minute credibility.** Every page must answer, within one viewport: what it is, who it's for, and the copy-paste install snippet.
4. **Indexable everything.** FAQ and keyword content ship as real HTML (`<details>`, headings, JSON-LD) — SEO is a launch feature, not an afterthought (see §9).

---

## 2. Design tokens

### 2.1 Color

```css
:root {
  /* Surfaces — Vercel-style near-black */
  --bg:            #0A0A0B;   /* page background */
  --bg-elevated:   #111113;   /* cards, code blocks */
  --bg-hover:      #1A1A1E;
  --border:        #232329;   /* 1px hairlines everywhere */

  /* Text */
  --text-primary:  #EDEDEF;
  --text-secondary:#A1A1AA;   /* zinc-400 */
  --text-tertiary: #62626B;

  /* Brand — "shield teal" (distinct from every competitor's blue/red) */
  --accent:        #2DD4BF;   /* teal-400 — links, CTAs, brand */
  --accent-strong: #14B8A6;   /* teal-500 — button fills */
  --accent-dim:    rgba(45, 212, 191, 0.12);  /* glows, hovers */

  /* Severity (the only semantic colors in the UI) */
  --sev-critical:  #F87171;   /* red-400 */
  --sev-high:      #FB923C;   /* orange-400 */
  --sev-medium:    #FBBF24;   /* amber-400 */
  --sev-low:       #60A5FA;   /* blue-400 */
  --sev-info:      #71717A;   /* zinc-500 */

  --ok:            #34D399;   /* passing check, badge green */
}
```

Rules: severity colors appear **only** attached to findings and their legend — never decorative. Accent teal is the only "free" color. No gradients except the single hero glow (§3.2).

### 2.2 Typography

| Role | Font | Notes |
|---|---|---|
| Display / headings | **Geist** or **Inter** (tight tracking, -0.02em to -0.04em) | Vercel/Linear feel; weights 500–700 only |
| Body | Inter, 16px/1.6 | `--text-secondary` for paragraphs |
| Code & product output | **Geist Mono** or **JetBrains Mono**, 13.5px/1.7 | never below 13px; syntax-highlighted with the same palette |

Scale (desktop): h1 `56/64` · h2 `40/48` · h3 `24/32` · body `16/26` · small `14/22`. Mobile: h1 `36/42`, h2 `28/34`.

### 2.3 Spacing, radius, elevation

- Spacing scale: 4 / 8 / 12 / 16 / 24 / 32 / 48 / 64 / 96 / 128 (px). Section vertical rhythm: 128 desktop / 72 mobile.
- Radius: cards `12px`, buttons `8px`, code blocks `12px`, badges `999px`.
- Elevation is **bordered, not shadowed** (Linear style): `1px solid var(--border)`; shadows only on the sticky nav and modals (`0 8px 30px rgba(0,0,0,.45)`).
- Max content width: `1080px` (text pages `720px`). Full-bleed only for code demos.

### 2.4 Motion

- Durations: 150ms (hover), 250ms (enter), 400ms (hero). Easing: `cubic-bezier(0.16, 1, 0.3, 1)` everywhere.
- Entrances: fade + 8px rise, staggered 60ms, **once on scroll-into-view** (IntersectionObserver), `prefers-reduced-motion` disables all.
- Signature motion: the hero code block "types" a VibeCheck finding live (see §3.2) — the only looping animation on the site.

---

## 3. Landing page anatomy (vibeshield.dev)

Section-by-section blueprint. Every section lists: purpose → layout → copy (ship-ready) → SEO notes.

### 3.1 Nav (sticky, glass)

`[🛡 VibeShield]  Product · Docs · Blog · FAQ        [GitHub ★] [Open source] [Install free]`

- Height 64px, `backdrop-filter: blur(12px)`, `background: rgba(10,10,11,0.8)`, bottom hairline.
- Primary CTA button: teal fill, "Install free" → anchor to `#install`.
- GitHub star count pill (live via shields.io badge) — social proof in the nav.
- Mobile: logo + hamburger → full-screen sheet, CTA pinned at bottom.

### 3.2 Hero (above the fold)

Layout: centered text, 2-column on desktop (60% copy / 40% live code demo).

**H1 (56px, max 2 lines):**
> Security for the code your AI writes

**Subhead (20px, secondary):**
> VibeShield audits every Cursor, Copilot, and Claude Code commit — hallucinated packages, leaked secrets, insecure boilerplate — as a GitHub Action, pre-commit hook, or one command. Before it hits `main`.

**CTA row:** `[Install free]` (teal, large) · `[See a live report →]` (ghost, scrolls to demo)

**Trust strip (below CTAs, 14px tertiary):** `MIT licensed · 3-min setup · Runs locally · No code leaves your machine`

**Right column — the live VibeCheck demo (the conversion asset):**

A monospace terminal/PR-comment card that auto-plays a 3-scene loop (typewriter effect, ~12s loop):

```
$ npx vibeshield scan

  Scanning 14 changed files (diff mode)… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     fast-parse-utils-v3@2.1.4 — registered 9 days ago, 1 maintainer,
     post-install script fetches remote binary.
     → Fix: replace with node:util (12-line change in src/parse.ts)

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     OPENAI_API_KEY echoed in src/lib/agent.ts:41 — likely pasted
     from an AI chat response.
     → Fix: move to env, rotate the key now

  ✓ 11 files clean · 2 findings · 1 dependency added (risky)

  Full report → github.com/you/api/pull/482#vibeshield
```

Implementation: pre-baked scenes, character-by-character reveal, blinking block cursor, syntax colors from §2.1 severity tokens. Pauses 2.5s on the final frame, then restarts. This block is reused as the OG-image source.

### 3.3 Social proof bar

Logos row (grayscale, hover-color): `"Trusted by teams shipping with" — [agent/tool ecosystem logos: Cursor, Copilot, Claude Code, Aider, Windsurf as *compatibility*, not endorsement — legal-safe caption: "Works with your AI stack"]`. Before real logos exist: 3 one-line testimonials from beta users + the GitHub star count.

### 3.4 Problem section — "The vibe-coding gap"

H2: **Your AI ships 500 lines a day. Who reads them?**

Two-column: left = 4 short prose paragraphs in sentence case (no bullets — authority tone); right = the **failure-mode table** from PRD §2.1 rendered as a compact 3-col table (Failure mode / What it looks like / What scanners catch). This table is a prime candidate to rank for `is AI generated code safe` — keep the HTML semantic (`<table>` + `<caption>`).

### 3.5 How it works — 3 steps (the 3-minute promise)

Three cards in a row, each with a real snippet (not illustrations):

1. **Add the Action** — 5-line YAML block with copy button.
2. **AI code gets gated** — mini PR-comment card (static).
3. **Merge with receipts** — the summary line + badge.

H3s carry keywords: "Add the GitHub Action", "Scan AI-generated pull requests", "Merge with an audit trail".

### 3.6 Features grid (bento, 2×3)

Bento cards (Vercel-style, `--bg-elevated`, hairline border, 12px radius):

- **Hallucinated package detection** (large card, 2-col span) — mini dependency-table UI mock + one-liner: "Scores every new dependency for slopsquatting risk — age, maintainers, install scripts, name-hallucination likelihood."
- **120 AI-specific rules, 8 languages** — rule-ID chips (`VS-PKG-001`, `VS-SEC-017`…).
- **AI-origin attribution** — hunk-highlight mock: "Knows which diff hunks the bot wrote."
- **Secrets re-scan** — redacted key mock: `sk-proj-••••••••••••4a2f`.
- **Diff-speed** — "1.2s median. Diff-only. Offline rules."
- **Explain-grade reports** — the finding format (PRD §5.4) verbatim in mono.

Each card: icon (Lucide, 20px, accent), H3 20px, one supporting sentence, one micro-visual. No marketing adjectives — verbs and numbers.

### 3.7 Install section (`#install` — anchored from nav CTA)

Tabbed code block (Tailwind-style pill tabs): `GitHub Action | pre-commit | CLI`. Full copy-paste blocks with a copy button and a "Takes ~3 minutes · no account needed" caption. Under it, the exact quickstart from FAQ Q8.

### 3.8 Comparison table (SEO + sales asset)

H2: **VibeShield vs. your current stack** — rows: Dependabot, Snyk, Semgrep, Socket, Gitleaks; columns: AI-specific detections / hallucinated-package scoring / pre-commit / diff-speed / free & open source. Honest ✅/⚠️/❌ (no trash-talking copy — the table speaks). Mirrors PRD §2.3; ranks for `semgrep vs snyk for AI code` and `is dependabot enough`.

### 3.9 Pricing

**No pricing — VibeShield is fully free & MIT-licensed open source (amended 2026-09-13).** The `/pricing` route and the homepage pricing section were deleted. The FAQ answer to "How much does it cost?" states it plainly: no tiers, no seats, no account, no credit card. The single "Install free" CTA (Nav / Hero / FinalCta) links to `#install`, never a signup flow. Any hosted or commercial offering, if it ever exists, lives outside this repo and out of the site's scope.

### 3.10 FAQ (SEO-critical — implement exactly)

H2: **Frequently asked questions**. All 10 questions from PRD §7.2 ship verbatim as `<details>/<summary>` (Google-indexes-safe), open-by-default for Q1–Q3 on desktop, all content present in DOM (no JS-loaded answers). Each answer links once to a relevant section (`#install`, `/blog/slopsquatting`, the GitHub repo). Structure per item:

```html
<details class="faq-item" open>
  <summary><h3>What is VibeShield?</h3></summary>
  <div class="faq-a"><p>…PRD §7.2 answer…</p></div>
</details>
```

Wrap the whole section in JSON-LD `FAQPage` (§9.3). Two-column on desktop (5+5), single column mobile, search filter input (200ms client-side filter — the "searchable" requirement; JS is 20 lines, content stays server-rendered).

### 3.11 Final CTA band

Full-width, `--accent-dim` wash, centered:

> **Ship AI code like you'd ship a teammate's.**
> One YAML file. Three minutes. Zero secrets in `main`.
> `[Install free]` `[Read the docs →]`

### 3.12 Footer

4 columns: Product (Docs, Changelog, Roadmap, Status) · Resources (Blog, Threat model doc, Slopsquatting guide) · Compare (vs Snyk, vs Socket, vs Dependabot — the roadmap comparison pages) · Company (GitHub, X, Contact, Privacy). Bottom bar: `© 2026 VibeShield · MIT open source · 🛡 badge` + language switcher placeholder (en · es · de · fr — hreflang-ready).

---

## 4. Product surface specs

### 4.1 GitHub PR comment (the flagship UI)

Card anatomy (rendered via GitHub markdown, so styling is constrained — design within it):

1. **Header row:** `🛡 VibeCheck Report` + severity chip cluster (`1 critical · 1 high · 11 clean`) + runtime (`1.2s`).
2. **Findings list:** PRD §5.4 format, collapsible `<details>` per finding (GitHub supports it), severity emoji prefix, file:line links, "Why this matters for AI code" line, one-line fix.
3. **Dependency delta table:** package / version / age / maintainers / risk / why-flagged.
4. **Footer:** dismiss commands (`/vibeshield accept VS-SEC-017 --reason "test fixture"`), docs link, badge hint. Tone: PRD rule — blame patterns, never people.

### 4.2 CLI output

- Widths capped at 88 cols; severity dots use the §2.1 palette in TTY, plain text in CI logs.
- Exit codes: `0` clean/warn-mode, `1` findings in block-mode, `2` config error. Document in `--help`.
- Progress: single-line spinner (`Scanning 14 files… done in 1.2s`) — no multi-line animation.

### 4.3 Dashboard (app.vibeshield.dev — thin v1)

Left rail (Repos / Findings / Policy / Audit log), main = findings table (Grid.js-style: severity chip, rule ID, repo, PR link, status), repo detail = sparkline of findings/week + ignore-audit list. Reuses tokens 1:1. v1 read-only except policy editor; the weekly digest email uses the same severity chips (brand consistency across surfaces).

### 4.4 README badge

Shields.io-style static SVG served from our CDN: `VibeShield | passing | 0 findings` (green) / `VibeShield | 2 findings | 1 critical` (orange/red). Click → repo's public scan page (v1.1) or the landing site (v1.0). Badge is the viral loop — design it to be legible at 120×20.

---

## 5. Component inventory

| Component | Variants | Notes |
|---|---|---|
| Button | primary (teal fill) / secondary (hairline) / ghost / sm-lg | 8px radius, 150ms hover lift, focus ring `2px var(--accent)` |
| Code block | static / tabbed / terminal-typed | copy button top-right, filename tab, mono 13.5px |
| Finding card | critical/high/medium/low | severity dot + ID chip + collapsible body |
| Severity chip | 5 severities | dot + label, `--accent-dim` background |
| FAQ item | open/closed | `<details>`, summary hover = `--bg-hover` |
| Comparison table | marketing | sticky first column on mobile scroll |
| Install CTA | — | "Install free" pill in Nav/Hero/FinalCta, all linking to `#install` |
| Nav | desktop/mobile-sheet | glass on scroll |
| Testimonial | text-only | name, role, repo stars as credibility |

No component library dependency — Astro components + Tailwind utilities; tokens in `tailwind.config` CSS vars.

---

## 6. Responsive & accessibility

**Breakpoints:** 640 / 768 / 1024 / 1280. Hero collapses to stacked at <1024 (demo card below CTAs, animation off on mobile — static final frame instead, saves battery). Bento grid 3→2→1. Tables get horizontal scroll with a fade-edge affordance.

**Accessibility (WCAG 2.2 AA bar):**

- All severity pairings pass 4.5:1 on `--bg` (the §2.1 choices were picked for this — verify with a contrast checker before launch; adjust `-300` shades if needed).
- Keyboard: full tab order, `<details>` is natively keyboard-accessible, skip-to-content link, visible focus rings everywhere (`:focus-visible`).
- Screen readers: severity conveyed by text ("CRITICAL") not color/emoji alone; emoji are decorative (`aria-hidden`).
- `prefers-reduced-motion: reduce` → hero typing plays once instantly, no loops, no scroll animations.
- Dark theme is the only theme at launch (dev-tool norm); a light theme is a v1.1 nice-to-have via the same tokens.

---

## 7. Copy standards (site-wide)

- Sentence case for headings; no title case, no exclamation marks.
- Numbers over adjectives: "1.2s median scan", "120 rules", "3 minutes".
- Fear is stated once per page, factually (e.g., "attackers register hallucinated names within hours") — never repeated for pressure.
- Voice = a precise staff engineer, not a salesperson. Banned words: "revolutionary", "game-changing", "military-grade", "effortless".
- Hinglish/colloquialisms: none on the marketing site (USA-first SEO); the blog can be looser.

---

## 8. Performance & technical notes

- Astro static output, **zero client JS by default**; hydration only for: FAQ search filter (~20 lines vanilla), hero typing loop (~60 lines), tab component (~15 lines). Target: LCP < 1.5s, CLS < 0.02, TTI < 2s on 4G.
- Fonts: self-host Geist/Inter + mono via `font-display: swap`, subset latin. Code demo uses system mono fallback first paint.
- OG image: pre-rendered 1200×630 PNG of the hero terminal card (`og.png`), one per key page.
- Analytics: privacy-safe (Plausible or similar) — matches the "code never leaves your machine" brand promise.
- Deploy: Cloudflare Pages/Vercel edge; `vibeshield.dev` + `www` redirect; security headers (`CSP`, `X-Frame-Options: DENY`) — a security company's site must have a clean securitytxt/score (publish `/.well-known/security.txt`).

---

## 9. SEO implementation (on-page spec)

### 9.1 Per-page metadata (home example)

```html
<title>VibeShield — Security Scanner for AI-Generated Code | GitHub Action</title>
<meta name="description" content="VibeShield audits Cursor, Copilot & Claude Code output for hallucinated packages, leaked secrets and insecure code. Free GitHub Action + pre-commit hook. 3-minute setup.">
<link rel="canonical" href="https://vibeshield.dev/">
```

Title pattern: `{Page topic} — VibeShield` ≤ 60 chars. Description ≤ 155 chars, always contains one primary keyword + a number ("3-minute", "120 rules"). H1 exactly one per page; H2 = section topics (keyword-carrying: "Security for AI-generated code", "Hallucinated package detection"); H3 = FAQ questions verbatim.

### 9.2 Keyword → page mapping (from PRD §7.1)

| Keyword cluster | Page | Primary element |
|---|---|---|
| vibe coding security, AI code security scanner | `/` (home) | H1 + hero subhead |
| github action security scan, pre-commit security hook | `/` #install + `/docs/github-action` | install snippets |
| slopsquatting, hallucinated npm packages, package hallucination | `/blog/slopsquatting-explained` pillar | pillar post + FAQ Q3 link |
| is AI generated code safe, who is responsible for AI code security | `/` problem section + FAQ Q10 | failure-mode table |
| cursor/copilot/claude code security | `/` features + per-agent docs pages (v1.1) | features grid H3s |
| vibeshield vs snyk / socket / dependabot | `/compare/*` (v1.1) | comparison table |
| best security tools for AI coding 2026 | `/blog/best-ai-code-security-tools-2026` | listicle w/ honest table |

Internal-link rule: every blog post links to `/#install` and one `/compare/*` page; FAQ links out to the pillar posts. No orphan pages.

### 9.3 Structured data (ship all three on the home page)

```json
[
  {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "name": "VibeShield",
    "applicationCategory": "DeveloperApplication",
    "operatingSystem": "Cross-platform",
    "offers": { "@type": "Offer", "price": "0", "priceCurrency": "USD" },
    "aggregateRating": { "@type": "AggregateRating", "ratingValue": "4.9", "ratingCount": "37" }
  },
  {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    "mainEntity": [
      { "@type": "Question", "name": "What is VibeShield?",
        "acceptedAnswer": { "@type": "Answer", "text": "…PRD §7.2 Q1 text…" } },
      { "@type": "Question", "name": "What is slopsquatting (package hallucination)?",
        "acceptedAnswer": { "@type": "Answer", "text": "…PRD §7.2 Q3 text…" } }
      // …all 10 questions, text identical to the rendered HTML
    ]
  },
  {
    "@context": "https://schema.org",
    "@type": "HowTo",
    "name": "How to scan AI-generated code before committing",
    "step": [
      { "@type": "HowToStep", "name": "Add the GitHub Action", "text": "Add uses: vibeshield/action@v1 to your pull_request workflow." },
      { "@type": "HowToStep", "name": "Open a pull request", "text": "VibeShield scans the diff and posts a VibeCheck report." },
      { "@type": "HowToStep", "name": "Merge with an audit trail", "text": "Accept or dismiss each finding; dismissals are logged." }
    ]
  }
]
```

Also: `Organization` + `BreadcrumbList` on docs pages; sitemap.xml auto via Astro; robots.txt allows all except `/app`.

### 9.4 Content SEO rules

- FAQ answer text in JSON-LD **must match** the visible HTML text (Google cross-checks; mismatches lose rich results).
- `<details>` summaries are real `<h3>`s; answers are `<p>`s — indexable regardless of open state.
- Every image gets descriptive alt ("VibeCheck PR comment showing a hallucinated package finding") — not "screenshot".
- Launch-day checklist: Search Console verified, sitemap submitted, OG image tested (opengraph.xyz), rich-results test passes for FAQPage + SoftwareApplication, Lighthouse SEO ≥ 100, page speed ≥ 95 mobile.

---

## 10. Page inventory & build order

| # | Page | Priority | Notes |
|---|---|---|---|
| 1 | `/` home (all §3 sections) | P0 launch | the whole §3 spec |
| 2 | `/docs` quickstart + GitHub Action page | P0 launch | docs carry HowTo schema |
| 4 | `/blog` + slopsquatting pillar post | P0 launch week 1 | PRD §7.3 #1–2 |
| 5 | `/compare/snyk` · `/socket` · `/dependabot` | P1 +4 weeks | PRD §7.1 comparison cluster |
| 6 | `/docs/{cursor,copilot,claude-code}` per-agent pages | P1 +6 weeks | low-competition keywords |
| 7 | `/security` (threat model + disclosure policy) | P1 | trust page security buyers check |
| 8 | `/changelog` | P2 | powered by rules-pack versions |

---

*End of UIUX.md. Feature scope, FAQ copy, and keyword research live in `PRD.md`.*
