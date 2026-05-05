# Tenant Admin Virtual Models UX Spec

## User Personas
- Tenant Admin: configures resilient routing quickly with minimal mistakes.
- Org/Team Admin: audits enabled aliases and target policies.

## Primary Flows
1. List Virtual Models
- Fast scan table: alias, strategy, targets, status.
- Search by alias.
- Empty state with direct CTA.

2. Create Virtual Model
- Single-page guided form with immediate validation.
- Required order: alias -> targets -> retry/fallback policy -> enabled state.

3. Edit Virtual Model
- Load existing record into same form.
- Dirty-state protection with browser unload and in-app switch confirmation.

4. Details/Preview
- Effective target chain preview chips (order + provider/model).

## Component Map
- `VirtualModelManagement` (page orchestration)
- List module (search + table + selection)
- Form module (alias/strategy/enabled)
- Target chain module (add/remove + up/down reordering)
- Policy editor module (retry attempts, delay, timeout, status code lists)
- Feedback module (success/error alert + loading skeleton)

## Validation & Error UX Matrix
- Alias missing -> inline submit block + error alert.
- Alias invalid pattern -> inline submit block + error alert.
- No targets -> submit block + error alert.
- Missing provider/model on any target -> submit block + error alert.
- Duplicate priorities -> submit block + error alert.
- Retry attempts out of range -> submit block + error alert.
- Retry delay out of range -> submit block + error alert.
- API list/detail/create/update errors -> mapped alert from API payload.
- Disable action -> confirmation prompt.
- Unsaved state navigation/switch -> confirmation prompt.

## Accessibility Notes
- Labeled fields with `Label` + `Input` pairings.
- Keyboard accessible controls for add/remove/move/save.
- Status expressed by text + badge, not color-only.
- Inputs and buttons preserve focusable semantic elements.

## Responsive Behavior
- Desktop: two-column list + editor layout.
- Mobile: stacked cards, full-width controls, maintained target ordering controls.

## References
- TrueFoundry virtual model concepts informed target-order and fallback mental model.
- Shahrah admin shell patterns and existing UI primitives preserved.
