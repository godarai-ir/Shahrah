"use client";

import { useState } from "react";
import { Activity, AlertTriangle, CheckCircle2, Database, Inbox, KeyRound, Trash2, Upload, X } from "lucide-react";

import { AppCard, AppSelectField, CompactValue, CommercialPlanUpgradeDialog, ConfirmDialog, DataListPanel, EmptyState, FilterCombobox, FormSection, HelpLink, IdentifierField, MetricCard, PageHeader, RowActionsMenu, SearchField, SectionHeader, SettingsCard, StatusBadge, TablePagination } from "@/components/app-ui";
import { Attachment, AttachmentAction, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
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

export default function DesignSystemPage() {
  const [selectValue, setSelectValue] = useState("active");
  const [filterValue, setFilterValue] = useState("openai");
  const [samplePage, setSamplePage] = useState(1);
  const [upgradeDialogOpen, setUpgradeDialogOpen] = useState(false);
  const requestsDisplay = formatCompactNumber(128_400, { locale: "en-US" });
  const spendDisplay = formatCompactCurrency(12_340, { locale: "en-US", currency: "USD" });
  const irrDisplay = appendDisplaySuffix(formatCompactNumber(845_000, { locale: "fa-IR" }), " IRR");

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-6 md:px-6">
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
        <div className="grid gap-4 lg:grid-cols-3">
          <EmptyState icon={Inbox} title="No records yet" description="Use EmptyState for empty lists and missing data." primaryAction={<Button size="sm">Create</Button>} />
          <AppCard title="Loading">
            <div className="flex flex-col gap-3">
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
                <TableCell className="text-muted-foreground">Today</TableCell>
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
        <AppCard title="Select field" description="Use AppSelectField for compact option sets in forms and toolbars.">
          <AppSelectField
            value={selectValue}
            onValueChange={setSelectValue}
            options={[
              { value: "active", label: "Active" },
              { value: "pending", label: "Pending" },
              { value: "disabled", label: "Disabled" },
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
        <AppCard title="Row actions" description="Use RowActionsMenu for table row action menus.">
          <RowActionsMenu
            label="Resource actions"
            actions={[
              { label: "Edit" },
              { label: "Members" },
              { label: "Delete", destructive: true, icon: <Trash2 /> },
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
    </main>
  );
}
