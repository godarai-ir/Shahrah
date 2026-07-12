# Component Usage

## PageHeader

Use at the top of route-level pages. It supports title, description, eyebrow, title-level secondary content through `titleAddon`, and page actions.

Use `titleAddon` for lightweight contextual links that must remain on the title row. Keep primary actions in `actions`.

## PageContainer

Use `PageContainer` as the outer layout for route-level pages. The default `wide` size, standard padding, and default spacing produce the established `max-w-7xl` page rhythm. Use `readable` for focused content and `full` for embedded workflows that must consume their parent width.

Set `padding="none"` only when the surrounding shell or composition already owns padding. Use `compact`, `default`, or `relaxed` spacing instead of rebuilding page-level gap classes. Do not nest `PageContainer` components; feature sections should use `FormSection`, `AppCard`, or normal layout elements.

## SectionHeader

Use inside pages to introduce major sections without creating another page title.

## EmptyState

Use for empty lists, missing data, and first-run states. Include an icon only when it clarifies the state.

Use `variant="surface"` for standalone page or panel states, `variant="inline"` inside an existing card/picker, and `variant="table"` inside a spanning `TableCell`. The inline and table variants intentionally avoid adding another border or background.

Use `density="compact"` for constrained pickers and small panels. Keep “no data” and “no search results” copy distinct in the owning feature, and preserve the table header when rendering a table empty state.

## OperationNotice

Use `OperationNotice` for transient operation feedback represented by a `kind` and `message`. Pass localized labels for success, error, warning, and info; keep notice state and dismissal behavior in the owning feature. Use raw `Alert` for persistent explanatory callouts that are not operation results.

## AppCard / SettingsCard

Use `AppCard` for product cards and `SettingsCard` for settings pages or grouped account configuration.

Do not put cards inside cards. If a settings surface already has `SettingsCard`, put form controls directly inside it or use an unframed `FormSection` variant.

## MetricCard

Use for dashboard metrics with a label, value, optional description, optional trend, and optional icon.

## CompactValue

Use for large counts and money values when the short form improves scanability, especially in metrics, logs, and traces.

Keep the exact value available through the built-in tooltip/title so users can still inspect the raw amount.

## DefinitionList

Use `DefinitionList` with `DefinitionListItem` for compact label/value metadata in detail pages and settings panels. Choose responsive `columns`, and use item-level `surface` only when each value needs a bordered visual container.

Technical values use `mono` with `valueDir="ltr"`. Use `wrap` for long identifiers that must remain fully visible; truncated string values expose their complete value through `title`. Values may be React nodes, so compose `CompactValue`, `StatusBadge`, or `CopyableValue` when their established behavior is needed.

Do not replace primary KPI cards with `DefinitionList`; continue using `MetricCard` for prominent metrics.

## DataTable Wrapper

The app already has `components/ui/data-table`. Keep using it where behavior is established. Add a product-level wrapper only when multiple pages need the same toolbar, state, and pagination structure.

Use `DataListPanel` when a list page needs the standard toolbar, loading/empty/content state switch, and an optional pagination footer. Keep filtering, fetching, and row rendering in the feature; pass those results into the panel slots.

Within table bodies, use `TableLoadingRows` for initial skeleton rows and `TableEmptyRow` for a correctly spanning `EmptyState`. Pass a localized loading label and the exact visible column count; feature-owned empty actions remain supported through `EmptyState` props.

Use `DateTimeValue` for visible timestamps so the UI emits a semantic `time` element, localized digits/calendar, a stable invalid-value fallback, and tabular numerals. Use the shared `formatDateTime` utility only when a formatted string is required by an interpolation or non-React API.

Use `TablePagination` for previous/next pagination. Pass localized summary and button labels from the feature namespace, and provide page/page-size values only when they add useful context.

Use the feature-level `GatewayPolicyRulesList` for the shared Rate Limit and Budget Limit page shell. It composes `PageHeader`, entitlement-aware create actions, notice feedback, `DataListPanel`, the create/edit `Sheet`, `ConfirmDialog`, and the plan-upgrade dialog.

Keep policy tables, columns, row actions, API calls, rule forms, and domain types in their owning feature. The shared shell must not introduce a generic gateway-rule model.

## Dropdowns And Lists

Use `RowActionsMenu` for table and card action menus instead of hand-composing `DropdownMenu` in each feature. Use callback actions for edit, activate/deactivate, revoke, and delete; use `href` for navigation and add `external` only for links that open in a new tab.

Start a new action group with `separatorBefore`. For asynchronous actions, pass `pending` and `pendingLabel`; the item stays disabled until the operation finishes. Set `stopPropagation` when the menu is rendered inside a clickable row or card.

Keep filter, column-selection, checkbox, radio, and other stateful menus on the underlying `DropdownMenu` primitives because they are not row-action menus.

Use `AppSelectField` for static option sets in forms and filter toolbars. It standardizes accessible labels, helper/error text, disabled state, RTL popup direction, `sm`/`default` control sizes, and guarantees every `SelectItem` is rendered inside a `SelectGroup`.

Pass `options` for a flat list or `groups` when options need section labels and separators. Individual options can be disabled. Use `optionTextDir="ltr"` for technical values such as model names while keeping the popup itself RTL.

With React Hook Form, keep the component controlled: pass `field.value`, `field.onChange`, and `field.onBlur` from `Controller`, and pass `fieldState.error?.message` to `error`. For simple forms, `watch` plus `setValue` with `shouldDirty`, `shouldTouch`, and `shouldValidate` is acceptable.

Use `FilterCombobox` for searchable filter dropdowns in table/list toolbars. It standardizes `ComboboxInput`, empty state, RTL text alignment, and option truncation.

Keep `components/ui/dropdown-menu`, `components/ui/select`, and `components/ui/combobox` as primitive escape hatches. Route components should prefer the app-ui wrappers unless they need a behavior the wrapper does not support.

## StatusBadge

Use for common statuses: `active`, `inactive`, `pending`, `success`, `warning`, `error`, and `neutral`.

## StatusOptionGroup

Use `StatusOptionGroup` when one setting must be selected from a small set of visually descriptive status options. Pass the controlled `value`, localized group `label`, option metadata, and `onValueChange`; use `columns` only to adjust responsive layout.

The component composes a single-selection `ToggleGroup`, preserves one selected value, and provides roving keyboard focus. Keep persistence and save state in the feature, pass `disabled` while loading or saving, and do not use standalone status-option buttons.

## Spinner And Async Actions

Use `Spinner` for indeterminate loading. Inside a button, compose it with `data-icon="inline-start"`, keep the localized pending label visible, disable the button, and set `aria-busy` on the button. Do not add `loading` or `isPending` props to `Button`.

For standalone loading feedback, pass a localized `label` to `Spinner`; this exposes `role="status"`. When an existing parent already owns the status message, omit the label so the decorative spinner remains hidden from assistive technology.

## ConfirmDialog

Use for destructive or high-impact confirmations. It composes the existing shadcn Dialog.

Use `open` and `onOpenChange` for feature-controlled dialogs, or pass `trigger` for the standard trigger-based flow. `onConfirm` may return a promise; provide `pendingLabel` so the dialog disables dismissal and communicates progress while the operation runs.

The dialog closes automatically after a successful confirmation and stays open when the promise rejects. Use `onConfirmError` to surface rejected operations. Set `closeOnConfirm={false}` only when the feature itself closes the dialog after updating its domain state.

## TypedConfirmDialog

Use `TypedConfirmDialog` only for high-impact destructive actions that require typing an exact technical identifier such as a team or virtual-user alias. Keep ordinary confirmations on `ConfirmDialog`.

Pass the localized input instruction separately from `expectedValue`. Confirmation remains disabled until the trimmed value matches exactly. Async confirmation disables dismissal, shows the shared `Spinner`, closes after success, and stays open when the promise rejects. Technical confirmation values default to LTR.

## Copyable Values And One-Time Secrets

Use `EntityIdentityHeader` inside entity detail cards when a title, technical identifier, status, and actions need one responsive hierarchy. Pass copy labels only when the identifier should be copyable; the component composes `CopyAction` and keeps identifiers LTR.

Use `CopyableValue` for technical values that need consistent clipboard behavior. It composes `InputGroup`, keeps raw values LTR, provides modern clipboard and legacy fallback behavior, announces successful copies, and optionally supports masked values with an explicit reveal action.

Use `CopyAction` when the value is already visible and only a compact copy affordance is needed. It owns clipboard fallback, pending/copied state, accessible labels, and reset timing; keep localized success or error notices in the owning feature through its callbacks.

Use `OneTimeSecretDialog` when a raw token or credential is returned only once. Pass the secret as controlled state and clear it through `onClear`; the component invokes cleanup for the footer action, overlay dismissal, Escape, and the Dialog close control.

Copy failures remain feature-owned through `onCopyError` so the surrounding page can use its localized error feedback. Never persist the raw secret after the dialog closes.

## CommercialPlanUpgradeDialog

Use when an action is unavailable because its commercial-plan entitlement is disabled. Pass localized title, description, cancel, and upgrade labels; the component standardizes the dialog composition and links to `/billing/plan` by default.

Keep the triggering action visible while entitlements load, but disable it until the entitlement response is resolved. When the feature is unavailable, open this dialog instead of starting the restricted workflow.

## FormSection

Use `FormSheet` for create/edit forms presented in a side sheet. It standardizes accessible title/description, responsive widths, scroll ownership, header/footer borders, and pending dismissal behavior. Keep form state and submission in the owning feature.

Use to group related fields with a title, description, content, and optional footer actions.

The title is optional for compact sections whose context is already established by the surrounding page or drawer header. Omitting both title and description removes the header column entirely.

Use `actions` for section-level controls that belong on the same row as the section title. Use `footer` for commit or navigation actions after the section content.

For page-level forms, prefer `FormSection` inside the page container. For compact settings cards, use it only when the section title adds useful structure.

## Field Family

Use `FieldGroup` to organize related controls and `Field` for each label/control/helper/error unit. Use `FieldLabel`, `FieldDescription`, and `FieldError` instead of hand-composed label and paragraph markup.

For validation, set `data-invalid` on `Field` and `aria-invalid` on the control. For disabled fields, set `data-disabled` on `Field` and `disabled` on the control.

Product wrappers such as `IdentifierField` and `AppSelectField` already compose the Field family; place them directly inside a `FieldGroup` rather than wrapping them in another `Field`.

Use `AppTextField` and `AppTextareaField` for ordinary labeled text controls. They standardize generated IDs, disabled/invalid semantics, description, and error rendering. Keep specialized inputs such as identifiers, OTP, currency, and search on their dedicated components.

## Permission Editing

Use `PermissionEditor` for controlled permission selection with search, selected/all views, grouped checkboxes, and group-level select/clear actions. Keep the permission catalog and selected keys in the feature; pass localized labels and formatting callbacks into the component.

Use `PermissionSummary` beside the editor when a role workflow needs selected-permission and covered-group counts. It may expose a clear-all action, but submission and validation remain responsibilities of the feature form.

Permission groups use semantic `FieldSet`/`FieldLegend`, and each checkbox uses the Field family. Technical permission keys remain LTR while group and description direction follows the page locale.
