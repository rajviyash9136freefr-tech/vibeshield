/**
 * Central site and legal configuration for VibeShield.
 * All legal pages, metadata tags, structured data, and SEO readers pull from here.
 * Placeholders marked with [PLACEHOLDER] or [CONFIRM] must be finalized by the project owner.
 */

export const siteConfig = {
  siteName: 'VibeShield',
  siteUrl: 'https://rajviyash9136freefr-tech.github.io',
  basePath: '/vibeshield',
  fullUrl: 'https://rajviyash9136freefr-tech.github.io/vibeshield',
  
  // Legal & Operator identity
  legalName: '[PLACEHOLDER: VibeShield Open Source Contributors]',
  contactEmail: '[PLACEHOLDER: legal@vibeshield.dev]',
  supportEmail: '[PLACEHOLDER: support@vibeshield.dev]',
  securityEmail: 'security@vibeshield.dev',
  address: '[PLACEHOLDER: New Delhi, India]',
  governingLaw: '[CONFIRM: Republic of India, jurisdiction of courts in New Delhi, India]',
  
  // Modification dates
  lastUpdated: 'September 19, 2026',
  legalLastUpdated: 'September 19, 2026',
  securityTxtExpires: '2027-09-19T00:00:00.000Z',

  // SEO Defaults
  defaultTitle: 'VibeShield — AI Code Security Scanner & Bug Hunter',
  defaultDescription: 'Static security scanner for AI vibe coding. Hunt hallucinated packages, leaked API keys, and insecure defaults before commit. 100% local, zero data egress.',
  
  // Social & Repository
  social: {
    github: 'https://github.com/rajviyash9136freefr-tech/vibeshield',
    repo: 'https://github.com/rajviyash9136freefr-tech/vibeshield',
    issues: 'https://github.com/rajviyash9136freefr-tech/vibeshield/issues',
    securityAdvisories: 'https://github.com/rajviyash9136freefr-tech/vibeshield/security/advisories/new',
    license: 'https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/LICENSE',
  },
} as const;

export type SiteConfig = typeof siteConfig;
