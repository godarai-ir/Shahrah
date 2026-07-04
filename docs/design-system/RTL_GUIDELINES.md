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
