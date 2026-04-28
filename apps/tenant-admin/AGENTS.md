# AGENTS.md — `apps/tenant-admin`

## Purpose
This file defines **mandatory** quality rules for UI copy, Persian translations, and i18n behavior in `tenant-admin`.
Any agent working in this app must follow these rules by default without needing extra user reminders.

## Non-Negotiable Defaults
1. Never ship hardcoded user-facing strings in components.
2. Always use `next-intl` keys (`useTranslations(...)`) for UI text, labels, helper text, buttons, empty states, and messages.
3. Preserve domain language for **LLM Models + Providers** (not generic wording).
4. Treat copywriting quality as product behavior, not as a cosmetic task.

## i18n Source of Truth
1. Primary locale is currently `fa` (see `i18n/request.ts`).
2. Keep `messages/fa.json` complete and coherent.
3. Keep `messages/en.json` structurally aligned with `fa.json` when English exists.
4. When adding a new key in one locale, add the corresponding key in the other locale in the same change.

## Copywriting Style (Product)
Use precise product terms:
- `Provider account`
- `Connection validation`
- `Enable models`
- `Usage access policy`
- `Account managers`
- `Users allowed to consume models`

Prefer short, operational, UI-native wording. Avoid verbose prose.

## Persian (`fa`) Quality Rules
1. Use natural Persian product copy, not literal word-by-word translation.
2. Keep technical terms where appropriate (`API Key`, `Secret`, `Discovery`, `Provider`).
3. Avoid mixed/awkward phrasing and inconsistent terminology across pages.
4. Keep action verbs clear and consistent across buttons/messages.

## Encoding and Corruption Guardrails
1. Files must be UTF-8.
2. Do not allow mojibake (`Ø...`, `Ù...`, replacement chars `�`) or `????` strings.
3. Before finalizing i18n changes, scan locale files for corrupted strings.

## UX Text Consistency
For each workflow, ensure:
1. Step titles, descriptions, CTA labels, and validation messages form one coherent flow.
2. Error and success tones are consistent across all steps.
3. Empty states are explicit and actionable.

## Provider-Specific Forms
For provider connection setup:
1. Step 1: provider selection only.
2. Step 2: provider-specific connection fields based on auth mode.
3. Step 3: models + access policy.
4. Text must reflect this flow exactly in both language files.

## Implementation Checklist (Required on i18n edits)
1. Add/update keys in `fa.json`.
2. Mirror structure in `en.json`.
3. Replace any newly introduced hardcoded UI text with translation keys.
4. Run lint for `tenant-admin`.
5. Verify no corrupted locale values.

## PR/Delivery Notes
When reporting i18n work:
1. List changed message namespaces.
2. Mention if any keys were renamed or moved.
3. Confirm hardcoded strings were not introduced.
4. Confirm corruption scan passed.

