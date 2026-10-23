# Dust Product Map (snapshot: Oct 2026)

Source: docs.dust.tt. Full page index: https://docs.dust.tt/llms.txt. Append `.md` to any docs URL for clean text. `UD` below = `https://docs.dust.tt/docs/user-documentation`.

Each area lists **what it is**, **design intent** (why it's built that way), and **UX details** worth borrowing. Re-fetch the live page before citing a specific number or behavior.

## Contents

1. Positioning and core objects
2. Agents and the Agent Builder
3. Default agents
4. Conversations: steering, branching, compaction, multi-agent
5. Conversational building and Sidekick
6. Skills (incl. self-improving skills)
7. Knowledge and data sources
8. Tools, MCP and credentials
9. Models: auto tiers, picker, reasoning effort
10. Memory
11. Triggers and automation
12. Frames
13. Pods
14. Spaces and governance
15. Credits, seats and spend control
16. Integrations (where agents show up)
17. Analytics and audit
18. Rollout and adoption playbook
19. Developer platform
20. Recent changelog themes

---

## 1. Positioning and core objects

- Tagline: **"Multiplayer AI"** — a shared workspace where teams build and use AI agents on company data.
- Objects: Workspace → Spaces (data + tool scope) → Agents (+ Skills, Knowledge, Tools) → Conversations / Pods / Frames / Triggers.
- Roles: **User** (uses agents), **Builder** (creates agents/skills), **Admin** (spaces, connections, tools, billing). Plus group managers for credit limits.
- Mental model: agents are named teammates you `@mention`; pick the right one, or let `@dust` route.
- Doc: `UD/getting-started/intro-to-dust`

## 2. Agents and the Agent Builder

- **Create** from: a conversation ("Convert to agent"), scratch/template, or YAML import.
- Builder sections: **Instructions** → **Tools & Knowledge** (choose space → connection → specific folders, plus a description per source so the agent knows when to use it) → **Settings** (handle = @mention name, description used for routing, tags with "Suggest").
- **Preview pane** next to the builder for continuous testing.
- **Publish**: published = visible to all who can access its data; unpublished = only you + editors. **Editors** list controls who can change it.
- Model button in Instructions (default "Standard" auto tier).
- Agents can be exported/imported as YAML; archived (soft) via API.
- Design intent: building is a writing task, not a config task — instructions first, capabilities second; testing is always one glance away.
- Docs: `UD/agents/create-your-first-agent`, `UD/agents/templates`, `UD/agents/llm-best-practices/how-to-write-effective-instructions`, `UD/agents/troubleshooting-your-agent`, `UD/agents/structured-output-format`

## 3. Default agents

- **@dust** — default agent over company knowledge; routes using agent descriptions; can now create/update/remove triggers by conversation.
- **@deep-dive** — long research agent; reached via the **Go Deep** skill, which hands off while keeping the calling agent's context.
- **@analyst** — data analysis over tables/warehouses.
- **@help** — product help (docs, open-source issues, community).
- Design intent: a useful product before anyone builds anything.
- Docs: `UD/agents/default-agents/dust`, `.../deep-dive-agent`, `.../analyst`, `.../help`, `UD/agents/go-deep`

## 4. Conversations

- **Steering**: live step-by-step view (thinking, tool calls, searches, file ops — each expandable). Send messages while the agent runs; they show as *pending*, queue, and are read after the current round of parallel actions. Stop keeps completed work. One agent per conversation, shown in the composer — no need to re-@mention. Doc: `UD/agents/steering-conversations`
- **Branching**: fork from any agent message into a new conversation seeded by a compacted summary; files/tool outputs copied; "Branched from…" lineage both ways; ~10–20 s, messages queue meanwhile. Doc: `UD/agents/branch-a-conversation`
- **Context compaction**: summarizes long history to keep agents within context. Doc: `UD/agents/context-compaction`
- **Composer**: `/` menu for skills, tools, knowledge (browse spaces/folders inline), `/model` for model + effort. Inline knowledge/tool chips.
- **Conversation timeline** on the right for jumping to earlier messages (Oct 2026). Unified search (conversations, members, settings) via command palette.
- Per-message **credit cost** shown; "Completed in N sec" opens a panel with skills loaded.
- Sharing conversations, rename/delete, thumbs up/down feedback per message (feeds Sidekick and self-improving skills).
- Agents can ask the user a question mid-run and wait for an answer; tool actions can require validation (approve/reject) — both exposed in the API.

## 5. Conversational building and Sidekick

- **Conversational building** (Sep 2026): enable the "Build Agents and Skills" skill with `/`, describe the change; each proposal arrives as a **suggestion card** (can bundle related changes) with Accept/Decline; side panel shows the change in full config. Can create, edit, rename, change model, archive, delete; admins can do workspace-wide refactors (naming conventions, extract a shared skill, migrate models). Permission-checked: the agent can only do what you can. Updates create versions. Doc: `UD/agents/conversational-building`
- **Sidekick** (inside Agent Builder): drafts instructions, recommends tools/skills/model, can research workspace data with permission; suggestions as **inline color-coded diffs** accepted one by one; tools/skills/model are one-click; conflicting manual edits dismiss suggestions; pending suggestions persist; each builder gets a private thread but sees all suggestions. Reads **thumbs feedback and metrics** of existing agents to propose substantive fixes (no cosmetic rewording). Free (doesn't consume credits), capped 100/user/day. Doc: `UD/agents/default-agents/agent-builder-sidekick`
- Design intent: the fastest way to build an agent is to talk to one, but the human stays the editor.

## 6. Skills

- Reusable bundle: name, description, **"when to use"** (drives auto-enabling), guidelines, plus attached tools (inline chips), sub-skills, knowledge (Notion/Drive/Confluence by search or URL).
- Loaded **just-in-time** only when relevant → smaller context. Single source of truth: edit once, all agents update.
- Default skills: Discover Knowledge, Discover Tools, Go Deep, Create Frames.
- Only builders/admins create; everyone can use (if they can access the skill's spaces). Copy link to share; favorites; Manage Skills page with search/filters/usage/bulk update; version history with author/date diff.
- Import from file/ZIP/repo, GitHub action import.
- **Discover Skills**: agents find and enable relevant skills mid-conversation (incl. user favorites).
- **Self-improving skills**: nightly analysis of conversations (thumbs, user corrections, missed tool calls, derailments) → proposed diffs in Skill Builder right pane → editor approves/declines. Opt-in per workspace and per skill; batch mode at half price (not ZDR-covered) or streaming mode (ZDR). No model training.
- Docs: `UD/agents/skills/skills-overview`, `.../skills-management`, `.../skill-examples`, `UD/agents/discover-skills`, `UD/agents/self-improving-skills`, `UD/agents/skill-availability`

## 7. Knowledge and data sources

- Sources: **Connections** (Google Drive, Notion, Confluence, Slack, GitHub, Intercom, Zendesk, Microsoft, Gong, Snowflake, BigQuery…, admin-managed), **Folders** (uploaded files, API-fed), **Websites** (crawled), **Conversation files** (incl. images), custom imports via Zapier/beta importers (Jira, Salesforce, HubSpot, Linear, Dropbox, Front, Guru).
- Agent methods: **Search** (RAG), **Table queries** (SQL over tables/warehouses), **Extract data** (structured extraction across many docs), **Include data** (put recent docs directly in context).
- **Discover Knowledge** skill: agent searches across company docs and warehouses without per-agent wiring.
- Design intent: pick the retrieval method by question shape (lookup vs. aggregation vs. recency), and teach users why LLMs fail at counting/parsing.
- Docs: `UD/data-sources/overview`, `UD/agents/knowledge/*`, `UD/agents/llm-best-practices/understanding-rag`, `.../understanding-llm-limitations`

## 8. Tools, MCP and credentials

- Default tools: web search/browse, file generation, image generation, voice/sound, data viz (now Frames), agent memory, run another agent, wake-ups, **Computer** (sandbox with per-Pod egress domains, env vars, secrets).
- ~50 third-party tools (Gmail, Calendar, Outlook, Teams, Jira, GitHub, HubSpot, Salesforce, Notion, Zendesk, Snowflake, Databricks, Power BI, Miro, Canva…). Admins add **remote MCP servers**; **client-side MCP** for local tools; Dust itself is an **MCP server**.
- **Credentials**: shared (admin OAuth/service account, same access for all) vs personal (each user connects; per-person attribution). Users disconnect/reconnect under Tools & Triggers.
- Tool settings tabs: **General / Tools & Stakes / Availability** — per-tool stakes decide whether actions need user approval; availability can be "on demand" for **Discover Tools / JIT tools**.
- Access: tools workspace-wide via Company space by default; restrict to specific spaces.
- Static egress IPs for allow-listing.
- Docs: `UD/agents/tools/index`, `UD/agents/discover-tools`, `UD/agents/tools/jit-tools`, `UD/admins/tools-management/*`

## 9. Models

- **Auto models**: Basic / Standard (default) / Premium (+ Ultra tier for most expensive models) — curated ordered candidate lists with fixed reasoning effort; first available candidate used. Not per-request routing.
- **Outage fallback**: auto tiers skip degraded providers and show which replacement is active; pinned models warn instead; failed messages offer "Retry with Standard tier"; never switches mid-run (could repeat side effects).
- **Picker**: in the composer and builder; applies to the next message only; "More models" for specific ones; `/model` keyboard command; reasoning effort chips (none → maximal), only supported levels shown.
- **Model access tiers**: admins cap models/efforts per workspace, group or user. EU-hosted-only option.
- Design intent: users choose *quality/cost level*, not vendor names; reliability is a product feature.
- Docs: `UD/agents/model-selection`, `UD/admins/usage-seats-and-credits/model-access-tiers`, `UD/admins/admin-governance/eu-hosted-models`

## 10. Memory

- Memory tool per agent and per user; agent decides what to save or user asks ("remember that…"); retrieved contextually.
- Users view/delete in the agent details drawer → Memory tab; private to the user; deletion is permanent.
- Doc: `UD/agents/tools/agent-memory`

## 11. Triggers and automation

- **Schedule** (described in plain language) and **Webhook** (GitHub, Jira, Linear, Zendesk, Fathom, or custom; payload filters; rate limit 42/24h default).
- Each run = a new conversation; default in your history or in a chosen **Pod** for team visibility.
- Triggers belong to their creator; credits from personal or workspace pool per trigger; review/disable with credit usage.
- **Wake-ups** tool: an agent schedules its own follow-up.
- Can be created by talking to @dust.
- Docs: `UD/agents/triggers/index`, `.../schedules`, `.../webhooks/*`, `.../credits-usage`, `UD/agents/tools/wake-ups`

## 12. Frames

- Interactive outputs (dashboards, reports, calculators, presentations) instead of static text; hover, filter, switch views.
- Created via **Create Frames** skill; edit by asking; **Revert last change** steps back versions.
- Share via token-gated links (public, company email domains, or internal, per admin policy); white-labeled Frames; export PDF/PNG/template code.
- Can be pinned as a Pod banner.
- Docs: `UD/agents/frames/overview`, `.../white-labeled-frames`

## 13. Pods

- Shared workspace for a team **and its agents**: **Conversations** (visible to all members, auto-indexed as context), **Tasks**, **Files** (uploads, linked company data, agent artifacts), **Frames** (pinned banner).
- Principle: *"everything a human can do, an agent can do"* — via the default **Pods skill** agents list Pods, post, manage tasks, read/upload files, search, and (as Editor) manage members.
- **Tasks**: description doubles as the agent's brief; optional assignee (human or agent); Open → In progress (linked conversation) → Done (agent adds rationale); ▶ button to start an agent on a task; Mine/All filter; sync with Jira/Asana/Linear via scheduled agent (warns about two-way loops).
- Visibility: Open (anyone joins) or Restricted (invite); roles Member / Editor.
- Example patterns: shared asset library, personal second brain, project management, support tickets, one Pod per customer, editorial production, competitive intelligence.
- Analytics: Pods ranked by credit consumption.
- Docs: `UD/pods/*`

## 14. Spaces and governance

- **Company Data** (default, everyone), **Open spaces**, **Restricted spaces** (members or IdP-provisioned groups). Admins create; connection data only by admins; folders/websites by builders (open) or any member (restricted).
- Agents/skills inherit access from the spaces their data/tools use.
- SSO, SAML, SCIM provisioning, group managers (team leads manage membership and credit limits), audit logs with event reference, account merging.
- Plan limits on number of spaces.
- Docs: `UD/admins/spaces-management`, `UD/admins/admin-governance/*`, `UD/admins/audit-logs/*`

## 15. Credits, seats and spend control

- Cost per interaction = **token credits** (model × context) + **action credits** per tool tier: Free (0: internal/orchestration/Pods), Basic (1: run agent, web search, Frame edit), Advanced (3: retrieval, Frame creation, file/image/audio generation, third-party/external MCP).
- Seats: Free (500 one-time), Pro (8,000/mo), Max (40,000/mo); credits belong to the user; no rollover. Workspace pool + overage.
- **Spend checkpoint**: a task pauses at 600 credits and asks the starter to continue or stop; admins toggle; per-agent exemption.
- Group managers set limits; temporary personal limits that auto-revert; credit requests.
- Per-message cost visible to users; breakdown by tools/models/agents; "Optimize credit consumption" and team training guides.
- Docs: `UD/admins/usage-seats-and-credits/*`, `https://docs.dust.tt/analytics`

## 16. Integrations

- Slack (mentions, auto-reply in channels, auto-join, workflows), Microsoft Teams, Zendesk sidebar, email-to-agent (send/forward), browser extension, Raycast (bind agent to a shortcut that rewrites selected text), Google Sheets add-on, meeting transcripts, Zapier/Make/n8n/Power Automate, Dust MCP server, CLI.
- Design intent: the agent comes to the work surface; Dust web app is the builder's home, not necessarily the user's.
- Docs: `UD/agents/integrations/*`

## 17. Analytics and audit

- Workspace analytics (export CSV/JSON), consumption analytics per call, Pods tab by consumption, agent/skill usage in Manage tables, audit logs, conversation consumption API, feedback API.

## 18. Rollout and adoption playbook

- Rollout guide: Welcome → Launch strategy → Team onboarding material → **Build your builder community** (find tinkerers per team, builder hub channel, meetups, template library, train-the-trainer) → Admin setup → **Designing use cases** (3–5 low-effort/high-value, one per pilot team, discovery workshop) → Educate & activate → Measure & expand (user survey).
- Use-case guides by function: Sales, Support, Marketing, Engineering, Data, Knowledge, People, Product, Collaboration.
- Design intent: Dust sells a change in how teams work; the product ships the playbook with it.
- Docs: `UD/getting-started/dust-rollout-guide/*`, `UD/getting-started/use-cases-and-guides/*`

## 19. Developer platform

- REST API (conversations with SSE events, messages, content fragments, files, feedback, mentions, agents CRUD/YAML, skills import, data sources/tables/rows upsert, search, spaces, triggers, client-side MCP registration), JavaScript SDK, CLI, OpenAPI/Postman, rate limits, Dust-Labs repo. Legacy "Dust Apps" deprecated in favor of agents + MCP.
- Docs: `https://docs.dust.tt/docs/developer-platform/*`, `https://docs.dust.tt/api-reference/*`

## 20. Recent changelog themes (Aug–Oct 2026)

- Building by conversation (agents, skills, triggers) with review cards.
- Cost governance everywhere: circuit breaker at 600 credits, agent/workspace checkpoints, Ultra tier, temporary limits, Pods consumption analytics.
- Model reliability: auto-tier outage fallback, reasoning effort levels in picker, `/model`.
- Discovery & navigation: unified search, conversation timeline, searchable Manage tables, `/` browsing of spaces/folders, inline knowledge chips.
- Localization: French UI with personal and workspace default language.
- Frames sharing to company email domains.
- Changelog: https://docs.dust.tt/docs/changelog (very long — fetch with offset)
