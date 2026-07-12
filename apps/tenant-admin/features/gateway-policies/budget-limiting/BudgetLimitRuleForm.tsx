"use client";

import { useEffect, useMemo, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { Plus, Trash2, X } from "lucide-react";

import { AppSelectField, FormSection, IdentifierField, PageContainer, PageHeader } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
  type PrincipalSearchOption,
} from "@/app/components/principal-search-input";
import { ModelEntitySelect, type ModelEntitySelectLabels, type ModelSelection } from "@/app/components/model-entity-select";

import { createBudgetLimitRule, updateBudgetLimitRule } from "./ruleApi";
import type { BudgetLimitCondition, BudgetLimitRule } from "./ruleTypes";
import { budgetLimitRuleSchema, type BudgetLimitRuleFormInput } from "./ruleValidation";

const fieldOptions = ["subject", "model", "metadata"] as const;
const metadataOperatorOptions = ["exists", "not_exists", "equals", "not_equals", "contains", "starts_with", "ends_with", "gt", "gte", "lt", "lte"] as const;

function subjectTokenFromSelection(selection: PrincipalEntitySelection) {
  return `${selection.type}:${selection.principal_id ?? selection.display_name}`;
}

function toSubjectSelection(token: string): PrincipalEntitySelection {
  const [typeRaw, ...rest] = token.split(":");
  const principalId = rest.join(":");
  const type = (typeRaw || "user") as PrincipalEntitySelection["type"];
  return {
    type,
    principal_id: principalId,
    display_name: principalId,
  };
}

async function requestJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, { cache: "no-store" });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload?.error?.message ?? payload?.message ?? `Request failed (${response.status})`);
  }
  return payload as T;
}

async function searchPrincipals(kind: PrincipalSearchKind, query: string): Promise<PrincipalSearchOption[]> {
  if (kind === "user" || kind === "team") {
    const payload = await requestJSON<{ data?: Array<{ principal_id: string; display_name: string; secondary?: string; type?: string }> }>(
      `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query)}&limit=20`,
    );
    return (payload.data ?? []).map((item) => ({
      principal_id: item.principal_id,
      display_name: item.display_name,
      secondary: item.secondary,
      type: kind,
    }));
  }
  if (kind === "org_unit") {
    const payload = await requestJSON<{ data?: Array<{ id: string; alias: string; display_name?: string }> }>("/api/gateway/lookups/organization-units");
    const q = query.trim().toLowerCase();
    return (payload.data ?? [])
      .filter((item) => `${item.display_name ?? ""} ${item.alias}`.toLowerCase().includes(q))
      .slice(0, 20)
      .map((item) => ({ principal_id: item.id, display_name: item.display_name || item.alias, secondary: item.alias, type: "org_unit" }));
  }
  if (kind === "virtual_account") {
    const payload = await requestJSON<{ data?: Array<{ id: string; alias: string; display_name?: string }> }>("/api/gateway/lookups/virtual-accounts");
    const q = query.trim().toLowerCase();
    return (payload.data ?? [])
      .filter((item) => `${item.display_name ?? ""} ${item.alias}`.toLowerCase().includes(q))
      .slice(0, 20)
      .map((item) => ({ principal_id: item.id, display_name: item.display_name || item.alias, secondary: item.alias, type: "virtual_account" }));
  }
  return [];
}

type BudgetLimitRuleFormProps = {
  initialRule?: BudgetLimitRule;
  embedded?: boolean;
  onCancel?: () => void;
  onSaved?: () => void | Promise<void>;
};

export function BudgetLimitRuleForm({ initialRule, embedded = false, onCancel, onSaved }: BudgetLimitRuleFormProps) {
  const t = useTranslations("budgetLimitRules");
  const router = useRouter();
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [subjectPickerSelection, setSubjectPickerSelection] = useState<Record<number, PrincipalEntitySelection | null>>({});
  const [subjectDisplayNameByRow, setSubjectDisplayNameByRow] = useState<Record<number, Record<string, string>>>({});
  const [modelPickerSelection, setModelPickerSelection] = useState<Record<number, ModelSelection | null>>({});
  const [providerAllowedModels, setProviderAllowedModels] = useState<ModelSelection[]>([]);
  const [virtualModels, setVirtualModels] = useState<ModelSelection[]>([]);

  const defaultValues = useMemo<Partial<BudgetLimitRuleFormInput>>(
    () => ({
      rule_id: initialRule?.rule_id ?? "",
      name: initialRule?.name ?? "",
      description: initialRule?.description ?? "",
      enabled: initialRule?.enabled ?? true,
      priority: initialRule?.priority ?? 100,
      conditions: initialRule?.conditions ?? { combinator: "AND", conditions: [] },
      limit_to: initialRule?.limit_to,
      unit: initialRule?.unit ?? "cost_per_month",
      budget_applies_per: initialRule?.budget_applies_per ?? undefined,
      block_on_budget_exceed: initialRule?.block_on_budget_exceed ?? true,
      alerts: initialRule?.alerts ?? { enabled: false, thresholds: [75, 90, 100], notification_target: [] },
    }),
    [initialRule],
  );

  const form = useForm<BudgetLimitRuleFormInput>({
    resolver: zodResolver(budgetLimitRuleSchema as never),
    mode: "onChange",
    defaultValues,
  });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "conditions.conditions" });
  // React Hook Form's watch API is intentionally reactive and cannot be memoized by React Compiler.
  // eslint-disable-next-line react-hooks/incompatible-library
  const watchedConditions = form.watch("conditions.conditions");
  const metadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => watchedConditions?.[x.i]?.field === "metadata").map((x) => x.i);
  const nonMetadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => watchedConditions?.[x.i]?.field !== "metadata").map((x) => x.i);
  const alertEnabled = form.watch("alerts.enabled");
  const fieldLabel: Record<(typeof fieldOptions)[number], string> = {
    subject: t("filters.fields.subject"),
    model: t("filters.fields.model"),
    metadata: t("filters.fields.metadata"),
  };
  const principalLabels: PrincipalEntitySelectLabels = {
    users: t("selectors.principal.users"),
    teams: t("selectors.principal.teams"),
    orgUnits: t("selectors.principal.orgUnits"),
    virtualAccounts: t("selectors.principal.virtualAccounts"),
    externalIdentities: t("selectors.principal.externalIdentities"),
    everyone: t("selectors.principal.everyone"),
    searchPlaceholder: t("selectors.principal.searchPlaceholder"),
    triggerPlaceholder: t("selectors.principal.triggerPlaceholder"),
    loading: t("selectors.principal.loading"),
    empty: t("selectors.principal.empty"),
  };
  const modelLabels: ModelEntitySelectLabels = {
    triggerPlaceholder: t("selectors.model.triggerPlaceholder"),
    searchPlaceholder: t("selectors.model.searchPlaceholder"),
    providerAllowed: t("selectors.model.providerAllowed"),
    virtualModels: t("selectors.model.virtualModels"),
    loading: t("selectors.model.loading"),
    empty: t("selectors.model.empty"),
  };

  useEffect(() => {
    void (async () => {
      try {
        const modelsPayload = await requestJSON<{ data?: Array<{ id?: string; model?: string; name?: string; provider_name?: string; provider?: string }> }>("/api/gateway/lookups/models");
        setProviderAllowedModels(
          (modelsPayload.data ?? [])
            .map((item) => ({
              id: item.model ?? item.name ?? item.id ?? "",
              name: item.model ?? item.name ?? item.id ?? "",
              kind: "provider_allowed" as const,
              secondary: item.provider_name ?? item.provider,
            }))
            .filter((item) => item.id && item.name),
        );
      } catch {}
      try {
        const virtualPayload = await requestJSON<{ data?: Array<{ id?: string; alias?: string }> }>("/api/virtual-models");
        setVirtualModels(
          (virtualPayload.data ?? [])
            .map((item) => ({
              id: item.id ?? item.alias ?? "",
              name: item.alias ?? item.id ?? "",
              kind: "virtual_model" as const,
            }))
            .filter((item) => item.id && item.name),
        );
      } catch {}
    })();
  }, []);

  function getUsedFields(excludeIndex?: number) {
    const used = new Set<string>();
    (watchedConditions ?? []).forEach((condition, index) => {
      if (excludeIndex !== undefined && index === excludeIndex) return;
      if (condition?.field && condition.field !== "metadata") used.add(condition.field);
    });
    return used;
  }

  function getNextAvailableField() {
    const used = getUsedFields();
    return fieldOptions.find((option) => option === "metadata" || !used.has(option));
  }

  function normalizeCondition(condition: BudgetLimitCondition): BudgetLimitCondition {
    if (condition.field === "metadata") return condition;
    const joined = condition.values?.[0] ?? "";
    const values = joined
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    return { ...condition, values, operator: "in" };
  }

  async function onSubmit(values: BudgetLimitRuleFormInput) {
    setIsSaving(true);
    setNotice(null);
    const parsed = budgetLimitRuleSchema.parse(values);
    const normalized = {
      ...parsed,
      conditions: {
        ...parsed.conditions,
        conditions: parsed.conditions.conditions.map((c) => normalizeCondition(c as BudgetLimitCondition) as never),
      },
    };
    try {
      if (initialRule) {
        await updateBudgetLimitRule(initialRule.rule_id, normalized);
      } else {
        await createBudgetLimitRule(normalized);
      }
      setNotice({ kind: "success", message: t("messages.saveSuccess") });
      if (onSaved) {
        await onSaved();
        return;
      }
      router.push("/gateway-policies/budget-limiting-rules");
      router.refresh();
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.saveError") });
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <PageContainer size={embedded ? "full" : "wide"}>
      <PageHeader
        eyebrow={t("title")}
        title={initialRule ? t("form.titleEdit") : t("form.titleCreate")}
        description={t("form.headerDescription")}
        actions={initialRule ? <Badge variant="secondary">{initialRule.rule_id}</Badge> : undefined}
      />

        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <form className="flex flex-col gap-6" onSubmit={form.handleSubmit(onSubmit)}>
          <FormSection
            title={t("form.sections.ruleIdentity")}
            description={t("form.descriptions.ruleIdentity")}
            className={embedded ? "md:grid-cols-1" : undefined}
          >
            <IdentifierField id="rule_id" label={t("form.fields.ruleId")} placeholder={t("form.placeholders.ruleId")} error={form.formState.errors.rule_id?.message} {...form.register("rule_id")} />
          </FormSection>

          <FormSection
            title={t("filters.title")}
            description={t("filters.description")}
            className={embedded ? "md:grid-cols-1" : undefined}
            actions={embedded ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={!getNextAvailableField()}
                onClick={() => {
                  const nextField = getNextAvailableField();
                  if (!nextField) return;
                  append({ field: nextField, operator: nextField === "metadata" ? "equals" : "in", values: undefined });
                }}
              >
                <Plus data-icon="inline-start" />
                {t("filters.add")}
              </Button>
            ) : undefined}
          >
            {!embedded ? (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm leading-6 text-muted-foreground">{t("filters.helper")}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={!getNextAvailableField()}
                  onClick={() => {
                    const nextField = getNextAvailableField();
                    if (!nextField) return;
                    append({ field: nextField, operator: nextField === "metadata" ? "equals" : "in", values: undefined });
                  }}
                >
                  <Plus data-icon="inline-start" />
                  {t("filters.add")}
                </Button>
              </div>
            ) : null}

            {fields.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border bg-muted/20 px-4 py-5 text-sm leading-6 text-muted-foreground">
                {t("filters.empty")}
              </div>
            ) : null}

            <div className="flex flex-col gap-3">
              {nonMetadataIndexes.map((index) => {
                const field = fields[index];
                const rowCondition = watchedConditions?.[index];
                const rowField = rowCondition?.field;
                const usedByOthers = getUsedFields(index);
                const subjectTokens = ((rowCondition?.values as Array<string | undefined> | undefined) ?? []).filter(
                  (token): token is string => typeof token === "string" && token.trim().length > 0,
                );
                const modelValues = ((rowCondition?.values as string[] | undefined) ?? []).filter((item) => typeof item === "string" && item.trim().length > 0);

                return (
                  <div key={field.id} className="rounded-xl bg-muted/30 p-3">
                    {rowField === "subject" ? (
                      <div className="flex flex-col gap-3">
                        <div className="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-2 md:grid-cols-[14rem_minmax(0,1fr)_auto_auto]">
                          <AppSelectField
                            ariaLabel={t("filters.fieldLabel")}
                            className="col-span-3 md:col-span-1"
                            triggerClassName="min-w-0 rounded-lg"
                            value={rowField}
                            onValueChange={(value) => form.setValue(`conditions.conditions.${index}.field`, value as BudgetLimitCondition["field"], { shouldDirty: true, shouldTouch: true, shouldValidate: true })}
                            options={fieldOptions
                              .filter((option) => option === rowField || !usedByOthers.has(option))
                              .map((option) => ({ value: option, label: fieldLabel[option] }))}
                          />
                          <PrincipalEntitySelect
                            id={`subject-selector-${index}`}
                            selected={subjectPickerSelection[index] ?? null}
                            onSelectedChange={(value) => setSubjectPickerSelection((prev) => ({ ...prev, [index]: value }))}
                            onSearch={searchPrincipals}
                            labels={principalLabels}
                            kinds={["user", "team", "org_unit", "virtual_account"]}
                            triggerClassName="h-10 min-h-10 rounded-lg"
                          />
                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            aria-label={t("actions.add")}
                            onClick={() => {
                              const selected = subjectPickerSelection[index];
                              if (!selected) return;
                              const token = subjectTokenFromSelection(selected);
                              if (subjectTokens.includes(token)) return;
                              form.setValue(`conditions.conditions.${index}.values`, [...subjectTokens, token], { shouldValidate: true });
                              setSubjectDisplayNameByRow((prev) => ({
                                ...prev,
                                [index]: {
                                  ...(prev[index] ?? {}),
                                  [token]: selected.display_name || selected.principal_id || token,
                                },
                              }));
                            }}
                          >
                            <Plus />
                          </Button>
                          <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label={t("filters.remove")}>
                            <Trash2 />
                          </Button>
                        </div>

                        <div className="flex flex-wrap gap-2">
                          {subjectTokens.map((token) => {
                            const item = toSubjectSelection(token);
                            const displayName = subjectDisplayNameByRow[index]?.[token] ?? item.display_name ?? item.principal_id ?? token;
                            return (
                              <span key={token} className="inline-flex items-center gap-1 rounded-md border bg-muted px-2 py-1 text-xs">
                                <span>
                                  {item.type}: {displayName}
                                </span>
                                <button
                                  type="button"
                                  className="rounded-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                                  aria-label={t("filters.removeSelection", { label: displayName })}
                                  onClick={() => {
                                    form.setValue(
                                      `conditions.conditions.${index}.values`,
                                      subjectTokens.filter((value) => value !== token),
                                      { shouldValidate: true },
                                    );
                                  }}
                                >
                                  <X className="size-3" />
                                </button>
                              </span>
                            );
                          })}
                        </div>
                      </div>
                    ) : (
                      <div className="flex flex-col gap-3">
                        <div className="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-2 md:grid-cols-[14rem_minmax(0,1fr)_auto_auto]">
                          <AppSelectField
                            ariaLabel={t("filters.fieldLabel")}
                            className="col-span-3 md:col-span-1"
                            triggerClassName="min-w-0 rounded-lg"
                            value={rowField}
                            onValueChange={(value) => form.setValue(`conditions.conditions.${index}.field`, value as BudgetLimitCondition["field"], { shouldDirty: true, shouldTouch: true, shouldValidate: true })}
                            options={fieldOptions
                              .filter((option) => option === rowField || !usedByOthers.has(option))
                              .map((option) => ({ value: option, label: fieldLabel[option] }))}
                          />
                          <ModelEntitySelect
                            id={`model-selector-${index}`}
                            selected={modelPickerSelection[index] ?? null}
                            onSelectedChange={(value) => setModelPickerSelection((prev) => ({ ...prev, [index]: value }))}
                            providerAllowedModels={providerAllowedModels}
                            virtualModels={virtualModels}
                            labels={modelLabels}
                            triggerClassName="h-10 rounded-lg"
                          />
                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            aria-label={t("actions.add")}
                            onClick={() => {
                              const selected = modelPickerSelection[index];
                              if (!selected) return;
                              if (modelValues.includes(selected.name)) return;
                              form.setValue(`conditions.conditions.${index}.values`, [...modelValues, selected.name], { shouldValidate: true });
                            }}
                          >
                            <Plus />
                          </Button>
                          <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label={t("filters.remove")}>
                            <Trash2 />
                          </Button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {modelValues.map((token) => (
                            <span key={token} className="inline-flex items-center gap-1 rounded-md border bg-muted px-2 py-1 text-xs">
                              <span>{token}</span>
                              <button
                                type="button"
                                className="rounded-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                                aria-label={t("filters.removeSelection", { label: token })}
                                onClick={() => {
                                  form.setValue(
                                    `conditions.conditions.${index}.values`,
                                    modelValues.filter((value) => value !== token),
                                    { shouldValidate: true },
                                  );
                                }}
                              >
                                <X className="size-3" />
                              </button>
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                );
              })}

              {metadataIndexes.length > 0 ? (
                <div className="rounded-xl bg-muted/30 p-3">
                  <div className="mb-2 text-sm font-medium">{t("filters.metadataTitle")}</div>
                  <div className="flex flex-col gap-2">
                    {metadataIndexes.map((index) => (
                      <div key={fields[index]?.id ?? `metadata-${index}`} className="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-2 md:grid-cols-[1fr_1fr_1fr_auto_auto]">
                        <Input className="col-span-3 md:col-span-1" placeholder={t("form.placeholders.metadataKey")} {...form.register(`conditions.conditions.${index}.metadataKey`)} />
                        <AppSelectField
                          ariaLabel={t("filters.operatorLabel")}
                          className="col-span-3 md:col-span-1"
                          triggerClassName="min-w-0 rounded-lg"
                          value={watchedConditions?.[index]?.operator ?? "equals"}
                          onValueChange={(value) => form.setValue(`conditions.conditions.${index}.operator`, value as BudgetLimitCondition["operator"], { shouldDirty: true, shouldTouch: true, shouldValidate: true })}
                          optionTextDir="ltr"
                          options={metadataOperatorOptions.map((option) => ({ value: option, label: option }))}
                        />
                        <Input placeholder={t("form.placeholders.metadataValue")} {...form.register(`conditions.conditions.${index}.value` as const)} />
                        <Button type="button" variant="outline" size="icon" onClick={() => append({ field: "metadata", operator: "equals", values: undefined })} aria-label={t("filters.addMetadata")}>
                          <Plus />
                        </Button>
                        <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label={t("filters.removeMetadata")}>
                          <Trash2 />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          </FormSection>

          <FormSection
            title={t("form.sections.budget")}
            description={t("form.descriptions.budget")}
            className={embedded ? "md:grid-cols-1" : undefined}
          >
            <div className="grid gap-4 lg:grid-cols-3">
              <Field className="min-w-0" data-invalid={Boolean(form.formState.errors.limit_to)}>
                <FieldLabel htmlFor="limit-to">{t("form.fields.limitTo")}</FieldLabel>
                <Input id="limit-to" type="number" min={1} step={1} inputMode="numeric" dir="ltr" aria-invalid={Boolean(form.formState.errors.limit_to)} {...form.register("limit_to", { valueAsNumber: true })} />
                <FieldError>{form.formState.errors.limit_to?.message}</FieldError>
              </Field>
              <AppSelectField
                className="min-w-0"
                triggerClassName="h-10 min-w-0 rounded-lg"
                label={t("form.fields.unit")}
                value={form.watch("unit")}
                onValueChange={(value) => form.setValue("unit", value as BudgetLimitRuleFormInput["unit"], { shouldValidate: true, shouldDirty: true })}
                options={(["cost_per_day", "cost_per_week", "cost_per_month"] as const).map((value) => ({ value, label: t(`units.${value}`) }))}
              />
              <AppSelectField
                className="min-w-0"
                triggerClassName="h-10 min-w-0 rounded-lg"
                label={t("form.fields.budgetAppliesPer")}
                value={form.watch("budget_applies_per") ?? "shared"}
                onValueChange={(value) => form.setValue("budget_applies_per", value === "shared" ? undefined : value as BudgetLimitRuleFormInput["budget_applies_per"], { shouldValidate: true, shouldDirty: true })}
                options={[
                  { value: "shared", label: t("form.options.applyPer.shared") },
                  { value: "tenant", label: t("form.options.applyPer.tenant") },
                  { value: "user", label: t("form.options.applyPer.user") },
                  { value: "team", label: t("form.options.applyPer.team") },
                  { value: "org_unit", label: t("form.options.applyPer.orgUnit") },
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
            className={embedded ? "md:grid-cols-1" : undefined}
          >
            <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-4 rounded-xl border border-border bg-background p-4">
              <div className="flex flex-col gap-1">
                <div className="font-medium">{t("form.blocking.title")}</div>
                <p className="text-sm leading-6 text-muted-foreground">{t("form.blocking.description")}</p>
              </div>
              <Switch
                aria-label={t("form.blocking.title")}
                checked={form.watch("block_on_budget_exceed")}
                onCheckedChange={(checked) => form.setValue("block_on_budget_exceed", checked, { shouldValidate: true, shouldDirty: true })}
              />
            </div>

            <div className="flex flex-col gap-4 rounded-xl border border-border bg-background p-4">
            <div className="flex items-center justify-between gap-4">
              <div className="flex flex-col gap-1">
                <div className="font-medium">{t("form.alerts.title")}</div>
                <p className="text-sm leading-6 text-muted-foreground">{t("form.alerts.description")}</p>
              </div>
              <Switch
                aria-label={t("form.alerts.title")}
                checked={alertEnabled}
                onCheckedChange={(checked) => form.setValue("alerts.enabled", checked, { shouldValidate: true, shouldDirty: true })}
              />
            </div>
            {alertEnabled ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label>{t("form.fields.thresholds")}</Label>
                  <Input
                    placeholder={t("form.placeholders.thresholds")}
                    defaultValue={(form.getValues("alerts.thresholds") ?? [75, 90, 95, 100]).join(",")}
                    onBlur={(event) => {
                      const values = event.target.value
                        .split(",")
                        .map((v) => Number(v.trim()))
                        .filter((v) => Number.isFinite(v));
                      form.setValue("alerts.thresholds", values);
                    }}
                  />
                </div>
                <AppSelectField
                  label={t("form.fields.channelType")}
                  value={(form.watch("alerts.notification_target")?.[0]?.type as string | undefined) ?? "email"}
                  onValueChange={(value) => form.setValue("alerts.notification_target", [{ type: value as "email" | "slack-webhook" | "slack-bot" }], { shouldDirty: true, shouldTouch: true, shouldValidate: true })}
                  options={[
                    { value: "email", label: t("form.options.channel.email") },
                    { value: "slack-webhook", label: t("form.options.channel.slackWebhook") },
                    { value: "slack-bot", label: t("form.options.channel.slackBot") },
                  ]}
                />
              </div>
            ) : null}
            </div>
            </div>
          </FormSection>

          <div className="sticky bottom-0 rounded-xl border border-border bg-background/95 p-4 shadow-sm backdrop-blur supports-[backdrop-filter]:bg-background/90">
            <div className="flex flex-wrap items-center justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => onCancel ? onCancel() : router.push("/gateway-policies/budget-limiting-rules")}>
                {t("actions.cancel")}
              </Button>
              <Button type="submit" disabled={!form.formState.isValid || isSaving}>
                {isSaving ? t("actions.saving") : t("actions.save")}
              </Button>
            </div>
          </div>
        </form>
    </PageContainer>
  );
}
