# Component Usage

## PageHeader

Use at the top of route-level pages. It supports title, description, eyebrow, title-level secondary content through `titleAddon`, and page actions.

Use `titleAddon` for lightweight contextual links that must remain on the title row. Keep primary actions in `actions`.

## SectionHeader

Use inside pages to introduce major sections without creating another page title.

## EmptyState

Use for empty lists, missing data, and first-run states. Include an icon only when it clarifies the state.

## AppCard / SettingsCard

Use `AppCard` for product cards and `SettingsCard` for settings pages or grouped account configuration.

Do not put cards inside cards. If a settings surface already has `SettingsCard`, put form controls directly inside it or use an unframed `FormSection` variant.

## MetricCard

Use for dashboard metrics with a label, value, optional description, optional trend, and optional icon.

## CompactValue

Use for large counts and money values when the short form improves scanability, especially in metrics, logs, and traces.

Keep the exact value available through the built-in tooltip/title so users can still inspect the raw amount.

## DataTable Wrapper

The app already has `components/ui/data-table`. Keep using it where behavior is established. Add a product-level wrapper only when multiple pages need the same toolbar, state, and pagination structure.

## Dropdowns And Lists

Use `RowActionsMenu` for table row actions instead of hand-composing `DropdownMenu` in each table. Use it for edit, members, activate/deactivate, revoke, and delete actions.

Use `AppSelectField` for compact static option sets in forms and filter toolbars. It standardizes trigger sizing, `SelectGroup`, RTL direction, and helper text.

Use `FilterCombobox` for searchable filter dropdowns in table/list toolbars. It standardizes `ComboboxInput`, empty state, RTL text alignment, and option truncation.

Keep `components/ui/dropdown-menu`, `components/ui/select`, and `components/ui/combobox` as primitive escape hatches. Route components should prefer the app-ui wrappers unless they need a behavior the wrapper does not support.

## StatusBadge

Use for common statuses: `active`, `inactive`, `pending`, `success`, `warning`, `error`, and `neutral`.

## ConfirmDialog

Use for destructive or high-impact confirmations. It composes the existing shadcn Dialog.

## FormSection

Use to group related fields with a title, description, content, and optional footer actions.

The title is optional for compact sections whose context is already established by the surrounding page or drawer header. Omitting both title and description removes the header column entirely.

Use `actions` for section-level controls that belong on the same row as the section title. Use `footer` for commit or navigation actions after the section content.

For page-level forms, prefer `FormSection` inside the page container. For compact settings cards, use it only when the section title adds useful structure.
