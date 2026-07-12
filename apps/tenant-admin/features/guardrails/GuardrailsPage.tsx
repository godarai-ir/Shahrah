"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Check, Copy, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";

import { AppSelectField, EmptyState, HelpLink, IdentifierField, OperationNotice } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";

import { createGuardrail, deleteGuardrail, GuardrailAPIError, listGuardrails, updateGuardrail } from "./api";
import type { GuardrailFormInput, GuardrailPayload, GuardrailType, TenantGuardrail } from "./types";

const emptyForm: GuardrailFormInput = {
  slug: "",
  name: "",
  description: "",
  type: "keyword_blocklist",
  operation: "validate",
  enforcement: "enforce",
  enabled: true,
  terms: "",
  patterns: "",
  pattern: "",
  replacement: "[***]",
  redactionText: "[***]",
  min: "",
  max: "",
  schema: '{\n  "type": "object"\n}',
  keys: "",
  keyMode: "all",
  languages: "",
  codeMode: "must_contain",
  tokenRegex: "",
  issuer: "",
  audience: "",
  sharedSecret: "",
  inlineJwks: "",
  jwksUri: "",
  introspectionUrl: "",
  introspectionAuthHeader: "",
  maxBytes: "1048576",
  allowedHosts: "",
  timeoutMs: "3000",
  allowPrivateNetworks: false,
};

const guardrailTypes: GuardrailType[] = [
  "keyword_blocklist",
  "regex",
  "regex_replace",
  "secrets_detection",
  "basic_pii",
  "jwt_validation",
  "sentence_count",
  "word_count",
  "character_count",
  "json_schema",
  "json_keys",
  "contains_code",
  "not_null",
  "inline_image_urls",
];

function formatDateTime(value: string, locale: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  const dateLocale = locale === "fa" ? "fa-IR-u-ca-persian" : "en-US";
  return new Intl.DateTimeFormat(dateLocale, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function lines(value: string) {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function labelForType(type: GuardrailType, t: ReturnType<typeof useTranslations<"guardrails">>) {
  return t(`types.${type}`);
}

function numberOrUndefined(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function parseJSONOrEmpty(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  return JSON.parse(trimmed) as unknown;
}

function stringFromConfig(config: Record<string, unknown>, key: string, fallback = "") {
  const value = config[key];
  return typeof value === "string" ? value : fallback;
}

function numberStringFromConfig(config: Record<string, unknown>, key: string) {
  const value = config[key];
  return typeof value === "number" ? String(value) : "";
}

function listStringFromConfig(config: Record<string, unknown>, key: string) {
  const value = config[key];
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string").join("\n") : "";
}

function jsonStringFromConfig(config: Record<string, unknown>, key: string, fallback: string) {
  const value = config[key];
  if (!value || typeof value !== "object") return fallback;
  return JSON.stringify(value, null, 2);
}

function isValidateOnlyType(type: GuardrailType) {
  return ["keyword_blocklist", "sentence_count", "word_count", "character_count", "json_schema", "json_keys", "contains_code", "not_null", "jwt_validation"].includes(type);
}

function isMutateOnlyType(type: GuardrailType) {
  return type === "regex_replace" || type === "inline_image_urls";
}

function operationForType(type: GuardrailType, operation: GuardrailFormInput["operation"]) {
  if (isValidateOnlyType(type)) return "validate";
  if (isMutateOnlyType(type)) return "mutate";
  return operation;
}

function labelForOperation(operation: GuardrailFormInput["operation"], t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (operation === "mutate") {
    return t.has("operations.mutate") ? t("operations.mutate") : "mutate";
  }
  return t.has("operations.validate") ? t("operations.validate") : "validate";
}

function labelForEnforcement(enforcement: GuardrailFormInput["enforcement"], t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (enforcement === "audit") {
    return t.has("enforcement.audit") ? t("enforcement.audit") : "audit";
  }
  if (enforcement === "enforce_ignore_error") {
    return t.has("enforcement.enforce_ignore_error") ? t("enforcement.enforce_ignore_error") : "enforce_ignore_error";
  }
  return t.has("enforcement.enforce") ? t("enforcement.enforce") : "enforce";
}

function formFromGuardrail(item: TenantGuardrail): GuardrailFormInput {
  return {
    slug: item.slug,
    name: item.name,
    description: item.description ?? "",
    type: item.type,
    operation: item.operation,
    enforcement: item.enforcement,
    enabled: item.enabled,
    terms: listStringFromConfig(item.config, "terms"),
    patterns: listStringFromConfig(item.config, "patterns"),
    pattern: stringFromConfig(item.config, "pattern"),
    replacement: stringFromConfig(item.config, "replacement", "[***]"),
    redactionText: typeof item.config.redaction_text === "string" ? item.config.redaction_text : "[***]",
    min: numberStringFromConfig(item.config, "min"),
    max: numberStringFromConfig(item.config, "max"),
    schema: jsonStringFromConfig(item.config, "schema", '{\n  "type": "object"\n}'),
    keys: listStringFromConfig(item.config, "keys"),
    keyMode: item.config.mode === "any" || item.config.mode === "none" ? item.config.mode : "all",
    languages: listStringFromConfig(item.config, "languages"),
    codeMode: item.config.mode === "must_not_contain" ? "must_not_contain" : "must_contain",
    tokenRegex: stringFromConfig(item.config, "token_regex"),
    issuer: stringFromConfig(item.config, "issuer"),
    audience: stringFromConfig(item.config, "audience"),
    sharedSecret: stringFromConfig(item.config, "shared_secret"),
    inlineJwks: jsonStringFromConfig(item.config, "inline_jwks", ""),
    jwksUri: stringFromConfig(item.config, "jwks_uri"),
    introspectionUrl: stringFromConfig(item.config, "introspection_url"),
    introspectionAuthHeader: stringFromConfig(item.config, "introspection_auth_header"),
    maxBytes: numberStringFromConfig(item.config, "max_bytes") || "1048576",
    allowedHosts: listStringFromConfig(item.config, "allowed_hosts"),
    timeoutMs: numberStringFromConfig(item.config, "timeout_ms") || "3000",
    allowPrivateNetworks: item.config.allow_private_networks === true,
  };
}

function buildPayload(form: GuardrailFormInput, mode: "create" | "edit"): GuardrailPayload {
  const type = form.type;
  const config: Record<string, unknown> = {};
  if (type === "keyword_blocklist") {
    config.terms = lines(form.terms);
  } else if (type === "regex") {
    config.patterns = lines(form.patterns);
    config.redaction_text = form.redactionText.trim() || "[***]";
  } else if (type === "regex_replace") {
    config.pattern = form.pattern.trim();
    config.replacement = form.replacement;
    config.case_sensitive = true;
  } else if (type === "sentence_count" || type === "word_count" || type === "character_count") {
    const min = numberOrUndefined(form.min);
    const max = numberOrUndefined(form.max);
    if (min !== undefined) config.min = min;
    if (max !== undefined) config.max = max;
  } else if (type === "json_schema") {
    config.schema = parseJSONOrEmpty(form.schema);
  } else if (type === "json_keys") {
    config.keys = lines(form.keys);
    config.mode = form.keyMode;
  } else if (type === "contains_code") {
    config.languages = lines(form.languages);
    config.mode = form.codeMode;
    config.must_contain = form.codeMode === "must_contain";
  } else if (type === "jwt_validation") {
    if (form.tokenRegex.trim()) config.token_regex = form.tokenRegex.trim();
    if (form.issuer.trim()) config.issuer = form.issuer.trim();
    if (form.audience.trim()) config.audience = form.audience.trim();
    if (form.sharedSecret.trim()) config.shared_secret = form.sharedSecret.trim();
    if (form.inlineJwks.trim()) config.inline_jwks = parseJSONOrEmpty(form.inlineJwks);
    if (form.jwksUri.trim()) config.jwks_uri = form.jwksUri.trim();
    if (form.introspectionUrl.trim()) config.introspection_url = form.introspectionUrl.trim();
    if (form.introspectionAuthHeader.trim()) config.introspection_auth_header = form.introspectionAuthHeader.trim();
  } else if (type === "inline_image_urls") {
    const maxBytes = numberOrUndefined(form.maxBytes);
    const timeoutMs = numberOrUndefined(form.timeoutMs);
    if (maxBytes !== undefined) config.max_bytes = maxBytes;
    if (timeoutMs !== undefined) config.timeout_ms = timeoutMs;
    config.allowed_hosts = lines(form.allowedHosts);
    config.allow_private_networks = form.allowPrivateNetworks;
  } else if (type !== "not_null") {
    config.redaction_text = form.redactionText.trim() || "[***]";
  }

  return {
    ...(mode === "create" ? { slug: form.slug.trim() } : {}),
    name: form.name.trim(),
    description: form.description.trim(),
    type,
    operation: operationForType(type, form.operation),
    enforcement: form.enforcement,
    enabled: form.enabled,
    config,
  };
}

function validateForm(form: GuardrailFormInput, t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,95}$/.test(form.slug.trim())) {
    return t("validation.slug");
  }
  if (!form.name.trim()) return t("validation.name");
  if (form.type === "keyword_blocklist" && lines(form.terms).length === 0) {
    return t("validation.terms");
  }
  if (form.type === "regex" && lines(form.patterns).length === 0) {
    return t("validation.patterns");
  }
  if (form.type === "regex_replace" && !form.pattern.trim()) return t("validation.pattern");
  if ((form.type === "sentence_count" || form.type === "word_count" || form.type === "character_count") && !form.min.trim() && !form.max.trim()) {
    return t("validation.range");
  }
  if (form.type === "json_schema") {
    try {
      parseJSONOrEmpty(form.schema);
    } catch {
      return t("validation.json");
    }
  }
  if (form.type === "json_keys" && lines(form.keys).length === 0) return t("validation.keys");
  if (form.type === "jwt_validation") {
    const hasVerifier = form.sharedSecret.trim() || form.inlineJwks.trim() || form.jwksUri.trim() || form.introspectionUrl.trim();
    if (!hasVerifier) return t("validation.jwtVerifier");
    if (form.inlineJwks.trim()) {
      try {
        parseJSONOrEmpty(form.inlineJwks);
      } catch {
        return t("validation.json");
      }
    }
  }
  return "";
}

function localizedError(error: unknown, fallback: string, t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (!(error instanceof GuardrailAPIError)) {
    return error instanceof Error ? error.message : fallback;
  }
  if (error.status === 409) return t("messages.duplicate");
  if (error.status === 400) return t("messages.invalid");
  if (error.status === 403) return t("messages.forbidden");
  if (error.status === 404) return t("messages.notFound");
  if (error.status && error.status >= 500) return t("messages.serverError");
  return fallback;
}

export function GuardrailsPage({ docsHref }: { docsHref: string }) {
  const t = useTranslations("guardrails");
  const locale = useLocale();
  const direction = locale === "fa" ? "rtl" : "ltr";
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<TenantGuardrail[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [editing, setEditing] = useState<TenantGuardrail | null>(null);
  const [pendingDelete, setPendingDelete] = useState<TenantGuardrail | null>(null);
  const [form, setForm] = useState<GuardrailFormInput>(emptyForm);
  const [formOpen, setFormOpen] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [copiedSelector, setCopiedSelector] = useState("");

  async function load() {
    setIsLoading(true);
    try {
      const response = await listGuardrails(query);
      setItems(response.data);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.loadError"), t) });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const tableItems = useMemo(() => items, [items]);
  const hasSearchQuery = query.trim().length > 0;

  function openCreateDialog() {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  }

  function openEditDialog(item: TenantGuardrail) {
    setEditing(item);
    setForm(formFromGuardrail(item));
    setFormOpen(true);
  }

  async function submitForm() {
    if (isSaving) return;
    const validationError = validateForm(form, t);
    if (validationError) {
      setNotice({ kind: "error", message: validationError });
      return;
    }
    setIsSaving(true);
    try {
      const payload = buildPayload(form, editing ? "edit" : "create");
      if (editing) {
        await updateGuardrail(editing.slug, payload);
      } else {
        await createGuardrail(payload);
      }
      setFormOpen(false);
      setNotice({ kind: "success", message: t("messages.saveSuccess") });
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.saveError"), t) });
    } finally {
      setIsSaving(false);
    }
  }

  async function confirmDelete() {
    if (!pendingDelete || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteGuardrail(pendingDelete.slug);
      setPendingDelete(null);
      setNotice({ kind: "success", message: t("messages.deleteSuccess") });
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.deleteError"), t) });
    } finally {
      setIsDeleting(false);
    }
  }

  async function copySelector(selector: string) {
    await navigator.clipboard.writeText(selector);
    setCopiedSelector(selector);
    window.setTimeout(() => setCopiedSelector(""), 1600);
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="min-w-0">
              <CardTitle className="flex items-center gap-2">
                <ShieldCheck data-icon="inline-start" aria-hidden="true" />
                {t("title")}
              </CardTitle>
              <CardDescription>
                {t("description")} {" "}
                <HelpLink href={docsHref} />
              </CardDescription>
            </div>
            <Button type="button" onClick={openCreateDialog}>
              <Plus data-icon="inline-start" aria-hidden="true" />
              {t("actions.new")}
            </Button>
          </div>
          <Input placeholder={t("searchPlaceholder")} value={query} onChange={(event) => setQuery(event.target.value)} />
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

          {tableItems.length === 0 && !isLoading ? (
            <EmptyState
              variant="inline"
              density="compact"
              title={hasSearchQuery ? t("emptySearch") : t("empty")}
            />
          ) : (
            <div className="overflow-x-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("table.name")}</TableHead>
                    <TableHead>{t("table.type")}</TableHead>
                    <TableHead>{t("table.execution")}</TableHead>
                    <TableHead>{t("table.status")}</TableHead>
                    <TableHead>{t("table.updatedAt")}</TableHead>
                    <TableHead>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tableItems.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div className="flex min-w-64 flex-col gap-1">
                          <span className="font-medium">{item.name}</span>
                          <div className="flex max-w-72 items-center gap-2">
                            <code className="truncate rounded bg-muted px-2 py-1 text-xs text-muted-foreground" dir="ltr">
                              {item.selector}
                            </code>
                            <Button
                              type="button"
                              size="icon"
                              variant="ghost"
                              aria-label={t("actions.copySelector")}
                              onClick={() => void copySelector(item.selector)}
                            >
                              {copiedSelector === item.selector ? (
                                <Check data-icon="inline-start" aria-hidden="true" />
                              ) : (
                                <Copy data-icon="inline-start" aria-hidden="true" />
                              )}
                            </Button>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{labelForType(item.type, t)}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          <Badge variant="secondary">{labelForOperation(item.operation, t)}</Badge>
                          <Badge variant="secondary">{labelForEnforcement(item.enforcement, t)}</Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={item.enabled ? "default" : "secondary"}>
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatDateTime(item.updated_at, locale)}</TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button type="button" size="icon" variant="outline" aria-label={t("actions.edit")} onClick={() => openEditDialog(item)}>
                            <Pencil data-icon="inline-start" aria-hidden="true" />
                          </Button>
                          <Button type="button" size="icon" variant="outline" aria-label={t("actions.delete")} onClick={() => setPendingDelete(item)}>
                            <Trash2 data-icon="inline-start" aria-hidden="true" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <Sheet open={formOpen} onOpenChange={setFormOpen}>
        <SheetContent side="left" className="flex h-full w-full max-w-2xl flex-col gap-0 p-0 sm:max-w-2xl">
          <SheetHeader className="border-b border-border px-6 py-5 text-start">
            <SheetTitle>{editing ? t("form.editTitle") : t("form.addTitle")}</SheetTitle>
            <SheetDescription className="sr-only">{t("form.description")}</SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto px-6 py-5">
            <div className="grid gap-3 md:grid-cols-2">
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-name">{t("form.name")}</Label>
                <Input id="guardrail-name" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
              </div>
              <IdentifierField
                id="guardrail-slug"
                label={t("form.identifier")}
                description={t("validation.slug")}
                pattern="[A-Za-z0-9][A-Za-z0-9_-]{1,95}"
                minLength={2}
                maxLength={96}
                value={form.slug}
                disabled={Boolean(editing)}
                onChange={(event) => setForm({ ...form, slug: event.target.value })}
              />
            </div>

            <div className="flex flex-col gap-2">
              <Label htmlFor="guardrail-description">{t("form.notes")}</Label>
              <Textarea id="guardrail-description" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
            </div>

            <div className="grid gap-3 md:grid-cols-3">
              <AppSelectField
                label={t("form.type")}
                dir={direction}
                value={form.type}
                onValueChange={(value) => {
                  const nextType = value as GuardrailType;
                  setForm({ ...form, type: nextType, operation: operationForType(nextType, form.operation) });
                }}
                options={guardrailTypes.map((type) => ({ value: type, label: t(`types.${type}`) }))}
              />
              <AppSelectField
                label={t("form.operation")}
                dir={direction}
                value={operationForType(form.type, form.operation)}
                disabled={isValidateOnlyType(form.type) || isMutateOnlyType(form.type)}
                onValueChange={(value) => setForm({ ...form, operation: value as GuardrailFormInput["operation"] })}
                options={[
                  { value: "validate", label: t("operations.validate") },
                  { value: "mutate", label: t("operations.mutate") },
                ]}
              />
              <AppSelectField
                label={t("form.enforcement")}
                dir={direction}
                value={form.enforcement}
                onValueChange={(value) => setForm({ ...form, enforcement: value as GuardrailFormInput["enforcement"] })}
                options={[
                  { value: "enforce", label: t("enforcement.enforce") },
                  { value: "audit", label: t("enforcement.audit") },
                  { value: "enforce_ignore_error", label: t("enforcement.enforce_ignore_error") },
                ]}
              />
            </div>

            {form.type === "keyword_blocklist" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-terms">{t("form.terms")}</Label>
                <Textarea
                  id="guardrail-terms"
                  dir="ltr"
                  value={form.terms}
                  placeholder="ignore previous instructions&#10;drop table"
                  onChange={(event) => setForm({ ...form, terms: event.target.value })}
                />
              </div>
            ) : null}

            {form.type === "regex" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-patterns">{t("form.patterns")}</Label>
                <Textarea
                  id="guardrail-patterns"
                  dir="ltr"
                  value={form.patterns}
                  placeholder="(?i)password\\s*[:=]\\s*\\S+"
                  onChange={(event) => setForm({ ...form, patterns: event.target.value })}
                />
              </div>
            ) : null}

            {form.type === "regex_replace" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-pattern">{t("form.pattern")}</Label>
                  <Textarea
                    id="guardrail-pattern"
                    dir="ltr"
                    value={form.pattern}
                    placeholder="(?i)password\\s*[:=]\\s*\\S+"
                    onChange={(event) => setForm({ ...form, pattern: event.target.value })}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-replacement">{t("form.replacement")}</Label>
                  <Input id="guardrail-replacement" dir="ltr" value={form.replacement} onChange={(event) => setForm({ ...form, replacement: event.target.value })} />
                </div>
              </div>
            ) : null}

            {form.type === "sentence_count" || form.type === "word_count" || form.type === "character_count" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-min">{t("form.min")}</Label>
                  <Input id="guardrail-min" dir="ltr" inputMode="numeric" value={form.min} onChange={(event) => setForm({ ...form, min: event.target.value })} />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-max">{t("form.max")}</Label>
                  <Input id="guardrail-max" dir="ltr" inputMode="numeric" value={form.max} onChange={(event) => setForm({ ...form, max: event.target.value })} />
                </div>
              </div>
            ) : null}

            {form.type === "json_schema" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-schema">{t("form.schema")}</Label>
                <Textarea id="guardrail-schema" dir="ltr" className="min-h-40 font-mono" value={form.schema} onChange={(event) => setForm({ ...form, schema: event.target.value })} />
              </div>
            ) : null}

            {form.type === "json_keys" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-keys">{t("form.keys")}</Label>
                  <Textarea id="guardrail-keys" dir="ltr" value={form.keys} placeholder="answer&#10;confidence" onChange={(event) => setForm({ ...form, keys: event.target.value })} />
                </div>
                <AppSelectField
                  label={t("form.keyMode")}
                  dir={direction}
                  value={form.keyMode}
                  onValueChange={(value) => setForm({ ...form, keyMode: value as GuardrailFormInput["keyMode"] })}
                  options={[
                    { value: "all", label: t("keyModes.all") },
                    { value: "any", label: t("keyModes.any") },
                    { value: "none", label: t("keyModes.none") },
                  ]}
                />
              </div>
            ) : null}

            {form.type === "contains_code" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-languages">{t("form.languages")}</Label>
                  <Textarea id="guardrail-languages" dir="ltr" value={form.languages} placeholder="sql&#10;python&#10;typescript" onChange={(event) => setForm({ ...form, languages: event.target.value })} />
                </div>
                <AppSelectField
                  label={t("form.codeMode")}
                  dir={direction}
                  value={form.codeMode}
                  onValueChange={(value) => setForm({ ...form, codeMode: value as GuardrailFormInput["codeMode"] })}
                  options={[
                    { value: "must_contain", label: t("codeModes.must_contain") },
                    { value: "must_not_contain", label: t("codeModes.must_not_contain") },
                  ]}
                />
              </div>
            ) : null}

            {form.type === "jwt_validation" ? (
              <div className="flex flex-col gap-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-token-regex">{t("form.tokenRegex")}</Label>
                    <Input id="guardrail-token-regex" dir="ltr" value={form.tokenRegex} placeholder="Bearer capture regex" onChange={(event) => setForm({ ...form, tokenRegex: event.target.value })} />
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-shared-secret">{t("form.sharedSecret")}</Label>
                    <Input id="guardrail-shared-secret" dir="ltr" value={form.sharedSecret} onChange={(event) => setForm({ ...form, sharedSecret: event.target.value })} />
                  </div>
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-issuer">{t("form.issuer")}</Label>
                    <Input id="guardrail-issuer" dir="ltr" value={form.issuer} onChange={(event) => setForm({ ...form, issuer: event.target.value })} />
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-audience">{t("form.audience")}</Label>
                    <Input id="guardrail-audience" dir="ltr" value={form.audience} onChange={(event) => setForm({ ...form, audience: event.target.value })} />
                  </div>
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-jwks-uri">{t("form.jwksUri")}</Label>
                    <Input id="guardrail-jwks-uri" dir="ltr" value={form.jwksUri} onChange={(event) => setForm({ ...form, jwksUri: event.target.value })} />
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-introspection">{t("form.introspectionUrl")}</Label>
                    <Input id="guardrail-introspection" dir="ltr" value={form.introspectionUrl} onChange={(event) => setForm({ ...form, introspectionUrl: event.target.value })} />
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-introspection-auth">{t("form.introspectionAuthHeader")}</Label>
                  <Input id="guardrail-introspection-auth" dir="ltr" value={form.introspectionAuthHeader} placeholder="Bearer ..." onChange={(event) => setForm({ ...form, introspectionAuthHeader: event.target.value })} />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-inline-jwks">{t("form.inlineJwks")}</Label>
                  <Textarea id="guardrail-inline-jwks" dir="ltr" className="min-h-32 font-mono" value={form.inlineJwks} onChange={(event) => setForm({ ...form, inlineJwks: event.target.value })} />
                </div>
              </div>
            ) : null}

            {form.type === "inline_image_urls" ? (
              <div className="flex flex-col gap-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-max-bytes">{t("form.maxBytes")}</Label>
                    <Input id="guardrail-max-bytes" dir="ltr" inputMode="numeric" value={form.maxBytes} onChange={(event) => setForm({ ...form, maxBytes: event.target.value })} />
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="guardrail-timeout">{t("form.timeoutMs")}</Label>
                    <Input id="guardrail-timeout" dir="ltr" inputMode="numeric" value={form.timeoutMs} onChange={(event) => setForm({ ...form, timeoutMs: event.target.value })} />
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="guardrail-allowed-hosts">{t("form.allowedHosts")}</Label>
                  <Textarea id="guardrail-allowed-hosts" dir="ltr" value={form.allowedHosts} placeholder="cdn.example.com&#10;images.example.com" onChange={(event) => setForm({ ...form, allowedHosts: event.target.value })} />
                </div>
                <div className="flex items-center justify-between rounded-lg border px-3 py-2">
                  <Label htmlFor="guardrail-private-networks">{t("form.allowPrivateNetworks")}</Label>
                  <Switch id="guardrail-private-networks" checked={form.allowPrivateNetworks} onCheckedChange={(checked) => setForm({ ...form, allowPrivateNetworks: checked })} />
                </div>
              </div>
            ) : null}

            {form.type !== "keyword_blocklist" && form.type !== "regex_replace" && form.type !== "sentence_count" && form.type !== "word_count" && form.type !== "character_count" && form.type !== "json_schema" && form.type !== "json_keys" && form.type !== "contains_code" && form.type !== "not_null" && form.type !== "jwt_validation" && form.type !== "inline_image_urls" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-redaction">{t("form.redactionText")}</Label>
                <Input id="guardrail-redaction" dir="ltr" value={form.redactionText} onChange={(event) => setForm({ ...form, redactionText: event.target.value })} />
              </div>
            ) : null}

            <div className="flex items-center justify-between rounded-lg border px-3 py-2">
              <Label htmlFor="guardrail-enabled">{t("form.enabled")}</Label>
              <Switch id="guardrail-enabled" checked={form.enabled} onCheckedChange={(checked) => setForm({ ...form, enabled: checked })} />
            </div>
          </div>

          <SheetFooter className="flex-row border-t border-border px-6 py-4">
            <Button type="button" variant="outline" disabled={isSaving} onClick={() => setFormOpen(false)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" disabled={isSaving} onClick={() => void submitForm()}>
              {isSaving ? t("actions.saving") : t("actions.save")}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Dialog open={pendingDelete !== null} onOpenChange={(open) => !open && setPendingDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>{t("deleteDialog.description", { name: pendingDelete?.name ?? "-" })}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" disabled={isDeleting} onClick={() => setPendingDelete(null)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" variant="destructive" disabled={isDeleting} onClick={() => void confirmDelete()}>
              {isDeleting ? t("deleteDialog.deleting") : t("actions.delete")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
