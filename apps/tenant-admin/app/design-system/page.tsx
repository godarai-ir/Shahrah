"use client";

import { useState } from "react";
import { Activity, AlertTriangle, CheckCircle2, Database, Flag, Inbox, KeyRound, OctagonX, Shield, Trash2, Upload, X } from "lucide-react";

import { ActiveFilterChips } from "@/components/app-ui/ActiveFilterChips";
import { AppCard } from "@/components/app-ui/AppCard";
import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppEmailField } from "@/components/app-ui/AppEmailField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { CheckboxOptionList } from "@/components/app-ui/CheckboxOptionList";
import { ChipInput } from "@/components/app-ui/ChipInput";
import { CodePanel } from "@/components/app-ui/CodePanel";
import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DataListPanel } from "@/components/app-ui/DataListPanel";
import { DateTimeRangeField } from "@/components/app-ui/DateTimeRangeField";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { EntityIdentityHeader } from "@/components/app-ui/EntityIdentityHeader";
import { ExpirationDurationField } from "@/components/app-ui/ExpirationDurationField";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { DebouncedTextFilter, RangeFilterField } from "@/components/app-ui/FilterField";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { FormSection } from "@/components/app-ui/FormSection";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { GodarAILogo } from "@/components/app-ui/GodarAILogo";
import { HelpLink } from "@/components/app-ui/HelpLink";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { JsonEditorField } from "@/components/app-ui/JsonEditor";
import { JsonViewer } from "@/components/app-ui/JsonViewer";
import { MarkdownRenderer } from "@/components/app-ui/MarkdownRenderer";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { ModalityFlow } from "@/components/app-ui/ModalityFlow";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { PermissionEditor } from "@/components/app-ui/PermissionEditor";
import { PermissionSummary } from "@/components/app-ui/PermissionSummary";
import { PromptMessageEditor } from "@/components/app-ui/PromptMessageEditor";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { RemovableChipList } from "@/components/app-ui/RemovableChipList";
import { ResizableSplitPane } from "@/components/app-ui/ResizableSplitPane";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { RunTelemetryStrip } from "@/components/app-ui/RunTelemetryStrip";
import { SearchField } from "@/components/app-ui/SearchField";
import { SectionHeader } from "@/components/app-ui/SectionHeader";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { ShellPageHeading } from "@/components/app-ui/ShellPageHeading";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { StatusOptionGroup } from "@/components/app-ui/StatusOptionGroup";
import { TabbedCodePanel } from "@/components/app-ui/TabbedCodePanel";
import { TablePagination } from "@/components/app-ui/TablePagination";
import { TableEmptyRow, TableLoadingRows } from "@/components/app-ui/TableStateRows";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { UnsavedChangesDialog } from "@/components/app-ui/UnsavedChangesDialog";
import { WizardStepper } from "@/components/app-ui/WizardStepper";
import { SearchPickerDialog } from "@/components/app-ui/SearchPickerDialog";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { SettingsAuditMeta } from "@/components/app-ui/SettingsAuditMeta";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Attachment, AttachmentAction, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Message, MessageContent, MessageHeader } from "@/components/ui/message";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { GatewayPolicyRulesList } from "@/features/gateway-policies/components";
import { ModelCardExampleGrid } from "@/components/app-ui/ModelCard.example";
import { formatCompactCurrency, formatCompactIrr, formatCompactNumber } from "@/lib/compact-number";

const tokens = [
  ["background", "bg-background text-foreground"],
  ["card", "bg-card text-card-foreground"],
  ["primary", "bg-primary text-primary-foreground"],
  ["primary-soft", "bg-primary-soft text-primary-soft-foreground"],
  ["secondary", "bg-secondary text-secondary-foreground"],
  ["muted", "bg-muted text-muted-foreground"],
  ["accent", "bg-accent text-accent-foreground"],
  ["success-soft", "bg-success-soft text-success-soft-foreground"],
  ["destructive", "bg-destructive text-destructive-foreground"],
];

const permissionOptions = [
  { key: "role:read", group: "roles", description: "Read role details" },
  { key: "role:manage", group: "roles", description: "Manage roles" },
  { key: "user:read", group: "users", description: "Read users" },
];

export default function DesignSystemPage() {
  const [selectValue, setSelectValue] = useState("active");
  const [filterValue, setFilterValue] = useState("openai");
  const [textFilterValue, setTextFilterValue] = useState("");
  const [rangeFilterValue, setRangeFilterValue] = useState<[number, number]>([20, 80]);
  const [samplePage, setSamplePage] = useState(1);
  const [permissionQuery, setPermissionQuery] = useState("");
  const [permissionMode, setPermissionMode] = useState<"selected" | "all">("all");
  const [selectedPermissions, setSelectedPermissions] = useState(["role:read"]);
  const [policyQuery, setPolicyQuery] = useState("");
  const [policySheetOpen, setPolicySheetOpen] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [demoSecret, setDemoSecret] = useState<string | null>(null);
  const [statusOption, setStatusOption] = useState<"metadata" | "full">("metadata");
  const [optionCard, setOptionCard] = useState<"block" | "flag" | "rewrite">("block");
  const [wizardStep, setWizardStep] = useState(1);
  const [typedConfirmOpen, setTypedConfirmOpen] = useState(false);
  const [typedConfirmNotice, setTypedConfirmNotice] = useState<OperationNoticeValue | null>(null);
  const [unsavedChangesOpen, setUnsavedChangesOpen] = useState(false);
  const [formSheetOpen, setFormSheetOpen] = useState(false);
  const [formDialogOpen, setFormDialogOpen] = useState(false);
  const [searchPickerOpen, setSearchPickerOpen] = useState(false);
  const [searchPickerQuery, setSearchPickerQuery] = useState("");
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const [sampleMessage, setSampleMessage] = useState("برای {{topic}} یک توضیح کوتاه بنویس.");
  const [sampleJson, setSampleJson] = useState('{"model":"gpt-5-nano","messages":[{"role":"user","content":"برای {{topic}} توضیح بنویس"}]}');
  const requestsDisplay = formatCompactNumber(128_400, { locale: "en-US" });
  const spendDisplay = formatCompactCurrency(12_340, { locale: "en-US", currency: "USD" });
  const irrDisplay = formatCompactIrr(845_000, { locale: "fa-IR" });

  return (
    <PageContainer as="main" spacing="relaxed" dir="ltr">
      <PageHeader
        eyebrow="Design System"
        title="Tenant Admin UI Foundation"
        description="Internal playground for product-level UI components, semantic tokens, and RTL-friendly layout patterns."
        titleAddon={<HelpLink href="/design-system">Title-level help</HelpLink>}
        actions={<Button>Primary action</Button>}
      />

      <AppCard title="Breadcrumb" description="Linked hierarchy for nested detail pages, with an accessible current location.">
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink href="/prompts">Prompts</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>booking_greeting</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </AppCard>

      <section className="flex flex-col gap-4">
        <SectionHeader
          title="GodarAI visual identity"
          description="Neutral application surfaces, indigo actions, restrained cyan data colors, and semantic operational states."
        />
        <div className="theme-canvas overflow-hidden rounded-lg border border-card-border">
          <div className="grid min-h-64 md:grid-cols-3">
            <aside className="theme-sidebar flex flex-col gap-4 border-e border-sidebar-border bg-sidebar p-4 text-sidebar-foreground">
              <div className="flex items-center gap-3">
                <span className="flex size-9 items-center justify-center rounded-lg bg-sidebar-accent p-1 ring-1 ring-sidebar-border">
                  <GodarAILogo decorative className="h-full w-auto max-w-full object-contain" />
                </span>
                <span className="flex flex-col gap-0.5">
                  <strong className="text-sm">GodarAI</strong>
                  <span className="text-xs text-sidebar-foreground/60">Calm developer console</span>
                </span>
              </div>
              <div className="border-s-2 border-sidebar-primary bg-sidebar-accent px-3 py-2 text-sm font-semibold text-sidebar-accent-foreground">
                Active navigation
              </div>
            </aside>
            <div className="grid content-start gap-4 p-4 sm:grid-cols-2 md:col-span-2">
              <MetricCard label="Requests" value="128.4K" description="Current period" icon={Activity} />
              <MetricCard label="Reliability" value="99.2%" description="Operational state" icon={CheckCircle2} tone="success" />
              <AppCard className="sm:col-span-2" title="Neutral data surface" description="Thin borders, restrained elevation, and compact hierarchy.">
                <div className="h-16 rounded-md border border-primary-subtle-border bg-primary-soft" />
              </AppCard>
            </div>
          </div>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <AppCard title="Automatic" description="Follows the active application theme.">
            <div className="flex min-h-28 items-center justify-center rounded-lg border border-card-border bg-card p-4">
              <GodarAILogo className="h-16 w-auto max-w-full object-contain" />
            </div>
          </AppCard>
          <AppCard title="Light surface" description="Navy and cyan mark for light backgrounds.">
            <div className="flex min-h-28 items-center justify-center rounded-lg border border-card-border bg-card p-4">
              <GodarAILogo variant="light" className="h-16 w-auto max-w-full object-contain" />
            </div>
          </AppCard>
          <AppCard title="Dark surface" description="Ice and cyan mark for dark backgrounds.">
            <div className="flex min-h-28 items-center justify-center rounded-lg border border-sidebar-border bg-sidebar p-4">
              <GodarAILogo variant="dark" className="h-16 w-auto max-w-full object-contain" />
            </div>
          </AppCard>
          <AppCard title="Monochrome" description="Single-color marks for constrained surfaces.">
            <div className="grid min-h-28 grid-cols-2 overflow-hidden rounded-lg border border-card-border">
              <span className="flex items-center justify-center bg-card p-3">
                <GodarAILogo variant="mono-navy" decorative className="h-12 w-auto max-w-full object-contain" />
              </span>
              <span className="flex items-center justify-center bg-sidebar p-3">
                <GodarAILogo variant="mono-white" decorative className="h-12 w-auto max-w-full object-contain" />
              </span>
            </div>
          </AppCard>
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader
          title="Shell page heading"
          description="Route-level pages pass their localized title, optional description, and contextual help link to AdminShell."
        />
        <div className="grid min-w-0 gap-3 md:grid-cols-2" dir="rtl">
          <div className="min-w-0 rounded-xl border border-border bg-background p-4">
            <ShellPageHeading
              as="div"
              title="قوانین محدودسازی نرخ"
              description="تعریف قوانین محدودسازی برای درخواست‌های گیت‌وی بر اساس کاربر، مدل و متادیتا."
            />
          </div>
          <div className="min-w-0 rounded-xl border border-border bg-background p-4">
            <ShellPageHeading as="div" title="پلن" />
          </div>
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Color tokens" description="Use semantic token classes instead of raw palette utilities." />
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {tokens.map(([name, className]) => (
            <div key={name} className={`rounded-lg border border-border p-4 ${className}`}>
              <p className="text-sm font-semibold">{name}</p>
              <p className="mt-1 text-xs opacity-80">{className}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-3">
        <MetricCard label="Requests" value={<CompactValue display={requestsDisplay} />} description="Current billing period" trend={<StatusBadge status="success">+12%</StatusBadge>} icon={Activity} />
        <MetricCard label="Provider accounts" value="8" description="Active connections" icon={Database} />
        <MetricCard label="Monthly spend" value={<CompactValue display={spendDisplay} />} description="Exact value stays in tooltip" icon={KeyRound} />
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader
          title="Model information card"
          description="RTL-first model identity, selection, pricing, modality, capability, and invocation-mode pattern."
        />
        <ModelCardExampleGrid />
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(16rem,1fr)]">
        <AppCard title="Permission editor" description="Controlled search, grouped selection, and selected/all views for role permissions.">
          <PermissionEditor
            permissions={permissionOptions}
            value={selectedPermissions}
            onValueChange={setSelectedPermissions}
            query={permissionQuery}
            onQueryChange={setPermissionQuery}
            mode={permissionMode}
            onModeChange={setPermissionMode}
            dir="ltr"
            labels={{
              selectedTab: "Selected",
              allTab: "All permissions",
              searchPlaceholder: "Search permission key or description",
              searchAriaLabel: "Search permissions",
              emptyTitle: "No matching permissions found",
              selectGroup: "Select group",
              clearGroup: "Clear group",
            }}
          />
        </AppCard>
        <PermissionSummary
          permissions={permissionOptions}
          value={selectedPermissions}
          title="Permission summary"
          labels={{
            selectedPermissions: "Selected permissions",
            coveredGroups: "Groups covered",
            clearAll: "Clear all permissions",
            empty: "No permissions are selected.",
          }}
          onClear={() => setSelectedPermissions([])}
        />
      </section>

      <section className="rounded-xl border border-border p-4">
        <GatewayPolicyRulesList
          title="Gateway policy rules"
          createAction={<Button type="button" onClick={() => setPolicySheetOpen(true)}>New rule</Button>}
          alertLabels={{ error: "Error", success: "Success" }}
          toolbar={<SearchField placeholder="Search rules" value={policyQuery} onChange={(event) => setPolicyQuery(event.target.value)} />}
          filtered={Boolean(policyQuery)}
          aside={<div className="rounded-xl border border-border p-4 text-sm text-muted-foreground">Feature-owned side panel</div>}
          itemCount={1}
          loading={false}
          loadingLabel="Loading rules"
          emptyIcon={Shield}
          emptyTitle="No rules"
          emptyDescription="Create the first policy rule."
          className="max-w-none px-0 py-0"
          ruleSheet={{
            open: policySheetOpen,
            onOpenChange: setPolicySheetOpen,
            title: "Create policy rule",
            description: "Configure the domain-specific policy form.",
            closeLabel: "Close",
            dir: "ltr",
            content: <div className="p-6 text-sm text-muted-foreground">Feature-owned policy form</div>,
          }}
          deleteDialog={{
            open: false,
            onOpenChange: () => undefined,
            title: "Delete policy rule",
            description: "This action cannot be undone.",
            cancelLabel: "Cancel",
            confirmLabel: "Delete",
            pendingLabel: "Deleting",
            pending: false,
            onConfirm: () => undefined,
          }}
          upgradeDialog={{
            open: false,
            onOpenChange: () => undefined,
            title: "Upgrade plan",
            description: "This feature requires a higher plan.",
            cancelLabel: "Cancel",
            actionLabel: "Upgrade",
          }}
        >
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Rule ID</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell className="font-mono" dir="ltr">tenant-default</TableCell>
                <TableCell><StatusBadge status="active">Active</StatusBadge></TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </GatewayPolicyRulesList>
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <AppCard title="Definition list" description="Use semantic label/value metadata for detail surfaces.">
          <DefinitionList columns={2}>
            <DefinitionListItem label="Provider" value="OpenAI" icon={Database} />
            <DefinitionListItem label="Model" value="gpt-5.4" mono valueDir="ltr" />
            <DefinitionListItem label="Request ID" value="req_demo_123456789" mono valueDir="ltr" wrap />
            <DefinitionListItem label="Status" value={<StatusBadge status="active">Active</StatusBadge>} />
          </DefinitionList>
        </AppCard>
        <AppCard title="Typed confirmation" description="Reserve typed confirmation for high-impact destructive actions.">
          <Button type="button" variant="destructive" onClick={() => setTypedConfirmOpen(true)}>Delete demo team</Button>
          <TypedConfirmDialog
            open={typedConfirmOpen}
            onOpenChange={(open) => {
              setTypedConfirmOpen(open);
              if (!open) setTypedConfirmNotice(null);
            }}
            title="Delete demo team"
            description="This action cannot be undone."
            expectedValue="demo-team"
            inputLabel="Type demo-team to confirm"
            cancelLabel="Cancel"
            confirmLabel="Delete"
            pendingLabel="Deleting"
            feedback={typedConfirmNotice ? (
              <OperationNotice
                notice={typedConfirmNotice}
                labels={{ success: "Success", error: "Error", warning: "Warning", info: "Information" }}
              />
            ) : null}
            onConfirm={() => Promise.reject(new Error("Demo dependency conflict"))}
            onConfirmError={() => setTypedConfirmNotice({ kind: "error", message: "The resource is still referenced. Resolve its dependencies and try again." })}
          />
        </AppCard>
        <AppCard title="Form sheet" description="Standard shell for create and edit workflows.">
          <Button type="button" onClick={() => setFormSheetOpen(true)}>Open form sheet</Button>
          <FormSheet
            open={formSheetOpen}
            onOpenChange={setFormSheetOpen}
            title="Create resource"
            description="Enter the resource details."
            footer={(
              <FormSheetActions
                secondary={<Button type="button" variant="outline" onClick={() => setFormSheetOpen(false)}>Cancel</Button>}
                primary={<Button type="button" onClick={() => setFormSheetOpen(false)}>Save</Button>}
              />
            )}
          >
            <FormSection title="General information"><Input placeholder="Resource name" /></FormSection>
          </FormSheet>
          <Button type="button" variant="outline" onClick={() => setFormDialogOpen(true)}>Open form dialog</Button>
          <FormDialog open={formDialogOpen} onOpenChange={setFormDialogOpen} title="Configure resource" description="Update the shared configuration." footer={<Button onClick={() => setFormDialogOpen(false)}>Save</Button>}><AppTextField label="Name" /></FormDialog>
          <Button type="button" variant="outline" onClick={() => setSearchPickerOpen(true)}>Open search picker</Button>
          <SearchPickerDialog open={searchPickerOpen} onOpenChange={setSearchPickerOpen} title="Choose model" searchValue={searchPickerQuery} onSearchValueChange={setSearchPickerQuery} searchPlaceholder="Search models"><Button variant="ghost">gpt-4.1</Button></SearchPickerDialog>
        </AppCard>
        <AppCard title="Status option group" description="Use for a single setting selected from visual status cards.">
          <StatusOptionGroup
            label="Capture mode"
            value={statusOption}
            onValueChange={setStatusOption}
            options={[
              {
                value: "metadata",
                badge: "metadata-only",
                title: "Metadata only",
                description: "Store request metadata without content.",
                icon: Database,
                tone: "info",
              },
              {
                value: "full",
                badge: "full-content",
                title: "Full content",
                description: "Store complete request and response content.",
                icon: Shield,
                tone: "success",
              },
            ]}
          />
        </AppCard>
        <AppCard title="Wizard stepper" description="Use for the step list of a multi-step create flow. Completed steps can be revisited; later steps stay disabled until reachable.">
          <WizardStepper
            aria-label="Create resource"
            steps={["Type", "Settings", "Behavior", "Review"].map((title, index) => ({
              id: title,
              title,
              description: index < wizardStep ? "Completed" : index === wizardStep ? "In progress" : undefined,
              state: index < wizardStep ? "complete" : index === wizardStep ? "active" : "idle",
              disabled: index > wizardStep + 1,
              onSelect: () => setWizardStep(index),
            }))}
          />
        </AppCard>
        <AppCard title="Option card group" description="Use for one choice among a few options that each need a title and a one-line effect. Cards are radios; arrow keys move the selection.">
          <OptionCardGroup
            aria-label="Enforcement"
            columns={3}
            value={optionCard}
            onValueChange={setOptionCard}
            options={[
              { value: "block", title: "Block and log", description: "The request is stopped and logged.", icon: OctagonX },
              { value: "flag", title: "Allow and flag", description: "The request passes and is flagged.", tip: "Useful before enforcing.", icon: Flag },
              { value: "rewrite", title: "Rewrite", description: "Unavailable for this rule type.", icon: Shield, disabled: true },
            ]}
          />
        </AppCard>
        <AppCard title="Copyable value" description="Use for technical identifiers and optionally masked values.">
          <CopyableValue
            value="sk_demo_123456789"
            masked
            allowReveal
            labels={{
              valueAriaLabel: "Demo secret",
              copy: "Copy",
              copying: "Copying",
              copied: "Copied",
              reveal: "Reveal",
              hide: "Hide",
            }}
          />
          <div className="flex items-center gap-2">
            <code dir="ltr" className="text-xs text-muted-foreground">provider:model</code>
            <CopyAction value="provider:model" copyLabel="Copy identifier" copiedLabel="Identifier copied" />
          </div>
        </AppCard>
        <AppCard title="One-time secret" description="Clears sensitive state whenever the dialog closes.">
          <Button
            type="button"
            onClick={() => {
              setDemoSecret("vat_demo_123456789");
              setSecretDialogOpen(true);
            }}
          >
            Show one-time secret
          </Button>
          <OneTimeSecretDialog
            open={secretDialogOpen}
            onOpenChange={setSecretDialogOpen}
            value={demoSecret}
            title="Copy secret now"
            description="This raw value is shown only once."
            warning="Store it securely before closing this dialog."
            closeLabel="Close"
            copyLabels={{
              valueAriaLabel: "One-time secret",
              copy: "Copy secret",
              copying: "Copying",
              copied: "Copied",
            }}
            onClear={() => setDemoSecret(null)}
          />
        </AppCard>
        <AppCard title="Refresh button" description="Icon-only manual refresh action for operational lists.">
          <RefreshButton label="Refresh data" />
        </AppCard>
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <AppCard title="Compact values" description="Use CompactValue when large numbers are easier to scan in abbreviated form.">
          <div className="grid gap-3 sm:grid-cols-3">
            <div className="rounded-lg border border-border/70 p-3">
              <div className="text-xs text-muted-foreground">Requests</div>
              <div className="mt-1 text-lg font-semibold"><CompactValue display={requestsDisplay} /></div>
            </div>
            <div className="rounded-lg border border-border/70 p-3">
              <div className="text-xs text-muted-foreground">USD cost</div>
              <div className="mt-1 text-lg font-semibold"><CompactValue display={spendDisplay} /></div>
            </div>
            <div className="rounded-lg border border-border/70 p-3">
              <div className="text-xs text-muted-foreground">IRR amount</div>
              <div className="mt-1 text-lg font-semibold"><CompactValue display={irrDisplay} /></div>
            </div>
          </div>
        </AppCard>
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <AppCard title="Standard card" description="Use AppCard for repeated product surfaces." actions={<Button variant="outline" size="sm">Action</Button>}>
          <p className="text-sm leading-6 text-muted-foreground">
            Cards should keep content compact, readable, and token-based.
          </p>
        </AppCard>

        <SettingsCard title="Settings card" description="Use for grouped account or tenant configuration." footer={<Button size="sm">Save</Button>}>
          <Input placeholder="Workspace name" />
          <Input placeholder="Support email" />
        </SettingsCard>
      </section>

      <FormSection
        title="Form section"
        description="Group related fields with a stable title, helper copy, content, and actions."
        actions={<Button size="sm" variant="outline">Add field</Button>}
        footer={<><Button variant="outline">Cancel</Button><Button>Submit</Button></>}
      >
        <Input placeholder="API base URL" dir="ltr" />
        <Input placeholder="Token alias" />
      </FormSection>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Field family" description="Compose labels, controls, helper text, and validation with FieldGroup and Field." />
        <FieldGroup className="grid gap-4 md:grid-cols-2">
          <DisplayNameField label="Display name" placeholder="Production workspace" description="Direction follows the first meaningful RTL or LTR letter." />
          <Field data-invalid>
            <FieldLabel htmlFor="playground-field-slug">Technical identifier</FieldLabel>
            <Input id="playground-field-slug" value="invalid identifier" readOnly aria-invalid />
            <FieldError>Use English letters, numbers, dashes, or underscores.</FieldError>
          </Field>
        </FieldGroup>
        <AppCard title="Text fields" description="Use the app field wrappers for consistent labels, helper text, and validation state.">
          <FieldGroup className="grid gap-4 md:grid-cols-2">
            <AppTextField label="Resource title" placeholder="Production resource" description="Shown throughout the tenant admin." />
            <AppEmailField label="Email" placeholder="user@example.com" invalidMessage="Enter a valid email address." requiredMessage="Email is required." required />
            <AppTextareaField label="Description" placeholder="Describe this resource" />
            <AppSwitchField label="Enable caching" description="Cache matching gateway requests." checked />
            <AppCodeTextareaField label="JSON schema" defaultValue={'{\n  "type": "object"\n}'} />
            <JsonEditorField label="JSON editor" value={sampleJson} variables={["topic"]} onValueChange={setSampleJson} />
            <AppNumberField label="Timeout" defaultValue={30} min={1} step={1} />
            <ExpirationDurationField defaultValue="30d" />
            <ModelNameValue value="openai:gpt-5-nano" />
          </FieldGroup>
        </AppCard>
        <AppCard title="JSON viewer" description="Use the read-only CodeMirror surface for structured JSON payloads and metadata.">
          <JsonViewer
            aria-label="Read-only JSON example"
            value={{ log_id: "log_123", status_code: 200, metadata: { environment: "production" } }}
            maxHeight="16rem"
          />
        </AppCard>
        <AppCard>
          <EntityIdentityHeader title="Production service" identifier="svc_production" description="Primary gateway service identity." status={<StatusBadge status="active">Active</StatusBadge>} copy={{ copyLabel: "Copy identifier", copiedLabel: "Identifier copied" }} actions={<Button variant="outline" size="sm">Edit</Button>} />
        </AppCard>
        <AppCard title="Date and time range" description="Technical timestamps remain LTR inside RTL pages.">
          <DateTimeRangeField from="2025-03-12T09:30:00Z" to="2025-03-12T10:30:00Z" fromLabel="From" toLabel="To" onFromChange={() => undefined} onToChange={() => undefined} />
        </AppCard>
        <AppCard title="Checkbox option list">
          <CheckboxOptionList label="Allowed models" options={[{ value: "gpt-4.1", label: "gpt-4.1", textDir: "ltr" }, { value: "claude", label: "claude", textDir: "ltr" }]} value={["gpt-4.1"]} onValueChange={() => undefined} emptyText="No models" />
        </AppCard>
        <AppCard title="Active filters"><ActiveFilterChips items={[{ key: "model", label: "Model", value: "gpt-4.1" }]} onRemove={() => undefined} getRemoveLabel={(item) => `Remove ${item.label}`} /></AppCard>
        <AppCard title="Chip input"><div className="flex flex-col gap-4"><ChipInput value={["429", "500"]} onValueChange={() => undefined} getRemoveLabel={(value) => `Remove ${value}`} placeholder="Add status code" /><RemovableChipList items={[{ value: "openai", label: "OpenAI", removeLabel: "Remove OpenAI" }]} onRemove={() => undefined} variant="outline" /></div></AppCard>
        <AppCard title="Settings audit metadata"><SettingsAuditMeta updatedAt="2025-03-12T09:30:00Z" updatedBy="Tenant admin" updatedAtLabel="Updated" updatedByLabel="By" fallback="Not set" locale="en" /></AppCard>
        <CodePanel title="Gateway request" description="A reusable titled code surface with optional actions." displayLanguage="cURL" code={'curl "$GATEWAY_URL/v1/models"'} language="bash" copy={{ copyLabel: "Copy code", copiedLabel: "Code copied" }} />
        <TabbedCodePanel title="Shell examples" tabsLabel="Shell" tabs={[{ value: "bash", label: "Bash", language: "bash", code: "curl https://gateway.example/v1/models" }, { value: "powershell", label: "PowerShell", language: "powershell", code: "Invoke-RestMethod https://gateway.example/v1/models" }]} copy={{ copyLabel: "Copy code", copiedLabel: "Code copied" }} />
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Identifier field" description="Use for user-defined technical identifiers and aliases." />
        <IdentifierField label="شناسه فنی" placeholder="team_support" />
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Search field" description="Use for list and toolbar search. The icon follows RTL inline-end placement." />
        <SearchField containerClassName="max-w-md" placeholder="جست‌وجو در کاربران" />
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Help link" description="Use for page-level and inline links to user guidance." />
        <div>
          <HelpLink href="/docs/quick-start" />
        </div>
        <div className="max-w-sm">
          <Progress value={45} aria-label="Example usage progress" />
        </div>
      </section>

      <section className="flex flex-col gap-4">
        <SectionHeader title="Status badges and states" />
        <AppCard title="Modality flow" description="Icon-only model input and output modalities with accessible tooltips.">
          <ModalityFlow
            inputModalities={["text", "image", "pdf", "video"]}
            outputModalities={["text", "audio"]}
            inputLabel="Input"
            outputLabel="Output"
            emptyLabel="N/A"
            getModalityLabel={(modality) => modality}
          />
        </AppCard>
        <div className="flex flex-wrap gap-2" aria-label="Semantic metadata chips">
          <Badge variant="info">Informational capability</Badge>
          <Badge variant="success">Recommended choice</Badge>
          <Badge variant="warning">Usage consideration</Badge>
          <Badge variant="secondary">Neutral metadata</Badge>
        </div>
        <div className="flex flex-wrap gap-2">
          <StatusBadge status="active">active</StatusBadge>
          <StatusBadge status="inactive">inactive</StatusBadge>
          <StatusBadge status="pending">pending</StatusBadge>
          <StatusBadge status="success">success</StatusBadge>
          <StatusBadge status="warning">warning</StatusBadge>
          <StatusBadge status="error">error</StatusBadge>
          <StatusBadge status="neutral">neutral</StatusBadge>
        </div>
        <OperationNotice
          notice={{ kind: "success", message: "The resource was updated successfully." }}
          labels={{ success: "Success", error: "Error", warning: "Warning", info: "Information" }}
        />
        <AppCard title="Transient feedback" description="Use toast for non-blocking operation results.">
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={() => showFeedbackToast({ kind: "success", title: "Saved successfully" })}>Success toast</Button>
            <Button type="button" variant="outline" onClick={() => showFeedbackToast({ kind: "error", title: "Request failed", description: "Try the operation again." })}>Error toast</Button>
            <Button type="button" variant="outline" onClick={() => showFeedbackToast({ kind: "warning", title: "Review required" })}>Warning toast</Button>
            <Button type="button" variant="outline" onClick={() => showFeedbackToast({ kind: "info", title: "Background work started" })}>Info toast</Button>
          </div>
        </AppCard>
        <div className="grid gap-4 lg:grid-cols-3">
          <EmptyState icon={Inbox} title="No records yet" description="Use the surface variant for standalone empty states." primaryAction={<Button size="sm">Create</Button>} />
          <AppCard title="Loading">
            <div className="flex flex-col gap-3">
              <Button type="button" disabled aria-busy="true" className="w-fit">
                <Spinner data-icon="inline-start" />
                Saving
              </Button>
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
              <Skeleton className="h-20 w-full" />
            </div>
          </AppCard>
          <AppCard title="Error state" actions={<StatusBadge status="error">error</StatusBadge>}>
            <div className="flex items-start gap-3 text-sm text-muted-foreground">
              <AlertTriangle className="mt-0.5 text-destructive" aria-hidden="true" />
              <p>Show concise, actionable error copy and keep destructive actions confirmed.</p>
            </div>
          </AppCard>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <AppCard title="Inline empty state" description="For existing cards and constrained panels.">
            <EmptyState variant="inline" density="compact" icon={Inbox} title="No matching records" />
          </AppCard>
          <AppCard title="Table empty state" description="For a spanning cell below a retained table header.">
            <Table><TableBody><TableEmptyRow colSpan={3} density="compact" title="No rows to display" /></TableBody></Table>
          </AppCard>
          <AppCard title="Table loading rows" description="Keep table geometry stable while the first page loads.">
            <Table><TableBody><TableLoadingRows colSpan={3} rowCount={2} label="Loading rows" /></TableBody></Table>
          </AppCard>
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-[1fr_20rem]">
        <DataListPanel
          aria-label="Resources"
          toolbar={<SearchField placeholder="Search resources" />}
          toolbarMeta="2 resources"
          pagination={
            <TablePagination
              label="Resource pages"
              summary={`Page ${samplePage} of 3`}
              page={samplePage}
              pageSize={25}
              pageSizeLabel="Rows per page"
              previousLabel="Previous"
              nextLabel="Next"
              canPrevious={samplePage > 1}
              canNext={samplePage < 3}
              onPrevious={() => setSamplePage((page) => Math.max(1, page - 1))}
              onNext={() => setSamplePage((page) => Math.min(3, page + 1))}
            />
          }
        >
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell className="font-medium">openai-production</TableCell>
                <TableCell><StatusBadge status="active">active</StatusBadge></TableCell>
                <TableCell className="text-muted-foreground"><DateTimeValue value="2025-03-12T09:30:00Z" locale="en" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell className="font-medium">budget-policy</TableCell>
                <TableCell><StatusBadge status="warning">review</StatusBadge></TableCell>
                <TableCell className="text-muted-foreground">Yesterday</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </DataListPanel>

        <AppCard title="Confirmation">
          <ConfirmDialog
            trigger={<Button variant="destructive">Open dialog</Button>}
            title="Confirm destructive action"
            description="Use ConfirmDialog before deleting or revoking important resources."
            cancelLabel="Cancel"
            confirmLabel="Confirm"
            pendingLabel="Confirming..."
            onConfirm={() => new Promise((resolve) => window.setTimeout(resolve, 800))}
          />
          <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <CheckCircle2 className="text-success" aria-hidden="true" />
            Dialog keeps title and footer structure consistent.
          </div>
          <Button className="mt-4" variant="outline" onClick={() => setUnsavedChangesOpen(true)}>Show unsaved changes</Button>
          <UnsavedChangesDialog open={unsavedChangesOpen} onOpenChange={setUnsavedChangesOpen} title="Unsaved changes" description="Your edits will be lost if you continue." stayLabel="Keep editing" discardLabel="Discard changes" onDiscard={() => undefined} />
        </AppCard>
        <AppCard title="Commercial plan restriction" description="Use one consistent dialog when an entitlement blocks an action.">
          <Button type="button" onClick={() => setUpgradeDialogOpen(true)}>Open upgrade dialog</Button>
          <CommercialPlanUpgradeDialog
            open={upgradeDialogOpen}
            onOpenChange={setUpgradeDialogOpen}
            title="Feature unavailable on this plan"
            description="Upgrade the commercial plan to enable this feature."
            cancelLabel="Not now"
            actionLabel="Upgrade plan"
          />
        </AppCard>
      </section>

      <section className="grid gap-4 lg:grid-cols-3">
        <AppCard title="Select field" description="Use AppSelectField for labelled, grouped, and validated option sets.">
          <AppSelectField
            label="Account status"
            description="Grouped options keep longer menus scannable."
            value={selectValue}
            onValueChange={setSelectValue}
            groups={[
              {
                label: "Available",
                options: [
                  { value: "active", label: "Active" },
                  { value: "pending", label: "Pending" },
                ],
              },
              {
                label: "Unavailable",
                options: [{ value: "disabled", label: "Disabled", disabled: true }],
              },
            ]}
          />
        </AppCard>
        <AppCard title="Filter combobox" description="Use FilterCombobox for searchable toolbar filters.">
          <FilterCombobox
            id="playground-filter-combobox"
            value={filterValue}
            onValueChange={setFilterValue}
            placeholder="Search provider"
            emptyText="No provider found"
            options={[
              { value: "openai", label: "OpenAI" },
              { value: "anthropic", label: "Anthropic" },
              { value: "google", label: "Google" },
            ]}
          />
        </AppCard>
        <AppCard title="Filter fields" description="Use debounced text and range filters in advanced filter toolbars.">
          <div className="grid gap-4 sm:grid-cols-2">
            <DebouncedTextFilter label="Request ID" value={textFilterValue} onValueChange={setTextFilterValue} mono />
            <RangeFilterField label="Latency" value={rangeFilterValue} min={0} max={100} minLabel="Minimum" maxLabel="Maximum" formatValue={(value) => `${value} ms`} onValueChange={setRangeFilterValue} onValueCommit={setRangeFilterValue} />
          </div>
        </AppCard>
        <AppCard title="Row actions" description="Use RowActionsMenu for table and card action menus.">
          <RowActionsMenu
            label="Resource actions"
            actions={[
              { label: "Edit", href: "/design-system" },
              { label: "Members" },
              { label: "Documentation", href: "https://ui.shadcn.com/docs/components/dropdown-menu", external: true },
              { label: "Delete", destructive: true, separatorBefore: true, icon: <Trash2 /> },
            ]}
          />
        </AppCard>
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <AppCard title="Playground composition" description="RTL split pane, compact message editing, and aligned run telemetry.">
          <div className="h-96 overflow-hidden rounded-lg border border-border" dir="rtl">
            <ResizableSplitPane
              direction="rtl"
              startLabel="قالب پرامپت"
              endLabel="پاسخ مدل"
              separatorLabel="تغییر اندازه قالب و پاسخ"
              start={<div className="h-full overflow-auto p-3"><PromptMessageEditor role="user" content={sampleMessage} highlightedVariables={["topic"]} roleOptions={[{ value: "system", label: "راهنما" }, { value: "user", label: "کاربر" }, { value: "assistant", label: "دستیار" }, { value: "tool", label: "ابزار" }]} roleLabel="نقش پیام" contentLabel="محتوای پیام" toolCallIdLabel="شناسه فراخوانی ابزار" duplicateLabel="تکثیر" removeLabel="حذف" dragLabel="دستگیره جابه‌جایی پیام" moveUpLabel="بالا" moveDownLabel="پایین" disableMoveUp disableMoveDown onRoleChange={() => undefined} onContentChange={setSampleMessage} onDuplicate={() => undefined} onRemove={() => undefined} onMoveUp={() => undefined} onMoveDown={() => undefined} /></div>}
              end={<div className="flex h-full flex-col justify-between gap-4 overflow-auto p-3"><MarkdownRenderer content={'## پاسخ نمونه\n\nاین پاسخ از **Markdown** و قطعه‌کد پشتیبانی می‌کند:\n\n```typescript\nconst ready: boolean = true;\n```'} /><RunTelemetryStrip unavailableLabel="—" copyLabel="کپی" copiedLabel="کپی شد" items={[{ key: "tokens", label: "مجموع توکن", value: "۱۸۲" }, { key: "duration", label: "زمان", value: "۸۴۲ ms" }, { key: "cost", label: "هزینه", value: "۹۵ ریال" }, { key: "trace", label: "شناسه رد", value: "trace_demo", rawValue: "trace_demo", technical: true }]} /></div>}
            />
          </div>
        </AppCard>
        <AppCard title="Attachment" description="Use Attachment for file and image selections with inline actions.">
          <Attachment orientation="vertical">
            <AttachmentMedia>
              <Upload aria-hidden="true" />
            </AttachmentMedia>
            <AttachmentContent>
              <AttachmentTitle>bank-receipt.png</AttachmentTitle>
              <AttachmentDescription>PNG · 820 KB</AttachmentDescription>
            </AttachmentContent>
            <AttachmentActions>
              <AttachmentAction aria-label="Remove bank-receipt.png">
                <X />
              </AttachmentAction>
            </AttachmentActions>
          </Attachment>
        </AppCard>
        <AppCard title="Message" description="Use Message for direction-aware conversation rows with sender metadata and a constrained surface.">
          <div className="flex flex-col gap-4" dir="rtl">
            <Message align="start">
              <MessageContent className="max-w-xl">
                <MessageHeader className="justify-between gap-2 px-1">
                  <strong className="text-foreground">پشتیبانی</strong>
                  <span>امروز، ۱۰:۳۰</span>
                </MessageHeader>
                <div className="w-full rounded-xl border border-card-border bg-card p-4 text-card-foreground shadow-xs">
                  پاسخ شما دریافت شد و در حال بررسی است.
                </div>
              </MessageContent>
            </Message>
            <Message align="end">
              <MessageContent className="max-w-xl group-data-[align=end]/message:ms-auto">
                <MessageHeader className="justify-between gap-2 px-1">
                  <strong className="text-foreground">شما</strong>
                  <span>امروز، ۱۰:۳۲</span>
                </MessageHeader>
                <div className="w-full rounded-xl border border-card-border bg-card p-4 text-card-foreground shadow-xs">
                  ممنون، منتظر پاسخ تیم پشتیبانی هستم.
                </div>
              </MessageContent>
            </Message>
          </div>
        </AppCard>
      </section>
    </PageContainer>
  );
}
