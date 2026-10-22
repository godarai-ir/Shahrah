---
version: 2
slug: "app-pages-home-tsx"
primary_target: "app/pages/Home.tsx"
related_targets: ["app/(site)/page.tsx"]
---

# Surface brief: Home

Scope: the marketing home (`app/pages/Home.tsx` and its section components). Mode: Persuade.

Audience and job: the technology or digital-transformation lead at an Iranian company decides whether to start the 15-day trial or book a demo. A developer finds their own path through the gateway card. Proof available today is only product truth (see PRODUCT.md); customer proof is absent and stays out of the page until it exists.

Constraints: Persian and RTL; Vazirmatn; copy follows docs/content-glossary.md; only claims confirmed in docs/site-architecture/04-open-questions.md; no invented customers, numbers or certifications; «بی‌گدار به آب نزنید» only in the hero and the closing band.

Memorable moment: the first product window, resting on its panel and crossing its edge, where an assistant's tool rows arrive one by one and the result resolves to a check, once, when it comes into view.

## Direction contract

THESIS: The Lit Workshop. The page shows real team work laid out in daylight, with the controls in plain sight, and refuses gradient heroes, decorative illustration and a wall of saturated panels.

OWN-WORLD: Cool daylight neutrals, one steady blue accent, four pastel tones at matched lightness that each mean one differentiator, 24px forms, soft ink-tinted shadows on product windows and interactive cards, Vazirmatn at large confident sizes.

STORY: The visitor sees an assistant finishing a real support task, understands there are two ways in (workplace for everyone, gateway for the technical team), meets the four controls on their tones, finds their own team's work, checks the enterprise controls, and starts the trial or books a demo.

FIRST VIEWPORT: Heading at display size with one accented phrase («هر روز») and the tagline, two calm sentences, the primary action at the reading start and the demo link beside it; the visual column holds the product tour slot, falling back to a drawn window that crosses the panel's edge.

FORM: Hero, one task end to end, two ways, four pillars, team tabs, enterprise checklist, three steps, grouped FAQ, dark closing band (DESIGN.md, docs/site-architecture/prompts/02-home.md).

OPEN: real GodarChat screenshots and the product tour for `heroDemo`; team media in `lib/home-media.ts`; links to the governance, cost-control and team pages appear automatically when their routes leave `pendingRoutes`.
