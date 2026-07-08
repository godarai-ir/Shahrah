"use client";

import { type ReactNode, useEffect, useMemo, useState } from "react";
import { useFieldArray, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { ChevronDown, ChevronUp, Plus, Trash2, X } from "lucide-react";

import { FormSection, PageHeader } from "@/components/app-ui";
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
const fieldLabel: Record<(typeof fieldOptions)[number], string> = {
  subject: "مبدا درخواست",
  model: "مدل‌های مجاز شده",
  metadata: "فراداده",
};

const metadataOperatorOptions = ["exists", "not_exists", "equals", "not_equals", "contains", "starts_with", "ends_with", "gt", "gte", "lt", "lte"] as const;

const metadataOperatorLabel: Record<(typeof metadataOperatorOptions)[number], string> = {
  exists: "دارای کلید",
  not_exists: "فاقد کلید",
  equals: "برابر با",
  not_equals: "نابرابر با",
  contains: "شامل",
  starts_with: "شروع با",
  ends_with: "پایان با",
  gt: "بزرگ‌تر از",
  gte: "بزرگ‌تر یا برابر",
  lt: "کوچک‌تر از",
  lte: "کوچک‌تر یا برابر",
};

const principalLabels: PrincipalEntitySelectLabels = {
  users: "کاربران",
  teams: "تیم‌ها",
  orgUnits: "واحد‌های سازمانی",
  virtualAccounts: "اکانت‌های مجازی",
  externalIdentities: "هویت خارجی",
  everyone: "همه",
  searchPlaceholder: "جست‌وجو...",
  triggerPlaceholder: "انتخاب مبدا",
  loading: "در حال بارگذاری...",
  empty: "موردی پیدا نشد",
};

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
  const modelLabels: ModelEntitySelectLabels = {
    triggerPlaceholder: "انتخاب مدل",
    searchPlaceholder: "جست‌وجوی مدل...",
    providerAllowed: "مدل‌های مجاز پروایدر",
    virtualModels: "مدل‌های مجازی",
    loading: "در حال بارگذاری...",
    empty: "موردی پیدا نشد",
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
        eyebrow="Gateway Policies"
        title={title}
        description="قانون محدودسازی را با ساختار استاندارد محصول تعریف کنید تا روی سوژه، مدل و فراداده اعمال شود."
        actions={initialRule ? <Badge variant="secondary">{initialRule.rule_id}</Badge> : undefined}
      />

      {notice ? (
        <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{notice.message}</AlertDescription>
        </Alert>
      ) : null}

      <form className="flex flex-col gap-6" onSubmit={form.handleSubmit(onSubmit)}>
        <FormSection title="هویت قانون" description="شناسه، نام نمایشی و وضعیت فعال بودن قانون را تنظیم کنید.">
          <div className="grid gap-4 lg:grid-cols-2">
            <FormField label="شناسه قانون *" htmlFor="rule_id" description="فقط از حروف کوچک انگلیسی، عدد، خط تیره و آندرلاین استفاده کنید.">
              <Input
                id="rule_id"
                placeholder={t("form.ruleIdPlaceholder")}
                aria-invalid={Boolean(form.formState.errors.rule_id)}
                {...form.register("rule_id")}
              />
              <FieldError message={form.formState.errors.rule_id?.message} />
            </FormField>

            <FormField label="نام نمایشی" htmlFor="name" description="برای تشخیص سریع قانون در فهرست‌ها و گزارش‌ها استفاده می‌شود.">
              <Input id="name" placeholder="مثال: محدودیت کاربران عمومی" {...form.register("name")} />
            </FormField>
          </div>

          <FormField label="توضیحات" htmlFor="description" description="هدف قانون یا سناریوی استفاده را برای اعضای تیم توضیح دهید.">
            <Textarea id="description" placeholder="این قانون برای چه نوع درخواست‌هایی اعمال می‌شود؟" {...form.register("description")} />
          </FormField>

          <SwitchRow
            title="وضعیت قانون"
            description="در صورت غیرفعال بودن، قانون ذخیره می‌شود اما روی ترافیک اعمال نخواهد شد."
            checked={Boolean(enabled)}
            onCheckedChange={(checked) => form.setValue("enabled", checked, { shouldValidate: true, shouldDirty: true })}
          />
        </FormSection>

        <FormSection title="شرایط و فیلترها" description="مشخص کنید قانون برای چه مبداها، مدل‌ها یا کلیدهای فراداده فعال شود.">
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <p className="text-sm leading-6 text-muted-foreground">
                می‌توانید برای سوژه، مدل یا چند ردیف فراداده شرط تعریف کنید. هر شرط جدید فقط از گزینه‌های معتبر این قانون استفاده می‌کند.
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
                هنوز فیلتری اضافه نشده است. از دکمه «افزودن فیلتر» برای تعریف مبدا، مدل یا فراداده استفاده کنید.
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
                      ariaLabel="نوع فیلتر"
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
                      افزودن
                    </Button>

                    <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label="حذف فیلتر">
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
                  <h3 className="text-base font-semibold text-card-foreground">فراداده</h3>
                  <p className="text-sm leading-6 text-muted-foreground">برای سناریوهای وابسته به کلیدهای متنی یا عددی متادیتا، چند شرط مجزا تعریف کنید.</p>
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
                            placeholder="کلید فراداده"
                            aria-invalid={Boolean(keyError)}
                            {...form.register(`conditions.conditions.${index}.metadataKey`)}
                          />

                          <InlineSelect
                            value={watchedConditions?.[index]?.operator ?? "equals"}
                            ariaLabel="عملگر فراداده"
                            onValueChange={(value) =>
                              form.setValue(`conditions.conditions.${index}.operator`, value as (typeof metadataOperatorOptions)[number], {
                                shouldValidate: true,
                                shouldDirty: true,
                              })
                            }
                            options={metadataOperatorOptions.map((option) => ({ value: option, label: metadataOperatorLabel[option] }))}
                          />

                          <Input
                            placeholder="مقدار"
                            aria-invalid={Boolean(valueError)}
                            {...form.register(`conditions.conditions.${index}.value` as const)}
                          />

                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            aria-label="افزودن ردیف فراداده"
                            onClick={() => append({ field: "metadata", operator: "equals", values: undefined })}
                          >
                            <Plus />
                          </Button>

                          <Button type="button" variant="outline" size="icon" onClick={() => remove(index)} aria-label={`حذف ردیف فراداده ${itemIndex + 1}`}>
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

        <FormSection title="محدودیت و بازه زمانی" description="ظرفیت مجاز قانون و بازه ارزیابی آن را مشخص کنید.">
          <div className="grid gap-4 lg:grid-cols-4">
            <FormField label="سقف محدودیت *" htmlFor="limit-max">
              <Input
                id="limit-max"
                type="number"
                min={1}
                aria-invalid={Boolean(form.formState.errors.limit?.max)}
                {...form.register("limit.max", { valueAsNumber: true })}
              />
              <FieldError message={form.formState.errors.limit?.max?.message} />
            </FormField>

            <FormField label="واحد محدودیت">
              <InlineSelect
                value={limitDimension}
                ariaLabel="واحد محدودیت"
                onValueChange={(value) => form.setValue("limit.dimension", value as "requests" | "tokens", { shouldValidate: true, shouldDirty: true })}
                options={[
                  { value: "requests", label: "تعداد درخواست" },
                  { value: "tokens", label: "توکن" },
                ]}
              />
            </FormField>

            <FormField label="بازه زمانی *">
              <InlineSelect
                value={selectedWindowUnit}
                ariaLabel="بازه زمانی"
                onValueChange={(value) =>
                  form.setValue("limit.window.unit", value as "second" | "minute" | "hour" | "day" | "week" | "month", {
                    shouldValidate: true,
                    shouldDirty: true,
                  })
                }
                options={[
                  { value: "minute", label: "در دقیقه" },
                  { value: "hour", label: "در ساعت" },
                  { value: "day", label: "در روز" },
                  { value: "week", label: "در هفته" },
                  { value: "month", label: "در ماه" },
                  { value: "second", label: "سفارشی (بر حسب ثانیه)" },
                ]}
              />
            </FormField>

            <FormField
              label="مقدار بازه *"
              htmlFor="limit-window-value"
              description={selectedWindowUnit === "second" ? "در حالت سفارشی، مقدار وارد شده بر حسب ثانیه محاسبه می‌شود." : undefined}
            >
              <Input
                id="limit-window-value"
                type="number"
                min={1}
                step={1}
                placeholder={selectedWindowUnit === "second" ? "مثال: 90" : "مثال: 1"}
                aria-invalid={Boolean(form.formState.errors.limit?.window?.value)}
                {...form.register("limit.window.value", { valueAsNumber: true })}
              />
              <FieldError message={form.formState.errors.limit?.window?.value?.message} />
            </FormField>
          </div>
        </FormSection>

        <FormSection title="تنظیمات پیشرفته" description="رفتار برخورد با تعارض‌ها و نحوه پاسخ‌دهی قانون را در این بخش مدیریت کنید.">
          <div className="flex flex-col gap-4">
            <Button
              type="button"
              variant="outline"
              className="w-full justify-between md:w-fit"
              onClick={() => setAdvancedOpen((prev) => !prev)}
            >
              <span>{advancedOpen ? "بستن تنظیمات پیشرفته" : "نمایش تنظیمات پیشرفته"}</span>
              {advancedOpen ? <ChevronUp data-icon="inline-end" /> : <ChevronDown data-icon="inline-end" />}
            </Button>

            {advancedOpen ? (
              <div className="flex flex-col gap-4 rounded-xl border border-border bg-muted/20 p-4">
                <div className="grid gap-4 lg:grid-cols-2">
                  <FormField label="اولویت" htmlFor="priority" description="عدد کمتر یا بیشتر، ترتیب اعمال را در زمان تعارض بین قوانین تعیین می‌کند.">
                    <Input id="priority" type="number" min={0} {...form.register("priority", { valueAsNumber: true })} />
                  </FormField>

                  <FormField label="رفتار تعارض">
                    <InlineSelect
                      value={conflictBehavior}
                      ariaLabel="رفتار تعارض"
                      onValueChange={(value) =>
                        form.setValue("conflict_behavior", value as "first_match" | "most_restrictive" | "apply_all", {
                          shouldValidate: true,
                          shouldDirty: true,
                        })
                      }
                      options={[
                        { value: "first_match", label: "اولین قانون منطبق" },
                        { value: "most_restrictive", label: "محدودکننده‌ترین قانون" },
                        { value: "apply_all", label: "اعمال همه قوانین منطبق" },
                      ]}
                    />
                  </FormField>
                </div>

                <div className="grid gap-4 lg:grid-cols-2">
                  <SwitchRow
                    title="حالت نرم"
                    description="به‌جای مسدود کردن درخواست، فقط ثبت و مانیتور می‌شود."
                    checked={Boolean(softLimit)}
                    onCheckedChange={(checked) => form.setValue("soft_limit", checked, { shouldValidate: true, shouldDirty: true })}
                  />

                  <SwitchRow
                    title="ارسال هدر Retry-After"
                    description="در پاسخ‌های محدود شده، زمان مناسب برای ارسال مجدد درخواست برگردانده می‌شود."
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
              پس از ذخیره، به فهرست قوانین محدودسازی نرخ برمی‌گردید.
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
}: {
  label: string;
  onRemove: () => void;
}) {
  return (
    <Badge variant="outline" className="gap-1 rounded-md px-2 py-1 text-xs font-normal">
      <span className="truncate">{label}</span>
      <button
        type="button"
        className="rounded-sm text-muted-foreground transition-colors hover:text-foreground"
        onClick={onRemove}
        aria-label={`حذف ${label}`}
      >
        <X className="size-3" />
      </button>
    </Badge>
  );
}

function SwitchRow({
  title,
  description,
  checked,
  onCheckedChange,
}: {
  title: string;
  description: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  return (
    <div className={cn("flex items-start justify-between gap-4 rounded-xl border border-border bg-card px-4 py-3")}>
      <div className="flex min-w-0 flex-col gap-1">
        <p className="text-sm font-medium text-card-foreground">{title}</p>
        <p className="text-xs leading-5 text-muted-foreground">{description}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  );
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return <p className="text-xs leading-5 text-destructive">{message}</p>;
}
