---
version: 3
slug: "app-pages-home-tsx"
primary_target: "app/pages/Home.tsx"
related_targets: ["app/(site)/page.tsx"]
---

# Surface brief: Home

Scope: the marketing home (`app/pages/Home.tsx` and its section components). Mode: Persuade.

Audience and job: the technology or digital-transformation lead at an Iranian company decides whether to start the 15-day trial or book a demo. A developer finds their own path through the gateway card. Proof available today is only product truth (see PRODUCT.md); customer proof is absent and stays out of the page until it exists.

Constraints: Persian and RTL; Estedad; copy follows docs/content-glossary.md in a bold, witty voice; only claims confirmed in docs/site-architecture/04-open-questions.md; no invented customers, numbers or certifications; the tagline «بی‌گدار به آب و AI نزنید!» once, as the hero line.

Memorable moment: the tagline in heavy Estedad on the navy field, beside a product window on a sand tile whose tool rows arrive one by one and resolve to a check.

## Direction contract

THESIS: The Marked Crossing. GodarAI marks where it is safe to cross: navy fields, warm paper, one signal orange per section, and headlines with a turn.

OWN-WORLD: Navy for the hero, the closing card and the footer; paper and linen sections alternating; flat white tiles with hairlines; sand under product windows; Estedad heavy for display and light for leads; no shadows.

STORY: The visitor sees an assistant finishing a real support task, understands there are two ways in (workplace for everyone, gateway for the technical team), meets the four controls as flat tiles, finds their own team's work, checks the enterprise controls, and starts the trial or books a demo.

FIRST VIEWPORT: On navy: the tagline at display size, a light lead, the orange primary and an on-navy outline button; the visual column holds the product tour slot, falling back to a drawn window centred on a sand tile.

FORM: Hero, one task end to end, two ways, four pillars, team tabs, enterprise checklist, three steps, grouped FAQ, dark closing band (DESIGN.md, docs/site-architecture/prompts/02-home.md).

OPEN: real GodarChat screenshots and the product tour for `heroDemo`; team media in `lib/home-media.ts`; links to the governance, cost-control and team pages appear automatically when their routes leave `pendingRoutes`.
