"use client";

import { useState } from "react";
import { Activity, AlertTriangle, CheckCircle2, Database, Inbox, KeyRound, Trash2, Upload, X } from "lucide-react";

import { AppCard, AppSelectField, ConfirmDialog, EmptyState, FilterCombobox, FormSection, MetricCard, PageHeader, RowActionsMenu, SectionHeader, SettingsCard, StatusBadge } from "@/components/app-ui";
import { Attachment, AttachmentAction, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-6 md:px-6">
      <PageHeader
        eyebrow="Design System"
        title="Tenant Admin UI Foundation"
        description="Internal playground for product-level UI components, semantic tokens, and RTL-friendly layout patterns."
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
        <MetricCard label="Requests" value="128K" description="Current billing period" trend={<StatusBadge status="success">+12%</StatusBadge>} icon={Activity} />
        <MetricCard label="Provider accounts" value="8" description="Active connections" icon={Database} />
        <MetricCard label="Access keys" value="24" description="Issued personal tokens" icon={KeyRound} />
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
        footer={<><Button variant="outline">Cancel</Button><Button>Submit</Button></>}
      >
        <Input placeholder="API base URL" dir="ltr" />
        <Input placeholder="Token alias" />
      </FormSection>

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
        <AppCard title="Table pattern" description="Keep tables readable in RTL and use status badges for state.">
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
        </AppCard>

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
