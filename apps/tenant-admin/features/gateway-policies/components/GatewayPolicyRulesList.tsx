"use client";

import * as React from "react";
import { createPortal } from "react-dom";
import type { LucideIcon } from "lucide-react";

import { CommercialPlanUpgradeDialog } from "@/components/app-ui/CommercialPlanUpgradeDialog";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { DataListPanel } from "@/components/app-ui/DataListPanel";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";

const GatewayPolicyRuleSheetFooterContext = React.createContext<HTMLElement | null>(null);

export function GatewayPolicyRuleSheetFooter({ children }: { children: React.ReactNode }) {
  const host = React.useContext(GatewayPolicyRuleSheetFooterContext);
  return host ? createPortal(children, host) : null;
}

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
  feedback?: React.ReactNode;
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
  /** The create control; the owning page decides whether it is a button or a type menu. */
  createAction: React.ReactNode;
  notice?: GatewayPolicyNotice | null;
  alertLabels: { error: React.ReactNode; success: React.ReactNode };
  /** Search and filter controls. */
  toolbar: React.ReactNode;
  /** Whether a search or filter is narrowing the list; hides the create action in the empty state. */
  filtered: boolean;
  /** Side column beside the list on wide screens, below it on narrow ones. */
  aside?: React.ReactNode;
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
  createAction,
  notice,
  alertLabels,
  toolbar,
  filtered,
  aside,
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
  const [ruleSheetFooterHost, setRuleSheetFooterHost] = React.useState<HTMLDivElement | null>(null);
  return (
    <>
      <PageContainer className={className}>
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? alertLabels.error : alertLabels.success}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <div className="grid min-w-0 items-start gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <DataListPanel
          aria-label={typeof title === "string" ? title : undefined}
          toolbar={toolbar}
          toolbarMeta={createAction}
          contentClassName="@container/gateway-rules min-w-0"
          loading={loading && itemCount === 0}
          loadingLabel={loadingLabel}
          empty={itemCount === 0}
          emptyState={(
            <EmptyState
              icon={emptyIcon}
              title={emptyTitle}
              description={emptyDescription}
              primaryAction={!filtered ? createAction : undefined}
            />
          )}
        >
          {children}
        </DataListPanel>
        {aside ? <aside className="min-w-0 xl:sticky xl:top-4">{aside}</aside> : null}
        </div>
      </PageContainer>

      <Sheet open={ruleSheet.open} onOpenChange={ruleSheet.onOpenChange}>
        <SheetContent
          side="left"
          closeLabel={ruleSheet.closeLabel}
          className="flex w-full flex-col overflow-hidden p-0 sm:max-w-3xl"
          dir={ruleSheet.dir}
        >
          <SheetHeader className="sr-only">
            <SheetTitle>{ruleSheet.title}</SheetTitle>
            <SheetDescription>{ruleSheet.description}</SheetDescription>
          </SheetHeader>
          <GatewayPolicyRuleSheetFooterContext.Provider value={ruleSheetFooterHost}>
            <div className="min-h-0 flex-1 overflow-y-auto">
              <React.Fragment key={ruleSheet.contentKey}>{ruleSheet.content}</React.Fragment>
            </div>
          </GatewayPolicyRuleSheetFooterContext.Provider>
          <SheetFooter className="shrink-0 border-t border-border px-6 py-4">
            <div ref={setRuleSheetFooterHost} className="w-full" />
          </SheetFooter>
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
        feedback={deleteDialog.feedback}
      />

      <CommercialPlanUpgradeDialog {...upgradeDialog} />
    </>
  );
}
