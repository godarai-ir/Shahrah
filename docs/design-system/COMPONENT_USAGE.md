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

Use `DataListPanel` when a list page needs the standard toolbar, loading/empty/content state switch, and an optional pagination footer. Keep filtering, fetching, and row rendering in the feature; pass those results into the panel slots.

Use `TablePagination` for previous/next pagination. Pass localized summary and button labels from the feature namespace, and provide page/page-size values only when they add useful context.

Use the feature-level `GatewayPolicyRulesList` for the shared Rate Limit and Budget Limit page shell. It composes `PageHeader`, entitlement-aware create actions, notice feedback, `DataListPanel`, the create/edit `Sheet`, `ConfirmDialog`, and the plan-upgrade dialog.

Keep policy tables, columns, row actions, API calls, rule forms, and domain types in their owning feature. The shared shell must not introduce a generic gateway-rule model.

## Dropdowns And Lists

Use `RowActionsMenu` for table row actions instead of hand-composing `DropdownMenu` in each table. Use it for edit, members, activate/deactivate, revoke, and delete actions.

Use `AppSelectField` for static option sets in forms and filter toolbars. It standardizes accessible labels, helper/error text, disabled state, RTL popup direction, `sm`/`default` control sizes, and guarantees every `SelectItem` is rendered inside a `SelectGroup`.

Pass `options` for a flat list or `groups` when options need section labels and separators. Individual options can be disabled. Use `optionTextDir="ltr"` for technical values such as model names while keeping the popup itself RTL.

With React Hook Form, keep the component controlled: pass `field.value`, `field.onChange`, and `field.onBlur` from `Controller`, and pass `fieldState.error?.message` to `error`. For simple forms, `watch` plus `setValue` with `shouldDirty`, `shouldTouch`, and `shouldValidate` is acceptable.

Use `FilterCombobox` for searchable filter dropdowns in table/list toolbars. It standardizes `ComboboxInput`, empty state, RTL text alignment, and option truncation.

Keep `components/ui/dropdown-menu`, `components/ui/select`, and `components/ui/combobox` as primitive escape hatches. Route components should prefer the app-ui wrappers unless they need a behavior the wrapper does not support.

## StatusBadge

Use for common statuses: `active`, `inactive`, `pending`, `success`, `warning`, `error`, and `neutral`.

## ConfirmDialog

Use for destructive or high-impact confirmations. It composes the existing shadcn Dialog.

Use `open` and `onOpenChange` for feature-controlled dialogs, or pass `trigger` for the standard trigger-based flow. `onConfirm` may return a promise; provide `pendingLabel` so the dialog disables dismissal and communicates progress while the operation runs.

The dialog closes automatically after a successful confirmation and stays open when the promise rejects. Use `onConfirmError` to surface rejected operations. Set `closeOnConfirm={false}` only when the feature itself closes the dialog after updating its domain state.

## CommercialPlanUpgradeDialog

Use when an action is unavailable because its commercial-plan entitlement is disabled. Pass localized title, description, cancel, and upgrade labels; the component standardizes the dialog composition and links to `/billing/plan` by default.

Keep the triggering action visible while entitlements load, but disable it until the entitlement response is resolved. When the feature is unavailable, open this dialog instead of starting the restricted workflow.

## FormSection

Use to group related fields with a title, description, content, and optional footer actions.

The title is optional for compact sections whose context is already established by the surrounding page or drawer header. Omitting both title and description removes the header column entirely.

Use `actions` for section-level controls that belong on the same row as the section title. Use `footer` for commit or navigation actions after the section content.

For page-level forms, prefer `FormSection` inside the page container. For compact settings cards, use it only when the section title adds useful structure.

## Field Family

Use `FieldGroup` to organize related controls and `Field` for each label/control/helper/error unit. Use `FieldLabel`, `FieldDescription`, and `FieldError` instead of hand-composed label and paragraph markup.

For validation, set `data-invalid` on `Field` and `aria-invalid` on the control. For disabled fields, set `data-disabled` on `Field` and `disabled` on the control.

Product wrappers such as `IdentifierField` and `AppSelectField` already compose the Field family; place them directly inside a `FieldGroup` rather than wrapping them in another `Field`.

## Permission Editing

Use `PermissionEditor` for controlled permission selection with search, selected/all views, grouped checkboxes, and group-level select/clear actions. Keep the permission catalog and selected keys in the feature; pass localized labels and formatting callbacks into the component.

Use `PermissionSummary` beside the editor when a role workflow needs selected-permission and covered-group counts. It may expose a clear-all action, but submission and validation remain responsibilities of the feature form.

Permission groups use semantic `FieldSet`/`FieldLegend`, and each checkbox uses the Field family. Technical permission keys remain LTR while group and description direction follows the page locale.
