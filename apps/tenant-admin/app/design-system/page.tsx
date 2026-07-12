"use client";

import { useState } from "react";
import { Activity, AlertTriangle, CheckCircle2, Database, Inbox, KeyRound, Shield, Trash2, Upload, X } from "lucide-react";

import { AppCard, AppSelectField, AppSwitchField, AppTextField, AppTextareaField, CodePanel, CompactValue, CommercialPlanUpgradeDialog, ConfirmDialog, CopyAction, CopyableValue, DataListPanel, DateTimeValue, DefinitionList, DefinitionListItem, EmptyState, EntityIdentityHeader, FilterCombobox, FormSection, FormSheet, HelpLink, IdentifierField, MetricCard, OneTimeSecretDialog, OperationNotice, PageContainer, PageHeader, PermissionEditor, PermissionSummary, RowActionsMenu, SearchField, SectionHeader, SettingsCard, StatusBadge, StatusOptionGroup, TableEmptyRow, TableLoadingRows, TablePagination, TypedConfirmDialog } from "@/components/app-ui";
import { Attachment, AttachmentAction, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { GatewayPolicyRulesList } from "@/features/gateway-policies/components";
import { appendDisplaySuffix, formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";

const tokens = [
  ["background", "bg-background text-foreground"],
  ["card", "bg-card text-card-foreground"],
  ["primary", "bg-primary text-primary-foreground"],
  ["secondary", "bg-secondary text-secondary-foreground"],
  ["muted", "bg-muted text-muted-foreground"],
  ["accent", "bg-accent text-accent-foreground"],
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
  const [samplePage, setSamplePage] = useState(1);
  const [permissionQuery, setPermissionQuery] = useState("");
  const [permissionMode, setPermissionMode] = useState<"selected" | "all">("all");
  const [selectedPermissions, setSelectedPermissions] = useState(["role:read"]);
  const [policyQuery, setPolicyQuery] = useState("");
  const [policySheetOpen, setPolicySheetOpen] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [demoSecret, setDemoSecret] = useState<string | null>(null);
  const [statusOption, setStatusOption] = useState<"metadata" | "full">("metadata");
  const [typedConfirmOpen, setTypedConfirmOpen] = useState(false);
  const [formSheetOpen, setFormSheetOpen] = useState(false);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const requestsDisplay = formatCompactNumber(128_400, { locale: "en-US" });
  const spendDisplay = formatCompactCurrency(12_340, { locale: "en-US", currency: "USD" });
  const irrDisplay = appendDisplaySuffix(formatCompactNumber(845_000, { locale: "fa-IR" }), " IRR");

  return (
    <PageContainer as="main" spacing="relaxed">
      <PageHeader
        eyebrow="Design System"
        title="Tenant Admin UI Foundation"
        description="Internal playground for product-level UI components, semantic tokens, and RTL-friendly layout patterns."
        titleAddon={<HelpLink href="/design-system">Title-level help</HelpLink>}
        actions={<Button>Primary action</Button>}
      />

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
          description="Shared list shell while table columns and domain actions remain feature-owned."
          createLabel="New rule"
          onCreate={() => setPolicySheetOpen(true)}
          alertLabels={{ error: "Error", success: "Success" }}
          query={policyQuery}
          onQueryChange={setPolicyQuery}
          searchPlaceholder="Search rules"
          resultCountLabel="1 result"
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
            onOpenChange={setTypedConfirmOpen}
            title="Delete demo team"
            description="This action cannot be undone."
            expectedValue="demo-team"
            inputLabel="Type demo-team to confirm"
            cancelLabel="Cancel"
            confirmLabel="Delete"
            pendingLabel="Deleting"
            onConfirm={() => undefined}
          />
        </AppCard>
        <AppCard title="Form sheet" description="Standard shell for create and edit workflows.">
          <Button type="button" onClick={() => setFormSheetOpen(true)}>Open form sheet</Button>
          <FormSheet open={formSheetOpen} onOpenChange={setFormSheetOpen} title="Create resource" description="Enter the resource details.">
            <FormSection title="General information"><Input placeholder="Resource name" /></FormSection>
          </FormSheet>
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
          <Field>
            <FieldLabel htmlFor="playground-field-name">Display name</FieldLabel>
            <Input id="playground-field-name" placeholder="Production workspace" />
            <FieldDescription>This name is visible to workspace members.</FieldDescription>
          </Field>
          <Field data-invalid>
            <FieldLabel htmlFor="playground-field-slug">Technical identifier</FieldLabel>
            <Input id="playground-field-slug" value="invalid identifier" readOnly aria-invalid />
            <FieldError>Use English letters, numbers, dashes, or underscores.</FieldError>
          </Field>
        </FieldGroup>
        <AppCard title="Text fields" description="Use the app field wrappers for consistent labels, helper text, and validation state.">
          <FieldGroup className="grid gap-4 md:grid-cols-2">
            <AppTextField label="Display name" placeholder="Production team" description="Shown throughout the tenant admin." />
            <AppTextareaField label="Description" placeholder="Describe this resource" />
            <AppSwitchField label="Enable caching" description="Cache matching gateway requests." checked />
          </FieldGroup>
        </AppCard>
        <AppCard>
          <EntityIdentityHeader title="Production service" identifier="svc_production" description="Primary gateway service identity." status={<StatusBadge status="active">Active</StatusBadge>} copy={{ copyLabel: "Copy identifier", copiedLabel: "Identifier copied" }} actions={<Button variant="outline" size="sm">Edit</Button>} />
        </AppCard>
        <CodePanel title="Gateway request" description="A reusable titled code surface with optional actions." displayLanguage="cURL" code={'curl "$GATEWAY_URL/v1/models"'} language="bash" copy={{ copyLabel: "Copy code", copiedLabel: "Code copied" }} />
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
      </section>
    </PageContainer>
  );
}
