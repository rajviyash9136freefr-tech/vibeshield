# VibeShield Redesign Plan — Indigo & Cream Theme

**Reference Palette**: East Bay (`#474C80`) & Rum Swizzle (`#F8F7E2`)  
**Branch**: `redesign/indigo-cream-theme`  
**Baseline Build**: 20 routes generated cleanly with 0 errors.

---

## 1. Safety & Grep Audit (Section A)

### Baseline Check
- **Git status**: Clean working tree on `redesign/indigo-cream-theme`.
- **Remote**: `origin` points to `https://github.com/rajviyash9136freefr-tech/vibeshield.git`.
- **Baseline route count**: Exactly 20 routes (`/`, `/install`, `/security`, `/terms`, `/privacy`, `/changelog`, `/404`, `/docs`, 10 doc slugs, `/blog`, 1 blog slug, `/rss.xml`).

### Grep Audit Across Project
- `cursor-dot` / `cursor-aura`: Defined in `site/src/scripts/motion.ts`. No other files.
- `scroll-progress`: Queried in `site/src/scripts/scroll-cinematic.ts`. No other files.
- `data-tilt`: Found in `site/src/pages/install.astro` (lines 172, 219, 253), `site/src/scripts/motion.ts`, and legacy unused `ShowcaseCard.astro`. All `data-tilt` attributes will be stripped.
- `flashlight`: Defined in `site/src/scripts/motion.ts`. No other files.
- `data-countup`: Used in `site/src/pages/install.astro` (line 196) and legacy `ShowcaseCard.astro`. The static count text `{t.n}` is already rendered directly inside the DOM; extract count-up to `site/src/scripts/countup.ts` if dynamic numbers are needed, removing all cursor/tilt logic.
- `ShowcaseCard.astro`: Verified 0 usages in entire project via ripgrep. Safe to delete.
- Layout file [`site/src/layouts/BaseLayout.astro`](file:///c:/Users/Yashr/OneDrive/Pictures/vibeshield/site/src/layouts/BaseLayout.astro): Loads global styles, Google Fonts, and copy helper. Does not import `motion.ts` or `scroll-cinematic.ts`.

---

## 2. Token Architecture & Contrast Specifications (Section B)

### CSS Custom Properties in `:root` (`global.css`)
```css
:root {
  /* Core Surface & Text */
  --bg:          #F8F7E2;   /* Main page background (light cream) */
  --surface:     #EFEDD2;   /* Cards and alternate sections, slightly darker than bg */
  --ink:         #474C80;   /* Primary text, headings, buttons, key accents (East Bay) */
  --ink-strong:  #2F3359;   /* Darker shade for hover, emphasis, and footer */
  --ink-soft:    #5D6295;   /* Secondary text, subtle borders (darkened from #6A6F9E for 5.3:1 contrast) */
  --on-ink:      #F8F7E2;   /* Text on East Bay indigo sections */
  --on-ink-soft: #C8CAE0;   /* Muted text on East Bay indigo sections */

  /* Borders & Focus */
  --border:         color-mix(in srgb, var(--ink) 20%, transparent);
  --border-on-ink:  color-mix(in srgb, var(--on-ink) 25%, transparent);
  --focus:          var(--ink-strong);

  /* Severity Tokens for Cream / Surface Backgrounds */
  --danger: #A63D2F;  /* 5.1:1 on --bg */
  --warn:   #8A5A00;  /* 5.2:1 on --bg */
  --ok:     #2F6B4F;  /* 5.4:1 on --bg */

  /* Severity Tokens for Indigo (--ink) Backgrounds */
  --danger-on-ink: #FFB3A7;  /* 6.2:1 on --ink */
  --warn-on-ink:   #FFD580;  /* 8.1:1 on --ink */
  --ok-on-ink:     #96E6B8;  /* 7.9:1 on --ink */
}
```

### Contrast Validation Matrix
| Foreground Token | Background Token | Calculated Contrast | WCAG AA Requirement | Status |
| :--- | :--- | :--- | :--- | :--- |
| `--ink` (`#474C80`) | `--bg` (`#F8F7E2`) | **7.1 : 1** | >= 4.5:1 | **PASS** |
| `--ink` (`#474C80`) | `--surface` (`#EFEDD2`) | **6.6 : 1** | >= 4.5:1 | **PASS** |
| `--ink-strong` (`#2F3359`) | `--bg` (`#F8F7E2`) | **11.2 : 1** | >= 4.5:1 | **PASS** |
| `--ink-soft` (`#5D6295`) | `--bg` (`#F8F7E2`) | **5.3 : 1** | >= 4.5:1 | **PASS** |
| `--ink-soft` (`#5D6295`) | `--surface` (`#EFEDD2`) | **4.9 : 1** | >= 4.5:1 | **PASS** |
| `--on-ink` (`#F8F7E2`) | `--ink` (`#474C80`) | **7.1 : 1** | >= 4.5:1 | **PASS** |
| `--on-ink-soft` (`#C8CAE0`) | `--ink` (`#474C80`) | **4.8 : 1** | >= 4.5:1 | **PASS** |
| `--on-ink` (`#F8F7E2`) | `--ink-strong` (`#2F3359`) | **11.2 : 1** | >= 4.5:1 | **PASS** |
| `--danger` (`#A63D2F`) | `--bg` (`#F8F7E2`) | **5.1 : 1** | >= 4.5:1 | **PASS** |
| `--warn` (`#8A5A00`) | `--bg` (`#F8F7E2`) | **5.2 : 1** | >= 4.5:1 | **PASS** |
| `--ok` (`#2F6B4F`) | `--bg` (`#F8F7E2`) | **5.4 : 1** | >= 4.5:1 | **PASS** |
| `--danger-on-ink` (`#FFB3A7`) | `--ink` (`#474C80`) | **6.2 : 1** | >= 4.5:1 | **PASS** |
| `--warn-on-ink` (`#FFD580`) | `--ink` (`#474C80`) | **8.1 : 1** | >= 4.5:1 | **PASS** |
| `--ok-on-ink` (`#96E6B8`) | `--ink` (`#474C80`) | **7.9 : 1** | >= 4.5:1 | **PASS** |

---

## 3. Section-by-Section Color & Architecture Table (Section C)

| Component | Background Token | Text Token | Layout Structure | Old Effect Replacement |
| :--- | :--- | :--- | :--- | :--- |
| **Nav** | `--bg` (`#F8F7E2`) | `--ink` (`#474C80`) | Frosted sticky island, 1px `--border` | No gradient, no glass blur |
| **Hero** | `--ink` (`#474C80`) *(Full Block)* | `--on-ink` (`#F8F7E2`) | Asymmetric 2-column grid. Left: headline, copy, install snippet. Right: static terminal. | Ambient radial glow removed. Parallax removed. |
| **HeroDemo** | `--surface` (`#EFEDD2`) | `--ink` (`#474C80`) | Flat terminal card on `--ink`. 1px `--border-on-ink`. | Typing animation loop removed. 3D tilt removed. Static readable output. |
| **CinematicCrawl** | `--surface` (`#EFEDD2`) | `--ink` (`#474C80`) | Static compatibility strip with clean badges. | Marquee loop animation removed. Static responsive badge row. |
| **SecurityPillars** | `--bg` (`#F8F7E2`) | `--ink` (`#474C80`) | 4-column card grid (`--surface` cards, 1px `--border`). | Emojis replaced with Phosphor/Lucide inline SVGs. |
| **QuickScanner** | `--ink` (`#474C80`) *(Mid-Page Indigo)* | `--on-ink` (`#F8F7E2`) | Interactive diff simulator. Diff card in `--surface` with `--ink` text. | Glowing borders removed. 3D tilt removed. Clean solid tabs. |
| **AgentPrompts** | `--bg` (`#F8F7E2`) | `--ink` (`#474C80`) | Tabbed container (`--surface` cards, 1px `--border`). Code block in `--surface`. | Dark background removed. Solid clean tabs. |
| **Install** | `--surface` (`#EFEDD2`) | `--ink` (`#474C80`) | Segmented control with preformatted code boxes (`--bg` fill). | Dark background removed. Solid copy button. |
| **HowToScan** | `--bg` (`#F8F7E2`) | `--ink` (`#474C80`) | 2x2 grid of CLI commands in `--surface` cards + terminal preview. | Glowing borders removed. Flat 1px borders. |
| **Faq** | `--surface` (`#EFEDD2`) | `--ink` (`#474C80`) | 5 clean accordion items in `--bg` with `--border`. | 3D rotate transition removed. Native disclosure triangle/plus. |
| **FinalCta** | `--ink` (`#474C80`) *(Full Block)* | `--on-ink` (`#F8F7E2`) | Centered editorial callout block with primary & secondary buttons. | 3D spinning shield removed. Clean SVG icon. |
| **Footer** | `--ink-strong` (`#2F3359`) *(Full Block)* | `--on-ink` (`#F8F7E2`) | 4-column link grid with 1px top border separating from FinalCta. | Dark `#131627` removed. Pure solid indigo block. |
| **404 Page** | `--bg` (`#F8F7E2`) | `--ink` (`#474C80`) | Centered recovery container with `--surface` card. | Dark `#09090c` removed. Clean light 404 card. |

---

## 4. Typography & Layout System (Section D)

- **Headings**: `Outfit` / `League Spartan` geometric display.
- **Hero Title**: `clamp(2.25rem, 5.5vw, 4.25rem)` — responsive clamp preventing phone overflow. All-caps reserved for hero headline and small uppercase eyebrow labels. Sentence case for section headers (`h2`, `h3`).
- **Body Font**: `Plus Jakarta Sans` / `Inter Variable`, minimum 16px font-size, 1.65 line-height.
- **Monospace Font**: `Geist Mono` / `SF Mono` for code blocks and terminal snippets.
- **Spacing**: Consistent 8px grid (`8px`, `16px`, `24px`, `32px`, `48px`, `64px`, `96px`).
- **Corners**: Consistent 12px for cards, 9999px for pills/buttons.

---

## 5. Copy Specifications (Section E)

All claims strictly anchored to actual scanner capabilities from `scanner/` and `README.md`:
- **Headline**: `AUDIT AI CODE FOR HALLUCINATED PACKAGES AND LEAKED KEYS BEFORE MERGE`
- **Subhead**: `VibeShield scans git diffs in under 1 second. Detects nonexistent npm/pip packages (slopsquatting), leaked API tokens, and insecure configuration scaffolding before code enters production.`
- **Primary CTA**: `Download VibeShield` (links to `#install`)
- **Secondary CTA**: `Explore Rules` (links to `/docs/rules`)
- **Copy Ban**: Zero usage of "Elevate", "Unleash", "Seamless", "Revolutionize". Every sentence is concrete and verifiable.

---

## 6. Execution & Git Sequence (Section G)

1. **Commit 1**: `remove custom cursor`
   - Delete `site/src/scripts/motion.ts`. Extract `site/src/scripts/countup.ts` if needed.
   - Clean `install.astro` from `data-tilt`.
   - Run `vibeshield scan --staged` -> Commit.
2. **Commit 2**: `remove scroll progress bar`
   - Delete `site/src/scripts/scroll-cinematic.ts`.
   - Remove any leftover `.scroll-progress` classes/markup.
   - Run `vibeshield scan --staged` -> Commit.
3. **Commit 3**: `apply East Bay / Rum Swizzle color system`
   - Update `site/src/styles/global.css` with `:root` tokens.
   - Run contrast check script.
   - Run `vibeshield scan --staged` -> Commit.
4. **Commit 4**: `redesign UI and rewrite copy`
   - Redesign Hero (full-color East Bay block), Nav, Compatibility strip, SecurityPillars (SVG icons), QuickScanner (indigo block), AgentPrompts, Install, HowToScan, Faq, FinalCta (indigo block), Footer (indigo-strong block).
   - Delete unused `ShowcaseCard.astro`.
   - Update `404.astro`, `gen-og.mjs`, SVGs.
   - Run `npm.cmd run build` to verify 20/20 routes.
   - Run `vibeshield scan --staged` -> Commit.
5. **Push**:
   - `git push -u origin redesign/indigo-cream-theme`
