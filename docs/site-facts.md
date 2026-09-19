# VibeShield — Verified Site & Product Facts

*Generated: 2026-09-19*
*Source: Full codebase audit of scanner (Go), distribution packages, Astro static site, and GitHub workflows.*

---

## 1. Data Collection & Forms
- **Forms / Inputs:** None. The website contains zero forms, input fields, search query parameters, or form submission endpoints.
- **Email Signup / Newsletter / Waitlist:** None.
- **Contact Forms:** None. Inquiries are directed via public GitHub Issues / GitHub Discussions or email.
- **Authentication / Accounts:** None. The software requires no accounts, passwords, API keys, logins, or OAuth permissions.
- **Payments / Billing:** None. The tool is 100% free open-source software (FOSS) released under the MIT License. There are no paid tiers, subscriptions, paywalls, or token meters.

## 2. Analytics, Cookies & Client-Side Storage
- **Analytics:** None. No Google Analytics, Plausible, PostHog, Mixpanel, Segment, or tracking pixels.
- **Cookies:** None. The application code sets zero first-party cookies and loads zero third-party cookie providers.
- **Local Storage / Session Storage:** None. Neither `localStorage` nor `sessionStorage` is accessed or written to by client scripts.
- **Third-Party Scripts & Widgets:** None. Zero third-party JavaScript libraries, chatbots, or tracking tag managers.
- **External CDN / Font Resources:**
  - The site loads Google Fonts (`fonts.googleapis.com` / `fonts.gstatic.com`) for Outfit and Plus Jakarta Sans.
  - Standard HTTP connection metadata (IP address, user agent) is transmitted to Google when fetching font files from Google CDN.
  - Local fallback fonts (`@fontsource-variable/geist`, `@fontsource-variable/inter`, `@fontsource-variable/geist-mono`) are bundled locally.
- **Hosting Infrastructure:**
  - The website is hosted as a static site on GitHub Pages (`rajviyash9136freefr-tech.github.io/vibeshield/`).
  - GitHub Inc. processes standard server logs (IP address, request headers, timestamp) in accordance with GitHub's Privacy Statement.

## 3. Product Code Processing & Storage (CLI, Pre-Commit, GitHub Action)
- **Local Execution:**
  - The scanner (`vibeshield` Go binary) operates 100% locally and offline.
  - Analysis is entirely static: AST parsing and RE2 regular-expression matching against the rule packs embedded in the compiled binary.
  - The binary contains no HTTP client or networking code for outbound code transmission.
  - The `--online` flag is a stub/reserved parameter in v3.0.0 (prints a notice and continues offline).
- **Zero Server Storage:**
  - VibeShield operates zero backend servers, cloud databases, or centralized dashboards.
  - User source code, repository metadata, and scan results are never transmitted to, processed by, or stored on any VibeShield server.
- **GitHub Action Mode:**
  - Runs in the user's own GitHub Actions runner environment.
  - Findings are posted directly to the user's pull request as a comment using the repository's native `GITHUB_TOKEN`.
  - Secrets identified by the scanner are automatically redacted (retaining only 4 characters on each side) before printing or posting.
  - No data is exfiltrated to any third-party service.

## 4. Financial & Commercial Facts
- **Paid Features:** None.
- **Payment Providers:** None (no Stripe, Lemon Squeezy, Paddle, PayPal).
- **Licensing:** Free, open source under MIT License.

---

## 5. Required Placeholders & Legal Confirmation
The following items are placeholders that must be confirmed or customized by the site operator before final publication:

- `[PLACEHOLDER: legalName]` — Legal name of the entity or individual operating VibeShield (e.g. "VibeShield Contributors" or personal/company name).
- `[PLACEHOLDER: contactEmail]` — Primary general/legal contact email address (e.g. `legal@vibeshield.dev` or `yashrajvi9136@gmail.com`).
- `[PLACEHOLDER: supportEmail]` — Support/community email address or GitHub repository URL.
- `[PLACEHOLDER: address]` — Physical or registered mailing address (optional for non-commercial open-source projects, required if commercial entity).
- `[CONFIRM: governingLaw]` — Jurisdiction and governing law. Operator is based in India (suggested: *Laws of the Republic of India, with jurisdiction in the courts of New Delhi, India*).
- `[CONFIRM: securityEmail]` — Security vulnerability reporting address (currently set to GitHub Security Advisories or `security@vibeshield.dev`).

> Note: The text generated for `/privacy`, `/terms`, and `/security` is based strictly on the verified technical facts above. This documentation does not constitute formal legal advice. A qualified attorney should review the final text to ensure full compliance with the Digital Personal Data Protection Act 2023 (India DPDP), EU General Data Protection Regulation (GDPR), California Consumer Privacy Act (CCPA), and applicable local regulations.
