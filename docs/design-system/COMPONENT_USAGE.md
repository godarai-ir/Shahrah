# Component Usage

## GodarAILogo

Use `GodarAILogo` for every tenant-admin brand mark. The default `auto` variant
tracks the application theme; use explicit `light` or `dark` variants only when
the containing surface has a fixed tone, such as the always-dark sidebar.

Keep the intrinsic logo ratio by setting a responsive height with `w-auto
max-w-full object-contain`. Use `decorative` when nearby visible text already
names GodarAI; otherwise pass the localized logo alt text. Email, favicon, PWA,
and social-preview contexts use their dedicated files under `/brand` instead of
this component.

## ShellPageHeading / AdminShell page metadata

Every route-level page rendered inside `AdminShell` must pass a `pageHeader` object with a localized `title`, an optional localized `description`, and an optional contextual `helpHref`. The shell renders `ShellPageHeading` as the page's only `h1` and uses `helpHref` for the existing documentation action, falling back to the documentation home when no contextual link is provided.

Keep route metadata explicit at the server page boundary. Do not derive descriptions from navigation labels or duplicate translated strings in a client route map. `AdminShell` forwards this server-owned metadata through the shell's internal page bridge so the shared route layout can remain mounted across navigation; feature components must not register headings themselves. Long titles and descriptions truncate in the sticky shell; the complete description remains available through its title text.

Primary page actions do not belong in the global shell action group. Place them in the nearest list toolbar or page workflow. Lightweight identity badges or legacy `PageHeader` actions may remain as a compact content action row during migration.

Pages outside `AdminShell` may continue to use `PageHeader` as their route heading. Embedded sheets, dialogs, and entity headers keep their own accessible titles and must not publish route metadata.

## PageHeader

Use at the top of standalone route-level pages that are not rendered inside `AdminShell`. It supports title, description, eyebrow, title-level secondary content through `titleAddon`, and page actions. When rendered under an `AdminShell` route with `pageHeader` metadata, its duplicated title and description are suppressed while remaining actions stay available as a compact content row.

Use `titleAddon` for lightweight contextual links that must remain on the title row. Keep primary actions in `actions`.

## CatalogBreadcrumb

Use `CatalogBreadcrumb` above public catalog provider and model detail surfaces.
The provider page renders `providers → current provider`; the model page renders
`providers → provider → current model`. Intermediate levels remain links and
the final level uses `BreadcrumbPage` so assistive technology receives
`aria-current="page"`. Keep labels localized and preserve automatic direction
for provider and model names.

## Breadcrumb

Use the shared `Breadcrumb` primitives for compact hierarchy on nested detail
pages. Every ancestor must be a real link, while the current entity uses
`BreadcrumbPage`. Do not reproduce breadcrumbs with plain text and separator
characters; the shared separator preserves RTL direction and accessible
navigation semantics.

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

Choose feedback by its UX lifetime and scope:

- Use `showFeedbackToast` for temporary, non-blocking operation results that do not require a decision. This includes successful create/update/delete/copy actions and failures whose originating surface has already closed or is otherwise unavailable. Import it only from `@/components/app-ui/feedback-toast`; feature code must not import `sonner` directly.
- Use `OperationNotice` inside an open `FormSheet`, `FormDialog`, `Sheet`, or confirmation surface when the user must read and act on the result without losing their current context. Validation and retryable submit failures belong beside the form or action that caused them.
- Use raw `Alert` for persistent explanatory, warning, degraded-state, or load-failure callouts that remain relevant until the underlying condition changes. A route-level load failure may stay near the affected content and should include an adjacent retry action when one is available.
- Use `ConfirmDialog` for destructive or high-impact actions that require an explicit user decision. It is backed by `AlertDialog`; rejected confirmation requests stay open and render their `feedback` inside the dialog.

Do not mirror one operation result simultaneously in a toast and an inline alert. Clear stale contextual feedback when its sheet or dialog closes or switches resources.

```tsx
showFeedbackToast({
  kind: "success",
  title: t("messages.saved"),
});
```

## AppCard / SettingsCard

Use `AppCard` for product cards and `SettingsCard` for settings pages or grouped account configuration.

Use `SettingsAuditMeta` for compact “updated at / updated by” metadata below settings controls. It composes `DateTimeValue`, provides semantic term/value markup, and centralizes fallbacks while the feature resolves the actor display label.

Do not put cards inside cards. If a settings surface already has `SettingsCard`, put form controls directly inside it or use an unframed `FormSection` variant.

## MetricCard

Use for dashboard metrics with a label, value, optional description, optional trend, and optional icon.

Use `density="compact"` for dense analytics grids and `tone` to communicate semantic icon emphasis. Do not create feature-local KPI wrappers only to rename `label`, `description`, or density props.

## ModelCard

Use the shared `ModelCard` for catalog and provider-account model grids. Keep the model identity, category badges, selection state, metrics, capability panels, and utility actions in the established hierarchy; do not rebuild the card inside individual model workflows.

Pass the full model-aware text through `selectedCheckboxLabel` for the Switch accessible name and a concise localized label such as “Enabled” through `selectedControlLabel` for visible UI. Technical actions such as FQN copy and code examples belong in `footerActions`, while the model detail link stays beside the title. Keep model names and technical values LTR without changing the surrounding RTL layout.

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

Use `ActiveFilterChips` below advanced filter controls to summarize applied filters and remove them individually. Pass localized accessible removal labels through `getRemoveLabel`; filter state and key interpretation remain feature-owned.

Use `ChipInput` for compact editable sets such as status codes, tags, or technical keys. Supply a domain parser when values need validation or normalization; the component owns keyboard commit, duplicate prevention, backspace removal, and accessible chip removal.

Use `RemovableChipList` when the values are selected by another control and only need controlled display and removal. `ChipInput` composes this primitive for its own committed values.

Use `DateTimeValue` for visible timestamps so the UI emits a semantic `time` element, localized digits/calendar, a stable invalid-value fallback, and tabular numerals. Use the shared `formatDateTime` utility only when a formatted string is required by an interpolation or non-React API.

Use `RefreshButton` for manual list refresh actions. It is icon-only, exposes the localized action through its accessible label and title, and swaps to a disabled spinner while `isLoading` is true.

Use `DateTimeRangeField` for controlled ISO from/to filters. It owns conversion to browser-local `datetime-local` values, LTR input direction, calendar adornments, Field semantics, and responsive two-column layout. Use `DateTimeField` when only one timestamp is required.

Use `ExpirationDurationField` for relative token expiry values accepted by the API. It keeps technical input LTR and monospace, provides accessible Field semantics, and shows localized examples such as `30m`, `24h`, `7d`, and `30d`. Reuse it for PAT and VAT creation instead of duplicating duration hints.

Use `ModelNameValue` for model identifiers in tables and compact metadata. It keeps the displayed identifier LTR, monospace, left-aligned, single-line, and directly copyable. Tables should scroll horizontally instead of breaking a technical identifier character-by-character on narrow screens. When the UI summarizes multiple models, pass the full comma-separated identifiers through `copyValue` while keeping the shorter summary in `value`.

Use `CodePanel` for titled snippets, raw API responses, and examples that need a language badge, description, copy action, or feature-owned actions. It composes the lightweight syntax-highlighted `CodeBlock`, avoiding an editor runtime for read-only examples while retaining line numbers, selection, and bounded scrolling. Set `titleHidden` when the surrounding context already provides a visible heading. Use `CodeViewer` directly only when a read-only surface genuinely needs CodeMirror search or keyboard editing semantics.

Use `TabbedCodePanel` when equivalent code examples are available in multiple languages or shells. It keeps every example in one `CodePanel`, renders the choices through the shared Tabs primitives, and updates the highlighted code surface and copy action together when the active tab changes. Keep tab labels concise, provide a localized `tabsLabel`, and do not render one card per language.

Use `MarkdownRenderer` for safe presentation of model-generated Markdown. It renders CommonMark and GFM structures through `react-markdown` and `remark-gfm`, keeps raw HTML inert, and delegates fenced code to the shared Prism-based `CodeBlock`. Use it for assistant responses and other read-only Markdown output; keep editable source in a code or text editor.

Use `JsonEditorField` for editable JSON surfaces. It uses the shared CodeMirror 6 editor with JSON parsing and diagnostics, line numbers, bracket matching, LTR editing, optional prompt-variable decoration, and valid-JSON formatting on blur.

Use `JsonViewer` for read-only structured JSON such as request payloads, response payloads, log summaries, and metadata. It provides CodeMirror syntax highlighting, line numbers, folding, search, selection, LTR rendering, and a bounded scroll area. Pass an accessible label for each viewer. Keep `CodeBlock` and `CodePanel` for non-JSON snippets or examples that need a title, language badge, copy action, or feature-owned actions.

Use `TablePagination` for previous/next pagination. Pass localized summary and button labels from the feature namespace, and provide page/page-size values only when they add useful context.

Use the feature-level `GatewayPolicyRulesList` for the gateway limit-rules page shell. It composes notice feedback, `DataListPanel` with a caller-owned `toolbar` and `createAction`, an optional `aside` column (the request simulator) that sits beside the list on wide screens and below it on narrow ones, the create/edit `Sheet`, `ConfirmDialog`, and the plan-upgrade dialog. Pass `filtered` so the empty state drops its create action while a search or filter narrows the list.

Rule rows read as one sentence from `ruleSentence`; do not reintroduce limit/window/dimension columns. Rule forms, row rendering, API calls, and domain types stay in `features/gateway-policies`.

## Dropdowns And Lists

Use `RowActionsMenu` for table and card action menus instead of hand-composing `DropdownMenu` in each feature. Use callback actions for edit, activate/deactivate, revoke, and delete; use `href` for navigation and add `external` only for links that open in a new tab.

Start a new action group with `separatorBefore`. For asynchronous actions, pass `pending` and `pendingLabel`; the item stays disabled until the operation finishes. Set `stopPropagation` when the menu is rendered inside a clickable row or card.

Keep filter, column-selection, checkbox, radio, and other stateful menus on the underlying `DropdownMenu` primitives because they are not row-action menus.

Use `AppSelectField` for static option sets in forms and filter toolbars. It standardizes accessible labels, helper/error text, disabled state, RTL popup direction, `sm`/`default` control sizes, and guarantees every `SelectItem` is rendered inside a `SelectGroup`.

Pass `options` for a flat list or `groups` when options need section labels and separators. Individual options can be disabled. Use `optionTextDir="ltr"` for technical values such as model names while keeping the popup itself RTL.

With React Hook Form, keep the component controlled: pass `field.value`, `field.onChange`, and `field.onBlur` from `Controller`, and pass `fieldState.error?.message` to `error`. For simple forms, `watch` plus `setValue` with `shouldDirty`, `shouldTouch`, and `shouldValidate` is acceptable.

Use `FilterCombobox` for searchable filter dropdowns in table/list toolbars. It standardizes `ComboboxInput`, empty state, RTL text alignment, and option truncation.

Use `FilterField` to label specialized toolbar controls with consistent compact Field semantics. Use `DebouncedTextFilter` for text filters that should not update query state on every keystroke, and `RangeFilterField` for a controlled two-thumb numeric range. Keep domain normalization and query serialization in the owning feature.

Use `SearchPickerDialog` as the shell for large entity pickers that require modal focus and search. It composes `FormDialog` and `SearchField`; grouping, result rendering, selection behavior, and domain-specific empty states remain feature-owned.

Keep `components/ui/dropdown-menu`, `components/ui/select`, and `components/ui/combobox` as primitive escape hatches. Route components should prefer the app-ui wrappers unless they need a behavior the wrapper does not support.

## StatusBadge

Use for common statuses: `active`, `inactive`, `pending`, `success`, `warning`, `error`, and `neutral`.

## EndpointModeBadge

Use `EndpointModeBadge` for a single model API endpoint mode such as `chat`,
`image`, `embedding`, `audio`, `video`, `rerank`, or `moderation`. It keeps the
technical label LTR and pairs every mode with a semantic color and recognizable
icon. Unknown modes intentionally receive the primary fallback treatment so
they remain visually distinct from neutral metadata.

Use `EndpointModeBadgeList` when a model supports multiple endpoint modes. It
deduplicates values, wraps responsively, and supports an optional display
limit. Do not recreate endpoint-mode chips with feature-local `Badge` loops.

## ModalityFlow

Use `ModalityFlow` for compact model input/output modality summaries in tables
and dense metadata surfaces. It keeps inputs on the right and outputs on the
left, separates them with a directional arrow, and exposes every icon through
a keyboard-accessible localized tooltip. Modality icons use semantic Badge
variants for color coding, including a dedicated PDF document mark. Pass
localized input, output, empty, and modality labels from the owning feature;
do not recreate modality icon mappings or merge both sides into one
undifferentiated badge list.

## StatusOptionGroup

Use `StatusOptionGroup` when one setting must be selected from a small set of visually descriptive status options. Pass the controlled `value`, localized group `label`, option metadata, and `onValueChange`; use `columns` only to adjust responsive layout.

The component composes a single-selection `ToggleGroup`, preserves one selected value, and provides roving keyboard focus. Keep persistence and save state in the feature, pass `disabled` while loading or saving, and do not use standalone status-option buttons.

## WizardStepper

Use `WizardStepper` for the step list of every multi-step create or edit flow, such as provider-account setup and the guardrail wizard. Each step has a `title`, an optional one-line `description` for its current value or scope, and a `state` of `complete`, `active`, or `idle`. The component renders localized step numbers, a check for completed steps, `aria-current="step"` on the active step, and an underline that marks progress.

Reachability stays in the feature: pass `disabled` for steps that cannot open yet and `onSelect` for steps that can. The active step never calls `onSelect`. Wrap the list in a labelled `nav` when it is the flow's primary navigation, and choose `columns` to match the step count. Do not build feature-local step indicators.

Pages with a sticky action footer pass `contentLayout="fill"` to `AdminShell`; the content area then stretches to the viewport height, so the page root can use `flex-1` and the footer `mt-auto` to stay at the bottom even when the step content is short.

## OptionCardGroup

Use `OptionCardGroup` when a form field chooses one of two to four options and each option needs a title plus a one-line explanation of its effect, such as a guardrail's action, its enforcement mode, or a verification method. It composes the shared `RadioGroup` with `RadioGroupCard` items, so the group is a `radiogroup`, every card is a `radio` with `aria-checked`, and arrow keys move the selection in the page direction.

Label the group with `aria-labelledby` pointing at the visible section heading or `FieldLegend`, or with `aria-label`. Each card is named by its `title` and described by its `description` and optional `tip`; use `tip` for a short emphasized note about when to prefer that option. Mark options that the current context does not allow as `disabled` instead of hiding them, and explain the lock next to the group heading. Use `density="compact"` for option sets inside a form card. Prefer `StatusOptionGroup` when the choice is a status with a badge, and `AppSelectField` when the options need no explanation.

## Spinner And Async Actions

Use `Spinner` for indeterminate loading. Inside a button, compose it with `data-icon="inline-start"`, keep the localized pending label visible, disable the button, and set `aria-busy` on the button. Do not add `loading` or `isPending` props to `Button`.

For standalone loading feedback, pass a localized `label` to `Spinner`; this exposes `role="status"`. When an existing parent already owns the status message, omit the label so the decorative spinner remains hidden from assistive technology.

## ConfirmDialog

Use for destructive or high-impact confirmations. It composes the shared shadcn `AlertDialog` primitives.

Use `open` and `onOpenChange` for feature-controlled dialogs, or pass `trigger` for the standard trigger-based flow. `onConfirm` may return a promise; provide `pendingLabel` so the dialog disables dismissal and communicates progress while the operation runs.

The dialog closes automatically after a successful confirmation and stays open when the promise rejects. Use `onConfirmError` to surface rejected operations. Set `closeOnConfirm={false}` only when the feature itself closes the dialog after updating its domain state.

## TypedConfirmDialog

Use `TypedConfirmDialog` only for high-impact destructive actions that require typing an exact technical identifier such as a team or virtual-user alias. Keep ordinary confirmations on `ConfirmDialog`.

Pass the localized input instruction separately from `expectedValue`. Confirmation remains disabled until the trimmed value matches exactly. Async confirmation disables dismissal, shows the shared `Spinner`, closes after success, and stays open when the promise rejects. Technical confirmation values default to LTR.

Pass dialog-scoped operation feedback through `feedback`, normally as an
`OperationNotice`. If confirmation fails while the dialog stays open, render
the failure there instead of relying on a page-level notice behind the modal.
Clear stale feedback when the dialog closes or switches to another resource.

Use `UnsavedChangesDialog` when leaving, replacing, or closing an editable surface would discard dirty form state. Keep the pending continuation in the owning feature and run it through `onDiscard`; the dialog provides explicit “keep editing” and “discard changes” actions instead of relying on `window.confirm`.

## Copyable Values And One-Time Secrets

Use `CopyPageLink` for the compact link-icon action beside public catalog
provider and model titles. It copies the current browser URL, exposes localized
idle, success, and failure labels, announces the result, and includes a legacy
clipboard fallback. Do not implement page-link copy state separately in each
detail header.

Use `TokenScale` wherever catalog pricing needs to identify whether an amount is
for one token or one million tokens. It renders the localized compact notation
and exposes the unabbreviated scale in a keyboard-accessible tooltip. Compose it
through `DisplayPrice` when both the localized currency amount and token scale
are shown together; do not rebuild this phrase in individual catalog pages.

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

Compose `FormSheetActions` inside the `footer` slot for Drawer actions. Pass the cancel or back action as `secondary` and the commit action as `primary`; this produces two equal-width actions with stable RTL ordering. Use `auxiliary` for neutral actions such as reset and `destructive` for delete. Form submit buttons may target a form in the scrollable body through the native `form` attribute. Do not place submit buttons inside the scrolling body or recreate sticky action cards.

Use `FormDialog` for focused configuration forms that belong in a modal. It standardizes accessible heading structure, responsive sizes, scroll containment, optional footer actions, and pending-state dismissal protection. Keep field state and submission in the feature.

Use to group related fields with a title, description, content, and optional footer actions.

The title is optional for compact sections whose context is already established by the surrounding page or drawer header. Omitting both title and description removes the header column entirely.

Use `actions` for section-level controls that belong on the same row as the section title. Use `footer` for commit or navigation actions after the section content.

For page-level forms, prefer `FormSection` inside the page container. For compact settings cards, use it only when the section title adds useful structure.

## Field Family

Use `FieldGroup` to organize related controls and `Field` for each label/control/helper/error unit. Use `FieldLabel`, `FieldDescription`, and `FieldError` instead of hand-composed label and paragraph markup.

For validation, set `data-invalid` on `Field` and `aria-invalid` on the control. For disabled fields, set `data-disabled` on `Field` and `disabled` on the control.

Product wrappers such as `IdentifierField` and `AppSelectField` already compose the Field family; place them directly inside a `FieldGroup` rather than wrapping them in another `Field`.

Use `AppTextField` and `AppTextareaField` for ordinary labeled text controls. They standardize generated IDs, disabled/invalid semantics, description, and error rendering. Keep specialized inputs such as identifiers, OTP, currency, and search on their dedicated components.

Use `DisplayNameField` whenever a form collects a user-facing display name. It derives the input direction from the first meaningful Unicode letter: Arabic-family and other RTL scripts select RTL, while LTR scripts select LTR. Leading whitespace, punctuation, and digits do not lock the direction. Do not set `dir` manually on display-name inputs.

Use `AppEmailField` for email addresses. It always renders an LTR email control with email input mode, autocomplete, spellcheck disabled, and accessible client-side validation on blur or invalid form submission. Pass localized `invalidMessage` and, for required fields, `requiredMessage`; do not recreate email direction or validation in feature forms.

Use `AppSwitchField` for boolean settings that need a label, optional explanation, validation, and consistent horizontal alignment. Its default surface is suitable for settings rows; pass `surface={false}` when the surrounding card or form section already owns the visual container.

Use `AppCodeTextareaField` for editable JSON, schemas, regular expressions, JWKS documents, and other technical multi-line values. It preserves the Field-family semantics while defaulting the control to LTR, monospace text, disabled spellcheck, and a readable code height.

Use `AppNumberField` for numeric configuration values. It composes `AppTextField` while standardizing number input type, LTR direction, tabular numerals, input mode, and validation rendering; pass domain-specific min, max, and step constraints from the feature.

## Permission Editing

Use `PermissionEditor` for controlled permission selection with search, selected/all views, grouped checkboxes, and group-level select/clear actions. Keep the permission catalog and selected keys in the feature; pass localized labels and formatting callbacks into the component.

Use `CheckboxOptionList` for compact, scrollable multi-selection lists such as allowed models or attached guardrails. It composes `FieldSet`, `FieldGroup`, and `Checkbox`, and owns loading and empty states while selection state remains feature-controlled.

Use `PermissionSummary` beside the editor when a role workflow needs selected-permission and covered-group counts. It may expose a clear-all action, but submission and validation remain responsibilities of the feature form.

Permission groups use semantic `FieldSet`/`FieldLegend`, and each checkbox uses the Field family. Technical permission keys remain LTR while group and description direction follows the page locale.

## Playground Composition

Use `ResizableSplitPane` for two sibling work surfaces that need a direction-aware desktop resize handle. Pass logical `start` and `end` content, the page direction, and localized accessible labels. The separator supports pointer and keyboard resizing; use a tabbed composition at narrow breakpoints rather than shrinking both panes below a readable width.

Use `PromptMessageEditor` for compact prompt-message rows. The owning feature controls the message role, content, tool-call identifier, ordering, duplication, removal, validation, and focus restoration. Always pass localized labels for every icon action. Technical tool-call identifiers remain LTR.

Use `RunTelemetryStrip` for an aligned set of per-run metrics such as token usage, latency, cost, effective model, cache state, and trace ID. Pass `null` for unavailable values instead of converting missing data to zero, and mark identifiers as technical. The component may expose a copy action when `rawValue` is provided.
