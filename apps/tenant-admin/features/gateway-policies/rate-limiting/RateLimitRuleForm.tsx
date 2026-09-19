"use client";

import { useMemo, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale, useTranslations } from "next-intl";
import { ArrowDownToLine, Ban, ChevronDown, ChevronUp, Gauge, OctagonX } from "lucide-react";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { FormSection } from "@/components/app-ui/FormSection";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { StatusOptionGroup } from "@/components/app-ui/StatusOptionGroup";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { GatewayPolicyRuleSheetFooter, RuleOverlapNote, RuleSentencePreview, RuleTargetsEditor } from "@/features/gateway-policies/components";
import { findOverlappingRules, type PolicyRule } from "@/features/gateway-policies/policy-rules";
import { buildRuleConditions, readRuleTargets } from "@/features/gateway-policies/rule-targets";
import { resolveRuleErrorMessage } from "@/features/gateway-policies/rule-error-message";
import { getLocaleDirection } from "@/lib/locale-direction";

import { createRateLimitRule, updateRateLimitRule } from "./ruleApi";
import type { RateLimitRule } from "./ruleTypes";
import { rateLimitRuleSchema, type RateLimitRuleFormInput } from "./ruleValidation";

const WINDOW_UNITS = ["minute", "hour", "day", "week", "month", "second"] as const;

type RateLimitRuleFormProps = {
  initialRule?: RateLimitRule;
  /** Every loaded rule, for the overlap warning. */
  rules: PolicyRule[];
  onCancel: () => void;
  onSaved: () => void | Promise<void>;
};

export function RateLimitRuleForm({ initialRule, rules, onCancel, onSaved }: RateLimitRuleFormProps) {
  const t = useTranslations("rateLimitRules");
  const tp = useTranslations("gatewayPolicies");
  const locale = useLocale();
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "error"; message: string } | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [now] = useState(() => new Date());
  const storedTargets = useMemo(() => readRuleTargets(initialRule?.conditions?.conditions), [initialRule]);

  const defaultValues = useMemo<RateLimitRuleFormInput>(
    () => ({
      rule_id: initialRule?.rule_id ?? "",
      name: initialRule?.name ?? "",
      description: initialRule?.description ?? "",
      enabled: initialRule?.enabled ?? true,
      priority: initialRule?.priority ?? 100,
      // Re-writing through the targets model upgrades legacy `kind:id` subjects to the shape the gateway reads.
      conditions: { combinator: "AND", conditions: buildRuleConditions(readRuleTargets(initialRule?.conditions?.conditions)) },
      limit: initialRule?.limit ?? { dimension: "requests", max: 60, window: { value: 1, unit: "minute" } },
      conflict_behavior: initialRule?.conflict_behavior ?? "most_restrictive",
      response_behavior: "return_429",
      soft_limit: initialRule?.soft_limit ?? false,
      retry_after_header: initialRule?.retry_after_header ?? true,
      effective_from: initialRule?.effective_from ?? "",
      expires_at: initialRule?.expires_at ?? "",
    }),
    [initialRule],
  );

  const form = useForm<RateLimitRuleFormInput>({
    resolver: zodResolver(rateLimitRuleSchema as never),
    mode: "onChange",
    defaultValues,
  });
  const values = useWatch({ control: form.control }) as RateLimitRuleFormInput;
  const conditions = values.conditions?.conditions;
  const errors = form.formState.errors;

  const draft: PolicyRule = { type: "rate", rule: { ...(values as RateLimitRule), conditions: { combinator: "AND", conditions: conditions ?? [] } } };
  const overlapping = values.enabled === false
    ? []
    : findOverlappingRules("rate", readRuleTargets(conditions), rules, now, initialRule ? `rate:${initialRule.rule_id}` : undefined);

  function set<K extends Parameters<typeof form.setValue>[0]>(name: K, value: Parameters<typeof form.setValue<K>>[1]) {
    form.setValue(name, value, { shouldValidate: true, shouldDirty: true });
  }

  async function onSubmit(submitted: RateLimitRuleFormInput) {
    setIsSaving(true);
    setNotice(null);
    const parsed = rateLimitRuleSchema.parse(submitted);
    try {
      if (initialRule) {
        await updateRateLimitRule(initialRule.rule_id, parsed);
      } else {
        await createRateLimitRule(parsed);
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
        <h2 className="text-base font-semibold md:text-lg">{initialRule ? tp("editor.editRateTitle") : tp("editor.newRateTitle")}</h2>
        <p className="text-sm leading-6 text-muted-foreground">{tp("editor.rateDescription")}</p>
      </div>

      <RuleSentencePreview rule={draft} />

      <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

      <form id="rate-limit-rule-form" className="flex flex-col gap-6" onSubmit={form.handleSubmit(onSubmit)}>
        <FormSection title={tp("editor.sections.identity")} className="md:grid-cols-1">
          <div className="grid gap-4 sm:grid-cols-2">
            <DisplayNameField id="rate-rule-name" label={t("form.fields.name")} placeholder={t("form.placeholders.name")} error={errors.name?.message} {...form.register("name")} />
            <IdentifierField id="rate-rule-id" label={t("form.fields.ruleId")} error={errors.rule_id?.message} {...form.register("rule_id")} />
          </div>
          <div className="flex flex-col gap-2">
            <FieldLabel htmlFor="rate-rule-description">{t("form.fields.description")}</FieldLabel>
            <Textarea id="rate-rule-description" placeholder={t("form.placeholders.description")} {...form.register("description")} />
          </div>
          <AppSwitchField
            id="rate-rule-enabled"
            label={t("form.switches.enabled.title")}
            description={t("form.switches.enabled.description")}
            checked={Boolean(values.enabled)}
            onCheckedChange={(checked) => set("enabled", checked)}
          />
        </FormSection>

        <FormSection title={tp("editor.sections.targets")} description={tp("editor.sections.targetsDescription")} className="md:grid-cols-1">
          <RuleTargetsEditor value={conditions} onChange={(next) => set("conditions.conditions", next)} disabled={isSaving} storedTargets={storedTargets} />
          <RuleOverlapNote overlapping={overlapping} variant="editor" />
        </FormSection>

        <FormSection title={tp("editor.sections.limit")} className="md:grid-cols-1">
          <fieldset className="flex flex-col gap-2">
            <legend className="sr-only">{tp("editor.sections.limit")}</legend>
            <div className="flex flex-wrap items-center gap-2 text-sm">
              <span>{tp("editor.limitSentence.prefix")}</span>
              <Input
                type="number"
                dir="ltr"
                min={1}
                className="w-28 tabular-nums"
                aria-label={t("form.fields.limitMax")}
                aria-invalid={Boolean(errors.limit?.max) || undefined}
                {...form.register("limit.max", { valueAsNumber: true })}
              />
              <AppSelectField
                ariaLabel={t("aria.limitDimension")}
                className="w-auto"
                triggerClassName="min-w-28 rounded-lg"
                value={values.limit?.dimension ?? "requests"}
                onValueChange={(value) => set("limit.dimension", value as "requests" | "tokens")}
                options={[
                  { value: "requests", label: tp("sentence.dimensions.requests") },
                  { value: "tokens", label: tp("sentence.dimensions.tokens") },
                ]}
              />
              <span>{tp("editor.limitSentence.per")}</span>
              <Input
                type="number"
                dir="ltr"
                min={1}
                step={1}
                className="w-20 tabular-nums"
                aria-label={t("form.fields.windowValue")}
                aria-invalid={Boolean(errors.limit?.window?.value) || undefined}
                {...form.register("limit.window.value", { valueAsNumber: true })}
              />
              <AppSelectField
                ariaLabel={t("aria.windowUnit")}
                className="w-auto"
                triggerClassName="min-w-28 rounded-lg"
                value={values.limit?.window?.unit ?? "minute"}
                onValueChange={(value) => set("limit.window.unit", value as (typeof WINDOW_UNITS)[number])}
                options={WINDOW_UNITS.map((unit) => ({ value: unit, label: tp(`sentence.units.${unit}`) }))}
              />
            </div>
            <FieldError>{errors.limit?.max?.message ?? errors.limit?.window?.value?.message}</FieldError>
          </fieldset>

          <StatusOptionGroup
            label={tp("editor.enforcement.label")}
            value={values.soft_limit ? "soft" : "hard"}
            columns={2}
            dir={getLocaleDirection(locale)}
            disabled={isSaving}
            onValueChange={(value) => set("soft_limit", value === "soft")}
            options={[
              { value: "hard", title: tp("editor.enforcement.hard.title"), description: tp("editor.enforcement.hard.description"), badge: "429", icon: OctagonX, tone: "success" },
              { value: "soft", title: tp("editor.enforcement.soft.title"), description: tp("editor.enforcement.soft.description"), badge: tp("editor.enforcement.soft.badge"), icon: Gauge, tone: "warning" },
            ]}
          />
        </FormSection>

        <FormSection title={tp("editor.sections.evaluation")} description={tp("editor.sections.evaluationDescription")} className="md:grid-cols-1">
          <AppNumberField
            id="rate-rule-priority"
            label={t("form.fields.priority")}
            description={t("form.descriptions.priority")}
            min={0}
            step={1}
            error={errors.priority?.message}
            {...form.register("priority", { valueAsNumber: true })}
          />
          <StatusOptionGroup
            label={tp("editor.afterMatch.label")}
            // apply_all enforces exactly like most_restrictive in the gateway; it keeps its stored value until changed.
            value={values.conflict_behavior === "first_match" ? "first_match" : "continue"}
            columns={2}
            dir={getLocaleDirection(locale)}
            disabled={isSaving}
            onValueChange={(value) => set("conflict_behavior", value === "first_match" ? "first_match" : "most_restrictive")}
            options={[
              { value: "continue", title: tp("editor.afterMatch.continue.title"), description: tp("editor.afterMatch.continue.description"), badge: tp("editor.afterMatch.continue.badge"), icon: ArrowDownToLine, tone: "success" },
              { value: "first_match", title: tp("editor.afterMatch.stop.title"), description: tp("editor.afterMatch.stop.description"), badge: tp("editor.afterMatch.stop.badge"), icon: Ban, tone: "info" },
            ]}
          />
        </FormSection>

        <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen}>
          <FormSection
            title={t("form.sections.advanced")}
            className="md:grid-cols-1"
            actions={(
              <CollapsibleTrigger asChild>
                <Button type="button" variant="ghost" size="sm">
                  <span>{advancedOpen ? t("form.advanced.toggleClose") : t("form.advanced.toggleOpen")}</span>
                  {advancedOpen ? <ChevronUp data-icon="inline-end" /> : <ChevronDown data-icon="inline-end" />}
                </Button>
              </CollapsibleTrigger>
            )}
          >
            <CollapsibleContent>
              <AppSwitchField
                id="rate-rule-retry-after"
                label={t("form.switches.retryAfterHeader.title")}
                description={t("form.switches.retryAfterHeader.description")}
                checked={Boolean(values.retry_after_header)}
                disabled={Boolean(values.soft_limit)}
                onCheckedChange={(checked) => set("retry_after_header", checked)}
              />
            </CollapsibleContent>
          </FormSection>
        </Collapsible>

        <GatewayPolicyRuleSheetFooter>
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={onCancel}>{t("actions.cancel")}</Button>}
            primary={<Button type="submit" form="rate-limit-rule-form" disabled={!form.formState.isValid || isSaving}>{isSaving ? t("actions.saving") : t("actions.save")}</Button>}
          />
        </GatewayPolicyRuleSheetFooter>
      </form>
    </div>
  );
}
