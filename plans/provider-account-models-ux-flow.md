# UX Flow Spec: Setup Provider Account & Manage Models (Shahrah)

## 1) Product Intent

Design a professional, low-friction flow for:
- connecting provider accounts (BYOK or virtual-key mode),
- adding and managing models,
- controlling access (collaborators),
- validating capability + pricing correctness before runtime.

Primary UX goal: users should complete setup in one pass, with zero ambiguity about what is required, what is optional, and what affects runtime access.

---

## 2) Key UX Problems in Current Experience

Current flow issues (from current screen and behavior):
- unclear hierarchy between Provider, Account, and Model.
- account setup and model setup are fragmented, not guided.
- capability/type selection is under-explained and easy to misconfigure.
- pricing mode (public/private) consequences are not visible enough.
- collaborator/access implications are hidden and hard to verify.
- weak state feedback: users cannot easily answer “is this account production-ready?”.

---

## 3) UX Principles

- Progressive disclosure: show essential actions first; reveal advanced options contextually.
- Single source of truth: one clear “effective access & capability” summary per account.
- Inline validation before save: prevent invalid state, don’t punish after submit.
- Status-first UI: readiness, health, and policy impact always visible.
- Safe defaults: public pricing, chat capability where valid, manager/user scaffolding.

---

## 4) Information Architecture

Use one top-level area: `AI Gateway > Providers & Models`

### Left Rail
- Connected Providers
  - OpenAI
  - Anthropic
  - Gemini
  - ...
- Add Provider

### Main Panel Modes
1. Provider Overview (list of accounts under selected provider)
2. Account Workspace (tabs)
   - `Models`
   - `Access`
   - `Credentials`
   - `Observability`

This removes context switching and keeps all account operations in one place.

---

## 5) Core User Flows

## Flow A: First-time Setup (Happy Path)

### Step A1: Select Provider
- Entry CTA: `Add Provider Account`
- Modal/drawer: provider picker with search + quick info (auth modes, supported APIs).
- On select: continue to account setup wizard.

### Step A2: Account Basics
Fields:
- Account Display Name (required, unique)
- Auth Mode (API Key / SigV4 / Service Account / No Credential)
- Toggle: `Virtual-key-only mode`

UX behavior:
- If `Virtual-key-only mode` ON, credential fields collapse and marked optional.
- Real-time name uniqueness check.

### Step A3: Credentials
- Credential input + optional secret reference.
- CTA: `Test Credentials`
- Result states:
  - Success: green status + timestamp
  - Warning: partial capability discovery
  - Fail: actionable error + retry

### Step A4: Add Models
Model selector table:
- Search model
- Select model
- Select capability/types (multi-select chips)
- Pricing mode (Public / Private)
- Private pricing input (conditional)

Validation rules inline:
- at least one type per selected model.
- private mode requires input or output price.
- provider capability mismatch blocked before save.

### Step A5: Access Setup
Two required zones:
- Managers (user/team only)
- Users (user/team/everyone)

Interactions:
- searchable principal picker
- dedupe prevention
- role explanation tooltips

### Step A6: Review & Activate
Summary panel:
- credential status
- selected models count
- capability coverage
- pricing coverage (resolved/unresolved)
- access snapshot

Primary CTA: `Activate Account`
Secondary: `Save Draft`

---

## Flow B: Manage Existing Account

From account card, open Account Workspace:

### Tab: Models
- Table columns:
  - Model display name
  - Effective capabilities
  - Pricing mode
  - Pricing resolved badge
  - Runtime status
- Per-row actions:
  - Edit
  - Disable/Enable
  - Remove

Bulk actions:
- Enable selected
- Disable selected
- Change pricing mode

### Tab: Access
- Managers and Users sections with chips
- Effective access preview:
  - test with `key_type/user_id/team_id`
  - shows allowed models immediately

### Tab: Credentials
- masked credential state
- rotate/update credential
- last test result
- test history (latest 3)

### Tab: Observability
- Effective model list API output visualization
- runtime denied reasons:
  - not in allowlist
  - capability mismatch
  - unknown model

---

## Flow C: Add Model to Existing Account

Entry points:
- `Add Model` button in Models tab
- empty-state CTA

Process:
1. Choose model from catalog
2. Select capabilities
3. Set pricing mode
4. Save
5. Immediate row insertion + optimistic success toast

If save fails:
- preserve draft
- highlight invalid fields
- show precise message from backend

---

## Flow D: Capability Mismatch Recovery

Scenario: model is selected for endpoint-incompatible type.

UX handling:
- hard validation on save.
- message format:
  - `Model X does not support capability Y for provider Z.`
- Quick fix actions:
  - remove unsupported capability
  - change model

---

## 6) Screen-by-Screen Requirements

## Screen 1: Provider Overview

Each account card shows:
- account name
- credential status
- models count
- access count
- last updated
- readiness badge:
  - `Ready`
  - `Needs Credentials`
  - `Needs Access`
  - `Pricing Incomplete`

Card actions:
- Open
- Quick Test
- Duplicate
- Delete

## Screen 2: Account Workspace / Models

Top bar:
- account name + provider icon
- Copy FQN
- readiness badge

Subheader KPIs:
- enabled models
- unresolved pricing
- capabilities spread

Main table with sticky header and row-level quick actions.

## Screen 3: Add/Edit Model Drawer

Sections:
1. Model identity
2. Capabilities
3. Pricing
4. Advanced settings

Footer:
- Save
- Save & Add Another
- Cancel

## Screen 4: Access Management

Two-column layout:
- Managers
- Users

Below: Effective Access Preview panel with filter inputs (`key_type`, `user_id`, `team_id`) and resulting model list.

## Screen 5: Credentials

- Auth mode display
- credential input/update area
- Test Credentials button
- latest test result card

---

## 7) State Model (UX)

All pages must handle:
- loading (skeletons)
- empty (helpful CTA)
- error (recoverable with retry)
- success (toast + persistent status)
- partial success (warning banner)

Recommended persistent banners:
- `Provider policy enforcement is active for this account.`
- `Some models have unresolved pricing.`

---

## 8) Critical Microcopy

- `Virtual-key-only mode: Applications can use platform keys without storing provider credential in this account.`
- `Public pricing uses catalog rates and may change over time.`
- `Private pricing overrides catalog rates for this account.`
- `Managers can edit account and access settings.`
- `Users can consume all enabled models in this account.`
- `At least one manager is required.`

---

## 9) Validation & Error UX

Inline validation before submit:
- account name required and unique
- model type required
- private pricing completeness
- manager role cannot be everyone

Error mapping examples:
- `invalid_model_types` -> highlight capability chips
- `invalid_models` -> highlight model selector row
- credential test fail -> credentials panel alert

Never show raw backend error unless in expandable “Technical details”.

---

## 10) Accessibility & Responsiveness

Accessibility:
- keyboard-only complete flow
- visible focus states
- WCAG AA contrast
- ARIA labels for icon-only buttons

Responsive:
- desktop: split panels
- tablet: collapsible left rail
- mobile: stacked cards + bottom-sheet drawers

---

## 11) Instrumentation Events

Track key events:
- `provider_account_create_started`
- `provider_account_created`
- `provider_account_credential_tested`
- `provider_account_models_saved`
- `provider_account_access_saved`
- `provider_account_activation_completed`
- `provider_account_setup_abandoned`

Attach metadata:
- provider
- auth_mode
- selected_model_count
- unresolved_pricing_count
- time_to_complete

---

## 12) Acceptance Criteria (UX)

1. New tenant admin can complete first account setup in < 5 minutes without docs.
2. Validation prevents runtime-breaking model/capability misconfiguration.
3. Access impact is visible before saving via effective-model preview.
4. User can identify account readiness state at a glance.
5. Error recovery preserves user input and supports one-click retry.

---

## 13) Implementation Phasing (Design Delivery)

### Phase UX-1 (Foundation)
- New IA + Provider Overview + Account Workspace shell
- Readiness badges + baseline states

### Phase UX-2 (Guided Setup)
- Setup wizard with credentials/model/access/review
- Inline validations and structured errors

### Phase UX-3 (Operational Excellence)
- Effective access preview
- observability tab
- bulk model actions

### Phase UX-4 (Polish)
- micro-interactions, keyboard polish, mobile hardening
- copy and localization QA

---

## 14) Design Handoff Checklist

- Full user flow diagrams (A/B/C/D)
- High-fidelity desktop/tablet/mobile screens
- Component inventory mapped to design system
- Interaction specs (states, transitions, errors)
- Copy deck (EN/FA)
- Analytics event schema
- QA checklist (functional + accessibility)
