"use client";

import * as React from "react";
import type { LucideIcon } from "lucide-react";
import { Plus } from "lucide-react";

import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { DataListPanel } from "@/components/app-ui/DataListPanel";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { SearchField } from "@/components/app-ui/SearchField";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";

export type GatewayPolicyNotice = {
  kind: "success" | "error";
  message: React.ReactNode;
};

export type GatewayPolicyRuleSheet = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  closeLabel: string;
  content: React.ReactNode;
  contentKey?: React.Key;
  dir?: "rtl" | "ltr";
};

export type GatewayPolicyDeleteDialog = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  cancelLabel: string;
  confirmLabel: string;
  pendingLabel: string;
  pending: boolean;
  onConfirm: () => void | Promise<void>;
};

export type GatewayPolicyUpgradeDialog = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  cancelLabel: string;
  actionLabel: string;
};

export type GatewayPolicyRulesListProps = {
  title: React.ReactNode;
  createLabel: React.ReactNode;
  onCreate: () => void;
  createDisabled?: boolean;
  notice?: GatewayPolicyNotice | null;
  alertLabels: { error: React.ReactNode; success: React.ReactNode };
  query: string;
  onQueryChange: (query: string) => void;
  searchPlaceholder: string;
  resultCountLabel: React.ReactNode;
  itemCount: number;
  loading: boolean;
  loadingLabel: string;
  emptyIcon: LucideIcon;
  emptyTitle: React.ReactNode;
  emptyDescription: React.ReactNode;
  children: React.ReactNode;
  ruleSheet: GatewayPolicyRuleSheet;
  deleteDialog: GatewayPolicyDeleteDialog;
  upgradeDialog: GatewayPolicyUpgradeDialog;
  className?: string;
};

export function GatewayPolicyRulesList({
  title,
  createLabel,
  onCreate,
  createDisabled,
  notice,
  alertLabels,
  query,
  onQueryChange,
  searchPlaceholder,
  resultCountLabel,
  itemCount,
  loading,
  loadingLabel,
  emptyIcon,
  emptyTitle,
  emptyDescription,
  children,
  ruleSheet,
  deleteDialog,
  upgradeDialog,
  className,
}: GatewayPolicyRulesListProps) {
  const createAction = (
    <Button type="button" onClick={onCreate} disabled={createDisabled}>
      <Plus data-icon="inline-start" />
      {createLabel}
    </Button>
  );

  return (
    <>
      <PageContainer className={className}>
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? alertLabels.error : alertLabels.success}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <DataListPanel
          aria-label={typeof title === "string" ? title : undefined}
          toolbar={(
            <SearchField
              containerClassName="max-w-xl"
              placeholder={searchPlaceholder}
              value={query}
              onChange={(event) => onQueryChange(event.target.value)}
            />
          )}
          toolbarMeta={(
            <div className="flex items-center gap-3">
              <span className="whitespace-nowrap">{resultCountLabel}</span>
              {createAction}
            </div>
          )}
          loading={loading && itemCount === 0}
          loadingLabel={loadingLabel}
          empty={itemCount === 0}
          emptyState={(
            <EmptyState
              icon={emptyIcon}
              title={emptyTitle}
              description={emptyDescription}
              primaryAction={!query ? createAction : undefined}
            />
          )}
        >
          {children}
        </DataListPanel>
      </PageContainer>

      <Sheet open={ruleSheet.open} onOpenChange={ruleSheet.onOpenChange}>
        <SheetContent
          side="left"
          closeLabel={ruleSheet.closeLabel}
          className="w-full overflow-y-auto p-0 sm:max-w-3xl"
          dir={ruleSheet.dir ?? "rtl"}
        >
          <SheetHeader className="sr-only">
            <SheetTitle>{ruleSheet.title}</SheetTitle>
            <SheetDescription>{ruleSheet.description}</SheetDescription>
          </SheetHeader>
          <React.Fragment key={ruleSheet.contentKey}>{ruleSheet.content}</React.Fragment>
        </SheetContent>
      </Sheet>

      <ConfirmDialog
        open={deleteDialog.open}
        onOpenChange={deleteDialog.onOpenChange}
        title={deleteDialog.title}
        description={deleteDialog.description}
        cancelLabel={deleteDialog.cancelLabel}
        confirmLabel={deleteDialog.confirmLabel}
        pendingLabel={deleteDialog.pendingLabel}
        pending={deleteDialog.pending}
        closeOnConfirm={false}
        onConfirm={deleteDialog.onConfirm}
      />

      <CommercialPlanUpgradeDialog {...upgradeDialog} />
    </>
  );
}
