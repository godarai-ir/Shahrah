"use client";

import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { ChipInput } from "@/components/app-ui/ChipInput";
import { FormSection } from "@/components/app-ui/FormSection";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { FieldDescription, FieldLabel } from "@/components/ui/field";
import { GatewayPolicyRuleSheetFooter, RuleOverlapNote, RuleSentencePreview, RuleTargetsEditor } from "@/features/gateway-policies/components";
import { findOverlappingRules, type PolicyRule } from "@/features/gateway-policies/policy-rules";
import { buildRuleConditions, readRuleTargets } from "@/features/gateway-policies/rule-targets";
import { resolveRuleErrorMessage } from "@/features/gateway-policies/rule-error-message";

import { createBudgetLimitRule, updateBudgetLimitRule } from "./ruleApi";
import type { BudgetLimitRule } from "./ruleTypes";
import { budgetLimitRuleSchema, DEFAULT_ALERT_THRESHOLDS, type BudgetLimitRuleFormInput } from "./ruleValidation";

const ON_EXCEED_VALUES = ["block", "alert"] as const;

/** Splits typed thresholds on commas or spaces, folds Persian digits, and keeps whole percentages. */
function parseThresholds(raw: string): string[] {
  const folded = raw.replace(/[۰-۹]/g, (digit) => String(digit.charCodeAt(0) - 0x06f0)).replace(/[٠-٩]/g, (digit) => String(digit.charCodeAt(0) - 0x0660));
  return folded.split(/[,،\s]+/).map((item) => item.trim().replace(/%$/, "")).filter((item) => /^\d{1,4}$/.test(item) && Number(item) >= 1 && Number(item) <= 1000);
}

/** The exceed behavior of a stored rule; rules saved before on_exceed fall back to the blocking flag. */
function storedOnExceed(rule?: BudgetLimitRule): "block" | "alert" {
  if (rule?.on_exceed) return rule.on_exceed;
  return rule?.block_on_budget_exceed === false ? "alert" : "block";
}

type BudgetLimitRuleFormProps = {
  initialRule?: BudgetLimitRule;
  /** Every loaded rule, for the overlap warning. */
  rules: PolicyRule[];
  onCancel: () => void;
  onSaved: () => void | Promise<void>;
};

export function BudgetLimitRuleForm({ initialRule, rules, onCancel, onSaved }: BudgetLimitRuleFormProps) {
  const t = useTranslations("budgetLimitRules");
  const tp = useTranslations("gatewayPolicies");
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [now] = useState(() => new Date());
  const storedTargets = useMemo(() => readRuleTargets(initialRule?.conditions?.conditions), [initialRule]);

  const defaultValues = useMemo<Partial<BudgetLimitRuleFormInput>>(
    () => ({
      rule_id: initialRule?.rule_id ?? "",
      name: initialRule?.name ?? "",
      description: initialRule?.description ?? "",
      enabled: initialRule?.enabled ?? true,
      priority: initialRule?.priority ?? 100,
      // Re-writing through the targets model upgrades legacy `kind:id` subjects to the shape the gateway reads.
      conditions: { combinator: "AND", conditions: buildRuleConditions(readRuleTargets(initialRule?.conditions?.conditions)) },
      limit: initialRule?.limit ?? "",
      unit: initialRule?.unit ?? "cost_per_month",
      budget_applies_per: initialRule?.budget_applies_per ?? undefined,
      on_exceed: storedOnExceed(initialRule),
      alert_thresholds: initialRule?.alert_thresholds?.length ? initialRule.alert_thresholds : [...DEFAULT_ALERT_THRESHOLDS],
    }),
    [initialRule],
  );

  const form = useForm<BudgetLimitRuleFormInput>({
    resolver: zodResolver(budgetLimitRuleSchema as never),
    mode: "onChange",
    defaultValues,
  });
  // React Hook Form's watch API is intentionally reactive and cannot be memoized by React Compiler.
  // eslint-disable-next-line react-hooks/incompatible-library
  const watched = form.watch();
  const conditions = watched.conditions?.conditions;
  const thresholds = watched.alert_thresholds ?? [];
  const draft: PolicyRule = {
    type: "budget",
    rule: { ...(watched as BudgetLimitRule), conditions: { combinator: "AND", conditions: conditions ?? [] } },
  };
  const overlapping = watched.enabled === false
    ? []
    : findOverlappingRules("budget", readRuleTargets(conditions), rules, now, initialRule ? `budget:${initialRule.rule_id}` : undefined);

  async function onSubmit(values: BudgetLimitRuleFormInput) {
    setIsSaving(true);
    setNotice(null);
    const parsed = budgetLimitRuleSchema.parse(values);
    try {
      if (initialRule) {
        await updateBudgetLimitRule(initialRule.rule_id, parsed);
      } else {
        await createBudgetLimitRule(parsed);
      }
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.saveSuccess") });
      await onSaved();
    } catch (error) {
      setNotice({ kind: "error", message: resolveRuleErrorMessage(error, (code) => t(`apiErrors.${code}`), t("messages.saveError")) });
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex flex-col gap-1">
        <h2 className="text-base font-semibold md:text-lg">{initialRule ? tp("editor.editBudgetTitle") : tp("editor.newBudgetTitle")}</h2>
        <p className="text-sm leading-6 text-muted-foreground">{tp("editor.budgetDescription")}</p>
      </div>

      <RuleSentencePreview rule={draft} />

      <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

        <form id="budget-limit-rule-form" className="flex flex-col gap-6" onSubmit={form.handleSubmit(onSubmit)}>
          <FormSection
            title={t("form.sections.ruleIdentity")}
            description={t("form.descriptions.ruleIdentity")}
            className="md:grid-cols-1"
          >
            <IdentifierField id="rule_id" label={t("form.fields.ruleId")} placeholder={t("form.placeholders.ruleId")} error={form.formState.errors.rule_id?.message} {...form.register("rule_id")} />
          </FormSection>

          <FormSection title={tp("editor.sections.targets")} description={tp("editor.sections.targetsDescription")} className="md:grid-cols-1">
            <RuleTargetsEditor
              value={conditions}
              onChange={(next) => form.setValue("conditions.conditions", next, { shouldValidate: true, shouldDirty: true })}
              disabled={isSaving}
              storedTargets={storedTargets}
            />
            <RuleOverlapNote overlapping={overlapping} variant="editor" />
          </FormSection>

          <FormSection
            title={t("form.sections.budget")}
            description={t("form.descriptions.budget")}
            className="md:grid-cols-1"
          >
            <div className="grid gap-4 lg:grid-cols-3">
              <AppNumberField id="limit" label={t("form.fields.limit")} min={0.01} step={0.01} inputMode="decimal" description={t("form.descriptions.limit")} error={form.formState.errors.limit?.message} fieldClassName="min-w-0" {...form.register("limit")} />
              <AppSelectField
                className="min-w-0"
                triggerClassName="h-10 min-w-0 rounded-lg"
                label={t("form.fields.unit")}
                value={watched.unit}
                onValueChange={(value) => form.setValue("unit", value as BudgetLimitRuleFormInput["unit"], { shouldValidate: true, shouldDirty: true })}
                options={(["cost_per_day", "cost_per_week", "cost_per_month"] as const).map((value) => ({ value, label: t(`units.${value}`) }))}
              />
              <AppSelectField
                className="min-w-0"
                triggerClassName="h-10 min-w-0 rounded-lg"
                label={t("form.fields.budgetAppliesPer")}
                value={watched.budget_applies_per ?? "shared"}
                onValueChange={(value) => form.setValue("budget_applies_per", value === "shared" ? undefined : value as BudgetLimitRuleFormInput["budget_applies_per"], { shouldValidate: true, shouldDirty: true })}
                options={[
                  { value: "shared", label: t("form.options.applyPer.shared") },
                  { value: "tenant", label: t("form.options.applyPer.tenant") },
                  { value: "user", label: t("form.options.applyPer.user") },
                  { value: "team", label: t("form.options.applyPer.team") },
                  { value: "model", label: t("form.options.applyPer.model") },
                  { value: "virtualaccount", label: t("form.options.applyPer.virtualAccount") },
                  { value: "metadata.project_id", label: t("form.options.applyPer.metadataProjectId") },
                ]}
              />
            </div>
          </FormSection>

          <FormSection
            title={t("form.sections.behavior")}
            description={t("form.descriptions.behavior")}
            className="md:grid-cols-1"
          >
            <OptionCardGroup
              aria-label={t("form.onExceed.label")}
              density="compact"
              value={watched.on_exceed ?? "block"}
              onValueChange={(value) => form.setValue("on_exceed", value, { shouldValidate: true, shouldDirty: true })}
              disabled={isSaving}
              options={ON_EXCEED_VALUES.map((value) => ({
                value,
                title: t(`form.onExceed.${value}.title`),
                description: t(`form.onExceed.${value}.description`),
              }))}
            />
            <div className="flex flex-col gap-2">
              <FieldLabel htmlFor="alert-thresholds">{t("form.fields.thresholds")}</FieldLabel>
              <ChipInput
                id="alert-thresholds"
                value={thresholds.map(String)}
                onValueChange={(values) => form.setValue("alert_thresholds", [...new Set(values.map(Number))].sort((x, y) => x - y), { shouldValidate: true, shouldDirty: true })}
                parseValues={parseThresholds}
                placeholder={t("form.placeholders.thresholds")}
                getRemoveLabel={(value) => t("form.thresholdRemove", { value })}
                disabled={isSaving}
              />
              <FieldDescription>{t("form.descriptions.thresholds")}</FieldDescription>
            </div>
          </FormSection>

          <GatewayPolicyRuleSheetFooter>
            <FormSheetActions
              secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={onCancel}>{t("actions.cancel")}</Button>}
              primary={<Button type="submit" form="budget-limit-rule-form" disabled={!form.formState.isValid || isSaving}>{isSaving ? t("actions.saving") : t("actions.save")}</Button>}
            />
          </GatewayPolicyRuleSheetFooter>
        </form>
    </div>
  );
}
