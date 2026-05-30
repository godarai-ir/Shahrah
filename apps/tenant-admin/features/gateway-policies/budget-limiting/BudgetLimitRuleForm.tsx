"use client";

import { useEffect, useMemo, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { Plus, Trash2, X } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
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

export function BudgetLimitRuleForm({ initialRule }: { initialRule?: BudgetLimitRule }) {
  const t = useTranslations("budgetLimitRules");
  const router = useRouter();
  const [isSaving, setIsSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [subjectPickerSelection, setSubjectPickerSelection] = useState<Record<number, PrincipalEntitySelection | null>>({});
  const [subjectDisplayNameByRow, setSubjectDisplayNameByRow] = useState<Record<number, Record<string, string>>>({});
  const [modelPickerSelection, setModelPickerSelection] = useState<Record<number, ModelSelection | null>>({});
  const [providerAllowedModels, setProviderAllowedModels] = useState<ModelSelection[]>([]);
  const [virtualModels, setVirtualModels] = useState<ModelSelection[]>([]);

  const defaultValues = useMemo<BudgetLimitRuleFormInput>(
    () => ({
      rule_id: initialRule?.rule_id ?? "",
      name: initialRule?.name ?? "",
      description: initialRule?.description ?? "",
      enabled: initialRule?.enabled ?? true,
      priority: initialRule?.priority ?? 100,
      conditions: initialRule?.conditions ?? { combinator: "AND", conditions: [] },
      limit_to: initialRule?.limit_to ?? 100,
      unit: initialRule?.unit ?? "cost_per_month",
      budget_applies_per: initialRule?.budget_applies_per ?? undefined,
      block_on_budget_exceed: initialRule?.block_on_budget_exceed ?? true,
      alerts: initialRule?.alerts ?? { enabled: false, thresholds: [75, 90, 100], notification_target: [] },
    }),
    [initialRule],
  );

  const form = useForm<BudgetLimitRuleFormInput>({ resolver: zodResolver(budgetLimitRuleSchema), mode: "onChange", defaultValues });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "conditions.conditions" });
  const watchedConditions = form.watch("conditions.conditions");
  const metadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => form.watch(`conditions.conditions.${x.i}.field`) === "metadata").map((x) => x.i);
  const nonMetadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => form.watch(`conditions.conditions.${x.i}.field`) !== "metadata").map((x) => x.i);
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
      router.push("/gateway-policies/budget-limiting-rules");
      router.refresh();
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.saveError") });
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {initialRule ? t("form.titleEdit") : t("form.titleCreate")}
          {initialRule ? <Badge variant="secondary">{initialRule.rule_id}</Badge> : null}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
          <section className="space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("form.sections.ruleIdentity")}</h2>
            <div className="space-y-2">
              <Label htmlFor="rule_id">{t("form.fields.ruleId")}</Label>
              <Input id="rule_id" placeholder={t("form.placeholders.ruleId")} {...form.register("rule_id")} />
              <FieldError message={form.formState.errors.rule_id?.message} />
            </div>
          </section>

          <Separator />
          <section className="space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("filters.title")}</h2>
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
                <Plus className="h-4 w-4" />
                {t("filters.add")}
              </Button>
            </div>

            <div className="space-y-3">
              {nonMetadataIndexes.map((index) => {
                const field = fields[index];
                const rowField = form.watch(`conditions.conditions.${index}.field`);
                const usedByOthers = getUsedFields(index);
                const subjectTokens = ((form.watch(`conditions.conditions.${index}.values`) as Array<string | undefined> | undefined) ?? []).filter(
                  (token): token is string => typeof token === "string" && token.trim().length > 0,
                );

                return (
                  <div key={field.id} className="rounded-lg border p-3">
                    {rowField === "subject" ? (
                      <div className="space-y-2">
                        <div className="grid gap-2 md:grid-cols-[1.2fr_2fr_auto_auto]">
                          <select className="h-10 rounded-md border px-2" {...form.register(`conditions.conditions.${index}.field`)}>
                            {fieldOptions
                              .filter((option) => option === rowField || !usedByOthers.has(option))
                              .map((option) => (
                                <option key={option} value={option}>
                                  {fieldLabel[option]}
                                </option>
                              ))}
                          </select>
                          <PrincipalEntitySelect
                            id={`subject-selector-${index}`}
                            selected={subjectPickerSelection[index] ?? null}
                            onSelectedChange={(value) => setSubjectPickerSelection((prev) => ({ ...prev, [index]: value }))}
                            onSearch={searchPrincipals}
                            labels={principalLabels}
                            kinds={["user", "team", "org_unit", "virtual_account"]}
                          />
                          <Button
                            type="button"
                            variant="outline"
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
                            {t("actions.add")}
                          </Button>
                          <Button type="button" variant="outline" onClick={() => remove(index)}>
                            <Trash2 className="h-4 w-4" />
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
                                  className="text-muted-foreground hover:text-foreground"
                                  onClick={() => {
                                    form.setValue(
                                      `conditions.conditions.${index}.values`,
                                      subjectTokens.filter((value) => value !== token),
                                      { shouldValidate: true },
                                    );
                                  }}
                                >
                                  <X className="h-3 w-3" />
                                </button>
                              </span>
                            );
                          })}
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-2">
                        <div className="grid gap-2 md:grid-cols-[1.2fr_2fr_auto_auto]">
                          <select className="h-10 rounded-md border px-2" {...form.register(`conditions.conditions.${index}.field`)}>
                            {fieldOptions
                              .filter((option) => option === rowField || !usedByOthers.has(option))
                              .map((option) => (
                                <option key={option} value={option}>
                                  {fieldLabel[option]}
                                </option>
                              ))}
                          </select>
                          <ModelEntitySelect
                            id={`model-selector-${index}`}
                            selected={modelPickerSelection[index] ?? null}
                            onSelectedChange={(value) => setModelPickerSelection((prev) => ({ ...prev, [index]: value }))}
                            providerAllowedModels={providerAllowedModels}
                            virtualModels={virtualModels}
                            labels={modelLabels}
                          />
                          <Button
                            type="button"
                            variant="outline"
                            onClick={() => {
                              const selected = modelPickerSelection[index];
                              if (!selected) return;
                              const modelValues = ((form.watch(`conditions.conditions.${index}.values`) as string[] | undefined) ?? []).filter(Boolean);
                              if (modelValues.includes(selected.name)) return;
                              form.setValue(`conditions.conditions.${index}.values`, [...modelValues, selected.name], { shouldValidate: true });
                            }}
                          >
                            {t("actions.add")}
                          </Button>
                          <Button type="button" variant="outline" onClick={() => remove(index)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {(((form.watch(`conditions.conditions.${index}.values`) as string[] | undefined) ?? []).filter((item) => item.trim().length > 0)).map((token) => (
                            <span key={token} className="inline-flex items-center gap-1 rounded-md border bg-muted px-2 py-1 text-xs">
                              <span>{token}</span>
                              <button
                                type="button"
                                className="text-muted-foreground hover:text-foreground"
                                onClick={() => {
                                  const modelValues = ((form.watch(`conditions.conditions.${index}.values`) as string[] | undefined) ?? []).filter(Boolean);
                                  form.setValue(
                                    `conditions.conditions.${index}.values`,
                                    modelValues.filter((value) => value !== token),
                                    { shouldValidate: true },
                                  );
                                }}
                              >
                                <X className="h-3 w-3" />
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
                <div className="rounded-lg border p-3">
                  <div className="mb-2 text-sm font-medium">{t("filters.metadataTitle")}</div>
                  <div className="space-y-2">
                    {metadataIndexes.map((index) => (
                      <div key={fields[index]?.id ?? `metadata-${index}`} className="grid gap-2 md:grid-cols-[1fr_1fr_1fr_auto_auto]">
                        <Input placeholder={t("form.placeholders.metadataKey")} {...form.register(`conditions.conditions.${index}.metadataKey`)} />
                        <select className="h-10 rounded-md border px-2" {...form.register(`conditions.conditions.${index}.operator`)}>
                          {metadataOperatorOptions.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                        <Input placeholder={t("form.placeholders.metadataValue")} {...form.register(`conditions.conditions.${index}.value` as const)} />
                        <Button type="button" variant="outline" onClick={() => append({ field: "metadata", operator: "equals", values: undefined })}>
                          <Plus className="h-4 w-4" />
                        </Button>
                        <Button type="button" variant="outline" onClick={() => remove(index)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          </section>

          <Separator />
          <section className="space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("form.sections.budget")}</h2>
            <div className="grid gap-3 md:grid-cols-3">
              <div className="space-y-2">
                <Label>{t("form.fields.limitTo")}</Label>
                <Input type="number" min={0.000001} step={0.01} {...form.register("limit_to", { valueAsNumber: true })} />
                <FieldError message={form.formState.errors.limit_to?.message} />
              </div>
              <div className="space-y-2">
                <Label>{t("form.fields.unit")}</Label>
                <select className="h-10 w-full rounded-md border px-2" {...form.register("unit")}>
                  <option value="cost_per_day">{t("units.cost_per_day")}</option>
                  <option value="cost_per_week">{t("units.cost_per_week")}</option>
                  <option value="cost_per_month">{t("units.cost_per_month")}</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>{t("form.fields.budgetAppliesPer")}</Label>
                <select className="h-10 w-full rounded-md border px-2" {...form.register("budget_applies_per")}>
                  <option value="">{t("form.options.applyPer.shared")}</option>
                  <option value="user">{t("form.options.applyPer.user")}</option>
                  <option value="model">{t("form.options.applyPer.model")}</option>
                  <option value="virtualaccount">{t("form.options.applyPer.virtualAccount")}</option>
                  <option value="metadata.project_id">{t("form.options.applyPer.metadataProjectId")}</option>
                </select>
              </div>
            </div>
          </section>

          <Separator />
          <section className="space-y-3 rounded-lg border p-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">{t("form.blocking.title")}</div>
                <p className="text-xs text-muted-foreground">{t("form.blocking.description")}</p>
              </div>
              <input
                type="checkbox"
                className="h-5 w-5"
                checked={form.watch("block_on_budget_exceed")}
                onChange={(event) => form.setValue("block_on_budget_exceed", event.target.checked)}
              />
            </div>
          </section>

          <section className="space-y-3 rounded-lg border p-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">{t("form.alerts.title")}</div>
                <p className="text-xs text-muted-foreground">{t("form.alerts.description")}</p>
              </div>
              <input
                type="checkbox"
                className="h-5 w-5"
                checked={alertEnabled}
                onChange={(event) => form.setValue("alerts.enabled", event.target.checked)}
              />
            </div>
            {alertEnabled ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="space-y-2">
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
                <div className="space-y-2">
                  <Label>{t("form.fields.channelType")}</Label>
                  <select
                    className="h-10 w-full rounded-md border px-2"
                    onChange={(event) => form.setValue("alerts.notification_target", [{ type: event.target.value as "email" | "slack-webhook" | "slack-bot" }])}
                    defaultValue={(form.getValues("alerts.notification_target")?.[0]?.type as string | undefined) ?? "email"}
                  >
                    <option value="email">{t("form.options.channel.email")}</option>
                    <option value="slack-webhook">{t("form.options.channel.slackWebhook")}</option>
                    <option value="slack-bot">{t("form.options.channel.slackBot")}</option>
                  </select>
                </div>
              </div>
            ) : null}
          </section>

          <div className="sticky bottom-0 flex justify-end gap-2 border-t bg-background pt-3">
            <Button type="button" variant="outline" onClick={() => router.push("/gateway-policies/budget-limiting-rules")}>
              {t("actions.cancel")}
            </Button>
            <Button type="submit" disabled={!form.formState.isValid || isSaving}>
              {isSaving ? t("actions.saving") : t("actions.save")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return <p className="text-xs text-destructive">{message}</p>;
}
