"use client";

import { memo, useState } from "react";
import { Plus, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { cn } from "@/lib/utils";

import type { SavedView } from "./logs-types";

type SavedViewsRowProps = {
  views: SavedView[];
  activeViewID: string | null;
  onApply: (view: SavedView) => void;
  onDelete: (viewID: string) => void | Promise<void>;
  onSaveCurrent: () => void;
  deletingViewID: string | null;
  t: ReturnType<typeof useTranslations>;
};

/**
 * One chip per saved view above the log table. The chip matching the current filters is pressed;
 * its delete control is always visible, while other chips reveal it on hover or keyboard focus.
 */
export const SavedViewsRow = memo(function SavedViewsRow({
  views,
  activeViewID,
  onApply,
  onDelete,
  onSaveCurrent,
  deletingViewID,
  t,
}: SavedViewsRowProps) {
  const [pendingDelete, setPendingDelete] = useState<SavedView | null>(null);

  return (
    <div className="flex flex-wrap items-center gap-2 text-xs" role="group" aria-label={t("savedViews.title")}>
      <span className="font-medium text-muted-foreground">{t("savedViews.rowLabel")}</span>
      {views.map((view) => {
        const active = view.saved_view_id === activeViewID;
        const deleting = view.saved_view_id === deletingViewID;
        return (
          <div
            key={view.saved_view_id}
            className={cn(
              "group/chip inline-flex h-7.5 max-w-60 items-center rounded-full border transition-colors",
              active
                ? "border-primary-subtle-border bg-primary-soft font-semibold text-primary-soft-foreground"
                : "border-border bg-card text-foreground hover:bg-accent",
              deleting && "opacity-60",
            )}
          >
            <button
              type="button"
              aria-pressed={active}
              title={view.description || view.name}
              onClick={() => onApply(view)}
              className="h-full min-w-0 truncate rounded-full px-3 outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
            >
              <span dir="auto">{view.name}</span>
            </button>
            <button
              type="button"
              aria-label={t("savedViews.deleteView", { name: view.name })}
              title={t("savedViews.deleteView", { name: view.name })}
              disabled={deleting}
              onClick={() => setPendingDelete(view)}
              className={cn(
                "-ms-1.5 me-1.5 size-5 shrink-0 items-center justify-center rounded-full text-current opacity-70 outline-none hover:bg-foreground/10 hover:opacity-100 focus-visible:ring-[3px] focus-visible:ring-ring/50",
                active ? "inline-flex" : "hidden group-focus-within/chip:inline-flex group-hover/chip:inline-flex",
              )}
            >
              <X aria-hidden="true" className="size-3" />
            </button>
          </div>
        );
      })}
      <button
        type="button"
        onClick={onSaveCurrent}
        className="inline-flex h-7.5 items-center gap-1 rounded-full border border-dashed border-border px-3 text-muted-foreground outline-none transition-colors hover:border-primary-subtle-border hover:text-foreground focus-visible:ring-[3px] focus-visible:ring-ring/50"
      >
        <Plus aria-hidden="true" className="size-3.5" />
        {t("savedViews.saveCurrent")}
      </button>
      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null);
        }}
        title={t("savedViews.deleteTitle")}
        description={t("savedViews.deleteDescription", { name: pendingDelete?.name ?? "" })}
        confirmLabel={t("actions.delete")}
        cancelLabel={t("actions.cancel")}
        pendingLabel={t("savedViews.deleting")}
        onConfirm={async () => {
          if (pendingDelete) await onDelete(pendingDelete.saved_view_id);
        }}
      />
    </div>
  );
});
