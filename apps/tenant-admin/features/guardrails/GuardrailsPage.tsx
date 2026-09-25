"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { Pencil, Plus, Trash2 } from "lucide-react";

import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { HelpLink } from "@/components/app-ui/HelpLink";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { SearchField } from "@/components/app-ui/SearchField";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

import { deleteGuardrail, listGuardrails } from "./api";
import { GUARDRAIL_CATALOG } from "./catalog";
import { localizedError } from "./errors";
import { formFromGuardrail } from "./form";
import { guardrailSentence } from "./sentence";
import type { TenantGuardrail } from "./types";
import { useGuardrailFormat } from "./use-guardrail-format";

const GUARDRAIL_COLUMN_CLASS = {
  name: "w-[52%]",
  type: "hidden w-36 @md:table-cell",
  execution: "hidden w-48 @lg:table-cell",
  status: "w-[28%]",
  updatedAt: "hidden w-40 @lg:table-cell",
  actions: "w-20 text-end",
} as const;

/** Reads the slug the wizard just saved and clears it from the URL so the row is highlighted once. */
function useJustSavedSlug() {
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const router = useRouter();
  const savedParam = searchParams.get("created") ?? searchParams.get("updated");
  const [justSaved] = useState(savedParam);

  useEffect(() => {
    if (savedParam) router.replace(pathname, { scroll: false });
  }, [pathname, router, savedParam]);

  return justSaved;
}

export function GuardrailsPage({ docsHref }: { docsHref: string }) {
  const t = useTranslations("guardrails");
  const locale = useLocale();
  const format = useGuardrailFormat();
  const justSaved = useJustSavedSlug();
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<TenantGuardrail[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<TenantGuardrail | null>(null);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const response = await listGuardrails(query);
      setItems(response.data);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.loadError"), format.t) });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const tableItems = useMemo(() => items, [items]);
  const hasSearchQuery = query.trim().length > 0;
  const pendingDeleteSentence = pendingDelete ? guardrailSentence(formFromGuardrail(pendingDelete), format) : "";

  async function confirmDelete() {
    if (!pendingDelete || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteGuardrail(pendingDelete.slug);
      setPendingDelete(null);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteSuccess") });
      await load();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: localizedError(error, t("messages.deleteError"), format.t) });
      throw error;
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <>
      <Card className="@container/guardrails">
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-2 @md:flex-row @md:items-center">
            <SearchField
              containerClassName="min-w-0 flex-1"
              placeholder={t("searchPlaceholder")}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <Button asChild className="w-full @md:w-auto">
              <Link href="/guardrails/new">
                <Plus data-icon="inline-start" aria-hidden="true" />
                {t("actions.new")}
              </Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

          {tableItems.length === 0 && !isLoading ? (
            <EmptyState
              variant="inline"
              density="compact"
              title={hasSearchQuery ? t("emptySearch") : t("empty")}
              primaryAction={!hasSearchQuery ? (
                <Button asChild size="sm">
                  <Link href="/guardrails/new">
                    <Plus data-icon="inline-start" aria-hidden="true" />
                    {t("actions.new")}
                  </Link>
                </Button>
              ) : undefined}
              secondaryAction={!hasSearchQuery ? <HelpLink href={docsHref} /> : undefined}
            />
          ) : (
            <div className="overflow-hidden rounded-lg border">
              <Table className="table-fixed text-xs lg:text-sm">
                <TableHeader>
                  <TableRow>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.name}>{t("table.name")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.type}>{t("table.type")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.execution}>{t("table.execution")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.status}>{t("table.status")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.updatedAt}>{t("table.updatedAt")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.actions}>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tableItems.map((item) => {
                    const TypeIcon = GUARDRAIL_CATALOG.find((entry) => entry.type === item.type)?.icon;
                    return (
                    <TableRow key={item.id} data-highlighted={item.slug === justSaved || undefined} className={cn(item.slug === justSaved && "bg-primary-soft hover:bg-primary-soft")}>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.name}>
                        <div className="flex min-w-0 flex-col gap-1">
                          <span className="truncate font-medium" title={item.name}>{item.name}</span>
                          <div className="flex min-w-0 items-center gap-2">
                            <code className="truncate rounded bg-muted px-2 py-1 text-xs text-muted-foreground" dir="ltr">
                              {item.selector}
                            </code>
                            <CopyAction value={item.selector} copyLabel={t("actions.copySelector")} copiedLabel={t("actions.copied")} size="icon" />
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.type}>
                        <span className="inline-flex min-w-0 items-center gap-1.5">
                          {TypeIcon ? <TypeIcon className="size-4 shrink-0 text-primary" aria-hidden="true" data-guardrail-type-icon={item.type} /> : null}
                          <span className="truncate">{t(`types.${item.type}`)}</span>
                        </span>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.execution} title={guardrailSentence(formFromGuardrail(item), format)}>
                        <span className="block truncate text-muted-foreground">
                          {t(`operations.${item.operation}`)} · {t(`enforcement.${item.enforcement}`)}
                        </span>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.status}>
                        <span className="inline-flex items-center gap-1.5">
                          <span className={cn("size-2 shrink-0 rounded-full", item.enabled ? "bg-success" : "bg-input-border")} aria-hidden="true" />
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </span>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.updatedAt}><DateTimeValue value={item.updated_at} locale={locale} format={{ year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }} /></TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.actions}>
                        <div className="flex justify-end gap-2">
                          <Button asChild size="icon" variant="outline">
                            <Link href={`/guardrails/${encodeURIComponent(item.slug)}/edit`} aria-label={t("actions.edit")}>
                              <Pencil data-icon="inline-start" aria-hidden="true" />
                            </Link>
                          </Button>
                          <Button type="button" size="icon" variant="outline" aria-label={t("actions.delete")} onClick={() => setPendingDelete(item)}>
                            <Trash2 data-icon="inline-start" aria-hidden="true" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingDelete(null);
            setDeleteNotice(null);
          }
        }}
        title={t("deleteDialog.title")}
        description={(
          <>
            {t("deleteDialog.description", { name: pendingDelete?.name ?? "-" })}
            {pendingDeleteSentence ? <span className="mt-2 block">{t("deleteDialog.rule", { sentence: pendingDeleteSentence })}</span> : null}
          </>
        )}
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.delete")}
        pendingLabel={t("deleteDialog.deleting")}
        onConfirm={confirmDelete}
        pending={isDeleting}
        closeOnConfirm={false}
        feedback={<OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />}
      />
    </>
  );
}
