"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Copy, Gauge, Pencil, Power, Trash2 } from "lucide-react";

import { RowActionsMenu, StatusBadge } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { GatewayPolicyRulesList } from "@/features/gateway-policies/components";
import { formatDateTime as formatSharedDateTime } from "@/lib/date-time";

import {
  createRateLimitRule,
  deleteRateLimitRule,
  getRateLimitEntitlements,
  listRateLimitRules,
  updateRateLimitRuleStatus,
} from "./ruleApi";
import type { RateLimitRule } from "./ruleTypes";
import { RateLimitRuleForm } from "./RateLimitRuleForm";

const FEATURE_RATE_LIMIT_RULES_ENABLED = "rate_limit_rules.enabled";

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

function formatJalaliDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return formatSharedDateTime(date, { locale: "fa", format: { year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", second: "2-digit" } });
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

export function RateLimitRulesPage() {
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
    <GatewayPolicyRulesList
      title={t("title")}
      createLabel={t("actions.new")}
      onCreate={handleCreateRule}
      createDisabled={!hasLoadedEntitlements}
      notice={notice}
      alertLabels={{ error: t("alerts.error"), success: t("alerts.success") }}
      query={query}
      onQueryChange={setQuery}
      searchPlaceholder={t("searchPlaceholder")}
      resultCountLabel={t("resultCount", { count: items.length })}
      itemCount={items.length}
      loading={isLoading}
      loadingLabel={t("messages.loading")}
      emptyIcon={Gauge}
      emptyTitle={query ? t("noResults.title") : t("emptyState.title")}
      emptyDescription={query ? t("noResults.description") : t("emptyState.description")}
      ruleSheet={{
        open: isCreateSheetOpen,
        onOpenChange: setIsCreateSheetOpen,
        title: t("form.addTitle"),
        description: t("form.headerDescription"),
        closeLabel: t("actions.close"),
        content: <RateLimitRuleForm embedded onCancel={() => setIsCreateSheetOpen(false)} onSaved={async () => { setIsCreateSheetOpen(false); await load(); }} />,
      }}
      deleteDialog={{
        open: pendingDeleteRule !== null,
        onOpenChange: (open) => !open && setPendingDeleteRule(null),
        title: t("deleteDialog.title"),
        description: t("deleteDialog.description", { ruleId: pendingDeleteRule?.rule_id ?? "-" }),
        cancelLabel: t("deleteDialog.cancel"),
        confirmLabel: t("deleteDialog.confirm"),
        pendingLabel: t("deleteDialog.deleting"),
        pending: isDeleting,
        onConfirm: confirmDeleteRule,
      }}
      upgradeDialog={{
        open: isUpgradeDialogOpen,
        onOpenChange: setIsUpgradeDialogOpen,
        title: t("upgradeDialog.title"),
        description: t("upgradeDialog.description"),
        cancelLabel: t("upgradeDialog.cancel"),
        actionLabel: t("upgradeDialog.action"),
      }}
    >
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
                {items.map((rule) => (
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
    </GatewayPolicyRulesList>
  );
}
