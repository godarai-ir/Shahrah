"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Copy, Gauge, Pencil, Plus, Power, Search, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { CommercialPlanUpgradeDialog, EmptyState, HelpLink, PageHeader, RowActionsMenu, StatusBadge } from "@/components/app-ui";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import {
  createRateLimitRule,
  deleteRateLimitRule,
  getRateLimitEntitlements,
  listRateLimitRules,
  updateRateLimitRuleStatus,
} from "./ruleApi";
import type { RateLimitRule } from "./ruleTypes";
import { RateLimitRuleForm } from "./RateLimitRuleForm";

const PERSIAN_DATE_TIME = new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
});
const FEATURE_RATE_LIMIT_RULES_ENABLED = "rate_limit_rules.enabled";

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

function formatJalaliDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return PERSIAN_DATE_TIME.format(date);
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

export function RateLimitRulesPage({ docsHref }: { docsHref?: string } = {}) {
  const t = useTranslations("rateLimitRules");
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [items, setItems] = useState<RateLimitRule[]>([]);
  const [pendingDeleteRule, setPendingDeleteRule] = useState<RateLimitRule | null>(null);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const response = await listRateLimitRules(query);
      setItems(response.data);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.loadError") });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    let isMounted = true;

    async function loadEntitlements() {
      try {
        const response = await getRateLimitEntitlements();
        if (!isMounted) return;
        setEntitlements(response);
      } catch {
        if (!isMounted) return;
        setEntitlements(null);
      } finally {
        if (isMounted) {
          setHasLoadedEntitlements(true);
        }
      }
    }

    void loadEntitlements();

    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const filtered = useMemo(() => items, [items]);
  const isRateLimitFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_RATE_LIMIT_RULES_ENABLED], false)
    : null;

  async function confirmDeleteRule() {
    if (!pendingDeleteRule || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteRateLimitRule(pendingDeleteRule.rule_id);
      setPendingDeleteRule(null);
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.deleteError") });
    } finally {
      setIsDeleting(false);
    }
  }

  function handleCreateRule() {
    if (isRateLimitFeatureEnabled === false) {
      setIsUpgradeDialogOpen(true);
      return;
    }

    setIsCreateSheetOpen(true);
  }

  return (
    <>
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 md:px-6">
        <PageHeader
          title={t("title")}
          description={t("description")}
          titleAddon={docsHref ? <HelpLink href={docsHref} /> : null}
          actions={
            <Button type="button" onClick={handleCreateRule} disabled={!hasLoadedEntitlements}>
              <Plus data-icon="inline-start" />
              {t("actions.new")}
            </Button>
          }
        />

        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4 sm:flex-row sm:items-center sm:justify-between">
            <InputGroup className="max-w-xl">
              <InputGroupAddon><Search aria-hidden="true" /></InputGroupAddon>
              <InputGroupInput placeholder={t("searchPlaceholder")} value={query} onChange={(event) => setQuery(event.target.value)} aria-label={t("searchPlaceholder")} />
            </InputGroup>
            <p className="text-sm text-muted-foreground">{t("resultCount", { count: filtered.length })}</p>
          </div>

        {isLoading && filtered.length === 0 ? (
          <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4" aria-label={t("messages.loading")}>
            {Array.from({ length: 5 }).map((_, index) => <Skeleton key={index} className="h-10 w-full" />)}
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Gauge}
            title={query ? t("noResults.title") : t("emptyState.title")}
            description={query ? t("noResults.description") : t("emptyState.description")}
            primaryAction={!query ? <Button type="button" onClick={handleCreateRule} disabled={!hasLoadedEntitlements}><Plus data-icon="inline-start" />{t("actions.new")}</Button> : undefined}
          />
        ) : (
          <div className="overflow-x-auto rounded-xl border border-border bg-card">
            <Table className="text-xs lg:text-sm">
              <TableHeader>
                <TableRow>
                  <TableHead className="px-1.5 lg:px-2">{t("table.ruleId")}</TableHead>
                  <TableHead className="px-1.5 lg:px-2">{t("table.limit")}</TableHead>
                  <TableHead className="px-1.5 lg:px-2">{t("table.window")}</TableHead>
                  <TableHead className="px-1.5 lg:px-2">{t("table.status")}</TableHead>
                  <TableHead className="px-1.5 lg:px-2">{t("table.updatedAt")}</TableHead>
                  <TableHead className="px-1.5 lg:px-2">{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="px-1.5 font-mono font-medium lg:px-2" dir="ltr">{rule.rule_id}</TableCell>
                    <TableCell className="px-1.5 lg:px-2">
                      <div className="flex items-center gap-2">
                        <span className="font-mono" dir="ltr">{rule.limit.max}</span>
                        <Badge variant="outline">{t(`form.options.dimension.${rule.limit.dimension}`)}</Badge>
                      </div>
                    </TableCell>
                    <TableCell className="px-1.5 lg:px-2">{`${rule.limit.window.value} ${t(`table.windowUnits.${rule.limit.window.unit}`)}`}</TableCell>
                    <TableCell className="px-1.5 lg:px-2">
                      <StatusBadge status={rule.enabled ? "active" : "inactive"}>{rule.enabled ? t("status.enabled") : t("status.disabled")}</StatusBadge>
                    </TableCell>
                    <TableCell className="px-1.5 lg:px-2">{formatJalaliDateTime(rule.updated_at)}</TableCell>
                    <TableCell className="px-1.5 lg:px-2">
                      <RowActionsMenu label={t("table.rowActions", { ruleId: rule.rule_id })} actions={[
                        { label: t("actions.edit"), icon: <Pencil />, onSelect: () => router.push(`/gateway-policies/rate-limiting-rules/${encodeURIComponent(rule.rule_id)}`) },
                        { label: rule.enabled ? t("actions.disable") : t("actions.enable"), icon: <Power />, onSelect: () => void updateRateLimitRuleStatus(rule.rule_id, !rule.enabled).then(load) },
                        { label: t("actions.duplicate"), icon: <Copy />, onSelect: () => void createRateLimitRule({ ...rule, rule_id: `${rule.rule_id}-copy`, name: `${rule.name ?? rule.rule_id} Copy` } as never).then(load) },
                        { label: t("actions.delete"), icon: <Trash2 />, destructive: true, onSelect: () => setPendingDeleteRule(rule) },
                      ]} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
        </div>
      </div>

      <Sheet open={isCreateSheetOpen} onOpenChange={setIsCreateSheetOpen}>
        <SheetContent side="left" closeLabel={t("actions.close")} className="w-full overflow-y-auto p-0 sm:max-w-3xl" dir="rtl">
          <SheetHeader className="sr-only">
            <SheetTitle>{t("form.addTitle")}</SheetTitle>
            <SheetDescription>{t("form.headerDescription")}</SheetDescription>
          </SheetHeader>
          <RateLimitRuleForm embedded onCancel={() => setIsCreateSheetOpen(false)} onSaved={async () => { setIsCreateSheetOpen(false); await load(); }} />
        </SheetContent>
      </Sheet>

      <Dialog open={pendingDeleteRule !== null} onOpenChange={(open) => !open && setPendingDeleteRule(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>{t("deleteDialog.description", { ruleId: pendingDeleteRule?.rule_id ?? "-" })}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" disabled={isDeleting} onClick={() => setPendingDeleteRule(null)}>
              {t("deleteDialog.cancel")}
            </Button>
            <Button type="button" variant="destructive" onClick={() => void confirmDeleteRule()} disabled={isDeleting}>
              {isDeleting ? t("deleteDialog.deleting") : t("deleteDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <CommercialPlanUpgradeDialog open={isUpgradeDialogOpen} onOpenChange={setIsUpgradeDialogOpen} title={t("upgradeDialog.title")} description={t("upgradeDialog.description")} cancelLabel={t("upgradeDialog.cancel")} actionLabel={t("upgradeDialog.action")} />
    </>
  );
}
