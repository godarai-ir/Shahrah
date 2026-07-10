"use client";

import { type ReactNode, useEffect, useMemo, useState } from "react";
import { useFieldArray, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { ChevronDown, ChevronUp, Plus, Trash2, X } from "lucide-react";

import { FormSection, IdentifierField, PageHeader } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
  type PrincipalSearchOption,
} from "@/app/components/principal-search-input";
import { ModelEntitySelect, type ModelEntitySelectLabels, type ModelSelection } from "@/app/components/model-entity-select";

import { createRateLimitRule, updateRateLimitRule } from "./ruleApi";
import type { RateLimitCondition, RateLimitRule } from "./ruleTypes";
import { rateLimitRuleSchema, type RateLimitRuleFormInput } from "./ruleValidation";

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
      .filter((item) => {
        const label = `${item.display_name ?? ""} ${item.alias}`.toLowerCase();
        return label.includes(q);
      })
      .slice(0, 20)
      .map((item) => ({ principal_id: item.id, display_name: item.display_name || item.alias, secondary: item.alias, type: "org_unit" }));
  }

  if (kind === "virtual_account") {
    const payload = await requestJSON<{ data?: Array<{ id: string; alias: string; display_name?: string }> }>("/api/gateway/lookups/virtual-accounts");
    const q = query.trim().toLowerCase();
    return (payload.data ?? [])
      .filter((item) => {
        const label = `${item.display_name ?? ""} ${item.alias}`.toLowerCase();
        return label.includes(q);
      })
      .slice(0, 20)
      .map((item) => ({ principal_id: item.id, display_name: item.display_name || item.alias, secondary: item.alias, type: "virtual_account" }));
  }

  return [];
}

export function RateLimitRuleForm({ initialRule }: { initialRule?: RateLimitRule }) {
  const t = useTranslations("rateLimitRules");
  const router = useRouter();
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [subjectPickerSelection, setSubjectPickerSelection] = useState<Record<number, PrincipalEntitySelection | null>>({});
  const [subjectDisplayNameByRow, setSubjectDisplayNameByRow] = useState<Record<number, Record<string, string>>>({});
  const [modelPickerSelection, setModelPickerSelection] = useState<Record<number, ModelSelection | null>>({});
  const [providerAllowedModels, setProviderAllowedModels] = useState<ModelSelection[]>([]);
  const [virtualModels, setVirtualModels] = useState<ModelSelection[]>([]);

  const defaultValues = useMemo<RateLimitRuleFormInput>(
    () => ({
      rule_id: initialRule?.rule_id ?? "",
      name: initialRule?.name ?? "",
      description: initialRule?.description ?? "",
      enabled: initialRule?.enabled ?? true,
      priority: initialRule?.priority ?? 100,
      conditions: initialRule?.conditions ?? { combinator: "AND", conditions: [] },
      limit: initialRule?.limit ?? { dimension: "requests", max: 1, window: { value: 1, unit: "hour" } },
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
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "conditions.conditions" });
  const watchedConditions = useWatch({ control: form.control, name: "conditions.conditions" });
  const selectedWindowUnit = useWatch({ control: form.control, name: "limit.window.unit" });
  const enabled = useWatch({ control: form.control, name: "enabled" });
  const limitDimension = useWatch({ control: form.control, name: "limit.dimension" });
  const conflictBehavior = useWatch({ control: form.control, name: "conflict_behavior" });
  const softLimit = useWatch({ control: form.control, name: "soft_limit" });
  const retryAfterHeader = useWatch({ control: form.control, name: "retry_after_header" });
  const fieldRows = fields.map((field, index) => ({
    field,
    index,
    rowField: watchedConditions?.[index]?.field ?? "subject",
  }));
  const metadataIndexes = fieldRows.filter((item) => item.rowField === "metadata").map((item) => item.index);
  const nonMetadataIndexes = fieldRows.filter((item) => item.rowField !== "metadata").map((item) => item.index);
  const fieldLabel: Record<(typeof fieldOptions)[number], string> = {
    subject: t("filters.fields.subject"),
    model: t("filters.fields.model"),
    metadata: t("filters.fields.metadata"),
  };
  const metadataOperatorLabel: Record<(typeof metadataOperatorOptions)[number], string> = {
    exists: t("filters.operators.exists"),
    not_exists: t("filters.operators.not_exists"),
    equals: t("filters.operators.equals"),
    not_equals: t("filters.operators.not_equals"),
    contains: t("filters.operators.contains"),
    starts_with: t("filters.operators.starts_with"),
    ends_with: t("filters.operators.ends_with"),
    gt: t("filters.operators.gt"),
    gte: t("filters.operators.gte"),
    lt: t("filters.operators.lt"),
    lte: t("filters.operators.lte"),
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
        const nextAllowed = (modelsPayload.data ?? [])
          .map((item) => {
            const name = item.model ?? item.name ?? item.id ?? "";
            return {
              id: name,
              name,
              kind: "provider_allowed" as const,
              secondary: item.provider_name ?? item.provider,
            };
          })
          .filter((item) => item.name);
        setProviderAllowedModels(nextAllowed);
      } catch {}

      try {
        const virtualPayload = await requestJSON<{ data?: Array<{ id?: string; alias?: string }> }>("/api/virtual-models");
        const nextVirtual = (virtualPayload.data ?? [])
          .map((item) => {
            const id = item.id ?? item.alias ?? "";
            const name = item.alias ?? item.id ?? "";
            return { id, name, kind: "virtual_model" as const };
          })
          .filter((item) => item.id && item.name);
        setVirtualModels(nextVirtual);
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

  function normalizeCondition(condition: RateLimitCondition): RateLimitCondition {
    if (condition.field === "metadata") return condition;
    const joined = condition.values?.[0] ?? "";
    const values = joined.split(",").map((item) => item.trim()).filter(Boolean);
    return { ...condition, values, operator: "in" };
  }

  async function onSubmit(values: RateLimitRuleFormInput) {
    setIsSaving(true);
    setNotice(null);
    const parsed = rateLimitRuleSchema.parse(values);
    const normalized = {
      ...parsed,
      conditions: {
        ...parsed.conditions,
        conditions: parsed.conditions.conditions.map((condition) => normalizeCondition(condition as RateLimitCondition) as never),
      },
    };
    try {
      if (initialRule) {
        await updateRateLimitRule(initialRule.rule_id, normalized);
      } else {
        await createRateLimitRule(normalized);
      }
      setNotice({ kind: "success", message: t("messages.saveSuccess") });
      router.push("/gateway-policies/rate-limiting-rules");
      router.refresh();
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.saveError") });
    } finally {
      setIsSaving(false);
    }
  }

  const title = initialRule ? t("form.editTitle") : t("form.addTitle");

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 md:px-6">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={title}
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
        <FormSection title={t("form.sections.identity")} description={t("form.descriptions.identity")}>
          <div className="grid gap-4 lg:grid-cols-2">
            <IdentifierField id="rule_id" label={t("form.fields.ruleId")} description={t("form.descriptions.ruleId")} placeholder={t("form.ruleIdPlaceholder")} error={form.formState.errors.rule_id?.message} {...form.register("rule_id")} />

            <FormField label={t("form.fields.name")} htmlFor="name" description={t("form.descriptions.name")}>
              <Input id="name" placeholder={t("form.placeholders.name")} {...form.register("name")} />
            </FormField>
          </div>

          <FormField label={t("form.fields.description")} htmlFor="description" description={t("form.descriptions.description")}>
            <Textarea id="description" placeholder={t("form.placeholders.description")} {...form.register("description")} />
          </FormField>

          <SwitchRow
            id="enabled"
            title={t("form.switches.enabled.title")}
            description={t("form.switches.enabled.description")}
            checked={Boolean(enabled)}
            onCheckedChange={(checked) => form.setValue("enabled", checked, { shouldValidate: true, shouldDirty: true })}
          />
        </FormSection>

        <FormSection title={t("filters.title")} description={t("filters.description")}>
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <p className="text-sm leading-6 text-muted-foreground">
                {t("filters.helper")}
              </p>
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
                {t("actions.addFilter")}
              </Button>
            </div>

            {fields.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border bg-muted/20 px-4 py-5 text-sm leading-6 text-muted-foreground">
                {t("filters.empty")}
              </div>
            ) : null}

            {nonMetadataIndexes.map((index) => {
              const field = fields[index];
              const rowField = watchedConditions?.[index]?.field ?? "subject";
              const usedByOthers = getUsedFields(index);
              const rowValues = watchedConditions?.[index]?.values;
              const subjectTokens = ((rowValues as Array<string | undefined> | undefined) ?? []).filter(
                (token): token is string => typeof token === "string" && token.trim().length > 0,
              );
              const modelValues = (((rowValues as string[] | undefined) ?? []).filter((item) => typeof item === "string" && item.trim().length > 0));

              return (
                <div key={field.id} className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
                  <div className="grid gap-3 xl:grid-cols-[14rem_minmax(0,1fr)_auto_auto]">
                    <InlineSelect
                      value={rowField}
                      ariaLabel={t("aria.filterFieldType")}
                      onValueChange={(value) =>
                        form.setValue(`conditions.conditions.${index}.field`, value as (typeof fieldOptions)[number], {
                          shouldValidate: true,
                          shouldDirty: true,
                        })
                      }
                      options={fieldOptions
                        .filter((option) => option === rowField || !usedByOthers.has(option))
                        .map((option) => ({ value: option, label: fieldLabel[option] }))}
                    />

                    {rowField === "subject" ? (
                      <PrincipalEntitySelect
                        id={`subject-selector-${index}`}
                        selected={subjectPickerSelection[index] ?? null}
                        onSelectedChange={(value) => setSubjectPickerSelection((prev) => ({ ...prev, [index]: value }))}
                        onSearch={searchPrincipals}
                        labels={principalLabels}
                        kinds={["user", "team", "org_unit", "virtual_account"]}
                      />
                    ) : (
                      <ModelEntitySelect
                        id={`model-selector-${index}`}
                        selected={modelPickerSelection[index] ?? null}
                        onSelectedChange={(value) => setModelPickerSelection((prev) => ({ ...prev, [index]: value }))}
                        providerAllowedModels={providerAllowedModels}
                        virtualModels={virtualModels}
                        labels={modelLabels}
                      />
                    )}

                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        if (rowField === "subject") {
                          const selected = subjectPickerSelection[index];
                          if (!selected) return;
                          const token = subjectTokenFromSelection(selected);
                          if (subjectTokens.includes(token)) return;
                          form.setValue(`conditions.conditions.${index}.values`, [...subjectTokens, token], { shouldValidate: true, shouldDirty: true });
                          setSubjectDisplayNameByRow((prev) => ({
                            ...prev,
                            [index]: {
                              ...(prev[index] ?? {}),
                              [token]: selected.display_name || selected.principal_id || token,
                            },
                          }));
                          return;
                        }

                        const selected = modelPickerSelection[index];
                        if (!selected || modelValues.includes(selected.name)) return;
                        form.setValue(`conditions.conditions.${index}.values`, [...modelValues, selected.name], { shouldValidate: true, shouldDirty: true });
                      }}
                    >
                      <Plus data-icon="inline-start" />
                      {t("actions.add")}
                    </Button>

                    <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label={t("aria.removeFilter")}>
                      <Trash2 />
                    </Button>
                  </div>

                  <div className="flex flex-wrap gap-2">
                    {rowField === "subject"
                      ? subjectTokens.map((token) => {
                          const item = toSubjectSelection(token);
                          const displayName = subjectDisplayNameByRow[index]?.[token] ?? item.display_name ?? item.principal_id ?? token;
                          return (
                            <SelectionChip
                              key={token}
                              label={`${item.type}: ${displayName}`}
                              removeLabel={t("aria.removeSelection", { label: `${item.type}: ${displayName}` })}
                              onRemove={() => {
                                form.setValue(
                                  `conditions.conditions.${index}.values`,
                                  subjectTokens.filter((value) => value !== token),
                                  { shouldValidate: true, shouldDirty: true },
                                );
                                setSubjectDisplayNameByRow((prev) => {
                                  const nextRow = { ...(prev[index] ?? {}) };
                                  delete nextRow[token];
                                  return { ...prev, [index]: nextRow };
                                });
                              }}
                            />
                          );
                        })
                      : modelValues.map((token) => (
                          <SelectionChip
                            key={token}
                            label={token}
                            removeLabel={t("aria.removeSelection", { label: token })}
                            onRemove={() => {
                              form.setValue(
                                `conditions.conditions.${index}.values`,
                                modelValues.filter((value) => value !== token),
                                { shouldValidate: true, shouldDirty: true },
                              );
                            }}
                          />
                        ))}
                  </div>

                  <FieldError message={form.formState.errors.conditions?.conditions?.[index]?.values?.message} />
                </div>
              );
            })}

            {metadataIndexes.length > 0 ? (
              <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4">
                <div className="flex flex-col gap-1">
                  <h3 className="text-base font-semibold text-card-foreground">{t("filters.metadataTitle")}</h3>
                  <p className="text-sm leading-6 text-muted-foreground">{t("filters.metadataDescription")}</p>
                </div>

                <div className="flex flex-col gap-3">
                  {metadataIndexes.map((index, itemIndex) => {
                    const operatorError = form.formState.errors.conditions?.conditions?.[index]?.operator?.message;
                    const keyError = form.formState.errors.conditions?.conditions?.[index]?.metadataKey?.message;
                    const valueError = form.formState.errors.conditions?.conditions?.[index]?.value?.message;

                    return (
                      <div key={fields[index]?.id ?? `metadata-${index}`} className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4">
                        <div className="grid gap-3 xl:grid-cols-[minmax(0,1.2fr)_14rem_minmax(0,1.2fr)_auto_auto]">
                          <Input
                            placeholder={t("form.placeholders.metadataKey")}
                            aria-invalid={Boolean(keyError)}
                            {...form.register(`conditions.conditions.${index}.metadataKey`)}
                          />

                          <InlineSelect
                            value={watchedConditions?.[index]?.operator ?? "equals"}
                            ariaLabel={t("aria.metadataOperator")}
                            onValueChange={(value) =>
                              form.setValue(`conditions.conditions.${index}.operator`, value as (typeof metadataOperatorOptions)[number], {
                                shouldValidate: true,
                                shouldDirty: true,
                              })
                            }
                            options={metadataOperatorOptions.map((option) => ({ value: option, label: metadataOperatorLabel[option] }))}
                          />

                          <Input
                            placeholder={t("form.placeholders.metadataValue")}
                            aria-invalid={Boolean(valueError)}
                            {...form.register(`conditions.conditions.${index}.value` as const)}
                          />

                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            aria-label={t("aria.addMetadataRow")}
                            onClick={() => append({ field: "metadata", operator: "equals", values: undefined })}
                          >
                            <Plus />
                          </Button>

                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            onClick={() => remove(index)}
                            aria-label={t("aria.removeMetadataRow", { index: itemIndex + 1 })}
                          >
                            <Trash2 />
                          </Button>
                        </div>

                        <div className="grid gap-2 md:grid-cols-3">
                          <FieldError message={keyError} />
                          <FieldError message={operatorError} />
                          <FieldError message={valueError} />
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ) : null}
          </div>
        </FormSection>

        <FormSection title={t("form.sections.limit")} description={t("form.descriptions.limit")}>
          <div className="grid gap-4 lg:grid-cols-4">
            <FormField label={t("form.fields.limitMax")} htmlFor="limit-max">
              <Input
                id="limit-max"
                type="number"
                min={1}
                aria-invalid={Boolean(form.formState.errors.limit?.max)}
                {...form.register("limit.max", { valueAsNumber: true })}
              />
              <FieldError message={form.formState.errors.limit?.max?.message} />
            </FormField>

            <FormField label={t("form.fields.limitDimension")}>
              <InlineSelect
                value={limitDimension ?? "requests"}
                ariaLabel={t("aria.limitDimension")}
                onValueChange={(value) => form.setValue("limit.dimension", value as "requests" | "tokens", { shouldValidate: true, shouldDirty: true })}
                options={[
                  { value: "requests", label: t("form.options.dimension.requests") },
                  { value: "tokens", label: t("form.options.dimension.tokens") },
                ]}
              />
            </FormField>

            <FormField label={t("form.fields.windowUnit")}>
              <InlineSelect
                value={selectedWindowUnit}
                ariaLabel={t("aria.windowUnit")}
                onValueChange={(value) =>
                  form.setValue("limit.window.unit", value as "second" | "minute" | "hour" | "day" | "week" | "month", {
                    shouldValidate: true,
                    shouldDirty: true,
                  })
                }
                options={[
                  { value: "minute", label: t("form.options.windowUnit.minute") },
                  { value: "hour", label: t("form.options.windowUnit.hour") },
                  { value: "day", label: t("form.options.windowUnit.day") },
                  { value: "week", label: t("form.options.windowUnit.week") },
                  { value: "month", label: t("form.options.windowUnit.month") },
                  { value: "second", label: t("form.options.windowUnit.second") },
                ]}
              />
            </FormField>

            <FormField
              label={t("form.fields.windowValue")}
              htmlFor="limit-window-value"
              description={selectedWindowUnit === "second" ? t("form.descriptions.windowSeconds") : undefined}
            >
              <Input
                id="limit-window-value"
                type="number"
                min={1}
                step={1}
                placeholder={selectedWindowUnit === "second" ? t("form.placeholders.windowCustom") : t("form.placeholders.windowDefault")}
                aria-invalid={Boolean(form.formState.errors.limit?.window?.value)}
                {...form.register("limit.window.value", { valueAsNumber: true })}
              />
              <FieldError message={form.formState.errors.limit?.window?.value?.message} />
            </FormField>
          </div>
        </FormSection>

        <FormSection title={t("form.sections.advanced")} description={t("form.descriptions.advanced")}>
          <div className="flex flex-col gap-4">
            <Button
              type="button"
              variant="outline"
              className="w-full justify-between md:w-fit"
              onClick={() => setAdvancedOpen((prev) => !prev)}
            >
              <span>{advancedOpen ? t("form.advanced.toggleClose") : t("form.advanced.toggleOpen")}</span>
              {advancedOpen ? <ChevronUp data-icon="inline-end" /> : <ChevronDown data-icon="inline-end" />}
            </Button>

            {advancedOpen ? (
              <div className="flex flex-col gap-4 rounded-xl border border-border bg-muted/20 p-4">
                <div className="grid gap-4 lg:grid-cols-2">
                  <FormField label={t("form.fields.priority")} htmlFor="priority" description={t("form.descriptions.priority")}>
                    <Input id="priority" type="number" min={0} {...form.register("priority", { valueAsNumber: true })} />
                  </FormField>

                  <FormField label={t("form.fields.conflictBehavior")}>
                    <InlineSelect
                      value={conflictBehavior ?? "most_restrictive"}
                      ariaLabel={t("aria.conflictBehavior")}
                      onValueChange={(value) =>
                        form.setValue("conflict_behavior", value as "first_match" | "most_restrictive" | "apply_all", {
                          shouldValidate: true,
                          shouldDirty: true,
                        })
                      }
                      options={[
                        { value: "first_match", label: t("form.options.conflictBehavior.first_match") },
                        { value: "most_restrictive", label: t("form.options.conflictBehavior.most_restrictive") },
                        { value: "apply_all", label: t("form.options.conflictBehavior.apply_all") },
                      ]}
                    />
                  </FormField>
                </div>

                <div className="grid gap-4 lg:grid-cols-2">
                  <SwitchRow
                    id="soft-limit"
                    title={t("form.switches.softLimit.title")}
                    description={t("form.switches.softLimit.description")}
                    checked={Boolean(softLimit)}
                    onCheckedChange={(checked) => form.setValue("soft_limit", checked, { shouldValidate: true, shouldDirty: true })}
                  />

                  <SwitchRow
                    id="retry-after-header"
                    title={t("form.switches.retryAfterHeader.title")}
                    description={t("form.switches.retryAfterHeader.description")}
                    checked={Boolean(retryAfterHeader)}
                    onCheckedChange={(checked) => form.setValue("retry_after_header", checked, { shouldValidate: true, shouldDirty: true })}
                  />
                </div>
              </div>
            ) : null}
          </div>
        </FormSection>

        <div className="sticky bottom-0 z-10 rounded-xl border border-border bg-background/95 p-4 shadow-sm backdrop-blur supports-[backdrop-filter]:bg-background/90">
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <p className="text-sm leading-6 text-muted-foreground">
              {t("form.stickyDescription")}
            </p>
            <div className="flex flex-wrap items-center justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => router.push("/gateway-policies/rate-limiting-rules")}>
                {t("actions.cancel")}
              </Button>
              <Button type="submit" disabled={!form.formState.isValid || isSaving}>
                {isSaving ? t("actions.saving") : t("actions.save")}
              </Button>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}

function FormField({
  label,
  htmlFor,
  description,
  children,
}: {
  label: string;
  htmlFor?: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <div className="flex min-w-0 flex-col gap-2">
      <Label htmlFor={htmlFor}>{label}</Label>
      {children}
      {description ? <p className="text-xs leading-5 text-muted-foreground">{description}</p> : null}
    </div>
  );
}

function InlineSelect({
  value,
  onValueChange,
  options,
  ariaLabel,
}: {
  value: string;
  onValueChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
  ariaLabel: string;
}) {
  return (
    <Select dir="rtl" value={value} onValueChange={onValueChange}>
      <SelectTrigger aria-label={ariaLabel}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent dir="rtl">
        <SelectGroup>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}

function SelectionChip({
  label,
  onRemove,
  removeLabel,
}: {
  label: string;
  onRemove: () => void;
  removeLabel: string;
}) {
  return (
    <Badge variant="outline" className="gap-1 rounded-md px-2 py-1 text-xs font-normal">
      <span className="truncate">{label}</span>
      <button
        type="button"
        className="rounded-sm text-muted-foreground transition-colors hover:text-foreground"
        onClick={onRemove}
        aria-label={removeLabel}
      >
        <X className="size-3" />
      </button>
    </Badge>
  );
}

function SwitchRow({
  id,
  title,
  description,
  checked,
  onCheckedChange,
}: {
  id: string;
  title: string;
  description: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  const descriptionId = `${id}-description`;

  return (
    <div className={cn("flex items-start justify-between gap-4 rounded-xl border border-border bg-card px-4 py-3")}>
      <div className="flex min-w-0 flex-col gap-1">
        <Label htmlFor={id} className="text-sm font-medium text-card-foreground">
          {title}
        </Label>
        <p id={descriptionId} className="text-xs leading-5 text-muted-foreground">
          {description}
        </p>
      </div>
      <Switch id={id} checked={checked} aria-describedby={descriptionId} onCheckedChange={onCheckedChange} />
    </div>
  );
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return <p className="text-xs leading-5 text-destructive">{message}</p>;
}
