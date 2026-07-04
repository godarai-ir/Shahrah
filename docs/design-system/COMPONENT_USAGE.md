# Component Usage

## PageHeader

Use at the top of route-level pages. It supports title, description, eyebrow, and actions.

## SectionHeader

Use inside pages to introduce major sections without creating another page title.

## EmptyState

Use for empty lists, missing data, and first-run states. Include an icon only when it clarifies the state.

## AppCard / SettingsCard

Use `AppCard` for product cards and `SettingsCard` for settings pages or grouped account configuration.

## MetricCard

Use for dashboard metrics with a label, value, optional description, optional trend, and optional icon.

## DataTable Wrapper

The app already has `components/ui/data-table`. Keep using it where behavior is established. Add a product-level wrapper only when multiple pages need the same toolbar, state, and pagination structure.

## StatusBadge

Use for common statuses: `active`, `inactive`, `pending`, `success`, `warning`, `error`, and `neutral`.

## ConfirmDialog

Use for destructive or high-impact confirmations. It composes the existing shadcn Dialog.

## FormSection

Use to group related fields with a title, description, content, and optional footer actions.
