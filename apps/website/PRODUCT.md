# Product

<!-- impeccable:product-schema 1 -->

Scope: the `apps/website` marketing and onboarding site only. `tenant-admin`, `support-admin`, `docs` and `blog` have their own surfaces and are not covered here.

## Platform

web

## Users

Primary: the technology or digital-transformation lead (IT manager, CDO) at an Iranian company who decides whether to bring AI into the daily work of every team. They need it to be safe, governable and affordable, and they read the site to judge fit, risk and cost before booking a demo or signing up with the gift credit.

Target organizations: private SMBs (roughly 10 to 200 people) and large private enterprises. Banks, government and semi-government bodies are not the primary target of the site, though On-Premise serves their needs when they arrive.

Secondary: developers and technical teams who want one OpenAI-compatible endpoint for many models, with keys, usage and spend under control. They have their own product (the gateway) and their own path through the site.

The site is Persian and right-to-left. Visitors read Persian first and expect Persian-native copy, not translated copy.

## Product Purpose

GodarAI (گدارAI) is two products of equal weight:

1. **AI Workplace:** an organization gives every team AI assistants that work with the team's own knowledge and tools for everyday work, while IT and finance keep control of access, data and spend.
2. **LLM Gateway:** one OpenAI-compatible API in front of 1200+ models from 21 providers, with keys, budgets, routing and logs.

The website exists to get an organization to sign up with the initial gift credit or book a private demo.

Success for a customer, three months in:

- every team uses the assistants daily, not a handful of enthusiasts;
- zero data leakage: no sensitive data leaves through unmanaged AI use;
- predictable cost: AI spend stays under set caps and is reportable to finance.

## Positioning

GodarAI is in the same category as Dust, Langdock and Claude Enterprise, built for Iranian organizations. Four differentiators carry the positioning:

1. **Enterprise security through semantic guardrails:** confidential data is detected by meaning, not keywords; each team has its own topic policy, checked on input and output; uploaded files and attachments pass through the same guardrails.
2. **AI governance by design:** team roles and access to models, assistants and tools; assistant creation can be restricted to a team or role; every request, response and settings change is logged (who, what, when); user and team sync from the organization's directory; SSO with the organization's identity provider in On-Premise deployments only.
3. **Observability and controllability:** usage and cost dashboards per team; request logs and per-request traces.
4. **Cost control and reduction:** budgets and caps per team and per key; routing simple work to cheaper models; response caching; payment in tomans with prepaid credit.

Local advantages a foreign product cannot truthfully claim:

- payment in tomans, stable access to foreign models, no international card or sanctions friction;
- deployment on the organization's own servers (On-Premise) in addition to the GodarAI cloud;
- sending and reading messages in messengers such as Telegram, WhatsApp and Bale, alongside common tools.
- compatibility with Iranian software: accounting and sales, routing/navigation apps, messengers and more («سازگار با نرم‌افزارهای ایرانی»).

Persian language support is not marketed as a feature or differentiator.

The real competitor named on the site is unmanaged, shadow use of ChatGPT by employees. Foreign workplace products (Dust, Langdock, Claude Enterprise) are design and product references only and are not named or compared on the site. Existing gateway comparisons (`/vs/litellm`, `/vs/portkey`) stay.

## Operating Context

- Chat workspace per organization (GodarChat); ready-made assistants and no-code assistant building shared with the team; knowledge through file upload and knowledge bases (RAG); live connections to tools through the MCP gateway (Jira, Notion, Git, CRM, the organization's own MCP servers).
- Messengers: assistants send and read messages in messengers such as Telegram, WhatsApp and Bale as tools; public copy uses exactly that phrasing («پیام‌رسان‌ها مانند تلگرام، واتساپ، بله و …») and does not name Eitaa separately. The assistant does not live inside the messenger as a bot.
- CRMs are reached through MCP and are not named individually («CRM از راه MCP»).
- Admin side: semantic guardrails (working with Persian text), prompt registry, model discovery and catalog, wallet and credits, budgets and rate limits (blocking on reaching a budget is configurable per rule), request logs and traces, settings-change audit, team roles, user sync, SSO (On-Premise).
- Data policy: GodarAI does not train models on customer data (no claim is made about upstream providers); log retention is configurable.
- Deployment and pricing: GodarAI cloud on a monthly subscription; On-Premise on an annual contract, requiring Kubernetes infrastructure, through questionnaire, technical session and installation. Official invoices are issued; a mid-period plan change is prorated.
- OpenAI models are stably available inside GodarAI; chat history from ChatGPT can be imported.
- Pricing (`/pricing`): no plans; every organization gets every feature. Seats per member: Free (500 credits once, up to 5 members per organization), Pro (2,000 credits a month) and Max (8,000 credits a month). Seat price = monthly credits × the active list price per credit (from the public pricing changelog) minus the seat discount, before VAT. Credit packs fill the organization pool, which pays for API, keys, triggers and guardrails. Prepaid, no auto-renewal. Price history is at `/pricing/changelog`.
- Getting started: there is no time-limited trial. Every new workspace gets a one-time gift credit of 500 credits to enter the panel and test the product; the credit never expires. Never say «۱۵ روز», «آزمایش ۱۵ روزه» or any trial length.
- Entry points on the site: sign-up with gift credit (`/onboarding`), demo booking (`/demo`), pricing, docs, catalog, blog, On-Premise request (`/on-premise`).
- Persian terminology follows `docs/content-glossary.md` (for example «درگاه»، «فضای کاری»، «گزارش رخداد»).

## Capabilities and Constraints

- Next.js App Router, RTL, Estedad (Vazirmatn as fallback). Pages under `app/(site)`; login, signup and onboarding sit outside the marketing shell. Run with `pnpm --filter @shahrah/website ...`.
- All existing routes are kept when pages are added or redesigned.
- Do not state a credit-to-dollar ratio in public copy.
- Not shipped today, never claim: PII masking as a guardrail feature; human approval of tool actions before execution; live intervention from the dashboard (pausing a key or switching a model without a deploy) as a marketed capability; automatic sync of sources such as Drive or Confluence; local open-weight models inside On-Premise; a usage-only cloud plan without subscription.
- Also not to claim: a sign-up without a bank card; filling customers' security questionnaires; security certifications; GodarAI-managed updates in On-Premise.
- Open: the publishable guardrail latency figure and its measurement conditions; whether user sync is available in the cloud or only On-Premise; whether Eitaa is among the supported messengers; plan prices and per-seat details come only from the live pricing API.

## Brand Commitments

- Name: «گدارAI» in Persian prose; `GodarAI` only in code, domains and identifiers.
- Tagline, fixed: «بی‌گدار به آب و AI نزنید!» (gadar = ford), quoted exactly with its half-space and exclamation mark, at most once per page.
- Assistants are called «دستیار». The word «هوشواره» is retired and must not appear in public copy.
- Logo comes only from `/brand` assets through the `GodarAILogo` component; never redrawn or recolored.
- Voice: bold and witty, a sharp, friendly expert. Short sentences; a pun when it earns its place (water, fords, depth: «اول عمق آب را بسنجید.»), never at the reader's expense. Headlines are two short clauses, the second a turn. Address the reader as «شما»; GodarAI speaks as «ما». Buttons are two- or three-word verbs. Persian-native; no calques from English, no stacks of adjectives, no emoji. Wit never bends a fact: every claim still follows Capabilities and Constraints.

## Evidence on Hand

- Real: 1200+ models from 21 providers; initial gift credit with no expiry; On-Premise request page; solutions pages; gateway comparison pages (LiteLLM, Portkey); feature pages for agents, MCP gateway, guardrails, observability, security, model discovery, prompt engineering; real product screenshots of GodarChat and the admin panel (to be supplied by the owner).
- Not available, never fabricate: customer logos, testimonials, case studies, certifications, press, usage statistics, savings percentages.
- The product mock UIs in `components/workplace/mocks.tsx` and the dashboards on feature pages are drawn in code with invented demo numbers; they are illustrative, to be replaced by real screenshots, and never presented as claims.

## Product Principles

1. Lead with the work, not the infrastructure: show an assistant finishing a real task with the team's knowledge and tools before talking about gateways.
2. Control is part of the promise: every capability is paired with who can use it, what it may see, how much it can spend and what is logged.
3. Say only what is true today; mark undecided facts instead of rounding them up.
4. Write for the Persian reader first; an English idea is rewritten, not translated.
5. Two products, two clear paths: the manager evaluates the Workplace, the developer evaluates the Gateway, and no page tries to persuade both at once.

## Accessibility & Inclusion

Persian, right-to-left, mobile-first reading. Light and dark themes are equal citizens: every page must be complete and checked in both. Keep visible keyboard focus, readable contrast (WCAG AA as the working bar) and reduced-motion support. Persian digits in running copy.
