"use client";

import { useEffect, useMemo, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { ChevronDown, ChevronUp, Plus, Trash2, X } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
  orgUnits: "واحدهای سازمانی",
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

  const form = useForm<RateLimitRuleFormInput>({ resolver: zodResolver(rateLimitRuleSchema), mode: "onChange", defaultValues });
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "conditions.conditions" });
  const watchedConditions = form.watch("conditions.conditions");
  const selectedWindowUnit = form.watch("limit.window.unit");
  const metadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => form.watch(`conditions.conditions.${x.i}.field`) === "metadata").map((x) => x.i);
  const nonMetadataIndexes = fields.map((f, i) => ({ f, i })).filter((x) => form.watch(`conditions.conditions.${x.i}.field`) !== "metadata").map((x) => x.i);
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
        const nextAllowed = (modelsPayload.data ?? []).map((item) => {
          const name = item.model ?? item.name ?? item.id ?? "";
          return {
            id: name,
            name,
            kind: "provider_allowed" as const,
            secondary: item.provider_name ?? item.provider,
          };
        }).filter((item) => item.name);
        setProviderAllowedModels(nextAllowed);
      } catch {}

      try {
        const virtualPayload = await requestJSON<{ data?: Array<{ id?: string; alias?: string }> }>("/api/virtual-models");
        const nextVirtual = (virtualPayload.data ?? []).map((item) => {
          const id = item.id ?? item.alias ?? "";
          const name = item.alias ?? item.id ?? "";
          return { id, name, kind: "virtual_model" as const };
        }).filter((item) => item.id && item.name);
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
    const normalized: RateLimitRuleFormInput = {
      ...values,
      conditions: {
        ...values.conditions,
        conditions: values.conditions.conditions.map((c) => normalizeCondition(c as RateLimitCondition) as never),
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

  return (
    <Card>
      <CardHeader>
        <CardTitle>{initialRule ? "ویرایش قانون" : "افزودن قانون جدید"}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? "خطا" : "موفق"}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
          <section className="space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">هویت قانون</h2>
            <div className="grid gap-3">
              <div className="space-y-2">
                <Label htmlFor="rule_id">شناسه قانون *</Label>
                <Input id="rule_id" placeholder="شناسه قانون را وارد کنید" {...form.register("rule_id")} />
                <FieldError message={form.formState.errors.rule_id?.message} />
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" {...form.register("enabled")} />
              فعال
            </label>
          </section>

          <Separator />
          <section className="space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">هنگام ورود درخواست</h2>
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
                افزودن فیلتر
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
                            {fieldOptions.filter((option) => option === rowField || !usedByOthers.has(option)).map((option) => (
                              <option key={option} value={option}>{fieldLabel[option]}</option>
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
                            افزودن
                          </Button>

                          <Button type="button" variant="outline" onClick={() => remove(index)} aria-label="حذف فیلتر">
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>

                        <div className="flex flex-wrap gap-2">
                          {subjectTokens.map((token) => {
                            const item = toSubjectSelection(token);
                            const displayName = subjectDisplayNameByRow[index]?.[token] ?? item.display_name ?? item.principal_id ?? token;
                            return (
                              <span key={token} className="inline-flex items-center gap-1 rounded-md border bg-muted px-2 py-1 text-xs">
                                <span>{item.type}: {displayName}</span>
                                <button
                                  type="button"
                                  className="text-muted-foreground hover:text-foreground"
                                  onClick={() => {
                                    form.setValue(`conditions.conditions.${index}.values`, subjectTokens.filter((value) => value !== token), { shouldValidate: true });
                                    setSubjectDisplayNameByRow((prev) => {
                                      const nextRow = { ...(prev[index] ?? {}) };
                                      delete nextRow[token];
                                      return { ...prev, [index]: nextRow };
                                    });
                                  }}
                                  aria-label="حذف مورد"
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
                          {fieldOptions.filter((option) => option === rowField || !usedByOthers.has(option)).map((option) => (
                            <option key={option} value={option}>{fieldLabel[option]}</option>
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
                            افزودن
                          </Button>

                        <Button type="button" variant="outline" onClick={() => remove(index)} aria-label="حذف فیلتر">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {(((form.watch(`conditions.conditions.${index}.values`) as string[] | undefined) ?? []).filter((item) => typeof item === "string" && item.trim().length > 0)).map((token) => (
                            <span key={token} className="inline-flex items-center gap-1 rounded-md border bg-muted px-2 py-1 text-xs">
                              <span>{token}</span>
                              <button
                                type="button"
                                className="text-muted-foreground hover:text-foreground"
                                onClick={() => {
                                  const modelValues = ((form.watch(`conditions.conditions.${index}.values`) as string[] | undefined) ?? []).filter(Boolean);
                                  form.setValue(`conditions.conditions.${index}.values`, modelValues.filter((value) => value !== token), { shouldValidate: true });
                                }}
                                aria-label="حذف مدل"
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
                  <div className="mb-2 text-sm font-medium">فراداده</div>
                  <div className="space-y-2">
                    {metadataIndexes.map((index, idx) => (
                      <div key={fields[index]?.id ?? `metadata-${index}`} className="grid gap-2 md:grid-cols-[1fr_1fr_1fr_auto_auto]">
                        <Input placeholder="کلید" {...form.register(`conditions.conditions.${index}.metadataKey`)} />
                        <select className="h-10 rounded-md border px-2" {...form.register(`conditions.conditions.${index}.operator`)}>
                          {metadataOperatorOptions.map((option) => (
                            <option key={option} value={option}>{metadataOperatorLabel[option]}</option>
                          ))}
                        </select>
                        <Input placeholder="مقدار" {...form.register(`conditions.conditions.${index}.value` as const)} />
                        <Button
                          type="button"
                          variant="outline"
                          aria-label="افزودن ردیف فراداده"
                          onClick={() => append({ field: "metadata", operator: "equals", values: undefined })}
                        >
                          <Plus className="h-4 w-4" />
                        </Button>
                        <Button type="button" variant="outline" onClick={() => remove(index)} aria-label="حذف ردیف فراداده">
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
            <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">اعمال محدودیت</h2>
            <div className="grid gap-3 md:grid-cols-4">
              <div className="space-y-2">
                <Label>سقف محدودیت *</Label>
                <Input type="number" min={1} {...form.register("limit.max", { valueAsNumber: true })} />
                <FieldError message={form.formState.errors.limit?.max?.message} />
              </div>
              <div className="space-y-2">
                <Label>واحد</Label>
                <select className="h-10 w-full rounded-md border px-2" {...form.register("limit.dimension")}>
                  <option value="requests">تعداد درخواست</option>
                  <option value="tokens">توکن</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>بازه زمانی *</Label>
                <select className="h-10 w-full rounded-md border px-2" {...form.register("limit.window.unit")}>
                  <option value="minute">در دقیقه</option>
                  <option value="hour">در ساعت</option>
                  <option value="day">در روز</option>
                  <option value="week">در هفته</option>
                  <option value="month">در ماه</option>
                  <option value="second">سفارشی (بر حسب ثانیه)</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>مقدار بازه *</Label>
                <Input
                  type="number"
                  min={1}
                  step={1}
                  placeholder={selectedWindowUnit === "second" ? "مثال: 90" : "مثال: 1"}
                  {...form.register("limit.window.value", { valueAsNumber: true })}
                />
                <FieldError message={form.formState.errors.limit?.window?.value?.message} />
              </div>
            </div>
            {selectedWindowUnit === "second" ? (
              <p className="text-xs text-muted-foreground">در حالت سفارشی، بازه به ثانیه محاسبه می‌شود (مثلاً 90 یعنی 90 ثانیه).</p>
            ) : null}
          </section>

          <Separator />
          <section className="space-y-3">
            <button type="button" className="flex items-center gap-2 text-sm font-medium" onClick={() => setAdvancedOpen((prev) => !prev)}>
              {advancedOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
              تنظیمات پیشرفته
            </button>
            {advancedOpen ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>اولویت</Label>
                  <Input type="number" min={0} {...form.register("priority", { valueAsNumber: true })} />
                </div>
                <div className="space-y-2">
                  <Label>رفتار تعارض</Label>
                  <select className="h-10 w-full rounded-md border px-2" {...form.register("conflict_behavior")}>
                    <option value="first_match">اولین قانون منطبق</option>
                    <option value="most_restrictive">محدودکننده‌ترین قانون</option>
                    <option value="apply_all">اعمال همه قوانین منطبق</option>
                  </select>
                </div>
                <label className="flex items-center gap-2 text-sm">
                  <input type="checkbox" {...form.register("soft_limit")} />
                  حالت نرم (فقط ثبت هشدار)
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input type="checkbox" {...form.register("retry_after_header")} />
                  ارسال هدر Retry-After
                </label>
              </div>
            ) : null}
          </section>

          <div className="sticky bottom-0 flex justify-end gap-2 border-t bg-background pt-3">
            <Button type="button" variant="outline" onClick={() => router.push("/gateway-policies/rate-limiting-rules")}>
              انصراف
            </Button>
            <Button type="submit" disabled={!form.formState.isValid || isSaving}>
              {isSaving ? "در حال ذخیره..." : "ذخیره"}
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
