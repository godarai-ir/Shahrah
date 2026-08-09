"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { CircleDollarSign, Copy, Pencil, Power, Trash2 } from "lucide-react";

import { CopyAction } from "@/components/app-ui/CopyAction";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { GatewayPolicyRulesList } from "@/features/gateway-policies/components";
import { formatDateTime as formatSharedDateTime } from "@/lib/date-time";

import {
  createBudgetLimitRule,
  deleteBudgetLimitRule,
  getBudgetLimitEntitlements,
  listBudgetLimitRules,
  updateBudgetLimitRuleStatus,
} from "./ruleApi";
import type { BudgetLimitRule } from "./ruleTypes";
import { BudgetLimitRuleForm } from "./BudgetLimitRuleForm";

const FEATURE_BUDGET_LIMIT_RULES_ENABLED = "budget_limit_rules.enabled";
const IRR_NUMBER_FORMAT = new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 0 });
const BUDGET_RULE_COLUMN_CLASS = {
  ruleId: "w-56",
  limit: "w-36",
  unit: "hidden w-32 @md:table-cell",
  status: "w-24 whitespace-nowrap",
  updatedAt: "hidden w-48 whitespace-nowrap @lg:table-cell",
  actions: "w-12 text-end",
} as const;

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

function formatJalaliDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return formatSharedDateTime(date, { locale: "fa", format: { year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", second: "2-digit" } });
}

function unitKey(unit: BudgetLimitRule["unit"]) {
  if (unit === "cost_per_day") return "cost_per_day";
  if (unit === "cost_per_week") return "cost_per_week";
  return "cost_per_month";
}

function formatIRR(value: number) {
  return `${IRR_NUMBER_FORMAT.format(value)} ریال`;
}

function readBooleanFeature(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") {
    if (value.trim().toLowerCase() === "true") return true;
    if (value.trim().toLowerCase() === "false") return false;
  }
  return fallback;
}

export function BudgetLimitRulesPage() {
  const t = useTranslations("budgetLimitRules");
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [items, setItems] = useState<BudgetLimitRule[]>([]);
  const [pendingDeleteRule, setPendingDeleteRule] = useState<BudgetLimitRule | null>(null);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);
  const [isRuleSheetOpen, setIsRuleSheetOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<BudgetLimitRule | null>(null);
  const [entitlements, setEntitlements] = useState<EffectiveEntitlements>(null);
  const [hasLoadedEntitlements, setHasLoadedEntitlements] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const response = await listBudgetLimitRules(query);
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
        const response = await getBudgetLimitEntitlements();
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

  const isBudgetLimitFeatureEnabled = hasLoadedEntitlements
    ? readBooleanFeature(entitlements?.features?.[FEATURE_BUDGET_LIMIT_RULES_ENABLED], false)
    : null;

  async function confirmDeleteRule() {
    if (!pendingDeleteRule || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteBudgetLimitRule(pendingDeleteRule.rule_id);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteSuccess") });
      setPendingDeleteRule(null);
      await load();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.deleteError") });
      throw error;
    } finally {
      setIsDeleting(false);
    }
  }

  function handleCreateRule() {
    if (isBudgetLimitFeatureEnabled === false) {
      setIsUpgradeDialogOpen(true);
      return;
    }

    setEditingRule(null);
    setIsRuleSheetOpen(true);
  }

  function handleEditRule(rule: BudgetLimitRule) {
    setEditingRule(rule);
    setIsRuleSheetOpen(true);
  }

  function handleRuleSheetOpenChange(open: boolean) {
    setIsRuleSheetOpen(open);
    if (!open) setEditingRule(null);
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
      itemCount={items.length}
      loading={isLoading}
      loadingLabel={t("messages.loading")}
      emptyIcon={CircleDollarSign}
      emptyTitle={query ? t("noResults.title") : t("emptyState.title")}
      emptyDescription={query ? t("noResults.description") : t("emptyState.description")}
      ruleSheet={{
        open: isRuleSheetOpen,
        onOpenChange: handleRuleSheetOpenChange,
        title: editingRule ? t("form.titleEdit") : t("form.titleCreate"),
        description: t("form.headerDescription"),
        closeLabel: t("actions.close"),
        contentKey: editingRule?.id ?? "create",
        content: (
          <BudgetLimitRuleForm
            embedded
            initialRule={editingRule ?? undefined}
            onCancel={() => handleRuleSheetOpenChange(false)}
            onSaved={async () => {
              handleRuleSheetOpenChange(false);
              await load();
            }}
          />
        ),
      }}
      deleteDialog={{
        open: pendingDeleteRule !== null,
        onOpenChange: (open) => {
          if (!open) {
            setPendingDeleteRule(null);
            setDeleteNotice(null);
          }
        },
        title: t("deleteDialog.title"),
        description: t("deleteDialog.description", { ruleId: pendingDeleteRule?.rule_id ?? "-" }),
        cancelLabel: t("deleteDialog.cancel"),
        confirmLabel: t("deleteDialog.confirm"),
        pendingLabel: t("deleteDialog.deleting"),
        pending: isDeleting,
        onConfirm: confirmDeleteRule,
        feedback: <OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />,
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
      <Table className="w-max min-w-full table-fixed text-xs lg:text-sm">
              <TableHeader>
                <TableRow>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.ruleId}>{t("table.ruleId")}</TableHead>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.limit}>{t("table.limit")}</TableHead>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.unit}>{t("table.unit")}</TableHead>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.status}>{t("table.status")}</TableHead>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.updatedAt}>{t("table.updatedAt")}</TableHead>
                  <TableHead className={BUDGET_RULE_COLUMN_CLASS.actions}>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className={`${BUDGET_RULE_COLUMN_CLASS.ruleId} min-w-0 font-mono font-medium`} dir="ltr">
                      <div className="flex min-w-0 items-center gap-1">
                        <span className="block min-w-0 flex-1 truncate" title={rule.rule_id}>{rule.rule_id}</span>
                        <CopyAction
                          className="shrink-0"
                          value={rule.rule_id}
                          copyLabel={t("actions.copyRuleId", { ruleId: rule.rule_id })}
                          copiedLabel={t("actions.copied")}
                        />
                      </div>
                    </TableCell>
                    <TableCell className={BUDGET_RULE_COLUMN_CLASS.limit}>
                      <span className="block truncate" title={formatIRR(rule.limit_to)}>{formatIRR(rule.limit_to)}</span>
                    </TableCell>
                    <TableCell className={BUDGET_RULE_COLUMN_CLASS.unit}>{t(`units.${unitKey(rule.unit)}`)}</TableCell>
                    <TableCell className={BUDGET_RULE_COLUMN_CLASS.status}>
                      <StatusBadge status={rule.enabled ? "active" : "inactive"}>
                        {rule.enabled ? t("status.enabled") : t("status.disabled")}
                      </StatusBadge>
                    </TableCell>
                    <TableCell className={BUDGET_RULE_COLUMN_CLASS.updatedAt}>{formatJalaliDateTime(rule.updated_at)}</TableCell>
                    <TableCell className={BUDGET_RULE_COLUMN_CLASS.actions}>
                      <RowActionsMenu
                        label={t("table.rowActions", { ruleId: rule.rule_id })}
                        actions={[
                          { label: t("actions.edit"), icon: <Pencil />, onSelect: () => handleEditRule(rule) },
                          { label: rule.enabled ? t("actions.disable") : t("actions.enable"), icon: <Power />, onSelect: () => void updateBudgetLimitRuleStatus(rule.rule_id, !rule.enabled).then(load) },
                          { label: t("actions.duplicate"), icon: <Copy />, onSelect: () => void createBudgetLimitRule({ ...rule, rule_id: `${rule.rule_id}-copy`, name: `${rule.name ?? rule.rule_id} ${t("actions.copySuffix")}` } as never).then(load) },
                          { label: t("actions.delete"), icon: <Trash2 />, destructive: true, onSelect: () => setPendingDeleteRule(rule) },
                        ]}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
      </Table>
    </GatewayPolicyRulesList>
  );
}

