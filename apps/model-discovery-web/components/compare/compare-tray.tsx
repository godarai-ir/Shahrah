"use client";

import Link from "next/link";
import { X } from "lucide-react";
import { useTranslations } from "next-intl";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCompareModels } from "@/components/compare/compare-provider";
import { MAX_COMPARE_MODELS, serializeCompareIds } from "@/lib/compare-models";

export function CompareTray() {
  const t = useTranslations("Compare");
  const pathname = usePathname();
  const { selectedIds, selectedCount, remove, clear, notice } = useCompareModels();

  if (selectedCount === 0 || pathname === "/compare") return null;

  return (
    <div className="fixed inset-x-0 bottom-0 z-30 border-t border-border bg-background/95 backdrop-blur">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-2 px-3 py-2 sm:px-4 md:flex-row md:items-center md:py-3">
        <div className="flex items-center gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <span className="truncate text-sm text-foreground">{t("tray")}</span>
            <Badge>{selectedCount}/{MAX_COMPARE_MODELS}</Badge>
          </div>
          <div className="ms-auto flex shrink-0 items-center gap-2 md:hidden">
            <Link
              href={`/compare?ids=${encodeURIComponent(serializeCompareIds(selectedIds))}`}
              aria-label={t("openPage")}
              className="inline-flex h-8 items-center rounded-md bg-primary px-2.5 text-xs font-medium text-primary-foreground hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {t("openCompare")}
            </Link>
            <Button type="button" variant="secondary" size="sm" onClick={clear} aria-label={t("clearSelection")}>
              {t("clearAll")}
            </Button>
          </div>
        </div>
        <div className="scrollbar-none flex min-w-0 gap-1 overflow-x-auto">
          {selectedIds.map((id) => (
            <button
              key={id}
              type="button"
              className="inline-flex shrink-0 items-center gap-1 rounded border border-border px-2 py-1 text-xs text-muted-foreground hover:border-primary hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              onClick={() => remove(id)}
              aria-label={t("removeAria", { id })}
            >
              <bdi dir="ltr" data-technical>{id}</bdi>
              <X className="size-3" aria-hidden="true" />
            </button>
          ))}
        </div>
        <div className="ms-auto hidden items-center gap-2 md:flex">
          <Link
            href={`/compare?ids=${encodeURIComponent(serializeCompareIds(selectedIds))}`}
            aria-label={t("openPage")}
            className="inline-flex h-8 items-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            {t("openCompare")}
          </Link>
          <Button type="button" variant="secondary" size="sm" onClick={clear} aria-label={t("clearSelection")}>{t("clearAll")}</Button>
        </div>
      </div>
      {notice ? <p className="mx-auto max-w-7xl px-4 pb-3 text-xs text-warning-soft-foreground" aria-live="polite">{notice}</p> : null}
    </div>
  );
}

