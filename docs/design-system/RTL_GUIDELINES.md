# RTL Guidelines

This application must support RTL layouts.

## Rules

- Prefer logical layout thinking over left/right assumptions.
- Avoid hardcoded `ml-*` and `mr-*` when `ms-*` and `me-*` or layout alternatives are more appropriate.
- Icons that imply direction should be checked in RTL.
- Tables should remain readable in RTL.
- Forms should align labels and helper text correctly.
- Dialogs, dropdowns, popovers, and sidebars must be visually checked in RTL.
- Avoid hardcoded `left-*` and `right-*` unless the behavior is intentional.
- Prefer `text-start` and `text-end` over `text-left` and `text-right` in reusable components.
- Search inputs with icons should use `start-*` and `ps-*` unless the icon is intentionally pinned to a physical side.
- Dropdown item indicators should prefer `start-*`/`end-*` and `ps-*`/`pe-*`; avoid `right-*`, `left-*`, `pr-*`, and `pl-*` in reusable menu primitives.
