"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ChevronDown, CircleDollarSign, Copy, Gauge, ListOrdered, Pencil, Plus, Power, Trash2 } from "lucide-react";

import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useAppLocale } from "@/hooks/use-locale-format";
import { formatDateTime } from "@/lib/date-time";
import { getLocaleDirection } from "@/lib/locale-direction";
import { matchesSearch } from "@/lib/search-normalize";
import { cn } from "@/lib/utils";

import { BudgetLimitRuleForm } from "./budget-limiting/BudgetLimitRuleForm";
import { createBudgetLimitRule, deleteBudgetLimitRule, listBudgetLimitRules, updateBudgetLimitRuleStatus } from "./budget-limiting/ruleApi";
import { GatewayPolicyRulesList, RuleOverlapNote, useRuleSentence } from "./components";
import {
  compareByEvaluationOrder,
  findOverlappingRules,
  isRuleLive,
  policyRuleKey,
  policyRuleName,
  policyRuleTargets,
  ruleSchedule,
  type PolicyRule,
  type PolicyRuleType,
} from "./policy-rules";
import { RateLimitRuleForm } from "./rate-limiting/RateLimitRuleForm";
import {
  createRateLimitRule,
  deleteRateLimitRule,
  getRateLimitEntitlements,
  listRateLimitRules,
  updateRateLimitRuleStatus,
} from "./rate-limiting/ruleApi";
import { RequestSimulator } from "./RequestSimulator";
import { resolveRuleErrorMessage } from "./rule-error-message";
import { subjectLabel } from "./rule-targets";

const TYPE_FILTERS = ["all", "rate", "budget", "disabled"] as const;
type TypeFilter = (typeof TYPE_FILTERS)[number];

const ENTITLEMENT_FEATURE: Record<PolicyRuleType, string> = {
  rate: "rate_limit_rules.enabled",
  budget: "budget_limit_rules.enabled",
};

function readBooleanFeature(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  return typeof value === "string" && value.trim().toLowerCase() === "true";
}

function isTypeFilter(value: string | null): value is TypeFilter {
  return TYPE_FILTERS.includes(value as TypeFilter);
}

/** `rate:<rule_id>` or `budget:<rule_id>`, as used in ?rule= and ?new=. */
function parseRuleRef(value: string | null): { type: PolicyRuleType; ruleID: string } | null {
  const match = value ? /^(rate|budget):(.+)$/.exec(value) : null;
  return match ? { type: match[1] as PolicyRuleType, ruleID: match[2] } : null;
}

function searchCandidates(item: PolicyRule, sentence: string): string[] {
  const targets = policyRuleTargets(item);
  const subject = targets.subject;
  return [
    item.rule.rule_id,
    item.rule.name ?? "",
    item.rule.description ?? "",
    sentence,
    ...(targets.models ?? []),
    ...(subject ? subject.ids.flatMap((id) => [id, subjectLabel(subject, id)]) : []),
    ...targets.metadata.map((condition) => condition.metadataKey ?? ""),
  ];
}

export function GatewayPolicyRulesPage() {
  const t = useTranslations("gatewayPolicies");
  const tRate = useTranslations("rateLimitRules");
  const tBudget = useTranslations("budgetLimitRules");
  const locale = useLocale();
  const appLocale = useAppLocale();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryString = searchParams.toString();
  const sentence = useRuleSentence();

  const [rules, setRules] = useState<PolicyRule[]>([]);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [missingRuleID, setMissingRuleID] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [entitled, setEntitled] = useState<Record<PolicyRuleType, boolean> | null>(null);
  const [upgradeType, setUpgradeType] = useState<PolicyRuleType | null>(null);
  const [pendingDelete, setPendingDelete] = useState<PolicyRule | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<{ kind: "error"; message: string } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [now] = useState(() => new Date());

  const urlState = useMemo(() => {
    const params = new URLSearchParams(queryString);
    const type = params.get("type");
    const newType = params.get("new");
    return {
      type: isTypeFilter(type) ? type : "all",
      rule: parseRuleRef(params.get("rule")),
      newType: newType === "rate" || newType === "budget" ? newType : null,
    } as const;
  }, [queryString]);

  const patchURL = useCallback(
    (patch: Record<string, string | null>) => {
      const params = new URLSearchParams(queryString);
      for (const [key, value] of Object.entries(patch)) {
        if (value) params.set(key, value);
        else params.delete(key);
      }
      const next = params.toString();
      router.replace(next ? `${pathname}?${next}` : pathname, { scroll: false });
    },
    [pathname, queryString, router],
  );

  const load = useCallback(async () => {
    setIsLoading(true);
    const [rate, budget] = await Promise.allSettled([listRateLimitRules(), listBudgetLimitRules()]);
    const next: PolicyRule[] = [
      ...(rate.status === "fulfilled" ? rate.value.data.map((rule) => ({ type: "rate" as const, rule })) : []),
      ...(budget.status === "fulfilled" ? budget.value.data.map((rule) => ({ type: "budget" as const, rule })) : []),
    ];
    const failure = rate.status === "rejected" ? rate.reason : budget.status === "rejected" ? budget.reason : null;
    setRules(next);
    setLoadError(failure);
    setHasLoaded(true);
    setIsLoading(false);
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  useEffect(() => {
    let active = true;
    void getRateLimitEntitlements()
      .then((response) => {
        if (!active) return;
        setEntitled({
          rate: readBooleanFeature(response?.features?.[ENTITLEMENT_FEATURE.rate]),
          budget: readBooleanFeature(response?.features?.[ENTITLEMENT_FEATURE.budget]),
        });
      })
      .catch(() => {
        if (active) setEntitled({ rate: false, budget: false });
      });
    return () => {
      active = false;
    };
  }, []);

  const sentences = useMemo(() => new Map(rules.map((item) => [policyRuleKey(item), sentence(item)])), [rules, sentence]);

  const visible = useMemo(() => {
    return rules
      .filter((item) => {
        if (urlState.type === "disabled") return !item.rule.enabled;
        return urlState.type === "all" || item.type === urlState.type;
      })
      .filter((item) => !query.trim() || matchesSearch(searchCandidates(item, sentences.get(policyRuleKey(item)) ?? ""), query))
      .sort(compareByEvaluationOrder);
  }, [query, rules, sentences, urlState.type]);
  const visibleRate = visible.filter((item) => item.type === "rate");
  const visibleBudget = visible.filter((item) => item.type === "budget");

  const editing = urlState.rule ? rules.find((item) => item.type === urlState.rule!.type && item.rule.rule_id === urlState.rule!.ruleID) ?? null : null;
  const sheetType = urlState.newType ?? editing?.type ?? null;
  const sheetOpen = Boolean(sheetType);

  useEffect(() => {
    if (hasLoaded && urlState.rule && !editing) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setMissingRuleID(urlState.rule.ruleID);
      patchURL({ rule: null });
    }
  }, [editing, hasLoaded, patchURL, urlState.rule]);

  const notice = loadError
    ? { kind: "error" as const, message: resolveRuleErrorMessage(loadError, (code) => tRate(`apiErrors.${code}`), t("messages.loadError")) }
    : missingRuleID
      ? { kind: "error" as const, message: t("messages.ruleNotFound", { ruleId: missingRuleID }) }
      : null;

  function openRule(item: PolicyRule) {
    setMissingRuleID(null);
    patchURL({ rule: policyRuleKey(item), new: null });
  }

  function closeSheet() {
    patchURL({ rule: null, new: null });
  }

  function startCreate(type: PolicyRuleType) {
    if (entitled && !entitled[type]) {
      setUpgradeType(type);
      return;
    }
    patchURL({ new: type, rule: null });
  }

  async function saved() {
    closeSheet();
    await load();
  }

  async function toggleStatus(item: PolicyRule) {
    try {
      if (item.type === "rate") await updateRateLimitRuleStatus(item.rule.rule_id, !item.rule.enabled);
      else await updateBudgetLimitRuleStatus(item.rule.rule_id, !item.rule.enabled);
      await load();
    } catch (error) {
      showFeedbackToast({ kind: "error", title: tRate("alerts.error"), description: resolveRuleErrorMessage(error, (code) => tRate(`apiErrors.${code}`), t("messages.statusError")) });
    }
  }

  async function duplicate(item: PolicyRule) {
    const rest: Partial<typeof item.rule> = { ...item.rule };
    delete rest.id;
    delete rest.created_at;
    delete rest.updated_at;
    const copy = { ...rest, rule_id: `${item.rule.rule_id}-copy`, name: t("copyName", { name: policyRuleName(item) }), enabled: false };
    try {
      if (item.type === "rate") await createRateLimitRule(copy as Parameters<typeof createRateLimitRule>[0]);
      else await createBudgetLimitRule(copy as Parameters<typeof createBudgetLimitRule>[0]);
      showFeedbackToast({ kind: "success", title: tRate("alerts.success"), description: t("messages.duplicated") });
      await load();
    } catch (error) {
      showFeedbackToast({ kind: "error", title: tRate("alerts.error"), description: resolveRuleErrorMessage(error, (code) => tRate(`apiErrors.${code}`), t("messages.duplicateError")) });
    }
  }

  async function confirmDelete() {
    if (!pendingDelete || isDeleting) return;
    setIsDeleting(true);
    try {
      if (pendingDelete.type === "rate") await deleteRateLimitRule(pendingDelete.rule.rule_id);
      else await deleteBudgetLimitRule(pendingDelete.rule.rule_id);
      showFeedbackToast({ kind: "success", title: tRate("alerts.success"), description: t("messages.deleted") });
      setPendingDelete(null);
      await load();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: resolveRuleErrorMessage(error, (code) => tRate(`apiErrors.${code}`), t("messages.deleteError")) });
      throw error;
    } finally {
      setIsDeleting(false);
    }
  }

  const createAction = (
    <DropdownMenu dir={getLocaleDirection(locale)}>
      <DropdownMenuTrigger asChild>
        <Button type="button" className="w-full sm:w-auto" disabled={!entitled}>
          <Plus data-icon="inline-start" />
          {t("actions.new")}
          <ChevronDown data-icon="inline-end" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onSelect={() => startCreate("rate")}>
          <Gauge />
          {t("types.rate")}
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => startCreate("budget")}>
          <CircleDollarSign />
          {t("types.budget")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  const filtered = Boolean(query.trim()) || urlState.type !== "all";

  function renderRow(item: PolicyRule) {
    const key = policyRuleKey(item);
    const schedule = ruleSchedule(item, now);
    const targets = policyRuleTargets(item);
    const overlapping = isRuleLive(item, now) ? findOverlappingRules(item.type, targets, rules, now, key) : [];
    const name = policyRuleName(item);

    return (
      <li key={key} className={cn("flex items-start gap-3 px-4 py-3", !item.rule.enabled && "opacity-70")}>
        {item.type === "rate" ? (
          <Badge variant="secondary" className="mt-0.5 min-w-10 justify-center font-mono tabular-nums" dir="ltr" title={t("priorityTitle", { priority: item.rule.priority })}>
            {item.rule.priority}
          </Badge>
        ) : (
          <span className="mt-0.5 flex min-w-10 justify-center text-muted-foreground" aria-hidden="true">
            <CircleDollarSign className="size-4" />
          </span>
        )}
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <div className="flex flex-wrap items-center gap-2">
            <button type="button" className="min-w-0 truncate text-start text-sm font-semibold hover:underline" onClick={() => openRule(item)}>
              {name}
            </button>
            {!item.rule.enabled ? <StatusBadge status="inactive">{t("status.disabled")}</StatusBadge> : null}
            {item.rule.enabled && schedule === "scheduled" && item.type === "rate" ? (
              <StatusBadge status="pending">{t("status.scheduled", { date: formatDateTime(item.rule.effective_from!, { locale: appLocale, format: { dateStyle: "medium" } }) })}</StatusBadge>
            ) : null}
            {item.rule.enabled && schedule === "expired" && item.type === "rate" ? (
              <StatusBadge status="neutral">{t("status.expired", { date: formatDateTime(item.rule.expires_at!, { locale: appLocale, format: { dateStyle: "medium" } }) })}</StatusBadge>
            ) : null}
            {item.type === "rate" && item.rule.enabled && schedule === "active" && item.rule.expires_at ? (
              <StatusBadge status="neutral">{t("status.expires", { date: formatDateTime(item.rule.expires_at, { locale: appLocale, format: { dateStyle: "medium" } }) })}</StatusBadge>
            ) : null}
            {item.type === "rate" && item.rule.soft_limit ? <StatusBadge status="warning">{t("status.countOnly")}</StatusBadge> : null}
            {item.type === "rate" && item.rule.conflict_behavior === "first_match" ? <StatusBadge status="neutral">{t("status.stopsEvaluation")}</StatusBadge> : null}
            {item.type === "budget" && !item.rule.block_on_budget_exceed ? <StatusBadge status="neutral">{t("status.noEffect")}</StatusBadge> : null}
          </div>
          <p className="text-sm leading-6 text-foreground">{sentences.get(key)}</p>
          {targets.legacySubject ? <p className="text-xs leading-5 text-destructive">{t("legacySubjectRow")}</p> : null}
          <RuleOverlapNote overlapping={overlapping} variant="row" />
        </div>
        <RowActionsMenu
          label={t("rowActions", { name })}
          actions={[
            { label: t("actions.edit"), icon: <Pencil />, onSelect: () => openRule(item) },
            { label: item.rule.enabled ? t("actions.disable") : t("actions.enable"), icon: <Power />, onSelect: () => void toggleStatus(item) },
            { label: t("actions.duplicate"), icon: <Copy />, onSelect: () => void duplicate(item) },
            { label: t("actions.delete"), icon: <Trash2 />, destructive: true, separatorBefore: true, onSelect: () => { setDeleteNotice(null); setPendingDelete(item); } },
          ]}
        />
      </li>
    );
  }

  const upgradeCopy = upgradeType === "budget" ? tBudget : tRate;

  return (
    <GatewayPolicyRulesList
      title={t("title")}
      createAction={createAction}
      notice={notice}
      alertLabels={{ error: tRate("alerts.error"), success: tRate("alerts.success") }}
      filtered={filtered}
      toolbar={(
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <SearchField
            containerClassName="lg:max-w-sm"
            placeholder={t("searchPlaceholder")}
            aria-label={t("searchPlaceholder")}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          <ToggleGroup
            type="single"
            value={urlState.type}
            onValueChange={(value) => {
              if (isTypeFilter(value)) patchURL({ type: value === "all" ? null : value });
            }}
            aria-label={t("filters.label")}
            className="w-fit rounded-lg border bg-muted/40 p-0.5"
          >
            {TYPE_FILTERS.map((filter) => (
              <ToggleGroupItem
                key={filter}
                value={filter}
                className="h-7 px-2.5 text-xs text-muted-foreground data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-xs"
              >
                {t(`filters.${filter}`)}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
        </div>
      )}
      aside={hasLoaded ? <RequestSimulator rules={rules} onOpenRule={openRule} /> : null}
      itemCount={visible.length}
      loading={isLoading && !hasLoaded}
      loadingLabel={t("messages.loading")}
      emptyIcon={ListOrdered}
      emptyTitle={filtered ? t("noResults.title") : t("emptyState.title")}
      emptyDescription={filtered ? t("noResults.description") : t("emptyState.description")}
      ruleSheet={{
        open: sheetOpen,
        onOpenChange: (open) => {
          if (!open) closeSheet();
        },
        title: sheetType === "budget"
          ? t(editing ? "editor.editBudgetTitle" : "editor.newBudgetTitle")
          : t(editing ? "editor.editRateTitle" : "editor.newRateTitle"),
        description: t("editor.sheetDescription"),
        closeLabel: tRate("actions.close"),
        dir: getLocaleDirection(locale),
        contentKey: urlState.newType ? `new:${urlState.newType}` : editing ? policyRuleKey(editing) : "none",
        content: sheetType === "rate" ? (
          <RateLimitRuleForm initialRule={editing?.type === "rate" ? editing.rule : undefined} rules={rules} onCancel={closeSheet} onSaved={saved} />
        ) : sheetType === "budget" ? (
          <BudgetLimitRuleForm initialRule={editing?.type === "budget" ? editing.rule : undefined} rules={rules} onCancel={closeSheet} onSaved={saved} />
        ) : null,
      }}
      deleteDialog={{
        open: pendingDelete !== null,
        onOpenChange: (open) => {
          if (!open) {
            setPendingDelete(null);
            setDeleteNotice(null);
          }
        },
        title: t("deleteDialog.title", { name: pendingDelete ? policyRuleName(pendingDelete) : "" }),
        description: pendingDelete ? t("deleteDialog.description", { sentence: sentences.get(policyRuleKey(pendingDelete)) ?? "" }) : "",
        cancelLabel: tRate("deleteDialog.cancel"),
        confirmLabel: tRate("deleteDialog.confirm"),
        pendingLabel: tRate("deleteDialog.deleting"),
        pending: isDeleting,
        onConfirm: confirmDelete,
        feedback: <OperationNotice notice={deleteNotice} labels={{ success: tRate("alerts.success"), error: tRate("alerts.error"), warning: tRate("alerts.error"), info: tRate("alerts.success") }} />,
      }}
      upgradeDialog={{
        open: upgradeType !== null,
        onOpenChange: (open) => {
          if (!open) setUpgradeType(null);
        },
        title: upgradeCopy("upgradeDialog.title"),
        description: upgradeCopy("upgradeDialog.description"),
        cancelLabel: upgradeCopy("upgradeDialog.cancel"),
        actionLabel: upgradeCopy("upgradeDialog.action"),
      }}
    >
      {visibleRate.length > 0 ? (
        <section aria-labelledby="gateway-rate-rules" className="overflow-hidden rounded-xl border border-border bg-card">
          <header className="flex flex-col gap-1 border-b border-border bg-muted/30 px-4 py-3">
            <h2 id="gateway-rate-rules" className="text-sm font-semibold">{t("sections.rate", { count: visibleRate.length })}</h2>
            <p className="text-xs leading-5 text-muted-foreground">{t("evaluationNote")}</p>
          </header>
          <ol className="divide-y divide-border">{visibleRate.map(renderRow)}</ol>
        </section>
      ) : null}
      {visibleBudget.length > 0 ? (
        <section aria-labelledby="gateway-budget-rules" className={cn("overflow-hidden rounded-xl border border-border bg-card", visibleRate.length > 0 && "mt-4")}>
          <header className="flex flex-col gap-1 border-b border-border bg-muted/30 px-4 py-3">
            <h2 id="gateway-budget-rules" className="text-sm font-semibold">{t("sections.budget", { count: visibleBudget.length })}</h2>
            <p className="text-xs leading-5 text-muted-foreground">{t("budgetEvaluationNote")}</p>
          </header>
          <ul className="divide-y divide-border">{visibleBudget.map(renderRow)}</ul>
        </section>
      ) : null}
    </GatewayPolicyRulesList>
  );
}
